package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
)

type registeredRoute struct{ method, path, handler string }
type catalogSource struct {
	files     []*ast.File
	functions map[string]*ast.FuncDecl
	routes    []registeredRoute
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate catalog source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func expressionName(e ast.Expr) string {
	switch e := e.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return e.Sel.Name
	case *ast.CallExpr:
		return expressionName(e.Fun)
	}
	return ""
}

func literalString(e ast.Expr) (string, bool) {
	v, ok := e.(*ast.BasicLit)
	if !ok || v.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(v.Value)
	return s, err == nil
}

func stringExpression(e ast.Expr, values map[string]string) (string, bool) {
	if s, ok := literalString(e); ok {
		return s, true
	}
	switch e := e.(type) {
	case *ast.Ident:
		s, ok := values[e.Name]
		return s, ok
	case *ast.BinaryExpr:
		if e.Op == token.ADD {
			a, aOK := stringExpression(e.X, values)
			b, bOK := stringExpression(e.Y, values)
			return a + b, aOK && bOK
		}
	}
	return "", false
}

func readCatalogSource(t *testing.T) catalogSource {
	t.Helper()
	source := catalogSource{functions: map[string]*ast.FuncDecl{}}
	files, err := filepath.Glob(filepath.Join(repositoryRoot(t), "server", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		source.files = append(source.files, file)
		for _, declaration := range file.Decls {
			fn, ok := declaration.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			source.functions[fn.Name.Name] = fn
			values := map[string]string{}
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				if assignment, ok := node.(*ast.AssignStmt); ok {
					for i, right := range assignment.Rhs {
						if i < len(assignment.Lhs) {
							if value, ok := stringExpression(right, values); ok {
								values[expressionName(assignment.Lhs[i])] = value
							}
						}
					}
				}
				call, ok := node.(*ast.CallExpr)
				if !ok || expressionName(call.Fun) != "HandleFunc" || len(call.Args) != 2 {
					return true
				}
				pattern, ok := stringExpression(call.Args[0], values)
				if !ok {
					if fn.Name.Name != "registerSideRoutes" {
						t.Errorf("unhandled route expression in %s; update catalog coverage extraction", fn.Name.Name)
					}
					return true
				}
				method, path, hasMethod := strings.Cut(pattern, " ")
				if !hasMethod {
					path, method = method, ""
				}
				source.routes = append(source.routes, registeredRoute{method, path, expressionName(call.Args[1])})
				return true
			})
		}
	}
	// Expand the actual parent/method literals, rather than freezing a side-route count.
	side := source.functions["registerSideRoutes"]
	if side == nil {
		t.Fatal("side-question route registrar missing")
	}
	var parents, methods []string
	ast.Inspect(side.Body, func(node ast.Node) bool {
		literal, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}
		array, ok := literal.Type.(*ast.ArrayType)
		if !ok {
			return true
		}
		if element, ok := array.Elt.(*ast.Ident); ok && element.Name == "string" {
			for _, item := range literal.Elts {
				if method, ok := literalString(item); ok {
					methods = append(methods, method)
				}
			}
		}
		if _, ok := array.Elt.(*ast.StructType); ok {
			for _, item := range literal.Elts {
				item, ok := item.(*ast.CompositeLit)
				if ok && len(item.Elts) > 0 {
					if parent, ok := literalString(item.Elts[0]); ok {
						parents = append(parents, parent)
					}
				}
			}
		}
		return true
	})
	if len(parents) == 0 || len(methods) == 0 {
		t.Fatal("cannot expand side-question route registration")
	}
	for _, parent := range parents {
		for _, method := range methods {
			source.routes = append(source.routes, registeredRoute{method, parent + "/side-questions", "handleSideQuestions"})
		}
	}
	return source
}

func operationPath(t *testing.T, operation Operation) string {
	t.Helper()
	parsed, err := url.Parse(operation.Path)
	if err != nil {
		t.Fatal(err)
	}
	return parsed.Path
}

func routeMatches(route registeredRoute, method, path string) bool {
	if route.method != "" && route.method != method {
		return false
	}
	if route.path == path {
		return true
	}
	if strings.HasSuffix(route.path, "/{rest...}") {
		return strings.HasPrefix(path, strings.TrimSuffix(route.path, "{rest...}"))
	}
	return false
}

func TestCatalogCoversEveryRegisteredRoute(t *testing.T) {
	source := readCatalogSource(t)
	operations := Operations()
	for _, route := range source.routes {
		found := false
		for _, operation := range operations {
			if routeMatches(route, operation.Method, operationPath(t, operation)) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("registered route has no named command: %s %s (%s)", route.method, route.path, route.handler)
		}
	}
	for _, operation := range operations {
		found := false
		for _, route := range source.routes {
			if routeMatches(route, operation.Method, operationPath(t, operation)) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("command has no matching backend route: %s: %s %s", operation.Name, operation.Method, operation.Path)
		}
	}
}

