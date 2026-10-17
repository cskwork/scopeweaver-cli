package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cskwork/scopeweaver-cli/internal/portable"
	"github.com/cskwork/scopeweaver-cli/locale"
)

var Version = "0.1.0"

const maxJSON = 32 << 20

type options struct {
	home, lang, output   string
	json, yes, dry, help bool
	timeout              time.Duration
}
type invocation struct {
	operation Operation
	path      string
	query     url.Values
	body      map[string]any
	uploads   []string
}
type App struct {
	In       io.Reader
	Out, Err io.Writer
	Connect  func(context.Context, string) (*portable.Connection, error)
}

func New() *App { return &App{In: os.Stdin, Out: os.Stdout, Err: os.Stderr, Connect: portable.Ensure} }

func (a *App) Run(ctx context.Context, args []string) int {
	o, rest, err := globals(args)
	if err != nil {
		return a.fail(o, "invalid_input", err, 2)
	}
	language, _ := locale.Normalize(o.lang)
	locale.SetServerDefault(language)
	if len(rest) == 0 {
		return a.help(o, "")
	}
	if rest[0] == "commands" && len(rest) == 1 {
		return a.print(o, map[string]any{"version": Version, "operations": Operations()})
	}
	switch rest[0] {
	case "help":
		return a.help(o, strings.Join(rest[1:], " "))
	case "version":
		return a.print(o, map[string]any{"name": "scopeweaver-cli", "version": Version})
	case "schema":
		prefix := strings.Join(rest[1:], " ")
		ops := selectOperations(prefix)
		if len(ops) == 0 && prefix != "" {
			return a.fail(o, "unknown_command", fmt.Errorf("unknown command: %s", prefix), 2)
		}
		return a.print(o, map[string]any{"version": Version, "operations": ops})
	case "repl":
		if len(rest) != 1 {
			return a.fail(o, "invalid_input", errors.New("repl takes no arguments"), 2)
		}
		return a.repl(ctx, o)
	case "runtime":
		if o.help {
			return a.help(o, "runtime")
		}
		return a.runtime(ctx, o, rest[1:])
	case "doctor":
		return a.doctor(ctx, o)
	}
	var op Operation
	n := 0
	for _, candidate := range Operations() {
		parts := strings.Fields(candidate.Name)
		if len(parts) <= len(rest) && len(parts) > n && strings.Join(rest[:len(parts)], " ") == candidate.Name {
			op = candidate
			n = len(parts)
		}
	}
	if n == 0 {
		if o.help {
			return a.help(o, strings.Join(rest, " "))
		}
		return a.fail(o, "unknown_command", fmt.Errorf("unknown command: %s", strings.Join(rest, " ")), 2)
	}
	if o.help {
		return a.print(o, op)
	}
	request, err := a.parse(op, rest[n:])
	if err != nil {
		return a.fail(o, "invalid_input", err, 2)
	}
	if o.dry {
		return a.print(o, map[string]any{"command": op.Name, "method": op.Method, "path": request.path, "query": redactQuery(request.query), "body": redact(request.body), "uploads": request.uploads, "effect": op.Effect})
	}
	if op.Effect != "read" && !o.yes {
		return a.fail(o, "confirmation_required", errors.New(msg(o.lang, "confirm")), 2)
	}
	conn, err := a.Connect(ctx, o.home)
	if err != nil {
		return a.fail(o, "runtime_unavailable", err, 3)
	}
	if err = a.execute(ctx, o, conn, request); err != nil {
		code := "request_failed"
		exit := 1
		var he *HTTPError
		if errors.As(err, &he) {
			code = he.Code
			if he.Status == 401 || he.Status == 403 {
				exit = 4
			}
		}
		if errors.Is(err, context.Canceled) {
			code = "cancelled"
			exit = 130
		}
		return a.fail(o, code, err, exit)
	}
	return 0
}
func globals(args []string) (options, []string, error) {
	exe, err := os.Executable()
	if err != nil {
		return options{}, nil, err
	}
	home := os.Getenv("SCOPEWEAVER_CLI_HOME")
	if home == "" {
		home = filepath.Join(filepath.Dir(exe), "state")
	}
	o := options{home: home, lang: "en", timeout: 60 * time.Second}
	if s := os.Getenv("SCOPEWEAVER_LANGUAGE"); s != "" {
		o.lang = s
	}
	var rest []string
	for i := 0; i < len(args); i++ {
		key, value, has := strings.Cut(args[i], "=")
		switch key {
		case "--json":
			o.json = true
		case "--yes":
			o.yes = true
		case "--dry-run":
			o.dry = true
		case "--help", "-h":
			o.help = true
		case "--version":
			rest = append(rest, "version")
		case "--home", "--lang", "--output", "--timeout":
			if !has {
				i++
				if i >= len(args) {
					return o, nil, fmt.Errorf("%s requires a value", key)
				}
				value = args[i]
			}
			switch key {
			case "--home":
				o.home = value
			case "--lang":
				o.lang = value
			case "--output":
				o.output = value
			case "--timeout":
				d, e := time.ParseDuration(value)
				if e != nil || d <= 0 {
					return o, nil, errors.New("--timeout requires a positive duration, for example 60s")
				}
				o.timeout = d
			}
		default:
			rest = append(rest, args[i])
		}
	}
	if o.lang != "en" && o.lang != "ko" && o.lang != "zh-Hans" {
		return o, nil, errors.New("--lang must be en, ko or zh-Hans")
	}
	o.home, err = filepath.Abs(o.home)
	return o, rest, err
}
func selectOperations(prefix string) []Operation {
	result := []Operation{}
	for _, op := range Operations() {
		if prefix == "" || op.Name == prefix || strings.HasPrefix(op.Name, prefix+" ") {
			result = append(result, op)
		}
	}
	return result
}
func (a *App) help(o options, prefix string) int {
	if o.json {
		return a.print(o, map[string]any{"version": Version, "commands": selectOperations(prefix), "builtins": []string{"commands", "schema", "doctor", "runtime start", "runtime status", "runtime stop", "repl", "version"}, "global_options": []string{"--home PATH", "--lang en|ko|zh-Hans", "--json", "--yes", "--dry-run", "--timeout 60s", "--output PATH"}})
	}
	fmt.Fprintln(a.Out, "scopeweaver-cli "+Version)
	fmt.Fprintln(a.Out, msg(o.lang, "usage"))
	fmt.Fprintln(a.Out, "  scopeweaver-cli <group> <action> [--parameter value] [--json]")
	fmt.Fprintln(a.Out, "  scopeweaver-cli schema <group> | doctor | runtime start|status|stop | repl")
	fmt.Fprintln(a.Out, "  --data @file.json | --data - | --field key=value | --query key=value")
	fmt.Fprintln(a.Out, "  --upload FILE (repeatable) | --output FILE | --yes | --dry-run")
	fmt.Fprintln(a.Out, msg(o.lang, "writes"))
	if prefix == "" {
		groups := map[string]bool{}
		for _, op := range Operations() {
			groups[strings.Fields(op.Name)[0]] = true
		}
		names := []string{}
		for k := range groups {
			names = append(names, k)
		}
		sort.Strings(names)
		fmt.Fprintln(a.Out, strings.Join(names, "  "))
	} else {
		ops := selectOperations(prefix)
		if len(ops) == 0 && prefix != "runtime" {
			return a.fail(o, "unknown_command", fmt.Errorf("unknown group: %s", prefix), 2)
		}
		for _, op := range ops {
			fmt.Fprintf(a.Out, "  %-48s %s %s [%s]\n", op.Name, op.Method, op.Path, op.Effect)
		}
	}
	return 0
}
func (a *App) parse(op Operation, args []string) (invocation, error) {
	r := invocation{operation: op, path: op.Path, query: url.Values{}, body: map[string]any{}}
	values := map[string]string{}
	dataSet := false
	params := map[string]Parameter{}
	for _, p := range op.Parameters {
		params[p.Name] = p
		params[strings.ReplaceAll(p.Name, "_", "-")] = p
	}
	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "--") {
			return r, fmt.Errorf("expected a named option, got %q", args[i])
		}
		key, value, has := strings.Cut(strings.TrimPrefix(args[i], "--"), "=")
		if !has {
			if p, ok := params[key]; ok && p.Type == "boolean" && (i+1 == len(args) || strings.HasPrefix(args[i+1], "--")) {
				value = "true"
			} else {
				i++
				if i >= len(args) {
					return r, fmt.Errorf("--%s requires a value", key)
				}
				value = args[i]
			}
		}
		switch key {
		case "data":
			if dataSet {
				return r, errors.New("--data may only be supplied once")
			}
			dataSet = true
			raw := []byte(value)
			var e error
			if value == "-" {
				raw, e = io.ReadAll(io.LimitReader(a.In, maxJSON+1))
			} else if strings.HasPrefix(value, "@") {
				file, openErr := os.Open(value[1:])
				if openErr != nil {
					e = openErr
				} else {
					raw, e = io.ReadAll(io.LimitReader(file, maxJSON+1))
					closeErr := file.Close()
					if e == nil {
						e = closeErr
					}
				}
			}
			if e != nil {
				return r, e
			}
			if len(raw) > maxJSON {
				return r, errors.New("input exceeds 32 MiB")
			}
			var data map[string]any
			if e = json.Unmarshal(raw, &data); e != nil || data == nil {
				return r, errors.New("--data must be a JSON object")
			}
			for k, v := range data {
				r.body[k] = v
			}
		case "field":
			k, v, ok := strings.Cut(value, "=")
			if !ok || k == "" {
				return r, errors.New("--field requires key=value")
			}
			var parsed any
			if json.Unmarshal([]byte(v), &parsed) != nil {
				parsed = v
			}
			r.body[k] = parsed
		case "query":
			k, v, ok := strings.Cut(value, "=")
			if !ok || k == "" {
				return r, errors.New("--query requires key=value")
			}
			r.query.Add(k, v)
		case "upload":
			r.uploads = append(r.uploads, value)
		default:
			p, ok := params[key]
			if !ok {
				return r, fmt.Errorf("unknown option --%s; inspect schema %s", key, op.Name)
			}
			values[p.Name] = value
		}
	}
	for _, p := range op.Parameters {
		v, ok := values[p.Name]
		if !ok {
			if p.Required {
				if p.In == "body" {
					if _, present := r.body[p.Name]; present {
						continue
					}
				}
				if p.In == "query" && r.query.Has(p.Name) {
					continue
				}
				return r, fmt.Errorf("missing --%s", strings.ReplaceAll(p.Name, "_", "-"))
			}
			continue
		}
		switch p.In {
		case "path":
			escaped := url.PathEscape(v)
			if strings.Contains(r.path, "{"+p.Name+"...}") {
				parts := strings.Split(v, "/")
				for i := range parts {
					parts[i] = url.PathEscape(parts[i])
				}
				escaped = strings.Join(parts, "/")
				r.path = strings.ReplaceAll(r.path, "{"+p.Name+"...}", escaped)
			}
			r.path = strings.ReplaceAll(r.path, "{"+p.Name+"}", escaped)
		case "query":
			r.query.Set(p.Name, v)
		case "body":
			parsed, e := typed(v, p.Type)
			if e != nil {
				return r, fmt.Errorf("--%s: %w", p.Name, e)
			}
			r.body[p.Name] = parsed
		}
	}
	if strings.Contains(r.path, "{") {
		return r, fmt.Errorf("unresolved path parameter: %s", r.path)
	}
	if len(r.body) > 0 && !op.Body {
		return r, errors.New("this command does not accept a JSON body")
	}
	if len(r.uploads) > 0 && op.Multipart == "" {
		return r, errors.New("this command does not accept uploads")
	}
	if op.Multipart != "" && len(r.uploads) == 0 {
		return r, errors.New("this command requires --upload FILE")
	}
	return r, nil
}
func typed(value, kind string) (any, error) {
	switch kind {
	case "integer":
		return strconv.ParseInt(value, 10, 64)
	case "number":
		return strconv.ParseFloat(value, 64)
	case "boolean":
		if value != "true" && value != "false" {
			return nil, errors.New("expected true or false")
		}
		return value == "true", nil
	case "object", "array":
		var v any
		if err := json.Unmarshal([]byte(value), &v); err != nil {
			return nil, err
		}
		if kind == "object" {
			if _, ok := v.(map[string]any); !ok {
				return nil, errors.New("expected a JSON object")
			}
		}
		if kind == "array" {
			if _, ok := v.([]any); !ok {
				return nil, errors.New("expected a JSON array")
			}
		}
		return v, nil
	default:
		return value, nil
	}
}

