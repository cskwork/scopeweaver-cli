package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cskwork/scopeweaver-cli/internal/portable"
)

func testApp() (*App, *bytes.Buffer, *bytes.Buffer) {
	out, errout := new(bytes.Buffer), new(bytes.Buffer)
	return &App{In: strings.NewReader(""), Out: out, Err: errout}, out, errout
}
func TestParsePreservesTypedJSONAndEncodesPaths(t *testing.T) {
	a, _, _ := testApp()
	op := Operation{Name: "tasks create", Path: "/api/tasks/{id}", Method: "POST", Body: true, Parameters: []Parameter{{Name: "id", In: "path", Required: true}, {Name: "enabled", In: "body", Type: "boolean"}, {Name: "limit", In: "query", Type: "integer"}}}
	r, e := a.parse(op, []string{"--id", "a/b ?", "--data", `{"name":"한글","nested":{"value":false}}`, "--enabled", "false", "--limit", "20"})
	if e != nil {
		t.Fatal(e)
	}
	if r.path != "/api/tasks/a%2Fb%20%3F" || r.body["enabled"] != false || r.query.Get("limit") != "20" {
		t.Fatalf("unexpected invocation: %+v", r)
	}
	if r.body["name"] != "한글" {
		t.Fatal("JSON content changed")
	}
}
func TestParseRejectsInvalidInputs(t *testing.T) {
	a, _, _ := testApp()
	op := Operation{Path: "/api/tasks/{id}", Body: true, Parameters: []Parameter{{Name: "id", In: "path", Required: true}, {Name: "enabled", In: "body", Type: "boolean"}}}
	for _, args := range [][]string{{}, {"--id", "1", "--unknown", "x"}, {"--id", "1", "--enabled", "maybe"}, {"--id", "1", "--data", "[]"}, {"--id", "1", "--data", "null"}, {"--id", "1", "--upload", "file"}} {
		if _, err := a.parse(op, args); err == nil {
			t.Fatalf("accepted invalid input: %v", args)
		}
	}
}
func TestBodyFieldsFromFileSatisfyRequiredParameters(t *testing.T) {
	a, _, _ := testApp()
	path := filepath.Join(t.TempDir(), "input.json")
	os.WriteFile(path, []byte(`{"name":"fixture"}`), 0600)
	op := Operation{Body: true, Parameters: []Parameter{{Name: "name", In: "body", Required: true}}}
	r, err := a.parse(op, []string{"--data", "@" + path})
	if err != nil || r.body["name"] != "fixture" {
		t.Fatalf("%+v %v", r, err)
	}
}
func TestJSONOutputAndAuthorization(t *testing.T) {
	a, out, _ := testApp()
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer local-token" || r.Header.Get("Accept-Language") != "ko" {
			t.Error("auth/locale missing")
		}
		io.WriteString(w, `{"ok":true,"name":"fixture"}`)
	}))
	defer backend.Close()
	err := a.execute(context.Background(), options{json: true, lang: "ko", timeout: time.Second}, &portable.Connection{URL: backend.URL, Token: "local-token"}, invocation{operation: Operation{Method: "GET", Response: "json"}, path: "/api/tasks"})
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]any
	if json.Unmarshal(out.Bytes(), &data) != nil || data["ok"] != true {
		t.Fatal(out.String())
	}
}
func TestWritesAreNotRetried(t *testing.T) {
	a, _, _ := testApp()
	calls := 0
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; http.Error(w, "conflict", 409) }))
	defer backend.Close()
	err := a.execute(context.Background(), options{timeout: time.Second}, &portable.Connection{URL: backend.URL}, invocation{operation: Operation{Method: "POST", Body: true}, path: "/api/tasks", body: map[string]any{}})
	if err == nil || calls != 1 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}
