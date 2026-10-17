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

Private repository: [cskwork/scopeweaver-cli](https://github.com/cskwork/scopeweaver-cli); [draft PR #1](https://github.com/cskwork/scopeweaver-cli/pull/1) targets dev and requires repository access. GitHub release archives and SHA256SUMS, with full portable folders, dependency notices, source provenance and file manifests. The remote exists; no merge, public tag or release is recorded. Five build targets exist; publication requires target-specific verification. Existing GUI build/release scripts are not CLI release automation.

## Release Notes Strategy

Describe actual standalone CLI features, supported model authentication, tested platforms, restrictions and limits. Preserve AGPL/upstream attribution. No claim of all-feature success from command-catalog parity or configuration-only tests. Paired release drafts are prepared at [docs/releases/v0.1.0.md](../docs/releases/v0.1.0.md) and [docs/ko/releases/v0.1.0.md](../docs/ko/releases/v0.1.0.md). Draft preparation does not establish release acceptance or authorize publication.

## CI Workflow Files

`.github/workflows/cli.yml`: focused tests/vet, native package and installed E2E on Windows, Linux and macOS. Triggered by PR, dev/main push or manual dispatch. Does not publish. Hosted run 39011629129 passed Ubuntu but exposed English Windows Unicode PostgreSQL-path and macOS /var-alias fixture failures. The 80f9016 fixes passed all three Windows, Ubuntu and macOS 26 arm64 jobs, including installed E2E, in run39012848998. Later source uses the latest PR checks as its final-head status. The updated workflow extracts the generated ZIP, restores Unix executable modes from archive metadata and tests that extracted executable; artifacts are retained only after the job passes. The workflow change is not itself a final-head hosted pass.

## Current Release Gates

- Release remains held under the user's all-pass condition: R7 Flash acceptance is unresolved; final-head CI and human review also remain. Clean R5 at ad814451 passed Windows core 62 calls / 60.46s, OAuth 137 / 205.12s and GLM-5.3-Flash/low 123 / 172.47s. Each real provider run covered chat/tools, queue, planner/worker, goal completion, finding/report and one owned GET.
- Historical flagship GLM-5.3 R5/low run at 2026-10-17 07:29:59 UTC failed before tool execution with an argument-object validation error (17,933 input / 25 output tokens; raw shape unknown). One extra fresh Z.AI-only correction now preserves strict decoding, rejects the full first message, shares a 10-minute deadline, rechecks policy/cancellation and sums all four usage fields. No network/quota/protocol/unknown-tool/duplicate-key correction is permitted. Provider red/green tests, vet and independent review passed; see docs/evidence/standalone-cli/glm-format-repair.json and glm-format-repair-review.md.
- R7 focused checks passed 339 test events, 0 failed, 18 skipped, plus vet; installed/paid-model checks are separate and platform/opt-in skips remain explicit.
- Rebuilt Windows R7 ChatGPT OAuth passed 131 installed calls / 163.68s, including real chat Read, queue/history, planner/worker, goal completion, exactly one owned GET, finding/report, persisted usage and successful cleanup. See docs/evidence/standalone-cli/candidate-r7-verification.json; final clean-head native CI is separate.
- Rebuilt Windows R7 flagship GLM-5.3/low passed 119 installed calls / 138.50s, including real chat Read, queue/resume, planner/worker, goal completion, exactly one owned GET, finding/report and persisted usage; cleanup succeeded. Zero correction attempts occurred: the pass verifies this current workflow, not live repair-branch recovery. See candidate-r7-verification.json. R7 includes later attachment/provider fixes; final clean-head native CI remains pending.
- R7 Flash remains unaccepted. Initial low failed after 136.98s at the exactly-one-finding assertion; actual count was not captured. Low diagnostic failed after 286.06s with two distinct report_finding calls (intent IDs 3/6, finding IDs 1/2) for the same owned marker/name. High reasoning failed after 482.38s: one worker WebFetch and one finding were recorded, but the final planner had not finished at the unchanged seven-minute task budget. All three runs had zero correction attempts. Underlying model/provider cause unknown; no broad engine/deduplication change. Preserve these failures separately from historical R5 Flash success; public release remains held.
- R5 Linux on native-ext4 WSL passed cold download/start in 6.466s and 62 calls in 10.008s, with zero owned processes and unchanged manifest. The initial missing-execute-bit harness failure is retained; retry used a fresh native copy and unchanged binary bytes. This was not a native Linux build or clean VM.
- Source 80f9016 is newer than R5. Windows PostgreSQL paths/cwd use verified ASCII 8.3 aliases for Unicode; unavailable aliases fail explicitly with ASCII-folder/home guidance. ASCII paths need no 8.3 support. Mac production guards are unchanged. Reproduced WSL fixture checks: 83 passed / 0 failed / 0 skipped. Windows filesystem checks: 81 passed / 4 symlink-privilege skips. Both Darwin architectures cross-compiled. Focused/notification checks: 313 passed / 0 failed / 18 skipped; vet and narrow independent source review passed. Hosted Windows, Ubuntu and macOS 26 arm64 integrated jobs passed at 80f9016 in run39012848998; latest PR checks govern later heads.
- Catalog union at 2026-10-17T07:21:13Z: 297 invoked / 297 positive / 0 error-only / 0 untested of 297. This spans candidates and means at least one asserted behavior, not all options/states or final certification. The historical R2–R6 union does not certify every option/state, live vendor or final native artifact. Do not sum call totals as unique commands.
- R4 owned fixtures proved 8 integration contracts, including MCP/ScopeSentry persistence, compatible model discovery and DeepSeek SSE parsing; no live-vendor or inference claim. R5 blocked recovery used 64 calls: genuine HTTP400, same intent blocked, 1 reopened, completed fact; no injected DB state. R6 notification used 29 calls: default/false blocked, explicit ARTEX_NOTIFY_ALLOW_LOCAL opt-in allowed one test send, actual finding delivery HTTP400 then same-ID retry sent HTTP200. Zero external sends/processes. Retain default blocking and explicit opt-in.
- Earlier lifecycle 130 calls plus 3 preflights and compaction 33 calls / 2 new completions retain their own receipts and limits. Per-call compaction paid records persisted; dashboard aggregates did not increase.
- All 3 installed updater commands passed in 20 calls against an owned nonforwarding GitHub-shaped HTTPS fixture with per-process CA: unknown CA rejected, pinned check/download/hash verification, actual apply/restart and rollback with same-home data preserved. No live GitHub release-cycle pass is claimed. Final-head native CI, human review and publication remain gates; consult latest PR checks.
- Public packaging uses approved first-use pinned runtime downloads; no separate app/database installer. Do not publish historical bundled Windows archives with unresolved redistribution rights.
- Exact agy gemini-4-argon-medium landing design retained. Historical 148/116-assertion campaigns do not certify this updated page; final bounded content QA has separate evidence after fact freeze.
- Original shared fixes [PR #38](https://github.com/cskwork/scopeweaver/pull/38) is ready and unmerged: all 7 required CI checks at 2273d50, 9/9 QA scenarios and 39 tests passed. This is separate from private CLI draft PR #1 targeting dev. No public tag, release or merge is recorded.

- Later six-line upstream attachment-manifest follow-up 58a79fe passed red-before/green-after checks: 31 test/subtest events (18 DB, 7 HTTP, 6 adjacent), 0 failed/skipped and vet. Full fresh-database suite: 230 top-level tests / 355 test events passed, 0 failed/skipped; the full Windows server suite was not repeated. It is newer than dated 80f9016; final-head native CI comes from latest PR checks. Final native CI package snapshots remain runtime-free and include hashes; no public release is claimed.
