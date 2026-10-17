package cli

import (
	"bytes"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Reuse completed subscription work; queue edits and evidence operations must not start another model run.
func runSubscriptionExtraE2E(t *testing.T, run func(...string) map[string]any, submit func(map[string]any, ...string) map[string]any, conversationID, taskID string) {
	t.Helper()
	task := run("tasks", "get", "--id", taskID)
	if task["status"] != "done" {
		t.Fatal("extra E2E requires the synthetic task to be done")
	}
	reporterRunning := 0
	for _, conversation := range codexRows(t, run("conversations", "list"), "conversations") {
		if conversation["agent_key"] == "reporter" && conversation["running"] == true {
			reporterRunning++
		}
	}
	t.Logf("Pre-existing reporter conversations running=%d; their conversation activity is excluded from the task exploration token total.", reporterRunning)
	// Done may precede a cancelled worker's final result; wait for actual task calls to settle before measuring.
	settleDeadline := time.Now().Add(30 * time.Second)
	var activeTask map[string]any
	for {
		stats := run("stats", "--task", taskID)
		var ok bool
		activeTask, ok = stats["active_task"].(map[string]any)
		if !ok || codexString(activeTask["id"]) != taskID {
			t.Fatal("task runtime counters unavailable for the usage baseline")
		}
		workerCalls, workerOK := activeTask["in_flight"].(float64)
		modelCalls, modelOK := activeTask["llm_in_flight"].(float64)
		if !workerOK || !modelOK {
			t.Fatal("task runtime counters have an unexpected type")
		}
		if workerCalls == 0 && modelCalls == 0 {
			break
		}
		if time.Now().After(settleDeadline) {
			t.Fatalf("pre-existing task work did not settle before the usage baseline: counters=%v", activeTask)
		}
		select {
		case <-t.Context().Done():
			t.Fatal(t.Context().Err())
		case <-time.After(250 * time.Millisecond):
		}
	}
	t.Log("Task usage baseline is quiescent: no running intents or active planner/worker calls; original whole-task invariance remains enforced.")
	usageBefore := run("exploration", "tokens", "get", "--task", taskID)
	state := run("conversations", "controls", "state", "--id", conversationID)
	initialRun := codexString(state["latest_run_id"])
	turnID := codexString(state["latest_turn_id"])
	if state["active"] != nil || initialRun == "" || turnID == "" {
		t.Fatal("extra E2E requires a completed real conversation")
	}
	attempts := codexRows(t, run("conversations", "controls", "attempts", "list", "--id", conversationID, "--turn-id", turnID), "attempts")
	found := false
	for _, attempt := range attempts {
		if codexString(attempt["run_id"]) != initialRun {
			continue
		}
		found = true
		if attempt["status"] != "completed" {
			t.Fatal("latest real attempt was not completed")
		}
		activities := codexRows(t, run("conversations", "controls", "attempts", "activities", "--id", conversationID, "--attempt-id", initialRun), "items")
		if len(activities) == 0 {
			t.Fatal("real attempt has no persisted activities")
		}
	}
	if !found {
		t.Fatal("latest real run absent from attempt history")
	}
	prompt := run("conversations", "controls", "turns", "prompt", "--id", conversationID, "--turn-id", turnID)
	if strings.TrimSpace(codexString(prompt["text"])) == "" {
		t.Fatal("real turn prompt was not persisted")
	}
	submit(map[string]any{"expected_revision": state["revision"]}, "conversations", "controls", "pause", "--id", conversationID)
	first := submit(map[string]any{"client_request_id": rand.Text(), "mode": "queue", "message": "Extra fixture A. Never execute this queued input."}, "conversations", "controls", "inputs", "send", "--id", conversationID)
	second := submit(map[string]any{"client_request_id": rand.Text(), "mode": "queue", "message": "Extra fixture B. Never execute this queued input."}, "conversations", "controls", "inputs", "send", "--id", conversationID)
	updated := submit(map[string]any{"expected_revision": first["revision"], "message": "Updated queued fixture A; never execute."}, "conversations", "controls", "inputs", "update", "--id", conversationID, "--input-id", codexString(first["id"]))
	updatedPrompt, ok := updated["prompt"].(map[string]any)
	if !ok || updatedPrompt["text"] != "Updated queued fixture A; never execute." {
		t.Fatal("queued input edit was not persisted")
	}
	state = run("conversations", "controls", "state", "--id", conversationID)
	ordered := submit(map[string]any{"expected_revision": state["revision"], "input_ids": []string{codexString(second["id"]), codexString(first["id"])}}, "conversations", "controls", "queue", "order", "--id", conversationID)
	inputs := codexRows(t, ordered, "inputs")
	if len(inputs) != 2 || codexString(inputs[0]["id"]) != codexString(second["id"]) || codexString(inputs[1]["id"]) != codexString(first["id"]) {
		t.Fatal("queued input order did not change")
	}
	for _, input := range inputs {
		run("conversations", "controls", "inputs", "cancel", "--id", conversationID, "--input-id", codexString(input["id"]), "--expected-revision", codexString(input["revision"]), "--yes")
	}
	state = run("conversations", "controls", "state", "--id", conversationID)
	if state["paused"] != true || state["active"] != nil || len(codexRows(t, state, "inputs")) != 0 || codexString(state["latest_run_id"]) != initialRun {
		t.Fatal("queue-only operations executed or retained pending inputs")
	}

	findings := codexRows(t, run("findings", "list", "--task-id", taskID, "--page", "1", "--limit", "20"), "items")
	if len(findings) != 1 || !strings.Contains(codexString(findings[0]["name"]), "SYNTHETIC LAB ONLY") {
		t.Fatal("expected exactly one real synthetic finding")
	}
	findingID := codexString(findings[0]["finding_id"])
	if findingID == "" {
		t.Fatal("standalone finding ID missing; exploration node IDs are not interchangeable")
	}
	detail := run("findings", "get", "--id", findingID)
	originalName := codexString(detail["name"])
	if detail["task_id"] != taskID || detail["severity"] != "low" {
		t.Fatal("synthetic finding provenance or severity mismatch")
	}
	lineage := run("findings", "lineage", "--id", findingID)
	if len(codexRows(t, lineage, "nodes")) == 0 {
		t.Fatal("real finding has no exploration lineage")
	}
	if len(codexRows(t, run("findings", "retests", "list", "--id", findingID), "retests")) != 0 {
		t.Fatal("unexpected retest before any retest request")
	}
	// Name edits do not publish status events or trigger a reporter run.
	changed := submit(map[string]any{"name": originalName + " [metadata round trip]"}, "findings", "update", "--id", findingID)
	if changed["name"] != originalName+" [metadata round trip]" {
		t.Fatal("finding metadata edit was not persisted")
	}
	restored := submit(map[string]any{"name": originalName}, "findings", "update", "--id", findingID)
	if restored["name"] != originalName {
		t.Fatal("finding name was not restored")
	}

	extraPositiveCommands := 18
	records := run("llm", "records", "list", "--task", taskID, "--size", "10")
	recorded := codexRows(t, records, "records")
	if len(recorded) == 0 {
		t.Log("LLM-record get remains unverified: codex-cli may meter usage without retaining raw/normalized request records; no synthetic record was inserted.")
	} else {
		record := run("llm", "records", "get", "--id", codexString(recorded[0]["id"]))
		if codexString(record["id"]) != codexString(recorded[0]["id"]) || record["task_id"] != taskID {
			t.Fatal("persisted LLM record read-back mismatch")
		}
		extraPositiveCommands++
	}
	intercepts := run("intercept", "task", "list", "--taskID", taskID)
	codexRows(t, intercepts, "items") // An empty history is valid; no approval decision is claimed.
	unsupported := submit(map[string]any{"provider": "codex-cli"}, "llm", "models", "list")
	if unsupported["ok"] != false || !strings.Contains(codexString(unsupported["error"]), "unsupported") {
		t.Fatal("Codex model discovery must report its unsupported capability explicitly")
	}

	settings := run("settings", "get")
	if settings["notify_enabled"] != false {
		t.Fatal("extra E2E requires notification delivery disabled")
	}
	submit(map[string]any{"traffic_capture": true}, "settings", "set")
	defer submit(map[string]any{"traffic_capture": settings["traffic_capture"]}, "settings", "set")
	payloads := map[string][]byte{"/first": []byte("EVIDENCE_FIRST_" + rand.Text()), "/second": []byte("EVIDENCE_SECOND_" + rand.Text())}
	fixture := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, ok := payloads[r.URL.Path]
		if !ok || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write(payload)
	}))
	defer fixture.Close()
	for _, path := range []string{"/first", "/second"} {
		response := submit(map[string]any{"kind": "http", "exec": map[string]any{"url": fixture.URL + path, "method": "GET", "timeout_ms": 5000, "use_recording_proxy": true}}, "tools", "custom", "test")
		if response["is_error"] != false || !strings.Contains(codexString(response["output"]), string(payloads[path])) {
			t.Fatal("owned loopback recording-proxy request failed")
		}
	}
	fixtureURL, err := url.Parse(fixture.URL)
	if err != nil {
		t.Fatal(err)
	}
	var exchanges []map[string]any
	var lastTraffic map[string]any
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		// Stored hosts omit ports; URL search keeps this fixture separate from other loopback services.
		lastTraffic = run("traffic", "list", "--host", fixtureURL.Hostname(), "--q", fixture.URL)
		exchanges = codexRows(t, lastTraffic, "exchanges")
		if len(exchanges) == 2 {
			break
		}
		select {
		case <-t.Context().Done():
			t.Fatal(t.Context().Err())
		case <-time.After(100 * time.Millisecond):
		}
	}
	if len(exchanges) != 2 {
		loopback := run("traffic", "list", "--host", fixtureURL.Hostname(), "--size", "500")
		t.Fatalf("both own loopback exchanges required: fixture=%s filtered=%v loopback_metadata=%v", fixture.URL, lastTraffic, loopback)
	}
	trafficIDs := map[string]string{}
	for _, exchange := range exchanges {
		for _, path := range []string{"/first", "/second"} {
			if exchange["url"] == fixture.URL+path {
				trafficIDs[path] = codexString(exchange["id"])
			}
		}
	}
	if trafficIDs["/first"] == "" || trafficIDs["/second"] == "" {
		t.Fatal("captured exchanges do not belong to the owned fixture")
	}
	before := run("findings", "traffic", "list", "--id", findingID)
	previousCount := len(codexRows(t, before, "bindings"))
	evidence := submit(map[string]any{"traffic_refs": []map[string]any{{"traffic_id": trafficIDs["/first"], "role": "proof", "note": "Owned first fixture"}, {"traffic_id": trafficIDs["/second"], "role": "supporting", "note": "Owned second fixture"}}}, "findings", "traffic", "bind", "--id", findingID)
	bindings := codexRows(t, evidence, "bindings")
	if len(bindings) != previousCount+2 {
		t.Fatal("actual traffic snapshots were not bound to the finding")
	}
	newIDs := map[string]string{}
	order := make([]string, len(bindings))
	firstPosition, secondPosition := -1, -1
	for i, binding := range bindings {
		order[i] = codexString(binding["id"])
		snapshot, ok := binding["snapshot"].(map[string]any)
		if !ok {
			t.Fatal("bound traffic snapshot missing")
		}
		switch codexString(snapshot["source_traffic_id"]) {
		case trafficIDs["/first"]:
			newIDs["/first"] = order[i]
			firstPosition = i
		case trafficIDs["/second"]:
			newIDs["/second"] = order[i]
			secondPosition = i
		}
	}
	if firstPosition < 0 || secondPosition < 0 {
		t.Fatal("bound snapshots lost original traffic provenance")
	}
	firstID := newIDs["/first"]
	evidenceDetail := run("findings", "traffic", "get", "--id", findingID, "--binding-id", firstID)
	preview, ok := evidenceDetail["response"].(map[string]any)
	if !ok || preview["content"] != string(payloads["/first"]) {
		t.Fatal("bound traffic detail is not the original response")
	}
	body := run("findings", "traffic", "body", "--id", findingID, "--binding-id", firstID, "--side", "response")
	if body["content"] != string(payloads["/first"]) {
		t.Fatal("bound response preview differs from original bytes")
	}
	downloaded := filepath.Join(t.TempDir(), "response.bin")
	run("findings", "traffic", "download", "--id", findingID, "--binding-id", firstID, "--side", "response", "--output", downloaded)
	raw, err := os.ReadFile(downloaded)
	if err != nil || !bytes.Equal(raw, payloads["/first"]) {
		t.Fatal("bound response download did not preserve original bytes")
	}
	evidence = submit(map[string]any{"version": evidence["version"], "note": "Updated owned evidence note", "role": "verification"}, "findings", "traffic", "update", "--id", findingID, "--binding-id", firstID)
	verified := false
	for _, binding := range codexRows(t, evidence, "bindings") {
		if binding["id"] == firstID {
			verified = binding["note"] == "Updated owned evidence note" && binding["role"] == "verification"
		}
	}
	if !verified {
		t.Fatal("evidence role/note edit was not persisted")
	}
	order[firstPosition], order[secondPosition] = order[secondPosition], order[firstPosition]
	evidence = submit(map[string]any{"version": evidence["version"], "binding_ids": order}, "findings", "traffic", "order", "--id", findingID)
	for i, binding := range codexRows(t, evidence, "bindings") {
		if codexString(binding["id"]) != order[i] {
			t.Fatal("evidence order did not change")
		}
	}
	for _, path := range []string{"/first", "/second"} {
		evidence = submit(map[string]any{"version": evidence["version"]}, "findings", "traffic", "delete", "--id", findingID, "--binding-id", newIDs[path])
	}
	if len(codexRows(t, evidence, "bindings")) != previousCount {
		t.Fatal("owned evidence cleanup changed pre-existing bindings")
	}
	if after := run("llm", "records", "list", "--task", taskID, "--size", "10"); after["total"] != records["total"] {
		t.Fatal("inference-free sidecar unexpectedly produced another task LLM record")
	}
	usageAfter := run("exploration", "tokens", "get", "--task", taskID)
	if !reflect.DeepEqual(usageBefore["total"], usageAfter["total"]) {
		t.Fatalf("inference-free sidecar changed the quiescent task usage ledger: before=%v after=%v", usageBefore["total"], usageAfter["total"])
	}
	state = run("conversations", "controls", "state", "--id", conversationID)
	if state["active"] != nil || codexString(state["latest_run_id"]) != initialRun {
		t.Fatal("inference-free sidecar started a conversation run")
	}
	t.Logf("Verified %d additional positive CLI operations plus explicit unsupported Codex model discovery; queued edits, real finding reads and byte-exact loopback evidence, without extra inference. Discard remains untested: it requires an unknown run, not an ordinary queued input.", extraPositiveCommands)
}
