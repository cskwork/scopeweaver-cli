package portable

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cskwork/scopeweaver-cli/internal/runtimebundle"
	"github.com/jackc/pgx/v5"
)

type postgresTools struct{ initdb, ctl, postgres string }
type databaseSecret struct {
	Password string `json:"password"`
}
type postgresProcess struct {
	tools               postgresTools
	data, dsn, adminDSN string
	port                int
}

// The lock belongs to the executable folder, not a home: distinct private
// homes share one portable runtime installation.
func ensurePostgres(ctx context.Context) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return ensurePostgresAt(ctx, filepath.Dir(exe), runtime.GOOS+"-"+runtime.GOARCH)
}
func ensurePostgresAt(ctx context.Context, root, target string) error {
	return ensurePostgresWith(ctx, root, target, runtimebundle.Ensure)
}
func ensurePostgresWith(ctx context.Context, root, target string, install func(context.Context, string, string) error) error {
	// Refuse redirected binary directories before checking an existing runtime
	// or creating the shared lock directory.
	if err := runtimebundle.CheckParents(filepath.Join(root, "runtime", "postgres", "bin")); err != nil {
		return err
	}
	if err := runtimebundle.Validate(root, target); err == nil {
		return nil
	}
	dir := filepath.Join(root, "runtime")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create portable runtime next to CLI (check folder permissions): %w", err)
	}
	lock, err := acquireLock(ctx, filepath.Join(dir, "install.lock"))
	if err != nil {
		return fmt.Errorf("lock portable runtime next to CLI (check folder permissions): %w", err)
	}
	defer lock.Close()
	return install(ctx, root, target)
}

func postgresBinaries() (postgresTools, error) {
	exe, err := os.Executable()
	if err != nil {
		return postgresTools{}, err
	}
	return postgresToolsAt(filepath.Join(filepath.Dir(exe), "runtime", "postgres", "bin"))
}
func postgresToolsAt(bin string) (postgresTools, error) {
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	tools := postgresTools{filepath.Join(bin, "initdb"+suffix), filepath.Join(bin, "pg_ctl"+suffix), filepath.Join(bin, "postgres"+suffix)}
	for _, path := range []string{tools.initdb, tools.ctl, tools.postgres} {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			return postgresTools{}, fmt.Errorf("portable PostgreSQL is missing %s; run an operational command to download the pinned runtime", path)
		}
		if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
			return postgresTools{}, fmt.Errorf("bundled PostgreSQL is not executable: %s", path)
		}
	}
	return tools, nil
}

var postgresVersionPattern = regexp.MustCompile(`PostgreSQL\)\s+(\d+)(?:\.\d+)*`)

