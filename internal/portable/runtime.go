// Package portable owns the CLI's private, persistent engine and database.
package portable

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cskwork/scopeweaver-cli/agent"
	"github.com/cskwork/scopeweaver-cli/locale"
	"github.com/cskwork/scopeweaver-cli/selfupdate"
	"github.com/cskwork/scopeweaver-cli/server"
	"github.com/golang-jwt/jwt/v5"
)

var ErrNotRunning = errors.New("private runtime is not running")

type Connection struct {
	URL                   string `json:"url"`
	Token                 string `json:"-"`
	PID                   int    `json:"pid"`
	instance              string
	shutdownResultVersion int
}
type runtimeRecord struct {
	URL                   string `json:"url"`
	PID                   int    `json:"pid"`
	Instance              string `json:"instance"`
	ShutdownResultVersion int    `json:"shutdown_result_version,omitempty"`
}
type shutdownResult struct {
	Instance string `json:"instance"`
	Error    string `json:"error,omitempty"`
}

type fileLock struct{ file *os.File }

func canonicalHome(home string) (string, error) {
	if strings.TrimSpace(home) == "" {
		return "", errors.New("runtime home must not be empty")
	}
	return filepath.Abs(home)
}
func privateDir(home string) string { return filepath.Join(home, "private") }
func prepareHome(home string) error {
	for _, dir := range []string{home, privateDir(home), filepath.Join(home, "workspace")} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	return protectDir(privateDir(home))
}
func tryLock(path string) (*fileLock, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = lockFile(f); err != nil {
		f.Close()
		return nil, err
	}
	return &fileLock{file: f}, nil
}
func (l *fileLock) Close() error { return l.file.Close() }
func acquireLock(ctx context.Context, path string) (*fileLock, error) {
	for {
		l, err := tryLock(path)
		if err == nil {
			return l, nil
		}
		if !lockBusy(err) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}
func activeRuntime(home string) (bool, error) {
	l, err := tryLock(filepath.Join(privateDir(home), "runtime.lock"))
	if err == nil {
		l.Close()
		return false, nil
	}
	if lockBusy(err) {
		return true, nil
	}
	return false, err
}
func freshToken(home string) (string, error) {
	key, err := os.ReadFile(filepath.Join(privateDir(home), "jwt.key"))
	if err != nil {
		return "", err
	}
	key = []byte(strings.TrimSpace(string(key)))
	if len(key) < 32 {
		return "", errors.New("invalid private runtime signing key")
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject: "ARTEX", IssuedAt: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString(key)
}
func validRuntimeURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "http" || u.Hostname() != "127.0.0.1" || u.Port() == "" ||
		u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return false
	}
	_, err = net.LookupPort("tcp", u.Port())
	return err == nil
}
func localClient() *http.Client {
	return &http.Client{Timeout: 2 * time.Second,
		Transport:     &http.Transport{Proxy: nil},
		CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("private runtime redirect refused") },
	}
}

// Status inspects an existing runtime only; it never initializes or starts one.
func Status(ctx context.Context, home string) (*Connection, error) {
	home, err := canonicalHome(home)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(filepath.Join(privateDir(home), "runtime.json"))
	if os.IsNotExist(err) {
		return nil, ErrNotRunning
	}
	if err != nil {
		return nil, err
	}
	var rec runtimeRecord
	if err = json.Unmarshal(b, &rec); err != nil {
		return nil, fmt.Errorf("invalid runtime metadata: %w", err)
	}
	if !validRuntimeURL(rec.URL) || rec.PID <= 0 || len(rec.Instance) != 32 {
		return nil, errors.New("invalid private runtime identity")
	}
	token, err := freshToken(home)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rec.URL+"/__portable/health", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	client := localClient()
	defer client.CloseIdleConnections()
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotRunning, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("private runtime health returned HTTP %d", res.StatusCode)
	}
	var got runtimeRecord
	if err = json.NewDecoder(io.LimitReader(res.Body, 4096)).Decode(&got); err != nil {
		return nil, fmt.Errorf("private runtime health: %w", err)
	}
	if got != rec {
		return nil, errors.New("private runtime identity does not match its state")
	}
	return &Connection{URL: rec.URL, Token: token, PID: rec.PID, instance: rec.Instance, shutdownResultVersion: rec.ShutdownResultVersion}, nil
}

