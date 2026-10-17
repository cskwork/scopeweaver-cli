package db

import "testing"

func TestCodexFormatMigrationPreservesProfilesAndReopens(t *testing.T) {
	dsn := testDSN(t)
	d, err := Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	ids := []int64{}
	defer func() {
		for _, id := range ids {
			_, _ = d.Exec(`DELETE FROM llm_profiles WHERE id=$1`, id)
		}
	}()
	for _, format := range []string{"openai", "anthropic", "openai-responses", "ollama"} {
		id, err := d.SaveProfile(&LLMProfile{Name: "t-codex-migration-" + format, Format: format, Model: "existing-model", APIKey: "fixture-preserved-key", ReasoningEffort: "high", Streaming: true})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	// Reproduce the prior supported-format constraint on the disposable database.
	if _, err := d.Exec(`ALTER TABLE llm_profiles DROP CONSTRAINT llm_profiles_format_check; ALTER TABLE llm_profiles ADD CONSTRAINT llm_profiles_format_check CHECK(format IN ('openai','anthropic','openai-responses','ollama'))`); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	for _, id := range ids {
		p, err := reopened.ProfileByID(id)
		if err != nil || p == nil {
			t.Fatalf("existing profile %d missing: %v", id, err)
		}
		if p.Model != "existing-model" || p.APIKey != "fixture-preserved-key" || p.ReasoningEffort != "high" || !p.Streaming {
			t.Fatalf("existing profile %d changed", id)
		}
	}
	p := &LLMProfile{Name: "t-codex-migration-oauth", Format: "codex-cli", Model: "gpt-6.1-sol", ReasoningEffort: "high", Streaming: true}
	id, err := reopened.SaveProfile(p)
	if err != nil {
		t.Fatal(err)
	}
	ids = append(ids, id)
	again, err := Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer again.Close()
	got, err := again.ProfileByID(id)
	if err != nil || got == nil || got.Format != p.Format || got.Model != p.Model || got.APIKey != "" || got.ReasoningEffort != p.ReasoningEffort {
		t.Fatalf("Codex persisted profile missing/changed: err=%v", err)
	}
	pool, err := again.PoolProfiles()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range pool {
		found = found || p.ID == id
	}
	if !found {
		t.Fatal("keyless OAuth profile absent from pool")
	}
	if _, err = again.SaveProfile(&LLMProfile{Name: "t-codex-invalid-format", Format: "invalid", Model: "m"}); err == nil {
		t.Fatal("format constraint was weakened")
	}
}

func TestCodexProfileCredentialTransitions(t *testing.T) {
	t.Setenv("ZAI_API_KEY", "")
	d, err := Open(testDSN(t))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	for _, tc := range []struct {
		name, oldFormat, oldBase, nextBase string
		keep                               bool
	}{
		{name: "cloud to OAuth", oldFormat: "openai"},
		{name: "cloud to Coding Plan", oldFormat: "openai", nextBase: "https://api.z.ai/api/v1"},
		{name: "Coding Plan to OAuth", oldFormat: "codex-cli", oldBase: "https://api.z.ai/api/v1"},
		{name: "Coding Plan blank edit", oldFormat: "codex-cli", oldBase: "https://api.z.ai/api/v1", nextBase: "https://api.z.ai/api/v1", keep: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &LLMProfile{Name: "t-codex-transition-" + tc.name, Format: tc.oldFormat, BaseURL: tc.oldBase, Model: "gpt-6.1-sol", APIKey: "fixture-stored-key"}
			if p.BaseURL != "" {
				p.Model = "glm-5.3"
			}
			id, err := d.SaveProfile(p)
			if err != nil {
				t.Fatal(err)
			}
			defer d.Exec(`DELETE FROM llm_profiles WHERE id=$1`, id)
			p.ID, p.Format, p.BaseURL, p.APIKey = id, "codex-cli", tc.nextBase, ""
			p.Model = "gpt-6.1-sol"
			if p.BaseURL != "" {
				p.Model = "glm-5.3"
			}
			if _, err := d.SaveProfile(p); err != nil {
				t.Fatal(err)
			}
			got, err := d.ProfileByID(id)
			if err != nil {
				t.Fatal(err)
			}
			if (got.APIKey != "") != tc.keep {
				t.Fatalf("stored credential retained=%v, want %v", got.APIKey != "", tc.keep)
			}
			ps, err := d.ListProfiles()
			if err != nil {
				t.Fatal(err)
			}
			for _, v := range ps {
				if v.ID == id && (v.APIKeyHint != "") != tc.keep {
					t.Fatalf("credential hint retained=%v want=%v", v.APIKeyHint != "", tc.keep)
				}
			}
		})
	}
}
