package tracore

import (
	"context"

	"github.com/tracore-io/sdk-go/internal/generated"
)

// WebhooksResource manages webhook endpoints and their deliveries. Endpoints are
// scoped to a schema and environment; delivery dispatch is environment-wide.
type WebhooksResource struct{ c *Client }

// List returns the webhook endpoints for a schema in an environment.
func (r *WebhooksResource) List(ctx context.Context, workspace, schemaKey string, params *ListParams) (*List[WebhookEndpoint], error) {
	var p ListParams
	if params != nil {
		p = *params
	}
	res, err := r.c.gen.GetWebhooksWithResponse(ctx, workspace, schemaKey, &generated.GetWebhooksParams{
		Env:      envEnum[generated.GetWebhooksParamsEnv](r.c.resolveEnv(p.Env)),
		Page:     intPtr(p.Page),
		PageSize: intPtr(p.PageSize),
	})
	if err != nil {
		return nil, transportError(err)
	}
	if res.JSON200 == nil {
		return nil, normalizeError(res.StatusCode(), res.Body)
	}
	return &List[WebhookEndpoint]{
		Data:     res.JSON200.Data,
		Page:     res.JSON200.Page,
		PageSize: res.JSON200.PageSize,
		Total:    res.JSON200.Total,
	}, nil
}

// Get returns a webhook endpoint by ID.
func (r *WebhooksResource) Get(ctx context.Context, id string) (*WebhookEndpoint, error) {
	res, err := r.c.gen.GetWebhookByIdWithResponse(ctx, id)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}

// Create creates a webhook endpoint for a schema in an environment. env is the
// target environment ("" uses the client default).
func (r *WebhooksResource) Create(ctx context.Context, workspace, schemaKey string, body CreateWebhookRequest, env Env) (*WebhookEndpoint, error) {
	res, err := r.c.gen.CreateWebhookWithResponse(ctx, workspace, schemaKey,
		&generated.CreateWebhookParams{Env: envEnum[generated.CreateWebhookParamsEnv](r.c.resolveEnv(env))}, body)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON201, res.StatusCode(), res.Body)
}

// Update updates a webhook endpoint.
func (r *WebhooksResource) Update(ctx context.Context, id string, body UpdateWebhookRequest) (*WebhookEndpoint, error) {
	res, err := r.c.gen.UpdateWebhookWithResponse(ctx, id, body)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}

// Delete deletes a webhook endpoint.
func (r *WebhooksResource) Delete(ctx context.Context, id string) error {
	res, err := r.c.gen.DeleteWebhookWithResponse(ctx, id)
	if err != nil {
		return transportError(err)
	}
	if res.StatusCode() >= 300 {
		return normalizeError(res.StatusCode(), res.Body)
	}
	return nil
}

// ListDeliveries returns the delivery attempts for a webhook endpoint.
func (r *WebhooksResource) ListDeliveries(ctx context.Context, id string, params *ListParams) (*List[WebhookDelivery], error) {
	var p ListParams
	if params != nil {
		p = *params
	}
	res, err := r.c.gen.GetWebhookDeliveriesWithResponse(ctx, id, &generated.GetWebhookDeliveriesParams{
		Page:     intPtr(p.Page),
		PageSize: intPtr(p.PageSize),
	})
	if err != nil {
		return nil, transportError(err)
	}
	if res.JSON200 == nil {
		return nil, normalizeError(res.StatusCode(), res.Body)
	}
	return &List[WebhookDelivery]{
		Data:     res.JSON200.Data,
		Page:     res.JSON200.Page,
		PageSize: res.JSON200.PageSize,
		Total:    res.JSON200.Total,
	}, nil
}

// SendTest dispatches a synthetic test event to a webhook endpoint. It reports
// whether the event was sent.
func (r *WebhooksResource) SendTest(ctx context.Context, id string, body SendTestWebhookRequest) (bool, error) {
	res, err := r.c.gen.SendTestWebhookWithResponse(ctx, id, body)
	if err != nil {
		return false, transportError(err)
	}
	if res.JSON200 == nil {
		return false, normalizeError(res.StatusCode(), res.Body)
	}
	return res.JSON200.Sent, nil
}
