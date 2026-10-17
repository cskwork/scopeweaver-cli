# Portable packaging

[한국어](README.ko.md)

Run from this repository root with Go 1.26.3 or newer:

```powershell
go run ./packaging -target windows-amd64
```

This creates a new `dist/scopeweaver-cli-windows-amd64` directory containing the
CLI binary, skills, licenses, pinned runtime sources and a SHA-256 manifest.
PostgreSQL and Visual C++ binaries are absent from the default package. The first
operational command downloads the pinned PostgreSQL 17.11 runtime, verifies its
SHA-256 before extraction, and atomically stores it under `runtime/postgres/`
beside the executable. On Windows it also downloads the pinned app-local Visual
C++ DLLs. Copy the entire folder to run it. First use requires HTTPS access and
write permission beside the executable; later starts reuse the installed runtime. Users do not need Go, Node, Docker,
a system PostgreSQL installation, or a PostgreSQL service. The packager does not
install anything or modify a database. The CLI's optional external tools and LLM
providers still need their own configuration.

Supported package targets are `windows-amd64`, `linux-amd64`, `linux-arm64`,
`darwin-amd64` and `darwin-arm64`. Linux targets use glibc, not Alpine/musl.
Windows requires Windows 10 or newer. macOS/Linux still use their normal system
libraries; an extracted or cross-compiled package is not proof of testing on that
operating system.

```powershell
# Use a new output directory; existing output is never overwritten.
go run ./packaging -target windows-amd64 -out dist/my-portable-folder

# Default packaging does not download PostgreSQL or Visual C++ binaries.
go run ./packaging -target windows-amd64 -offline -out dist/offline-build

# Local verification only: include pinned runtime binaries.
go run ./packaging -target windows-amd64 -bundle-runtime -out dist/local-bundled-test

# Prepare the runtime for development before the CLI binary is available.
go run ./packaging -target windows-amd64 -runtime-only -out dist/runtime-only

go test ./packaging -count=1
```

`-bundle-runtime` is for local verification; `-runtime-only` implies it and omits
the CLI executable. Only these modes download runtime artifacts into `dist/cache`.
The default build does not access that cache. A cached file with a wrong hash fails rather than
being replaced. A failed build leaves its output directory for inspection; choose
a fresh directory for the next attempt.

`../internal/runtimebundle/postgres-sources.json` embeds every publisher artifact
and SHA-256 in the CLI; `packaging/postgres-sources.json` is checked for equality. The publisher
is [Zonky Embedded Postgres Binaries](https://github.com/zonkyio/embedded-postgres-binaries),
which packages EDB binaries on Windows/macOS and builds PostgreSQL for Linux. Its
bundles are intended for embedded testing; this CLI uses a private local database.
The original [PostgreSQL Windows page](https://www.postgresql.org/download/windows/)
and [EDB binary page](https://www.enterprisedb.com/download-postgresql-binaries)
document the binary distribution. This reduced bundle includes `initdb`,
`pg_ctl` and `postgres`; it does not include `psql`.

On Windows, first operational use fetches three app-local Microsoft Visual C++ DLLs
from the pinned Microsoft VCLibs package. Its checksum is independently recorded
in [Microsoft's winget manifest](https://github.com/microsoft/winget-pkgs/blob/master/manifests/m/Microsoft/VCLibs/Desktop/14/14.0.33321.0/Microsoft.VCLibs.Desktop.14.installer.yaml).
No runtime installer runs. Source URLs, hashes and the original Microsoft license
are included in the default output; runtime binaries are fetched directly from the publisher. [Microsoft documents app-local deployment](https://learn.microsoft.com/en-us/cpp/windows/redistributing-visual-cpp-files)
and limits redistribution to licensed Visual Studio users. **Publishing bundled local-test runtime binaries requires the applicable
Microsoft redistribution rights and compliance with the bundled LGPL libraries'
source obligations.** Default download mode does not itself establish legal
clearance; local packaging is not permission to publish third-party binaries.

Archive extraction uses Go's rooted filesystem API, rejects absolute paths,
escaping links, Windows device names, duplicate files and special device entries.
The same validated extraction implementation serves packaging and first use.
Internal library links are materialized as regular copies, so cross-packaging on
Windows needs no symlink privilege. The shipped CLI includes `github.com/ulikunitz/xz` to extract the pinned runtime.

`MANIFEST.json` records the source revision, dirty status, target, PostgreSQL
`runtime_mode` (`download` by default, `bundled` for local verification),
source URL/hash and SHA-256 for every regular output file except the manifest
itself. `POSTGRES-SOURCES.json` includes the pinned PostgreSQL and Windows runtime
sources. `licenses/SOURCES.json` records the source/hash of preserved notices.
The Go dependency license directory contains licenses from modules linked into
the CLI and the Go standard-library license. Keep all notices with the package.

Runtime installation is serialized across homes sharing an executable folder.
Downloads and extraction use a fresh staging directory; interrupted attempts
retry without exposing partial runtime files. No installer, service or global
configuration runs. An existing runtime folder is never replaced automatically.
Incomplete existing folders fail with restoration guidance; private data and
PostgreSQL major-version checks remain unchanged. Help, schema, status, doctor
and dry-run commands do not download the runtime.

Linux first-use installation requires a filesystem supporting atomic no-replace
rename. WSL Windows-mounted folders such as `/mnt/c` and `/mnt/d` can reject
this operation with `EINVAL`; installation then fails without exposing a partial
runtime. Extract the portable package into the WSL Linux filesystem before first
use. Native ext4 in the existing WSL environment passed the bounded verification;
that does not establish support for every Linux filesystem or a clean Linux VM.
