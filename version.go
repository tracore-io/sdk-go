package tracore

// Version is this SDK module's own semantic version. It is the version the publish
// workflow turns into the git tag vX.Y.Z (and thus the version `go get` resolves);
// it is independent of the API contract version below. sync.yml bumps it on regen.
const Version = "0.3.1"

// APIVersion is the Tracore API contract version this SDK was generated from — the
// `api-vX.Y.Z` release / .api-version, NOT this module's own Version. It is sent on
// every request as the `Tracore-Version` header and exposed as Client.APIVersion.
// Kept in sync with .api-version by sync.yml.
const APIVersion = "0.6.0"
