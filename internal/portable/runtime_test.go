package portable

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestRuntimeHandlerRequiresAuthenticationForEveryEndpoint(t *testing.T) {
	key := []byte(strings.Repeat("s", 32))
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true; w.WriteHeader(204) })
	handler := runtimeHandler(next, key, runtimeRecord{PID: 123}, func() { t.Fatal("unauthenticated shutdown") })
	for _, path := range []string{"/api/health", "/api/auth/init", "/api/tasks", "/__portable/health", "/__portable/shutdown"} {
		called = false
		req := httptest.NewRequest(http.MethodPost, path, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusUnauthorized || called {
			t.Fatalf("%s: status=%d forwarded=%v", path, response.Code, called)
		}
	}
}
func signedTestToken(t *testing.T, key []byte, expires time.Time) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: "ARTEX", ExpiresAt: jwt.NewNumericDate(expires)}).SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return token
}
func TestRuntimeHandlerHealthAndShutdown(t *testing.T) {
	key := []byte(strings.Repeat("s", 32))
	stopped := false
	record := runtimeRecord{URL: "http://127.0.0.1:12345", PID: 123, Instance: strings.Repeat("a", 32)}
	handler := runtimeHandler(http.NotFoundHandler(), key, record, func() { stopped = true })
	for _, tc := range []struct {
		method, path, token string
		status              int
	}{
		{"GET", "/__portable/health", signedTestToken(t, key, time.Now().Add(time.Hour)), 200},
		{"GET", "/__portable/shutdown", signedTestToken(t, key, time.Now().Add(time.Hour)), 405},
		{"POST", "/__portable/shutdown", signedTestToken(t, key, time.Now().Add(-time.Hour)), 401},
		{"POST", "/__portable/shutdown", signedTestToken(t, key, time.Now().Add(time.Hour)), 202},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		req.Header.Set("Authorization", "Bearer "+tc.token)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != tc.status {
			t.Fatalf("%s %s: %d", tc.method, tc.path, response.Code)
		}
		if tc.path == "/__portable/health" && response.Code == 200 {
			var got runtimeRecord
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil || got != record {
				t.Fatalf("health identity=%+v err=%v", got, err)
			}
		}
	}
	if !stopped {
		t.Fatal("authenticated shutdown did not cancel runtime")
	}
}
func TestStatusDoesNotStartOrCreateHome(t *testing.T) {
	home := filepath.Join(t.TempDir(), "absent")
	_, err := Status(context.Background(), home)
	if !errors.Is(err, ErrNotRunning) {
		t.Fatalf("status: %v", err)
	}
	if _, err = os.Stat(home); !os.IsNotExist(err) {
		t.Fatalf("status created home: %v", err)
	}
}
func TestRuntimeIdentityRejectsRedirectOrRemoteOrigin(t *testing.T) {
	for _, raw := range []string{"http://example.com:1234", "http://127.0.0.1", "https://127.0.0.1:1234", "http://user@127.0.0.1:1234", "http://127.0.0.1:1234/evil", "http://127.0.0.1:1234?token=x", "http://127.0.0.1:99999"} {
		if validRuntimeURL(raw) {
			t.Fatalf("accepted %q", raw)
		}
	}
	if !validRuntimeURL("http://127.0.0.1:1234") {
		t.Fatal("rejected loopback")
	}
}
func TestRuntimeLockPreventsConcurrentOwnersAndReleases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime.lock")
	first, err := tryLock(path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tryLock(path)
	if !lockBusy(err) {
		first.Close()
		t.Fatalf("second owner acquired lock: %v", err)
	}
	first.Close()
	next, err := tryLock(path)
	if err != nil {
		t.Fatalf("released lock remains owned: %v", err)
	}
	next.Close()
}
func TestIsolatedEnvironmentPreservesProviderOnly(t *testing.T) {
	input := []string{"PATH=tools", "ARTEX_PG_DSN=live", "SCOPEWEAVER_CONFIG=source", "ARTEX_CONFIG=source", "ARTEX_SKILL_DIR=source", "PGOPTIONS=bad", "SCOPEWEAVER_TOKEN=secret", "ARTEX_LLM_API_KEY=provider", "SCOPEWEAVER_LLM_BASE_URL=local", "SCOPEWEAVER_LANGUAGE=ko", "SCOPEWEAVER_CODEX_PATH=official-codex.exe"}
	got := isolatedEnv(input)
	joined := strings.Join(got, "\n")
	for _, entry := range input[1:7] {
		if strings.Contains(joined, entry) {
			t.Fatalf("inherited %s", entry)
		}
	}
	for _, entry := range []string{input[0], input[7], input[8], input[9], input[10]} {
		if !strings.Contains(joined, entry) {
			t.Fatalf("lost %s", entry)
		}
	}
}
func TestDatabaseMajorMismatchPreservesCluster(t *testing.T) {
	data := t.TempDir()
	path := filepath.Join(data, "PG_VERSION")
	if err := os.WriteFile(path, []byte("16\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := validateClusterVersion(data, "17"); err == nil || !strings.Contains(err.Error(), "automatic major upgrades are disabled") {
		t.Fatalf("mismatch error=%v", err)
	}
	b, _ := os.ReadFile(path)
	if string(b) != "16\n" {
		t.Fatal("cluster metadata changed")
	}
}
func TestMissingBundleIsActionable(t *testing.T) {
	_, err := postgresToolsAt(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "download the pinned runtime") {
		t.Fatalf("missing bundle: %v", err)
	}
}
func TestDatabaseCredentialsAreNotRegeneratedForExistingCluster(t *testing.T) {
	_, err := newDatabaseSecret(filepath.Join(t.TempDir(), "database.json"), true)
	if err == nil || !strings.Contains(err.Error(), "restore") {
		t.Fatalf("existing cluster missing credentials: %v", err)
	}
}
func TestConnectionJSONDoesNotExposeToken(t *testing.T) {
	b, err := json.Marshal(Connection{URL: "http://127.0.0.1:1", Token: "secret", PID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "secret") || strings.Contains(string(b), "token") {
		t.Fatalf("credentials leaked: %s", b)
	}
}

// TestPostgresLauncherHelper reproduces pg_ctl returning while its background
// postgres process retains inherited output handles. It is invoked only by the
// regression below, with no PostgreSQL binaries or database required.
func TestPostgresLauncherHelper(t *testing.T) {
	if os.Getenv("PORTABLE_POSTGRES_HELPER") != "1" {
		return
	}
	if os.Args[len(os.Args)-1] == "hold-output" {
		time.Sleep(3 * time.Second)
		os.Exit(0)
	}
	child := exec.Command(os.Args[0], "-test.run=^TestPostgresLauncherHelper$", "--", "hold-output")
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr
	hideCommand(child)
	if err := child.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(child.Process.Pid)
	_ = child.Process.Release()
	os.Exit(0)
}
func TestPostgresLauncherDoesNotWaitForBackgroundOutputHandles(t *testing.T) {
	t.Setenv("PORTABLE_POSTGRES_HELPER", "1")
	home := t.TempDir()
	if err := os.MkdirAll(privateDir(home), 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	started := time.Now()
	output, err := runPostgres(ctx, os.Args[0], "-test.run=^TestPostgresLauncherHelper$", "--", "-D", filepath.Join(home, "postgres"))
	elapsed := time.Since(started)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		t.Fatalf("helper did not report its owned child's PID: %q", output)
	}
	if elapsed < 2*time.Second {
		child, findErr := os.FindProcess(pid)
		if findErr == nil {
			_ = child.Kill()
			_, _ = child.Wait()
		}
	} else {
		t.Fatalf("launcher waited for background output handles: %s", elapsed)
	}
}

func TestRestartLauncherHonorsCapturedExecutablePath(t *testing.T) {
	home := t.TempDir()
	if err := prepareHome(home); err != nil {
		t.Fatal(err)
	}
	captured := filepath.Join(home, "captured-executable-that-does-not-exist")
	if err := spawnRuntimeAt(captured, home); err == nil {
		t.Fatal("restart substituted the current executable for the captured launch path")
	}
}

func TestStopReportsOnlyItsAuthenticatedInstanceShutdownError(t *testing.T) {
	for _, tc := range []struct {
		name, resultInstance, resultError, wantError string
		version                                      int
		missing                                      bool
	}{
		{name: "drain timeout", resultInstance: strings.Repeat("a", 32), resultError: "drain background chats: context deadline exceeded", wantError: "drain background chats: context deadline exceeded", version: 1},
		{name: "advertised missing result fails", version: 1, missing: true, wantError: "shutdown result is missing"},
		{name: "advertised stale result fails", resultInstance: strings.Repeat("b", 32), version: 1, wantError: "does not match its authenticated instance"},
		{name: "legacy stale failure ignored", resultInstance: strings.Repeat("b", 32), resultError: "stale failure"},
		{name: "legacy missing result ignored", missing: true},
		{name: "successful instance ignores startup marker", resultInstance: strings.Repeat("a", 32), version: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			if err := prepareHome(home); err != nil {
				t.Fatal(err)
			}
			key := []byte(strings.Repeat("s", 32))
			if err := os.WriteFile(filepath.Join(privateDir(home), "jwt.key"), key, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(privateDir(home), "startup-error.txt"), []byte("stale startup failure"), 0600); err != nil {
				t.Fatal(err)
			}
			owner, err := tryLock(filepath.Join(privateDir(home), "runtime.lock"))
			if err != nil {
				t.Fatal(err)
			}
			defer owner.Close()
			srv := httptest.NewUnstartedServer(nil)
			rec := runtimeRecord{URL: "http://" + srv.Listener.Addr().String(), PID: os.Getpid(), Instance: strings.Repeat("a", 32), ShutdownResultVersion: tc.version}
			srv.Config.Handler = runtimeHandler(http.NotFoundHandler(), key, rec, func() {
				if !tc.missing {
					if err := writeJSONFile(filepath.Join(privateDir(home), "shutdown-result.json"), shutdownResult{Instance: tc.resultInstance, Error: tc.resultError}); err != nil {
						t.Error(err)
					}
				}
				if err := owner.Close(); err != nil {
					t.Error(err)
				}
			})
			srv.Start()
			defer srv.Close()
			if err := writeJSONFile(filepath.Join(privateDir(home), "runtime.json"), rec); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			err = Stop(ctx, home)
			if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("stop lost shutdown error: %v", err)
				}
			} else if err != nil {
				t.Fatalf("stale failure affected stop: %v", err)
			}
		})
	}
}
