package codexprovider

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Autumn-27/norma/llm"
)

const maxProtocolLine = 8 << 20
const invocationTimeout = 10 * time.Minute

type wireMessage struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code int `json:"code"`
	} `json:"error"`
}
type wireResult struct {
	message wireMessage
	err     error
}
type session struct {
	cmd              *exec.Cmd
	stdin            io.WriteCloser
	output           <-chan wireResult
	done             <-chan error
	stop             chan struct{}
	nextID           int
	threadID, turnID string
	usage            llm.Usage
	final            string
	finalID          string
	completed        bool
}

func (p *provider) run(ctx context.Context, req llm.CompletionRequest, prompt string) (string, llm.Usage, error) {
	ctx, cancel := context.WithTimeout(ctx, invocationTimeout)
	defer cancel()
	executable, err := resolveExecutable()
	if err != nil {
		return "", llm.Usage{}, err
	}
	scratch, err := os.MkdirTemp("", "scopeweaver-codex-")
	if err != nil {
		return "", llm.Usage{}, fmt.Errorf("codex: create isolated working directory: %w", err)
	}
	defer os.RemoveAll(scratch)
	runtimeOptions := p.options
	// Thinking is a harness hint, not a universal Codex effort value. In
	// particular, gpt-6.1-sol does not support "none"; preserve its configured
	// effort during context compaction. GLM explicitly supports low reasoning
	// as the documented replacement for disabled thinking.
	if req.Thinking == "disabled" && runtimeOptions.BaseURL == ZAIBaseURL {
		runtimeOptions.ReasoningEffort = "low"
	}
	args, env, err := processConfig(runtimeOptions, scratch)
	if err != nil {
		return "", llm.Usage{}, err
	}
	if p.options.BaseURL == ZAIBaseURL {
		env = replaceEnv(env, "CODEX_HOME", scratch)
	}
	s, err := startSession(executable, args, env, scratch)
	if err != nil {
		return "", llm.Usage{}, err
	}
	defer s.close()
	if _, err = s.request(ctx, "initialize", map[string]any{
		"clientInfo":   map[string]any{"name": "scopeweaver_cli", "version": "1"},
		"capabilities": map[string]any{"experimentalApi": true},
	}); err != nil {
		return "", s.usage, err
	}
	if err = s.writeContext(ctx, map[string]any{"method": "initialized"}); err != nil {
		return "", s.usage, err
	}
	if p.options.BaseURL == "" {
		if err = s.verifyChatGPTLogin(ctx); err != nil {
			return "", s.usage, err
		}
	}
	config, err := s.isolatedThreadConfig(ctx)
	if err != nil {
		return "", s.usage, err
	}
	// Private wire form: the Z.AI/GLM route requests arguments as a nested JSON
	// object, removing the JSON-inside-a-string double encoding where R20
	// observed invalid inner JSON inside an otherwise valid envelope; the OAuth
	// route keeps its known-working strict string schema and instructions.
	instructions := bridgeInstructions
	schema := outputSchema(req.Tools)
	if p.options.BaseURL == ZAIBaseURL {
		instructions = zaiBridgeInstructions
		schema = zaiOutputSchema(req.Tools)
	}
	threadParams := map[string]any{
		"model": p.options.Model, "ephemeral": true, "cwd": scratch,
		"approvalPolicy": "never", "sandbox": "read-only", "environments": []any{},
		"dynamicTools": []any{}, "runtimeWorkspaceRoots": []any{}, "selectedCapabilityRoots": []any{},
		"baseInstructions": instructions, "developerInstructions": strings.Join(req.System, "\n\n") + "\n\n" + instructions, "config": config,
	}
	if p.options.BaseURL == ZAIBaseURL {
		threadParams["modelProvider"] = "scopeweaver_zai"
	} else {
		threadParams["modelProvider"] = "openai"
	}
	result, err := s.request(ctx, "thread/start", threadParams)
	if err != nil {
		return "", s.usage, err
	}
	var started struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if json.Unmarshal(result, &started) != nil || started.Thread.ID == "" {
		return "", s.usage, errors.New("codex: thread/start returned no thread ID")
	}
	s.threadID = started.Thread.ID
	if err = s.verifyNoMCP(ctx); err != nil {
		return "", s.usage, err
	}
	turnParams := map[string]any{
		"threadId": s.threadID, "input": []any{map[string]any{"type": "text", "text": prompt}},
		"environments": []any{}, "outputSchema": schema, "approvalPolicy": "never",
		"summary": "none", "runtimeWorkspaceRoots": []any{}, "sandboxPolicy": map[string]any{"type": "readOnly"},
	}

	result, err = s.request(ctx, "turn/start", turnParams)
	if err != nil {
		return "", s.usage, err
	}
	var turnStarted struct {
		Turn struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	if json.Unmarshal(result, &turnStarted) != nil || turnStarted.Turn.ID == "" {
		return "", s.usage, errors.New("codex: turn/start returned no turn ID")
	}
	if s.turnID != "" && s.turnID != turnStarted.Turn.ID {
		return "", s.usage, errors.New("codex: turn/start ID does not match notifications")
	}
	s.turnID = turnStarted.Turn.ID
	for {
		if s.completed {
			if s.final == "" {
				return "", s.usage, errors.New("codex: completed turn has no structured assistant output")
			}
			return s.final, s.usage, nil
		}
		msg, err := s.receive(ctx)
		if err != nil {
			return "", s.usage, err
		}
		completed, err := s.notification(msg)
		if err != nil {
			return "", s.usage, err
		}
		if completed {
			if s.final == "" {
				return "", s.usage, errors.New("codex: completed turn has no structured assistant output")
			}
			return s.final, s.usage, nil
		}
	}
}

func startSession(path string, args, env []string, cwd string) (*session, error) {
	cmd := exec.Command(path, args...)
	cmd.Env = env
	cmd.Dir = cwd
	cmd.WaitDelay = 2 * time.Second
	// Diagnostics may contain prompts or authentication data. Drain them without
	// retaining or forwarding any bytes, including on process failure.
	cmd.Stderr = io.Discard
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("codex: stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("codex: stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("codex: start official CLI: %w", err)
	}
	output := make(chan wireResult, 32)
	stop := make(chan struct{})
	go func() {
		defer close(output)
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 64<<10), maxProtocolLine)
		for scanner.Scan() {
			var msg wireMessage
			err := json.Unmarshal(scanner.Bytes(), &msg)
			if err != nil {
				err = errors.New("codex: malformed app-server message")
			}
			select {
			case output <- wireResult{msg, err}:
			case <-stop:
				return
			}
			if err != nil {
				return
			}
		}
		if scanner.Err() != nil {
			select {
			case output <- wireResult{err: errors.New("codex: app-server output exceeded bounds or could not be read")}:
			case <-stop:
			}
		}
	}()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	return &session{cmd: cmd, stdin: stdin, output: output, done: done, stop: stop}, nil
}

