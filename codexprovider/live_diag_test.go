package codexprovider

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Autumn-27/norma/llm"
)

// Opt-in one-shot diagnostic through the production bridge with a smaller,
// one-tool request. It does not recreate the full engine conversation. Rejected
// arguments are classified without recording their contents.
// Bounded: glm-5.3/low, at most 2 subprocess attempts (Complete's correction
// budget) under the shared 10-minute invocation timeout. Enable with
// CODEXPROVIDER_GLM_DIAG=1 plus inherited ZAI_API_KEY. Never logs raw tool
// content, prompts or credentials.
func TestLiveGLMArgumentShapeDiagnostic(t *testing.T) {
	if os.Getenv("CODEXPROVIDER_GLM_DIAG") != "1" {
		t.Skip("set CODEXPROVIDER_GLM_DIAG=1 (plus ZAI_API_KEY) to opt into the one-shot live shape diagnostic")
	}
	if strings.TrimSpace(os.Getenv("ZAI_API_KEY")) == "" {
		t.Fatal("CODEXPROVIDER_GLM_DIAG=1 requires ZAI_API_KEY in the environment")
	}
	p, err := New(Options{Model: "glm-5.3", BaseURL: ZAIBaseURL, ReasoningEffort: "low"})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	marker := "CODEX_READ_" + func() string {
		b := make([]byte, 12)
		_, _ = rand.Read(b)
		return fmt.Sprintf("%x", b)
	}()
	source := filepath.Join(dir, "codex-e2e-marker.txt")
	if err := os.WriteFile(source, []byte(marker), 0600); err != nil {
		t.Fatal(err)
	}
	// Uses the E2E request shape with only the owned Read tool available.
	req := llm.CompletionRequest{
		System: []string{"You are a careful assistant in an authorized isolated integration test."},
		Messages: []llm.Message{llm.UserText(fmt.Sprintf(
			"This is an authorized isolated CLI integration test. Use the Norma Read tool exactly once to read this owned test file: %s. Do not use any other tool or any path outside that file. Then answer exactly MARKER=<the complete file contents>. Do not guess the contents.", source))},
		Tools: []llm.ToolSchema{{
			Name:        "Read",
			Description: "Reads a file from the local filesystem.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"file_path": map[string]any{"type": "string", "description": "The absolute path to the file to read"},
				},
				"required": []string{"file_path"},
			},
		}},
	}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Minute)
	defer cancel()
	start := time.Now()
	message, stop, usage, err := p.Complete(ctx, req)
	elapsed := time.Since(start)
	// Content-free outcome only: error text (carries the classifier detail),
	// stop reason, block inventory, tool names, usage fields.
	kinds := make([]string, 0, len(message.Content))
	tools := make([]string, 0, len(message.Content))
	var readInput json.RawMessage
	for _, block := range message.Content {
		kinds = append(kinds, string(block.Type))
		if block.Type == llm.BlockToolUse {
			tools = append(tools, block.Name)
			readInput = block.Input
		}
	}
	t.Logf("live diagnostic: elapsed=%s err=%v stop=%q blocks=%v tool_names=%v usage(in=%d out=%d cache_read=%d cache_write=%d) marker_present=%v",
		elapsed.Truncate(time.Millisecond), err, stop, kinds, tools,
		usage.InputTokens, usage.OutputTokens, usage.CacheReadTokens, usage.CacheWriteTokens,
		strings.Contains(fmt.Sprint(message.Content), marker))
	if err != nil {
		t.Fatalf("live GLM completion failed (content-free detail above): %v", err)
	}
	if stop != "tool_use" || len(tools) != 1 || tools[0] != "Read" {
		t.Fatal("completion did not request exactly one Read tool")
	}
	var args struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal(readInput, &args); err != nil {
		t.Fatal("Read arguments did not decode")
	}
	got, want := filepath.Clean(args.FilePath), filepath.Clean(source)
	matches := got == want
	if runtime.GOOS == "windows" {
		matches = strings.EqualFold(got, want)
	}
	if !matches {
		t.Fatal("Read did not target the owned file")
	}
	if strings.Contains(fmt.Sprint(message.Content), marker) {
		t.Fatal("completion returned file contents before the tool executed")
	}
}
