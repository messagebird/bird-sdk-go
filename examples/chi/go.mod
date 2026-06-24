// Framework composite (ADR-0054): the chi handler example lives in its own
// module so the web-framework dependency never enters the SDK's go.mod. The
// replace points at the in-repo SDK; it is monorepo-only and never published.
module github.com/messagebird/bird-sdk-go/examples/chi

go 1.24.0

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/messagebird/bird-sdk-go v0.0.0
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/oapi-codegen/runtime v1.4.1 // indirect
)

replace (
	github.com/messagebird-dev/bird/clients/conformance => ../../../conformance
	github.com/messagebird/bird-sdk-go => ../..
)
