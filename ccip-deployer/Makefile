# CCIP Deployer Makefile

.PHONY: build run clean test deps tidy dev dev-config dev-verbose dev-race

# Variables
BINARY_NAME=ccip-deployer
CONFIG_FILE=config.yaml

# Build the application
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) .

# Run the application with default config
run: build
	@echo "🚀 Running $(BINARY_NAME)..."
	./bin/$(BINARY_NAME) $(CONFIG_FILE)

# Run with a specific config file
run-config: build
	@echo "🚀 Running $(BINARY_NAME) with config: $(CONFIG)"
	./bin/$(BINARY_NAME) $(CONFIG)

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -rf bin/

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	go mod download

# Tidy dependencies
tidy:
	@echo "🧽 Tidying dependencies..."
	go mod tidy

# Install the binary
install: build
	@echo "📦 Installing $(BINARY_NAME)..."
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/

# Create sample config
sample-config:
	@echo "📝 Creating sample config..."
	@echo "home_chain:" > config.yaml
	@echo "  chain_id: 11155111  # Sepolia" >> config.yaml
	@echo "  name: \"Ethereum Sepolia\"" >> config.yaml
	@echo "  rpc_url: \"https://sepolia.infura.io/v3/YOUR_PROJECT_ID\"" >> config.yaml
	@echo "  private_key: \"YOUR_PRIVATE_KEY_HEX\"" >> config.yaml
	@echo "" >> config.yaml
	@echo "feed_chain:" >> config.yaml
	@echo "  chain_id: 43113  # Avalanche Fuji" >> config.yaml
	@echo "  name: \"Avalanche Fuji\"" >> config.yaml
	@echo "  rpc_url: \"https://api.avax-test.network/ext/bc/C/rpc\"" >> config.yaml
	@echo "  private_key: \"YOUR_PRIVATE_KEY_HEX\"" >> config.yaml
	@echo "" >> config.yaml
	@echo "chains:" >> config.yaml
	@echo "  - chain_id: 80002  # Polygon Amoy" >> config.yaml
	@echo "    name: \"Polygon Amoy\"" >> config.yaml
	@echo "    rpc_url: \"https://rpc-amoy.polygon.technology\"" >> config.yaml
	@echo "    private_key: \"YOUR_PRIVATE_KEY_HEX\"" >> config.yaml
	@echo "" >> config.yaml
	@echo "deployment:" >> config.yaml
	@echo "  rmn_enabled: true" >> config.yaml
	@echo "  fresh_deployment: true" >> config.yaml
	@echo "" >> config.yaml
	@echo "existing_contracts: {}" >> config.yaml
	@echo "✅ Sample config created as config.yaml"
	@echo "⚠️  Please update the RPC URLs and private keys before running!"

# Help
help:
	@echo "CCIP Deployer Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build          Build the application"
	@echo "  make run            Build and run with default config"
	@echo "  make run-config CONFIG=file.yaml   Run with specific config"
	@echo ""
	@echo "Development (with traces):"
	@echo "  make dev            Run with go run (better stack traces)"
	@echo "  make dev-config CONFIG=file.yaml   Run with go run and specific config"
	@echo "  make dev-verbose    Run with debug logging for detailed traces"
	@echo "  make dev-race       Run with race detection"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean          Clean build artifacts"
	@echo "  make test           Run tests"
	@echo "  make deps           Download dependencies"
	@echo "  make tidy           Tidy dependencies"
	@echo "  make install        Install binary to GOPATH/bin"
	@echo "  make sample-config  Create sample configuration file"
	@echo "  make help           Show this help"

# Run with go run for better traces (development)
dev: 
	@echo "🔧 Running in development mode with go run..."
	go run . $(CONFIG_FILE)

# Run with go run and specific config (development)
dev-config:
	@echo "🔧 Running in development mode with config: $(CONFIG)"
	go run . $(CONFIG)

# Run with verbose logging and traces
dev-verbose:
	@echo "🔧 Running in verbose development mode..."
	@echo "Setting log level to DEBUG for detailed traces..."
	CL_LOG_LEVEL=debug go run . $(CONFIG_FILE)

