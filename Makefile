.PHONY: build test lint dev ensure-ui-stub web-build generate

build: web-build
	go build -o bin/dockyard ./cmd/dockyard

web-build:
	pnpm --dir web install
	pnpm --dir web build

test: ensure-ui-stub
	go test ./...

lint:
	go vet ./...
	@if [ -f web/package.json ]; then pnpm --dir web lint; fi
	@./scripts/check-no-radix.sh

dev:
	@echo "Run: DOCKYARD_DEV_UI=http://localhost:3000 go run ./cmd/dockyard serve & pnpm --dir web dev"

ensure-ui-stub:
	@./scripts/ensure-ui-stub.sh

generate:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate
