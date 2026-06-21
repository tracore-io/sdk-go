package tracore

import (
	"context"
	"os"
	"testing"
)

// --- Unit tests (no network) ---

func TestNewRequiresAPIKey(t *testing.T) {
	if _, err := New(ClientOptions{}); err == nil {
		t.Fatal("expected error when APIKey is empty")
	}
}

func TestNewDefaults(t *testing.T) {
	c, err := New(ClientOptions{APIKey: "k"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.APIVersion != APIVersion {
		t.Fatalf("APIVersion = %q, want %q", c.APIVersion, APIVersion)
	}
	if c.Workspaces == nil || c.User == nil || c.User.ProviderKeys == nil {
		t.Fatal("resource groups not wired")
	}
}

func TestNormalizeError(t *testing.T) {
	body := []byte(`{"error":{"code":"NOT_FOUND","message":"Workspace not found","details":{"slug":"x"}}}`)
	te := normalizeError(404, body)
	if te.Status != 404 || te.Code != "NOT_FOUND" || te.Message != "Workspace not found" {
		t.Fatalf("unexpected: %+v", te)
	}
	if te.Details["slug"] != "x" {
		t.Fatalf("details not parsed: %+v", te.Details)
	}
	// Non-envelope body falls back to a status message.
	fb := normalizeError(500, []byte("oops"))
	if fb.Status != 500 || fb.Message == "" {
		t.Fatalf("fallback unexpected: %+v", fb)
	}
}

func TestAsTracoreError(t *testing.T) {
	var err error = &TracoreError{Status: 403, Code: "forbidden"}
	te, ok := AsTracoreError(err)
	if !ok || te.Status != 403 {
		t.Fatalf("AsTracoreError failed: %v %+v", ok, te)
	}
}

// --- Contract smoke tests (require a Prism mock; skipped if unreachable) ---
//
// Set TRACORE_MOCK_URL to the mock base URL. CI starts Prism on
// http://127.0.0.1:4010 from the committed spec. Locally, `go test` without a
// mock skips these (a Status-0 transport error means "not reachable").

func mockClient(t *testing.T) *Client {
	t.Helper()
	base := os.Getenv("TRACORE_MOCK_URL")
	if base == "" {
		base = "http://127.0.0.1:4010"
	}
	c, err := New(ClientOptions{APIKey: "test-key", BaseURL: base})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func skipIfUnreachable(t *testing.T, err error) {
	t.Helper()
	if te, ok := AsTracoreError(err); ok && te.Status == 0 {
		t.Skipf("mock server not reachable: %v", err)
	}
}

func TestContractWorkspacesList(t *testing.T) {
	c := mockClient(t)
	_, err := c.Workspaces.List(context.Background(), nil)
	skipIfUnreachable(t, err)
	if err != nil {
		t.Fatalf("Workspaces.List: %v", err)
	}
}

func TestContractUserPlan(t *testing.T) {
	c := mockClient(t)
	_, err := c.User.Plan(context.Background())
	skipIfUnreachable(t, err)
	if err != nil {
		t.Fatalf("User.Plan: %v", err)
	}
}