func TestRedirectRefusedAndTokenNotForwarded(t *testing.T) {
	a, _, _ := testApp()
	forwarded := false
	external := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { forwarded = true }))
	defer external.Close()
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, external.URL, 302) }))
	defer backend.Close()
	err := a.execute(context.Background(), options{timeout: time.Second}, &portable.Connection{URL: backend.URL, Token: "secret"}, invocation{operation: Operation{Method: "GET"}, path: "/api/tasks"})
	if err == nil || forwarded {
		t.Fatalf("redirect followed: err=%v forwarded=%v", err, forwarded)
	}
}
func TestUploadUsesBackendMultipartField(t *testing.T) {
	a, _, _ := testApp()
	path := filepath.Join(t.TempDir(), "fixture 한글.txt")
	os.WriteFile(path, []byte("fixture bytes"), 0600)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, h, err := r.FormFile("file")
		if err != nil {
			t.Error(err)
			http.Error(w, "missing", 400)
			return
		}
		defer f.Close()
		raw, _ := io.ReadAll(f)
		if string(raw) != "fixture bytes" || h.Filename != filepath.Base(path) {
			t.Error("upload changed")
		}
		io.WriteString(w, `{"ok":true}`)
	}))
	defer backend.Close()
	err := a.execute(context.Background(), options{timeout: time.Second}, &portable.Connection{URL: backend.URL}, invocation{operation: Operation{Method: "POST", Multipart: "file", Response: "json"}, path: "/api/workspace/upload", uploads: []string{path}})
	if err != nil {
		t.Fatal(err)
	}
}
func TestDownloadIntegrityAndNoOverwrite(t *testing.T) {
	a, _, _ := testApp()
	path := filepath.Join(t.TempDir(), "result.bin")
	payload := []byte{0, 1, 2, 255, 13, 10}
	if err := a.download(options{output: path}, bytes.NewReader(payload)); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(got, payload) {
		t.Fatal("download changed bytes")
	}
	if err := a.download(options{output: path}, strings.NewReader("overwrite")); err == nil {
		t.Fatal("overwrote existing file")
	}
	got, _ = os.ReadFile(path)
	if !bytes.Equal(got, payload) {
		t.Fatal("original file changed")
	}
}
func TestServerSentEventsPreserveIDsAndMultilinePayload(t *testing.T) {
	out := new(bytes.Buffer)
	input := ": heartbeat\nid: 12\nevent: progress\ndata: {\"done\":false}\n\nid: 13\ndata: line one\ndata: line two\n\n"
	if err := streamEvents(context.Background(), strings.NewReader(input), out); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatal(out.String())
	}
	var data map[string]any
	json.Unmarshal([]byte(lines[1]), &data)
	if data["id"] != "13" || data["data"] != "line one\nline two" || data["event"] != "message" {
		t.Fatal(data)
	}
}
func TestStreamCancellation(t *testing.T) {
	a, _, _ := testApp()
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: {}\n\n")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer backend.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := a.execute(ctx, options{timeout: time.Second}, &portable.Connection{URL: backend.URL}, invocation{operation: Operation{Method: "GET", Response: "sse"}, path: "/events"})
	if err == nil {
		t.Fatal("stream ignored cancellation")
	}
}
func TestHelpAndDryRunNeverStartRuntime(t *testing.T) {
	a, _, _ := testApp()
	a.Connect = func(context.Context, string) (*portable.Connection, error) {
		t.Fatal("started runtime for discovery")
		return nil, nil
	}
	for _, args := range [][]string{{"--help"}, {"commands", "--json"}, {"schema", "--json"}, {"tasks", "list", "--dry-run", "--json"}} {
		if code := a.Run(context.Background(), args); code != 0 {
			t.Fatalf("%v: %d", args, code)
		}
	}
}
func TestMutationNeedsExplicitConfirmation(t *testing.T) {
	a, _, errout := testApp()
	a.Connect = func(context.Context, string) (*portable.Connection, error) {
		t.Fatal("started runtime before confirmation")
		return nil, nil
	}
	op := Operation{}
	for _, candidate := range Operations() {
		if candidate.Effect != "read" && len(candidate.Parameters) == 0 && candidate.Multipart == "" {
			op = candidate
			break
		}
	}
	if op.Name == "" {
		t.Fatal("catalog has no parameter-free mutation")
	}
	args := append(strings.Fields(op.Name), "--json")
	if code := a.Run(context.Background(), args); code != 2 || !strings.Contains(errout.String(), "confirmation_required") {
		t.Fatalf("%s: %d %s", op.Name, code, errout.String())
	}
}
func TestDryRunRedactsNestedSecrets(t *testing.T) {
	got := redact(map[string]any{"api_key": "secret", "nested": map[string]any{"password": "secret2", "name": "keep"}})
	raw, _ := json.Marshal(got)
	if strings.Contains(string(raw), "secret") || !strings.Contains(string(raw), "keep") {
		t.Fatal(string(raw))
	}
}
func TestREPLQuotesAndNoShellExpansion(t *testing.T) {
	args, err := splitLine("tasks create --data '{\"name\":\"two words\"}' --field 'goal=$(echo nope)'")
	if err != nil || len(args) != 6 || args[3] != "{\"name\":\"two words\"}" || args[5] != "goal=$(echo nope)" {
		t.Fatalf("%v %v", args, err)
	}
	if _, err = splitLine("tasks 'unterminated"); err == nil {
		t.Fatal("accepted unterminated quote")
	}
}

