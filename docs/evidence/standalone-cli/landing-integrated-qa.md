# Integrated landing content check — 2026-10-17

**PASS for the integrated local landing page.** All 79 content/resource assertions pass; changed English/Korean copy wraps within 1440px, 390px and 320px viewports. No landing broken paths or stale subscription-adapter copy remain. This is a bounded content/integration check, not a new design audit or product-release approval.

## What was checked

Source: `D:/Git/scopeweaver-cli/landing`. CSS and JavaScript hashes exactly match the previously QA-approved snapshot. Therefore the unchanged clipboard interaction matrix was not rerun. The six language/viewport combinations were loaded from an owned loopback server using existing Playwright Chromium, captured, and reviewed through raw viewport/section screenshots and content contact sheets. No installs, external models, external page visits or publication occurred.

- All six document widths exactly match their viewports: 1440/1440, 390/390 and 320/320 in EN/KO. Changed hero, workflow, runtime/model requirements, limitations and provenance paragraphs remain within the page. Long diagnostics wrap and code surfaces retain local scrolling.
- All local scripts, stylesheet, SVG and self-hosted fonts load. No external resource requests, request failures, console warnings/errors or page exceptions were recorded. All fragment targets resolve. The planned repository remains unpublished plain text rather than a fake live release link. External upstream/reference URLs were not visited in this local-only pass.
- First-use pinned runtime downloads and one-time Internet need are disclosed in both languages. Operational startup is distinguished from discovery/diagnosis, which does not download or start the engine. These statements match current README.md's download-mode description.
- Built-in Codex OAuth support is described as using the additional official Codex CLI and account access. The 79 installed-CLI-invocation acceptance claim is supported by `dist/codex-agent-e2e-20261017/summary.json`: passed=true, official ChatGPT OAuth, real chat/branch/restart/control and planner/worker synthetic finding/report evidence. That summary explicitly describes an earlier Windows candidate, and the landing keeps final-package repetition pending rather than transferring that pass to the final artifact.
- GLM-5.3/max remains explicitly blocked by `insufficient_quota`; key classification/account entitlement stay unconfirmed and no completed GLM inference is claimed. This pass checked the coordinator-supplied status was preserved consistently in both languages; it did not perform another GLM request or inspect credential state.
- Stale provenance wording found during this check was corrected by the coordinator before captures: runtime licenses/metadata now refers to downloaded runtimes. The updated wording passes in both languages. Historical bundled-runtime E2E claims are retained as historical test evidence, with earlier archives distinguished from the final public download-mode package.
- Unreleased status, lack of public download/repository and local-isolated usage restrictions remain visible. The architecture JSON remains explicitly illustrative.

## Separate documentation observation

At evidence-read time, current README.md linked `docs/ai-providers.md`, but that file did not exist. This is outside the landing page (the landing does not link it), was reported to the coordinator, and is not counted as a landing failure. The source documentation work may create it separately. No docs/source files were edited by this reviewer.

## Evidence and limits

`evidence.json` contains 79 passing checks, exact scoped copy, resource results, widths and hashes. `*-viewport.png`, `*-full.png`, `*-requirements.png`, `*-limits.png` are raw captures; `*-content-review.png` contains diagnostic contact sheets. `process-check.json` records zero remaining QA node processes. The browser is disconnected/closed and the loopback server stopped.

The preceding 18/20 bounded design verdict is retained; this content-only pass does not rescore the interface. Physical touch, other browser engines, assistive-technology speech, product runtime setup and final-package AI acceptance were not exercised here.

## SHA-256 snapshot

- index.html: `9035ba987ef1d4e92d01e4f135363b6b3aeb1a527f4e80cbec36a19cf077f4ac`
- styles.css: `14ad3cb0bfe79fdcebd61a117bf97cf0a1e12f89c158634f17fa1a35f265be16`
- script.js: `ac31f043f637a5e47f41768d19f0a0e437f1dcabc2f4559f4dbbaf48d98e7368`
