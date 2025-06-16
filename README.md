# CCIP (Cross-Chain Interoperability Protocol) Deployment Package

A complete Kurtosis package for deploying and configuring Chainlink's Cross-Chain Interoperability Protocol (CCIP) infrastructure across multiple blockchain networks.

## 🎯 Overview

This package automates the deployment of a complete CCIP infrastructure including:
- **Decentralized Oracle Network (DON)** with 6 Chainlink nodes
- **Home Chain Contracts** (CapabilitiesRegistry, CCIPHome, RMNHome)
- **Chain-specific CCIP Contracts** (Router, OnRamp, OffRamp, FeeQuoter)
- **Cross-chain Lanes** for bidirectional communication
- **OCR3 Configuration** for commit and execution plugins

## 📋 What This Package Returns

The package returns a comprehensive deployment result:

```starlark
{
  "contracts_addresses": {
    "home_chain_contracts": {
      "capReg": "0x...",          # CapabilitiesRegistry address
      "ccipHome": "0x...",        # CCIPHome contract address  
      "rmnHome": "0x..."          # RMNHome contract address
    },
    "chains_contracts": {
      "ethereum-1": {
        "router": "0x...",        # CCIP Router
        "onRamp": "0x...",        # OnRamp for outgoing messages
        "offRamp": "0x...",       # OffRamp for incoming messages
        "feeQuoter": "0x...",     # Fee calculation contract
        "rmnProxy": "0x...",      # Risk Management Network proxy
        "tokenAdminRegistry": "0x...",
        "registryModule": "0x...",
        "linkToken": "0x...",
        "nonceManager": "0x..."
      },
      "ethereum-2": {
        # ... same structure for each chain
      }
    }
  },
  "chainlink_nodes": [
    # Array of 6 Chainlink node configurations
  ]
}
```

## 🏗️ CCIP Architecture Role

### Core Components

1. **Home Chain Infrastructure**
   - **CapabilitiesRegistry**: Central registry for DON capabilities
   - **CCIPHome**: Core CCIP coordination contract
   - **RMNHome**: Risk Management Network configuration

2. **Chain-specific Components**
   - **Router**: Entry point for CCIP messages on each chain
   - **OnRamp**: Processes outgoing cross-chain messages
   - **OffRamp**: Processes incoming cross-chain messages  
   - **FeeQuoter**: Calculates cross-chain transaction fees

3. **Oracle Network**
   - **6 Chainlink Nodes**: Provide decentralized consensus
   - **OCR3 Plugins**: Commit and Exec plugins for message processing
   - **P2P Network**: Secure communication between nodes

### Message Flow
```
Chain A                    Oracle Network                 Chain B
  ↓                           ↓                            ↓
Router → OnRamp → [OCR3 Commit] → [OCR3 Exec] → OffRamp → Router
```

## 🚀 Quick Start

### 1. Basic Usage
```bash
kurtosis run github.com/your-org/ccip-package --args-file config.yaml
```

### 2. With Custom Configuration
```bash
kurtosis run github.com/your-org/ccip-package '{
  "chains": [
    {
      "name": "ethereum-1",
      "chain_id": 9215983,
      "chain_selector": "1234567890",
      "rpc_url": "https://your-rpc-url.com",
      "private_key": "0x...",
      "existing_contracts": {
        "link_token": "0x514910771AF9Ca656af840dff83E8264EcF986CA",
        "weth9": "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
      }
    }
  ]
}'
```

## ⚙️ Configuration

### Required Configuration Structure

```yaml
chains:
  - chain_id: 9215983                    # EVM Chain ID
    chain_selector: "1234567890"         # CCIP Chain Selector (unique identifier)
    name: "ethereum-1"                   # Human-readable chain name
    rpc_url: "https://..."               # JSON-RPC endpoint
    private_key: "0x..."                 # Deployer private key (32 bytes hex)
    faucet: "https://faucet.chain.link/ethereum/mainnet"  # Optional: for funding nodes
    existing_contracts:
      link_token: "0x514910771AF9Ca656af840dff83E8264EcF986CA"
      link_eth_feed: "0xdc530d9457755926550b59e8eccdae7624181557"
      eth_usd_feed: "0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419"
      weth9: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
```

### Configuration Parameters Explained

| Parameter | Type | Description | Required |
|-----------|------|-------------|----------|
| `chain_id` | number | EVM chain ID (e.g., 1 for Ethereum mainnet) | ✅ |
| `chain_selector` | string | Unique CCIP identifier for the chain | ✅ |
| `name` | string | Human-readable chain identifier | ✅ |
| `rpc_url` | string | HTTP/HTTPS JSON-RPC endpoint | ✅ |
| `private_key` | string | Deployer account private key (without 0x prefix) | ✅ |
| `faucet` | string | Faucet URL for funding nodes | ✅ |
| `existing_contracts.link_token` | string | LINK token contract address | ✅ |
| `existing_contracts.weth9` | string | Wrapped ETH contract address | ✅ |
| `existing_contracts.link_eth_feed` | string | LINK/ETH price feed | ❌ |
| `existing_contracts.eth_usd_feed` | string | ETH/USD price feed | ❌ |

