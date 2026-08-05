package tracore

import "github.com/tracore-io/sdk-go/internal/generated"

// DefaultBaseURL is the production Tracore API origin used when ClientOptions.BaseURL
// is empty.
const DefaultBaseURL = "https://api.tracore.io"

// Env names a Tracore environment for environment-scoped requests. The zero value
// ("") means "use the client default" (and, absent that, the server default of
// production).
type Env string

const (
	EnvProduction  Env = "production"
	EnvStaging     Env = "staging"
	EnvDevelopment Env = "development"
)

// ListParams are the optional query parameters for paginated, environment-scoped
// list endpoints. A nil *ListParams (or zero fields) requests server defaults.
type ListParams struct {
	// Page is the 1-based page number. 0 means the server default.
	Page int
	// PageSize is the items per page (1-100). 0 means the server default.
	PageSize int
	// Env overrides the client's default environment for this call.
	Env Env
}

// List is a paginated result set. It mirrors the API's PaginatedResponse envelope.
type List[T any] struct {
	Data     []T `json:"data"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

// Re-exported model types. Callers use these instead of the internal generated
// package (which is import-restricted to this module). They are type aliases, so
// they are identical to the generated types — no conversion is needed.
type (
	Workspace                 = generated.Workspace
	CreateWorkspaceRequest    = generated.CreateWorkspaceRequest
	UpdateWorkspaceRequest    = generated.UpdateWorkspaceRequest
	SchemaFamily              = generated.SchemaFamily
	SchemaFamilyWithVersion   = generated.SchemaFamilyWithVersion
	SchemaVersion             = generated.SchemaVersion
	CreateSchemaFamilyRequest = generated.CreateSchemaFamilyRequest
	Document                  = generated.Document
	CreateDocumentRequest     = generated.CreateDocumentRequest
	Run                       = generated.Run
	RunListItem               = generated.RunListItem
	RunStatus                 = generated.RunStatus
	WebhookEndpoint           = generated.WebhookEndpoint
	WebhookDelivery           = generated.WebhookDelivery
	CreateWebhookRequest      = generated.CreateWebhookRequest
	UpdateWebhookRequest      = generated.UpdateWebhookRequest
	SendTestWebhookRequest    = generated.SendTestWebhookRequest
	Environment               = generated.Environment
	EnvironmentWithResolved   = generated.EnvironmentWithResolved
	UpdateEnvironmentRequest  = generated.UpdateEnvironmentRequest
	ProviderType              = generated.ProviderType
	ProviderKey               = generated.ProviderKey
	ProviderKeyPutRequest     = generated.ProviderKeyPutRequest
	ProviderKeyListResponse   = generated.ProviderKeyListResponse
	ProviderKeyTestResponse   = generated.ProviderKeyTestResponse
	UserPlanResponse          = generated.UserPlanResponse
	DocumentExtractRequest    = generated.DocumentExtractRequest
	ExtractAcceptedResponse   = generated.ExtractAcceptedResponse
	ApiErrorCode              = generated.ApiErrorCode
	WebhookEventType          = generated.WebhookEventType
	WebhookPayload            = generated.WebhookPayload
)

// Run statuses. A run in Completed, Failed, or ValidationFailed is terminal.
const (
	RunStatusCompleted        = generated.RunStatusCompleted
	RunStatusFailed           = generated.RunStatusFailed
	RunStatusValidationFailed = generated.RunStatusValidationFailed
	RunStatusPending          = generated.RunStatusPending
	RunStatusProcessing       = generated.RunStatusProcessing
)