// Ensure starts the portable engine once and reuses it across CLI invocations.
func Ensure(ctx context.Context, home string) (*Connection, error) {
	home, err := canonicalHome(home)
	if err != nil {
		return nil, err
	}
	if err = prepareHome(home); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 12*time.Minute)
	defer cancel()
	start, err := acquireLock(ctx, filepath.Join(privateDir(home), "start.lock"))
	if err != nil {
		return nil, err
	}
	defer start.Close()
	if c, err := Status(ctx, home); err == nil {
		return c, nil
	}
	active, err := activeRuntime(home)
	if err != nil {
		return nil, err
	}
	if !active {
		if err = ensurePostgres(ctx); err != nil {
			return nil, err
		}
		if _, err = postgresBinaries(); err != nil {
			return nil, err
		}
		if err = spawnRuntime(home); err != nil {
			return nil, err
		}
	}
	readyCtx, readyCancel := context.WithTimeout(ctx, 90*time.Second)
	defer readyCancel()
	ctx = readyCtx
	for {
		if c, err := Status(ctx, home); err == nil {
			return c, nil
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("private runtime did not become ready: %w; see %s", ctx.Err(), filepath.Join(privateDir(home), "runtime.log"))
		case <-time.After(150 * time.Millisecond):
		}
		// A child may still be starting before it obtains its lock. Inspect the
		// failure marker rather than interpreting a briefly free lock as failure.
		if b, err := os.ReadFile(filepath.Join(privateDir(home), "startup-error.txt")); err == nil && len(b) > 0 {
			return nil, fmt.Errorf("private runtime startup failed: %s", strings.TrimSpace(string(b)))
		}
	}
}
func spawnRuntime(home string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return spawnRuntimeAt(exe, home)
}

