package codexprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Autumn-27/norma/llm"
)

// Extend the existing subprocess, recording only whether fixed feedback arrived.
func fakeArgumentRepair(input json.RawMessage, normal string) string {
	mode := os.Getenv("SCOPEWEAVER_CODEX_FAKE_ARGUMENT_MODE")
	if mode == "" {
		return normal
	}
	var blocks []struct{ Text string }
	_ = json.Unmarshal(input, &blocks)
	feedback := len(blocks) == 1 && strings.Contains(blocks[0].Text, "ARGUMENT_FORMAT_CORRECTION")
	f, err := os.OpenFile(os.Getenv("SCOPEWEAVER_CODEX_FAKE_ARGUMENT_SESSIONS"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		os.Exit(30)
	}
	_, _ = f.WriteString(map[bool]string{false: "initial\n", true: "repair\n"}[feedback])
	_ = f.Close()
	bad := `{"assistant_text":"must not leak","tool_calls":[{"name":"echo","arguments":"{\"value\":\"must not execute\"}"},{"name":"echo","arguments":"[]"}]}`
	switch mode {
	case "transport", "native", "quota":
		return "fake-" + mode
	case "valid":
		return `{"assistant_text":"hello","tool_calls":[{"name":"echo","arguments":"{\"value\":\"hello\"}"}]}`
	case "malformed":
		return `{"assistant_text":"broken"`
	case "unknown":
		return `{"assistant_text":"","tool_calls":[{"name":"shell","arguments":"{}"}]}`
	case "duplicate":
		return `{"assistant_text":"","tool_calls":[{"name":"echo","arguments":"{\"a\":1,\"a\":2}"}]}`
	case "mixed-unknown":
		return `{"assistant_text":"","tool_calls":[{"name":"echo","arguments":"[]"},{"name":"shell","arguments":"{}"}]}`
	case "mixed-duplicate":
		return `{"assistant_text":"","tool_calls":[{"name":"echo","arguments":"[]"},{"name":"echo","arguments":"{\"a\":1,\"a\":2}"}]}`
	case "invalid":
		return bad
	}
	if !feedback {
		return bad
	}
	if strings.HasPrefix(mode, "repair-") && mode != "repair-malformed" {
		return "fake-" + strings.TrimPrefix(mode, "repair-")
	}
	if mode == "repair-malformed" {
		return `{"assistant_text":"broken"`
	}
	return `{"assistant_text":"hello","tool_calls":[{"name":"echo","arguments":"{\"value\":\"hello\"}"}]}`
}

func argumentRepairProvider(t *testing.T, mode, base string) (*provider, string) {
	t.Helper()
	_ = fakeProvider(t, "normal")
	sessions := filepath.Join(t.TempDir(), "sessions.txt")
	t.Setenv("SCOPEWEAVER_CODEX_FAKE_ARGUMENT_MODE", mode)
	t.Setenv("SCOPEWEAVER_CODEX_FAKE_ARGUMENT_SESSIONS", sessions)
	options := Options{Model: "normal", BaseURL: base}
	if base == ZAIBaseURL {
		options.Model = "glm-5.3"
		options.APIKey = "fixture-key"
	}
	p, err := New(options)
	if err != nil {
		t.Fatal(err)
	}
	return p.(*provider), sessions
}
func requireArgumentSessions(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil || string(data) != want {
		t.Fatalf("sessions %q, error %v; want %q", data, err, want)
	}
}

var oneArgumentUsage = llm.Usage{InputTokens: 100, OutputTokens: 10, CacheReadTokens: 20, CacheWriteTokens: 3}
var twoArgumentUsage = llm.Usage{InputTokens: 200, OutputTokens: 20, CacheReadTokens: 40, CacheWriteTokens: 6}

func TestGLMArgumentRepairSubprocess(t *testing.T) {
	p, sessions := argumentRepairProvider(t, "repair", ZAIBaseURL)
	message, stop, usage, err := p.Complete(context.Background(), toolRequest())
	if err != nil {
		t.Fatal(err)
	}
	if message.Text() != "hello" || stop != "tool_use" || len(message.ToolUses()) != 1 || string(message.ToolUses()[0].Input) != `{"value":"hello"}` {
		t.Fatalf("unexpected repaired completion %#v %q", message, stop)
	}
	if usage != twoArgumentUsage {
		t.Fatalf("usage %#v", usage)
	}
	requireArgumentSessions(t, sessions, "initial\nrepair\n")
}

func TestGLMArgumentRepairBufferedStream(t *testing.T) {
	for _, mode := range []string{"repair", "invalid", "repair-malformed"} {
		t.Run(mode, func(t *testing.T) {
			p, sessions := argumentRepairProvider(t, mode, ZAIBaseURL)
			var events []llm.StreamEvent
			var failure error
			for event, err := range p.Stream(context.Background(), toolRequest()) {
				if err != nil {
					failure = err
					continue
				}
				events = append(events, event)
			}
			if len(events) == 0 || events[0].Type != llm.SEMessageStart || events[0].Usage != twoArgumentUsage {
				t.Fatalf("events %#v", events)
			}
			for _, event := range events[1:] {
				if event.Usage != (llm.Usage{}) {
					t.Fatal("usage charged twice")
				}
			}
			if mode == "repair" {
				if failure != nil || len(events) != 6 || events[1].Text != "hello" {
					t.Fatalf("events %#v, failure %v", events, failure)
				}
			} else if failure == nil || len(events) != 1 {
				t.Fatalf("partial output or missing failure: %#v %v", events, failure)
			}
			requireArgumentSessions(t, sessions, "initial\nrepair\n")
		})
	}
}

func TestArgumentRepairOnlyForGLMFormatError(t *testing.T) {
	for _, tc := range []struct {
		mode, base string
		success    bool
	}{
		{"valid", ZAIBaseURL, true}, {"repair", "", false}, {"malformed", ZAIBaseURL, false}, {"unknown", ZAIBaseURL, false}, {"duplicate", ZAIBaseURL, false}, {"mixed-unknown", ZAIBaseURL, false}, {"mixed-duplicate", ZAIBaseURL, false}, {"transport", ZAIBaseURL, false}, {"native", ZAIBaseURL, false}, {"quota", ZAIBaseURL, false},
	} {
		t.Run(tc.mode+tc.base, func(t *testing.T) {
			p, sessions := argumentRepairProvider(t, tc.mode, tc.base)
			_, _, usage, err := p.Complete(context.Background(), toolRequest())
			wantUsage := oneArgumentUsage
			if tc.mode == "transport" || tc.mode == "native" {
				wantUsage = llm.Usage{}
			}
			if (err == nil) != tc.success || usage != wantUsage {
				t.Fatalf("usage %#v, error %v", usage, err)
			}
			requireArgumentSessions(t, sessions, "initial\n")
		})
	}
}

func TestGLMArgumentRepairRechecksPolicyAndCancellation(t *testing.T) {
	for _, mode := range []string{"policy", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			p, sessions := argumentRepairProvider(t, "repair", ZAIBaseURL)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			checks := 0
			p.options.PolicyResolver = func() (bool, []string) {
				checks++
				if checks == 2 && mode == "cancel" {
					cancel()
				}
				return checks == 2 && mode == "policy", nil
			}
			_, _, usage, err := p.Complete(ctx, toolRequest())
			if err == nil || usage != oneArgumentUsage || checks != 2 {
				t.Fatalf("checks %d, usage %#v, error %v", checks, usage, err)
			}
			if mode == "cancel" && !errors.Is(err, context.Canceled) {
				t.Fatalf("cancellation %v", err)
			}
			if mode == "policy" && !strings.Contains(err.Error(), "local-LLM-only") {
				t.Fatalf("policy %v", err)
			}
			requireArgumentSessions(t, sessions, "initial\n")
		})
	}
}

