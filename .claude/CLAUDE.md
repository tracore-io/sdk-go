# sdk-go — repo guide

Official Go SDK for the Tracore API. Module path `github.com/tracore-io/sdk-go`,
root package `tracore`. Generated from the OpenAPI contract owned by the
`tracore-io/tracore` monorepo; this repo never hand-describes the API.

## Layout

- `internal/generated/client.gen.go` — low-level client, produced by `oapi-codegen`.
  Carries a `DO NOT EDIT` header, committed, **never hand-edited**. Import-restricted
  to this module (callers use the re-exported aliases in `types.go`).
- `client.go`, `errors.go`, `polling.go`, `types.go`, `*.go` resource files — the
  hand-written ergonomic layer (package `tracore`). This is what you extend.
- `cmd/downconvert/` — minimal OpenAPI 3.1 -> 3.0 rewrite (oapi-codegen does not yet
  support 3.1). Run before codegen by `go generate`.
- `examples/` — runnable, compile-checked usage examples.
- `version.go` — `Version` (this module's SemVer, drives the publish tag) and
  `APIVersion` (the API contract version, the `Tracore-Version` header).

## Regenerating

```bash
go generate ./...   # downconvert openapi.yaml -> openapi.gen.yaml, then run oapi-codegen
```

`openapi.yaml` (the spec) and `openapi.gen.yaml` (the 3.0 build artifact) are
gitignored; the generated client IS committed. In CI, `sync.yml` fetches + verifies
the pinned spec from the public release before regenerating.

## Conventions

- Every public method takes `context.Context` first and returns `(*T, error)` (or
  `error` for deletes). Failures are always `*TracoreError` (use `AsTracoreError`).
- Match the existing resource-file pattern: thin wrappers over the generated
  `…WithResponse` calls + `decode(...)` for the success/error split.
- gofmt-clean, `go vet`-clean. No emojis anywhere.
- Pin optional ints with `intPtr`, env enums with `envEnum[...]`.

## Pipeline

This repo subscribes to the `api-released` dispatch from the monorepo. `sync.yml`
regenerates and opens a draft `needs-polish` PR; the `polish-sdk-pr` skill finishes
it locally; a human merges; `publish.yml` tags `vX.Y.Z`. Full runbook:
`docs/releasing-sdks.md` in the monorepo. The `polish-sdk-pr` skill here is synced
read-only from the monorepo — never hand-edit it.
