# Standalone CLI verification

[한국어](ko/CLI-VERIFICATION.md)

Recorded 2026-10-17 on Windows amd64. Engine source pinned to ScopeWeaver commit 2df0b307dcb817a459abebd3a4c240a2019598db. Tests use disposable local state and PostgreSQL17.11. Live OAuth tests consume the configured subscription; only synthetic files and an owned loopback HTTP fixture were used. No online target testing or external notification sends occurred.

## Current candidate evidence (2026-10-17)

The first public v0.1.0 release is **held**. These dated R3/R4 results supersede earlier status below; they do not certify the final public package.

- Built-in chat, planner and worker agents use `codex-cli` profiles through official Codex CLI 0.162.1. ChatGPT OAuth and Z.AI Coding Plan remain separate provider routes. See [provider setup and limits](ai-providers.md).
- Windows R3 passed **125 installed OAuth CLI invocations in 161.24 seconds**, including real chat tools, queue/resume, planner/worker completion, one owned loopback GET, a synthetic finding and a report. Windows core E2E passed **62 invocations in 34.39 seconds**. [R3 receipt](evidence/standalone-cli/candidate-r3-verification.json).
- Windows R4 passed **127 GLM-5.3-Flash/low invocations in 153.87 seconds** with a replacement Coding Plan key: chat/tool results, branch/restart, paused queue/resume, planner intent, worker WebFetch, a synthetic LOW finding, goal completion and report. The fixture observed exactly one GET and zero unexpected target requests. Flagship GLM-5.3 reached worker evidence and finding registration but its final planner call exceeded the seven-minute task E2E budget; **its full workflow is not accepted**. Earlier key errors and tool-format failures are historical, not the current Flash result. [R4 Flash receipt](evidence/standalone-cli/glm-flash-e2e-r4.json).
- The R3 OAuth lifecycle campaign used **130 CLI calls plus 3 separate preflights**. Active chat/task and side-question shutdown/drain, restart, steering, editing, retry and branching were checked. The side question was observed running and persisted cancelled; this does not establish paid side-inference completion. Zero owned processes remained. The R3 compaction gap was resolved separately in R4. [Lifecycle receipt](evidence/standalone-cli/oauth-lifecycle-r3.json).
- R4 OAuth compaction passed **33 CLI calls and 2 new model completions**. Conversation and task context shrank, synthetic history markers and published summaries survived restart, and paid per-call records were retained. Existing dashboard/aggregate usage did not increase; do not claim compaction costs were included in those aggregates. [Compaction receipt](evidence/standalone-cli/oauth-compaction-r4.json).
- Linux R3 ran on the native ext4 filesystem of existing Ubuntu WSL2 as a non-root user. First-use download/start took **6.679 seconds** and **62 installed E2E invocations passed in 10.579 seconds**. This was a Windows-cross-compiled binary executed on Linux, not a native Linux build or clean VM. No native macOS runtime E2E is recorded. [Linux receipt](evidence/standalone-cli/candidate-r3-linux.json).
- The exact-name catalog union snapshot recorded at **2026-10-17T06:45:31Z** invoked **284 of 297 operations**, with **283 positive, 1 error-only and 13 untested**. It combines version-specific campaigns; positive means at least one asserted behavior, not every option/state or final-candidate certification. [Coverage snapshot](evidence/standalone-cli/catalog-coverage-union-20261017.json).
- The R4 remaining-local campaign added **10 previously nonpositive operation contracts** through **346 CLI calls across 35 names**. It verified actual OAuth completions, legacy chat admissions, authentic interception allow/deny history and task-intent controls. Zero owned processes remained. Exploratory fixture/harness failures are preserved; this is not a claim that every attempted assertion passed. [Campaign receipt](evidence/standalone-cli/remaining-local-r4.json).
- The newest final-flush write-error propagation and localization changes pass source tests but are newer than the tested R4 binary. A rebuilt public candidate and its installed checks remain required. Earlier Windows stream cancellation now returns 130; installed R3 checks passed without forced termination. [R3 receipt](evidence/standalone-cli/candidate-r3-verification.json).
- Public packages use the approved first-use download of checksum-pinned PostgreSQL/runtime files into the writable portable folder. No separate application/database installer or OS service is needed. Earlier bundled archives remain local test artifacts and are not the public package.
- The exact user-requested agy `gemini-4-argon-medium` landing design is retained with paired EN/KO content. Initial browser confirmation passed 148 assertions across six layouts; later factual updates have separate bounded QA and do not inherit that count.
- No remote, hosted CI run, human review, public tag or release exists. Final rebuilt-package checks, required external integrations, the online updater release cycle and target-specific evidence remain open. Catalog coverage and candidate passes do not waive these release gates.

