# Independent final CLI source review

Date: 2026-10-17
Candidate: ad814451c9a69aa9ffd9c61c84b484a9f7165475
Comparison base: c86c7e4a78d52cd3831de41494792fa16f6f145a
Verdict: PASS within the source-review scope. No concrete blocking correctness finding was identified.

This is an independent source review, not final-binary acceptance, human review, release authorization, or an all-features-pass claim. No production files changed, tests rerun, live inference, credential reads, external service calls, commits, or publication occurred. Only this ignored review receipt was written. The worktree was clean at review start and remained clean after source inspection. Source-only git diff --check passed.

## Reviewed contracts

- Read CONTRIBUTING.md, AGENTS.md, global rules, prior release-followup-review.txt and chat-shutdown-review.md, and current paired-contract documentation. Prior provider findings and shutdown fixes were considered rather than reopened without evidence.
- codexprovider/provider.go, isolation.go, protocol.go and catalog.go: configuration-only construction; pinned OAuth/Z.AI provider routes; filtered authentication/endpoint overrides; intentional OAuth home retention versus disposable Z.AI home; disabled direct tools/plugins/apps/hooks/skills; enumerated MCP isolation and status verification before inference; denial of unexpected app-server requests or tool items; fresh ephemeral thread per completion; strict complete action-envelope parsing, duplicate-key/UTF-8/object checks and bounds; one cumulative usage snapshot; safe quota/rate errors and recoverable retry handling; bounded process cancellation and cleanup.
- GLM-5.3-Flash is included in the exact accepted Z.AI model list and catalog alongside GLM-5.3. Disabled thinking maps GLM to low; OAuth compaction retains configured effort instead of forcing unsupported none. Existing tests explicitly cover both cases and both catalog entries.
- internal/runtimebundle plus internal/portable/postgres.go: HTTPS and checksum boundary; capped downloads and staging; path-safe extraction rooted in os.Root; no overwrite of existing runtime/cache entries; atomic no-replace publication on Windows/Linux/Darwin; executable-root install lock shared by distinct homes; preservation of incomplete existing installs and actionable native-filesystem diagnostics.
- internal/cli/cli.go cancellation classification consults caller context as well as the transport error, preserving native Windows exit 130 when interrupted pipe reads report another error. No newly shipped debug/probe command was found in production command sources or the operation catalog.
- server/chat_shutdown.go and chat/side admission, dispatch and publisher integration: admission barriers precede draining; chat WaitGroup additions serialize with closure; captured side runs complete terminal persistence before done; publisher completion precedes final snapshot/context flush; final write errors propagate and repeated drains preserve the same failure. Runtime cancels and drains before manager/database teardown, with bounded failure reported explicitly.
- internal/portable/runtime.go: authenticated health identity includes shutdown-result version; Stop requires the matching nonce-bound result for version 1 after owner release; missing or stale receipts fail; legacy behavior remains compatible. Receipt publication occurs after drain/database teardown and before ownership release.
- agent/provider.go, db/config.go, db/schema.sql and profile handlers: provider construction and readiness integrate the adapter; credential retention is restricted for Codex transitions; additive format constraint preserves existing formats and storage. Profile save does not invoke inference.
- Proxy URL and invalid transcript-directory fixes follow their existing contracts with functional regression oracles. Packaging uses the shared runtime pins and includes provider documentation and manifests; public mode omits bundled runtime bytes.

## Verification and qualification

Reviewed the existing tests for cancellation, endpoint/auth isolation, GLM metadata/home cleanup, structured arguments, OAuth/GLM effort, no-replace install behavior, final flush write failure and nonce receipt mismatch. I did not independently rerun tests or final packaged E2E. The tracked final-source receipt reports 132 passes, 0 failures and 18 named platform/live-opt-in skips, plus focused vet and fresh-database regressions; those remain implementation evidence rather than tests run by this review.

Catalog parity remains 297 named operations. The coverage union distinguishes 284 invoked names, 283 positive contracts, 1 error-only/unsupported and 13 untested; it does not equate invocation totals with unique coverage. Current docs distinguish OAuth R3 and Flash R4 successes from the GLM-5.3 final planner timeout, unproven aggregate compaction accounting, historical binaries and remaining external/update/platform release gates. No contradiction requiring a source blocker was found.

Final rebuilt-package checks are separate acceptance work. This PASS does not clear publication, hosted CI, human-review, live integration or native macOS gates.
