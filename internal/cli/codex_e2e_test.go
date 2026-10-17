package cli

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cskwork/scopeweaver-cli/db"
)

// Each subscription route requires its own explicit opt-in. Auth files are never
// read by these tests; every target and mutable path belongs to its fixture.
func TestPortableCLICodexEndToEnd(t *testing.T) {
	runPortableCLICodexEndToEnd(t, "SCOPEWEAVER_CLI_CODEX_E2E_BINARY", "ChatGPT OAuth", "gpt-6.1-sol", "")
}

func TestPortableCLIGLMEndToEnd(t *testing.T) {
	runPortableCLICodexEndToEnd(t, "SCOPEWEAVER_CLI_GLM_E2E_BINARY", "Z.AI GLM Coding Plan", "glm-5.3", "https://api.z.ai/api/v1")
}

func TestPortableCLIGLMFlashEndToEnd(t *testing.T) {
	runPortableCLICodexEndToEnd(t, "SCOPEWEAVER_CLI_GLM_FLASH_E2E_BINARY", "Z.AI GLM Flash Coding Plan", "glm-5.3-flash", "https://api.z.ai/api/v1")
}

func runPortableCLICodexEndToEnd(t *testing.T, binaryEnv, providerName, model, baseURL string) {
	t.Helper()
	binary := os.Getenv(binaryEnv)
	if binary == "" {
		t.Skipf("set %s to opt into real %s inference with a packaged executable", binaryEnv, providerName)
	}
	binary, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(binary); err != nil || !info.Mode().IsRegular() {
		t.Fatalf("packaged executable unavailable: %v", err)
	}
	home, cwd := filepath.Join(t.TempDir(), "codex portable state 한글"), t.TempDir()
	env := codexE2EEnvironment(os.Environ())
	prerequisites := "official Codex executable and ChatGPT login"
	if baseURL != "" {
		key := os.Getenv("ZAI_API_KEY")
		if strings.TrimSpace(key) == "" {
			t.Fatal("explicit GLM E2E opt-in requires ZAI_API_KEY in the process environment")
		}
		// Only the opted-in GLM child gets this secret; never put it in JSON or argv.
		env = append(env, "ZAI_API_KEY="+key)
		prerequisites = "official Codex executable and ZAI_API_KEY credentials/quota"
	}
	if path := os.Getenv("SCOPEWEAVER_CODEX_PATH"); path != "" {
		env = append(env, "SCOPEWEAVER_CODEX_PATH="+path)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Minute)
	defer cancel()
	calls := 0
	execute := func(callCtx context.Context, args ...string) ([]byte, error) {
		cmd := exec.CommandContext(callCtx, binary, append([]string{"--home", home, "--json"}, args...)...)
		cmd.Dir, cmd.Env, cmd.WaitDelay = cwd, env, 5*time.Second
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("%v: %w; stdout=%s; stderr=%s", args, err, stdout.String(), stderr.String())
		}
		return stdout.Bytes(), nil
	}
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer stopCancel()
		if _, err := execute(stopCtx, "runtime", "stop", "--yes"); err != nil {
			t.Errorf("owned runtime cleanup: %v", err)
		}
		if baseURL != "" {
			// Count only the fixed safe diagnostic; never expose raw prompts or model output.
			log, err := os.ReadFile(filepath.Join(home, "private", "runtime.log"))
			if err != nil {
				t.Errorf("read owned GLM correction evidence: %v", err)
			} else {
				t.Logf("GLM argument-format correction attempts: %d", bytes.Count(log, []byte("codex: GLM argument-format correction started (attempt 2 of 2; prior response not executed)")))
			}
		}
	}()
	run := func(args ...string) map[string]any {
		t.Helper()
		calls++
		callCtx, stop := context.WithTimeout(ctx, 100*time.Second)
		defer stop()
		raw, err := execute(callCtx, args...)
		if err != nil {
			t.Fatal(err)
		}
		var value any
		if err := json.Unmarshal(raw, &value); err != nil {
			t.Fatalf("%v returned invalid JSON: %s", args, raw)
		}
		if object, ok := value.(map[string]any); ok {
			return object
		}
		return map[string]any{"data": value}
	}
	submit := func(body map[string]any, args ...string) map[string]any {
		t.Helper()
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(cwd, fmt.Sprintf("request-%d.json", calls))
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
		return run(append(args, "--data", "@"+path, "--yes")...)
	}
	pause := func(duration time.Duration) {
		t.Helper()
		select {
		case <-ctx.Done():
			t.Fatalf("%s E2E overall timeout: %v", providerName, ctx.Err())
		case <-time.After(duration):
		}
	}
	waitChat := func(id, previousRun string) map[string]any {
		t.Helper()
		deadline := time.Now().Add(3 * time.Minute)
		var state map[string]any
		for time.Now().Before(deadline) {
			state = run("conversations", "controls", "state", "--id", id)
			latest := codexString(state["latest_run_id"])
			if latest != "" && latest != previousRun {
				switch state["latest_run_status"] {
				case "completed":
					if state["active"] == nil {
						return state
					}
				case "failed", "cancelled", "interrupted", "stopped", "unknown":
					messages := run("conversations", "messages", "list", "--id", id)
					var failures []map[string]any
					for _, item := range codexRows(t, messages, "items") {
						if item["kind"] == "result" && item["is_error"] == true {
							failures = append(failures, run("conversations", "messages", "get", "--id", id, "--seq", codexString(item["seq"])))
						}
					}
					t.Fatalf("real %s chat did not complete: state=%v activities=%v terminal_details=%v; verify %s and SCOPEWEAVER_CODEX_PATH", providerName, state, messages, failures, prerequisites)
				}
			}
			pause(2 * time.Second)
		}
		t.Fatalf("real %s chat timeout: state=%v; verify %s and SCOPEWEAVER_CODEX_PATH", providerName, state, prerequisites)
		return nil
	}

	run("runtime", "start")
	if health := run("health"); health["ok"] != true {
		t.Fatalf("unhealthy packaged runtime: %v", health)
	}
	submit(map[string]any{"language": "en", "workers": 1, "llm_max_concurrent": 1, "web_search_enabled": false, "notify_enabled": false, "llm_record": true}, "settings", "set")
	profileData := map[string]any{"name": providerName + " installed E2E", "format": "codex-cli", "model": model, "base_url": baseURL, "reasoning_effort": "low", "streaming": true}
	if baseURL != "" {
		profileData["context_window_k"] = 1000
	}
	profile := submit(profileData, "llm", "profiles", "save")
	profileID := codexString(profile["id"])
	if profileID == "" {
		t.Fatalf("profile was not persisted: %v", profile)
	}
	run("llm", "profiles", "activate", "--id", profileID, "--yes")
	for _, key := range []string{"auto", "goals", "planner", "worker"} {
		submit(map[string]any{"max_turns": 8, "run_seconds": 180, "web_search": false, "interactive_shell": false}, "agents", "config", "set", "--key", key)
	}

	marker := "CODEX_READ_" + rand.Text()
	relativeFile := "codex-e2e-marker.txt"
	run("workspace", "write", "--path", relativeFile, "--content", marker, "--yes")
	conversation := run("conversations", "create", "--agent-key", "auto", "--title", providerName+" real tool proof", "--llm-profile-id", profileID, "--yes")
	conversationID := codexString(conversation["id"])
	if conversationID == "" {
		t.Fatalf("conversation missing: %v", conversation)
	}
	source := filepath.Join(home, "workspace", relativeFile)
	firstPrompt := fmt.Sprintf("This is an authorized isolated CLI integration test. Use the Norma Read tool exactly once to read this owned test file: %s. Do not use Bash, web, other tools, or any path outside that file. Then answer exactly MARKER=<the complete file contents>. Do not guess the contents. Do not create tasks or modify files.", source)
	admission := submit(map[string]any{"client_request_id": rand.Text(), "mode": "send", "message": firstPrompt}, "conversations", "controls", "inputs", "send", "--id", conversationID)
	if codexString(admission["id"]) == "" {
		t.Fatalf("chat admission missing ID: %v", admission)
	}
	firstState := waitChat(conversationID, "")
	firstRun := codexString(firstState["latest_run_id"])
	messages := run("conversations", "messages", "list", "--id", conversationID, "--limit", "100")
	readCalls := map[string]bool{}
	verifiedRead, verifiedAnswer := false, false
	for _, item := range codexRows(t, messages, "items") {
		if item["run_id"] != firstRun {
			continue
		}
		if item["kind"] == "tool_use" && item["tool"] == "Read" {
			readCalls[codexString(item["tool_use_id"])] = true
		}
		if (item["kind"] == "tool_result" && item["tool"] == "Read") || item["kind"] == "result" {
			detail := run("conversations", "messages", "get", "--id", conversationID, "--seq", codexString(item["seq"]))
			text := codexString(detail["detail"])
			if item["is_error"] != true && strings.Contains(text, marker) {
				if item["kind"] == "tool_result" && readCalls[codexString(item["tool_use_id"])] {
					verifiedRead = true
				}
				if item["kind"] == "result" && strings.Contains(text, "MARKER="+marker) {
					verifiedAnswer = true
				}
			}
		}
	}
	if len(readCalls) != 1 || !verifiedRead || !verifiedAnswer {
		t.Fatalf("real Norma Read/result chain unproven: calls=%d toolResult=%v finalAnswer=%v activities=%v", len(readCalls), verifiedRead, verifiedAnswer, messages)
	}
	usage, ok := firstState["billed_usage"].(map[string]any)
	if !ok || codexNumber(usage["input_tokens"]) <= 0 || codexNumber(usage["output_tokens"]) <= 0 {
		t.Fatalf("chat authoritative billed usage missing: %v", firstState)
	}

	branch := submit(map[string]any{"client_request_id": rand.Text(), "turn_id": firstState["latest_turn_id"], "expected_revision": firstState["revision"], "title": providerName + " preserved history"}, "conversations", "controls", "branch", "--id", conversationID)
	session, ok := branch["session"].(map[string]any)
	if !ok {
		t.Fatalf("branch session missing: %v", branch)
	}
	branchID := codexString(session["conversation_id"])
	if branchID == "" || branchID == conversationID {
		t.Fatalf("branch did not create a distinct conversation: %v", branch)
	}
	run("workspace", "delete", "--path", relativeFile, "--yes")
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source marker file still available: %v", err)
	}
	run("runtime", "stop", "--yes")
	run("runtime", "start")
	branchState := run("conversations", "controls", "state", "--id", branchID)
	priorBranchRun := codexString(branchState["latest_run_id"])
	submit(map[string]any{"expected_revision": branchState["revision"]}, "conversations", "controls", "pause", "--id", branchID)
	submit(map[string]any{"client_request_id": rand.Text(), "mode": "queue", "message": "From the prior conversation history only, answer exactly REMEMBERED=<the complete marker you already read>. The file has been deleted. Do not call any tool, read any file, guess, or add explanations."}, "conversations", "controls", "inputs", "send", "--id", branchID)
	queued := run("conversations", "controls", "state", "--id", branchID)
	if queued["paused"] != true || queued["active"] != nil || len(codexRows(t, queued, "inputs")) != 1 {
		t.Fatalf("paused admission was not durably queued: %v", queued)
	}
	submit(map[string]any{"expected_revision": queued["revision"]}, "conversations", "controls", "resume", "--id", branchID)
	followupState := waitChat(branchID, priorBranchRun)
	followupRun := codexString(followupState["latest_run_id"])
	followup := run("conversations", "messages", "list", "--id", branchID, "--limit", "100")
	remembered := false
	for _, item := range codexRows(t, followup, "items") {
		if item["run_id"] != followupRun {
			continue
		}
		if item["kind"] == "tool_use" {
			t.Fatalf("history follow-up unexpectedly reread via a tool: %v", item)
		}
		if item["kind"] == "result" && item["is_error"] != true {
			detail := run("conversations", "messages", "get", "--id", branchID, "--seq", codexString(item["seq"]))
			remembered = strings.Contains(codexString(detail["detail"]), "REMEMBERED="+marker)
		}
	}
	if !remembered {
		t.Fatalf("branched follow-up lost prior tool result: %v", followup)
	}
	totals := run("tokens", "conversations")
	billed := false
	for _, row := range codexRows(t, totals, "data") {
		billed = billed || (codexString(row["llm_profile_id"]) == profileID && codexNumber(row["input_tokens"]) > 0 && codexNumber(row["output_tokens"]) > 0)
	}
	if !billed {
		t.Fatalf("persisted conversation usage ledger missing: %v", totals)
	}
	t.Logf("Verified real %s chat, paired Norma Read execution, positive billed usage, branch history and paused queue/resume", providerName)

	labMarker := "SYNTHETIC_LAB_" + rand.Text()
	var fixtureRequests, unexpectedRequests atomic.Int32
	fixture := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/synthetic-debug" {
			unexpectedRequests.Add(1)
			http.NotFound(w, r)
			return
		}
		fixtureRequests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"synthetic_lab_only": true, "public_debug_marker": labMarker, "notice": "Intentionally public toy debug data, never a real credential"})
	}))
	t.Cleanup(fixture.Close)
	target := fixture.URL + "/synthetic-debug"
	// The strict URL rule requires a URL-bearing asset, not the auto-seeded IP.
	// Company scope attaches the pre-registered endpoint before planning starts.
	company := submit(map[string]any{"name": providerName + " owned loopback fixture", "scope": []any{map[string]any{"kind": "ip", "value": "127.0.0.1"}}}, "companies", "create")
	companyID := codexString(company["id"])
	if companyID == "" {
		t.Fatalf("fixture company missing: %v", company)
	}
	registered := submit(map[string]any{"assets": []any{map[string]any{"type": "endpoint", "url": target, "method": "GET", "ip": "127.0.0.1"}}}, "assets", "insert")
	registeredRows := codexRows(t, registered, "results")
	if len(registeredRows) != 1 {
		t.Fatalf("exact fixture endpoint was not registered: %v", registered)
	}
	endpointID := codexString(registeredRows[0]["id"])
	listed := run("assets", "list", "--type", "endpoint", "--dsl", "url=="+target)
	endpointRows := codexRows(t, listed, "assets")
	if len(endpointRows) != 1 || codexString(endpointRows[0]["id"]) != endpointID || endpointRows[0]["url"] != target || endpointRows[0]["method"] != "GET" || codexString(endpointRows[0]["company_id"]) != companyID {
		t.Fatalf("fixture endpoint row/company attribution invalid: %v", listed)
	}
	description := fmt.Sprintf("Owned synthetic local lab only; the sole allowed target is %s. The exact GET endpoint is already registered as asset #%s. Planner: verify that endpoint with list_assets and bind only asset_ids [%s]; the broader IP asset is not allowed by the exact URL rule. This intentionally public toy endpoint returns a public_debug_marker, which is fake lab data and never a real credential. No other host, port, URL, shell command, discovery, exploit, web search, or notification is allowed. Planner: create exactly ONE worker intent to fetch this endpoint with WebFetch and register the observed lab-only debug exposure; do not fetch or report it yourself. Worker: make one GET to that exact URL, use the actual response marker as evidence, and report one LOW finding named SYNTHETIC LAB ONLY public debug marker with explicit synthetic-only scope. Do not invent data or findings elsewhere. Finish after recording that finding. Planner: prove the single goal using the recorded finding and stop the task.", target, endpointID, endpointID)
	task := submit(map[string]any{"name": providerName + " synthetic loopback task", "description": description, "goal": "Record exactly one SYNTHETIC LAB ONLY low-severity finding documenting the public debug marker observed by a worker from the sole owned endpoint, then finish.", "llm_profile_id": json.Number(profileID), "company_ids": []any{json.Number(companyID)}, "timeout_seconds": 420, "seed_first_intent": false, "coverage_enabled": false, "intercept_rules": []any{map[string]any{"enabled": true, "action": "allow", "kind": "exact_url", "pattern": target, "note": "Only this owned synthetic fixture"}}}, "tasks", "create")
	taskID := codexString(task["id"])
	if taskID == "" {
		t.Fatalf("task missing ID: %v", task)
	}
	taskAssets := run("assets", "list", "--type", "endpoint", "--dsl", "url=="+target, "--task-id", taskID)
	taskEndpointRows := codexRows(t, taskAssets, "assets")
	if len(taskEndpointRows) != 1 || codexString(taskEndpointRows[0]["id"]) != endpointID {
		t.Fatalf("task creation did not attach the exact fixture endpoint: %v", taskAssets)
	}
	deadline := time.Now().Add(7 * time.Minute)
	done := false
	for time.Now().Before(deadline) {
		state := run("tasks", "get", "--id", taskID)
		if state["status"] == "done" {
			done = true
			break
		}
		if state["status"] == "failed" {
			t.Fatalf("synthetic planner/worker task failed: %v", state)
		}
		pause(3 * time.Second)
	}
	if !done {
		run("tasks", "control", "--id", taskID, "--action", "pause", "--yes")
		// Preserve full tool and terminal model errors before cleanup removes the database.
		trace := run("exploration", "activity", "list", "--task", taskID, "--limit", "500")
		for _, row := range codexRows(t, trace, "items") {
			if row["is_error"] == true || (row["kind"] == "tool_result" && row["tool"] == "add_intent") {
				detail := run("exploration", "activity", "get", "--task", taskID, "--seq", codexString(row["seq"]))
				t.Logf("task failure detail: %v", detail)
			}
		}
		t.Fatalf("planner/worker task did not finish within budget: task=%s activities=%v", taskID, trace)
	}
	if fixtureRequests.Load() == 0 || unexpectedRequests.Load() != 0 {
		t.Fatalf("owned fixture requests=%d unexpected=%d", fixtureRequests.Load(), unexpectedRequests.Load())
	}
	taskTokens := run("exploration", "tokens", "get", "--task", taskID)
	plannerUsage, workerUsage := false, false
	for _, row := range codexRows(t, taskTokens, "workers") {
		if codexNumber(row["input_tokens"]) <= 0 || codexNumber(row["output_tokens"]) <= 0 {
			continue
		}
		worker := codexString(row["worker"])
		plannerUsage = plannerUsage || worker == "planner"
		workerUsage = workerUsage || strings.HasPrefix(worker, "work#")
	}
	if !plannerUsage || !workerUsage {
		t.Fatalf("real planner/worker usage unproven: %v", taskTokens)
	}
	activity := run("exploration", "activity", "list", "--task", taskID, "--limit", "500")
	plannerIntent, workerFinding, workerHTTPResult := false, false, false
	workerHTTPCalls := map[string]bool{}
	for _, row := range codexRows(t, activity, "items") {
		worker := codexString(row["worker"])
		isWorker := strings.HasPrefix(worker, "work#")
		if row["kind"] == "tool_use" {
			plannerIntent = plannerIntent || (worker == "planner" && row["tool"] == "add_intent")
			workerFinding = workerFinding || (isWorker && row["tool"] == "report_finding")
			if isWorker && row["tool"] == "WebFetch" {
				workerHTTPCalls[codexString(row["tool_use_id"])] = true
			}
		}
		if isWorker && row["kind"] == "tool_result" && row["is_error"] != true && workerHTTPCalls[codexString(row["tool_use_id"])] {
			detail := run("exploration", "activity", "get", "--task", taskID, "--seq", codexString(row["seq"]))
			workerHTTPResult = workerHTTPResult || strings.Contains(codexString(detail["detail"]), labMarker)
		}
	}
	if !plannerIntent || !workerFinding || !workerHTTPResult {
		t.Fatalf("planner intent=%v worker finding=%v paired worker HTTP evidence=%v; activities=%v", plannerIntent, workerFinding, workerHTTPResult, activity)
	}
	report := run("report", "get", "--task", taskID)
	if text := codexString(report["data"]); !strings.Contains(text, labMarker) || !strings.Contains(text, "SYNTHETIC LAB ONLY") {
		t.Fatalf("report lacks actual synthetic finding evidence: %v", report)
	}
	runSubscriptionExtraE2E(t, run, submit, branchID, taskID)
	t.Logf("Verified %d installed CLI invocations; real %s planner/worker through Codex transport, owned loopback HTTP evidence (%d GETs), persisted usage and report. Disposable home: %s", calls, providerName, fixtureRequests.Load(), home)
}

