.PHONY: build test lint dev ensure-ui-stub web-build generate e2e release-snapshot

VERSION := $(shell node -p "require('./package.json').version" 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

build: web-build
	go build -ldflags "$(LDFLAGS)" -o bin/dockyard ./cmd/dockyard

web-build:
	pnpm --dir web install
	pnpm --dir web build

test: ensure-ui-stub
	go test ./...

lint: ensure-ui-stub
	go vet ./...
	@if [ -f web/package.json ]; then pnpm --dir web lint; fi
	@./scripts/check-no-radix.sh

dev:
	@echo "Run: DOCKYARD_DEV_UI=http://localhost:3000 go run ./cmd/dockyard serve & pnpm --dir web dev"

ensure-ui-stub:
	@./scripts/ensure-ui-stub.sh

generate:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate

e2e:
	chmod +x e2e/smoke.sh
	./e2e/smoke.sh

e2e-ui: build
	chmod +x scripts/e2e-webserver.sh
	E2E_FRESH=1 pnpm --dir e2e install
	pnpm --dir e2e exec playwright install chromium
	E2E_FRESH=1 pnpm --dir e2e test

release-snapshot:
	goreleaser release --snapshot --clean --skip=publish
