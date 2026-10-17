# Release Rules
<!-- last-analyzed: 2026-10-17 -->

## Version Sources

- `cmd/scopeweaver-cli/main.go`: `var version = "0.1.0"`, linker-overridable.
- `internal/cli/cli.go`: `var Version = "0.1.0"`, overwritten by main at startup.
- No public CLI tags/releases exist. The first public minor version is v0.1.0; do not inherit the original GUI application's version.

## Release Trigger

Manual GitHub release after all agreed gates pass. User authorized the first public minor release conditionally on full working-function E2E, subscription AI support and landing page completion. No automatic publishing workflow exists. CONTRIBUTING requires feature/fix -> dev -> main, maintainer authorization and human review. Conditional release authorization does not waive unresolved evidence or human review.

## Test Gate

- `go test ./internal/cli ./internal/portable ./internal/runtimebundle ./packaging ./selfupdate ./codexprovider -count=1`
- `go vet ./internal/cli ./internal/portable ./internal/runtimebundle ./packaging ./selfupdate ./codexprovider ./server ./llmrec`
- Native portable package build and installed CLI E2E via `SCOPEWEAVER_CLI_E2E_BINARY` on every published platform.
- Fresh disposable `scopeweaver_test_*` databases through ARTEX_TEST_DSN for engine suites; never production state.
- Independent agent-driven functional E2E and fixes rechecked on rebuilt package.
- Live GLM Coding Plan and Codex OAuth workflows, including real internal tool use for the approved built-in AI scope. Authentication smoke tests alone are insufficient.
- No unresolved failures concealed by skips. Record environmental failures and verify on a capable target.
- Landing page built with user-requested agy `gemini-4-argon-medium`, then desktop/mobile/link/claim QA.

## Registry / Distribution

Planned repository: cskwork/scopeweaver-cli. GitHub release archives and SHA256SUMS, with full portable folders, dependency notices, source provenance and file manifests. Local repository has no remote as of this analysis. Five build targets exist; publication requires target-specific verification. Existing GUI build/release scripts are not CLI release automation.

## Release Notes Strategy

Describe actual standalone CLI features, supported model authentication, tested platforms, restrictions and limits. Preserve AGPL/upstream attribution. No claim of all-feature success from command-catalog parity or configuration-only tests. No public notes prepared until scope and acceptance evidence are complete.

## CI Workflow Files

`.github/workflows/cli.yml`: focused tests/vet, native package and installed E2E on Windows, Linux and macOS. Triggered by PR, dev/main push or manual dispatch. Does not publish. Hosted CI has not run.

## Current Release Gates

- Release is held. Windows R3 OAuth passed 125 installed invocations in 161.24 seconds; R4 GLM-5.3-Flash/low passed 127 in 153.87 seconds through real chat/tools, queue, planner/worker, one owned loopback GET, finding, goal completion and report. Flagship GLM-5.3 reached worker/finding but timed out at the final planner call; its full workflow is not accepted. Earlier key/quota and tool-format failures are historical.
- R3 lifecycle passed 130 calls plus 3 preflights for active chat/task and side-question drain/restart, steering, editing, retry and branch. R4 compaction passed 33 calls and 2 new completions with restart persistence. Paid per-call records were retained; aggregate dashboard usage did not increase.
- Windows R3 core passed 62 calls in 34.39 seconds. Linux R3 ran on WSL2 native ext4: first-use download/start 6.679 seconds, 62 installed calls 10.579 seconds. No native macOS runtime E2E or clean-host baseline is recorded. Native target evidence remains required for each published platform.
- The current exact-name catalog union snapshot records 284 invoked, 283 positive, 1 error-only and 13 untested of 297 operations. It combines candidate versions and does not prove every option/state or final-package success.
- The newest final-flush error propagation/localization fix passes source tests but is newer than tested R4. Rebuild the public candidate after documentation freeze and rerun its required installed checks.
- The approved first-use pinned runtime download mode is the public packaging path. No separate application/database installer is required. Do not publish older bundled Windows test archives with unresolved CRT/LGPL redistribution rights.
- The exact requested agy gemini-4-argon-medium landing design is retained. Initial 148-assertion QA and later factual-content QA are separate evidence campaigns.
- Required external integration evidence, online updater release-cycle evidence, hosted CI and human review remain incomplete. Windows symlink privilege prevented one historical full-server test; the exact regression passed in WSL.
- No remote, public tag or release exists. Conditional authorization does not waive these gates or CONTRIBUTING's review route.
