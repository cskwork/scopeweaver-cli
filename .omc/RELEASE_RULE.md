# Release Rules
<!-- last-analyzed: 2026-10-17T13:39:54.413623+00:00 -->

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
- R22 frozen36 source SHAe1d76986da63cd9b05beb8a88978d7875f0e529be685e817435e33b2b9c3ff48 source-reviewPASS; prior28 accepted CLI files unchanged.
- Focused186/0/19 test/subtest events, focused/fullclean vet0 and packaging0. Platform/opt-in skips remain explicit.
- R22 binary970d7037b4567912fb09254821d110ce16c08a39b4772dec6e531eda1235d4c9 matrix IN PROGRESS: GLM-low165.04s/core31.28s/Flash-low157.8s/OAuth153.42s PASS; Flash-high FAILED470.50s before goalproof at unchanged420s taskbudget. R22 acceptance FAIL4/5; bounded R23 planner-context investigation underway.
- R19/R20 failures and rejected incompleteR21 preserved. R22 requests nested arguments only forZAI; OAuth prior format unchanged. Local strict validation remains authoritative; no guarantee of ZAI schema enforcement.
- Original PR40 atffe260435fed6576394d613991aaf6e8010d3f48 has all7 checksPASS (39036051785/39036051720), UI14steps and65frontendtests; OPEN/UNMERGED. Evidence .sdlc/archive/261017-glm-agent-contracts in originalworktree.
- Original PR38 merged2026-10-17T08:06UTC. Prior cache's unmerged claim was stale.
- Held landingSHA6617524d1c180f033aa23102450bba3f781f645550227ba4aa4dac80c2b21bb9 passed six viewport/language cases,107pairedspans, controls/assets/nooverflow/noerrors; ownedserver/browser stopped. Final text requires finalQA.
- CLI commit/push, final nativeCI/artifacts, maintainerreview, integration/promotion and postmergeE2E remain open. No public tag/release.

## First-Time Setup Gaps
Manual publication is intentional. No tags until firstv0.1.0 release. dist and .omc are ignored. No release workflow scaffolding needed.