func TestREPLPreservesWindowsPaths(t *testing.T) {
	args, err := splitLine("workspace upload --upload \"C:\\Users\\a\\My Documents\\fixture.txt\"")
	if err != nil || args[3] != "C:\\Users\\a\\My Documents\\fixture.txt" {
		t.Fatalf("%q %v", args, err)
	}
}
func TestREPLReadsStdinJSONAndPreservesNextCommand(t *testing.T) {
	a, out, errout := testApp()
	a.In = strings.NewReader("workspace write --data - --dry-run --json\n{\"path\":\"fixture.txt\",\"content\":\"ok\"}\nversion --json\nquit\n")
	if code := a.Run(context.Background(), []string{"repl", "--json"}); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errout)
	}
	decoder := json.NewDecoder(out)
	var first, second map[string]any
	if decoder.Decode(&first) != nil || decoder.Decode(&second) != nil || first["command"] != "workspace write" || second["version"] != Version {
		t.Fatalf("output=%s", out)
	}
}
func TestRuntimeDryRunCannotConnect(t *testing.T) {
	a, out, _ := testApp()
	a.Connect = func(context.Context, string) (*portable.Connection, error) {
		t.Fatal("dry-run started runtime")
		return nil, nil
	}
	for _, action := range []string{"start", "stop"} {
		if code := a.Run(context.Background(), []string{"runtime", action, "--yes", "--dry-run", "--json"}); code != 0 {
			t.Fatalf("dry-run code=%d", code)
		}
	}
	if !strings.Contains(out.String(), "dry_run") {
		t.Fatal(out)
	}
}
func TestCommandsFamilyIsNotShadowedByDiscovery(t *testing.T) {
	a, out, _ := testApp()
	if code := a.Run(context.Background(), []string{"commands", "list", "--dry-run", "--json"}); code != 0 {
		t.Fatal(code)
	}
	if !strings.Contains(out.String(), "/api/commands") {
		t.Fatal(out)
	}
}

func TestLocalizedErrorKeepsDiagnosticSeparate(t *testing.T) {
	a, _, errout := testApp()
	if code := a.Run(context.Background(), []string{"--lang", "ko", "--json", "does-not-exist"}); code != 2 {
		t.Fatal(code)
	}
	var value map[string]map[string]any
	if err := json.Unmarshal(errout.Bytes(), &value); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(fmt.Sprint(value["error"]["message"]), "명령") || !strings.Contains(fmt.Sprint(value["error"]["detail"]), "does-not-exist") {
		t.Fatal(value)
	}
}
func TestBackendErrorRedactsSubmittedCredential(t *testing.T) {
	a, _, _ := testApp()
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "invalid key super-secret-fixture", 400) }))
	defer backend.Close()
	err := a.execute(context.Background(), options{timeout: time.Second}, &portable.Connection{URL: backend.URL}, invocation{operation: Operation{Method: "POST", Body: true}, path: "/api/llm", body: map[string]any{"api_key": "super-secret-fixture"}})
	if err == nil || strings.Contains(err.Error(), "super-secret-fixture") {
		t.Fatal(err)
	}
}
