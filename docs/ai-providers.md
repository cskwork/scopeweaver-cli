# AI provider setup and limits

[한국어](ko/ai-providers.md) · [CLI architecture](cli.md) · [Verification](CLI-VERIFICATION.md)

Built-in chat, planner and worker agents can use `codex-cli` as an LLM profile format. The adapter sends text and tool descriptions to the official Codex app-server over stdio. Norma, the engine's agent harness, executes the requested ScopeWeaver tools and returns their results to the next completion. It does not delegate tool execution to Codex.

## ChatGPT login

Use an official Codex CLI installation with a ChatGPT login. The adapter was verified with Codex 0.162.1; compatibility with other or future versions is not established. Codex is an additional dependency and is not downloaded by the portable PostgreSQL bootstrap.

```sh
codex --version
codex login
codex login status
```

The browser login is managed by official Codex. ScopeWeaver does not extract, copy or send OAuth tokens itself. An OpenAI API-key login is not accepted by this adapter's ChatGPT route. Model access and limits still depend on the signed-in account. See [official authentication guidance](https://learn.chatgpt.com/docs/auth).

Resolve Codex from PATH, or set `SCOPEWEAVER_CODEX_PATH` to its native executable before starting the portable runtime. On Windows, the override must name a `.exe`, not a `.cmd` wrapper. If installed through the official npm package, the adapter can resolve its native Windows executable without invoking Node.js.

Save the following as `chatgpt-profile.json`:

```json
{
  "name": "ChatGPT via Codex",
  "format": "codex-cli",
  "model": "gpt-6.1-sol",
  "base_url": "",
  "api_key": "",
  "reasoning_effort": "high",
  "streaming": true
}
```

```sh
scopeweaver-cli schema llm profiles save --json
scopeweaver-cli llm profiles save --data @chatgpt-profile.json --yes --json
scopeweaver-cli llm profiles list --json
scopeweaver-cli llm profiles activate --id 1 --yes --json
scopeweaver-cli llm test --provider codex-cli --model gpt-6.1-sol --profile-id 1 --yes --json
```

Replace `1` with the `id` returned by save. Saving validates configuration without launching Codex or contacting the model; `llm test` makes a real model request. Leave `base_url`, `api_key` and `proxy` empty for this route. Activate the profile to use it as the default; explicit task/agent profile bindings retain their usual precedence.

## Z.AI GLM Coding Plan

The adapter supports the pinned Z.AI Responses endpoint with `glm-5.3` and `glm-5.3-flash`. It starts official Codex with a temporary clean Codex home, a provider-specific credential and the included metadata for these two GLM models. It does not alter the user's Codex configuration or use the ChatGPT login for Z.AI.

Provide `ZAI_API_KEY` in the environment of the process that starts the portable runtime. If the runtime is already running, stop it with `runtime stop --yes`, set the environment securely, then run the next operational command. Alternatively, supply `api_key` in a private profile JSON file outside the browsable workspace. Never commit credentials or include them in evidence.

Save this keyless configuration as `zai-profile.json` when using `ZAI_API_KEY`:

```json
{
  "name": "Z.AI Coding Plan via Codex",
  "format": "codex-cli",
  "model": "glm-5.3",
  "base_url": "https://api.z.ai/api/v1",
  "reasoning_effort": "max",
  "streaming": true,
  "context_window_k": 1000
}
```

```sh
scopeweaver-cli llm profiles save --data @zai-profile.json --yes --json
scopeweaver-cli llm profiles activate --id 2 --yes --json
scopeweaver-cli llm test --provider codex-cli --model glm-5.3 --base-url https://api.z.ai/api/v1 --profile-id 2 --reasoning-effort max --yes --json
```

Replace `2` with the returned profile ID. A base URL ending in `/responses` is normalized to the same root. Use `"model": "glm-5.3-flash"` in the same profile to select Flash, and pass that same model to `llm test`. Other endpoints and Z.AI models are rejected. Both models accept `low`, `high` or `max`; disabled-thinking requests map to `low`.