func codexString(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func codexNumber(value any) float64 {
	number, _ := value.(float64)
	return number
}

func codexRows(t *testing.T, object map[string]any, key string) []map[string]any {
	t.Helper()
	values, ok := object[key].([]any)
	if !ok {
		t.Fatalf("expected %s array: %v", key, object)
	}
	rows := make([]map[string]any, 0, len(values))
	for _, value := range values {
		row, ok := value.(map[string]any)
		if !ok {
			t.Fatalf("expected %s object row: %v", key, value)
		}
		rows = append(rows, row)
	}
	return rows
}

func TestCodexFixtureExactEndpointAllowRule(t *testing.T) {
	target := "http://127.0.0.1:63134/synthetic-debug"
	allow := []db.AssetInterceptRule{{Kind: "exact_url", Pattern: target, Enabled: true}}
	for _, tc := range []struct {
		name    string
		urls    []string
		allowed bool
	}{
		{name: "IP asset has no URL"},
		{name: "exact registered endpoint", urls: []string{target}, allowed: true},
		{name: "other path on same loopback IP", urls: []string{"http://127.0.0.1:63134/unrelated"}},
		{name: "same path on different loopback port", urls: []string{"http://127.0.0.1:63135/synthetic-debug"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := db.EvaluateAssetGate(nil, allow, nil, []string{"127.0.0.1"}, tc.urls)
			if got.Allowed != tc.allowed {
				t.Fatalf("allowed=%v want=%v reason=%s", got.Allowed, tc.allowed, got.Reason)
			}
		})
	}
}

// Windows environment names are case insensitive; OAuth must not retain a
// differently-cased Coding Plan key after the general fixture filter.
func codexE2EEnvironment(inherited []string) []string {
	filtered := withoutModelEnvironment(inherited)
	out := filtered[:0]
	for _, item := range filtered {
		key, _, _ := strings.Cut(item, "=")
		if !strings.EqualFold(key, "ZAI_API_KEY") {
			out = append(out, item)
		}
	}
	return out
}

func TestCodexE2EEnvironmentExcludesCodingPlanKey(t *testing.T) {
	filtered := codexE2EEnvironment([]string{"PATH=fixture", "ZAI_API_KEY=fixture-upper", "zai_api_key=fixture-lower"})
	if len(filtered) != 1 || filtered[0] != "PATH=fixture" {
		t.Fatal("OAuth fixture inherited a Coding Plan key or lost its executable path")
	}
}