## 🔧 Deployment Steps

The package executes the following deployment sequence:

### Phase 1: Infrastructure Setup
1. **Deploy Chainlink Nodes** (6 nodes with PostgreSQL databases)
2. **Setup Hardhat Environment** for contract deployment
3. **Configure Network Parameters** for each chain

### Phase 2: Home Chain Deployment
1. **Deploy CapabilitiesRegistry** - Central capability management
2. **Deploy CCIPHome** - Core CCIP coordination contract
3. **Deploy RMNHome** - Risk Management Network
4. **Configure Node Operators** - Register node operator
5. **Add Nodes** - Register all 6 DON nodes with capabilities

### Phase 3: Chain-specific Deployment
For each configured chain:
1. **Deploy Prerequisites**:
   - RMN Proxy (& RMNMock - instead of RMNRemote - to point to)
   - Token Admin Registry  
   - Registry Module
   - Router (entry point)

2. **Deploy Core CCIP Contracts**:
   - Nonce Manager
   - Fee Quoter
   - OnRamp (outgoing messages)
   - OffRamp (incoming messages)

### Phase 4: Cross-chain Configuration
1. **Configure CCIP Lanes** - Bidirectional communication paths
2. **Setup OCR3 Plugins**:
   - Commit Plugin (message commitment)
   - Exec Plugin (message execution)
3. **Create Chainlink Jobs** - Bootstrap and CCIP job specs

### Phase 5: Oracle Configuration
1. **Generate OCR3 Configs** for commit and exec plugins
2. **Configure Node Keys**:
   - Signer keys (Ed25519 for off-chain consensus)
   - Transmitter keys (Ethereum addresses for on-chain transactions)
   - Encryption keys (X25519 for secure communication)

## 🔐 Key Management

### Node Key Types
The package manages three distinct key types per node:

1. **Signer Key** (`ocr_key.off_chain_key`)
   - Ed25519 key for OCR consensus signing
   - Used in off-chain consensus protocol
   - NOT the Ethereum signing key

2. **Transmitter Key** (`eth_key`)
   - Ethereum address for submitting transactions
   - Pays gas fees for on-chain operations
   - Actual blockchain interaction key

3. **Encryption Key** (`ocr_key.config_key`)
   - X25519 key for encrypted node communication
   - Secures off-chain message passing

## 🌐 Network Requirements

### Minimum Requirements
- **2+ EVM-compatible chains** (for cross-chain functionality)
- **Funded deployer accounts** on all chains
- **RPC endpoints** with sufficient rate limits
- **LINK and WETH contracts** deployed on each chain

### Production Considerations
- Use **7-13 nodes** for production deployments (modify `DON_NODES_COUNT`)
- Ensure **geographic distribution** of nodes
- Implement **proper key management** (HSM/secure key storage)
- Monitor **gas prices** and **funding levels**

## 🛠️ Development & Testing

### Local Development
```bash
# Clone the repository
git clone <repository-url>
cd ccip-package

# Edit configuration
cp config.yaml my-config.yaml
# Edit my-config.yaml with your settings

# Deploy
kurtosis run . --args-file my-config.yaml
```

### Monitoring Deployment
```bash
# Check enclave status
kurtosis enclave inspect <enclave-name>

# View node logs
kurtosis service logs <enclave-name> chainlink-ccip-node-0

# Access node UI
# URLs will be displayed in the deployment output
```

## 🔍 Troubleshooting

### Common Issues

1. **Insufficient Funds**
   - Ensure deployer accounts have sufficient native tokens
   - Check gas price estimation

2. **RPC Rate Limits**
   - Use professional RPC providers
   - Implement proper rate limiting

3. **Contract Deployment Failures**
   - Verify existing contract addresses
   - Check network connectivity

4. **Node Connectivity Issues**
   - Verify P2P port accessibility
   - Check firewall configurations

### Debug Commands
```bash
# View all services
kurtosis enclave inspect <enclave-name>

# Check specific contract deployment
kurtosis service logs <enclave-name> hardhat-ccip-contracts

# Access node shell
kurtosis service shell <enclave-name> chainlink-ccip-node-0
```

## 📚 Dependencies

- **Kurtosis**: Orchestration platform
- **Chainlink Nodes**: Oracle network infrastructure  
- **Hardhat**: Ethereum development environment
- **PostgreSQL**: Node database storage
- **CCIP Contracts**: Cross-chain messaging protocol

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🔗 Additional Resources

- [Chainlink CCIP Documentation](https://docs.chain.link/ccip)
- [Kurtosis Documentation](https://docs.kurtosis.com/)
- [OCR Protocol Specification](https://research.chain.link/ocr.pdf)
- [CCIP Architecture Deep Dive](https://blog.chain.link/ccip-architecture/) 