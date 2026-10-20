# Changelog

All notable changes to the Tracore Go SDK are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); this SDK follows SemVer.
Each entry notes the Tracore API contract version (`api-vX.Y.Z`) it was built from.

## 0.3.0 — api-v0.6.0

### Added

- `ExtractOptions.Schema`, `SchemaName`, `Description`: pass a JSON Schema inline to
  `Client.Extract` / `Client.ExtractAsync`. The API creates the schema family on
  first use and versions it by hash afterwards. Works for both `File`+`FileName`
  and `DocumentID` inputs. The API always accepted these fields; the SDK never
  exposed them, although the docs quickstart used them.

## 0.2.0 — api-v0.6.0

### Added

- `Client.ExtractAsync`: fire-and-forget extract returning the API's 202 accepted
  envelope (`ExtractAcceptedResponse`: run id + `pending`) with no follow-up GET,
  so webhook-driven consumers never lose the id of a live run (tracore#36).
- Exported error-code constants: client-side `ErrCodeTransport` / `ErrCodeConfig` /
  `ErrCodePollingTimeout` plus the full server `error.code` vocabulary from the
  contract's `ApiErrorCode` schema (`ErrCodeNotFound`, `ErrCodeValidation`,
  `ErrCodeUnauthorized`, `ErrCodeInternal`, `ErrCodePlanLimitExceeded`,
  `ErrCodeRateLimited`, `ErrCodeForbiddenKey`, `ErrCodeInvalidProvider`,
  `ErrCodeInvalidModel`, `ErrCodeNoKey`, plus internal billing/webhook-surface
  codes). Key-validation failures are `ErrCodeValidation` with
  `Details["code"] == "invalid_key"`, not a top-level code (tracore#38).
- `TracoreError.Unwrap`: transport errors keep their underlying cause, so
  `errors.Is(err, context.Canceled)` and `net.Error` timeout checks see through
  the SDK wrapper (tracore#39).
- Re-exported contract types `ApiErrorCode`, `WebhookEventType`, and
  `WebhookPayload` (typed webhook delivery body) from the regenerated client
  (tracore#38, tracore#40).

### Changed

- Regenerated from api-v0.6.0: `Run.confidence` is now a 64-bit float
  (`format: double`), and the generated types include the `ApiErrorCode` and
  `WebhookPayload` schemas (tracore#37, tracore#40).
- `Extract` documents that retrying it is not idempotent (billed POST + follow-up
  GET); use `ExtractAsync` when you only need the run id.
- Poll-loop exhaustion now returns `Status 0` with code `ErrCodePollingTimeout`
  instead of a fabricated HTTP 408 — consistent with the `Status 0` = client-side
  failure semantics. Callers matching on `Status == 408` must switch to the code.

## 0.1.0 — api-v0.5.0

Initial release.

### Added

- `tracore.Client` with resource groups: `Workspaces`, `Schemas`, `Documents`,
  `Runs`, `Webhooks`, `Environments`, and `User.ProviderKeys`.
- `Client.Extract` convenience: extract by document ID or inline file upload, with
  optional polling to a terminal run status.
- `Runs.Poll` and `PollOptions` (interval, max attempts, progress callback).
- Typed errors via `TracoreError` and `AsTracoreError`.
- Generated low-level client (`internal/generated`) produced by `oapi-codegen` from
  the OpenAPI contract, regenerated on every API release.
