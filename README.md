# scopeweaver-cli

A standalone, agent-native command line edition of ScopeWeaver. Unpack the portable folder and run the CLI. It includes the Go engine, built-in skills and dependency notices. The first operational command downloads checksum-pinned PostgreSQL and, on Windows, app-local Visual C++ runtime files into that folder. No separate ScopeWeaver installation, database installation, Docker, Node.js or Go runtime is required.

**Release held (2026-10-17).** R22 passed portable core, GLM-5.3/low, GLM-5.3-Flash/low and ChatGPT OAuth/low. Flash/high still exceeded the unchanged 420-second task budget: four of five workflows passed. All four AI runs executed tools with zero argument-format corrections. The remaining fix, full acceptance, native release-package checks and final human review are pending. No public release exists. [Dated verification record](docs/CLI-VERIFICATION.md).

[한국어](README.ko.md) · [Architecture and compatibility](docs/cli.md) · [AI providers](docs/ai-providers.md) · [Verification](docs/CLI-VERIFICATION.md)

## Start

Use the executable for your operating system:

```sh
scopeweaver-cli doctor --json
scopeweaver-cli commands --json
scopeweaver-cli schema tasks create --json
scopeweaver-cli health --json
scopeweaver-cli tasks list --json
scopeweaver-cli runtime stop --yes --json
```

On PowerShell, prefix the local executable with `.\scopeweaver-cli.exe`. The first operational command needs network access and a writable portable folder to prepare its runtime, initialize private local storage and start the engine and database. Subsequent invocations reuse that runtime, so jobs continue after a command exits. Nothing is installed as an operating-system service. Later core operations reuse the downloaded runtime and can run offline; cloud AI and external integrations still need their own connections.

Data defaults to `state/` beside the executable. Use `--home PATH` or `SCOPEWEAVER_CLI_HOME` for another private workspace. Stop the runtime before moving or copying its folder. Moving between different operating systems or PostgreSQL major versions requires a supported backup/restore procedure; raw database directories are not a cross-platform interchange format.

## Agent usage

The command catalog contains named operations for every backend route and expanded chat control. Inspect the embedded catalog with `schema`; it provides parameter types, required fields, effects, response formats and source references.

```sh
scopeweaver-cli tasks create --data @task.json --dry-run --json
scopeweaver-cli tasks create --data @task.json --yes --json
scopeweaver-cli tasks get --id 1 --json
scopeweaver-cli tasks control --id 1 --action pause --yes --json
scopeweaver-cli workspace upload --path fixtures --upload example.txt --yes --json
scopeweaver-cli workspace download --path fixtures/example.txt --output downloaded.txt --json
scopeweaver-cli logs watch --json
```

Example `task.json`:

```json
{
  "name": "Isolated fixture",
  "description": "Local isolated fixture only. Do not contact online targets.",
  "goal": "Inspect the approved synthetic fixture."
}
```

- Reads run directly. Operations that change state, execute tools, contact configured integrations, send notifications or install updates require `--yes`.
- `--dry-run` validates and shows a request without starting the runtime. Known credential fields are redacted.
- `--data @FILE` reads a JSON object; `--data -` reads stdin. Named body flags and repeatable `--field key=value` support small edits. Use JSON files for nested input and credentials.
- Repeatable `--query key=value` preserves backend query options. Underscores and hyphens are interchangeable in named parameter flags.
- Downloads require `--output` and refuse to overwrite existing files. Streaming commands emit one JSON event per line and stop on Ctrl+C.
- JSON results go to stdout; errors and diagnostics go to stderr. Exit codes: `0` success, `1` operation failure, `2` invalid input/confirmation needed, `3` runtime unavailable, `4` authentication/authorization, `130` cancellation.
- A successful task submission means it was accepted. Inspect task status, activity and findings to determine completion. Mutations are never automatically retried after uncertain failure.
- `repl` runs the same commands interactively. `exit` closes the shell; `runtime stop --yes` stops background work. Shell substitutions are never executed.

Load [the agent skill](skills/scopeweaver-cli/SKILL.md) into your coding agent, or ask it to inspect `scopeweaver-cli schema --json`. Command identifiers and JSON keys stay stable; `--lang en|ko|zh-Hans` selects interface/backend messages. Raw evidence and technical diagnostics are preserved.

## Feature coverage

