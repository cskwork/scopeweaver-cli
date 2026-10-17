package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPortableCLIEndToEnd(t *testing.T) {
	binary := os.Getenv("SCOPEWEAVER_CLI_E2E_BINARY")
	if binary == "" {
		t.Skip("set SCOPEWEAVER_CLI_E2E_BINARY to the packaged executable")
	}
	binary, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(binary); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(t.TempDir(), "portable state 한글")
	cwd := t.TempDir()
	calls := 0
	run := func(args ...string) map[string]any {
		t.Helper()
		calls++
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binary, append([]string{"--home", home, "--json"}, args...)...)
		cmd.Dir = cwd
		cmd.Env = withoutModelEnvironment(os.Environ())
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("%v failed: %v\nstdout=%s\nstderr=%s", args, err, stdout.String(), stderr.String())
		}
		var data map[string]any
		if err := json.Unmarshal(stdout.Bytes(), &data); err != nil {
			var other any
			if json.Unmarshal(stdout.Bytes(), &other) != nil {
				t.Fatalf("%v invalid JSON: %s", args, stdout.String())
			}
			return map[string]any{"data": other}
		}
		return data
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()
		output, err := exec.CommandContext(ctx, binary, "--home", home, "runtime", "stop", "--yes", "--json").CombinedOutput()
		if err != nil {
			t.Errorf("runtime cleanup failed: %v %s", err, output)
		}
	}()
	run("version")
	run("schema", "tasks", "create")
	run("runtime", "start")
	health := run("health")
	if health["ok"] != true {
		t.Fatalf("unhealthy: %#v", health)
	}
	reads := []string{
		"tasks list", "categories list", "templates list", "archives list", "companies list",
		"assets list", "findings list", "findings stats", "agents list", "conversations list",
		"llm profiles list", "llm pool status", "llm retry-policy get", "llm records list",
		"tools list", "mcp list", "skills list", "settings get", "notify channels list",
		"intercept rules list", "intercept pending list", "asset-intercept rules list",
		"traffic hosts list", "commands list", "commands stats", "logs history", "tokens daily",
	}
	for _, name := range reads {
		if len(selectOperations(name)) == 0 {
			t.Fatalf("E2E command absent: %s", name)
		}
		run(strings.Fields(name)...)
	}
	category := run("categories", "create", "--name", "Portable fixture", "--yes")
	if category["id"] == nil {
		t.Fatal(category)
	}
	run("templates", "create", "--name", "Fixture template", "--description", "Local isolated fixture only", "--goal", "Verify persistence without contacting any target", "--yes")
	task := run("tasks", "create", "--name", "Portable task", "--description", "Local isolated fixture; do not contact targets.", "--goal", "Verify task persistence", "--yes")
	taskID := fmt.Sprint(task["id"])
	if taskID == "" || taskID == "<nil>" {
		t.Fatal(task)
	}
	run("tasks", "control", "--id", taskID, "--action", "pause", "--yes")
	paused := run("tasks", "get", "--id", taskID)
	if paused["paused"] != true {
		t.Fatal(paused)
	}
	run("tasks", "goals", "create", "--id", taskID, "--text", "Inspect synthetic fixture", "--yes")
	run("tasks", "constraints", "create", "--id", taskID, "--text", "Only isolated fixture data", "--kind", "deny", "--yes")
	for _, name := range []string{"tasks goals list", "tasks constraints list", "tasks coverage get", "tasks scope list"} {
		if len(selectOperations(name)) == 0 {
			t.Fatalf("E2E command absent: %s", name)
		}
		run(append(strings.Fields(name), "--id", taskID)...)
	}
	company := run("companies", "create", "--name", "Portable fixture company", "--yes")
	run("companies", "get", "--id", fmt.Sprint(company["id"]))
	run("agents", "create", "--key", "portable_fixture", "--name", "Portable fixture agent", "--yes")
	conv := run("conversations", "create", "--agent-key", "portable_fixture", "--title", "Portable conversation", "--yes")
	convID := fmt.Sprint(conv["id"])
	run("conversations", "messages", "list", "--id", convID)
	run("workspace", "mkdir", "--path", "fixtures", "--yes")
	run("workspace", "write", "--path", "fixtures/unicode.txt", "--content", "portable 한글 中文", "--yes")
	read := run("workspace", "read", "--path", "fixtures/unicode.txt")
	if !strings.Contains(fmt.Sprint(read), "portable 한글 中文") {
		t.Fatal(read)
	}
	source := filepath.Join(cwd, "upload.bin")
	payload := []byte{0, 1, 2, 255, 13, 10}
	if err := os.WriteFile(source, payload, 0600); err != nil {
		t.Fatal(err)
	}
	run("workspace", "upload", "--path", "fixtures", "--upload", source, "--yes")
	downloaded := filepath.Join(cwd, "download.bin")
	run("workspace", "download", "--path", "fixtures/upload.bin", "--output", downloaded)
	got, err := os.ReadFile(downloaded)
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatalf("download mismatch: %v %v", got, err)
	}
	run("report", "get", "--task", taskID)
	exportPath := filepath.Join(cwd, "findings.zip")
	run("findings", "export", "--format", "md-zip", "--output", exportPath)
	if info, err := os.Stat(exportPath); err != nil || info.Size() == 0 {
		t.Fatalf("empty export: %v", err)
	}
	// Adding a goal intentionally resumes the task (server/goals_api.go).
	if state := run("tasks", "get", "--id", taskID); state["paused"] != false {
		t.Fatalf("new goal did not revive task: %#v", state)
	}
	run("tasks", "control", "--id", taskID, "--action", "pause", "--yes")
	run("runtime", "stop", "--yes")
	run("runtime", "start")
	persisted := run("tasks", "get", "--id", taskID)
	if persisted["name"] != "Portable task" || persisted["paused"] != true {
		t.Fatalf("restart lost state: %#v", persisted)
	}
	// Move only this test's stopped data directory. The next invocation must recompute paths.
	run("runtime", "stop", "--yes")
	moved := filepath.Join(filepath.Dir(home), "moved portable state")
	if err := os.Rename(home, moved); err != nil {
		t.Fatal(err)
	}
	home = moved
	run("runtime", "start")
	if got := run("tasks", "get", "--id", taskID); got["name"] != "Portable task" {
		t.Fatalf("relocation lost state: %#v", got)
	}
	// One failed command must remain a failure to the caller and keep JSON stderr separate.
	cmd := exec.Command(binary, "--home", home, "--json", "tasks", "get", "--id", "999999999")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = cwd
	if err := cmd.Run(); err == nil {
		t.Fatal("missing task exited successfully")
	}
	if stdout.Len() != 0 || !json.Valid(stderr.Bytes()) {
		t.Fatalf("failure output stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	t.Logf("Verified %d installed CLI invocations against bundled PostgreSQL; no model or target calls. Source cwd: %s", calls, cwd)
}
func withoutModelEnvironment(env []string) []string {
	out := []string{}
	for _, item := range env {
		key, _, _ := strings.Cut(item, "=")
		if strings.Contains(key, "API_KEY") || strings.HasPrefix(key, "ANTHROPIC_") || strings.HasPrefix(key, "OPENAI_") || strings.HasPrefix(key, "ARTEX_") || strings.HasPrefix(key, "SCOPEWEAVER_") {
			continue
		}
		out = append(out, item)
	}
	return out
}
