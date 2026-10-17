# Independent agent CLI E2E, 2026-10-17

Artifact: c86c7e4, `D:\Git\scopeweaver-cli\dist\release\scopeweaver-cli-0.1.0-windows-amd64\scopeweaver-cli.exe`.
321 real CLI invocations. 205/297 catalog operations invoked, including 193 with successful behavior and 12 with error-only coverage. 8 indirectly referenced in existing sources/tests; 84 unexecuted with explicit reasons. No claim that all 297 pass.

## Results

- Real bundled PostgreSQL; isolated home and unrelated working directory.
- Create/update/read-back/delete: categories, templates, company+scope, tasks, assets, constraints+scope, agents, conversations, skills/files/ZIP upload, prompts/variables/preview, disabled triggers/MCP/tool metadata, LLM profiles/retry policy, global settings, intercept rules/judge.
- Binary upload/download checked byte-for-byte; Unicode/multiline workspace and skills checked by read-back. Existing download overwrite rejected. Traversal-like write remains inside workspace; private read rejected.
- Actual task archive ready -> restore preserved name, paused state, constraints and scope. Batch archive/restore/delete and conversation batch delete exercised.
- Runtime stop/restart preserved task, conversation, workspace, skill and LLM profile state. Final runtime stopped; 0 owned runtime/PostgreSQL processes.
- Markdown report and ZIP/JSON exports inspected; no synthetic findings injected.
- Own-loopback HTTP custom tool succeeded direct and with explicit recording proxy; captured traffic exchange read back fixture marker.

## Findings requiring coordinator action

1. Automatic recording-proxy URL is double-prefixed in server/customtool.go. Two actual requests failed; direct/explicit-proxy controls succeeded. This is a reproducible engine defect in the tested artifact.
2. Embedded catalog marks some required values optional: tasks asset-refs list asset_id; llm records by-model task; source confirms llm records delete task, chat upload scope/id and agents update name. Coordinator received evidence for correction.
3. Existing replacement semantics matter: agent metadata update clears omitted description; trigger PATCH needs a full trigger condition group. Missing message/activity read returns empty detail. These are observed inherited behaviors, not silent passes.

## Limits

- Actual model-backed chat and task execution, tool/approval requests, generated findings/retests/evidence bindings, retry/edit/branch/queue/compact controls require authorized provider/inference; not exercised in this run.
- No live external ScopeSentry, MCP/search service, notification delivery or update download/install.
- Streaming event receipt/resume and keyboard cancellation not independently exercised.

35 initial expectation mismatches remain in calls.jsonl and are adjudicated individually in report.json. They include rejected exploratory inputs, incorrect harness shape/state assumptions, catalog defects and the confirmed recording-proxy defect. No result was removed.

## Operation classification

