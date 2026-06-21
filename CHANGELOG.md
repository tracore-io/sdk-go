# Changelog

All notable changes to the Tracore Go SDK are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); this SDK follows SemVer.
Each entry notes the Tracore API contract version (`api-vX.Y.Z`) it was built from.

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
