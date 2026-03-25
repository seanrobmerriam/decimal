.PHONY: test bench lint clean build build-js build-wasm build-all install test-all fmt vet coverage benchmark help

# Default target
help:
	@echo "Decimal Money Library - Build Targets"
	@echo ""
	@echo "  make test        - Run all Go tests with race detection"
	@echo "  make bench       - Run Go benchmarks"
	@echo "  make lint        - Run linters (golint, vet)"
	@echo "  make fmt         - Format all Go code"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make build       - Build all Go packages"
	@echo "  make build-js    - Build JavaScript/TypeScript package"
	@echo "  make build-wasm  - Build Rust WebAssembly"
	@echo "  make build-all   - Build everything"
	@echo "  make coverage    - Generate test coverage report"
	@echo "  make install     - Install dependencies"

# Test targets
test:
	cd go && go test -v -race ./...

test-all: test test-js

test-js:
	cd js && npm test

# Benchmark targets
bench:
	cd go && go test -bench=. -benchmem -benchtime=1s ./...

benchmark: bench

# Lint and format
fmt:
	cd go && go fmt ./...
	cd rust && cargo fmt
	cd js && npm run format

lint:
	cd go && go vet ./...
	cd rust && cargo clippy -- -D warnings
	cd js && npm run lint

vet:
	cd go && go vet ./...

# Build targets
build:
	cd go && go build -v ./...

build-all: build build-js build-wasm

build-js:
	cd js && npm install && npm run build

build-wasm:
	cd rust && cargo build --target wasm32-unknown-unknown --release

build-release:
	cd go && go build -v -release ./...

# Dependencies
install:
	cd go && go mod download
	cd js && npm install
	cd rust && cargo fetch

# Coverage
coverage:
	cd go && go test -v -race -coverprofile=coverage.out ./...
	cd go && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at go/coverage.html"

# Clean
clean:
	cd go && go clean
	cd go && rm -f coverage.out coverage.html
	cd js && rm -rf dist node_modules/.cache
	cd rust && cargo clean

# CI simulation
ci: fmt lint test bench coverage

# Development helpers
dev:
	cd go && go run ./cmd/calculator

run-web:
	cd web && python3 -m http.server 8080

# Documentation
docs:
	cd go && go doc -all > docs/api.md

# Release (requires semantic versioning)
release-tag:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make release-tag VERSION=v1.2.3"; \
		exit 1; \
	fi
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

# Docker
docker-build:
	docker build -t decimal-money:latest .

docker-test:
	docker run decimal-money:latest go test ./...
