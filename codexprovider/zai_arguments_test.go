package codexprovider

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/Autumn-27/norma/llm"
)

// R22 focused regressions for the Z.AI nested-object arguments wire form.
// The R20 full-engine failure (receipt da96e7a1…, invalid-json 1 byte after one
// correction) localized to the JSON-inside-a-string arguments field inside an
// otherwise valid envelope. The fix requests a nested JSON object on the Z.AI
// route only; OAuth keeps its strict string schema. These tests pin:
//   - both wire forms decode with byte-identical accepted input semantics;
//   - the exact R20 hard-failure priority for duplicate keys in ANY valid JSON
//     representation (the R21 partial regressed this for legacy arrays);
//   - repairable shapes stay content-free and Z.AI-correction-eligible;
//   - the requested schema and instructions are route-specific.

var r22Tools = []llm.ToolSchema{{Name: "echo", Description: "echo", InputSchema: map[string]any{"type": "object"}}}

// envelopeWith builds an action envelope whose single call carries arguments as
// the given JSON value (raw literal for object/array/scalar kinds, or a JSON
// string for the legacy form when legacy is a Go string).
func envelopeWith(argumentsValue any) string {
	raw, err := json.Marshal(map[string]any{"assistant_text": "", "tool_calls": []any{
		map[string]any{"name": "echo", "arguments": argumentsValue},
	}})
	if err != nil {
		panic(err)
	}
	return string(raw)
}

func TestDecodeEnvelopeAcceptsObjectFormArguments(t *testing.T) {
	envelope := `{"assistant_text":"reading","tool_calls":[{"name":"echo","arguments":{"value":"hello","path":"D:\\owned\\표시.txt"}}]}`
	message, stop, err := decodeEnvelope(envelope, r22Tools)
	if err != nil {
		t.Fatalf("object-form arguments rejected: %v", err)
	}
	if stop != "tool_use" || len(message.ToolUses()) != 1 || message.ToolUses()[0].Name != "echo" {
		t.Fatalf("unexpected message %#v stop %q", message, stop)
	}
	// Accepted nested-object bytes pass through untransformed.
	if got := string(message.ToolUses()[0].Input); got != `{"value":"hello","path":"D:\\owned\\표시.txt"}` {
		t.Fatalf("input bytes not preserved: %s", got)
	}
	var input map[string]any
	if json.Unmarshal(message.ToolUses()[0].Input, &input) != nil || input["path"] != `D:\owned\표시.txt` {
		t.Fatalf("input not the equivalent object: %s", message.ToolUses()[0].Input)
	}
}

func TestDecodeEnvelopeAcceptsMixedObjectAndLegacyStringForms(t *testing.T) {
	tools := append([]llm.ToolSchema{}, r22Tools[0], llm.ToolSchema{Name: "Read", Description: "read", InputSchema: map[string]any{"type": "object"}})
	envelope := `{"assistant_text":"","tool_calls":[` +
		`{"name":"Read","arguments":{"file_path":"x"}},` +
		`{"name":"echo","arguments":"{\"value\":\"hello\"}"}]}`
	message, stop, err := decodeEnvelope(envelope, tools)
	if err != nil || stop != "tool_use" {
		t.Fatalf("mixed forms rejected: %v stop %q", err, stop)
	}
	calls := message.ToolUses()
	if len(calls) != 2 {
		t.Fatalf("calls %d", len(calls))
	}
	var read map[string]any
	if json.Unmarshal(calls[0].Input, &read) != nil || read["file_path"] != "x" {
		t.Fatalf("object-form input wrong: %s", calls[0].Input)
	}
	if string(calls[1].Input) != `{"value":"hello"}` {
		t.Fatalf("string-form input wrong: %s", calls[1].Input)
	}
}

func TestObjectAndLegacyFormsProduceEquivalentInputs(t *testing.T) {
	objectForm := envelopeWith(map[string]any{"value": "hello", "n": 2})
	legacyForm := envelopeWith(`{"n":2,"value":"hello"}`)
	m1, _, err1 := decodeEnvelope(objectForm, r22Tools)
	m2, _, err2 := decodeEnvelope(legacyForm, r22Tools)
	if err1 != nil || err2 != nil {
		t.Fatalf("decode errors: %v %v", err1, err2)
	}
	var a, b map[string]any
	_ = json.Unmarshal(m1.ToolUses()[0].Input, &a)
	_ = json.Unmarshal(m2.ToolUses()[0].Input, &b)
	if len(a) != len(b) || a["value"] != b["value"] || a["n"] != b["n"] {
		t.Fatalf("forms not equivalent: %v vs %v", a, b)
	}
}

