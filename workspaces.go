package tracore

import (
	"context"

	"github.com/tracore-io/sdk-go/internal/generated"
)

// WorkspacesResource manages workspaces.
type WorkspacesResource struct{ c *Client }

// List returns the caller's workspaces.
func (r *WorkspacesResource) List(ctx context.Context, params *ListParams) (*List[Workspace], error) {
	var p ListParams
	if params != nil {
		p = *params
	}
	res, err := r.c.gen.GetWorkspacesWithResponse(ctx, &generated.GetWorkspacesParams{
		Page:     intPtr(p.Page),
		PageSize: intPtr(p.PageSize),
	})
	if err != nil {
		return nil, transportError(err)
	}
	if res.JSON200 == nil {
		return nil, normalizeError(res.StatusCode(), res.Body)
	}
	return &List[Workspace]{
		Data:     res.JSON200.Data,
		Page:     res.JSON200.Page,
		PageSize: res.JSON200.PageSize,
		Total:    res.JSON200.Total,
	}, nil
}

// Get returns a workspace by slug.
func (r *WorkspacesResource) Get(ctx context.Context, slug string) (*Workspace, error) {
	res, err := r.c.gen.GetWorkspaceBySlugWithResponse(ctx, slug)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}

// Create creates a workspace.
func (r *WorkspacesResource) Create(ctx context.Context, body CreateWorkspaceRequest) (*Workspace, error) {
	res, err := r.c.gen.CreateWorkspaceWithResponse(ctx, body)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON201, res.StatusCode(), res.Body)
}

// Update updates a workspace.
func (r *WorkspacesResource) Update(ctx context.Context, slug string, body UpdateWorkspaceRequest) (*Workspace, error) {
	res, err := r.c.gen.UpdateWorkspaceWithResponse(ctx, slug, body)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}

// Delete deletes a workspace and all of its schemas, documents, and runs.
func (r *WorkspacesResource) Delete(ctx context.Context, slug string) error {
	res, err := r.c.gen.DeleteWorkspaceWithResponse(ctx, slug)
	if err != nil {
		return transportError(err)
	}
	if res.StatusCode() >= 300 {
		return normalizeError(res.StatusCode(), res.Body)
	}
	return nil
}
