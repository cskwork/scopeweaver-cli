The failed OAuth rerun exposed a fixture mismatch: the planner bound an IP asset to a task that allowed only one exact URL. The gate correctly rejected that broader asset. No provider or product gate was changed.

Normally, the planner creates an intent bound to target assets; add_intent validates those assets before creating any worker work. The fixture created a task from a URL, and server/server.go seed() registered the host as an IP asset. It did not register the precise endpoint. In the 2026-10-17 13:46:47 +09:00 run, task 1 found no endpoint row, chose asset 1, and received ids=[0]. No worker or HTTP GET followed. The run failed after 463.79 seconds. oauth-intent-22.json and oauth-intent-23.json preserve the exact request and rejection.

Relevant as-is setup:

    target := fixture.URL + "/synthetic-debug"
    // tasks create had the URL in description and an exact_url allow rule,
    // but no registered endpoint or attached company scope.

1. The URL describes the test target but does not itself create an endpoint row.
2. Automatic seeding creates the host IP. Binding that row gives the gate IP candidates, with no URL candidate.

To-be setup (relevant statements; unchanged task fields omitted):

    company := submit(companyWithLoopbackScope, "companies", "create")
    registered := submit(exactGETEndpointWithLoopbackIP, "assets", "insert")
    listed := run("assets", "list", "--type", "endpoint", "--dsl", "url=="+target)
    // Verify the returned ID, exact URL, GET method and company attribution.
    // tasks create now supplies company_ids=[companyID].
    // Its description tells the planner to verify and bind only endpointID.

1. The company supplies deterministic asset visibility to this disposable task.
2. Registering the precise endpoint supplies the URL candidate required by the existing allow rule.
3. Readback verifies actual persisted data, not a guessed asset ID.
4. Task creation attaches company scope and its asset snapshot inside the existing transaction, before planning starts; it cannot race a later attach command.
5. The planner is given the verified endpoint ID and still must make its own real add_intent call. The unchanged exact_url rule continues to reject broader IP assets.

Real failed row substituted into existing gate logic:

| Candidate | IP candidates | URL candidates | exact_url result |
| --- | --- | --- | --- |
| Observed asset 1, type=ip | 127.0.0.1 | empty | denied |
| Correct endpoint row required by fixture | 127.0.0.1 | http://127.0.0.1:63134/synthetic-debug | allowed |
| Counterexample, same IP and port | 127.0.0.1 | http://127.0.0.1:63134/unrelated | denied |
| Counterexample, another port | 127.0.0.1 | http://127.0.0.1:63135/synthetic-debug | denied |

In agent/tools.go addOneIntent, CheckAssetsInterceptForLanguage loads the bound row. db/asset_intercept_match.go interceptCandidates returns no URLs for the observed IP row. matchOne exact_url iterates URL candidates only, so it finds no match. EvaluateAssetGate then rejects the allowlist miss. A URL mentioned in the intent summary is not substituted for the asset's URL. This distinction preserves the scope check.

TestCodexFixtureExactEndpointAllowRule passes all four table cases through the existing exported EvaluateAssetGate function. It makes no model, database or HTTP calls. Existing DB gate tests and CLI unit/vet checks are recorded separately. The corrected full live workflow still requires the parent's final-candidate run; this explanation does not count that pending run as passing.
