package codexprovider

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Autumn-27/norma/llm"
)

// The test binary is an app-server subprocess, not an in-process mock. These
// checks cover JSON-RPC ordering, process cancellation, and concurrent requests.
func TestMain(m *testing.M) {
	if os.Getenv("SCOPEWEAVER_CODEX_FAKE") == "1" {
		if os.Getenv("SCOPEWEAVER_CODEX_FAKE_STDIN_STALL") == "1" {
			time.Sleep(time.Minute)
			os.Exit(0)
		}
		fakeAppServer()
		os.Exit(0)
	}
	os.Exit(m.Run())
}
func fakeAppServer() {
	if expected := os.Getenv("SCOPEWEAVER_CODEX_FAKE_EXPECT_EFFORT"); expected != "" {
		encoded, _ := json.Marshal(expected)
		found := false
		for _, arg := range os.Args[1:] {
			found = found || arg == "model_reasoning_effort="+string(encoded)
		}
		if !found {
			os.Exit(28)
		}
	}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 64<<10), maxProtocolLine)
	send := func(value any) { data, _ := json.Marshal(value); fmt.Println(string(data)) }
	model, thread, turn := "", "fixture-thread", "fixture-turn"
	for scanner.Scan() {
		var req struct {
			ID     json.RawMessage
			Method string
			Params map[string]json.RawMessage
		}
		if json.Unmarshal(scanner.Bytes(), &req) != nil {
			os.Exit(21)
		}
		response := func(value any) { send(map[string]any{"id": req.ID, "result": value}) }
		event := func(method string, params any) { send(map[string]any{"method": method, "params": params}) }
		switch req.Method {
		case "initialize":
			var capabilities map[string]bool
			_ = json.Unmarshal(req.Params["capabilities"], &capabilities)
			if !capabilities["experimentalApi"] {
				os.Exit(22)
			}
			response(map[string]any{"serverInfo": map[string]any{"name": "fake"}})
		case "initialized":
		case "account/read":
			kind := os.Getenv("SCOPEWEAVER_CODEX_FAKE_ACCOUNT_TYPE")
			if kind == "" {
				kind = "chatgpt"
			}
			response(map[string]any{"account": map[string]any{"type": kind}})
		case "config/read":
			response(map[string]any{"config": map[string]any{"mcp_servers": map[string]any{fakeMCPName(): map[string]any{"command": "never-execute"}}}})
		case "thread/start":
			_ = json.Unmarshal(req.Params["model"], &model)
			var config map[string]any
			_ = json.Unmarshal(req.Params["config"], &config)
			if string(req.Params["ephemeral"]) != "true" || string(req.Params["environments"]) != "[]" || config["mcp_servers.fixture_mcp.enabled"] != false {
				os.Exit(23)
			}
			var base string
			_ = json.Unmarshal(req.Params["baseInstructions"], &base)
			var developer string
			_ = json.Unmarshal(req.Params["developerInstructions"], &developer)
			if !strings.Contains(developer, bridgeInstructions) || (model != "text" && !strings.Contains(developer, "fixture-system")) {
				os.Exit(27)
			}
			if base != bridgeInstructions {
				os.Exit(24)
			}
			response(map[string]any{"thread": map[string]any{"id": thread}})
		case "mcpServerStatus/list":
			status := "disabled"
			tools := map[string]any{}
			if model == "active-mcp" {
				status = "ready"
				tools["execute"] = map[string]any{}
			}
			if len(req.Params["cursor"]) == 0 {
				response(map[string]any{"data": []any{map[string]any{"runtimeStatus": status, "tools": tools}}, "nextCursor": "page2"})
			} else {
				response(map[string]any{"data": []any{}, "nextCursor": nil})
			}
		case "turn/start":
			if string(req.Params["environments"]) != "[]" || len(req.Params["outputSchema"]) == 0 || len(req.Params["effort"]) != 0 {
				os.Exit(25)
			}
			response(map[string]any{"turn": map[string]any{"id": turn}})
			if model == "retry" {
				event("error", map[string]any{"threadId": thread, "turnId": turn, "willRetry": true, "error": map[string]any{"message": "fixture-secret-diagnostic", "codexErrorInfo": "rateLimitExceeded"}})
			}
			if model == "quota" || model == "rate" {
				code := "usageLimitExceeded"
				if model == "rate" {
					code = "rateLimitExceeded"
				}
				event("error", map[string]any{"threadId": thread, "turnId": turn, "willRetry": false, "error": map[string]any{"message": "fixture-secret-diagnostic", "codexErrorInfo": code}})
				time.Sleep(time.Minute)
				return
			}
			if model == "hang" {
				time.Sleep(time.Minute)
				return
			}
			if model == "approval" {
				send(map[string]any{"id": 99, "method": "item/commandExecution/requestApproval", "params": map[string]any{}})
				time.Sleep(time.Minute)
				return
			}
			if model == "builtin" {
				event("item/started", map[string]any{"threadId": thread, "turnId": turn, "item": map[string]any{"type": "commandExecution", "id": "cmd"}})
				time.Sleep(time.Minute)
				return
			}
			text := `{"assistant_text":"hello","tool_calls":[{"name":"echo","arguments":"{\"value\":\"hello\"}"}]}`
			switch model {
			case "glm-5.3", "glm-5.3-flash":
				home := os.Getenv("CODEX_HOME")
				catalog, err := os.ReadFile(home + string(os.PathSeparator) + "models.json")
				// A system temp alias can differ from the physical working-directory path.
				homeInfo, homeErr := os.Stat(home)
				cwdInfo, cwdErr := os.Stat(".")
				if err != nil || string(catalog) != zaiModelCatalog || homeErr != nil || cwdErr != nil || !os.SameFile(homeInfo, cwdInfo) || os.Getenv("SCOPEWEAVER_CODEX_ZAI_API_KEY") == "" {
					os.Exit(28)
				}
				encoded, _ := json.Marshal(map[string]any{"assistant_text": home, "tool_calls": []any{}})
				text = string(encoded)
			case "text":
				text = `{"assistant_text":"hello","tool_calls":[]}`
			case "unknown":
				text = `{"assistant_text":"","tool_calls":[{"name":"shell","arguments":"{}"}]}`
			case "malformed":
				text = `{"assistant_text":"broken"`
			case "oversize":
				text = strings.Repeat("x", maxOutputBytes+1)
			case "context":
				var input []struct{ Text string }
				_ = json.Unmarshal(req.Params["input"], &input)
				data, _ := json.Marshal(map[string]any{"assistant_text": input[0].Text, "tool_calls": []any{}})
				text = string(data)
			}
			item := map[string]any{"type": "agentMessage", "id": "answer", "phase": "final_answer", "text": text}
			event("item/started", map[string]any{"threadId": thread, "turnId": turn, "item": map[string]any{"type": "agentMessage", "id": "answer", "text": ""}})
			event("item/completed", map[string]any{"threadId": thread, "turnId": turn, "item": item})
			for i := 0; i < 2; i++ {
				event("thread/tokenUsage/updated", map[string]any{"threadId": thread, "turnId": turn, "tokenUsage": map[string]any{"total": map[string]any{"inputTokens": 100, "outputTokens": 10, "cachedInputTokens": 20, "cacheWriteInputTokens": 3}}})
			}
			status := "completed"
			if model == "failed" {
				status = "failed"
			}
			event("turn/completed", map[string]any{"threadId": thread, "turn": map[string]any{"id": turn, "status": status, "items": []any{item}}})
		default:
			os.Exit(26)
		}
	}
}
func fakeProvider(t *testing.T, model string) llm.Provider {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("SCOPEWEAVER_CODEX_PATH", exe)
	t.Setenv("SCOPEWEAVER_CODEX_FAKE", "1")
	p, err := New(Options{Model: model})
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func toolRequest() llm.CompletionRequest {
	return llm.CompletionRequest{System: []string{"fixture-system"}, Messages: []llm.Message{llm.UserText("hello")}, Tools: []llm.ToolSchema{{Name: "echo", InputSchema: map[string]any{"type": "object"}}}}
}
func TestCompleteAndBufferedStream(t *testing.T) {
	p := fakeProvider(t, "normal")
	message, stop, usage, err := p.Complete(context.Background(), toolRequest())
	if err != nil {
		t.Fatal(err)
	}
	if message.Text() != "hello" || stop != "tool_use" || len(message.ToolUses()) != 1 || string(message.ToolUses()[0].Input) != `{"value":"hello"}` {
		t.Fatalf("bad completion: %#v %s", message, stop)
	}
	want := llm.Usage{InputTokens: 100, OutputTokens: 10, CacheReadTokens: 20, CacheWriteTokens: 3}
	if usage != want {
		t.Fatalf("usage %#v", usage)
	}
	accumulator := llm.NewAccumulator()
	var events []llm.StreamEvent
	for event, err := range p.Stream(context.Background(), toolRequest()) {
		if err != nil {
			t.Fatal(err)
		}
		events = append(events, event)
		accumulator.Add(event)
	}
	if accumulator.Usage != want || accumulator.StopReason != "tool_use" || accumulator.Message().Text() != "hello" || len(accumulator.Message().ToolUses()) != 1 {
		t.Fatalf("bad buffered stream %#v", accumulator)
	}
	if events[0].Usage != want || events[len(events)-2].Usage != (llm.Usage{}) {
		t.Fatal("usage must appear exactly once on message_start")
	}
	if message.ToolUses()[0].ID == accumulator.Message().ToolUses()[0].ID {
		t.Fatal("tool IDs must be unique across calls")
	}
}
func TestStatelessContextAndConcurrency(t *testing.T) {
	p := fakeProvider(t, "context")
	req := toolRequest()
	req.Messages = append(req.Messages, llm.Message{Role: llm.RoleAssistant, Content: []llm.ContentBlock{{Type: llm.BlockToolUse, ID: "old", Name: "old-tool", Input: json.RawMessage(`{}`)}}}, llm.Message{Role: llm.RoleUser, Content: []llm.ContentBlock{llm.ToolResultText("old", "result", false)}})
	req.Tools = []llm.ToolSchema{{Name: "deferred-now", InputSchema: map[string]any{"type": "object"}}}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			message, stop, _, err := p.Complete(context.Background(), req)
			if err != nil {
				t.Error(err)
				return
			}
			if stop != "end_turn" || !strings.Contains(message.Text(), `"name":"deferred-now"`) || !strings.Contains(message.Text(), `"tool_use_id":"old"`) || !strings.Contains(message.Text(), `"system_segments":["fixture-system"]`) {
				t.Error("current tools and translated history were not supplied")
			}
		}()
	}
	wg.Wait()
}
func TestSubprocessFailuresAndCancellation(t *testing.T) {
	for _, model := range []string{"unknown", "malformed", "oversize", "approval", "builtin", "failed", "active-mcp"} {
		t.Run(model, func(t *testing.T) {
			p := fakeProvider(t, model)
			_, _, _, err := p.Complete(context.Background(), toolRequest())
			if err == nil {
				t.Fatal("unsafe or malformed response accepted")
			}
		})
	}
	t.Run("cancel", func(t *testing.T) {
		p := fakeProvider(t, "hang")
		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer cancel()
		start := time.Now()
		_, _, _, err := p.Complete(ctx, toolRequest())
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected deadline, got %v", err)
		}
		if time.Since(start) > 4*time.Second {
			t.Fatal("cancellation did not clean up promptly")
		}
	})
	t.Run("early-stream-consumer", func(t *testing.T) {
		p := fakeProvider(t, "text")
		count := 0
		for _, err := range p.Stream(context.Background(), llm.CompletionRequest{Messages: []llm.Message{llm.UserText("hi")}}) {
			if err != nil {
				t.Fatal(err)
			}
			count++
			break
		}
		if count != 1 {
			t.Fatal("no event")
		}
	})
}
func TestEnvelopeValidation(t *testing.T) {
	tools := toolRequest().Tools
	for _, text := range []string{
		`{}`, `null`, `{"assistant_text":"a","assistant_text":"b","tool_calls":[]}`, `{"assistant_text":"","tool_calls":[{"name":"echo","arguments":"{\"value\":1,\"value\":2}"}]}`, `{"assistant_text":"","tool_calls":null}`, `{"assistant_text":"","tool_calls":[],"unknown":true}`,
		`{"assistant_text":"","tool_calls":[]} {}`,
		`{"assistant_text":"","tool_calls":[{"name":"echo","arguments":"[]"}]}`,
		`{"assistant_text":"","tool_calls":[{"name":"echo","arguments":"null"}]}`,
		`{"assistant_text":"","tool_calls":[{"name":"echo","arguments":"{broken}"}]}`,
		`{"assistant_text":"","tool_calls":[{"name":"echo","arguments":"{} {}"}]}`,
		`{"assistant_text":"","tool_calls":[{"name":"echo"}]}`,
	} {
		if _, _, err := decodeEnvelope(text, tools); err == nil {
			t.Errorf("accepted %s", text)
		}
	}
	if _, _, err := decodeEnvelope(`{"assistant_text":"","tool_calls":[{"name":"echo","arguments":"{}"}]}`, nil); err == nil {
		t.Fatal("tool called during tools-free completion")
	}
}
func TestValidationAndProcessConfiguration(t *testing.T) {
	t.Setenv("ZAI_API_KEY", "")
	for _, options := range []Options{{}, {Model: "x", BaseURL: "https://example.com"}, {Model: "x", APIKey: "secret"}, {Model: "x", BaseURL: ZAIBaseURL}, {Model: "x", ReasoningEffort: "invented"}, {Model: "x", ReasoningEffort: "max"}} {
		if ValidateOptions(options) == nil {
			t.Errorf("accepted %#v", options)
		}
	}
	if NormalizeBaseURL(" https://api.z.ai/api/v1/responses/ ") != ZAIBaseURL {
		t.Fatal("endpoint normalization")
	}
	t.Setenv("ZAI_API_KEY", "fixture-zai-secret")
	t.Setenv("OPENAI_API_KEY", "fixture-openai-secret")
	t.Setenv("OPENAI_BASE_URL", "https://untrusted.invalid")
	options := Options{Model: "glm-5.3", BaseURL: ZAIBaseURL, ReasoningEffort: "high"}
	if err := ValidateOptions(options); err != nil {
		t.Fatal(err)
	}
	args, env, err := processConfig(options, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(args, " "), "fixture-") {
		t.Fatal("credential leaked into argv")
	}
	joined := strings.Join(env, "\n")
	if strings.Contains(joined, "fixture-openai-secret") || strings.Contains(joined, "https://untrusted.invalid") || !strings.Contains(joined, "SCOPEWEAVER_CODEX_ZAI_API_KEY=fixture-zai-secret") {
		t.Fatal("incorrect child credential isolation")
	}
	for _, required := range []string{"features.apps=false", "features.hooks=false", "features.plugins=false", "features.shell_tool=false", "features.multi_agent=false", "features.skills=false", "web_search=disabled", "tools.view_image=false", "project_doc_max_bytes=0", "notify=[]", "analytics.enabled=false", `otel.exporter="none"`, `otel.trace_exporter="none"`, `otel.metrics_exporter="none"`, "otel.log_user_prompt=false", "otel.log_agent_responses=false", "otel.log_guardian_assessments=false"} {
		found := false
		for _, arg := range args {
			if arg == required {
				found = true
			}
		}
		if !found {
			t.Errorf("missing isolation setting %s", required)
		}
	}
	p, err := New(Options{Model: "fixture", LocalLLMOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, err = p.Complete(context.Background(), toolRequest())
	if err == nil || !strings.Contains(err.Error(), "local-LLM-only") {
		t.Fatal("local-only policy not checked")
	}
	dynamic := false
	p, err = New(Options{Model: "fixture", PolicyResolver: func() (bool, []string) { return dynamic, nil }})
	if err != nil {
		t.Fatal(err)
	}
	dynamic = true
	_, _, _, err = p.Complete(context.Background(), toolRequest())
	if err == nil || !strings.Contains(err.Error(), "local-LLM-only") {
		t.Fatal("dynamic policy not checked")
	}
	req := toolRequest()
	req.Tools = append(req.Tools, req.Tools[0])
	if _, err := completionContext(req); err == nil {
		t.Fatal("duplicate tools accepted")
	}

}

func TestEarlyConsumerRetainsFullUsage(t *testing.T) {
	p := fakeProvider(t, "normal")
	accumulator := llm.NewAccumulator()
	for event, err := range p.Stream(context.Background(), toolRequest()) {
		if err != nil {
			t.Fatal(err)
		}
		accumulator.Add(event)
		break
	}
	if accumulator.Usage != (llm.Usage{InputTokens: 100, OutputTokens: 10, CacheReadTokens: 20, CacheWriteTokens: 3}) {
		t.Fatal("early consumer lost completed model usage")
	}
}
func TestUnsupportedImageInputFailsBeforeProcess(t *testing.T) {
	t.Setenv("SCOPEWEAVER_CODEX_PATH", "does-not-exist")
	p, err := New(Options{Model: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	for _, blocks := range [][]llm.ContentBlock{
		{{Type: "image", Text: "image-data"}},
		{{Type: llm.BlockToolResult, Content: []llm.ContentBlock{{Type: "image"}}}},
	} {
		_, _, _, err := p.Complete(context.Background(), llm.CompletionRequest{Messages: []llm.Message{{Role: llm.RoleUser, Content: blocks}}})
		if err == nil || !strings.Contains(err.Error(), "unsupported content block") {
			t.Fatalf("image silently degraded or process started: %v", err)
		}
	}
}

func TestSubscriptionLoginRejectsAPIKeyAccount(t *testing.T) {
	p := fakeProvider(t, "normal")
	t.Setenv("SCOPEWEAVER_CODEX_FAKE_ACCOUNT_TYPE", "apiKey")
	_, _, _, err := p.Complete(context.Background(), toolRequest())
	if err == nil || !strings.Contains(err.Error(), "ChatGPT subscription login required") {
		t.Fatalf("API key login used for subscription profile: %v", err)
	}
}
func TestBufferedStreamInvalidOutputRetainsUsage(t *testing.T) {
	p := fakeProvider(t, "unknown")
	usage := llm.Usage{}
	errorsSeen := 0
	for event, err := range p.Stream(context.Background(), toolRequest()) {
		if event.Type == llm.SEMessageStart {
			usage.Add(event.Usage)
		}
		if err != nil {
			errorsSeen++
		}
		if event.Type == llm.SEToolUseStart {
			t.Fatal("invalid tool escaped validation")
		}
	}
	if usage.InputTokens != 100 || usage.OutputTokens != 10 || errorsSeen != 1 {
		t.Fatalf("usage or terminal error lost: %#v errors=%d", usage, errorsSeen)
	}
}

func TestCancellationDuringBlockedInput(t *testing.T) {
	_ = fakeProvider(t, "normal")
	t.Setenv("SCOPEWEAVER_CODEX_FAKE_STDIN_STALL", "1")
	executable, err := resolveExecutable()
	if err != nil {
		t.Fatal(err)
	}
	args, env, err := processConfig(Options{Model: "normal"}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s, err := startSession(executable, args, env, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.close()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err = s.request(ctx, "oversized-input-fixture", map[string]any{"text": strings.Repeat("x", 2<<20)})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("blocked stdin did not cancel: %v", err)
	}
	if time.Since(start) > 3*time.Second {
		t.Fatal("blocked input cleanup exceeded bound")
	}
}

func fakeMCPName() string {
	if name := os.Getenv("SCOPEWEAVER_CODEX_FAKE_MCP_NAME"); name != "" {
		return name
	}
	return "fixture_mcp"
}
func TestUnsupportedMCPNameFailsBeforeThreadCreation(t *testing.T) {
	p := fakeProvider(t, "normal")
	for _, name := range []string{"fixture.mcp", `fixture"mcp`, "fixture mcp", "한글"} {
		t.Setenv("SCOPEWEAVER_CODEX_FAKE_MCP_NAME", name)
		_, _, _, err := p.Complete(context.Background(), toolRequest())
		if err == nil || !strings.Contains(err.Error(), "cannot isolate MCP server") {
			t.Fatalf("unsupported server name did not fail before thread creation: %v", err)
		}
	}
}

func TestProviderSpecificModelAndReasoningValidation(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.3-flash"} {
		for _, effort := range []string{"", "low", "high", "max"} {
			if err := ValidateOptions(Options{Model: model, BaseURL: ZAIBaseURL, APIKey: "fixture", ReasoningEffort: effort}); err != nil {
				t.Errorf("valid %s effort %s rejected: %v", model, effort, err)
			}
		}
		for _, effort := range []string{"medium", "xhigh", "none"} {
			if ValidateOptions(Options{Model: model, BaseURL: ZAIBaseURL, APIKey: "fixture", ReasoningEffort: effort}) == nil {
				t.Errorf("unsupported %s effort %s accepted", model, effort)
			}
		}
	}
	for _, model := range []string{"glm-5", "glm-5.3-flashx", "arbitrary"} {
		if ValidateOptions(Options{Model: model, BaseURL: ZAIBaseURL, APIKey: "fixture"}) == nil {
			t.Errorf("unsupported GLM model %s accepted", model)
		}
	}
}
func TestOAuthEndpointAndAuthenticationOverridesAreIsolated(t *testing.T) {
	for _, name := range []string{"OPENAI_BASE_URL", "OPENAI_API_KEY", "OPENAI_UNRECOGNIZED_OVERRIDE", "CODEX_REFRESH_TOKEN_URL_OVERRIDE", "CODEX_REVOKE_TOKEN_URL_OVERRIDE", "CODEX_APP_SERVER_LOGIN_CLIENT_ID", "CODEX_ACCESS_TOKEN", "CODEX_UNRECOGNIZED_OVERRIDE"} {
		t.Setenv(name, "fixture-danger")
	}
	t.Setenv("CODEX_HOME", t.TempDir())
	args, env, err := processConfig(Options{Model: "fixture", ReasoningEffort: "high"}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, "\n")
	for _, value := range []string{`openai_base_url="https://chatgpt.com/backend-api/codex"`, `chatgpt_base_url="https://chatgpt.com/backend-api/"`, `model_reasoning_effort="high"`} {
		if !strings.Contains(joined, value) {
			t.Errorf("missing pinned setting %s", value)
		}
	}
	if strings.Contains(strings.Join(env, "\n"), "fixture-danger") {
		t.Fatal("authentication or endpoint override leaked into child")
	}
	if !strings.Contains(strings.Join(env, "\n"), "CODEX_HOME="+os.Getenv("CODEX_HOME")) {
		t.Fatal("intentional OAuth home was not preserved")
	}
}
func TestGLMCatalogAndHomeLifecycle(t *testing.T) {
	_ = fakeProvider(t, "normal")
	for _, model := range []string{"glm-5.3", "glm-5.3-flash"} {
		t.Run(model, func(t *testing.T) {
			p, err := New(Options{Model: model, BaseURL: ZAIBaseURL, APIKey: "fixture-zai", ReasoningEffort: "max"})
			if err != nil {
				t.Fatal(err)
			}
			message, _, _, err := p.Complete(context.Background(), toolRequest())
			if err != nil {
				t.Fatal(err)
			}
			if message.Text() == "" {
				t.Fatal("fake did not verify isolated catalog")
			}
			if _, err := os.Stat(message.Text()); !os.IsNotExist(err) {
				t.Fatalf("temporary Codex home was retained: %v", err)
			}
		})
	}
	var catalog struct {
		Models []struct {
			Slug             string   `json:"slug"`
			ContextWindow    int      `json:"context_window"`
			MaxContextWindow int      `json:"max_context_window"`
			DefaultReasoning string   `json:"default_reasoning_level"`
			InputModalities  []string `json:"input_modalities"`
			Reasoning        []struct {
				Effort string `json:"effort"`
			} `json:"supported_reasoning_levels"`
		} `json:"models"`
	}
	if err := json.Unmarshal([]byte(zaiModelCatalog), &catalog); err != nil {
		t.Fatal(err)
	}
	if len(catalog.Models) != 2 {
		t.Fatalf("expected two declared GLM models, got %d", len(catalog.Models))
	}
	for i, slug := range []string{"glm-5.3", "glm-5.3-flash"} {
		model := catalog.Models[i]
		if model.Slug != slug || model.ContextWindow != 1048576 || model.MaxContextWindow != 1048576 || model.DefaultReasoning != "max" || strings.Join(model.InputModalities, ",") != "text" || len(model.Reasoning) != 3 {
			t.Fatalf("official %s metadata was not preserved", slug)
		}
		for j, effort := range []string{"low", "high", "max"} {
			if model.Reasoning[j].Effort != effort {
				t.Errorf("%s reasoning level %d: got %s, want %s", slug, j, model.Reasoning[j].Effort, effort)
			}
		}
	}
}
func TestRecoverableRetryAndTerminalQuotaClassification(t *testing.T) {
	t.Run("retry-success", func(t *testing.T) {
		p := fakeProvider(t, "retry")
		message, stop, _, err := p.Complete(context.Background(), toolRequest())
		if err != nil || message.Text() != "hello" || stop != "tool_use" {
			t.Fatalf("recoverable error interrupted model retry: %v", err)
		}
	})
	for _, model := range []string{"quota", "rate"} {
		t.Run(model, func(t *testing.T) {
			p := fakeProvider(t, model)
			_, _, _, err := p.Complete(context.Background(), toolRequest())
			if err == nil {
				t.Fatal("terminal error accepted")
			}
			if strings.Contains(err.Error(), "fixture-secret-diagnostic") {
				t.Fatal("raw provider message escaped")
			}
			if strings.Contains(err.Error(), "quota exhausted") != (model == "quota") {
				t.Fatalf("quota/rate classification confused: %v", err)
			}
		})
	}
	s := &session{threadID: "t", turnID: "u"}
	_, err := s.notification(wireMessage{Method: "turn/completed", Params: json.RawMessage(`{"threadId":"t","turn":{"id":"u","status":"failed","items":[],"error":{"message":"private fixture","codexErrorInfo":"usageLimitExceeded"}}}`)})
	if err == nil || !strings.Contains(err.Error(), "quota exhausted") || strings.Contains(err.Error(), "private") {
		t.Fatalf("failed turn lost safe quota classification: %v", err)
	}
}

func TestActionSchemaRequiresCompleteObjectArguments(t *testing.T) {
	tools := []llm.ToolSchema{{Name: "add_intent", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"intents": map[string]any{"type": "array"}}, "required": []string{"intents"}}}}
	// The actual GLM planner diagnostic returned a two-byte JSON array string for
	// add_intent. Its schema requires an object containing the intents property.
	bad := `{"assistant_text":"","tool_calls":[{"name":"add_intent","arguments":"[]"}]}`
	if _, _, err := decodeEnvelope(bad, tools); err == nil {
		t.Fatal("observed nonobject arguments were accepted")
	}
	schema := outputSchema(tools)
	arguments := schema["properties"].(map[string]any)["tool_calls"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any)["arguments"].(map[string]any)
	if description, _ := arguments["description"].(string); !strings.Contains(description, "COMPLETE INPUT OBJECT") {
		t.Fatal("arguments have no complete-object guidance")
	}
	for _, value := range []string{`{}`, `{"intents":[{"summary":"fixture"}]}`, " \n{\n\"items\": []\n}\n"} {
		raw, _ := json.Marshal(map[string]any{"assistant_text": "", "tool_calls": []any{map[string]any{"name": "add_intent", "arguments": value}}})
		if _, _, err := decodeEnvelope(string(raw), tools); err != nil {
			t.Errorf("valid object rejected: %q: %v", value, err)
		}
	}
	for _, value := range []string{`[]`, `null`, `"{}"`, "", `true`, `42`} {
		raw, _ := json.Marshal(map[string]any{"assistant_text": "", "tool_calls": []any{map[string]any{"name": "add_intent", "arguments": value}}})
		if _, _, err := decodeEnvelope(string(raw), tools); err == nil {
			t.Errorf("nonobject arguments accepted: %q", value)
		}
	}
	if !strings.Contains(bridgeInstructions, "COMPLETE INPUT OBJECT") || !strings.Contains(bridgeInstructions, "If no valid tool invocation is needed, return tool_calls as an empty array") {
		t.Fatal("bridge omitted object and no-action guidance")
	}
	// Braces alone do not make valid JSON. The production parser remains strict.
	if _, _, err := decodeEnvelope(`{"assistant_text":"","tool_calls":[{"name":"add_intent","arguments":"{broken}"}]}`, tools); err == nil {
		t.Fatal("schema guidance weakened malformed JSON validation")
	}
}

func TestDisabledThinkingPreservesSupportedOAuthProfileEffort(t *testing.T) {
	fakeProvider(t, "normal")
	t.Setenv("SCOPEWEAVER_CODEX_FAKE_EXPECT_EFFORT", "high")
	p, err := New(Options{Model: "normal", ReasoningEffort: "high"})
	if err != nil {
		t.Fatal(err)
	}
	req := toolRequest()
	req.Thinking = "disabled"
	if _, _, _, err = p.Complete(t.Context(), req); err != nil {
		t.Fatal(err)
	}
}

func TestDisabledThinkingUsesLowForGLM(t *testing.T) {
	fakeProvider(t, "glm-5.3")
	t.Setenv("SCOPEWEAVER_CODEX_FAKE_EXPECT_EFFORT", "low")
	p, err := New(Options{Model: "glm-5.3", BaseURL: ZAIBaseURL, APIKey: "fixture", ReasoningEffort: "max"})
	if err != nil {
		t.Fatal(err)
	}
	req := toolRequest()
	req.Thinking = "disabled"
	if _, _, _, err = p.Complete(t.Context(), req); err != nil {
		t.Fatal(err)
	}
}