| Operation | Classification | Positive runtime behavior | Reason |
| --- | --- | --- | --- |
| agents config set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents create | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents prompt preview | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents prompt reset | runtime_exercised | no | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents prompt set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents prompts list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents triggers create | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents triggers list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents update | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents variables list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents visibility get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents visibility set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents wrapup reset | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents wrapup set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents wrapup task-timeout reset | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| agents wrapup task-timeout set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| archives delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| archives delete batch | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| archives get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| archives list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| archives restore | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| archives restore batch | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| asset-intercept rules create | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| asset-intercept rules delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| asset-intercept rules list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| asset-intercept rules set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| asset-intercept rules toggle | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| assets counts | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| assets delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| assets export | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| assets insert | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| assets list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| audit list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| auth change-password | indirectly_existing_tests | no | Synthetic credential lifecycle belongs to inherited server tests; not initialized in this independent home. |
| auth init | indirectly_existing_tests | no | Not initialized to keep this run focused on portable bearer-auth runtime; no credential read. |
| auth login | indirectly_existing_tests | no | Not logged in; no credentials used. |
| auth status | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| categories create | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| categories delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| categories list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| categories update | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| chat mentions search | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| chat send | unavailable | no | Requires model-backed chat/session/side-question execution or resulting request; provider inference excluded. |
| chat upload | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| commands list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| commands stats | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| companies create | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| companies delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| companies get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| companies list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| companies reattribute | indirectly_existing_tests | no | Company scope attribution read-back exercised, global reattribute not repeated. |
| companies scope set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| conversations controls attempts activities | unavailable | no | No model-backed turns/runs/queued input fixture; conversation state/context and invalid revision pause exercised. |
| conversations controls attempts list | unavailable | no | No model-backed turns/runs/queued input fixture; conversation state/context and invalid revision pause exercised. |
| conversations controls branch | unavailable | no | No model-backed turns/runs/queued input fixture; conversation state/context and invalid revision pause exercised. |
| conversations controls compact | unavailable | no | No model-backed turns/runs/queued input fixture; conversation state/context and invalid revision pause exercised. |
| conversations controls context | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| conversations controls edit | unavailable | no | No model-backed turns/runs/queued input fixture; conversation state/context and invalid revision pause exercised. |
| conversations controls inputs cancel | unavailable | no | No model-backed turns/runs/queued input fixture; conversation state/context and invalid revision pause exercised. |
| conversations controls inputs discard | unavailable | no | No model-backed turns/runs/queued input fixture; conversation state/context and invalid revision pause exercised. |
| conversations controls inputs followup | unavailable | no | No model-backed turns/runs/queued input fixture; conversation state/context and invalid revision pause exercised. |
| conversations controls inputs send | unavailable | no | No model-backed turns/runs/queued input fixture; conversation state/context and invalid revision pause exercised. |
| conversations controls inputs update | unavailable | no | No model-backed turns/runs/queued input fixture; conversation state/context and invalid revision pause exercised. |
| conversations controls pause | runtime_exercised | no | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| conversations controls queue order | unavailable | no | No model-backed turns/runs/queued input fixture; conversation state/context and invalid revision pause exercised. |
| conversations controls resume | unavailable | no | No model-backed turns/runs/queued input fixture; conversation state/context and invalid revision pause exercised. |
| conversations controls retry | unavailable | no | No model-backed turns/runs/queued input fixture; conversation state/context and invalid revision pause exercised. |
| conversations controls state | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| conversations controls turns prompt | unavailable | no | No model-backed turns/runs/queued input fixture; conversation state/context and invalid revision pause exercised. |
| conversations create | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| conversations delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| conversations delete batch | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| conversations list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| conversations messages get | runtime_exercised | no | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| conversations messages list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| conversations messages send | unavailable | no | Requires model-backed chat/session/side-question execution or resulting request; provider inference excluded. |
| conversations profile set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| conversations side-questions ask | unavailable | no | Requires model-backed chat/session/side-question execution or resulting request; provider inference excluded. |
| conversations side-questions clear | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| conversations side-questions list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| conversations stop | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| conversations update | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| exploration activity get | runtime_exercised | no | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| exploration activity history | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| exploration activity list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| exploration activity watch | unavailable | no | Long-running event stream/cancellation not independently exercised; direct/history read counterparts exercised where available. |
| exploration frontier list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| exploration graph get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| exploration intents list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| exploration main-sessions create | unavailable | no | Requires model-backed chat/session/side-question execution or resulting request; provider inference excluded. |
| exploration main-sessions list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| exploration nodes list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| exploration tokens get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| findings asset-tree get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| findings deepen | unavailable | no | No CLI create-finding command and no AI-generated finding fixture; did not seed database/API behind CLI. Findings list/stats/export and missing-findings rejection exercised. |
| findings delete | unavailable | no | No CLI create-finding command and no AI-generated finding fixture; did not seed database/API behind CLI. Findings list/stats/export and missing-findings rejection exercised. |
| findings export | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| findings get | runtime_exercised | no | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| findings groups list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| findings lineage | unavailable | no | No CLI create-finding command and no AI-generated finding fixture; did not seed database/API behind CLI. Findings list/stats/export and missing-findings rejection exercised. |
| findings list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| findings retests active | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| findings retests list | unavailable | no | No CLI create-finding command and no AI-generated finding fixture; did not seed database/API behind CLI. Findings list/stats/export and missing-findings rejection exercised. |
| findings retests start | unavailable | no | No CLI create-finding command and no AI-generated finding fixture; did not seed database/API behind CLI. Findings list/stats/export and missing-findings rejection exercised. |
| findings stats | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| findings traffic bind | unavailable | no | No CLI create-finding command and no AI-generated finding fixture; did not seed database/API behind CLI. Findings list/stats/export and missing-findings rejection exercised. |
| findings traffic body | unavailable | no | No CLI create-finding command and no AI-generated finding fixture; did not seed database/API behind CLI. Findings list/stats/export and missing-findings rejection exercised. |
| findings traffic delete | unavailable | no | No CLI create-finding command and no AI-generated finding fixture; did not seed database/API behind CLI. Findings list/stats/export and missing-findings rejection exercised. |
| findings traffic download | unavailable | no | No CLI create-finding command and no AI-generated finding fixture; did not seed database/API behind CLI. Findings list/stats/export and missing-findings rejection exercised. |
| findings traffic get | unavailable | no | No CLI create-finding command and no AI-generated finding fixture; did not seed database/API behind CLI. Findings list/stats/export and missing-findings rejection exercised. |
| findings traffic list | unavailable | no | No CLI create-finding command and no AI-generated finding fixture; did not seed database/API behind CLI. Findings list/stats/export and missing-findings rejection exercised. |
| findings traffic order | unavailable | no | No CLI create-finding command and no AI-generated finding fixture; did not seed database/API behind CLI. Findings list/stats/export and missing-findings rejection exercised. |
| findings traffic update | unavailable | no | No CLI create-finding command and no AI-generated finding fixture; did not seed database/API behind CLI. Findings list/stats/export and missing-findings rejection exercised. |
| findings update | unavailable | no | No CLI create-finding command and no AI-generated finding fixture; did not seed database/API behind CLI. Findings list/stats/export and missing-findings rejection exercised. |
| health | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| intercept history execution | unavailable | no | No executed agent tool/approval fixture; rules and disabled judge configuration CRUD exercised. |
| intercept history get | runtime_exercised | no | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| intercept history list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| intercept judge get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| intercept judge set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| intercept judge usage | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| intercept pending decide | unavailable | no | No executed agent tool/approval fixture; rules and disabled judge configuration CRUD exercised. |
| intercept pending get | runtime_exercised | no | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| intercept pending list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| intercept rules create | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| intercept rules delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| intercept rules list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| intercept rules set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| intercept rules toggle | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| intercept task list | unavailable | no | No executed agent tool/approval fixture; rules and disabled judge configuration CRUD exercised. |
| intercept tool-config get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| intercept tool-config set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| llm get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| llm models list | unavailable | no | Requires model/search service interaction, prohibited in this independent run. |
| llm pool reset | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| llm pool status | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| llm profiles activate | indirectly_existing_tests | no | Profile save/list/delete exercised; activation/inference not exercised. |
| llm profiles delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| llm profiles list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| llm profiles save | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| llm records by-model | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| llm records delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| llm records get | unavailable | no | No provider inference, hence no recorded LLM response ID exists. |
| llm records list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| llm records tasks | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| llm retry-policy get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| llm retry-policy set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| llm set | indirectly_existing_tests | no | Profile config path exercised; global provider binding not changed. |
| llm test | unavailable | no | Requires model/search service interaction, prohibited in this independent run. |
| logs history | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| logs list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| logs watch | unavailable | no | Long-running event stream/cancellation not independently exercised; direct/history read counterparts exercised where available. |
| maintenance gc | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| mcp delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| mcp list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| mcp refresh | unavailable | no | Requires live integration; only disabled local metadata CRUD was authorized/exercised. |
| mcp save | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| mcp tools list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| notify channels create | runtime_exercised | no | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| notify channels delete | unavailable | no | Notification destination rejected as loopback; notifications/sends prohibited and no delivery fixture exists. |
| notify channels list | runtime_exercised | no | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| notify channels test | unavailable | no | Notification destination rejected as loopback; notifications/sends prohibited and no delivery fixture exists. |
| notify channels update | unavailable | no | Notification destination rejected as loopback; notifications/sends prohibited and no delivery fixture exists. |
| notify deliveries list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| notify deliveries retry | unavailable | no | Notification destination rejected as loopback; notifications/sends prohibited and no delivery fixture exists. |
| notify meta | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| report get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| settings get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| settings python detect | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| settings set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| settings web-search test | unavailable | no | Requires model/search service interaction, prohibited in this independent run. |
| side-questions cancel | unavailable | no | Requires model-backed chat/session/side-question execution or resulting request; provider inference excluded. |
| side-questions watch | unavailable | no | Long-running event stream/cancellation not independently exercised; direct/history read counterparts exercised where available. |
| skills create | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| skills delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| skills dirs create | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| skills files delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| skills files list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| skills files read | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| skills files write | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| skills list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| skills meta set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| skills missing | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| skills upload | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| skills usage | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| stats | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| sync scopesentry datasource set | unavailable | no | Requires external ScopeSentry integration; external target/service activity excluded. |
| sync scopesentry projects list | unavailable | no | Requires external ScopeSentry integration; external target/service activity excluded. |
| sync scopesentry run | unavailable | no | Requires external ScopeSentry integration; external target/service activity excluded. |
| sync scopesentry status | unavailable | no | Requires external ScopeSentry integration; external target/service activity excluded. |
| sync scopesentry tasks list | unavailable | no | Requires external ScopeSentry integration; external target/service activity excluded. |
| tasks active set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks archive | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks archive batch | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks asset-refs list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks assets attach | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks assets detach | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks category batch-set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks category set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks chat controls attempts activities | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat controls attempts list | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat controls branch | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat controls compact | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat controls context | runtime_exercised | no | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks chat controls edit | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat controls inputs cancel | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat controls inputs discard | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat controls inputs followup | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat controls inputs send | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat controls inputs update | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat controls pause | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat controls queue order | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat controls resume | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat controls retry | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat controls state | runtime_exercised | no | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks chat controls turns prompt | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat side-questions ask | unavailable | no | No task chat transcript/run created because sending invokes model; task chat unavailable rejection/status/side-list/clear exercised. |
| tasks chat side-questions clear | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks chat side-questions list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks chat status | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks chat stop | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks constraints create | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks constraints delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks constraints list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks constraints update | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks control | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks control batch | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks coverage get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks coverage-graph get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks create | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks goals create | unavailable | no | Goal mutation schedules/resumes planning and AI was excluded; existing packaged E2E exercised goal creation (62-invocation run), not this run. |
| tasks goals delete | unavailable | no | Goal mutation schedules/resumes planning and AI was excluded; existing packaged E2E exercised goal creation (62-invocation run), not this run. |
| tasks goals list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks goals update | unavailable | no | Goal mutation schedules/resumes planning and AI was excluded; existing packaged E2E exercised goal creation (62-invocation run), not this run. |
| tasks intent-assets list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks intents control | unavailable | no | No AI exploration intent fixture; no provider calls or database seeding permitted. |
| tasks intents message | unavailable | no | No AI exploration intent fixture; no provider calls or database seeding permitted. |
| tasks intents rerun | unavailable | no | No AI exploration intent fixture; no provider calls or database seeding permitted. |
| tasks intents rerun-blocked | unavailable | no | No AI exploration intent fixture; no provider calls or database seeding permitted. |
| tasks intents side-questions ask | unavailable | no | No AI exploration intent fixture; no provider calls or database seeding permitted. |
| tasks intents side-questions clear | unavailable | no | No AI exploration intent fixture; no provider calls or database seeding permitted. |
| tasks intents side-questions list | unavailable | no | No AI exploration intent fixture; no provider calls or database seeding permitted. |
| tasks intercept-rules create | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks intercept-rules delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks intercept-rules list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks intercept-rules set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks intercept-rules toggle | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks llm resolution | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks llm set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks scope add | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks scope delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks scope list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tasks update | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| templates create | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| templates delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| templates list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| templates update | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tokens conversations | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tokens daily | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tokens usage | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tools custom create | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tools custom delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tools custom set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tools custom test | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tools list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tools reset | runtime_exercised | no | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| tools set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| traffic blob download | unavailable | no | Captured synthetic response was small; no oversized traffic blob was generated. Traffic exchange capture/read-back exercised. |
| traffic delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| traffic exchange get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| traffic hosts delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| traffic hosts list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| traffic list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| traffic purge | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| triggers delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| triggers update | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| update apply | unavailable | no | Release/update lifecycle was outside permitted scope; first public stable release does not exist. No installs/publication performed. |
| update check | unavailable | no | Release/update lifecycle was outside permitted scope; first public stable release does not exist. No installs/publication performed. |
| update rollback | unavailable | no | Release/update lifecycle was outside permitted scope; first public stable release does not exist. No installs/publication performed. |
| update watch | unavailable | no | Release/update lifecycle was outside permitted scope; first public stable release does not exist. No installs/publication performed. |
| visibility get | indirectly_existing_tests | no | Per-resource tool/MCP visibility not independently exercised; skill visibility exercised. |
| visibility set | indirectly_existing_tests | no | Per-resource tool/MCP visibility not independently exercised; skill visibility exercised. |
| visibility skill get | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| visibility skill set | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| workspace delete | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| workspace download | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| workspace list | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| workspace mkdir | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| workspace read | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| workspace upload | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
| workspace write | runtime_exercised | yes | Actual packaged executable invocation; coverage is bounded by individual assertions in calls.jsonl. |
