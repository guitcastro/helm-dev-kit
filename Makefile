.PHONY: all build test clean install example coverage lint staticcheck security check docker-build docker-run test-integration

# Variables
BINARY_NAME=helm-dev-kit
VERSION?=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=${VERSION}"
COVERAGE_FILE=coverage.out

# Build the CLI tool
all: build

# Build the binary
build:
	go build ${LDFLAGS} -o ${BINARY_NAME} ./cmd/helm-dev-kit

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-amd64 ./cmd/helm-dev-kit
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-arm64 ./cmd/helm-dev-kit
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-amd64 ./cmd/helm-dev-kit
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-arm64 ./cmd/helm-dev-kit
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-windows-amd64.exe ./cmd/helm-dev-kit

# Run all tests
test:
	go test -v -cover ./...

# Run tests with race detection
test-race:
	go test -race -v ./...

# Run integration tests
test-integration:
	go test -v ./tests/...

# Generate test coverage
coverage:
	go test -coverprofile=${COVERAGE_FILE} ./...
	go tool cover -html=${COVERAGE_FILE} -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	rm -f ${BINARY_NAME}
	rm -rf output/
	rm -rf dist/
	rm -f ${COVERAGE_FILE}
	rm -f coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run the example (updated for directory-only processing)
example: build
	mkdir -p examples/sample-config
	cp examples/deployment.hcl examples/sample-config/
	cp examples/configmap.hcl examples/sample-config/
	mkdir -p output
	./${BINARY_NAME} examples/sample-config output
	@echo "Generated Helm chart in output/"
	rm -rf examples/sample-config

# Format code
fmt:
	go fmt ./...

# Run basic linter
lint:
	go vet ./...

# Run golangci-lint
golangci-lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
		echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Run staticcheck
staticcheck:
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not installed, skipping..."; \
		echo "Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

# Run security check
security:
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed, skipping..."; \
		echo "Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Run all quality checks
check: fmt golangci-lint staticcheck security test

# Build Docker image
docker-build:
	docker build -t ${BINARY_NAME}:${VERSION} .
	docker tag ${BINARY_NAME}:${VERSION} ${BINARY_NAME}:latest

# Run Docker container
docker-run: docker-build
	docker run --rm -v $(PWD)/examples:/input -v $(PWD)/output:/output ${BINARY_NAME}:latest /input /output

# Install the CLI tool
install: build
	cp ${BINARY_NAME} $(GOPATH)/bin/

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the CLI tool"
	@echo "  build-all       - Build for multiple platforms"
	@echo "  test            - Run all tests"
	@echo "  test-race       - Run tests with race detection"
	@echo "  test-integration- Run integration tests"
	@echo "  coverage        - Generate test coverage report"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Install and tidy dependencies"
	@echo "  example         - Run example conversion"
	@echo "  fmt             - Format code"
	@echo "  lint            - Run basic linter (go vet)"
	@echo "  golangci-lint   - Run golangci-lint"
	@echo "  staticcheck     - Run staticcheck"
	@echo "  security        - Run security check (gosec)"
	@echo "  check           - Run all quality checks"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run Docker container"
	@echo "  install         - Install CLI tool to GOPATH/bin"
	@echo "  help            - Show this help message"
