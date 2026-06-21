// Package tracore is the official Go SDK for the Tracore API — a schema-driven
// document-to-JSON extraction service.
//
//	client, err := tracore.New(tracore.ClientOptions{APIKey: os.Getenv("TRACORE_API_KEY")})
//	if err != nil { log.Fatal(err) }
//
//	run, err := client.Extract(ctx, "my-workspace", "invoice",
//		tracore.ExtractInput{DocumentID: "doc-123"},
//		&tracore.ExtractOptions{Poll: true})
//	if err != nil { log.Fatal(err) }
//	fmt.Println(run.Status, run.ExtractedData)
package tracore

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/tracore-io/sdk-go/internal/generated"
)

// ClientOptions configures a Client.
type ClientOptions struct {
	// APIKey authenticates every request (sent as the x-api-key header). Required.
	APIKey string
	// BaseURL overrides the API origin. Defaults to DefaultBaseURL.
	BaseURL string
	// Environment is the default environment for environment-scoped requests.
	// Individual calls may override it. Empty defers to the server default.
	Environment Env
	// HTTPClient overrides the underlying *http.Client (timeouts, proxies, etc.).
	HTTPClient *http.Client
}

// Client is the Tracore API client. Construct one with New and use its resource
// groups (Workspaces, Schemas, Documents, Runs, Webhooks, Environments, User) or
// the top-level Extract convenience.
type Client struct {
	gen        *generated.ClientWithResponses
	defaultEnv Env

	// APIVersion is the API contract version this SDK was generated from, sent as
	// the Tracore-Version header. It is the API version, not this module's Version.
	APIVersion string

	// Workspaces manages workspaces.
	Workspaces *WorkspacesResource
	// Schemas manages schema families and versions.
	Schemas *SchemasResource
	// Documents manages documents (upload, download, delete).
	Documents *DocumentsResource
	// Runs lists, fetches, and polls extraction runs.
	Runs *RunsResource
	// Webhooks manages webhook endpoints and deliveries.
	Webhooks *WebhooksResource
	// Environments manages environments within a workspace.
	Environments *EnvironmentsResource
	// User exposes user-scoped resources for the authenticated user.
	User *UserResource
}

// New constructs a Client. It returns an error only for invalid configuration; no
// network call is made here.
func New(opts ClientOptions) (*Client, error) {
	if opts.APIKey == "" {
		return nil, &TracoreError{Code: "config_error", Message: "APIKey is required"}
	}
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	editor := func(_ context.Context, req *http.Request) error {
		req.Header.Set("x-api-key", opts.APIKey)
		// Advisory header: tells the API which contract version this SDK targets.
		req.Header.Set("Tracore-Version", APIVersion)
		return nil
	}
	genOpts := []generated.ClientOption{generated.WithRequestEditorFn(editor)}
	if opts.HTTPClient != nil {
		genOpts = append(genOpts, generated.WithHTTPClient(opts.HTTPClient))
	}
	gen, err := generated.NewClientWithResponses(baseURL, genOpts...)
	if err != nil {
		return nil, transportError(err)
	}

	c := &Client{gen: gen, defaultEnv: opts.Environment, APIVersion: APIVersion}
	c.Workspaces = &WorkspacesResource{c}
	c.Schemas = &SchemasResource{c}
	c.Documents = &DocumentsResource{c}
	c.Runs = &RunsResource{c}
	c.Webhooks = &WebhooksResource{c}
	c.Environments = &EnvironmentsResource{c}
	c.User = &UserResource{c: c, ProviderKeys: &UserProviderKeysResource{c}}
	return c, nil
}

// resolveEnv applies the client default to an empty per-call env.
func (c *Client) resolveEnv(env Env) Env {
	if env != "" {
		return env
	}
	return c.defaultEnv
}

// envEnum casts a resolved env into a generated per-operation env enum pointer,
// or nil when unset (so the server applies its own default).
func envEnum[E ~string](env Env) *E {
	if env == "" {
		return nil
	}
	e := E(env)
	return &e
}

