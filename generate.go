package tracore

// Regeneration entrypoint. `go generate ./...` runs these in order:
//  1. down-convert the 3.1 spec to 3.0 (oapi-codegen does not support 3.1 yet),
//  2. regenerate the low-level client into internal/generated.
// openapi.yaml is fetched + verified by sync.yml (gitignored); openapi.gen.yaml
// is the down-converted build artifact (also gitignored). The generated client
// (internal/generated/client.gen.go) IS committed and must never be hand-edited.
//
//go:generate go run ./cmd/downconvert openapi.yaml openapi.gen.yaml
//go:generate go tool oapi-codegen -config oapi-codegen.yaml openapi.gen.yaml
