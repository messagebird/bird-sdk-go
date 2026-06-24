// Framework composite (ADR-0054): the Fiber handler example lives in its own
// module so the web-framework dependency never enters the SDK's go.mod. The
// replace points at the in-repo SDK; it is monorepo-only and never published.
module github.com/messagebird/bird-sdk-go/examples/fiber

go 1.24.0

require (
	github.com/gofiber/fiber/v2 v2.52.8
	github.com/messagebird/bird-sdk-go v0.0.0
)

require (
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/oapi-codegen/runtime v1.4.1 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
)

replace (
	github.com/messagebird-dev/bird/clients/conformance => ../../../conformance
	github.com/messagebird/bird-sdk-go => ../..
)