func (s *session) close() {
	close(s.stop)
	// Only the process started by this request is touched. Closing input plus
	// killing it also bounds cleanup when the server is stalled or awaiting input.
	s.stdin.Close()
	s.cmd.Process.Kill()
	select {
	case <-s.done:
	case <-time.After(3 * time.Second):
	}
}

func (s *session) write(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("codex: encode app-server request: %w", err)
	}
	data = append(data, '\n')
	if _, err := s.stdin.Write(data); err != nil {
		return errors.New("codex: app-server input closed")
	}
	return nil
}

func (s *session) receive(ctx context.Context) (wireMessage, error) {
	select {
	case <-ctx.Done():
		return wireMessage{}, ctx.Err()
	case result, ok := <-s.output:
		if !ok {
			return wireMessage{}, errors.New("codex: app-server exited before completion")
		}
		if result.err != nil {
			return wireMessage{}, result.err
		}
		if result.message.Method != "" && len(result.message.ID) > 0 {
			// Never grant approvals, answer dynamic tools, or return authentication.
			// Closing the owned process denies the request without supplying data.
			return wireMessage{}, errors.New("codex: unexpected app-server request denied")
		}
		return result.message, nil
	}
}

func (s *session) request(ctx context.Context, method string, params any) (json.RawMessage, error) {
	s.nextID++
	id := s.nextID
	if err := s.writeContext(ctx, map[string]any{"id": id, "method": method, "params": params}); err != nil {
		return nil, err
	}
	for {
		msg, err := s.receive(ctx)
		if err != nil {
			return nil, err
		}
		if len(msg.ID) > 0 {
			if string(msg.ID) != strconv.Itoa(id) {
				return nil, errors.New("codex: unexpected app-server response ID")
			}
			if msg.Error != nil {
				return nil, fmt.Errorf("codex: %s failed (JSON-RPC code %d; provider diagnostic withheld)", method, msg.Error.Code)
			}
			if len(msg.Result) == 0 {
				return nil, errors.New("codex: app-server response has no result")
			}
			return msg.Result, nil
		}
		if _, err := s.notification(msg); err != nil {
			return nil, err
		}
	}
}