// R21-P2 counterexample, now pinned: a legacy string whose valid JSON content is
// an array with duplicate keys must keep the R20 nonrepairable priority.
func TestLegacyStringDuplicateKeysInsideArrayFailClosed(t *testing.T) {
	for _, content := range []string{
		// array containing a duplicate-key object
		`[{"a":1,"a":2}]`,
		// duplicate keys in the object itself
		`{"a":1,"a":2}`,
		// duplicate keys nested under an object
		`{"nest":[{"a":1,"a":2}]}`,
		// duplicates nested inside an array
		`[{"deep":{"x":{"k":1,"k":2}}}]`,
	} {
		_, _, err := decodeEnvelope(envelopeWith(content), r22Tools)
		if err == nil {
			t.Fatalf("duplicate keys accepted: %s", content)
		}
		if errors.Is(err, errArgumentObject) {
			t.Fatalf("duplicate keys classified repairable (R21-P2 regression): %s: %v", content, err)
		}
		if !strings.Contains(err.Error(), "ambiguous") {
			t.Fatalf("unexpected error class: %v", err)
		}
	}
}

// Object-form duplicate keys are rejected by the whole-envelope ambiguity scan
// before any call is classified; they must never reach the repair path.
func TestObjectFormDuplicateKeysFailClosed(t *testing.T) {
	envelope := `{"assistant_text":"","tool_calls":[{"name":"echo","arguments":{"a":1,"a":2}}]}`
	_, _, err := decodeEnvelope(envelope, r22Tools)
	if err == nil || errors.Is(err, errArgumentObject) || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("object-form duplicates must fail closed without retry: %v", err)
	}
}

// A repairable first call must not outrank a later strict failure: the whole
// envelope is validated before any error class is chosen (R20 priority).
func TestRepairableBeforeAmbiguousStillFailsClosed(t *testing.T) {
	envelope := `{"assistant_text":"","tool_calls":[` +
		`{"name":"echo","arguments":[]},` +
		`{"name":"echo","arguments":"{\"a\":1,\"a\":2}"}]}`
	_, _, err := decodeEnvelope(envelope, r22Tools)
	if err == nil || errors.Is(err, errArgumentObject) || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("later ambiguity must dominate a repairable shape: %v", err)
	}
}

// A valid object-form call followed by an unknown tool fails closed with the
// nonrepairable unknown-tool class; no call executes and no retry starts.
func TestObjectFormValidCallThenUnknownToolFailsClosed(t *testing.T) {
	envelope := `{"assistant_text":"","tool_calls":[` +
		`{"name":"echo","arguments":{"value":"x"}},` +
		`{"name":"shell","arguments":{}}]}`
	_, _, err := decodeEnvelope(envelope, r22Tools)
	if err == nil || errors.Is(err, errArgumentObject) || !strings.Contains(err.Error(), "unavailable tool") {
		t.Fatalf("unknown tool must keep the strict class: %v", err)
	}
}

// Raw non-object arguments values stay repairable with exact content-free
// detail: kind and trimmed byte count only. A raw JSON null reports "nil"
// like an absent value (matches the restored R20 detail contract test).
func TestRawNonObjectArgumentsShapesRepairableWithDetail(t *testing.T) {
	for _, tc := range []struct {
		name, value, detail string
	}{
		{"null", `null`, `; rejected call 1 tool "echo": nil, 4 bytes`},
		{"absent", ``, `; rejected call 1 tool "echo": nil, 0 bytes`},
		{"array", `[]`, `; rejected call 1 tool "echo": not-object:array, 2 bytes`},
		{"number", `42`, `; rejected call 1 tool "echo": not-object:number, 2 bytes`},
		{"boolean", `true`, `; rejected call 1 tool "echo": not-object:boolean, 4 bytes`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			envelope := `{"assistant_text":"","tool_calls":[{"name":"echo","arguments":` + tc.value + `}]}`
			if tc.value == `` {
				envelope = `{"assistant_text":"","tool_calls":[{"name":"echo"}]}`
			}
			_, _, err := decodeEnvelope(envelope, r22Tools)
			if !errors.Is(err, errArgumentObject) {
				t.Fatalf("repairable identity lost: %v", err)
			}
			if !strings.Contains(err.Error(), tc.detail) {
				t.Fatalf("detail mismatch: got %s want substring %s", err.Error(), tc.detail)
			}
		})
	}
	// The legacy string form keeps R20 content classification verbatim.
	for _, tc := range []struct {
		content, detail string
	}{
		{`x`, `; rejected call 1 tool "echo": invalid-json, 1 bytes`},
		{`null`, `; rejected call 1 tool "echo": not-object:null, 4 bytes`},
	} {
		_, _, err := decodeEnvelope(envelopeWith(tc.content), r22Tools)
		if !errors.Is(err, errArgumentObject) || !strings.Contains(err.Error(), tc.detail) {
			t.Fatalf("legacy classification changed: %v", err)
		}
	}
}