// intPtr returns a pointer to v, or nil when v is 0 (the "unset" sentinel for
// optional integer query params like page/pageSize).
func intPtr(v int) *int {
	if v == 0 {
		return nil
	}
	return &v
}

// ExtractInput selects what to extract from: either an existing document by ID
// (DocumentID) or an inline file upload (File + FileName). Provide exactly one.
type ExtractInput struct {
	// DocumentID extracts from an already-uploaded document.
	DocumentID string
	// File is the document content for an inline upload.
	File io.Reader
	// FileName is the display/file name for an inline upload.
	FileName string
}

// ExtractOptions tunes an Extract call.
type ExtractOptions struct {
	// Env overrides the client's default environment.
	Env Env
	// VersionNumber pins a schema version. 0 uses the latest.
	VersionNumber int
	// Model is an optional `<provider>/<model>` override for this run.
	Model string
	// Poll, when true, blocks until the run reaches a terminal status.
	Poll bool
	// PollOptions tunes the poll loop (only used when Poll is true).
	PollOptions *PollOptions
}

// Extract triggers an extraction run and, when ExtractOptions.Poll is set, waits
// for it to complete. It is a convenience over Documents.Upload + the schema
// extract endpoint + Runs.Poll.
func (c *Client) Extract(ctx context.Context, workspace, schemaKey string, input ExtractInput, opts *ExtractOptions) (*Run, error) {
	var o ExtractOptions
	if opts != nil {
		o = *opts
	}
	env := c.resolveEnv(o.Env)

	var runID string
	switch {
	case input.DocumentID != "":
		body := generated.DocumentExtractRequest{}
		if o.VersionNumber != 0 {
			body.VersionNumber = &o.VersionNumber
		}
		if o.Model != "" {
			body.Model = &o.Model
		}
		res, err := c.gen.ExtractDocumentByIdWithResponse(ctx, workspace, schemaKey, input.DocumentID,
			&generated.ExtractDocumentByIdParams{Env: envEnum[generated.ExtractDocumentByIdParamsEnv](env)}, body)
		if err != nil {
			return nil, transportError(err)
		}
		accepted, derr := decode(res.JSON202, res.StatusCode(), res.Body)
		if derr != nil {
			return nil, derr
		}
		runID = accepted.RunId
	case input.File != nil && input.FileName != "":
		fields := map[string]string{}
		if o.VersionNumber != 0 {
			fields["versionNumber"] = strconv.Itoa(o.VersionNumber)
		}
		if o.Model != "" {
			fields["model"] = o.Model
		}
		reqBody, contentType, err := multipartBody(input.File, input.FileName, fields)
		if err != nil {
			return nil, transportError(err)
		}
		res, err := c.gen.ExtractDocumentWithBodyWithResponse(ctx, workspace, schemaKey,
			&generated.ExtractDocumentParams{Env: envEnum[generated.ExtractDocumentParamsEnv](env)}, contentType, reqBody)
		if err != nil {
			return nil, transportError(err)
		}
		accepted, derr := decode(res.JSON202, res.StatusCode(), res.Body)
		if derr != nil {
			return nil, derr
		}
		runID = accepted.RunId
	default:
		return nil, &TracoreError{Code: "config_error", Message: "Extract requires either DocumentID or File+FileName"}
	}

	if o.Poll {
		return pollRun(ctx, c, runID, o.PollOptions)
	}
	res, err := c.gen.GetRunByIdWithResponse(ctx, runID)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}

// multipartBody builds a multipart/form-data body with a single "file" part plus
// a "name" field and any extra string fields. Returned alongside its content type.
func multipartBody(file io.Reader, name string, fields map[string]string) (io.Reader, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if err := w.WriteField("name", name); err != nil {
		return nil, "", err
	}
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			return nil, "", err
		}
	}
	part, err := w.CreateFormFile("file", name)
	if err != nil {
		return nil, "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, "", err
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return &buf, w.FormDataContentType(), nil
}
