// Command examples demonstrates the Tracore Go SDK end to end. It is compiled by
// CI (so it never goes stale) but is meant to be read, not run as-is — it needs a
// real TRACORE_API_KEY and an invoice.pdf in the working directory.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	tracore "github.com/tracore-io/sdk-go"
)

func main() {
	client, err := tracore.New(tracore.ClientOptions{
		APIKey:      os.Getenv("TRACORE_API_KEY"),
		Environment: tracore.EnvProduction,
	})
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// 1. Create a workspace.
	ws, err := client.Workspaces.Create(ctx, tracore.CreateWorkspaceRequest{Name: "My Workspace"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("workspace:", ws.Slug)

	// 2. List workspaces (paginated).
	page, err := client.Workspaces.List(ctx, &tracore.ListParams{Page: 1, PageSize: 20})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("you have %d workspace(s)\n", page.Total)

	// 3. Define a schema.
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

	// 4. Extract from a file, polling until the run completes.
	f, err := os.Open("invoice.pdf")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	run, err := client.Extract(ctx, ws.Slug, "invoice",
		tracore.ExtractInput{File: f, FileName: "invoice.pdf"},
		&tracore.ExtractOptions{
			Poll: true,
			PollOptions: &tracore.PollOptions{
				Interval: 2 * time.Second,
				OnProgress: func(r *tracore.Run) {
					fmt.Println("status:", r.Status)
				},
			},
		})
	if err != nil {
		// Typed error handling.
		if te, ok := tracore.AsTracoreError(err); ok {
			log.Fatalf("extract failed: status=%d code=%s msg=%s", te.Status, te.Code, te.Message)
		}
		log.Fatal(err)
	}

	if run.Status == tracore.RunStatusCompleted && run.ExtractedData != nil {
		fmt.Println("extracted:", *run.ExtractedData)
	}
}
