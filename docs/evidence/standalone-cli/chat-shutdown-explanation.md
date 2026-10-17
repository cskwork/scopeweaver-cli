# Confirmed shutdown ordering bug and bounded fix

ScopeWeaver admits durable chat inputs, claims one run, then launches a background dispatcher. HTTP 202 acknowledges admission; it does not mean inference, transcript writes or run finalization finished. Portable shutdown normally cancels the runtime, drains HTTP handlers and closes the manager/database.

On 2026-10-17, the installed Windows candidate closed the database while its reporter dispatcher was still unwinding. The saved log shows five `sql: database is closed` errors at 14:21:52: usage, LLM recording, conversation activity, uncertain transcript persistence, and FinishChatRun. The exact affected row is conversation 2 / input c987b5fd2992b74520f7bcca8a993e62 / run 764cee16ff014cb0d4bccdaca1f4bf80. Restart recovery intentionally left the input unknown, paused its gate and preserved 48,995 input / 115 output tokens. Unknown status is not the defect.

## As-is behavior

```go
// dispatchChatInput
1 go func() {
2     defer func() { pg.FinishChatRun(context.Background(), run.ID, status) }()
3     s.executeChatRun(ctx, *run)
4 }()
// portable.Run
5 httpSrv.Shutdown(shutdownCtx)
6 return retErr // deferred mgr.Close() closes pg
```

1. The accepting HTTP handler can return while this goroutine still executes.
2. Finalization writes chat_control_runs, chat_control_inputs and the gate after inference/persistence unwinds.
3. Cancellation stops inference, but cancellation observation does not mean deferred writes have finished.
4. The HTTP server has no ownership of this goroutine.
5. Shutdown waits for HTTP handlers, which already returned 202.
6. Closing pg here races lines 2–3. Substituting run 764cee16ff014cb0d4bccdaca1f4bf80 gives the saved FinishChatRun error; recovery, rather than graceful finalization, then reconciles its running row on restart.

## To-be behavior (source excerpts)

```go
// server/chat_dispatch.go:88,126,129,142
1 if s.chatClosing || s.ctx.Err() != nil { ... return context.Canceled }
2 s.chatWG.Add(1) // under chatMu, before the goroutine can be launched
3 go func() {
4     defer s.chatWG.Done()
5     defer func() { ... pg.FinishChatRun(context.Background(), run.ID, status) ... }()
6     s.executeChatRun(ctx, *run)
7 }()
// server/chat_shutdown.go:9
8 s.chatMu.Lock(); s.chatClosing = true
9 go func() { s.chatWG.Wait(); close(s.chatDrainDone) }()
10 s.chatMu.Unlock()
11 select { case <-done: return nil; case <-ctx.Done(): return ctx.Err() }
// portable.Run:495,542 (defer ordering)
12 httpSrv.Shutdown(shutdownCtx)
13 cancel(agent.AbortShutdown); srv.DrainChats(fiveSecondContext)
14 mgr.Close()
```

1. Refuse new dispatches once cancellation or draining starts.
2. Registration and closing share chatMu, so no positive Add can race a drain started from a zero count.
3–7. LIFO defers make line 5 finish before line 4 decrements the count. All activity/transcript writes from executeChatRun also precede Done.
8–10. Closing prevents further registration. Wait runs without holding chatMu or Manager.mu; finalizers can acquire their normal locks.
11. An uncooperative tool/provider returns an explicit deadline failure instead of holding shutdown indefinitely.
12–14. Runtime cancellation/HTTP drain precede the background drain; the later-registered drain defer executes before the manager-close defer. The extra drain budget is 5 seconds, within the existing Stop 75-second bound together with HTTP 5 seconds and PostgreSQL 60 seconds.

New user admission also checks root cancellation immediately, then serializes its final durable insertion with chatClosing under chatMu. Rejected requests get HTTP503 and cannot leave an extra queued input.

## Real disposable database substitution

The test used ARTEX_TEST_DSN pointing only to a newly created scopeweaver_test_* database, with a deterministic provider that observes cancellation, waits on an explicit release channel, and writes a final conversation activity. No live model was called.

| Observed row / behavior | During blocked provider | After drain | After actual Close and DB reopen |
|---|---|---|---|
| run 81d21bc972cd98799d0c60f1b0f3dd76 | running | unknown | unknown |
| gate active run | present | null | persisted by FinishChatRun |
| gate paused | unchanged | true | preserved |
| queued follow-up count | 1 | 1 | not dispatched |
| finalization marker activity count | 0 | 1 | 1 |
| late user chat | HTTP503 | refused | no additional queued row |

An initial 20ms drain with the provider deliberately unreleased returned DeadlineExceeded. Releasing the provider then allowed a subsequent bounded drain to finish, including the real database writes. Both fresh databases were dropped by the helper, and the owned fixture runtime was stopped.