func TestGLMArgumentRepairFailureRetainsPaidUsage(t *testing.T) {
	for _, mode := range []string{"transport", "quota", "hang", "native"} {
		t.Run(mode, func(t *testing.T) {
			p, sessions := argumentRepairProvider(t, "repair-"+mode, ZAIBaseURL)
			ctx := context.Background()
			if mode == "hang" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
				defer cancel()
			}
			_, _, usage, err := p.Complete(ctx, toolRequest())
			wantUsage := twoArgumentUsage
			if mode == "transport" || mode == "native" {
				wantUsage = oneArgumentUsage
			}
			if err == nil || usage != wantUsage {
				t.Fatalf("usage %#v, error %v", usage, err)
			}
			if mode == "hang" && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("deadline %v", err)
			}
			requireArgumentSessions(t, sessions, "initial\nrepair\n")
		})
	}
}

func TestArgumentRepairContextBoundAndSafeDiagnostic(t *testing.T) {
	prefix := strings.Repeat("x", maxContextBytes-len(argumentRepairFeedback))
	prompt, err := argumentRepairContext(prefix)
	if err != nil || len(prompt) != maxContextBytes {
		t.Fatalf("bounded feedback length %d, error %v", len(prompt), err)
	}
	if _, err := argumentRepairContext(prefix + "x"); err == nil {
		t.Fatal("oversize correction accepted")
	}
	var diagnostic bytes.Buffer
	oldWriter := log.Writer()
	log.SetOutput(&diagnostic)
	defer log.SetOutput(oldWriter)
	p, _ := argumentRepairProvider(t, "repair", ZAIBaseURL)
	req := toolRequest()
	req.Messages = []llm.Message{llm.UserText("fixture-secret-prompt")}
	if _, _, _, err := p.Complete(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if strings.Count(diagnostic.String(), argumentRepairDiagnostic) != 1 || strings.Contains(diagnostic.String(), "fixture-secret") || strings.Contains(diagnostic.String(), "must not leak") {
		t.Fatalf("unsafe diagnostic %q", diagnostic.String())
	}
}

func TestGLMArgumentRepairEarlyConsumerUsage(t *testing.T) {
	p, sessions := argumentRepairProvider(t, "repair", ZAIBaseURL)
	events := 0
	for event, err := range p.Stream(context.Background(), toolRequest()) {
		if err != nil || event.Type != llm.SEMessageStart || event.Usage != twoArgumentUsage {
			t.Fatalf("event %#v, error %v", event, err)
		}
		events++
		break
	}
	if events != 1 {
		t.Fatal("missing paid usage")
	}
	requireArgumentSessions(t, sessions, "initial\nrepair\n")
}
