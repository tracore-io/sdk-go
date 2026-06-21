package tracore

import (
	"context"

	"github.com/tracore-io/sdk-go/internal/generated"
)

// SchemasResource manages schema families and their immutable versions. Schemas are
// workspace-scoped (shared across environments).
type SchemasResource struct{ c *Client }

// List returns the schema families in a workspace.
func (r *SchemasResource) List(ctx context.Context, workspace string, params *ListParams) (*List[SchemaFamily], error) {
	var p ListParams
	if params != nil {
		p = *params
	}
	res, err := r.c.gen.GetSchemaFamiliesWithResponse(ctx, workspace, &generated.GetSchemaFamiliesParams{
		Page:     intPtr(p.Page),
		PageSize: intPtr(p.PageSize),
	})
	if err != nil {
		return nil, transportError(err)
	}
	if res.JSON200 == nil {
		return nil, normalizeError(res.StatusCode(), res.Body)
	}
	return &List[SchemaFamily]{
		Data:     res.JSON200.Data,
		Page:     res.JSON200.Page,
		PageSize: res.JSON200.PageSize,
		Total:    res.JSON200.Total,
	}, nil
}

// Get returns a schema family by key.
func (r *SchemasResource) Get(ctx context.Context, workspace, schemaKey string) (*SchemaFamily, error) {
	res, err := r.c.gen.GetSchemaFamilyByKeyWithResponse(ctx, workspace, schemaKey)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}

// Create creates a schema family and its first version. Returns the family with
// the resolved version.
func (r *SchemasResource) Create(ctx context.Context, workspace string, body CreateSchemaFamilyRequest) (*SchemaFamilyWithVersion, error) {
	res, err := r.c.gen.CreateSchemaFamilyWithResponse(ctx, workspace, body)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}

// ListVersions returns the versions of a schema family.
func (r *SchemasResource) ListVersions(ctx context.Context, workspace, schemaKey string, params *ListParams) (*List[SchemaVersion], error) {
	var p ListParams
	if params != nil {
		p = *params
	}
	res, err := r.c.gen.GetSchemaVersionsWithResponse(ctx, workspace, schemaKey, &generated.GetSchemaVersionsParams{
		Page:     intPtr(p.Page),
		PageSize: intPtr(p.PageSize),
	})
	if err != nil {
		return nil, transportError(err)
	}
	if res.JSON200 == nil {
		return nil, normalizeError(res.StatusCode(), res.Body)
	}
	return &List[SchemaVersion]{
		Data:     res.JSON200.Data,
		Page:     res.JSON200.Page,
		PageSize: res.JSON200.PageSize,
		Total:    res.JSON200.Total,
	}, nil
}

// GetVersion returns a specific schema version by number.
func (r *SchemasResource) GetVersion(ctx context.Context, workspace, schemaKey string, versionNumber int) (*SchemaVersion, error) {
	res, err := r.c.gen.GetSchemaVersionWithResponse(ctx, workspace, schemaKey, versionNumber)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}