// spawnRuntimeAt deliberately uses the captured launch path. After an update
// swaps a running executable, os.Executable can identify its renamed old file.
func spawnRuntimeAt(exe, home string) error {
	_ = os.Remove(filepath.Join(privateDir(home), "startup-error.txt"))
	out, err := os.OpenFile(filepath.Join(privateDir(home), "runtime.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer out.Close()
	cmd := exec.Command(exe, "__runtime", "--home", home)
	cmd.Dir = filepath.Join(home, "workspace")
	cmd.Env = isolatedEnv(os.Environ())
	cmd.Stdout = out
	cmd.Stderr = out
	detach(cmd)
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("start private runtime: %w", err)
	}
	return cmd.Process.Release()
}

// Stop requests an authenticated graceful shutdown, then waits for ownership
// to be released. It never kills a PID obtained from a stale metadata file.
func Stop(ctx context.Context, home string) error {
	home, err := canonicalHome(home)
	if err != nil {
		return err
	}
	if _, err = os.Stat(privateDir(home)); os.IsNotExist(err) {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 75*time.Second)
	defer cancel()
	start, err := acquireLock(ctx, filepath.Join(privateDir(home), "start.lock"))
	if err != nil {
		return err
	}
	defer start.Close()
	c, statusErr := Status(ctx, home)
	if statusErr != nil {
		active, err := activeRuntime(home)
		if err != nil {
			return err
		}
		if active {
			return fmt.Errorf("cannot authenticate running private runtime: %w", statusErr)
		}
		return stopOrphanPostgres(ctx, home)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL+"/__portable/shutdown", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	client := localClient()
	defer client.CloseIdleConnections()
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("private shutdown returned HTTP %d", res.StatusCode)
	}
	for {
		active, err := activeRuntime(home)
		if err != nil {
			return err
		}
		if !active {
			return readShutdownResult(home, c.instance, c.shutdownResultVersion > 0)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("runtime shutdown incomplete: %w", ctx.Err())
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// A result belongs only to the instance Stop authenticated and asked to shut
// down. Earlier startup failures and stale results must not poison a later stop.
func readShutdownResult(home, instance string, required bool) error {
	b, err := os.ReadFile(filepath.Join(privateDir(home), "shutdown-result.json"))
	if os.IsNotExist(err) {
		if required {
			return errors.New("private runtime shutdown result is missing")
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("read private shutdown result: %w", err)
	}
	var result shutdownResult
	if err = json.Unmarshal(b, &result); err != nil {
		return fmt.Errorf("invalid private shutdown result: %w", err)
	}
	if result.Instance != instance && required {
		return errors.New("private runtime shutdown result does not match its authenticated instance")
	}
	if result.Instance == instance && result.Error != "" {
		return fmt.Errorf("private runtime shutdown failed: %s", result.Error)
	}
	return nil
}

func isolatedEnv(env []string) []string {
	result := make([]string, 0, len(env))
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		key = strings.ToUpper(key)
		if strings.HasPrefix(key, "PG") {
			continue
		}
		if strings.HasPrefix(key, "ARTEX_") || strings.HasPrefix(key, "SCOPEWEAVER_") {
			if !strings.HasPrefix(key, "ARTEX_LLM_") && !strings.HasPrefix(key, "SCOPEWEAVER_LLM_") &&
				key != "ARTEX_LOCAL_LLM_ONLY" && key != "SCOPEWEAVER_LOCAL_LLM_ONLY" &&
				key != "ARTEX_LOCAL_LLM_ALLOWED_CIDRS" && key != "SCOPEWEAVER_LOCAL_LLM_ALLOWED_CIDRS" &&
				key != "SCOPEWEAVER_LANGUAGE" && key != "SCOPEWEAVER_CODEX_PATH" {
				continue
			}
		}
		result = append(result, entry)
	}
	return result
}

// Run is the hidden child entrypoint. It must execute in its own process:
// engine configuration and agent hooks are process globals.
func Run(ctx context.Context, home string) (retErr error) {
	home, retErr = canonicalHome(home)
	if retErr != nil {
		return retErr
	}
	if retErr = prepareHome(home); retErr != nil {
		return retErr
	}
	// Capture this before Bootstrap can rename or replace the running binary.
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	restart := false
	defer func() {
		if retErr != nil {
			_ = os.WriteFile(filepath.Join(privateDir(home), "startup-error.txt"), []byte(retErr.Error()+"\n"), 0600)
		}
		if restart {
			if err := spawnRuntimeAt(exe, home); err != nil {
				retErr = errors.Join(retErr, err)
			}
		}
	}()
	owner, err := tryLock(filepath.Join(privateDir(home), "runtime.lock"))
	if err != nil {
		return fmt.Errorf("runtime already owned: %w", err)
	}
	defer owner.Close()
	instance := ""
	defer func() {
		if instance == "" {
			return
		}
		result := shutdownResult{Instance: instance}
		if retErr != nil {
			result.Error = retErr.Error()
		}
		if err := writeJSONFile(filepath.Join(privateDir(home), "shutdown-result.json"), result); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("write private shutdown result: %w", err))
		}
	}()
	_ = os.Remove(filepath.Join(privateDir(home), "startup-error.txt"))
	action, state := selfupdate.Bootstrap()
	server.SetBootUpdateState(state)
	if action == selfupdate.Restart {
		restart = true
		return nil
	}
	if err = os.Chdir(filepath.Join(home, "workspace")); err != nil {
		return err
	}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		upper := strings.ToUpper(key)
		if strings.HasPrefix(upper, "PG") || strings.HasPrefix(upper, "ARTEX_") || strings.HasPrefix(upper, "SCOPEWEAVER_") {
			keep := false
			for _, allowed := range isolatedEnv([]string{entry}) {
				if allowed == entry {
					keep = true
				}
			}
			if !keep {
				if err = os.Unsetenv(key); err != nil {
					return err
				}
			}
		}
	}
	configPath := filepath.Join(privateDir(home), "config.json")

	if _, err = os.Stat(configPath); os.IsNotExist(err) {
		if err = os.WriteFile(configPath, []byte("{}\n"), 0600); err != nil {
			return err
		}
	}
	skillDir, err := seedSkills(ctx, filepath.Join(filepath.Dir(exe), "skills"), home)
	if err != nil {
		return err
	}
	for key, value := range map[string]string{
		"SCOPEWEAVER_CONFIG": configPath, "ARTEX_CONFIG": configPath,
		"ARTEX_SKILL_DIR": skillDir,
	} {
		if err = os.Setenv(key, value); err != nil {
			return err
		}
	}
	lang, _ := locale.FromEnv(os.Getenv)
	locale.SetServerDefault(lang)
	pg, err := startPostgres(ctx, home)
	if err != nil {
		return err
	}
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		retErr = errors.Join(retErr, pg.stop(stopCtx))
	}()
	if err = os.Setenv("ARTEX_PG_DSN", pg.dsn); err != nil {
		return err
	}
	runtimeCtx, cancel := context.WithCancelCause(context.Background())
	defer cancel(agent.AbortShutdown)
	go func() {
		select {
		case <-ctx.Done():
			cancel(agent.AbortShutdown)
		case <-runtimeCtx.Done():
		}
	}()
	server.StartLogCapture()
	proxyPort, err := freePort()
	if err != nil {
		return err
	}
	mgr, err := server.NewManager(filepath.Join(home, "workspace"), net.JoinHostPort("127.0.0.1", fmt.Sprint(proxyPort)))
	if err != nil {
		return err
	}
	defer func() {
		cancel(agent.AbortShutdown)
		_ = mgr.Close()
	}()
	settle := time.AfterFunc(selfupdate.SettleDelay, selfupdate.Settle)
	defer settle.Stop()
	srv, err := server.NewWithError(runtimeCtx, mgr, os.Getenv("ARTEX_SKILL_DIR"), filepath.Join(home, "workspace"), privateDir(home))
	if err != nil {
		return err
	}
	defer func() {
		cancel(agent.AbortShutdown)
		drainCtx, drainCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer drainCancel()
		if err := srv.DrainChats(drainCtx); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("drain background chats: %w", err))
		}
	}()
	key, err := os.ReadFile(filepath.Join(privateDir(home), "jwt.key"))
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer listener.Close()
	nonce := make([]byte, 16)
	if _, err = rand.Read(nonce); err != nil {
		return err
	}
	rec := runtimeRecord{URL: "http://" + listener.Addr().String(), PID: os.Getpid(), Instance: hex.EncodeToString(nonce), ShutdownResultVersion: 1}
	instance = rec.Instance
	httpSrv := &http.Server{Handler: runtimeHandler(srv.Handler(), []byte(strings.TrimSpace(string(key))), rec, func() { cancel(agent.AbortShutdown) }), ReadHeaderTimeout: 10 * time.Second}
	served := make(chan error, 1)
	go func() { served <- httpSrv.Serve(listener) }()
	recordPath := filepath.Join(privateDir(home), "runtime.json")
	if err = writeJSONFile(recordPath, rec); err != nil {
		httpSrv.Close()
		return err
	}
	defer os.Remove(recordPath)
	log.Printf("ScopeWeaver CLI private runtime ready on %s (pid=%d)", rec.URL, rec.PID)
	select {
	case <-runtimeCtx.Done():
	case <-server.RestartRequested():
		restart = true
		cancel(agent.AbortShutdown)
	case err = <-served:
		if !errors.Is(err, http.ErrServerClosed) {
			retErr = err
		}
		cancel(agent.AbortShutdown)
	}
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err = httpSrv.Shutdown(shutdownCtx); err != nil {
		_ = httpSrv.Close()
		retErr = errors.Join(retErr, err)
	}
	return retErr
}
func runtimeHandler(next http.Handler, key []byte, rec runtimeRecord, shutdown func()) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		token, err := jwt.ParseWithClaims(raw, &jwt.RegisteredClaims{}, func(*jwt.Token) (any, error) { return key, nil },
			jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithSubject("ARTEX"))
		if err != nil || !token.Valid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"error":"Private runtime authentication required"}`)
			return
		}
		switch r.URL.Path {
		case "/__portable/health":
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(rec)
		case "/__portable/shutdown":
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_, _ = io.WriteString(w, `{"ok":true}`)
			shutdown()
		default:
			next.ServeHTTP(w, r)
		}
	})
}
func writeJSONFile(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".state-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(append(b, '\n'))
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	// The runtime owns this file exclusively. Remove permits replacement on Windows.
	if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(tmp, path)
}
