package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cskwork/scopeweaver-cli/agent"
	"github.com/cskwork/scopeweaver-cli/db"
	"github.com/cskwork/scopeweaver-cli/internal/dbtest"
)

func TestCodexConnectionDiagnosticsPreserveKeylessRoute(t *testing.T) {
	t.Setenv("SCOPEWEAVER_CONFIG", filepath.Join(t.TempDir(), "missing.json"))
	t.Setenv("SCOPEWEAVER_LOCAL_LLM_ONLY", "true")
	t.Setenv("ZAI_API_KEY", "")
	s := &Server{llmCfg: agent.ConfigFrom("openai", "cloud", "", "fixture-active-key", "")}
	for _, tc := range []struct{ name, base, proxy, want string }{
		{name: "OAuth uses Codex with no inherited key", want: "local-LLM-only"},
		{name: "unapproved endpoint", base: "https://gateway.example/v1", want: "base URL"},
		{name: "ZAI requires its own key", base: "https://api.z.ai/api/v1/responses", want: "Z.AI requires"},
		{name: "unsupported proxy", proxy: "http://127.0.0.1:1", want: "proxy"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := testCodexConnection(t, s, tc.base, tc.proxy)
			diagnostic, _ := result["error"].(string)
			if result["ok"] != false || !strings.Contains(diagnostic, tc.want) {
				t.Fatalf("diagnostic=%v; want %q", result, tc.want)
			}
		})
	}
	s.llmCfg = agent.ConfigFrom("codex-cli", "glm-5.3", "https://api.z.ai/api/v1", "fixture-coding-plan-key", "")
	result := testCodexConnection(t, s, "https://api.z.ai/api/v1/responses", "")
	if diagnostic, _ := result["error"].(string); !strings.Contains(diagnostic, "local-LLM-only") {
		t.Fatalf("same Coding Plan credential was not retained: %v", result)
	}
}

func TestCodexProfileDTOAndDiscovery(t *testing.T) {
	t.Setenv("ZAI_API_KEY", "")
	oauth := llmProfileDTO(&db.LLMProfile{ID: 1, Format: "codex-cli", Model: "gpt-6.1-sol"})
	if !oauth.APIKeyOptional {
		t.Fatal("OAuth profile incorrectly requires an API key")
	}
	zai := llmProfileDTO(&db.LLMProfile{ID: 2, Format: "codex-cli", Model: "glm-5.3", BaseURL: "https://api.z.ai/api/v1"})
	if zai.APIKeyOptional {
		t.Fatal("Coding Plan profile incorrectly marked keyless")
	}
	var calls atomic.Int32
	fixture := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
	defer fixture.Close()
	raw, _ := json.Marshal(map[string]any{"provider": "codex-cli", "base_url": fixture.URL, "api_key": "fixture-key"})
	s := &Server{}
	response := httptest.NewRecorder()
	s.pgListModels(response, httptest.NewRequest(http.MethodPost, "/api/llm/models", bytes.NewReader(raw)))
	if response.Code != http.StatusOK || calls.Load() != 0 || !strings.Contains(response.Body.String(), "model discovery is unsupported") {
		t.Fatalf("discovery HTTP=%d calls=%d body=%s", response.Code, calls.Load(), response.Body)
	}
}

func TestCodexKeylessProfileResolutionAndSharedFactory(t *testing.T) {
	dbtest.DSN(t)
	m, err := NewManager(t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	s := &Server{m: m, provByProfile: map[int64]*provEntry{}}
	p := &db.LLMProfile{Name: "t-codex-shared-profile", Format: "codex-cli", Model: "gpt-6.1-sol", ReasoningEffort: "high", Streaming: true, RatePerMinute: 1000}
	id, err := m.pg.SaveProfile(p)
	if err != nil {
		t.Fatal(err)
	}
	defer m.pg.Exec(`DELETE FROM llm_profiles WHERE id=$1`, id)
	p.ID = id
	cfg, ready := s.loadProfileConfig(id)
	if !ready || !cfg.Ready() || cfg.Provider() != "codex-cli" || cfg.APIKey != "" {
		t.Fatalf("configuration ready=%v provider=%q keyPresent=%v", ready, cfg.Provider(), cfg.APIKey != "")
	}
	if !s.resolutionFromProfile(p, "task_chain").Available {
		t.Fatal("keyless task profile unavailable")
	}
	first, cfg, ok := s.providerForProfile(id)
	if !ok || first == nil || cfg.Provider() != "codex-cli" {
		t.Fatal("profile did not use Codex factory")
	}
	second, _, ok := s.providerForProfile(id)
	if !ok || first != second {
		t.Fatal("profile callers do not share one provider and limiter")
	}
}

func testCodexConnection(t *testing.T, s *Server, base, proxy string) map[string]any {
	t.Helper()
	// Omitted model uses the real provider defaults: OAuth GPT or the pinned GLM.
	raw, err := json.Marshal(map[string]any{"provider": "codex-cli", "base_url": base, "proxy": proxy, "api_key": "", "streaming": false})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	s.testLLM(response, httptest.NewRequest(http.MethodPost, "/api/llm/test", bytes.NewReader(raw)))
	var result map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil || response.Code != http.StatusOK {
		t.Fatalf("HTTP %d response=%s decode error=%v", response.Code, response.Body, err)
	}
	return result
}
