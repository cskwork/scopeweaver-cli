# Exact CLI operation coverage

**297/297 exact names invoked: 297 positive-behavior, 0 error-only, 0 indirect-only, 0 untested.** Positive means at least one asserted behavior, not every option/state or final-release certification. Invocation totals are never added as unique coverage. Every exact name has at least one positive contract; separate release limits remain below.

R3 lifecycle adds ten positives: genuine unknown-run discard and undelivered-steer followup for conversation/task, task resume, conversation/task edit and retry, and task branch. Actual completed replies, idempotent followup, persisted restart recovery and inherited branch history were asserted. The 130-call campaign plus three separate preflight calls observed simultaneous model-selected loopback tool calls and a running side question, then stopped in0.824s with a nonce-bound success receipt. No campaign database-closed finalization errors or remaining owned processes were observed. Original wrong tool_call oracle and failed compact receipts are preserved; corrected tool_use evidence proves the actual tool requests.

R4 adds both compact operations: 33 CLI calls and two actual completion requests, without seed inference. Genuine retained history shrank on the same estimate basis by15.67%/14.94%; marker summaries, count1/manual source, completed turn/attempt history, idempotency and restart persistence passed. Numeric paid usage persisted in LLM records52/53; existing aggregate ledgers stayed unchanged, so aggregate inclusion of compaction cost is not claimed.

R3 OAuth full workflow passed125 calls in161.24s. R4 GLM Flash full workflow passed127 calls in153.87s, including actual paired Read, planner/worker loopback finding/report and all19 sidecar positives. GLM5.3 diagnostic remains failed: worker finding completed, but planner final turn stalled6m13s and the workflow exceeded its budget. Flash success does not relabel GLM5.3 as passed.


R4 remaining local work adds ten positives: both legacy chat sends and the LLM connection test, authentic approval pending/detail/decision/navigation, worker pause, human message continuation with a real synthetic fact, and eligible exhausted-intent rerun. Deny caused0 HTTP requests; allow in an independent conversation caused1 owned GET, with paired persisted tool history. Intent4 rerun entered running and completed done, independently adding27681 input/71 output tokens to its session. The346-call campaign (338main plus8 read-only diagnostics;35 exact catalog names) ended model work in568.3s within840s; cleanup found zero owned processes and unchanged R4 hash. [The sanitized receipt](remaining-local-r4.json) preserves exploratory harness failures and expected invalid/duplicate decision errors; it does not label every invocation passed. Intent control proves pause, not every control action.


R4 owned integration fixtures add eight positives in 35 CLI calls and 23 actual HTTP requests: MCP refresh, five ScopeSentry datasource/list/status/import commands, OpenAI-compatible model discovery and DeepSeek search-SSE parsing. Nonempty remote results, exact arguments, refreshed cache and typed imported asset/company readbacks passed. No live vendor, external notification or provider inference was used. [The sanitized receipt](integrations-r4.json) preserves corrected asset-type/default-filter and HTTP-header-case harness failures. `llm models list` now has positive OpenAI-compatible proof; Codex-cli discovery remains explicitly unsupported.


R5 blocked-intent recovery closes one operation with 64 CLI calls: a genuine owned provider HTTP 400 put intent 2 into blocked, repairing only the fixture enabled bulk rerun to reopen 1, and the same intent finished done with actual engine record_fact node 4. The initial graph-field oracle failure is retained. [Blocked-intent receipt](blocked-intent-r5.json).

R6 notifications close two operations with 29 CLI calls: unset/false local-relay configuration stayed blocked; explicit opt-in 1 survived child-runtime isolation and performed a real owned synchronous test. An actual synthetic worker finding generated delivery 1, the owned sink returned 400, then retry sent that same delivery/finding with 200 after only sink repair. Exactly 3 owned notification requests and 0 external notifications occurred. This resolves the old R4 environment-drop gap on the tested dirty-tree clone; its copied manifest was not regenerated, so it is not a final release artifact. [Notification receipt](notification-r6.json).