## Earlier subscription and setup campaigns (historical, superseded above)

The first public v0.1.0 release is **held**. This section supersedes the earlier preparation status below.

- Built-in planner, worker and chat support for codex-cli is implemented. Official Codex CLI0.162.1 manages ChatGPT OAuth; the Z.ai route selects GLM-5.3 through its documented Responses endpoint. See [provider setup and limits](ai-providers.md).
- A packaged Windows candidate passed 79 real CLI invocations in 145.04 seconds. The test proved actual chat tool results and usage, branch/restart history, paused queue/resume, planner-created intent, worker WebFetch, a synthetic LOW finding and a report. The sole loopback endpoint received exactly one GET; zero owned processes remained. [Receipt](evidence/standalone-cli/codex-oauth-e2e-initial.json). This candidate predates the final provider isolation changes, so a final-package repeat is required.
- Independent re-review closed four provider findings: endpoint override isolation, authentication environment filtering, quota-versus-rate-limit classification, and announced recoverable retries. Provider tests/vet and fresh-database profile/schema tests passed. Race instrumentation could not build with the installed 32-bit Windows C compiler.
- The first supplied key returned insufficient_quota for GLM-5.3/max, not a successful inference. The user identifies the key as a Coding Plan key. A direct built-in-provider call, without the diagnostic relay, also returned usageLimitExceeded with zero tokens. Z.ai changed key management on October16; Standard keys do not use Coding Plan quota. [Sanitized receipt](evidence/standalone-cli/glm-53-recheck.json), [official key guide](https://docs.z.ai/devpack/notice/api-key-management).
- The user approved automatic first-use downloads of pinned PostgreSQL/runtime files into the portable folder. The public archive excludes those runtime binaries. Existing bundled packages remain local historical test artifacts; they must not be published with unresolved redistribution rights.
- The user approved the staged documents for agy. The landing page was generated with exact gemini-4-argon-medium at medium effort, then checked independently. Both agy calls timed out before a final summary while producing files; the coordinator finished the remaining technical fixes. Initial confirmation passed 148 browser assertions across six EN/KO layouts. Updated factual copy and integration have a separate check.
- A replacement credential subsequently completed built-in GLM-5.3/max inference and an independent official Codex GLM-5.3-Flash/low probe. This changes the earlier key-failure status; full GLM workflow acceptance is separate. [Probe receipt](evidence/standalone-cli/glm-replacement-key-success.json).
- Final Windows download-mode candidate64ba8d29 passed93 independent CLI invocations, concurrent two-home cold setup, four-client startup, authentication, crash adoption, offline full-folder relocation,1608unchanged hashes and real junction guards;0owned processes remained. [Receipt](evidence/standalone-cli/bootstrap-agent-e2e.json).
- Linux native-filesystem WSL testing passed real first-use download in8.012seconds and62installed E2E invocations in11.515seconds;0owned processes remained. WSL Windows-mounted drvfs lacks the required atomic no-replace rename; this is diagnosed explicitly without weakening safety. [Receipt](evidence/standalone-cli/linux-first-use-e2e.json).
- The next OAuth run exposed a fixture mismatch: an IP-only asset cannot satisfy its strict exact-URL allowlist. The fixture now registers and attaches the owned endpoint before planning. The original failure is retained and scope protections remain unchanged. [Explanation and counterexamples](evidence/standalone-cli/subscription-e2e-scope-fixture.md).
- Final subscription workflow repetition, native platform/hosted CI evidence, external integrations and human review remain release gates. No remote, public tag or release has been created.

- Independent later campaigns exercised 54 non-AI CLI invocations and 146 advanced OAuth invocations. They verified authentication changes, notification metadata without delivery, exact 400,011-byte traffic download, task chat, side questions and a fresh-fetch finding retest. Deepen completed a worker but did not repeat the fetch because the existing constraint forbade it. These are version-specific results, not final-candidate acceptance. [Local features](evidence/standalone-cli/local-feature-e2e.json), [advanced OAuth](evidence/standalone-cli/advanced-oauth-e2e.json).
- These campaigns exposed two defects: native Windows stream interruption returned exit1 instead of130, and shutdown closed the database before admitted background chat finalizers settled. The stream regression now passes in source; the bounded chat-drain fix and rebuilt-package checks are in progress. [Stream regression](evidence/standalone-cli/stream-cancellation-regression.json).
- The GLM full workflow reached real chat tool use and goal setup, then rejected a planner call whose arguments encoded an array instead of the required input object. The bridge keeps strict rejection and now makes the complete-object contract explicit. A synthetic GLM probe accepted the schema constraint and returned a valid object in2.96seconds; full planner verification is still pending. [Synthetic format probe](evidence/standalone-cli/glm-argument-pattern-probe.json).

## Earlier release preparation (superseded where noted)

The first public v0.1.0 release is **not ready**. These results supplement the historical baseline below; no release or remote was created.

- An independent agent ran 321 real CLI invocations against c86c7e4: 205 of 297 operations were invoked, including 193 positive and 12 error-only cases. Eight more had existing-test references; 84 remained unavailable with explicit reasons. This is not an all-features-pass result.
- The agent proved task archive/restore, batch operations, persisted CRUD, Unicode/binary files, reports, restarts and local HTTP traffic capture. It found the automatic HTTP recording proxy incorrectly prefixed a complete URL. The engine fix preserves the URL; a functional regression failed before the fix and passed afterward. Direct and explicit proxy modes form the controls.
- Fixed catalog required-field metadata and descriptions from actual handler/runtime evidence, including asset references, LLM records, upload destinations, agent updates, finding follow-up and archive batches. Conditional fields remain conditional; trigger/agent replacement behavior is described.
- Fixed startup recovery accepting a file as transcript storage on Windows. The existing regression now passes. The installed executable also rejects the invalid storage, preserves its contents and starts after the fixture is moved aside.
- Fixed the usage-recording test fixture to initialize the same ledger table production startup initializes. The fresh-database llmrec suite passes all 23 tests. The Windows server suite after the startup fix reports 492 passed, 1 failed and 4 skipped; the remaining failure is Windows symlink privilege, not a passing result.
- Linux amd64 packaged E2E on an earlier dirty build (before the final catalog and recording-proxy fixes) under Ubuntu WSL2 passed all 62 invocations using real bundled PostgreSQL17.11. The exact symlink regression passed without a skip; zero owned processes remained. Windows-built Linux files needed executable permissions restored before running. This is not a clean Linux VM or complete Linux engine-suite test.
- Official Codex CLI0.162.1, already logged in through ChatGPT, completed a synthetic authentication smoke test with gpt-6.1-sol and no tool calls. This verifies the login route only; built-in ScopeWeaver AI has no subscription adapter yet and was not tested through this login. Z.ai documents [GLM Coding Plan through Codex](https://docs.z.ai/devpack/tool/codex); the present ZAI_API_KEY failed that official Codex route. Z.ai returned JSON error code401, "token expired or incorrect", under HTTP200; Codex reported an incomplete stream. No GLM inference completed. See [sanitized provider receipt](evidence/standalone-cli/glm-coding-plan-probe.json).
- Portable folders now include both provenance documents and the verification evidence directory referenced by their READMEs.

The corrected Windows package then passed all 43 independent-agent regression invocations, including 9 required-field failures before runtime creation and both captured proxy exchanges. The focused run passed 62 tests/subtests including installed E2E, with 0 failures and 10 Unix-specific updater skips; changed-package vet passed. See [current test counts](evidence/standalone-cli/release-followup-tests.json), [baseline operation coverage](evidence/standalone-cli/agent-e2e-baseline.md), and [corrected-package regression](evidence/standalone-cli/agent-e2e-regression.md).

The updated Linux candidate, including the final catalog and recording-proxy fixes, also passed all 62 packaged E2E invocations and all four requested filesystem/HTTP regressions (0 failures/skips), leaving 0 owned processes. See [final WSL receipt](evidence/standalone-cli/wsl-final-summary.json). WSL and clean-host limitations still apply.

Release remains gated on the subscription-AI scope decision and implementation/live E2E, the requested agy landing page and QA, native target/hosted CI evidence, human review and dependency redistribution evidence. The bundled Windows CRT grant and matching LGPL source/build inputs remain unresolved. Microsoft describes its [redistribution requirements](https://learn.microsoft.com/en-us/cpp/windows/redistributing-visual-cpp-files?view=msvc-170). Inclusion in a local test package is not permission to publish it.

At this earlier stage automatic approval review rejected agy document sharing. The user subsequently approved the exact staged files, and the calls described above ran.

## Earlier standalone baseline

- Catalog:297 named operations,857 parameters. Source tests cover every registered API route, generated side-question routes and both chat-control dispatch trees.287 JSON responses,5 downloads,4 event streams,1 text response.
- Focused suites:61 passing tests/subtests across CLI(25), runtime(17), packaging(13), updater(6).11 skips: the packaged E2E opt-in was run separately;10 inherited updater checks require Unix shell/executable semantics on this Windows run.
- Installed packaged executable E2E:62 invocations from an unrelated temporary directory. Real bundled PostgreSQL; discovery, reads across domains, tasks, goals, constraints, companies, agents/conversations, Unicode file writes, binary upload/download integrity, Markdown/ZIP output, failure exit codes, pause persistence, restart and relocated home.
- Full portable folder: copied package, launched with default home from C:\Windows\Temp and a Windows-system-only PATH, wrote Unicode/multiline content, stopped, moved the complete folder into a Unicode path, read the same content, and stopped.0 owned processes remained.1,573 original package file hashes matched before and after. Final packaging also includes both READMEs and paired CLI/verification docs.
- Runtime lifecycle: four simultaneous startup clients reused one runtime PID; client exit preserved runtime; auth-protected health/init rejected unauthenticated requests401; inherited database/config/skill settings did not select another installation; owned runtime crash recovered the authenticated surviving database.
- Original database suite:336 passed,0 failed/skipped. Original agent suite:534 passed,0 failed/skipped. Each suite used a separately created scopeweaver_test_* database and dropped it afterward.
- Independent review: fixed wrong Linux post-update executable path, runtime dry-run mutation, REPL Windows backslash loss, REPL stdin buffering and interrupted skill-seed publication. Re-review passed.
- Changed-package go vet passed. Runtime Linux/Darwin cross-compilation passed. Linux PostgreSQL17.11 executed under existing WSL; macOS arm64 archive extraction passed.

Additional engine packages:19 packages checked,17 passed and2 had no tests;291 top-level tests passed (545 including subtests). Eight skips included seven database-dependent tests, rerun separately with fresh databases, and one Windows symlink-privilege check.

## Earlier failures and limits (updated above)

Additional database-dependent suites: evidence17 passed with no failures/skips; llmrec22 passed and1 failed with no skips. TestSideUsageRecordedOnceOnConsumerCancellation failed because relation llm_usage was absent (SQLSTATE42P01); the exact test also fails against the untouched original source. This is a third inherited test failure, not a passing check. Both test databases were dropped. See core-evidence-llmrec-evidence.json and local baseline-core-llmrec-test.jsonl.

The original server suite reported491 passed,2 failed,4 skipped. Both failures also reproduced against the untouched original checkout (observed HEAD5681d41334804bb9411a85632bf5aefbb4f9cd70). The new CLI did not change their source or weaken the checks:

1. TestChatBranchCopyRejectsDestinationSymlink: Windows denied symlink creation, “A required privilege is not held by the client.”
2. TestChatStartupRecoveryFailureStopsConstruction: the existing constructor returned a usable server when the transcript path was a file.

Skipped server checks: TestAgentAdapterE2E, TestChatControlsE2E, TestLiveContextReview and TestTaskArchivePackageSkipsSymlink. No claim of an entirely passing inherited server suite is made.

A clean Windows VM and full Linux/macOS CLI runtime E2E were not run. Current-host tests with an isolated PATH establish independence from the source checkout and separately installed application/database; they do not establish all possible OS baselines. Hosted CI is configured but has not run because this repository has not been published.

External LLM, ScopeSentry, notification delivery, custom executables and MCP integrations remain configuration-dependent. Catalog parity plus existing engine tests is not evidence that every external service was contacted. No stable CLI release exists yet, so online update-download execution is not verified; updater unit checks and the corrected relaunch path were verified.

## Evidence and reproduction

[Evidence directory](evidence/standalone-cli/) contains the packaged E2E log, lifecycle/relocation receipts, core-suite counts and independent-review record. Detailed local JSON logs remain under dist/core-*-test.jsonl, dist/baseline-core-server-test.jsonl, dist/focused-tests.jsonl.

    go test ./internal/cli ./internal/portable ./internal/runtimebundle ./packaging ./selfupdate ./codexprovider -count=1
    go vet ./internal/cli ./internal/portable ./internal/runtimebundle ./packaging ./selfupdate ./codexprovider
    go run ./packaging -target windows-amd64
    # Set SCOPEWEAVER_CLI_E2E_BINARY to the package's executable.
    go test ./internal/cli -run '^TestPortableCLIEndToEnd$' -count=1 -v

Preserve the limitations above when describing this build.

Verbatim third-party license files retain publisher bytes and hashes, including their original whitespace. The broad Git whitespace check flags those files; the source-only whitespace check passes. Git attributes disable line-ending conversion for the pinned license files.