func postgresMajor(output string) (string, error) {
	match := postgresVersionPattern.FindStringSubmatch(output)
	if len(match) != 2 {
		return "", fmt.Errorf("cannot identify bundled PostgreSQL version: %s", strings.TrimSpace(output))
	}
	major, err := strconv.Atoi(match[1])
	if err != nil || major < 10 {
		return "", errors.New("bundled PostgreSQL 10 or newer is required")
	}
	return match[1], nil
}
func validateClusterVersion(data, major string) error {
	b, err := os.ReadFile(filepath.Join(data, "PG_VERSION"))
	if err != nil {
		return err
	}
	existing := strings.TrimSpace(string(b))
	if existing != major {
		return fmt.Errorf("database cluster requires PostgreSQL %s but bundle contains %s; keep the matching portable release or explicitly migrate a backup (automatic major upgrades are disabled)", existing, major)
	}
	return nil
}
func runPostgres(ctx context.Context, exe string, args ...string) (string, error) {
	cmd, err := newPostgresCommand(ctx, exe, args...)
	if err != nil {
		return "", err
	}
	// pg_ctl's background postgres child can inherit stdout/stderr handles on
	// Windows. Pipes make CombinedOutput wait for that child's entire lifetime.
	// Real files let us wait only for pg_ctl while retaining startup diagnostics.
	var capture *os.File
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "-D" {
			capture, err = os.OpenFile(filepath.Join(filepath.Dir(args[i+1]), "private", "postgres-command.log"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
			break
		}
	}
	if capture == nil && err == nil {
		capture, err = os.CreateTemp("", "scopeweaver-pg-command-*.log")
		if err == nil {
			defer os.Remove(capture.Name())
		}
	}
	if err != nil {
		return "", err
	}
	cmd.Stdout = capture
	cmd.Stderr = capture
	err = cmd.Run()
	_ = capture.Close()
	out, readErr := os.ReadFile(capture.Name())
	if readErr != nil {
		return "", errors.Join(err, readErr)
	}
	output := string(out)
	if err != nil {
		if len(output) > 3000 {
			output = output[len(output)-3000:]
		}
		return string(out), fmt.Errorf("%s failed: %w: %s", filepath.Base(exe), err, strings.TrimSpace(output))
	}
	return output, nil
}
func newDatabaseSecret(path string, clusterExists bool) (databaseSecret, error) {
	var secret databaseSecret
	b, err := os.ReadFile(path)
	if err == nil {
		if err = json.Unmarshal(b, &secret); err != nil || len(secret.Password) < 32 {
			return secret, errors.New("invalid private database credentials; restore the matching private/database.json backup")
		}
		return secret, nil
	}
	if !os.IsNotExist(err) {
		return secret, err
	}
	if clusterExists {
		return secret, errors.New("database cluster exists without its private credentials; restore the matching private/database.json backup")
	}
	random := make([]byte, 32)
	if _, err = rand.Read(random); err != nil {
		return secret, err
	}
	secret.Password = hex.EncodeToString(random)
	return secret, writeJSONFile(path, secret)
}
func freePort() (int, error) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	err = listener.Close()
	return port, err
}
func databaseDSN(port int, password, database string) string {
	u := url.URL{Scheme: "postgres", User: url.UserPassword("scopeweaver", password), Host: net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), Path: "/" + database}
	q := url.Values{"sslmode": {"disable"}, "connect_timeout": {"3"}}
	u.RawQuery = q.Encode()
	return u.String()
}
func startPostgres(ctx context.Context, home string) (result *postgresProcess, retErr error) {
	tools, err := postgresBinaries()
	if err != nil {
		return nil, err
	}
	output, err := runPostgres(ctx, tools.postgres, "--version")
	if err != nil {
		return nil, err
	}
	major, err := postgresMajor(output)
	if err != nil {
		return nil, err
	}
	data := filepath.Join(home, "postgres")
	_, statErr := os.Stat(filepath.Join(data, "PG_VERSION"))
	clusterExists := statErr == nil
	if statErr != nil && !os.IsNotExist(statErr) {
		return nil, statErr
	}
	secret, err := newDatabaseSecret(filepath.Join(privateDir(home), "database.json"), clusterExists)
	if err != nil {
		return nil, err
	}
	if clusterExists {
		if err = validateClusterVersion(data, major); err != nil {
			return nil, err
		}
	} else {
		if err = os.MkdirAll(data, 0700); err != nil {
			return nil, err
		}
		if err = protectDir(data); err != nil {
			return nil, err
		}
		pwPath := filepath.Join(privateDir(home), "initdb-password")
		if err = os.WriteFile(pwPath, []byte(secret.Password+"\n"), 0600); err != nil {
			return nil, err
		}
		defer os.Remove(pwPath)
		if _, err = runPostgres(ctx, tools.initdb, "-D", data, "-U", "scopeweaver", "--pwfile="+pwPath,
			"--auth=scram-sha-256", "--encoding=UTF8", "--no-locale"); err != nil {
			return nil, err
		}
	}
	process := &postgresProcess{tools: tools, data: data}
	running, err := postgresRunning(ctx, tools, data)
	if err != nil {
		return nil, err
	}
	if running {
		port, err := postmasterPort(data)
		if err != nil {
			return nil, err
		}
		process.port = port
	} else {
		process.port, err = freePort()
		if err != nil {
			return nil, err
		}
		opts := fmt.Sprintf("-h 127.0.0.1 -p %d -c unix_socket_directories=", process.port)
		_, err = runPostgres(ctx, tools.ctl, "-D", data, "-l", filepath.Join(privateDir(home), "postgres.log"),
			"-w", "-t", "60", "-o", opts, "start")
		if err != nil {
			// pg_ctl can be cancelled after its child starts. Recover only this
			// credential-authenticated cluster; never signal a recorded PID.
			cleanCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			_ = stopOrphanPostgres(cleanCtx, home)
			return nil, err
		}
	}
	process.adminDSN = databaseDSN(process.port, secret.Password, "postgres")
	process.dsn = databaseDSN(process.port, secret.Password, "scopeweaver_cli")
	if err = process.verify(ctx); err != nil {
		return nil, err
	}
	defer func() {
		if retErr != nil {
			cleanCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			retErr = errors.Join(retErr, process.stop(cleanCtx))
		}
	}()
	conn, err := pgx.Connect(ctx, process.adminDSN)
	if err != nil {
		return nil, errors.New("cannot authenticate private PostgreSQL")
	}
	defer conn.Close(context.Background())
	var exists bool
	if err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname='scopeweaver_cli')").Scan(&exists); err != nil {
		return nil, fmt.Errorf("inspect private database: %w", err)
	}
	if !exists {
		if _, err = conn.Exec(ctx, "CREATE DATABASE scopeweaver_cli"); err != nil {
			return nil, fmt.Errorf("create private database: %w", err)
		}
	}
	return process, nil
}
func postgresRunning(ctx context.Context, tools postgresTools, data string) (bool, error) {
	cmd, err := newPostgresCommand(ctx, tools.ctl, "-D", data, "status")
	if err != nil {
		return false, err
	}
	output, err := cmd.CombinedOutput()
	if err == nil {
		return true, nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 3 {
		return false, nil
	}
	return false, fmt.Errorf("inspect private PostgreSQL: %w: %s", err, strings.TrimSpace(string(output)))
}
func postmasterPort(data string) (int, error) {
	b, err := os.ReadFile(filepath.Join(data, "postmaster.pid"))
	if err != nil {
		return 0, err
	}
	lines := strings.Split(string(b), "\n")
	if len(lines) < 4 {
		return 0, errors.New("private PostgreSQL has invalid ownership metadata")
	}
	port, err := strconv.Atoi(strings.TrimSpace(lines[3]))
	if err != nil || port < 1 || port > 65535 {
		return 0, errors.New("private PostgreSQL has invalid port metadata")
	}
	return port, nil
}
func (p *postgresProcess) verify(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, p.adminDSN)
	if err != nil {
		return errors.New("cannot authenticate private PostgreSQL; refusing to adopt or stop another database")
	}
	defer conn.Close(context.Background())
	var actual string
	if err = conn.QueryRow(ctx, "SHOW data_directory").Scan(&actual); err != nil {
		return fmt.Errorf("verify private PostgreSQL ownership: %w", err)
	}
	actual, err = filepath.Abs(actual)
	if err != nil {
		return err
	}
	expected, err := filepath.Abs(p.data)
	if err != nil {
		return err
	}
	matches := filepath.Clean(actual) == filepath.Clean(expected)
	if runtime.GOOS == "windows" {
		actualInfo, err := os.Stat(actual)
		if err != nil {
			return fmt.Errorf("verify PostgreSQL data directory: %w", err)
		}
		expectedInfo, err := os.Stat(expected)
		if err != nil {
			return err
		}
		matches = os.SameFile(actualInfo, expectedInfo)
	}
	if !matches {
		return errors.New("PostgreSQL data directory does not match this private runtime; refusing to adopt or stop it")
	}
	return nil
}
func (p *postgresProcess) stop(ctx context.Context) error {
	running, err := postgresRunning(ctx, p.tools, p.data)
	if err != nil {
		return err
	}
	if !running {
		return nil
	}
	if err = p.verify(ctx); err != nil {
		return err
	}
	_, err = runPostgres(ctx, p.tools.ctl, "-D", p.data, "-w", "-t", "60", "-m", "fast", "stop")
	return err
}
func stopOrphanPostgres(ctx context.Context, home string) error {
	data := filepath.Join(home, "postgres")
	if _, err := os.Stat(filepath.Join(data, "PG_VERSION")); os.IsNotExist(err) {
		return nil
	}
	tools, err := postgresBinaries()
	if err != nil {
		return err
	}
	running, err := postgresRunning(ctx, tools, data)
	if err != nil || !running {
		return err
	}
	secret, err := newDatabaseSecret(filepath.Join(privateDir(home), "database.json"), true)
	if err != nil {
		return err
	}
	port, err := postmasterPort(data)
	if err != nil {
		return err
	}
	return (&postgresProcess{tools: tools, data: data, port: port, adminDSN: databaseDSN(port, secret.Password, "postgres")}).stop(ctx)
}
