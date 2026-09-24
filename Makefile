.PHONY: build test lint fmt generate prepare

# Worktree-agnostic build cache keys — see "Go in worktrees" in the root
# AGENTS.md. Exported so direct `go build`/`go test` recipes inherit it.
export GOFLAGS := -trimpath

prepare:
	@if [ -x ../../tools/bin/beak ]; then \
	  $(MAKE) generate; \
	elif [ ! -f internal/oapi/oapi.gen.go ] || [ ! -f eventtypes.gen.go ] || [ ! -f openenums.gen.go ] || [ ! -f callerrules.gen.go ] || [ ! -f audiences.gen.go ]; then \
	  echo "Generated Go SDK sources missing from standalone mirror" >&2; exit 1; \
	fi

build: prepare
	go build ./...
	@# Framework-composite examples are isolated modules (ADR-0054) — build each
	@# so a docs snippet can never reference code that no longer compiles.
	@for mod in examples/*/go.mod; do \
	  [ -f "$$mod" ] || continue; \
	  echo "  build $$(dirname $$mod)"; \
	  ( cd "$$(dirname $$mod)" && go build ./... ) || exit 1; \
	done

test: prepare
	go test ./...

lint: prepare
	golangci-lint run ./...

fmt: prepare
	gofumpt -w .
	gci write .

generate:
	../../tools/bin/beak run sdk-go:events
	../../tools/bin/beak run clients:caller-detection-generate
	../../tools/bin/beak run clients:sdk-go-generate
