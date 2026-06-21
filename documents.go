package tracore

import (
	"context"
	"io"

	"github.com/tracore-io/sdk-go/internal/generated"
)

// DocumentsResource manages documents. Documents are environment-scoped.
type DocumentsResource struct{ c *Client }

// List returns the documents for a schema in an environment.
func (r *DocumentsResource) List(ctx context.Context, workspace, schemaKey string, params *ListParams) (*List[Document], error) {
	var p ListParams
	if params != nil {
		p = *params
	}
	res, err := r.c.gen.GetDocumentsWithResponse(ctx, workspace, schemaKey, &generated.GetDocumentsParams{
		Env:      envEnum[generated.GetDocumentsParamsEnv](r.c.resolveEnv(p.Env)),
		Page:     intPtr(p.Page),
		PageSize: intPtr(p.PageSize),
	})
	if err != nil {
		return nil, transportError(err)
	}
	if res.JSON200 == nil {
		return nil, normalizeError(res.StatusCode(), res.Body)
	}
	return &List[Document]{
		Data:     res.JSON200.Data,
		Page:     res.JSON200.Page,
		PageSize: res.JSON200.PageSize,
		Total:    res.JSON200.Total,
	}, nil
}

// Get returns a document by ID.
func (r *DocumentsResource) Get(ctx context.Context, id string) (*Document, error) {
	res, err := r.c.gen.GetDocumentByIdWithResponse(ctx, id)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}

// Create creates a document from text content. env is the target environment
// ("" uses the client default).
func (r *DocumentsResource) Create(ctx context.Context, workspace, schemaKey string, body CreateDocumentRequest, env Env) (*Document, error) {
	res, err := r.c.gen.CreateDocumentWithResponse(ctx, workspace, schemaKey,
		&generated.CreateDocumentParams{Env: envEnum[generated.CreateDocumentParamsEnv](r.c.resolveEnv(env))}, body)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON201, res.StatusCode(), res.Body)
}

// Upload uploads a file as a document. env is the target environment ("" uses the
// client default).
func (r *DocumentsResource) Upload(ctx context.Context, workspace, schemaKey string, file io.Reader, name string, env Env) (*Document, error) {
	reqBody, contentType, err := multipartBody(file, name, nil)
	if err != nil {
		return nil, transportError(err)
	}
	res, err := r.c.gen.CreateDocumentWithBodyWithResponse(ctx, workspace, schemaKey,
		&generated.CreateDocumentParams{Env: envEnum[generated.CreateDocumentParamsEnv](r.c.resolveEnv(env))}, contentType, reqBody)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON201, res.StatusCode(), res.Body)
}

// Download returns the original document file bytes.
func (r *DocumentsResource) Download(ctx context.Context, id string) ([]byte, error) {
	res, err := r.c.gen.DownloadDocumentWithResponse(ctx, id)
	if err != nil {
		return nil, transportError(err)
	}
	if res.StatusCode() >= 300 {
		return nil, normalizeError(res.StatusCode(), res.Body)
	}
	return res.Body, nil
}

// Delete deletes a document, its stored file, and all associated runs.
func (r *DocumentsResource) Delete(ctx context.Context, id string) error {
	res, err := r.c.gen.DeleteDocumentWithResponse(ctx, id)
	if err != nil {
		return transportError(err)
	}
	if res.StatusCode() >= 300 {
		return normalizeError(res.StatusCode(), res.Body)
	}
	return nil
}
