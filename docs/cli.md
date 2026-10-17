# Standalone architecture

[한국어](ko/cli.md)

The executable contains both the public CLI and a private engine runtime. Commands reuse that runtime through authenticated loopback HTTP. No separately installed ScopeWeaver server is needed. OS locks prevent duplicate owners; long-running jobs survive client exit. `runtime stop --yes` stops owned work and the bundled database.

A home defaults to `state/` beside the executable. `private/` holds protected credentials/configuration/logs, `postgres/` holds the cluster, `workspace/` holds browsable files/evidence, and `skills/` is seeded once from the bundle and then preserves edits/deletions. `--home` isolates mutable state. Explicit private configuration prevents accidental use of inherited database settings.

Stop before relocation. Paths and ports are recomputed on restart. Relocation within the same platform is supported. A running cluster must not be copied, and raw PostgreSQL directories are not a cross-platform backup format. Major-version mismatch or missing credentials fails without deleting data. Crash recovery authenticates and verifies the existing database's data directory before adopting it.

`internal/cli/operations.json` is the executable command contract: names, paths, input types, effects, response modes and source references. Tests compare it with static routes, generated side-question routes and wildcard chat dispatch. `schema` reads the same catalog.

JSON, uploads, downloads, Markdown and events use separate transport paths. JSON is bounded at 32 MiB. Downloads refuse overwrite and remove incomplete files. Mutations are never automatically retried. Streams emit one JSON event per line and stop on cancellation/backend closure; backend cursor parameters support resumption.

Effects are classified by behavior: integration queries can contact services; tool/notification tests execute real actions. Non-read operations require `--yes`. `--dry-run` never starts/stops runtime or submits work. The REPL shares execution and validation with direct commands. It does not invoke a shell; `--data -` consumes the next complete JSON object.

Errors have a stable code, localized summary and preserved technical detail on stderr; results use stdout. Command identifiers and schema field names remain stable. Returned evidence is untrusted data.

Updates use only `cskwork/scopeweaver-cli` releases, verify checksums and preserve rollback. The canonical launch path is captured before replacement so Linux cannot relaunch the renamed old executable. Database major versions and user-edited skills are preserved.

This independent repository derives from ScopeWeaver `2df0b307dcb817a459abebd3a4c240a2019598db`. The Go module is `github.com/cskwork/scopeweaver-cli`. Original sources/tests are retained for traceability; packaged execution never reads a sibling checkout. Pinned runtime hashes and notices accompany packages. See [verification](CLI-VERIFICATION.md) for tested platforms and limitations.