type protocolItem struct{ Type, ID, Text, Phase string }

func (s *session) checkItem(item protocolItem, completed bool) error {
	switch item.Type {
	case "userMessage", "reasoning":
	case "agentMessage":
		if completed && item.Phase != "commentary" {
			if len(item.Text) > maxOutputBytes {
				return errors.New("codex: structured output exceeds 2 MiB")
			}
			if s.finalID != "" && s.finalID != item.ID {
				return errors.New("codex: more than one final assistant output")
			}
			s.final, s.finalID = item.Text, item.ID
		}
	default:
		return errors.New("codex: unexpected built-in tool or non-text item; completion rejected")
	}
	return nil
}

func (s *session) notification(msg wireMessage) (bool, error) {
	if msg.Method == "" {
		return false, errors.New("codex: unexpected app-server message")
	}
	switch msg.Method {
	case "item/started", "item/completed":
		var event struct {
			ThreadID string       `json:"threadId"`
			TurnID   string       `json:"turnId"`
			Item     protocolItem `json:"item"`
		}
		if json.Unmarshal(msg.Params, &event) != nil {
			return false, errors.New("codex: malformed item notification")
		}
		if err := s.checkIDs(event.ThreadID, event.TurnID); err != nil {
			return false, err
		}
		return false, s.checkItem(event.Item, msg.Method == "item/completed")
	case "thread/tokenUsage/updated":
		var event struct {
			ThreadID   string `json:"threadId"`
			TurnID     string `json:"turnId"`
			TokenUsage struct {
				Total struct {
					Input      int `json:"inputTokens"`
					Output     int `json:"outputTokens"`
					Cached     int `json:"cachedInputTokens"`
					CacheWrite int `json:"cacheWriteInputTokens"`
				} `json:"total"`
			} `json:"tokenUsage"`
		}
		if json.Unmarshal(msg.Params, &event) != nil {
			return false, errors.New("codex: malformed usage notification")
		}
		if err := s.checkIDs(event.ThreadID, event.TurnID); err != nil {
			return false, err
		}
		u := event.TokenUsage.Total
		if u.Input < 0 || u.Output < 0 || u.Cached < 0 || u.CacheWrite < 0 || u.Cached > u.Input {
			return false, errors.New("codex: invalid token usage")
		}
		// total is cumulative for this one ephemeral thread. Repeated notifications
		// replace the snapshot rather than charging the same tokens again.
		s.usage = llm.Usage{InputTokens: u.Input, OutputTokens: u.Output, CacheReadTokens: u.Cached, CacheWriteTokens: u.CacheWrite}
	case "turn/completed":
		var event struct {
			ThreadID string `json:"threadId"`
			Turn     struct {
				ID, Status string
				Error      json.RawMessage
				Items      []protocolItem
			} `json:"turn"`
		}
		if json.Unmarshal(msg.Params, &event) != nil {
			return false, errors.New("codex: malformed completion notification")
		}
		if err := s.checkIDs(event.ThreadID, event.Turn.ID); err != nil {
			return false, err
		}
		if len(event.Turn.Error) > 0 && string(event.Turn.Error) != "null" {
			return false, safeTurnError(event.Turn.Error)
		}
		if event.Turn.Status != "completed" {
			return false, errors.New("codex: turn failed or was interrupted (provider diagnostic withheld)")
		}
		for _, item := range event.Turn.Items {
			if err := s.checkItem(item, true); err != nil {
				return false, err
			}
		}
		s.completed = true
		return true, nil
	case "error":
		var event struct {
			ThreadID  string          `json:"threadId"`
			TurnID    string          `json:"turnId"`
			Error     json.RawMessage `json:"error"`
			WillRetry bool            `json:"willRetry"`
		}
		if json.Unmarshal(msg.Params, &event) != nil {
			return false, errors.New("codex: malformed error notification")
		}
		if err := s.checkIDs(event.ThreadID, event.TurnID); err != nil {
			return false, err
		}
		if event.WillRetry {
			return false, nil
		}
		return false, safeTurnError(event.Error)
	default:
		// Legacy raw codex/event notifications mirror the v2 stream and must not be
		// charged twice. Tool-specific notifications fail closed even without items.
		if strings.HasPrefix(msg.Method, "item/") && !strings.HasPrefix(msg.Method, "item/agentMessage/") && !strings.HasPrefix(msg.Method, "item/reasoning/") {
			return false, errors.New("codex: unexpected built-in tool notification")
		}
	}
	return false, nil
}
func (s *session) checkIDs(thread, turn string) error {
	if s.threadID != "" && thread != s.threadID {
		return errors.New("codex: notification belongs to another thread")
	}
	if s.turnID == "" && turn != "" {
		s.turnID = turn
	}
	if s.turnID != "" && turn != s.turnID {
		return errors.New("codex: notification belongs to another turn")
	}
	return nil
}

