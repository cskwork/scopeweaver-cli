# Contributing to scopeweaver-cli

[한국어](CONTRIBUTING.ko.md)

Read README.md, docs/cli.md and docs/CLI-VERIFICATION.md. Preserve AGPL-3.0, attribution and local-isolated usage restrictions. Existing engine sources/tests retain behavior and evidence.

Create feature/fix branches from dev and target dev. Promotion to main, remote creation and publication require maintainer authorization and human review. Preserve unrelated work.

The command contract is internal/cli/operations.json. New operations need named commands, parameter/effect/response metadata and source-drift tests. Classify effects by actual handler behavior. Never weaken checks to obtain a pass.

Keep English/Korean README and docs paired. Interface summaries support en/ko/zh-Hans; raw evidence and technical diagnostics remain intact.

Run:

    go test ./internal/cli ./internal/portable ./packaging ./selfupdate -count=1
    go vet ./internal/cli ./internal/portable ./packaging ./selfupdate
    go run ./packaging -target windows-amd64

Set SCOPEWEAVER_CLI_E2E_BINARY to the resulting executable and run:

    go test ./internal/cli -run '^TestPortableCLIEndToEnd$' -count=1 -v

Choose the proper target on Linux/macOS. E2E must invoke the packaged executable from another working directory with disposable state. Core database tests require fresh scopeweaver_test_* databases via ARTEX_TEST_DSN. Never use user databases or substitute mocked success for required runtime verification. Report inherited failures/skips.

Runtime changes need auth, startup concurrency, crash recovery, graceful shutdown, complete stopped-folder relocation and per-home skills checks. Preserve clusters on version mismatch or missing credentials.

Packages must include pinned runtime checksums, dependency notices and file manifests. Review third-party redistribution requirements before publication. No workflow publishes automatically. Original workflows are retained under docs/upstream-workflows as historical reference only.
