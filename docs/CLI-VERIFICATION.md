# Standalone CLI verification

[한국어](ko/CLI-VERIFICATION.md)

Recorded 2026-10-17 on Windows amd64. Engine source pinned to ScopeWeaver commit 2df0b307dcb817a459abebd3a4c240a2019598db. Tests use disposable local state and PostgreSQL17.11; no real targets, paid model calls or external notification sends.

## Passed

- Catalog:297 named operations,857 parameters. Source tests cover every registered API route, generated side-question routes and both chat-control dispatch trees.287 JSON responses,5 downloads,4 event streams,1 text response.
- Focused suites:61 passing tests/subtests across CLI(25), runtime(17), packaging(13), updater(6).11 skips: the packaged E2E opt-in was run separately;10 inherited updater checks require Unix shell/executable semantics on this Windows run.
- Installed packaged executable E2E:62 invocations from an unrelated temporary directory. Real bundled PostgreSQL; discovery, reads across domains, tasks, goals, constraints, companies, agents/conversations, Unicode file writes, binary upload/download integrity, Markdown/ZIP output, failure exit codes, pause persistence, restart and relocated home.
- Full portable folder: copied package, launched with default home from C:\Windows\Temp and a Windows-system-only PATH, wrote Unicode/multiline content, stopped, moved the complete folder into a Unicode path, read the same content, and stopped.0 owned processes remained.1,573 original package file hashes matched before and after. Final packaging also includes both READMEs and paired CLI/verification docs.
- Runtime lifecycle: four simultaneous startup clients reused one runtime PID; client exit preserved runtime; auth-protected health/init rejected unauthenticated requests401; inherited database/config/skill settings did not select another installation; owned runtime crash recovered the authenticated surviving database.
- Original database suite:336 passed,0 failed/skipped. Original agent suite:534 passed,0 failed/skipped. Each suite used a separately created scopeweaver_test_* database and dropped it afterward.
- Independent review: fixed wrong Linux post-update executable path, runtime dry-run mutation, REPL Windows backslash loss, REPL stdin buffering and interrupted skill-seed publication. Re-review passed.
- Changed-package go vet passed. Runtime Linux/Darwin cross-compilation passed. Linux PostgreSQL17.11 executed under existing WSL; macOS arm64 archive extraction passed.

Additional engine packages:19 packages checked,17 passed and2 had no tests;291 top-level tests passed (545 including subtests). Eight skips included seven database-dependent tests, rerun separately with fresh databases, and one Windows symlink-privilege check.

## Inherited failures and limits

Additional database-dependent suites: evidence17 passed with no failures/skips; llmrec22 passed and1 failed with no skips. TestSideUsageRecordedOnceOnConsumerCancellation failed because relation llm_usage was absent (SQLSTATE42P01); the exact test also fails against the untouched original source. This is a third inherited test failure, not a passing check. Both test databases were dropped. See core-evidence-llmrec-evidence.json and local baseline-core-llmrec-test.jsonl.

The original server suite reported491 passed,2 failed,4 skipped. Both failures also reproduced against the untouched original checkout (observed HEAD5681d41334804bb9411a85632bf5aefbb4f9cd70). The new CLI did not change their source or weaken the checks:

1. TestChatBranchCopyRejectsDestinationSymlink: Windows denied symlink creation, “A required privilege is not held by the client.”
2. TestChatStartupRecoveryFailureStopsConstruction: the existing constructor returned a usable server when the transcript path was a file.

Skipped server checks: TestAgentAdapterE2E, TestChatControlsE2E, TestLiveContextReview and TestTaskArchivePackageSkipsSymlink. No claim of an entirely passing inherited server suite is made.

A clean Windows VM and full Linux/macOS CLI runtime E2E were not run. Current-host tests with an isolated PATH establish independence from the source checkout and separately installed application/database; they do not establish all possible OS baselines. Hosted CI is configured but has not run because this repository has not been published.

External LLM, ScopeSentry, notification delivery, custom executables and MCP integrations remain configuration-dependent. Catalog parity plus existing engine tests is not evidence that every external service was contacted. No stable CLI release exists yet, so online update-download execution is not verified; updater unit checks and the corrected relaunch path were verified.

## Evidence and reproduction

[Evidence directory](evidence/standalone-cli/) contains the packaged E2E log, lifecycle/relocation receipts, core-suite counts and independent-review record. Detailed local JSON logs remain under dist/core-*-test.jsonl, dist/baseline-core-server-test.jsonl, dist/focused-tests.jsonl.

    go test ./internal/cli ./internal/portable ./packaging ./selfupdate -count=1
    go vet ./internal/cli ./internal/portable ./packaging ./selfupdate
    go run ./packaging -target windows-amd64
    # Set SCOPEWEAVER_CLI_E2E_BINARY to the package's executable.
    go test ./internal/cli -run '^TestPortableCLIEndToEnd$' -count=1 -v

Preserve the limitations above when describing this build.

Verbatim third-party license files retain publisher bytes and hashes, including their original whitespace. The broad Git whitespace check flags those files; the source-only whitespace check passes. Git attributes disable line-ending conversion for the pinned license files.