func resolveExecutable() (string, error) {
	if override := strings.TrimSpace(os.Getenv("SCOPEWEAVER_CODEX_PATH")); override != "" {
		path, err := exec.LookPath(override)
		if err != nil {
			return "", errors.New("codex: SCOPEWEAVER_CODEX_PATH is not an executable")
		}
		if runtime.GOOS == "windows" && strings.ToLower(filepath.Ext(path)) != ".exe" {
			return "", errors.New("codex: SCOPEWEAVER_CODEX_PATH must name a native executable")
		}
		return path, nil
	}
	path, err := exec.LookPath("codex")
	if err == nil && (runtime.GOOS != "windows" || strings.EqualFold(filepath.Ext(path), ".exe")) {
		return path, nil
	}
	if runtime.GOOS == "windows" {
		// npm's .cmd shim needs a shell. Resolve only its official native package.
		roots := []string{filepath.Join(os.Getenv("APPDATA"), "npm")}
		if path != "" {
			roots = append([]string{filepath.Dir(path)}, roots...)
		}
		triplet := "x86_64-pc-windows-msvc"
		platform := "codex-win32-x64"
		if runtime.GOARCH == "arm64" {
			triplet = "aarch64-pc-windows-msvc"
			platform = "codex-win32-arm64"
		}
		for _, root := range roots {
			candidate := filepath.Join(root, "node_modules", "@openai", "codex", "node_modules", "@openai", platform, "vendor", triplet, "bin", "codex.exe")
			if info, e := os.Stat(candidate); e == nil && !info.IsDir() {
				return candidate, nil
			}
		}
	}
	return "", errors.New("codex: official Codex CLI executable not found; install it and run codex login, or set SCOPEWEAVER_CODEX_PATH")
}

func (s *session) writeContext(ctx context.Context, value any) error {
	result := make(chan error, 1)
	go func() { result <- s.write(value) }()
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		s.stdin.Close()
		s.cmd.Process.Kill()
		select {
		case <-result:
		case <-time.After(time.Second):
		}
		return ctx.Err()
	}
}

// Classify only documented machine-readable codes; raw diagnostics may contain
// prompts, credentials, or provider response bodies and must not escape here.
func safeTurnError(raw json.RawMessage) error {
	var detail struct {
		Info json.RawMessage `json:"codexErrorInfo"`
	}
	var code string
	if json.Unmarshal(raw, &detail) == nil && json.Unmarshal(detail.Info, &code) == nil {
		switch code {
		case "usageLimitExceeded":
			return errors.New("codex: quota exhausted (usageLimitExceeded)")
		case "rateLimitExceeded":
			return errors.New("codex: rate limit exceeded (rateLimitExceeded)")
		}
	}
	return errors.New("codex: app-server reported an error (provider diagnostic withheld)")
}