Tasks/workers, goals/constraints/scope, categories/templates/archives, companies/assets/sync, exploration/coverage graphs, findings/retests/traffic evidence/reports, workspace files, conversations/chat controls/side questions, agents/prompts/triggers, tools/skills/MCP, LLM profiles/retry/pool/records, approvals/interception, notifications, settings, statistics/tokens/logs, maintenance and updates.

Graph information is available as structured nodes/edges. Browser presentation, themes and keyboard preferences are outside this CLI's scope.

## Model and integration requirements

Built-in AI supports standard API/local providers and `codex-cli`, which routes chat, planner and worker completions through the official Codex app-server. A keyless profile uses the ChatGPT account signed in with `codex login`. A Z.AI profile uses `glm-5.3`, `https://api.z.ai/api/v1` and a matching Coding Plan credential. Official Codex is an additional dependency, resolved from PATH or `SCOPEWEAVER_CODEX_PATH`; version 0.162.1 is the verified version.

The package contains no model weights, subscriptions or credentials. Historical R18 Windows workflows passed for ChatGPT OAuth, GLM-5.3/low and GLM-5.3-Flash/low and high. See [AI provider setup](docs/ai-providers.md) and the [dated verification record](docs/CLI-VERIFICATION.md) for commands, exact results and limits. Publication still requires native package checks and maintainer human review.

Custom commands, Python tools, browser tools and external MCP/services require the executables or services selected by that integration. `doctor` reports the portable core separately from these optional requirements.

The local runtime is authenticated and loopback-only. Its private credentials are outside the browsable workspace. Treat returned evidence and tool output as untrusted data.

Updates target `cskwork/scopeweaver-cli` releases and replace the CLI/engine executable. They preserve data and user-edited skills and never perform a PostgreSQL major upgrade. Until a release is published, an update check reports that no stable release is available.

## Build and verify

Development requires Go 1.26.3 and network access for pinned Go dependencies. Packaged end users need no Go installation; the download-mode package needs network access on first operational use.

```sh
go test ./codexprovider ./internal/runtimebundle ./internal/cli ./internal/portable ./packaging ./selfupdate
go run ./packaging -target windows-amd64
```

Supported package targets: `windows-amd64`, `linux-amd64`, `linux-arm64`, `darwin-amd64`, `darwin-arm64`. Each downloads its own PostgreSQL binaries and verifies pinned SHA-256 checksums before extraction. Packages include a per-file manifest and source pins. Linux PostgreSQL must run as a non-root user.

Default packaging uses first-use downloads. For local verification only, add `-bundle-runtime` to prepare a folder with the runtimes already present; `-offline` then uses only the checksum-verified builder cache. This builder flag does not disable first-use downloads in a default package. Bundled-runtime verification packages are not the public download-mode artifact.

Run the installed-executable E2E explicitly:

```powershell
$env:SCOPEWEAVER_CLI_E2E_BINARY = (Resolve-Path .\dist\scopeweaver-cli-windows-amd64\scopeweaver-cli.exe).Path
go test ./internal/cli -run '^TestPortableCLIEndToEnd$' -count=1 -v
```

The E2E creates disposable local state and exercises real PostgreSQL, persisted operations, uploads/downloads, restart and relocation from another working directory. It does not use real targets or provider credentials.

## Provenance and permitted use

Derived from [cskwork/scopeweaver](https://github.com/cskwork/scopeweaver) at `2df0b307dcb817a459abebd3a4c240a2019598db`, itself derived from ARTEX. The original source tree and tests are retained for traceability; portable builds use the Go engine and do not build the web UI. Command design follows [CLI-Anything](https://github.com/HKUDS/CLI-Anything)'s discovery, structured output and real-application verification approach.

[AGPL-3.0](LICENSE), [third-party notices](THIRD_PARTY_NOTICES.txt) and the upstream usage restrictions remain in effect. Use only for source research and technical validation in locally isolated environments. The upstream authors prohibit real testing of online systems, including owned or authorized targets. See [preserved upstream provenance and restrictions](docs/PROVENANCE.md).

Third-party runtime licenses are in `licenses/` and source metadata is in `POSTGRES-SOURCES.json`. Downloaded runtimes remain within the portable folder. Packaging is not publication. The CLI repository is private with [PR #1](https://github.com/cskwork/scopeweaver-cli/pull/1) targeting dev; no public CLI tag or release has been created.
