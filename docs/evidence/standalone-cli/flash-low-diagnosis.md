# R7 Flash/low diagnostic review

Confirmed: the diagnostic repeat failed after 286.06 seconds because two findings existed, not because the expected name was missing or differently capitalized. Source: dist/candidate-r7-glm-flash-diagnostic-e2e.jsonl and dist/candidate-r7-glm-flash-diagnostic-output.txt (2026-10-17, Asia/Seoul).

- First report_finding input: intent_id=3, asset_ids=[1], severity=low, vulnclass=information_disclosure_debug. Success returned finding_id=1 / finding_node_id=5; persisted timestamp 16:51:02.
- Second input: intent_id=6, same asset, same owned endpoint and observed synthetic marker, severity=low, vulnclass=Information Disclosure - Debug Endpoint. Success returned finding_id=2 / finding_node_id=7; persisted timestamp 16:51:36. Its evidence explicitly refers to the earlier worker observation from intent 3.
- Both rows have exactly the expected name, SYNTHETIC LAB ONLY public debug marker, and task_id=1. The captured attempt count for argument-format correction is zero. Real chat/Read/result, branch history and paused queue/resume had already passed. The complete workflow still failed its exactly-one-finding assertion.

The direct cause of this assertion failure is confirmed: two distinct successful tool calls produced two independent committed rows. Why the agent workflow issued the second call remains unknown; these receipts do not establish a model, provider, concurrency, or engine root cause. They also do not establish that changing reasoning effort will prevent repetition.

Waiting is not justified as a finding-count repair: agent/tools.go:1264-1286 awaits recording before success; evidence/store.go:177-188 calls db.RecordFindingTx inside WithEvidenceTx; db/finding_traffic.go:343-416 atomically inserts graph and standalone rows. The asynchronous reporter updates the existing detailed report (server/orchestration.go:430-458). It does not materialize or remove these finding rows. A report-ready wait is relevant only to detailed report completeness, using a nonempty report plus report_stale=false.

Name normalization cannot fix this observed case: both names already match, while count=2 violates the requirement. Test-side deduplication would hide a second successful write. Engine deduplication would change product behavior and needs its own identity/semantics requirements, especially because these calls supplied different vulnclass values; it is not an evidence-backed adapter repair or validation adjustment here. Preserve the unchanged exactly-one assertion and this failure receipt.

A Flash/high run is a bounded configuration discriminator: same R7 binary and assertions, changing only the profile reasoning effort through test-only SCOPEWEAVER_CLI_GLM_E2E_EFFORT=high. The helper accepts only low/high/max and defaults to low (internal/cli/codex_e2e_test.go:167-175).

If that run passes, the defensible claim is: this R7 packaged built-in GLM-5.3-Flash/high run completed the existing isolated end-to-end workflow, including the exactly-one finding assertion. Report its exact receipt, duration, usage and observed correction count. Keep the Flash/low failure visible. One pass does not prove high caused the improvement, general reliability, low-mode repair, flagship GLM-5.3 acceptance, or live argument-format recovery when the correction count is zero. Describe any later detailed reporter completion only if separately observed.

Read-only review: no source edits, tests, API calls, credentials or database access. Only this ignored receipt was written. Diagnostic JSONL SHA-256: EB6A0C309CC8AD0B106DF56150073654607B8EA8B1BD0437B5A102F0E41354C2.