Follow [Z.AI's official Codex guide](https://docs.z.ai/devpack/tool/codex) and check the key type under its [API key management guidance](https://docs.z.ai/devpack/notice/api-key-management). Standard, Coding Plan and Compatible keys have different usage rules; a Standard key does not use Coding Plan quota. The user reports that the supplied key is a Personal Coding Plan/Team Plan key. The first supplied key returned insufficient_quota; a direct built-in call also returned usageLimitExceeded with zero reported tokens. A replacement key subsequently completed built-in GLM-5.3 inference and an official Codex GLM-5.3-Flash probe. Clean Windows R5 GLM-5.3-Flash/low passed 123 real CLI invocations in 172.47 seconds, covering chat/tool results, queue/resume, planner, worker, exactly one owned loopback GET, finding, goal completion and report. An earlier GLM-5.3 workflow reached worker evidence and finding registration before a final-planner timeout. Its later R5/low run failed on 2026-10-17 at 07:29:59 UTC, before any tool executed, with `codex: tool arguments must be a JSON object string` (17,933 input / 25 output tokens). The raw argument value was not captured; its shape and underlying cause are unknown.

The rebuilt Windows R7 GLM-5.3/low workflow subsequently passed **119 installed CLI invocations in 138.50 seconds**, including real chat Read, queue, planner/worker, goal completion, one owned GET, finding/report and persisted usage; cleanup succeeded. **No format-correction attempt occurred**: this proves the current workflow on that run, not live recovery through the new branch. [R7 receipt](evidence/standalone-cli/candidate-r7-verification.json). R7 Flash acceptance remains **unresolved**. The initial low run failed after 136.98 seconds at the exactly-one-finding assertion without capturing its actual count. A low diagnostic failed after 286.06 seconds with two distinct `report_finding` calls (intent IDs 3/6, finding IDs 1/2) for the same owned marker/name. High reasoning also failed after 482.38 seconds: one worker fetch and one finding were recorded, but the final planner did not finish within the seven-minute task budget. All three runs had zero correction attempts. The underlying model/provider cause is unknown; the historical R5 Flash pass does not certify R7. **The public release remains held under the user's all-pass condition**, as well as final-head CI and human review. These results do not establish the cause of the earlier account response. See the dated verification record for full workflow acceptance.

## Behavior and compatibility limits

- Each completion launches an owned app-server process and a fresh ephemeral thread. The complete current Norma conversation, roles, tool IDs/results and available tool schemas are translated into JSON again. ScopeWeaver history, branches and restart persistence remain in the engine; this is not native Codex thread replay.
- Streaming is buffered until the completed structured response is validated. Text and tool events then enter the usual agent stream. There are no live token deltas before completion, and incomplete or malformed tool requests never reach Norma.
- Only Z.AI argument-object validation failures allow one additional fresh invocation with fixed format-correction feedback. The entire first response is rejected before tool execution; strict decoding is unchanged. Both attempts share one 10-minute deadline, cancellation and `local_llm_only` are rechecked, and input/output/cache-read/cache-write usage is summed once. OAuth, network, quota, protocol, unknown-tool and duplicate-key failures do not trigger this correction. [Source tests/vet](evidence/standalone-cli/glm-format-repair.json) and [independent review](evidence/standalone-cli/glm-format-repair-review.md) passed; live correction recovery is not yet established.
- Codex shell, web, image, MCP, plugins, apps, skills, hooks and subagents are disabled for these completions. MCP isolation is checked before inference; unsupported MCP names fail explicitly. Unexpected tool items or app-server requests reject the completion. ScopeWeaver's own configured tools/MCP integrations still execute through Norma with their existing controls.
- Project instructions and environment access are disabled. ChatGPT OAuth retains the official Codex home for login and inherits its global instructions, including global `AGENTS.md`. Z.AI uses a disposable clean home. Account-global instructions can therefore affect the OAuth route.
- Text, thinking and structured tool history are supported. Images and other non-text inputs are rejected. The adapter bounds translated context to 32 MiB, structured output to 2 MiB and tool calls to 128 per completion. Each invocation has a 10-minute timeout, subject to earlier caller cancellation.
- `max_tokens`, temperature, stop sequences and disabled-thinking requests are advisory prompt hints, not API-enforced controls. OAuth compaction preserves the configured model effort instead of forcing unsupported `none`; GLM maps disabled thinking to `low`. Profile context-window values guide engine compaction and do not change the provider's actual capacity. OAuth reasoning accepts empty, `none`, `minimal`, `low`, `medium`, `high` or `xhigh`; the selected model may impose additional restrictions.
- Profile proxy settings are rejected. Custom session headers, HTTP-provider retry options and raw HTTP request/response capture are unsupported by this transport. Normal engine retry/failover and rate/concurrency controls still apply where configured. Usage comes from app-server token-usage notifications; raw Codex diagnostics are withheld.
- Both routes are cloud providers. `local_llm_only` is checked on every invocation and blocks them before Codex starts; an allowed CIDR cannot convert this route into local inference. Saving a profile or seeing it as configuration-ready does not prove executable, login, entitlement or network readiness.
- Model discovery through `llm models list` is unsupported for `codex-cli`; configure the exact model explicitly. Package preparation and core offline reuse do not make cloud inference offline.

## Verification scope

Windows R7 OAuth passed **131 installed CLI invocations in 163.68 seconds**, with real chat Read, queue/history, planner/worker, goal completion, one owned GET, finding/report, persisted usage and successful cleanup. [R7 receipt](evidence/standalone-cli/candidate-r7-verification.json). Earlier clean Windows R5 OAuth passed 137 installed CLI invocations in 205.12 seconds. A separate lifecycle campaign checked active chat/task and side-question drain/restart, steering, edits, retry and branch behavior with 130 calls plus 3 preflights; a cancelled side question does not prove completed side inference. Windows R5 Flash passed the full built-in workflow above. These are candidate-specific results, not every-model, every-platform or final-package acceptance. [Dated receipts and remaining gates](CLI-VERIFICATION.md).

R4 OAuth compaction passed 33 CLI calls with 2 new model completions. Per-call paid records and summaries survived restart, but existing conversation/dashboard aggregate usage did not increase. Compaction costs are not established as included in those aggregates. The source at `80f9016` is newer than R5 and passed all three hosted Windows, Ubuntu and macOS jobs in run39012848998. Later source and final release status follow the latest PR checks and human review; this dated pass does not certify every option/state or later artifact. The repository is private and draft PR #1 targets `dev`; no public CLI release exists.

Implementation references in the source checkout: `codexprovider/provider.go`, `codexprovider/protocol.go`, `codexprovider/isolation.go`, `agent/provider.go` and `internal/cli/operations.json`. The official [app-server documentation](https://learn.chatgpt.com/docs/app-server) describes the versioned protocol; this adapter's verified baseline remains 0.162.1.