Counterexamples: an interrupted run is NOT converted to completed; its status remains unknown. A queued follow-up is NOT claimed during shutdown. A provider that ignores cancellation does NOT make drain wait forever. A request arriving after shutdown is NOT acknowledged with 202.

## Terminal shutdown error propagation

The previous Stop loop returned nil immediately after runtime.lock became free, even if Run returned a drain timeout. startup-error.txt is published after ownership release and may represent a prior startup, so reading it from Stop would be unsafe.

Now Run publishes a private shutdown-result.json with its existing random Instance nonce before releasing runtime.lock. Status retains that nonce from the authenticated, identity-checked health response. After its own authenticated shutdown request and ownership release, Stop reads only the matching result. The local HTTP regression proves a matching drain timeout returns an error, an older instance's error is ignored, and a successful instance ignores stale startup-error.txt.

Counterexample: a stale instance bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb cannot poison a successful stop of authenticated instance aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa. New runtime private metadata and authenticated health advertise shutdown_result_version=1. For that protocol, missing, malformed or mismatched receipts are errors, including a failed disk publication. Missing/stale-result compatibility applies only to legacy version0 runtimes. Six local HTTP cases verify current drain failure, required missing result, required stale result, legacy missing result, legacy stale result and success despite stale startup-error.txt.

## Evidence and limits

- before.jsonl: disabling only DrainChats with a Go source overlay fails `blocked drain: <nil>`.
- after-r3.jsonl: 9 focused fresh-DB dispatcher/queue/retest/shutdown tests pass, including both actual close/reopen readbacks. Original first-slice after.jsonl is preserved.
- before-stop.jsonl: old Stop overlay fails `stop lost shutdown error: <nil>`.
- unit.jsonl: portable suite passes. Server has261 passing test/subtest events,187 DB-dependent skips with no ARTEX_TEST_DSN, and1 inherited failure: TestChatBranchCopyRejectsDestinationSymlink cannot create a Windows symlink because the required privilege is missing. This check was not weakened or skipped.
- go vet ./server ./internal/portable and git diff --check pass.
- original-shutdown-observation.log and original-shutdown-recovery.json preserve the independent installed failure.

Installed real-model shutdown recheck is pending the parent's rebuilt candidate. The reviewed side-question and checkpoint writers are included; unrelated background producers remain outside this demonstrated defect. No changes were made to codexprovider, CLI cancellation, schema, user databases or credentials.


## R3 side-question and checkpoint extension

Independent review found two additional concrete writers: runSide terminal answer/usage persistence and the side publisher loop, whose cancellation branch flushes both side snapshots and chat contexts. Neither belongs to chatWG. Waiting only chatWG could return immediately when an admitted side question was the only active work.

The final drain now executes these numbered steps inside its existing bounded worker:

```go
1 s.side.commands.Lock()
2 s.side.closing = true
3 sideDone = s.cancelSideWhere(func(sidequestion.Parent) bool { return true })
4 s.side.commands.Unlock()
5 s.chatWG.Wait()
6 for _, ch := range sideDone { <-ch }
7 <-s.side.done
8 s.flushSideSnapshots(); s.flushChatContexts()
9 close(s.chatDrainDone)
```

1–4. Existing side admission and registration already hold commands, so closing and collecting completion channels cannot miss a newly registered side run. New POST requests also check root cancellation and closing under the same lock, returning503.
5. Main chats finish all their writes.
6. Each existing side run closes its own done channel after terminal answer/usage persistence.
7. The checkpoint loop closes side.done after its cancellation flush finishes.
8. Final flushing catches a checkpoint emitted by chat unwind after the publisher's initial cancellation flush. The publisher and tracked producers have now settled.
9. Only then may portable teardown close the database. The caller's same5-second context still bounds waiting; no Manager or DB mutex is held during completion waits.

Real disposable row: side request03323ade-4bd5-4abf-9fa5-4a749fbdda5a persisted status=cancelled, answer=shutdown side marker and input_tokens=19. Snapshot version2 with text=final shutdown checkpoint persisted too. Both were read again after actual manager Close and fresh DB connection reopen. Blocking the publisher produced DeadlineExceeded; releasing it while keeping the side run blocked produced DeadlineExceeded again. Releasing the side provider then completed the drain. A late side-question POST returned503.

Counterexample: no main chats active does NOT imply a safe database close while a side answer or checkpoint flush is still running. The old chat-only drain overlay fails incomplete side drain returned<nil> (before-r3.jsonl). The old optional-receipt overlay fails both required missing/stale result cases (before-result.jsonl). Final portable-r3.jsonl and fresh DB after-r3.jsonl pass. The independent narrow R3 review found no remaining concrete issue in these changes; installed real-model recheck belongs to the parent's rebuilt candidate.
