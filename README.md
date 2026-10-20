# sdk-go

<!-- tracore-api-version:start -->
> Built against Tracore API v0.6.0.
<!-- tracore-api-version:end -->

Official Go SDK for the [Tracore API](https://tracore.io) — schema-driven document-to-JSON extraction.

```bash
go get github.com/tracore-io/sdk-go@latest
```

```go
import tracore "github.com/tracore-io/sdk-go"
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	tracore "github.com/tracore-io/sdk-go"
)

func main() {
	client, err := tracore.New(tracore.ClientOptions{
		APIKey: os.Getenv("TRACORE_API_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// Create a workspace.
	ws, err := client.Workspaces.Create(ctx, tracore.CreateWorkspaceRequest{Name: "My Workspace"})
	if err != nil {
		log.Fatal(err)
	}

	// Define a schema.
	if _, err := client.Schemas.Create(ctx, ws.Slug, tracore.CreateSchemaFamilyRequest{
		SchemaKey: "invoice",
		Name:      "Invoice Schema",
		Definition: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"invoiceNumber": map[string]interface{}{"type": "string"},
				"total":         map[string]interface{}{"type": "number"},
			},
		},
	}); err != nil {
		log.Fatal(err)
	}

	// Extract from a file, polling until the run completes.
	f, err := os.Open("invoice.pdf")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	run, err := client.Extract(ctx, ws.Slug, "invoice",
		tracore.ExtractInput{File: f, FileName: "invoice.pdf"},
		&tracore.ExtractOptions{Poll: true})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(run.Status, run.ExtractedData)
}
```

### Inline schema

Skip the separate `Schemas.Create` call by passing the schema to `Extract`. The
schema family is created on first use; later calls create a new version only when
the schema changes.

```go
run, err := client.Extract(ctx, "my-workspace", "invoice",
	tracore.ExtractInput{File: f, FileName: "invoice.pdf"},
	&tracore.ExtractOptions{
		SchemaName: "Invoice",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"invoiceNumber": map[string]any{"type": "string"},
				"total":         map[string]any{"type": "number"},
			},
		},
		Poll: true,
	})
```

## Resources

- `client.Workspaces` — manage workspaces
- `client.Schemas` — manage schema families and versions
- `client.Documents` — upload, download, and manage documents
- `client.Runs` — list, fetch, and poll extraction runs
- `client.Webhooks` — configure webhook endpoints and inspect deliveries
- `client.Environments` — manage environments
- `client.User.ProviderKeys` — manage the user's AI provider keys (BYOK)
- `client.Extract(...)` — one-call extract (by document ID or inline upload), with optional polling
- `client.ExtractAsync(...)` — fire-and-forget extract returning the 202 accepted envelope (`runId`, `pending`)

Note that `Extract` is not idempotent across retries: it issues a billed POST plus
at least one follow-up GET, so retrying the whole call can start a second billed
run. Webhook-driven consumers should use `ExtractAsync` — it performs the single
POST and returns the accepted envelope as-is, so the run id is never lost to a
failed follow-up request:

```go
accepted, err := client.ExtractAsync(ctx, ws.Slug, "invoice",
	tracore.ExtractInput{DocumentID: docID}, nil)
if err != nil {
	// Status != 0: the API rejected the request, no run exists — safe to retry.
	// Status == 0 (ErrCodeTransport): outcome unknown — the connection may have
	// dropped after the server accepted; retry only if a duplicate run is acceptable.
}
store(accepted.RunId) // the run.completed webhook will quote this id
```

## Error handling

Every call returns a `*tracore.TracoreError` on failure, carrying the HTTP status,
the API error code, and any structured details.

```go
ws, err := client.Workspaces.Get(ctx, "nonexistent")
if err != nil {
	if te, ok := tracore.AsTracoreError(err); ok {
		fmt.Println(te.Status)  // 404
		fmt.Println(te.Code)    // tracore.ErrCodeNotFound
		fmt.Println(te.Message) // "Workspace not found"
	}
}
```

Error codes are exported as constants — the client-side synthetic codes
(`tracore.ErrCodeTransport` for network/decoding failures with `Status == 0`,
`tracore.ErrCodeConfig` for invalid client-side input, `tracore.ErrCodePollingTimeout`
for an exhausted poll loop) plus the server's `error.code` vocabulary
(`tracore.ErrCodeNotFound`, `tracore.ErrCodePlanLimitExceeded`, ...). Transport
failures keep their underlying
cause reachable through `errors.Is`/`errors.As` (`Unwrap`), so checks like
`errors.Is(err, context.Canceled)` or `net.Error` timeout detection work through
the SDK wrapper.

## Versioning

This SDK is generated from the Tracore OpenAPI contract. Its module version
(`vX.Y.Z`, what `go get` resolves) follows SemVer; the API contract version it was
built against is shown above and sent on every request as the `Tracore-Version`
header. See [CHANGELOG.md](./CHANGELOG.md).

## License

MIT — see [LICENSE](./LICENSE).
