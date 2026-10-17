package agent

import (
	"strings"
	"testing"

	"github.com/Autumn-27/norma/llm"
)

func TestCodexConfigDefaultsAndKeylessReadiness(t *testing.T) {
	t.Setenv("ZAI_API_KEY", "")
	for _, tc := range []struct {
		name, base, key, model string
		ready                  bool
	}{
		{name: "ChatGPT OAuth", model: "gpt-6.1-sol", ready: true},
		{name: "Coding Plan", base: " https://api.z.ai/api/v1/responses/ ", key: "fixture-zai-key", model: "glm-5.3", ready: true},
		{name: "Coding Plan needs key", base: "https://api.z.ai/api/v1", model: "glm-5.3"},
		{name: "arbitrary gateway", base: "https://gateway.example/v1", key: "fixture-key", model: "gpt-6.1-sol"},
		{name: "local gateway", base: "http://127.0.0.1:1234/v1", model: "gpt-6.1-sol"},
		{name: "OAuth must not use API key", key: "fixture-key", model: "gpt-6.1-sol"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ConfigFrom("codex-cli", "", tc.base, tc.key, "")
			if cfg.Provider() != "codex-cli" || cfg.Model != tc.model || cfg.Ready() != tc.ready {
				t.Fatalf("provider=%q model=%q ready=%v; want model=%q ready=%v", cfg.Provider(), cfg.Model, cfg.Ready(), tc.model, tc.ready)
			}
			provider, err := cfg.NewProvider()
			if (err == nil) != tc.ready || (provider != nil) != tc.ready {
				t.Fatalf("provider configured=%v err=%v", provider != nil, err)
			}
			if strings.Contains(tc.base, "api.z.ai") && cfg.BaseURL != "https://api.z.ai/api/v1" {
				t.Fatalf("endpoint not normalized: %q", cfg.BaseURL)
			}
		})
	}
}

func TestCodexEnvironmentDoesNotInheritOtherProviderKeys(t *testing.T) {
	isolateLLMEnvironment(t)
	t.Setenv("SCOPEWEAVER_LLM_PROVIDER", "codex-cli")
	t.Setenv("ZAI_API_KEY", "")
	cfg, ready := FromEnv()
	if !ready || cfg.Provider() != "codex-cli" || cfg.APIKey != "" {
		t.Fatalf("OAuth ready=%v provider=%q inheritedKey=%v", ready, cfg.Provider(), cfg.APIKey != "")
	}
	t.Setenv("SCOPEWEAVER_LLM_BASE_URL", "https://api.z.ai/api/v1/responses")
	cfg, ready = FromEnv()
	if ready || cfg.APIKey != "" {
		t.Fatalf("ZAI inherited another provider credential: ready=%v inheritedKey=%v", ready, cfg.APIKey != "")
	}
	t.Setenv("ZAI_API_KEY", "fixture-coding-plan-key")
	cfg, ready = FromEnv()
	if !ready || cfg.APIKey != "" || cfg.Model != "glm-5.3" {
		t.Fatalf("ZAI environment readiness=%v model=%q copiedKey=%v", ready, cfg.Model, cfg.APIKey != "")
	}
	t.Setenv("SCOPEWEAVER_LLM_BASE_URL", "")
	cfg, ready = FromEnv()
	if !ready || cfg.APIKey != "" || cfg.Model != "gpt-6.1-sol" {
		t.Fatal("ZAI credential leaked into OAuth configuration")
	}
}

func TestCodexFactoryRoutesBeforeHTTPAndEnforcesLivePolicy(t *testing.T) {
	cfg := ConfigFrom("codex-cli", "", "", "", "")
	strict := false
	cfg.PolicyResolver = func() (bool, []string) { return strict, nil }
	cfg.ReasoningEffort = "high"
	cfg.RatePerMinute = 1000
	p, err := cfg.NewProvider()
	if err != nil {
		t.Fatal(err)
	}
	// A cached provider must obey a policy change before looking for Codex/login.
	strict = true
	if _, _, _, err := p.Complete(t.Context(), llm.CompletionRequest{}); err == nil || !strings.Contains(err.Error(), "local-LLM-only") {
		t.Fatalf("Complete policy error=%v", err)
	}
	events := 0
	for _, err := range p.Stream(t.Context(), llm.CompletionRequest{}) {
		events++
		if err == nil || !strings.Contains(err.Error(), "local-LLM-only") {
			t.Fatalf("Stream policy error=%v", err)
		}
	}
	if events != 1 {
		t.Fatalf("Stream policy error events=%d want=1", events)
	}
	cfg.ReasoningEffort = "max"
	if _, err := cfg.NewProvider(); err == nil {
		t.Fatal("unsupported Codex effort accepted")
	}
	cfg.ReasoningEffort = "high"
	cfg.Proxy = "http://127.0.0.1:1"
	if _, err := cfg.NewProvider(); err == nil || !strings.Contains(err.Error(), "proxy") {
		t.Fatalf("unsupported proxy error=%v", err)
	}
	if cfg.Ready() {
		t.Fatal("unsupported proxy marked ready")
	}
}
