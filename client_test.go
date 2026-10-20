package tracore

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/tracore-io/sdk-go/internal/generated"
)

// roundTripperFunc adapts a function to http.RoundTripper for stub transports.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// --- Unit tests (no network) ---

// Drift guard: every hand-written server-side ErrCode* constant must be a member
// of the contract's generated ApiErrorCode enum. Catches a constant surviving a
// contract rename/removal (the reverse direction — a new contract code without a
// public constant — is caught by regen review, not compilable assertions).
func TestServerErrCodesMatchContract(t *testing.T) {
	for _, code := range []string{
		ErrCodeNotFound,
		ErrCodeValidation,
		ErrCodeUnauthorized,
		ErrCodeInternal,
		ErrCodePlanLimitExceeded,
		ErrCodeRateLimited,
		ErrCodeForbiddenKey,
		ErrCodeInvalidProvider,
		ErrCodeInvalidModel,
		ErrCodeNoKey,
		ErrCodeNotImplemented,
		ErrCodeEmailNotVerified,
		ErrCodeStripeNotConfigured,
		ErrCodeNoStripeCustomer,
		ErrCodeCheckoutMissingUser,
		ErrCodeInvalidSignature,
	} {
		if !generated.ApiErrorCode(code).Valid() {
			t.Errorf("ErrCode %q is not in the contract's ApiErrorCode enum", code)
		}
	}
}

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

func TestTransportErrorPreservesCause(t *testing.T) {
	// errors.Is must see through the SDK wrapper to the original cause.
	cause := context.Canceled
	wrapped := transportError(cause)
	if wrapped.Code != ErrCodeTransport {
		t.Fatalf("Code = %q, want %q", wrapped.Code, ErrCodeTransport)
	}
	if !errors.Is(wrapped, context.Canceled) {
		t.Fatal("errors.Is(wrapped, context.Canceled) = false, want true")
	}
	// The cause survives end-to-end through the real request path. A stub
	// transport keeps this deterministic (no dial-vs-cancel race).
	c, err := New(ClientOptions{APIKey: "k", HTTPClient: &http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, context.Canceled
		}),
	}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.Workspaces.List(context.Background(), nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("errors.Is(err, context.Canceled) = false, err = %v", err)
	}
	// Pure HTTP envelope errors have no cause.
	if envErr := normalizeError(404, nil); errors.Unwrap(envErr) != nil {
		t.Fatalf("envelope error should have nil cause, got %v", errors.Unwrap(envErr))
	}
}

func TestExtractAsyncReturnsAcceptedEnvelope(t *testing.T) {
	var gets int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/extract"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"runId":"run_123","status":"pending"}`))
		case r.Method == http.MethodGet:
			gets++
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c, err := New(ClientOptions{APIKey: "k", BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	accepted, err := c.ExtractAsync(context.Background(), "ws", "schema",
		ExtractInput{DocumentID: "doc_1"}, nil)
	if err != nil {
		t.Fatalf("ExtractAsync: %v", err)
	}
	if accepted.RunId != "run_123" || string(accepted.Status) != "pending" {
		t.Fatalf("unexpected envelope: %+v", accepted)
	}
	if gets != 0 {
		t.Fatalf("ExtractAsync issued %d follow-up GETs, want 0", gets)
	}
}

func TestExtractAsyncRequiresInput(t *testing.T) {
	c, err := New(ClientOptions{APIKey: "k"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.ExtractAsync(context.Background(), "ws", "schema", ExtractInput{}, nil)
	te, ok := AsTracoreError(err)
	if !ok || te.Code != ErrCodeConfig {
		t.Fatalf("want config_error, got %v", err)
	}
}

var inlineSchema = map[string]any{
	"type":       "object",
	"properties": map[string]any{"total": map[string]any{"type": "number"}},
}

// captureExtract starts a server that records the extract request and answers 202.
func captureExtract(t *testing.T, inspect func(r *http.Request)) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inspect(r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"runId":"run_123","status":"pending"}`))
	}))
	t.Cleanup(srv.Close)
	c, err := New(ClientOptions{APIKey: "k", BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestExtractUploadSendsInlineSchemaBeforeFile(t *testing.T) {
	var parts []string
	got := map[string]string{}
	c := captureExtract(t, func(r *http.Request) {
		mr, err := r.MultipartReader()
		if err != nil {
			t.Errorf("MultipartReader: %v", err)
			return
		}
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Errorf("NextPart: %v", err)
				return
			}
			b, _ := io.ReadAll(p)
			parts = append(parts, p.FormName())
			got[p.FormName()] = string(b)
		}
	})

	_, err := c.ExtractAsync(context.Background(), "ws", "invoice",
		ExtractInput{File: strings.NewReader("%PDF-1.4"), FileName: "invoice.pdf"},
		&ExtractOptions{Schema: inlineSchema, SchemaName: "Invoice", Description: "Supplier invoices"})
	if err != nil {
		t.Fatalf("ExtractAsync: %v", err)
	}

	if len(parts) == 0 || parts[len(parts)-1] != "file" {
		t.Fatalf("file must be the last part, got order %v", parts)
	}
	var schema map[string]any
	if err := json.Unmarshal([]byte(got["schema"]), &schema); err != nil {
		t.Fatalf("schema field is not JSON: %q", got["schema"])
	}
	if !reflect.DeepEqual(schema, inlineSchema) {
		t.Fatalf("schema = %v, want %v", schema, inlineSchema)
	}
	if got["schemaName"] != "Invoice" || got["description"] != "Supplier invoices" || got["name"] != "invoice.pdf" {
		t.Fatalf("unexpected fields: %v", got)
	}
	if got["file"] != "%PDF-1.4" {
		t.Fatalf("file = %q", got["file"])
	}
}

func TestExtractByDocumentIDSendsInlineSchema(t *testing.T) {
	var body map[string]any
	c := captureExtract(t, func(r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
	})

	_, err := c.ExtractAsync(context.Background(), "ws", "invoice",
		ExtractInput{DocumentID: "doc_1"},
		&ExtractOptions{Schema: inlineSchema, SchemaName: "Invoice", Description: "Supplier invoices"})
	if err != nil {
		t.Fatalf("ExtractAsync: %v", err)
	}

	want := map[string]any{"schema": inlineSchema, "schemaName": "Invoice", "description": "Supplier invoices"}
	if !reflect.DeepEqual(body, want) {
		t.Fatalf("body = %v, want %v", body, want)
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
