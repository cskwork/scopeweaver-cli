# Release Rules
<!-- last-analyzed: 2026-10-18 -->

## Version Sources
- cmd/scopeweaver-cli/main.go: var version = "0.1.0" (linker override).
- internal/cli/cli.go: var Version = "0.1.0" (main overrides at startup).
- First public minor release is v0.1.0. Already set; no bump to v0.2.0.
- No current CLI tags. Inherited frontend/adapter versions are not separate publication targets.

## Release Trigger
Manual GitHub release after the user's all-pass condition. No automatic publishing workflow, intentionally.
User explicitly authorizes making scopeweaver-cli PUBLIC AT RELEASE. Do not ask again for visibility permission.
Route feat/standalone-cli -> dev -> main. CONTRIBUTING requires maintainer authorization and human review; publication authorization does not certify failed tests or replace that final review.
Original ScopeWeaver PR40 is a separate PR-only deliverable, not authorized for merging here.

## Test Gate
- Focused tests/vet for internal/cli, internal/portable, internal/runtimebundle, packaging, selfupdate, codexprovider; full clean-source go vet ./....
- Five unchanged packaged Windows workflows: core, GLM-5.3/low, Flash/low, OAuth/low, Flash/high.
- Keep one existing correction, shared ten-minute completion budget and 420-second task budget; no model/effort or assertion substitutions.
- Fresh disposable DB engine tests, independent agent workflows and candidate-specific Linux race evidence; never user databases.
- Native clean-head Windows amd64, Linux amd64, macOS arm64 ZIP extraction, first-use/installed core E2E and exact manifest/hash checks.
- Landing EN/KO at320/390/1440, controls, links/assets and claims QA; preserve requested agy gemini-4-argon-medium design.

## Registry / Distribution
Currently PRIVATE cskwork/scopeweaver-cli, default main, draft PR1 -> dev. Make PUBLIC at release under explicit user authorization.
Publish verified Windows amd64, Linux amd64, macOS arm64 ZIPs and SHA256SUMS with notices/provenance/manifests.
Default runtime-free packages use approved pinned first-use downloads. Never publish bundled-runtime test clones.
Core needs no separate ScopeWeaver/database installation; optional subscription AI needs official Codex/account or Coding Plan credentials.
No npm/PyPI/Docker publication in scope. Never include keys, private state or user data.

## Release Notes Strategy
Paired docs/releases/v0.1.0.md and docs/ko/releases/v0.1.0.md. Describe actual features, tested platforms/authentication, downloads and limits; preserve AGPL/attribution.
The historical297-operation ledger proves one positive contract per name, not every option/state or final artifact. Historical passes never certify later artifacts.

## CI Workflow Files
.github/workflows/cli.yml: PR, dev/main push, manual dispatch; three native OS jobs. Focused tests/vet, package, ZIP/hash, extraction with Unix modes, first-use health and installed core E2E; upload only after pass. No publishing job.
Final ZIP check: exact three targets, one clean source revision, SHA/CRC, exact manifest allowlist/every payload hash, Unix executable mode, no runtime/private-state paths.

## Current Release Gates
- R24 acceptance PASS 5/5 on the frozen dirty Windows source snapshot based on650dc82910e5e36668161a5276455506f64713fe. Four production hashes and binary8f91bbb7d1ce764f890e699c2fa32400405e5a81c4cbf03bc22c6bff0da83a77 are in docs/evidence/standalone-cli/candidate-r24-verification.json.
- Core62calls/19.14s; GLM-low106/82.63; Flash-low117/114.84; OAuth-low115/108.12; Flash-high182/317.83. All native exits0, unchanged420-second task and600-second shared completion budgets, three ZAI correction counts0; OAuth correction N/A.
- Runtime-free public manifest181payload files; local test clone1482runtime files copied andSHA-verified. Publish only runtime-free archives with first-use pinned downloads.
- Focused186/0/19, focused/fullclean vet0 and packaging0. CLI agent607/0/0, original598/0/0, exact formatted new actual-planner tests31/0/0. Native Linux focused race49/0/0 and fullagent race607/0/0; native exits0, no race report,1175snapshot hashes unchanged, ownedcluster stopped/noPID. Controlled regression-provider usage synthetic; actual usage comes from real installed CLI workflows.
- Locale1986production keys x3languages complete. Landing validator33/0; negative-fixture testevents19/0/0. Legacy single-page validator fixed and.nojekyll present. Final text-only landing update retains107EN/107KO elements and design assets; fresh browser QA passed EN/KO at320/390/1440, controls and assets/links with no overflow or console/page errors.
- Earlier checkpoint650dc82 passed all3nativeOS CI/jobs and archive checks; historical pass does not certify R24/finalcleanhead. New final CLI head native CI/archives are PENDING.
- Original PR40 shared correction committed and pushed at247219ee623757a313796557a22ee64bb1b4905e; new-head7CI checks/humanreview PENDING. Earlierffe260435fed6576394d613991aaf6e8010d3f48 green checks are historical only. Original PR40 remains PR-only/unmerged here. Original PR38 merged2026-10-17T08:06UTC.
- R22 and older failed outcomes/receipts remain historical in paired verification docs; no earlier pass certifies a later artifact. Windows inherited symlink-privilege failure reproduces unchanged650dc82 and is not weakened.
- Finalcleanhead native CI/archives, maintainerhumanreview, CLI integration/promotion, postmergeE2E and manualpublicrelease remain open. Publicvisibility at release already authorized; do not ask again. No public tag/release claimed.

## First-Time Setup Gaps
Manual publication is intentional. No tags until firstv0.1.0 release. dist is ignored; .omc/RELEASE_RULE.md is tracked. No release workflow scaffolding needed.
