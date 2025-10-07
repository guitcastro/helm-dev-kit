.PHONY: all build test clean install example

# Build the CLI tool
all: build

# Build the binary
build:
	go build -o helm-dev-kit ./cmd/helm-dev-kit

# Run all tests
test:
	go test -v -cover ./...

# Run tests with race detection
test-race:
	go test -race -v ./...

# Clean build artifacts
clean:
	rm -f helm-dev-kit
	rm -rf output/

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run the example
example: build
	mkdir -p output
	./helm-dev-kit examples/deployment.hcl output my-app
	@echo "Generated Helm chart in output/my-app"

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	go vet ./...

# Install the CLI tool
install: build
	cp helm-dev-kit $(GOPATH)/bin/

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the CLI tool"
	@echo "  test       - Run all tests"
	@echo "  test-race  - Run tests with race detection"
	@echo "  clean      - Clean build artifacts"
	@echo "  deps       - Install and tidy dependencies"
	@echo "  example    - Run example conversion"
	@echo "  fmt        - Format code"
	@echo "  lint       - Run linter"
	@echo "  install    - Install CLI tool to GOPATH/bin"
	@echo "  help       - Show this help message"
