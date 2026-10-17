# GLM Coding Plan diagnostic receipt

Historical first-key failure, recorded 2026-10-17. A replacement key later succeeded with built-in GLM-5.3 and official Codex GLM-5.3-Flash; see glm-replacement-key-success.json. No API key, token, request prompt or authorization header is included.

The owner identifies the key as a Personal Coding Plan / Team Plan key. Both official client routes selected glm-5.3-flash. A local diagnostic relay verified that the authorization value matched the supplied key and forwarded the official client's request unchanged to the documented Z.ai host.

| Client | Route | Result |
| --- | --- | --- |
| Codex CLI0.162.1 | https://api.z.ai/api/v1/responses | HTTP200 event stream, insufficient_quota, no inference |
| Claude Code | https://api.z.ai/api/anthropic/v1/messages?beta=true | HTTP429, code1113, insufficient balance or no resource package |

Codex request ID: 2026101712480204c038d1e1774b82.
First Claude request ID: 2026101712515308038e7328ed40a5.
Claude retried the same upstream error until the bounded probe stopped it; it did not complete a model response.

GLM-5.3/max also failed through official Codex and a direct built-in provider call without the relay. The built-in call returned usageLimitExceeded with zero reported tokens.

This narrows the failure beyond the ScopeWeaver adapter but does not establish the underlying account cause. Z.ai support can use the request IDs to investigate plan/seat entitlement and quota routing. No purchase, subscription change or global credential configuration was made.

Official setup: https://docs.z.ai/devpack/tool/codex and https://docs.z.ai/devpack/tool/claude.
Key rules: https://docs.z.ai/devpack/notice/api-key-management.
