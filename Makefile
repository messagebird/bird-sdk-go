.PHONY: build test lint fmt ci generate

# Worktree-agnostic build cache keys — see "Go in worktrees" in the root
# AGENTS.md. Exported so direct `go build`/`go test` recipes inherit it.
export GOFLAGS := -trimpath

build:
	go build ./...
	@# Framework-composite examples are isolated modules (ADR-0054) — build each
	@# so a docs snippet can never reference code that no longer compiles.
	@for mod in examples/*/go.mod; do \
	  [ -f "$$mod" ] || continue; \
	  echo "  build $$(dirname $$mod)"; \
	  ( cd "$$(dirname $$mod)" && go build ./... ) || exit 1; \
	done

test:
	go test -race ./...

lint:
	golangci-lint run ./...

fmt:
	gofumpt -w .
	gci write .

ci: lint test build

# Regenerate internal/oapi from the OpenAPI public bundle. The low-level
# client is filtered to the curated operations (oapi-codegen.yaml); models are
# kept whole so the webhook event union survives. Mirrors clients/api-go.
generate:
	@test -f ../../backend/openapi/.generated/openapi.public.bundle.yaml || \
		{ echo "Error: openapi bundle not found. Run 'make openapi-bundle' from the repo root first."; exit 1; }
	go run ../../backend/scripts/openapi-compat.go \
		--strip-go-type \
		../../backend/openapi/.generated/openapi.public.bundle.yaml \
		../../backend/openapi/.generated/openapi.compat.sdkgo.yaml
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.0 \
		-config oapi-codegen.yaml \
		../../backend/openapi/.generated/openapi.compat.sdkgo.yaml
