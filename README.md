# 🌉 CCIP Package for Kurtosis 🚀

> *Deploy complete Cross-Chain Interoperability Protocol (CCIP) infrastructure in isolated test environments with one command.*

## 🌐 Overview

This package allows you to deploy and configure a full **CCIP v1.6 infrastructure** in isolated, reproducible environments using [Kurtosis](https://kurtosis.com/). CCIP enables secure cross-chain communication, allowing smart contracts on one blockchain to send messages, transfer tokens, and execute functions on another blockchain.

## 🚀 Key Capabilities

- 🛠️ **Complete CCIP Infrastructure Deployment**:
  - Deploy 6 Chainlink nodes (1 bootstrap + 5 plugin nodes) with DON configuration
  - Set up Job Distributor for centralized job management
  - Deploy all required CCIP v1.6 contracts on multiple chains
  - Configure cross-chain lanes for bidirectional communication

- ⚡ **Cross-Chain Services**:
  - 🌉 **Token Transfer**: Secure cross-chain token movements with burn/mint or lock/release mechanisms
  - 📨 **Arbitrary Messaging**: Send any data between chains with programmable receivers
  - 🔒 **Risk Management Network (RMN)**: Optional fraud detection and risk mitigation (configuration only)
  - 💰 **Fee Management**: Dynamic pricing based on gas costs and network conditions

- 🏗️ **Multi-Chain Architecture**:
  - **Home Chain**: Primary chain hosting CapabilitiesRegistry and CCIPHome contracts
  - **Feed Chain**: Chain providing LINK/ETH price feeds for fee calculations
  - **Remote Chains**: Additional chains participating in CCIP network
  - **Lane Configuration**: Automatic setup of bidirectional communication lanes

- 🔧 **Flexible Deployment Options**:
  - Use existing contracts or deploy fresh infrastructure
  - Support for custom private keys and RPC endpoints
  - RMN integration ready (contracts deployed, nodes require separate setup)
  - Configurable for testnet, mainnet, or devnet environments

## ✨ Features

- **CCIP v1.6 Support**: Latest CCIP protocol with enhanced security and efficiency
- **Multi-Chain DON**: Single Decentralized Oracle Network serving multiple chains
- **Automated Lane Setup**: Bidirectional lanes between all configured chains
- **Job Distributor Integration**: Centralized job management and distribution
- **Contract Flexibility**: Deploy new contracts or use existing ones
- **Bootstrap & Plugin Nodes**: Proper OCR consensus with 3f+1 fault tolerance
- **Price Feed Integration**: LINK/ETH feeds for accurate fee calculations

## 📋 Prerequisites

| Requirement | Version |
|-------------|---------|
| [Kurtosis](https://docs.kurtosis.com/)    | >= 0.47.0 | 
| Docker/Kubernetes      | >= 20.10.0 |
| Chainlink Nodes | >= 2.23.0 |
| Disk Space  | >= 10GB |

## 🏃 Quick Start

<div style="display: flex; align-items: flex-start;">
<div style="flex: 1;">

1️⃣ Create a `config.yaml` file:

```yaml
home_chain:
  chain_id: 1337
  name: "ethereum-1"
  rpc_url: "http://host.docker.internal:8545"
  private_key: "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

feed_chain:
  chain_id: 1338
  name: "ethereum-2"
  rpc_url: "http://host.docker.internal:8546"
  private_key: "0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"

chains:
  - chain_id: 1339
    name: "polygon"
    rpc_url: "http://host.docker.internal:8547"
    private_key: "0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a"

deployment:
  rmn_enabled: false

existing_contracts:
  link_token: ""
  link_eth_feed: ""
  eth_usd_feed: ""
```

2️⃣ Run the package:

```bash
kurtosis run github.com/your-org/ccip-package --args-file config.yaml
```

3️⃣ Access your infrastructure:

| Component | Endpoint |
|-----------|----------|
| Chainlink Nodes | http://localhost:6688-6693 |
| Job Distributor | gRPC: localhost:50051 |

</div>
</div>

## ⚙️ Configuration Reference

### 🏠 Home Chain Configuration (required)

The home chain hosts the core CCIP contracts including CapabilitiesRegistry and CCIPHome.

| Parameter | Required | Description |
|-----------|----------|-------------|
| `chain_id` | ✅ Yes | Chain ID as integer |
| `name` | ✅ Yes | Human-readable chain name |
| `rpc_url` | ✅ Yes | HTTP RPC endpoint URL |
| `private_key` | ✅ Yes | Private key for deployments (hex format) |

### 📊 Feed Chain Configuration (required)

The feed chain provides LINK/ETH price feeds for fee calculations.

| Parameter | Required | Description |
|-----------|----------|-------------|
| `chain_id` | ✅ Yes | Chain ID as integer |
| `name` | ✅ Yes | Human-readable chain name |
| `rpc_url` | ✅ Yes | HTTP RPC endpoint URL |
| `private_key` | ✅ Yes | Private key for deployments |

### 🌐 Additional Chains (optional)

Additional chains that will participate in the CCIP network.

| Parameter | Required | Description |
|-----------|----------|-------------|
| `chain_id` | ✅ Yes | Chain ID as integer |
| `name` | ✅ Yes | Human-readable chain name |
| `rpc_url` | ✅ Yes | HTTP RPC endpoint URL |
| `private_key` | ✅ Yes | Private key for deployments |

### 🚀 Deployment Options

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `rmn_enabled` | ❌ No | `false` | Enable RMN verification (contracts only) |

### 📝 Existing Contracts (optional)

Specify existing contracts to avoid redeployment:

| Parameter | Required | Description |
|-----------|----------|-------------|
| `link_token` | ❌ No | Existing LINK token contract address |
| `link_eth_feed` | ❌ No | Existing LINK/ETH price feed address |
| `eth_usd_feed` | ❌ No | Existing ETH/USD price feed address |

### 🔗 Chainlink Node Configuration (auto-generated)

If fewer than 6 nodes are provided, the package automatically generates:

| Node Type | Count | Role |
|-----------|-------|------|
| Bootstrap | 1 | Network coordination and P2P discovery |
| Plugin | 5 | OCR consensus participation and job execution |

## 🏗️ Architecture Deep Dive

### 🌟 CCIP v1.6 Overview

CCIP (Cross-Chain Interoperability Protocol) v1.6 is Chainlink's latest protocol for secure cross-chain communication. This package deploys the complete infrastructure required for a production-ready CCIP network.

### 📐 Network Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Home Chain    │    │   Feed Chain    │    │  Remote Chain   │
│                 │    │                 │    │                 │
│ CapabilitiesReg │    │   PriceFeed     │    │   OnRamp        │
│ CCIPHome        │    │   LINK/ETH      │    │   OffRamp       │
│ RMNHome         │    │   ETH/USD       │    │   Routers       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │      DON        │
                    │                 │
                    │  1 Bootstrap    │
                    │  5 Plugin Nodes │
                    │                 │
                    │ Job Distributor │
                    └─────────────────┘
```

### 🏢 Core Contracts Deployed

#### Home Chain Contracts
| Contract | Purpose |
|----------|---------|
| **CapabilitiesRegistry** | Central registry for all DON capabilities and configurations |
| **CCIPHome** | Core CCIP configuration and state management |
| **RMNHome** | Risk Management Network configuration hub |

#### Per-Chain Contracts
| Contract | Purpose |
|----------|---------|
| **OnRamp** | Handles outgoing cross-chain messages and token transfers |
| **OffRamp** | Processes incoming messages and executes cross-chain calls |
| **EVM2EVMOnRamp** | Legacy v1.2 compatibility layer |
| **EVM2EVMOffRamp** | Legacy v1.2 compatibility layer |
| **Router** | User-facing interface for CCIP interactions |
| **TokenPool** | Manages token locking/unlocking or burning/minting |
| **PriceRegistry** | Chain-specific gas price and token price management |

#### Supporting Contracts
| Contract | Purpose |
|----------|---------|
| **LINK Token** | Native token for node payments and staking |
| **BurnMintERC677** | Standard CCIP-compatible token implementation |
| **LinkToken** | LINK implementation with CCIP compatibility |
| **WETH9** | Wrapped ETH for standardized token handling |

### 🔗 DON (Decentralized Oracle Network) Configuration

#### Node Roles & Responsibilities

**Bootstrap Node (1x)**
- **P2P Network Coordination**: Acts as initial connection point for plugin nodes
- **Network Discovery**: Helps plugin nodes find and connect to each other
- **No Job Execution**: Pure networking role, doesn't participate in OCR rounds

**Plugin Nodes (5x)**
- **OCR Consensus**: Participate in Off-Chain Reporting consensus mechanisms
- **Job Execution**: Execute CCIP-specific jobs for cross-chain operations
- **3f+1 Fault Tolerance**: 5 nodes provide tolerance for up to 1-2 Byzantine faults
- **Commit & Execute**: Handle both commit and execution phases of CCIP

#### 🎯 Jobs Deployed Per Node

**CCIP Commit Jobs**
```toml
name = "ccip-commit-{source_chain_id}-{dest_chain_id}"
type = "ccip-commit"
# Monitors source chain for new messages
# Creates merkle roots of message batches
# Submits commit reports to destination chain
```

**CCIP Execute Jobs**
```toml
name = "ccip-execute-{source_chain_id}-{dest_chain_id}"
type = "ccip-execute" 
# Monitors destination chain for committed messages
# Executes cross-chain calls and token transfers
# Updates message status and handles gas estimation
```

**Bootstrap Job** (Bootstrap node only)
```toml
name = "ccip-bootstrap"
type = "bootstrap"
# Provides P2P networking services
# No consensus participation
```

### 🛣️ Lane Configuration

CCIP operates on **lanes** - unidirectional paths between chains. This package automatically configures **bidirectional lanes** between all chains.

#### Lane Example: Chain A ↔ Chain B
```
Chain A → Chain B Lane:
├── OnRamp (Chain A) → OffRamp (Chain B)
├── Commit Job: Monitors Chain A, reports to Chain B
└── Execute Job: Monitors Chain B, executes messages

Chain B → Chain A Lane:
├── OnRamp (Chain B) → OffRamp (Chain A)  
├── Commit Job: Monitors Chain B, reports to Chain A
└── Execute Job: Monitors Chain A, executes messages
```

#### Lane Matrix for 3 Chains
| From/To | Chain A | Chain B | Chain C |
|---------|---------|---------|---------|
| Chain A | - | Lane A→B | Lane A→C |
| Chain B | Lane B→A | - | Lane B→C |
| Chain C | Lane C→A | Lane C→B | - |

**Total Lanes**: n(n-1) = 3×2 = 6 lanes for 3 chains

### 🎛️ Job Distributor

The Job Distributor is a centralized service that manages job deployment across all nodes:

**Features:**
- **gRPC API**: Listens on port 50051 for job management requests
- **Node Registration**: Nodes register themselves on startup
- **Job Propagation**: Distributes jobs to appropriate nodes based on capabilities
- **CSA Authentication**: Uses Chainlink Service Agreement keys for security

**Workflow:**
1. **Node Startup**: Nodes register with Job Distributor
2. **DON Configuration**: CCIP deployer configures DON with node details
3. **Job Creation**: Jobs are created for each lane and capability
4. **Job Distribution**: Job Distributor pushes jobs to appropriate nodes
5. **Execution**: Nodes execute jobs and report back status

### 🔒 Risk Management Network (RMN)

RMN provides an additional security layer for CCIP by detecting and preventing fraudulent cross-chain messages.

**Deployment Status:**
- ✅ **Contracts Deployed**: RMNHome and configuration contracts
- ✅ **DON Integration**: OCR configured to expect RMN verification when enabled
- ❌ **RMN Nodes**: NOT automatically deployed (requires separate infrastructure)

**RMN Components (when fully deployed):**
- **RMN Nodes**: Independent verification nodes (rageproxy + afn2proxy containers)
- **Fraud Detection**: Monitors for invalid state transitions
- **Emergency Stop**: Can halt lanes if fraud is detected
- **Decentralized Verification**: Multiple independent RMN operators

### 💰 Fee Mechanism

CCIP uses a sophisticated fee model to ensure economic sustainability:

**Fee Components:**
1. **Execution Fee**: Cost of executing the message on destination chain
2. **Data Availability Fee**: Cost of storing message data
3. **Network Fee**: CCIP protocol usage fee
4. **Premium**: Market-based pricing adjustments

**Price Feeds Required:**
- **LINK/ETH**: For converting LINK payments to gas costs
- **ETH/USD**: For USD-denominated pricing
- **Gas Price Feeds**: Real-time gas price monitoring per chain

### 🛠️ Package Components

#### 1. CCIP Deployer (`ccip-deployer/`)
**Technology**: Go application with Starlark wrapper
**Purpose**: Deploys and configures all CCIP contracts and infrastructure

**Key Files:**
- `main.go`: Core deployment orchestration
- `ccip_deployer.go`: Contract deployment logic
- `don.go`: DON configuration and node management
- `main.star`: Starlark wrapper for Kurtosis integration

#### 2. Chainlink Node Package
**Technology**: Imported Kurtosis package
**Purpose**: Deploys the 6-node DON with proper configuration

**Features:**
- PostgreSQL databases per node
- Multi-chain RPC configuration
- P2P networking setup
- Job template pre-loading

#### 3. Job Distributor (`job-distributor/`)
**Technology**: Go gRPC service
**Purpose**: Centralized job management and distribution

**API Endpoints:**
- `RegisterNode`: Node registration
- `ProposeJob`: Job creation and validation
- `UpdateJob`: Job modification
- `DeleteJob`: Job removal

## 🔄 Deployment Flow

### Phase 1: Infrastructure Setup
1. **PostgreSQL Deployment**: One database per Chainlink node
2. **Chainlink Node Deployment**: 6 nodes with multi-chain configuration
3. **Job Distributor Startup**: gRPC service initialization

### Phase 2: Contract Deployment
1. **Token Contracts**: LINK, WETH, and test tokens
2. **Price Feeds**: LINK/ETH and ETH/USD oracles
3. **Core CCIP Contracts**: CapabilitiesRegistry, CCIPHome, RMNHome
4. **Per-Chain Contracts**: OnRamp, OffRamp, Router, TokenPools

### Phase 3: DON Configuration
1. **Node Registration**: Nodes register with Job Distributor
2. **DON Setup**: Configure DON in CapabilitiesRegistry
3. **OCR Configuration**: Set up consensus parameters
4. **Capability Registration**: Register CCIP capabilities

### Phase 4: Lane Configuration
1. **Lane Creation**: Set up OnRamp → OffRamp mappings
2. **Fee Configuration**: Configure pricing parameters
3. **Token Support**: Enable supported tokens per lane
4. **Rate Limiting**: Set transfer limits and windows

### Phase 5: Job Deployment
1. **Job Generation**: Create commit/execute jobs for each lane
2. **Job Distribution**: Push jobs to appropriate nodes
3. **Job Activation**: Start job execution
4. **Health Monitoring**: Verify job status and node health

## 📊 Monitoring & Debugging

### Node Access
Each Chainlink node exposes a web interface:
- **Node 0 (Bootstrap)**: http://localhost:6688
- **Node 1-5 (Plugin)**: http://localhost:6689-6693

### Job Distributor Monitoring
- **gRPC Health Check**: `grpcurl -plaintext localhost:50051 list`
- **Container Logs**: `kurtosis service logs job-distributor`

### Contract Verification
The deployer outputs all deployed contract addresses for verification on block explorers.

## 🎯 Use Cases

### Development & Testing
- **dApp Development**: Test cross-chain functionality locally
- **Integration Testing**: Validate end-to-end cross-chain workflows
- **Load Testing**: Stress-test CCIP infrastructure

### Production Setup
- **Testnet Deployment**: Deploy on Sepolia, Goerli, Mumbai
- **Mainnet Preparation**: Production-ready configuration templates
- **Custom Networks**: Support for any EVM-compatible chains

### Research & Innovation
- **Protocol Research**: Experiment with CCIP parameters
- **Security Analysis**: Test attack vectors and mitigations
- **Performance Optimization**: Benchmark throughput and latency

## 🔧 Advanced Configuration

### Custom Token Pools
```yaml
token_pools:
  USDC:
    type: "burn_mint"
    decimals: 6
    burn_enabled: true
  WETH:
    type: "lock_release"  
    pool_type: "native"
```

### Lane-Specific Settings
```yaml
lanes:
  "1337->1338":
    rate_limiter_enabled: true
    rate_limiter_capacity: "1000000000000000000000"
    rate_limiter_rate: "1000000000000000000"
    fee_token_config:
      network_fee_usd_cents: 50
```

### RMN Configuration
```yaml
rmn:
  enabled: true
  curse_subjects:
    - "0x1234..." # OnRamp addresses that can be cursed
  blessing_subjects:
    - "0x5678..." # OffRamp addresses that can be blessed
```

## 🚨 Security Considerations

### Private Key Management
- **Development Only**: The package uses hardcoded private keys for testing
- **Production**: Use secure key management systems
- **Rotation**: Implement key rotation for long-term deployments

### Network Security
- **Firewall Rules**: Restrict access to node interfaces
- **TLS/SSL**: Enable encryption for production deployments
- **Authentication**: Use strong passwords and API keys

### Smart Contract Security
- **Audited Contracts**: CCIP v1.6 contracts are audited by multiple firms
- **Upgrade Procedures**: Follow proper governance for contract upgrades
- **Emergency Procedures**: Understand pause and emergency stop mechanisms

## 📚 Additional Resources

- [CCIP Documentation](https://docs.chain.link/ccip)
- [Chainlink Node Documentation](https://docs.chain.link/chainlink-nodes)
- [Kurtosis Documentation](https://docs.kurtosis.com/)
- [CCIP GitHub Repository](https://github.com/smartcontractkit/chainlink-ccip)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Test your changes with local deployments
4. Submit a pull request with detailed description

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

⭐ **Star this repo if you find it useful!** ⭐

Built with ❤️ for the cross-chain future 🌉
