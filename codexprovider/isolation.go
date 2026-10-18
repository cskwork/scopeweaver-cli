package codexprovider

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func processConfig(options Options, scratch string) ([]string, []string, error) {
	settings := []string{
		"project_doc_max_bytes=0", "apps._default.enabled=false", "features.apps=false",
		"features.shell_tool=false", "features.hooks=false", "features.multi_agent=false",
		"features.plugins=false", "features.skills=false", "web_search=disabled", "tools.view_image=false",
		"notify=[]", "analytics.enabled=false",
		`otel.exporter="none"`, `otel.trace_exporter="none"`, `otel.metrics_exporter="none"`,
		"otel.log_user_prompt=false", "otel.log_agent_responses=false", "otel.log_guardian_assessments=false",
	}
	if options.ReasoningEffort != "" {
		encoded, _ := json.Marshal(options.ReasoningEffort)
		settings = append(settings, "model_reasoning_effort="+string(encoded))
	}
	env := make([]string, 0, len(os.Environ())+1)
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		upper := strings.ToUpper(name)
		if strings.HasPrefix(upper, "OPENAI_") || strings.HasPrefix(upper, "CODEX_") && upper != "CODEX_HOME" {
			continue
		}
		switch strings.ToUpper(name) {
		case "OPENAI_API_KEY", "OPENAI_BASE_URL", "CODEX_API_KEY", "ANTHROPIC_API_KEY", "ANTHROPIC_BASE_URL", "ZAI_API_KEY", "SCOPEWEAVER_CODEX_ZAI_API_KEY":
			continue
		}
		env = append(env, entry)
	}
	if options.BaseURL == ZAIBaseURL {
		key := options.APIKey
		if strings.TrimSpace(key) == "" {
			key = os.Getenv("ZAI_API_KEY")
		}
		if strings.TrimSpace(key) == "" {
			return nil, nil, errors.New("codex: Z.AI requires an API key or ZAI_API_KEY")
		}
		catalogPath := filepath.Join(scratch, "models.json")
		if err := os.WriteFile(catalogPath, []byte(zaiModelCatalog), 0600); err != nil {
			return nil, nil, errors.New("codex: could not write isolated GLM model catalog")
		}
		encodedPath, _ := json.Marshal(catalogPath)
		settings = append(settings, "model_catalog_json="+string(encodedPath))
		env = append(env, "SCOPEWEAVER_CODEX_ZAI_API_KEY="+key)
		settings = append(settings, "model_provider=scopeweaver_zai",
			`model_providers.scopeweaver_zai.name="ScopeWeaver Z.AI"`,
			`model_providers.scopeweaver_zai.base_url="https://api.z.ai/api/v1"`,
			`model_providers.scopeweaver_zai.env_key="SCOPEWEAVER_CODEX_ZAI_API_KEY"`,
			`model_providers.scopeweaver_zai.wire_api="responses"`,
			"model_providers.scopeweaver_zai.requires_openai_auth=false")
	} else {
		settings = append(settings, "model_provider=openai",
			`openai_base_url="https://chatgpt.com/backend-api/codex"`,
			`chatgpt_base_url="https://chatgpt.com/backend-api/"`)
	}
	args := make([]string, 0, len(settings)*2+2)
	for _, setting := range settings {
		args = append(args, "--config", setting)
	}
	args = append(args, "app-server", "--stdio")
	return args, env, nil
}

func (s *session) isolatedThreadConfig(ctx context.Context) (map[string]any, error) {
	result, err := s.request(ctx, "config/read", map[string]any{"includeLayers": false, "cwd": s.cmd.Dir})
	if err != nil {
		return nil, err
	}
	var response struct {
		Config map[string]json.RawMessage `json:"config"`
	}
	if json.Unmarshal(result, &response) != nil || response.Config == nil {
		return nil, errors.New("codex: could not inspect effective configuration")
	}
	overrides := map[string]any{}
	raw := response.Config["mcp_servers"]
	if len(raw) == 0 {
		raw = response.Config["mcpServers"]
	}
	if len(raw) > 0 && string(raw) != "null" {
		var servers map[string]json.RawMessage
		if json.Unmarshal(raw, &servers) != nil {
			return nil, errors.New("codex: malformed MCP configuration")
		}
		for name := range servers {
			// App-server splits dotted override keys literally; it does not unquote
			// TOML components. Fail before thread creation for unsupported names.
			if name == "" || strings.IndexFunc(name, func(r rune) bool {
				return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-')
			}) >= 0 {
				return nil, errors.New("codex: cannot isolate MCP server with a name outside letters, numbers, underscore or hyphen; use a clean Codex home/login or rename that server in Codex configuration")
			}
			overrides["mcp_servers."+name+".enabled"] = false
		}
	}
	return overrides, nil
}

func (s *session) verifyNoMCP(ctx context.Context) error {
	cursor := ""
	seen := map[string]bool{}
	for page := 0; page < 128; page++ {
		params := map[string]any{"threadId": s.threadID}
		if cursor != "" {
			params["cursor"] = cursor
		}
		result, err := s.request(ctx, "mcpServerStatus/list", params)
		if err != nil {
			return err
		}
		var response struct {
			Data *[]struct {
				RuntimeStatus string                     `json:"runtimeStatus"`
				Tools         map[string]json.RawMessage `json:"tools"`
				Resources     []json.RawMessage          `json:"resources"`
				Templates     []json.RawMessage          `json:"resourceTemplates"`
			} `json:"data"`
			NextCursor *string `json:"nextCursor"`
		}
		if json.Unmarshal(result, &response) != nil || response.Data == nil {
			return errors.New("codex: malformed MCP isolation status")
		}
		for _, server := range *response.Data {
			if server.RuntimeStatus != "disabled" || len(server.Tools) > 0 || len(server.Resources) > 0 || len(server.Templates) > 0 {
				return errors.New("codex: MCP isolation failed; enabled server or exposed capabilities remain")
			}
		}
		if response.NextCursor == nil || *response.NextCursor == "" {
			return nil
		}
		cursor = *response.NextCursor
		if seen[cursor] {
			return errors.New("codex: repeated MCP status cursor")
		}
		seen[cursor] = true
	}
	return errors.New("codex: MCP status pagination exceeded bounds")
}

func (s *session) verifyChatGPTLogin(ctx context.Context) error {
	result, err := s.request(ctx, "account/read", map[string]any{"refreshToken": false})
	if err != nil {
		return err
	}
	var response struct {
		Account *struct {
			Type string `json:"type"`
		} `json:"account"`
	}
	if json.Unmarshal(result, &response) != nil || response.Account == nil || response.Account.Type != "chatgpt" {
		return errors.New("codex: ChatGPT subscription login required; run codex login using the official CLI")
	}
	return nil
}
func replaceEnv(env []string, name, value string) []string {
	out := make([]string, 0, len(env)+1)
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if !strings.EqualFold(key, name) {
			out = append(out, entry)
		}
	}
	return append(out, name+"="+value)
}
