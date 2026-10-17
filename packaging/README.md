# Portable packaging

[한국어](README.ko.md)

Run from this repository root with Go 1.26.3 or newer:

```powershell
go run ./packaging -target windows-amd64
```

This creates a new `dist/scopeweaver-cli-windows-amd64` directory containing the
CLI binary, its private PostgreSQL 17.11 runtime, skills, licenses and a SHA-256
manifest. Copy the entire folder to run it. Users do not need Go, Node, Docker,
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

# Rebuild without network after the pinned downloads have been cached.
go run ./packaging -target windows-amd64 -offline -out dist/offline-build

# Prepare the runtime for development before the CLI binary is available.
go run ./packaging -target windows-amd64 -runtime-only -out dist/runtime-only

go test ./packaging -count=1
```

`-runtime-only` is a development bundle and does not include the CLI executable.
Downloads stay in `dist/cache`. A cached file with a wrong hash fails rather than
being replaced. A failed build leaves its output directory for inspection; choose
a fresh directory for the next attempt.

`postgres-sources.json` pins every publisher artifact and SHA-256. The publisher
is [Zonky Embedded Postgres Binaries](https://github.com/zonkyio/embedded-postgres-binaries),
which packages EDB binaries on Windows/macOS and builds PostgreSQL for Linux. Its
bundles are intended for embedded testing; this CLI uses a private local database.
The original [PostgreSQL Windows page](https://www.postgresql.org/download/windows/)
and [EDB binary page](https://www.enterprisedb.com/download-postgresql-binaries)
document the binary distribution. This reduced bundle includes `initdb`,
`pg_ctl` and `postgres`; it does not include `psql`.

On Windows, the bundle also carries three app-local Microsoft Visual C++ DLLs
from the pinned Microsoft VCLibs package. Its checksum is independently recorded
in [Microsoft's winget manifest](https://github.com/microsoft/winget-pkgs/blob/master/manifests/m/Microsoft/VCLibs/Desktop/14/14.0.33321.0/Microsoft.VCLibs.Desktop.14.installer.yaml).
No runtime installer runs. Source URLs, hashes and the original Microsoft license
are included in the output. [Microsoft documents app-local deployment](https://learn.microsoft.com/en-us/cpp/windows/redistributing-visual-cpp-files)
and limits redistribution to licensed Visual Studio users. **External publication
requires the applicable Microsoft redistribution rights and compliance with the
bundled LGPL libraries' source obligations.** Local packaging is not clearance
to publish these third-party binaries.

Archive extraction uses Go's rooted filesystem API, rejects absolute paths,
escaping links, Windows device names, duplicate files and special device entries.
Internal library links are materialized as regular copies, so cross-packaging on
Windows needs no symlink privilege. The `github.com/ulikunitz/xz` dependency is used
only to build this packaging tool, not by the shipped CLI.

`MANIFEST.json` records the source revision, dirty status, target, PostgreSQL
source URL/hash and SHA-256 for every regular output file except the manifest
itself. `POSTGRES-SOURCES.json` includes the pinned PostgreSQL and Windows runtime
sources. `licenses/SOURCES.json` records the source/hash of preserved notices.
The Go dependency license directory contains licenses from modules linked into
the CLI and the Go standard-library license. Keep all notices with the package.
