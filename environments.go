package tracore

import (
	"context"

	"github.com/tracore-io/sdk-go/internal/generated"
)

// EnvironmentsResource manages the environments within a workspace. Each workspace
// has three: production (default) plus locked staging and development.
type EnvironmentsResource struct{ c *Client }

// List returns the environments in a workspace.
func (r *EnvironmentsResource) List(ctx context.Context, workspace string, params *ListParams) (*List[Environment], error) {
	var p ListParams
	if params != nil {
		p = *params
	}
	res, err := r.c.gen.GetEnvironmentsWithResponse(ctx, workspace, &generated.GetEnvironmentsParams{
		Page:     intPtr(p.Page),
		PageSize: intPtr(p.PageSize),
	})
	if err != nil {
		return nil, transportError(err)
	}
	if res.JSON200 == nil {
		return nil, normalizeError(res.StatusCode(), res.Body)
	}
	return &List[Environment]{
		Data:     res.JSON200.Data,
		Page:     res.JSON200.Page,
		PageSize: res.JSON200.PageSize,
		Total:    res.JSON200.Total,
	}, nil
}

// Get returns an environment by slug, including its resolved provider/model.
func (r *EnvironmentsResource) Get(ctx context.Context, workspace, envSlug string) (*EnvironmentWithResolved, error) {
	res, err := r.c.gen.GetEnvironmentBySlugWithResponse(ctx, workspace, generated.EnvironmentSlug(envSlug))
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}

// Update updates an environment's provider/model assignment.
func (r *EnvironmentsResource) Update(ctx context.Context, workspace, envSlug string, body UpdateEnvironmentRequest) (*EnvironmentWithResolved, error) {
	res, err := r.c.gen.UpdateEnvironmentWithResponse(ctx, workspace, generated.EnvironmentSlug(envSlug), body)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}