# Run with race detection (development)
dev-race:
	@echo "🔧 Running with race detection..."
	go run -race . $(CONFIG_FILE) 
	@echo "  - chain_id: 80002  # Polygon Amoy" >> config.yaml
	@echo "    name: \"Polygon Amoy\"" >> config.yaml
	@echo "    rpc_url: \"https://rpc-amoy.polygon.technology\"" >> config.yaml
	@echo "    private_key: \"YOUR_PRIVATE_KEY_HEX\"" >> config.yaml
	@echo "" >> config.yaml
	@echo "deployment:" >> config.yaml
	@echo "  rmn_enabled: true" >> config.yaml
	@echo "  fresh_deployment: true" >> config.yaml
	@echo "" >> config.yaml
	@echo "existing_contracts: {}" >> config.yaml
	@echo "✅ Sample config created as config.yaml"
	@echo "⚠️  Please update the RPC URLs and private keys before running!"

# Help
help:
	@echo "CCIP Deployer Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build          Build the application"
	@echo "  make run            Build and run with default config"
	@echo "  make run-config CONFIG=file.yaml   Run with specific config"
	@echo ""
	@echo "Development (with traces):"
	@echo "  make dev            Run with go run (better stack traces)"
	@echo "  make dev-config CONFIG=file.yaml   Run with go run and specific config"
	@echo "  make dev-verbose    Run with debug logging for detailed traces"
	@echo "  make dev-race       Run with race detection"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean          Clean build artifacts"
	@echo "  make test           Run tests"
	@echo "  make deps           Download dependencies"
	@echo "  make tidy           Tidy dependencies"
	@echo "  make install        Install binary to GOPATH/bin"
	@echo "  make sample-config  Create sample configuration file"
	@echo "  make help           Show this help"

# Run with go run for better traces (development)
dev: 
	@echo "🔧 Running in development mode with go run..."
	go run . $(CONFIG_FILE)

# Run with go run and specific config (development)
dev-config:
	@echo "🔧 Running in development mode with config: $(CONFIG)"
	go run . $(CONFIG)

# Run with verbose logging and traces
dev-verbose:
	@echo "🔧 Running in verbose development mode..."
	@echo "Setting log level to DEBUG for detailed traces..."
	CL_LOG_LEVEL=debug go run . $(CONFIG_FILE)

# Run with race detection (development)
dev-race:
	@echo "🔧 Running with race detection..."
	go run -race . $(CONFIG_FILE) 
	@echo "CCIP Deployer Commands:"
	@echo "  make setup          - Initialize Go modules"
	@echo "  make deploy         - Run full automatic deployment"
	@echo "  make deploy-interactive - Run interactive step-by-step deployment" 
	@echo "  make deploy-auto    - Run automatic step-by-step deployment"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Configuration:"
	@echo "  Edit config.yaml with your testnet details before deploying"

# Setup Go modules
setup:
	go mod tidy
	@echo "✅ Dependencies installed"

# Full automatic deployment
deploy: setup
	@echo "🚀 Starting full CCIP deployment..."
	go run main.go config.yaml

# Interactive step-by-step deployment
deploy-interactive: setup
	@echo "🚀 Starting interactive CCIP deployment..."
	go run deploy_step_by_step.go interactive config.yaml

# Automatic step-by-step deployment  
deploy-auto: setup
	@echo "🚀 Starting automatic step-by-step CCIP deployment..."
	go run deploy_step_by_step.go auto config.yaml

# Custom config deployment
deploy-config: setup
	@if [ -z "$(CONFIG)" ]; then \
		echo "❌ Please specify CONFIG file: make deploy-config CONFIG=myconfig.yaml"; \
		exit 1; \
	fi
	@echo "🚀 Deploying with config: $(CONFIG)"
	go run main.go $(CONFIG)

# Clean build artifacts
clean:
	go clean
	@echo "✅ Clean complete"

# Development helpers
fmt:
	go fmt ./...

vet:
	go vet ./...

# Quick local testnet example
local-testnets:
	@echo "Starting local testnets with Anvil..."
	@echo "Terminal 1: anvil --port 8545 --chain-id 31337"
	@echo "Terminal 2: anvil --port 8546 --chain-id 31338" 
	@echo "Terminal 3: anvil --port 8547 --chain-id 31339"
	@echo "Then run: make deploy" 