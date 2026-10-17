# Narrow shutdown review, 2026-10-17

R3 source-only re-review of the current uncommitted shutdown changes against HEAD. No production files changed, tests rerun, inference performed, or runtime shutdown exercised during this review.

## Result

Both previously reported P2 findings are closed. No remaining concrete issue was found within these fixes. Native verification of the rebuilt candidate remains a separate acceptance check.

## Closed: side-question finalization outside the drain

server/chat_shutdown.go now closes side admission under side.commands before snapshotting and cancelling registered runs. Side-question admission uses that same mutex and checks both side.closing and server-context cancellation, so a new run cannot register after the snapshot. Each captured run.done closes after terminal persistence. The drain waits for tracked chat finalization, captured side runs and publisher completion, then performs the final side-snapshot and chat-context flush before signaling completion. This covers checkpoints emitted during chat unwind after the publisher's cancellation flush.

The existing chatMu barrier still prevents WaitGroup additions after closing begins. The runtime keeps the database open throughout the bounded drain. A timeout remains an explicit failure rather than evidence of completed finalization.

## Closed: failed terminal-result publication could appear successful

internal/portable/runtime.go now advertises private shutdown_result_version=1 in the authenticated runtime record and captures that value in Connection. After the matching runtime releases ownership, Stop requires a matching nonce-bound terminal receipt for version 1. Missing or mismatched receipts return an explicit error, including when publication failed. Legacy version 0 retains compatibility with an absent receipt.

Deferred ordering remains correct: chat drain precedes manager/database teardown; the terminal result includes teardown errors and is written before ownership release. The version is checked through the existing authenticated health-record equality, so an unrelated stale result cannot satisfy the new runtime's receipt requirement.

## Scope and evidence

Reviewed server/chat_shutdown.go, related chat admission/finalization and side-question/publisher paths, and internal/portable/runtime.go's connection, Stop, terminal-result and teardown ordering. Focused test results were reported by the implementation agent, not independently rerun here. This receipt is a source review, not a claim that the final executable passed native shutdown E2E.