func TestCatalogExpandsEveryChatControlAction(t *testing.T) {
	source := readCatalogSource(t)
	expected := []string{
		"GET ", "GET /turns/{turn_id}/prompt", "GET /attempts", "GET /attempts/{attempt_id}/activities", "GET /context",
		"POST /compact", "POST /retry", "POST /edit", "POST /branch", "POST /inputs", "PATCH /inputs/{input_id}", "DELETE /inputs/{input_id}",
		"POST /inputs/{input_id}/followup", "POST /inputs/{input_id}/discard", "POST /pause", "POST /resume", "POST /queue/order",
	}
	for _, parent := range []string{"/api/conversations/{id}/controls", "/api/tasks/{id}/chat/controls"} {
		var actual []string
		for _, operation := range Operations() {
			path := operationPath(t, operation)
			if path == parent || strings.HasPrefix(path, parent+"/") {
				actual = append(actual, operation.Method+" "+strings.TrimPrefix(path, parent))
				if strings.HasPrefix(parent, "/api/tasks/") {
					assertParameter(t, operation, "query", "seg", true)
				}
			}
		}
		slices.Sort(actual)
		want := slices.Clone(expected)
		slices.Sort(want)
		if !reflect.DeepEqual(actual, want) {
			t.Errorf("chat control matrix differs for %s\ngot %v\nwant %v", parent, actual, want)
		}
	}
	// New dispatcher branches must update the matrix, even though a wildcard still matches.
	for fnName, want := range map[string][]string{
		"chatControls":     {"", "turns", "prompt", "attempts", "compact", "context", "retry", "edit", "branch", "inputs", "followup", "discard", "pause", "resume", "queue/order"},
		"chatAttemptsHTTP": {"activities"},
	} {
		found := map[string]bool{}
		fn := source.functions[fnName]
		if fn == nil {
			t.Fatalf("missing dispatcher %s", fnName)
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			comparison, ok := node.(*ast.BinaryExpr)
			if !ok || (comparison.Op != token.EQL && comparison.Op != token.NEQ) {
				return true
			}
			for _, pair := range [][2]ast.Expr{{comparison.X, comparison.Y}, {comparison.Y, comparison.X}} {
				if text, ok := literalString(pair[1]); ok {
					isRoute := expressionName(pair[0]) == "rest"
					if index, ok := pair[0].(*ast.IndexExpr); ok && expressionName(index.X) == "parts" {
						isRoute = true
					}
					if isRoute {
						found[text] = true
					}
				}
			}
			return true
		})
		var actual []string
		for s := range found {
			actual = append(actual, s)
		}
		slices.Sort(actual)
		slices.Sort(want)
		if !reflect.DeepEqual(actual, want) {
			t.Errorf("%s routing changed; review expanded commands: got %v want %v", fnName, actual, want)
		}
	}
}

func assertParameter(t *testing.T, operation Operation, in, name string, required bool) {
	t.Helper()
	for _, parameter := range operation.Parameters {
		if parameter.In == in && parameter.Name == name {
			if required && !parameter.Required {
				t.Errorf("%s needs required %s parameter %s", operation.Name, in, name)
			}
			return
		}
	}
	t.Errorf("%s missing %s parameter %s", operation.Name, in, name)
}

func sourceQueryFields(source catalogSource, handler string, seen map[string]bool) map[string]bool {
	out := map[string]bool{}
	if seen[handler] {
		return out
	}
	seen[handler] = true
	fn := source.functions[handler]
	if fn == nil {
		return out
	}
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		name := expressionName(call.Fun)
		if name == "Get" && len(call.Args) == 1 {
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if ok {
				receiver := expressionName(selector.X)
				if receiver == "q" || receiver == "query" || receiver == "Query" {
					if field, ok := literalString(call.Args[0]); ok {
						out[field] = true
					}
				}
			}
		}
		if name == "queryInt" && len(call.Args) >= 2 {
			if field, ok := literalString(call.Args[1]); ok {
				out[field] = true
			}
		}
		if name != "extractToken" && name != "verifyJWT" && source.functions[name] != nil {
			requestPassed := false
			ast.Inspect(call, func(node ast.Node) bool {
				if id, ok := node.(*ast.Ident); ok && (id.Name == "r" || id.Name == "q" || id.Name == "query") {
					requestPassed = true
				}
				return true
			})
			if requestPassed {
				for field := range sourceQueryFields(source, name, seen) {
					out[field] = true
				}
			}
		}
		return true
	})
	return out
}

