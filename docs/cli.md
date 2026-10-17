# Standalone architecture

[한국어](ko/cli.md) · [AI providers](ai-providers.md)

The executable contains both the public CLI and a private engine runtime. Commands reuse that runtime through authenticated loopback HTTP. No separately installed ScopeWeaver server is needed. OS locks prevent duplicate owners; long-running jobs survive client exit. `runtime stop --yes` stops owned work and its database.

The public download-mode folder includes the CLI/engine, built-in skills, notices, source pins and a file manifest. The first operational command downloads and verifies pinned PostgreSQL archives and, on Windows, app-local Visual C++ runtime files before starting the engine. They are kept under `runtime/` beside the executable and shared by its separate homes. This requires a writable portable folder and network access on first use, with explicit failures for missing permissions, unavailable downloads or checksum mismatches. It registers no operating-system service. After preparation, core operations reuse the runtime offline; cloud models and external integrations retain their own requirements. Discovery and dry-run commands do not download it.

Default packaging uses this first-use path. `go run ./packaging -target windows-amd64 -bundle-runtime` prepares a local verification folder with those runtimes already present. The builder's `-offline` flag accepts only its verified download cache; it is not a runtime network policy. Bundled-runtime verification folders are not the public artifact. Pinned checksums and local tests do not establish publication rights for third-party binaries.

A home defaults to `state/` beside the executable. `private/` holds protected credentials/configuration/logs, `postgres/` holds the cluster, `workspace/` holds browsable files/evidence, and `skills/` is seeded once from the bundle and then preserves edits/deletions. `--home` isolates mutable state. Explicit private configuration prevents accidental use of inherited database settings.

Stop before relocation. Paths and ports are recomputed on restart. Relocation within the same platform is supported. A running cluster must not be copied, and raw PostgreSQL directories are not a cross-platform backup format. Major-version mismatch or missing credentials fails without deleting data. Crash recovery authenticates and verifies the existing database's data directory before adopting it.

Notification delivery blocks loopback and link-local destinations by default. For a trusted local relay, the existing explicit `ARTEX_NOTIFY_ALLOW_LOCAL=1` setting is passed to the private runtime; it is never enabled automatically. This opt-in allows both loopback and link-local delivery. Stop the home and set the variable before starting it again when changing this policy. Other inherited database/configuration paths remain isolated.

On Windows, PostgreSQL needs ASCII command-line paths. Unicode portable folders and homes use verified ASCII 8.3 aliases when the filesystem provides them. If no usable alias exists, startup fails explicitly without moving data or changing system settings; use an ASCII folder and home. ASCII paths do not require 8.3 support.

`internal/cli/operations.json` is the executable command contract: names, paths, input types, effects, response modes and source references. Tests compare it with static routes, generated side-question routes and wildcard chat dispatch. `schema` reads the same catalog.

JSON, uploads, downloads, Markdown and events use separate transport paths. JSON is bounded at 32 MiB. Downloads refuse overwrite and remove incomplete files. Mutations are never automatically retried. Streams emit one JSON event per line and stop on cancellation/backend closure; backend cursor parameters support resumption.

Effects are classified by behavior: integration queries can contact services; tool/notification tests execute real actions. Non-read operations require `--yes`. `--dry-run` never starts/stops runtime or submits work. The REPL shares execution and validation with direct commands. It does not invoke a shell; `--data -` consumes the next complete JSON object.

Errors have a stable code, localized summary and preserved technical detail on stderr; results use stdout. Command identifiers and schema field names remain stable. Returned evidence is untrusted data.

Built-in chat, planner and workers share the LLM provider factory. The `codex-cli` format uses official Codex app-server over stdio for ChatGPT OAuth or the pinned Z.AI GLM-5.3 Responses route. It translates the current Norma conversation into each fresh ephemeral thread, validates the entire structured completion, then lets Norma execute the requested tools. Streaming is buffered; Codex's direct tools, MCP and plugins are disabled. OAuth uses the official Codex login/home and inherits global instructions; Z.AI uses a clean temporary home. Both routes fail closed under `local_llm_only`. Official Codex is an additional dependency; only 0.162.1 has been verified. See [provider setup and exact limits](ai-providers.md).

Updates use only `cskwork/scopeweaver-cli` releases, verify checksums and preserve rollback. The canonical launch path is captured before replacement so Linux cannot relaunch the renamed old executable. Database major versions and user-edited skills are preserved.

This independent repository derives from ScopeWeaver `2df0b307dcb817a459abebd3a4c240a2019598db`. The Go module is `github.com/cskwork/scopeweaver-cli`. Original sources/tests are retained for traceability; packaged execution never reads a sibling checkout. Pinned runtime hashes and notices accompany packages; default packages fetch runtime bytes on first use. See [verification](CLI-VERIFICATION.md) for tested platforms and limitations.
