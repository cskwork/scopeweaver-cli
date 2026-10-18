# Final download-mode package independent E2E

PASS on the dirty candidate at c86c7e4a78d52cd3831de41494792fa16f6f145a.

Binary SHA-256: `64ba8d29e2632c7fbbb320ebc622a7fcc327c3359fb745fd7008488c798c887f`.

- 31 manual CLI invocations plus the existing 62-invocation packaged E2E: 93 total. Every CLI expected result passed. Core E2E ran from another cwd with owned disposable state and unreachable HTTP(S)/ALL proxies.
- Cold clone contained no runtime. Two simultaneous clients with different homes bootstrapped the real checksum-pinned PostgreSQL17.11 and app-local VC DLL files, then both reported healthy. No installer ran. Cold health durations: [26.393, 26.601] seconds.
- Four simultaneous clients for one home reused exactly one persistent engine PID. Unauthenticated runtime health and auth/init returned401.
- Forced crash of the owned engine left6 authenticated database processes alive; restart adopted the same database processes and read back persisted Unicode content.
- Stopped the runtime, verified no owned processes, checked both resolved move paths stayed within this run's scratch directory, and moved the complete portable folder to a Unicode path. Offline workspace and custom skill read-back succeeded. Custom skill remained absent from another home. All1608 immutable package/runtime file hashes matched after restart/relocation.
- Actual Windows runtime-parent and bin-parent directory junctions were rejected with exit3 before download/target write. Both target sentinel files remained byte-identical.
- Source package126files remained unchanged and its runtime directory is still absent. Final process inventory found0 owned runtime/PostgreSQL processes.

Evidence: `summary.json`, `calls.jsonl`, `portable-e2e.log`, `portable-e2e.json`, `redirect-guard.json`. Harness assertion corrections are retained in summary.json: normalize PostgreSQL Windows forward-slash paths and scope redirect cleanup separately from a concurrent lifecycle test. Neither was a product failure.

Limits: current Windows host, no clean VM or full macOS/Linux package E2E in this run. No provider inference, external targets, notifications, system installation or publication. Runtime pins were enforced by the successful bootstrap verifier; its temporary source archives are not retained.