type HTTPError struct {
	Status        int
	Code, Message string
}

func (e *HTTPError) Error() string { return fmt.Sprintf("HTTP %d: %s", e.Status, e.Message) }
func (a *App) execute(ctx context.Context, o options, c *portable.Connection, r invocation) error {
	base, err := url.Parse(c.URL)
	if err != nil {
		return err
	}
	if base.Scheme != "http" || (base.Hostname() != "127.0.0.1" && base.Hostname() != "::1") {
		return errors.New("runtime must use a loopback HTTP origin")
	}
	targetURL, err := url.Parse(strings.TrimRight(c.URL, "/") + r.path)
	if err != nil {
		return err
	}
	query := targetURL.Query()
	for key, values := range r.query {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	targetURL.RawQuery = query.Encode()
	target := targetURL.String()
	var body io.Reader
	var closeUpload func()
	contentType := "application/json"
	if r.operation.Multipart != "" {
		reader, writer := io.Pipe()
		mw := multipart.NewWriter(writer)
		contentType = mw.FormDataContentType()
		body = reader
		uploadCtx, cancel := context.WithCancel(ctx)
		closeUpload = func() { cancel(); reader.Close() }
		go func() {
			var writeErr error
			defer func() {
				if writeErr == nil {
					writeErr = mw.Close()
				}
				writer.CloseWithError(writeErr)
			}()
			for _, path := range r.uploads {
				if uploadCtx.Err() != nil {
					writeErr = uploadCtx.Err()
					return
				}
				f, e := os.Open(path)
				if e != nil {
					writeErr = e
					return
				}
				part, e := mw.CreateFormFile(r.operation.Multipart, filepath.Base(path))
				if e == nil {
					_, e = io.Copy(part, f)
				}
				f.Close()
				if e != nil {
					writeErr = e
					return
				}
			}
		}()
	} else if r.operation.Body {
		raw, e := json.Marshal(r.body)
		if e != nil {
			return e
		}
		body = bytes.NewReader(raw)
	}
	if closeUpload != nil {
		defer closeUpload()
	}
	requestCtx := ctx
	var cancel context.CancelFunc
	if r.operation.Response != "sse" {
		requestCtx, cancel = context.WithTimeout(ctx, o.timeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(requestCtx, r.operation.Method, target, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept-Language", o.lang)
	if body != nil {
		req.Header.Set("Content-Type", contentType)
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.ResponseHeaderTimeout = o.timeout
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 65536))
		message := strings.TrimSpace(string(raw))
		var data map[string]any
		if json.Unmarshal(raw, &data) == nil {
			if s, ok := data["error"].(string); ok {
				message = s
			}
		}
		message = strings.ReplaceAll(message, c.Token, "[REDACTED]")
		return &HTTPError{resp.StatusCode, "backend_error", redactMessage(message, r.body)}
	}
	switch r.operation.Response {
	case "sse":
		return streamEvents(requestCtx, resp.Body, a.Out)
	case "download":
		if o.output == "" {
			return errors.New(msg(o.lang, "output"))
		}
		return a.download(o, resp.Body)
	case "text":
		raw, e := io.ReadAll(io.LimitReader(resp.Body, maxJSON+1))
		if e != nil {
			return e
		}
		if len(raw) > maxJSON {
			return errors.New("response exceeds 32 MiB")
		}
		if o.output != "" {
			return a.download(o, bytes.NewReader(raw))
		}
		if o.json {
			a.print(o, map[string]any{"data": string(raw)})
			return nil
		}
		_, e = fmt.Fprint(a.Out, string(raw))
		return e
	default:
		if resp.StatusCode == 204 {
			a.print(o, map[string]any{"ok": true})
			return nil
		}
		raw, e := io.ReadAll(io.LimitReader(resp.Body, maxJSON+1))
		if e != nil {
			return e
		}
		if len(raw) > maxJSON {
			return errors.New("response exceeds 32 MiB; narrow the query")
		}
		var data any
		if e = json.Unmarshal(raw, &data); e != nil {
			return fmt.Errorf("backend returned invalid JSON: %w", e)
		}
		if a.print(o, data) != 0 {
			return errors.New("write output failed")
		}
		return nil
	}
}
func (a *App) download(o options, body io.Reader) error {
	target, err := filepath.Abs(o.output)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		f.Close()
		if !ok {
			os.Remove(target)
		}
	}()
	n, err := io.Copy(f, body)
	if err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	ok = true
	a.print(o, map[string]any{"output": target, "bytes": n})
	return nil
}
func streamEvents(ctx context.Context, reader io.Reader, out io.Writer) error {
	scan := bufio.NewScanner(reader)
	scan.Buffer(make([]byte, 64*1024), 4<<20)
	var event, id string
	var data []string
	emit := func() error {
		if len(data) == 0 {
			event = ""
			return nil
		}
		raw := strings.Join(data, "\n")
		var payload any
		if json.Unmarshal([]byte(raw), &payload) != nil {
			payload = raw
		}
		if event == "" {
			event = "message"
		}
		err := json.NewEncoder(out).Encode(map[string]any{"event": event, "id": id, "data": payload})
		event = ""
		data = nil
		return err
	}
	for scan.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		line := scan.Text()
		if line == "" {
			if err := emit(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		key, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")
		switch key {
		case "event":
			event = value
		case "id":
			id = value
		case "data":
			data = append(data, value)
		}
	}
	if err := scan.Err(); err != nil {
		return err
	}
	return emit()
}
func (a *App) runtime(ctx context.Context, o options, args []string) int {
	if len(args) != 1 {
		return a.fail(o, "invalid_input", errors.New("runtime requires start, status or stop"), 2)
	}
	if args[0] != "start" && args[0] != "status" && args[0] != "stop" {
		return a.fail(o, "invalid_input", errors.New("runtime requires start, status or stop"), 2)
	}
	if o.dry {
		return a.print(o, map[string]any{"command": "runtime " + args[0], "home": o.home, "dry_run": true})
	}
	switch args[0] {
	case "start":
		c, e := a.Connect(ctx, o.home)
		if e != nil {
			return a.fail(o, "runtime_unavailable", e, 3)
		}
		return a.print(o, map[string]any{"running": true, "home": o.home, "pid": c.PID})
	case "status":
		c, e := portable.Status(ctx, o.home)
		if e != nil {
			return a.print(o, map[string]any{"running": false, "home": o.home, "reason": e.Error()})
		}
		return a.print(o, map[string]any{"running": true, "home": o.home, "pid": c.PID})
	case "stop":
		if !o.yes {
			return a.fail(o, "confirmation_required", errors.New(msg(o.lang, "confirm")), 2)
		}
		if e := portable.Stop(ctx, o.home); e != nil {
			return a.fail(o, "runtime_error", e, 3)
		}
		return a.print(o, map[string]any{"running": false, "home": o.home})
	default:
		return a.fail(o, "invalid_input", errors.New("runtime requires start, status or stop"), 2)
	}
}
func (a *App) doctor(ctx context.Context, o options) int {
	exe, _ := os.Executable()
	root := filepath.Dir(exe)
	missing := []string{}
	suffix := ""
	if strings.HasSuffix(strings.ToLower(exe), ".exe") {
		suffix = ".exe"
	}
	for _, name := range []string{"postgres", "initdb", "pg_ctl"} {
		if _, e := os.Stat(filepath.Join(root, "runtime", "postgres", "bin", name+suffix)); e != nil {
			missing = append(missing, name)
		}
	}
	c, err := portable.Status(ctx, o.home)
	result := map[string]any{"version": Version, "home": o.home, "portable_root": root, "missing_runtime_files": missing, "running": err == nil, "external_requirements": []string{"Configured LLM provider or local inference service for AI operations", "Executables/services selected by custom tools and MCP integrations"}}
	if c != nil {
		result["pid"] = c.PID
	}
	a.print(o, result)
	if len(missing) > 0 {
		return 3
	}
	return 0
}
func (a *App) repl(ctx context.Context, o options) int {
	reader := bufio.NewReader(a.In)
	if !o.json {
		fmt.Fprintln(a.Err, msg(o.lang, "repl"))
	}
	exitCode := 0
	for {
		line, readErr := reader.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			return a.fail(o, "input_error", readErr, 1)
		}
		if len(line) > maxJSON {
			return a.fail(o, "invalid_input", errors.New("input exceeds 32 MiB"), 2)
		}
		line = strings.TrimSpace(line)
		if line == "exit" || line == "quit" {
			return exitCode
		}
		if line != "" {
			args, err := splitLine(line)
			if err != nil {
				exitCode = a.fail(o, "invalid_input", err, 2)
			} else {
				for i := 0; i < len(args); i++ {
					stdinArg := args[i] == "--data=-" || (args[i] == "--data" && i+1 < len(args) && args[i+1] == "-")
					if !stdinArg {
						continue
					}
					decoder := json.NewDecoder(io.LimitReader(reader, maxJSON+1))
					var raw json.RawMessage
					err = decoder.Decode(&raw)
					reader = bufio.NewReader(io.MultiReader(decoder.Buffered(), reader))
					if err != nil {
						return a.fail(o, "invalid_input", fmt.Errorf("stdin JSON: %w", err), 2)
					}
					if len(raw) > maxJSON {
						return a.fail(o, "invalid_input", errors.New("input exceeds 32 MiB"), 2)
					}
					if args[i] == "--data=-" {
						args[i] = "--data=" + string(raw)
					} else {
						args[i+1] = string(raw)
						i++
					}
				}
				defaults := []string{"--home", o.home, "--lang", o.lang, "--timeout", o.timeout.String()}
				if o.json {
					defaults = append(defaults, "--json")
				}
				if o.dry {
					defaults = append(defaults, "--dry-run")
				}
				if code := a.Run(ctx, append(defaults, args...)); code != 0 {
					exitCode = code
				}
			}
		}
		if ctx.Err() != nil {
			return 130
		}
		if readErr == io.EOF {
			return exitCode
		}
	}
}
func splitLine(s string) ([]string, error) {
	var result []string
	var word strings.Builder
	var quote rune
	active := false
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		if quote != 0 {
			if c == '\\' && quote == '"' && i+1 < len(runes) && (runes[i+1] == '"' || runes[i+1] == '\\') {
				i++
				word.WriteRune(runes[i])
				active = true
				continue
			}
			if c == quote {
				quote = 0
			} else {
				word.WriteRune(c)
			}
			active = true
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			active = true
			continue
		}
		if c == ' ' || c == '\t' {
			if active {
				result = append(result, word.String())
				word.Reset()
				active = false
			}
			continue
		}
		word.WriteRune(c)
		active = true
	}
	if quote != 0 {
		return nil, errors.New("unterminated quoted argument")
	}
	if active {
		result = append(result, word.String())
	}
	return result, nil
}

