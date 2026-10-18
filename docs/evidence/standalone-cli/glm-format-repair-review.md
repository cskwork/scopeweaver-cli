# GLM format repair: independent skeptic review

Verdict: no source-level blocker in the reviewed patch. Conditional on the coordinator's focused checks and real packaged flagship GLM-5.3/low workflow; this review is not a live-success claim.

Scope: read-only provider/decoder/protocol/test review. No credentials, live API calls, or tests were run by this reviewer. Only this ignored receipt was written. CONTRIBUTING.md and the machine rules were read. No code, documentation, or publication changes were made.

## Evidence and diagnosis boundary

The latest supplied R5 receipt, dist/candidate-r5-glm-53-final-e2e.jsonl, records the built-in flagship glm-5.3/low chat failing at `codex: tool arguments must be a JSON object string`, after 17,933 input / 25 output tokens and before any tool was called. The latest raw argument value was not captured. The older regression fixture documents a raw `[]` for add_intent. These are not evidence that the latest raw value was also `[]`, or that the underlying root cause is definitively the model rather than provider/adapter behavior.

Normal path: a fresh official Codex thread returns a text action envelope; decodeEnvelope validates the whole completion; Stream then publishes validated Norma tool descriptions; Norma executes them. The failure occurred during envelope decoding, before that execution boundary. The patch supplies a bounded adapter recovery opportunity for that observed rejection class. It cannot independently establish provider conformance or flagship end-to-end acceptance.

## Reviewed invariants

- provider.go: Complete permits exactly one additional fresh invocation, only for the Z.AI route and the dedicated argument-object error. Native-tool, app-server/protocol, quota/rate, network/process, unknown-tool and ordinary malformed-envelope failures do not enter this branch.
- Complete creates one parent ten-minute deadline covering both invocations. It rechecks context and local-LLM-only policy before the second invocation; the existing per-run deadline cannot extend the parent deadline.
- The original prompt and current tool schemas remain present. Correction feedback is fixed text, includes the complete input-object requirement, and never copies the failed output or provider diagnostic. The fixed started diagnostic exposes neither prompts nor credentials.
- decodeEnvelope retains strict outer validation, UTF-8, output bytes, maximum call count, duplicate-key/depth checks and strict object acceptance. Bad arguments do not become coerced objects. Duplicate/depth ambiguity remains non-repairable.
- Initial concern was mixed envelopes: a bad argument in call 1 could hide an unavailable tool in call 2 and provoke a paid repair contrary to policy. The patch resolves this by deferring the repairable error until all calls have been examined; mixed unknown and duplicate fixtures cover this case.
- The rejected first message is never returned or streamed, including its valid earlier tool calls or assistant text. This prevents duplicate execution of an initially valid subset. The second invocation uses the existing isolated run path; no native Codex tools or dynamicTools rewrite is introduced.
- Usage.Add sums input/output/cache-read/cache-write snapshots before returning any second-attempt error. Stream emits the aggregate once on message_start. This fits llmrec/llmrec.go:205-220, whose per-provider-call accounting uses field maxima and forwards deltas. Emitting separate attempt events would risk losing totals; this patch does not do so.

## Test coverage inspected, not executed here

argument_repair_test.go covers repaired success, repeated invalid output, malformed second output, no first-response text/tool leakage, exactly two subprocess attempts, all four usage fields once, OAuth non-repair, ordinary malformed/unknown/duplicate output non-repair, mixed unknown/duplicate envelopes, and policy/cancellation changes between invocations. Existing provider tests cover protocol denial, usage snapshots, early consumers and subprocess cancellation.

The frozen final tests also cover first and second transport/native/quota failures, paid second-attempt deadline failure, feedback context bounds, one fixed non-sensitive diagnostic, and an early stream consumer retaining both attempts usage. The implementation agent reports go test ./codexprovider -count=1 -v passed in 4.793 seconds, go vet ./codexprovider passed and git diff --check passed; those checks were not rerun by this reviewer. Final file hashes were independently re-read after the freeze notification. No source-review blocker remains; real packaged flagship E2E remains the acceptance gate.

## Cheapest useful live discriminator

Use the packaged built-in flagship glm-5.3/low route, ideally existing TestPortableCLIGLMEndToEnd. Its first phase reproduces the owned Read/random-marker contract and asserts exactly one Read plus paired result and final answer. Later phases cover planner/worker, synthetic finding/report and exactly one owned loopback GET. Flash success and direct synthetic API/schema probes do not certify this flagship route.

A live correction-start diagnostic proves an attempt, not successful recovery. Full E2E success with zero correction diagnostics proves the workflow on that run but not that the new repair branch recovered a live failure. Keep those evidence claims separate. If the flagship still fails, retain the failure and hold acceptance instead of substituting Flash.

## Reviewed file SHA-256

- codexprovider/provider.go: 88031274C62CA44854C6D9FA27255E965405B871ACEEED8EAD04917A55C53CAF
- codexprovider/provider_test.go: B17E0A10CC9C86CD7B2F42E10BE23013E68A8289A720F05403E80036AD5F5D93
- codexprovider/argument_repair_test.go: 78C74D1CBA6E6BA6C2500ED1BC5555225AB38EBCA50529B3606F55B448966FBD

These hashes bound this source review; edits after them require a narrow re-review.

