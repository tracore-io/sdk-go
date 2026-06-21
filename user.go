package tracore

import "context"

// UserResource groups user-scoped resources for the authenticated user.
type UserResource struct {
	c *Client
	// ProviderKeys manages the user's AI provider keys (BYOK).
	ProviderKeys *UserProviderKeysResource
}

// Plan returns the authenticated user's plan and current usage counters.
func (r *UserResource) Plan(ctx context.Context) (*UserPlanResponse, error) {
	res, err := r.c.gen.GetUserPlanWithResponse(ctx)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}

// UserProviderKeysResource manages the current user's AI provider keys (BYOK).
// Keys are user-scoped (one per provider, reused across the user's workspaces).
// These endpoints require session-cookie auth; API-key auth is rejected by design.
type UserProviderKeysResource struct{ c *Client }

// List returns the user's provider keys (masked previews, never plaintext).
func (r *UserProviderKeysResource) List(ctx context.Context) (*ProviderKeyListResponse, error) {
	res, err := r.c.gen.GetUserProviderKeysWithResponse(ctx)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}

// Get returns a single provider key's metadata.
func (r *UserProviderKeysResource) Get(ctx context.Context, provider ProviderType) (*ProviderKey, error) {
	res, err := r.c.gen.GetUserProviderKeyWithResponse(ctx, provider)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}

// Set creates, replaces, or updates the user's key for a provider. Supply
// ProviderKeyPutRequest.Key to test-and-store a key, or only DefaultModel to change
// the profile default model without re-encrypting.
func (r *UserProviderKeysResource) Set(ctx context.Context, provider ProviderType, body ProviderKeyPutRequest) (*ProviderKey, error) {
	res, err := r.c.gen.SetUserProviderKeyWithResponse(ctx, provider, body)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}

// Test re-tests the stored provider key against the provider and returns the new
// status.
func (r *UserProviderKeysResource) Test(ctx context.Context, provider ProviderType) (*ProviderKeyTestResponse, error) {
	res, err := r.c.gen.TestUserProviderKeyWithResponse(ctx, provider)
	if err != nil {
		return nil, transportError(err)
	}
	return decode(res.JSON200, res.StatusCode(), res.Body)
}

// Remove deletes the user's key for a provider, reverting any environment that
// referenced it to the system Mistral fallback.
func (r *UserProviderKeysResource) Remove(ctx context.Context, provider ProviderType) error {
	res, err := r.c.gen.DeleteUserProviderKeyWithResponse(ctx, provider)
	if err != nil {
		return transportError(err)
	}
	if res.StatusCode() >= 300 {
		return normalizeError(res.StatusCode(), res.Body)
	}
	return nil
}