func TestCatalogNamesEveryLiteralQueryOption(t *testing.T) {
	source := readCatalogSource(t)
	for _, operation := range Operations() {
		var handler string
		for _, route := range source.routes {
			if route.path == operationPath(t, operation) && route.method == operation.Method {
				handler = route.handler
				break
			}
		}
		if handler == "" || handler == "handleSideQuestions" {
			continue
		}
		queries := sourceQueryFields(source, handler, map[string]bool{})
		fixed, _ := url.Parse(operation.Path)
		for name := range queries {
			if !fixed.Query().Has(name) {
				assertParameter(t, operation, "query", name, false)
			}
		}
	}
}

func TestCatalogMetadataAndConsequentialOptions(t *testing.T) {
	operations := Operations()
	byName := map[string]Operation{}
	placeholder := regexp.MustCompile(`\{([^}]+)\}`)
	for _, operation := range operations {
		if _, exists := byName[operation.Name]; exists {
			t.Errorf("duplicate command %s", operation.Name)
		}
		byName[operation.Name] = operation
		if operation.Description == "" || operation.Source == "" {
			t.Errorf("missing provenance/help for %s", operation.Name)
		}
		if !slices.Contains([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}, operation.Method) {
			t.Errorf("invalid method for %s", operation.Name)
		}
		if !slices.Contains([]string{"json", "text", "download", "sse"}, operation.Response) {
			t.Errorf("invalid response for %s", operation.Name)
		}
		if !slices.Contains([]string{"read", "write", "execute", "destructive", "admin"}, operation.Effect) {
			t.Errorf("invalid effect for %s", operation.Name)
		}
		if operation.Method != "GET" && operation.Effect == "read" {
			t.Errorf("mutation mislabeled as read: %s", operation.Name)
		}
		if operation.Multipart != "" && operation.Multipart != "file" && operation.Multipart != "files" {
			t.Errorf("invalid multipart field for %s", operation.Name)
		}
		for _, match := range placeholder.FindAllStringSubmatch(operationPath(t, operation), -1) {
			assertParameter(t, operation, "path", strings.TrimSuffix(match[1], "..."), true)
		}
		seen := map[string]bool{}
		for _, parameter := range operation.Parameters {
			key := parameter.In + ":" + parameter.Name
			if seen[key] {
				t.Errorf("duplicate parameter %s in %s", key, operation.Name)
			}
			seen[key] = true
			if !slices.Contains([]string{"path", "query", "body"}, parameter.In) || !slices.Contains([]string{"string", "integer", "number", "boolean", "object", "array"}, parameter.Type) {
				t.Errorf("invalid parameter %s in %s", key, operation.Name)
			}
			if parameter.In == "body" && !operation.Body {
				t.Errorf("body field without body support in %s", operation.Name)
			}
		}
		sourceFile, line, ok := strings.Cut(operation.Source, ":")
		if !ok {
			t.Errorf("invalid source %s", operation.Source)
			continue
		}
		contents, err := os.ReadFile(filepath.Join(repositoryRoot(t), filepath.FromSlash(sourceFile)))
		if err != nil {
			t.Error(err)
			continue
		}
		number, err := strconv.Atoi(line)
		if err != nil || number < 1 || number > len(strings.Split(string(contents), "\n")) {
			t.Errorf("invalid source line %s", operation.Source)
		}
	}
	for name, response := range map[string]string{"report get": "text", "findings export": "download", "assets export": "download", "workspace download": "download", "traffic blob download": "download", "findings traffic body": "json", "findings traffic download": "download", "logs watch": "sse", "exploration activity watch": "sse", "side-questions watch": "sse", "update watch": "sse"} {
		operation, ok := byName[name]
		if !ok || operation.Response != response {
			t.Errorf("%s must have response %s", name, response)
		}
	}
	for _, name := range []string{"sync scopesentry status", "sync scopesentry projects list", "sync scopesentry tasks list", "tools custom test", "intercept pending decide", "tasks create", "tasks intents control", "traffic purge", "update apply"} {
		operation, ok := byName[name]
		if !ok || operation.Effect == "read" {
			t.Errorf("%s must declare its consequences", name)
		}
	}
	for _, field := range []string{"category_id", "llm_profile_ids", "source_task_ids", "company_ids", "timeout_seconds", "plan_heartbeat_seconds", "seed_first_intent", "coverage_enabled", "intercept_rules"} {
		assertParameter(t, byName["tasks create"], "body", field, false)
	}
	for _, field := range []string{"delete_assets", "delete_traffic", "delete_files", "delete_findings", "delete_llm_records"} {
		assertParameter(t, byName["tasks delete"], "body", field, false)
	}
	for _, name := range []string{"conversations controls inputs send", "tasks chat controls inputs send"} {
		assertParameter(t, byName[name], "body", "client_request_id", true)
	}
	for _, name := range []string{"conversations controls inputs update", "tasks chat controls queue order"} {
		assertParameter(t, byName[name], "body", "expected_revision", true)
	}
	for _, name := range []string{"workspace upload", "chat upload", "skills upload"} {
		if byName[name].Multipart != "file" {
			t.Errorf("%s must send the actual multipart field file", name)
		}
	}
}