R5 updater closes three operations with 20 CLI calls and 25 assertions in nonroot WSL Linux: check/apply/rollback exercised the pinned GitHub host/repository/checksum/ZIP and executable smoke gates through an owned, nonforwarding TLS 1.3 HTTPS fixture with per-process CA trust. Untrusted CA was rejected first; global trust stayed unchanged. Runtime versions 0.1.0→0.1.1→0.1.0 and PIDs 901→1054→1165 were observed automatically before any CLI could start the runtime; the same company marker survived, and rollback restored the exact original SHA256. This is not a published GitHub release or native Windows/macOS updater proof. The 0.1.1 source was pinned to ad814451 and all 1120 blobs matched after Windows archive CRLF normalization; Go inherited dirty enclosing-parent VCS metadata. A later optional read found native/tmp scratch absent, cause unknown; acceptance/readback receipts were already durable. [Updater receipt](updater-r5.json).

| Candidate | SHA256 | Bounded proof |
| --- | --- | --- |
| R2 | `246bf17e9aa783b6f91367b0b3e2374156e71c1ea619a7e2ff908aeef26fe835` |132-call OAuth full PASS; native stream cancellation exit130; no-model stored-state slice |
| R3 | `4efa7d565fa14c46275da49caf680c03387abe15582926bb0df19e9b1322fddd` |Lifecycle/recovery and125-call OAuth PASS; compact failure and GLM5.3 incomplete workflow retained |
| R4 | `4e0c9d74b903f67248ed8c870b8d42ed39e465cb7db9eac5d2e57dc8303c91d2` |Both compact positives,127-call GLM Flash full PASS, ten remaining local functional contracts and eight local protocol integration contracts |
| R5 Windows | `4a1704cdd12c908afe2d2390c88a8041c49b1cc816a9a1d60f3f82346215a0a5` |Genuine blocked-intent bulk recovery |
| R6 Windows test clone | `bc36fa7b7f1e378b6d159b0b38d67cb59625e262207cf219f8cbb39aec4bc461` |Explicit local notification opt-in, default rejection and genuine failed-delivery retry |
| R5 WSL Linux updater | `43ad8cd490d96e80cd902ca7bb65f5d4efb102a686b635085b9e16ad16a0e7e2` → `7e02d901769f31ac1a4a4d86e577b4e981445d6fc6234cf28489e226f47e7591` → original |Owned HTTPS fixture apply/rollback with actual automatic restarts |


This is a historical union across earlier baseline and R2/R3/R4/R5/R6 candidates. R2–R4 derive from dirty c86c7e4-era work; the updater fixture source is pinned ad814451 with the separately recorded EOL/VCS qualification. Current reconciliation HEAD 80f9016 is not claimed as every tested binary revision. R2 frozen sidecar source 5447b8f… and reconstructed matching main source 13c0a74a… provenance remain in the JSON; no compiled-test manifest is implied for later runs.

No exact catalog operation remains without positive proof. Release qualification remains separate:

| Residual limit | What the coverage does not establish |
| --- | --- |
| All states/options and final artifact | One positive contract per name across historical candidates; no every-state/every-flag or final native artifact certification. |
| Live vendor services/delivery | Owned MCP/ScopeSentry/model-list/search/notification fixtures prove installed protocols, not live vendor delivery. |
| Published GitHub release | Owned TLS updater fixture proves actual installed replacement/rollback; no public release was created, and no native Windows/macOS updater cycle ran. |
| Final CI and human review | Native final artifact/CI results and required human review must be assessed independently. Coverage alone does not clear public release. |

`findings deepen` proves linked worker execution; fresh HTTP deepening was declined under the original exactly-once constraint. Disabled notification CRUD and earlier empty interception-list reads remain qualified; separate authentic approval decisions and paired execution navigation are now proven. No release-wide all-functions pass is claimed.

Machine-readable classifications, new operation provenance and receipt hashes: [catalog-coverage-union-20261017.json](catalog-coverage-union-20261017.json).
