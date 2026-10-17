# Corrected-package agent E2E regression

PASS: 43 actual packaged CLI invocations, all expected outcomes met.

- Binary: `D:\Git\scopeweaver-cli\dist\release-recheck-20261017\scopeweaver-cli.exe`
- SHA-256: `dec0b7b7e3dc9b39b8c9343f30d6551008cc5076d8cd0c8fdbc0db0522e0d240`
- All 9 newly required fields rejected omission with exit 2 before runtime/home creation: agents update name; archives delete/restore batch archive_ids; chat upload scope/id; findings deepen description; llm records by-model/delete task; tasks asset-refs list asset_id.
- Positive required-value requests verified task asset references, empty LLM record aggregates/deletion, agent metadata read-back, session attachment byte count, real archive batch restore/deletion and preserved paused task state.
- Direct, explicit-proxy and automatic-recording-proxy requests all succeeded against the agent's own loopback HTTP fixture. Exactly 2 proxied exchanges were captured and their response marker read back through CLI.
- findings deepen with description reached the backend and rejected a nonexistent finding; no inference ran.
- Owned runtime stopped; 0 matching runtime/PostgreSQL processes remained.

Original c86c7e4 baseline remains preserved separately: 321 invocations, 205/297 operations invoked, including the before-fix recording-proxy defect. This regression adds proof of correction, without rewriting the baseline or claiming provider/integration paths passed.

No provider inference, external target, notification send, installation or publication occurred.