// Changes to request structs should expose their new fields as named CLI flags.
// Parse the actual decoder targets; response structs and database rows do not
// become invented command arguments.
func TestCatalogNamesDecodedBodyFields(t *testing.T) {
	source := readCatalogSource(t)
	types := map[string]ast.Expr{}
	for _, pkg := range []string{"server", "db", "chatstate", "intercept"} {
		files, err := filepath.Glob(filepath.Join(repositoryRoot(t), pkg, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range files {
			if strings.HasSuffix(name, "_test.go") {
				continue
			}
			file, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, declaration := range file.Decls {
				declaration, ok := declaration.(*ast.GenDecl)
				if !ok {
					continue
				}
				for _, spec := range declaration.Specs {
					if spec, ok := spec.(*ast.TypeSpec); ok {
						types[pkg+"."+spec.Name.Name] = spec.Type
					}
				}
			}
		}
	}
	var fields func(ast.Expr, string, map[string]bool) []string
	fields = func(expression ast.Expr, pkg string, seen map[string]bool) []string {
		switch expression := expression.(type) {
		case *ast.StarExpr:
			return fields(expression.X, pkg, seen)
		case *ast.Ident:
			key := pkg + "." + expression.Name
			if seen[key] {
				return nil
			}
			seen[key] = true
			return fields(types[key], pkg, seen)
		case *ast.SelectorExpr:
			return fields(expression.Sel, expressionName(expression.X), seen)
		case *ast.StructType:
			var result []string
			for _, field := range expression.Fields.List {
				if len(field.Names) == 0 {
					result = append(result, fields(field.Type, pkg, seen)...)
					continue
				}
				tag := ""
				if field.Tag != nil {
					raw, _ := strconv.Unquote(field.Tag.Value)
					tag = reflect.StructTag(raw).Get("json")
				}
				if tag == "-" {
					continue
				}
				for _, name := range field.Names {
					if !name.IsExported() {
						continue
					}
					key := strings.Split(tag, ",")[0]
					if key == "" {
						key = strings.ToLower(name.Name)
					}
					result = append(result, key)
				}
			}
			return result
		}
		return nil
	}
	var decodedFields func(string, map[string]bool) map[string]bool
	decodedFields = func(handler string, seen map[string]bool) map[string]bool {
		result := map[string]bool{}
		if seen[handler] {
			return result
		}
		seen[handler] = true
		fn := source.functions[handler]
		if fn == nil {
			return result
		}
		variables := map[string]ast.Expr{}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			if declaration, ok := node.(*ast.ValueSpec); ok && declaration.Type != nil {
				for _, name := range declaration.Names {
					variables[name.Name] = declaration.Type
				}
			}
			return true
		})
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			name := expressionName(call.Fun)
			decode := strings.HasPrefix(strings.ToLower(name), "decode")
			if name == "Decode" {
				// Only count json.NewDecoder(r.Body), not a remote provider response.
				decode = false
				ast.Inspect(call.Fun, func(node ast.Node) bool {
					if selector, ok := node.(*ast.SelectorExpr); ok && selector.Sel.Name == "Body" && expressionName(selector.X) == "r" {
						decode = true
					}
					return true
				})
			}
			if decode {
				for _, argument := range call.Args {
					if address, ok := argument.(*ast.UnaryExpr); ok && address.Op == token.AND {
						for _, field := range fields(variables[expressionName(address.X)], "server", map[string]bool{}) {
							result[field] = true
						}
					}
				}
			}
			if source.functions[name] != nil {
				requestPassed := false
				ast.Inspect(call, func(node ast.Node) bool {
					if id, ok := node.(*ast.Ident); ok && id.Name == "r" {
						requestPassed = true
					}
					return true
				})
				if requestPassed {
					for field := range decodedFields(name, seen) {
						result[field] = true
					}
				}
			}
			return true
		})
		return result
	}
	for _, operation := range Operations() {
		if !operation.Body {
			continue
		}
		handler := ""
		for _, route := range source.routes {
			if route.path == operationPath(t, operation) && route.method == operation.Method {
				handler = route.handler
				break
			}
		}
		if handler == "" || handler == "handleSideQuestions" {
			continue
		}
		for field := range decodedFields(handler, map[string]bool{}) {
			if handler == "editFindingTraffic" {
				// The dispatcher accepts a shared struct but each action uses distinct fields.
				if operation.Method == "DELETE" && field != "version" {
					continue
				}
				if operation.Method == "PUT" && field != "version" && field != "binding_ids" {
					continue
				}
				if operation.Method == "PATCH" && field == "binding_ids" {
					continue
				}
			}
			assertParameter(t, operation, "body", field, false)
		}
	}
}