var secretKey = regexp.MustCompile("(?i)(password|api[_-]?key|token|secret|authorization|credential)")

func redact(v any) any {
	switch value := v.(type) {
	case map[string]any:
		r := map[string]any{}
		for k, item := range value {
			if secretKey.MatchString(k) {
				r[k] = "[REDACTED]"
			} else {
				r[k] = redact(item)
			}
		}
		return r
	case []any:
		r := make([]any, len(value))
		for i, item := range value {
			r[i] = redact(item)
		}
		return r
	default:
		return value
	}
}
func (a *App) print(o options, value any) int {
	e := json.NewEncoder(a.Out)
	e.SetEscapeHTML(false)
	if !o.json {
		e.SetIndent("", "  ")
	}
	if err := e.Encode(value); err != nil {
		return 1
	}
	return 0
}
func (a *App) fail(o options, code string, err error, exit int) int {
	if o.json {
		json.NewEncoder(a.Err).Encode(map[string]any{"error": map[string]any{"code": code, "message": errorSummary(o.lang, code), "detail": err.Error()}})
	} else {
		fmt.Fprintln(a.Err, errorSummary(o.lang, code)+": "+err.Error())
	}
	return exit
}
func msg(lang, key string) string {
	messages := map[string][3]string{
		"usage":   {"Usage", "사용법", "用法"},
		"writes":  {"State-changing commands require --yes. Inspect --dry-run first.", "상태를 변경하려면 --yes가 필요합니다. 먼저 --dry-run으로 확인하세요.", "修改状态的命令需要 --yes。请先使用 --dry-run 检查。"},
		"confirm": {"This operation changes state or executes work; pass --yes to confirm.", "이 작업은 상태를 변경하거나 실행을 시작합니다. 확인하려면 --yes를 지정하세요.", "此操作会修改状态或执行任务；请使用 --yes 确认。"},
		"output":  {"Download commands require --output PATH; existing files are never overwritten.", "다운로드하려면 --output PATH가 필요합니다. 기존 파일은 덮어쓰지 않습니다.", "下载命令需要 --output PATH；不会覆盖现有文件。"},
		"error":   {"Error", "오류", "错误"},
		"repl":    {"Interactive mode. Enter help, or exit. Commands use shell-style quotes; no shell is executed.", "대화형 모드입니다. help 또는 exit를 입력하세요. 셸 형식 따옴표를 지원하며 셸을 실행하지 않습니다.", "交互模式。输入 help 或 exit。支持 shell 风格引号，但不会执行 shell。"},
	}
	v := messages[key]
	switch lang {
	case "ko":
		return v[1]
	case "zh-Hans":
		return v[2]
	default:
		return v[0]
	}
}
