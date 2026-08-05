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

## Resources

- `client.Workspaces` — manage workspaces
- `client.Schemas` — manage schema families and versions
- `client.Documents` — upload, download, and manage documents
- `client.Runs` — list, fetch, and poll extraction runs
- `client.Webhooks` — configure webhook endpoints and inspect deliveries
- `client.Environments` — manage environments
- `client.User.ProviderKeys` — manage the user's AI provider keys (BYOK)
- `client.Extract(...)` — one-call extract (by document ID or inline upload), with optional polling

## Error handling

Every call returns a `*tracore.TracoreError` on failure, carrying the HTTP status,
the API error code, and any structured details.

```go
ws, err := client.Workspaces.Get(ctx, "nonexistent")
if err != nil {
	if te, ok := tracore.AsTracoreError(err); ok {
		fmt.Println(te.Status)  // 404
		fmt.Println(te.Code)    // "NOT_FOUND"
		fmt.Println(te.Message) // "Workspace not found"
	}
}
```

## Versioning

This SDK is generated from the Tracore OpenAPI contract. Its module version
(`vX.Y.Z`, what `go get` resolves) follows SemVer; the API contract version it was
built against is shown above and sent on every request as the `Tracore-Version`
header. See [CHANGELOG.md](./CHANGELOG.md).

## License

MIT — see [LICENSE](./LICENSE).
