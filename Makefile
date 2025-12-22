.PHONY: build install generate-sdks test clean help

# Default version for local development
VERSION ?= 0.1.0

# Build the provider binary
build:
	@echo "Building pulumi-resource-dex (version: $(VERSION))..."
	@mkdir -p bin
	@go build -ldflags "-X 'github.com/kotaicode/pulumi-provider-dex/pkg/provider.Version=$(VERSION)'" -o bin/pulumi-resource-dex ./cmd/pulumi-resource-dex
	@echo "✓ Built bin/pulumi-resource-dex"

# Install the provider locally (requires Pulumi CLI)
install: build
	@echo "Installing provider..."
	@pulumi plugin install resource dex v0.1.0 --file bin/pulumi-resource-dex || true
	@echo "✓ Provider installed"

# Generate schema.json (requires Pulumi CLI)
# First install the provider, then get its schema, then fix metadata
generate-schema: build
	@echo "Generating schema.json..."
	@pulumi package get-schema ./bin/pulumi-resource-dex > schema.json || (echo "⚠ Schema generation failed (Pulumi CLI required)" && exit 1)

# Generate language SDKs (requires Pulumi CLI)
# Use schema.json if available, otherwise install provider and use plugin name
generate-sdks: build
	@echo "Generating SDKs..."
	@mkdir -p sdk
	@if [ -f schema.json ]; then \
		echo "Using schema.json for SDK generation..."; \
		pulumi package gen-sdk schema.json --language typescript --out sdk/typescript || echo "⚠ TypeScript SDK generation failed"; \
		pulumi package gen-sdk schema.json --language go --out sdk/go || echo "⚠ Go SDK generation failed"; \
		pulumi package gen-sdk schema.json --language python --out sdk/python || echo "⚠ Python SDK generation failed"; \
	else \
		echo "schema.json not found, installing provider and using plugin name..."; \
		pulumi plugin install resource dex v0.1.0 --file bin/pulumi-resource-dex 2>/dev/null || true; \
		pulumi package gen-sdk dex --language typescript --out sdk/typescript || echo "⚠ TypeScript SDK generation failed"; \
		pulumi package gen-sdk dex --language go --out sdk/go || echo "⚠ Go SDK generation failed"; \
		pulumi package gen-sdk dex --language python --out sdk/python || echo "⚠ Python SDK generation failed"; \
	fi
	@echo "✓ SDKs generated in sdk/"

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf sdk/
	@rm -f pulumi-resource-dex
	@echo "✓ Cleaned"

# Start local Dex for testing
dex-up:
	@echo "Starting Dex with docker-compose..."
	@docker-compose up -d
	@echo "✓ Dex started at localhost:5557 (gRPC) and http://localhost:5556 (web)"

# Stop local Dex
dex-down:
	@echo "Stopping Dex..."
	@docker-compose down
	@echo "✓ Dex stopped"

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the provider binary"
	@echo "  install         - Build and install the provider locally"
	@echo "  generate-schema - Generate schema.json (requires Pulumi CLI)"
	@echo "  generate-sdks   - Generate language SDKs (requires Pulumi CLI)"
	@echo "  test            - Run tests"
	@echo "  clean           - Clean build artifacts"
	@echo "  dex-up          - Start local Dex with docker-compose"
	@echo "  dex-down        - Stop local Dex"
	@echo "  help            - Show this help message"