// The requested wire form is route-specific: Z.AI asks for a nested object,
// OAuth keeps the strict-compatible string, and all other envelope constraints
// (required fields, additionalProperties:false, tool-name enum) are identical.
func TestZAIOutputSchemaRequestsObjectAndOAuthKeepsString(t *testing.T) {
	schemaArguments := func(schema map[string]any) map[string]any {
		t.Helper()
		items := schema["properties"].(map[string]any)["tool_calls"].(map[string]any)["items"].(map[string]any)
		if required, ok := items["required"].([]string); !ok || len(required) != 2 || required[0] != "name" || required[1] != "arguments" {
			t.Fatalf("call item required fields changed: %v", items["required"])
		}
		if items["additionalProperties"] != false || schema["additionalProperties"] != false {
			t.Fatal("additionalProperties:false constraints were weakened")
		}
		return items["properties"].(map[string]any)["arguments"].(map[string]any)
	}
	if args := schemaArguments(zaiOutputSchema(r22Tools)); args["type"] != "object" {
		t.Fatalf("ZAI arguments schema type = %v, want object", args["type"])
	}
	if args := schemaArguments(outputSchema(r22Tools)); args["type"] != "string" {
		t.Fatalf("OAuth arguments schema type = %v, want string", args["type"])
	}
	toolsFree := zaiOutputSchema(nil)
	if calls := toolsFree["properties"].(map[string]any)["tool_calls"].(map[string]any); calls["maxItems"] != 0 {
		t.Fatal("tools-free ZAI schema must still forbid calls")
	}
}

// Instructions and the single correction feedback are route-consistent: the
// Z.AI variant teaches the nested-object form and never the string contract;
// the OAuth instructions are unchanged.
func TestInstructionsAndFeedbackAreRouteSpecific(t *testing.T) {
	if !strings.Contains(zaiBridgeInstructions, "nested JSON object") ||
		!strings.Contains(zaiBridgeInstructions, "COMPLETE INPUT OBJECT") ||
		strings.Contains(zaiBridgeInstructions, "JSON object strings") {
		t.Fatal("Z.AI bridge instructions must teach the nested-object representation")
	}
	if !strings.Contains(bridgeInstructions, "JSON object strings") ||
		strings.Contains(bridgeInstructions, "nested JSON object") {
		t.Fatal("OAuth bridge instructions must keep the string representation")
	}
	if !strings.Contains(argumentRepairFeedback, "nested JSON object") ||
		strings.Contains(argumentRepairFeedback, "arguments string") {
		t.Fatal("correction feedback must teach the nested-object representation")
	}
}

// Subprocess round trip through the real bridge: the Z.AI route sends the
// object-form schema and instructions (asserted by the fake app-server), and a
// nested-object response decodes without any correction being started.
func TestGLMObjectFormSubprocessRoundTrip(t *testing.T) {
	p, sessions := argumentRepairProvider(t, "object-valid", ZAIBaseURL)
	message, stop, usage, err := p.Complete(context.Background(), toolRequest())
	if err != nil {
		t.Fatal(err)
	}
	if message.Text() != "hello" || stop != "tool_use" || len(message.ToolUses()) != 1 {
		t.Fatalf("unexpected completion %#v stop %q", message, stop)
	}
	if got := string(message.ToolUses()[0].Input); got != `{"value":"hello"}` {
		t.Fatalf("object-form input wrong: %s", got)
	}
	if usage != oneArgumentUsage {
		t.Fatalf("usage %#v", usage)
	}
	requireArgumentSessions(t, sessions, "initial\n")
}
