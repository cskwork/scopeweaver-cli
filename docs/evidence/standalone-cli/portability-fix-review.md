# Independent portability diff review

Result: qualified PASS. No blocking correctness, security, or compatibility defect found in the reviewed changes versus ad81445.

## Reviewed scope
internal/portable/{postgres.go,pgpath_command.go,pgpath_unix.go,pgpath_windows.go,postgres_windows_test.go,runtime.go,runtime_test.go,bootstrap_test.go}; internal/runtimebundle/download_test.go; codexprovider/provider_test.go; docs/cli.md; docs/ko/cli.md. CONTRIBUTING.md, nearby production callers, runtime parent guards, notify/http.go, and the relevant receipts were read. No source, credentials, or user state were modified. tests_rerun=false: this was a read-only independent review of code and existing evidence.

## Windows PostgreSQL
All production tool invocations are covered: postgres --version, initdb (-D and --pwfile), pg_ctl status/start/stop (-D and -l). Absolute executable/home paths are established by os.Executable and canonicalHome; changing the Windows child cwd to its binary directory therefore does not reinterpret current production path arguments. The adapter keeps individual exec arguments and leaves -o and its option string intact, retaining Go exec quoting.

Existing Unicode paths use GetShortPathName and require an ASCII result whose os.Stat identity equals the original via os.SameFile. Missing log files resolve only the existing parent and retain the intended basename. ASCII paths bypass short-name lookup. Alias resolution does not move data, change OS settings, or bypass runtimebundle.CheckParents. PostgreSQL adoption/stop still requires credential authentication before comparing SHOW data_directory to the intended directory; Windows SameFile handles short/long aliases while rejecting a different physical directory. Startup rollback, orphan authentication, private output capture, and password-file cleanup remain intact.

Existing Windows regression output in dist/post-platform-focused-20261017.jsonl records PASS for Unicode command identity/argument preservation, ASCII without short names, and unsupported Unicode diagnostics. The specifically named dist/windows-unicode-fix-20261017 folder was absent when reviewed. This review does not independently certify a rebuilt installed package or English Windows host after the fix. Unicode support remains conditional on usable ASCII aliases; the unsupported-path unit test directly checks the diagnostic helper rather than creating an actual no-8.3 volume.

## macOS fixtures
Canonicalization occurs on newly owned temporary fixtures before the deliberately malicious runtime symlinks are created. Production parent guards are unchanged. Exact symlink-path assertions prevent an unrelated /var alias from falsely satisfying the malicious-link oracle; DeadlineExceeded prevents an unrelated error from falsely satisfying lock cancellation. The fake provider verifies same physical home/cwd directory plus the existing catalog/key conditions instead of relying on path spelling.

dist/hosted-cli-first-failures.log confirms the original /var parent rejections and provider mock failures. dist/macos-ci-fix-20261017/summary.json and native-after logs record red/green native WSL symlinked-TMPDIR reproduction, Windows tests with explicit symlink-privilege skips, and Darwin amd64/arm64 cross-compilation. Native macOS runtime/test execution after the change remains unverified by this review.

## Notification boundary and documentation
isolatedEnv permits only the existing notify.AllowLocalTargetsEnv constant in addition to its previous allowlist. Values are preserved exactly, absence adds nothing, and unknown notification/database configuration stays filtered. Both runtime spawning and the runtime's own environment sanitization use this same filter. Existing notification validation/dial guards still interpret only 1 or true as opt-in; default and false stay blocked. English/Korean docs state the loopback/link-local scope and stop/set/restart requirement.

dist/r6-notification-e2e-20261017/{REPORT.md,summary.json,cleanup.json} records 29 installed calls: unset/false rejection without sink requests, opted-in owned test delivery, genuine synthetic finding/event HTTP400 failure, same-delivery retry HTTP200 with identical payload, and rejection of the existing local channel after restart with false. Exactly three owned sink requests and zero remaining owned processes are recorded. This was a dirty-tree disposable test clone with no external delivery or paid inference; link-local delivery was not contacted and the cloned manifest was not regenerated.

## Checks and remaining gates
Targeted git diff --check passed. No broad tests were rerun and no release claim is made. Coordinator's final package rebuild, installed Windows verification, and hosted macOS/Windows rerun remain the delivery gates. No additional change is requested by this review.
