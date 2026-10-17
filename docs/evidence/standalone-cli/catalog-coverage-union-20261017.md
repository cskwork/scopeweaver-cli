# Exact remaining CLI coverage

**284/297 exact names invoked: 283 positive-behavior, 1 error-only/unsupported, 0 indirect-only, 13 untested.** Positive means at least one asserted behavior, not every option/state or final-release certification. Invocation totals are never added as unique coverage. Fourteen names still lack positive proof.

R3 lifecycle adds ten positives: genuine unknown-run discard and undelivered-steer followup for conversation/task, task resume, conversation/task edit and retry, and task branch. Actual completed replies, idempotent followup, persisted restart recovery and inherited branch history were asserted. The 130-call campaign plus three separate preflight calls observed simultaneous model-selected loopback tool calls and a running side question, then stopped in0.824s with a nonce-bound success receipt. No campaign database-closed finalization errors or remaining owned processes were observed. Original wrong tool_call oracle and failed compact receipts are preserved; corrected tool_use evidence proves the actual tool requests.

R4 adds both compact operations: 33 CLI calls and two actual completion requests, without seed inference. Genuine retained history shrank on the same estimate basis by15.67%/14.94%; marker summaries, count1/manual source, completed turn/attempt history, idempotency and restart persistence passed. Numeric paid usage persisted in LLM records52/53; existing aggregate ledgers stayed unchanged, so aggregate inclusion of compaction cost is not claimed.

R3 OAuth full workflow passed125 calls in161.24s. R4 GLM Flash full workflow passed127 calls in153.87s, including actual paired Read, planner/worker loopback finding/report and all19 sidecar positives. GLM5.3 diagnostic remains failed: worker finding completed, but planner final turn stalled6m13s and the workflow exceeded its budget. Flash success does not relabel GLM5.3 as passed.


R4 remaining local work adds ten positives: both legacy chat sends and the LLM connection test, authentic approval pending/detail/decision/navigation, worker pause, human message continuation with a real synthetic fact, and eligible exhausted-intent rerun. Deny caused0 HTTP requests; allow in an independent conversation caused1 owned GET, with paired persisted tool history. Intent4 rerun entered running and completed done, independently adding27681 input/71 output tokens to its session. The346-call campaign (338main plus8 read-only diagnostics;35 exact catalog names) ended model work in568.3s within840s; cleanup found zero owned processes and unchanged R4 hash. [The sanitized receipt](remaining-local-r4.json) preserves exploratory harness failures and expected invalid/duplicate decision errors; it does not label every invocation passed. Intent control proves pause, not every control action.

| Candidate | SHA256 | Bounded proof |
| --- | --- | --- |
| R2 | `246bf17e9aa783b6f91367b0b3e2374156e71c1ea619a7e2ff908aeef26fe835` |132-call OAuth full PASS; native stream cancellation exit130; no-model stored-state slice |
| R3 | `4efa7d565fa14c46275da49caf680c03387abe15582926bb0df19e9b1322fddd` |Lifecycle/recovery and125-call OAuth PASS; compact failure and GLM5.3 incomplete workflow retained |
| R4 | `4e0c9d74b903f67248ed8c870b8d42ed39e465cb7db9eac5d2e57dc8303c91d2` |Both compact positives,127-call GLM Flash full PASS and ten remaining local functional contracts |

These are intermediate dirty candidates based on source revision c86c7e4a78d52cd3831de41494792fa16f6f145a. The historical union also retains earlier candidates and failures. Later final flush-error propagation changes require separate verification. R2 frozen sidecar source5447b8f… and reconstructed matching main source13c0a74a… provenance remain in the JSON; no compiled-test manifest is implied for later runs.

| Pending proof | Exact commands |
| --- | --- |
| Integration service/delivery fixture (9) | `mcp refresh`; `notify channels test`; `notify deliveries retry`; `settings web-search test`; `sync scopesentry datasource set`; `sync scopesentry projects list`; `sync scopesentry run`; `sync scopesentry status`; `sync scopesentry tasks list` |
| Release/update lifecycle (3) | `update apply`; `update check`; `update rollback` |
| Unsupported Codex model discovery (1) | `llm models list` |
| Authentic blocked-intent recovery (1) | `tasks intents rerun-blocked` |

Bulk intent recovery still needs a genuine blocked intent; an empty set is not recovery proof. Owned service fixtures can prove integration contracts but cannot establish live-vendor delivery. Update apply/rollback need a compatible release lifecycle and authorization. Codex model discovery is explicitly unsupported, not a successful discovery capability.

`findings deepen` proves linked worker execution; fresh HTTP deepening was declined under the original exactly-once constraint. Disabled notification CRUD and earlier empty interception-list reads remain qualified; separate authentic approval decisions and paired execution navigation are now proven. No release-wide all-functions pass is claimed.

Machine-readable classifications, new operation provenance and receipt hashes: [catalog-coverage-union-20261017.json](catalog-coverage-union-20261017.json).
