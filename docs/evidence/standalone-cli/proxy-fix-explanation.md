# HTTP recording proxy failure and correction

Environment: c86c7e4 Windows portable CLI, independently exercised on 2026-10-17 against an owned loopback fixture. Introduction commit upstream is unknown; the faulty line exists in the pinned source and this baseline.

Normally Manager.ProxyAddr returns the complete proxy URL. The HTTP custom tool passes it to net/http, and the recording proxy captures the outgoing request before forwarding it to the fixture. The custom tool added another scheme, so net/http tried to resolve a host named http instead.

Before, server/customtool.go:

    proxyStr = "http://" + addr

This concatenates a scheme with the manager's already complete URL. A real fixture value was `addr = "http://127.0.0.1:61347"`, yielding `http://http://127.0.0.1:61347` and a DNS lookup for `http`.

After:

    proxyStr = addr

This keeps the manager's URL unchanged. The paired CA selection on the next line is unchanged. The real corrected-package request succeeded; the agent read the captured response marker back through `traffic exchange get`.

Counterexample: an explicitly supplied `exec.proxy` already worked because it bypasses this automatic-proxy branch. Direct requests also succeeded. Both controls still pass after correction.

Evidence: agent-e2e-baseline.md, agent-e2e-regression.md, and TestHTTPToolUsesManagerProxyURL (failed before the fix, passed after it). No user or external targets were contacted.
