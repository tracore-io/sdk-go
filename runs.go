package tracore

import (
	"context"

	"github.com/tracore-io/sdk-go/internal/generated"
)

// RunsResource lists, fetches, and polls extraction runs. Runs are
// environment-scoped.
type RunsResource struct{ c *Client }

// List returns the runs for a schema in an environment.
func (r *RunsResource) List(ctx context.Context, workspace, schemaKey string, params *ListParams) (*List[RunListItem], error) {
	var p ListParams
	if params != nil {
		p = *params
	}
	res, err := r.c.gen.GetRunsWithResponse(ctx, workspace, schemaKey, &generated.GetRunsParams{
		Env:      envEnum[generated.GetRunsParamsEnv](r.c.resolveEnv(p.Env)),
		Page:     intPtr(p.Page),
		PageSize: intPtr(p.PageSize),
	})
	if err != nil {
		return nil, transportError(err)
	}
	if res.JSON200 == nil {
		return nil, normalizeError(res.StatusCode(), res.Body)
	}
	return &List[RunListItem]{
		Data:     res.JSON200.Data,
		Page:     res.JSON200.Page,
		PageSize: res.JSON200.PageSize,
		Total:    res.JSON200.Total,
	}, nil
}

// Get returns a run by ID.
func (r *RunsResource) Get(ctx context.Context, id string) (*Run, error) {
	res, err := r.c.gen.GetRunByIdWithResponse(ctx, id)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}

// Poll blocks until a run reaches a terminal status (completed, failed, or
// validation_failed), ctx is cancelled, or the attempt budget is exhausted.
func (r *RunsResource) Poll(ctx context.Context, id string, opts *PollOptions) (*Run, error) {
	return pollRun(ctx, r.c, id, opts)
}
