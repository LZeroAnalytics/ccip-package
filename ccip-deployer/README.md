# CCIP Deployer

A simple Go tool for deploying Chainlink CCIP (Cross-Chain Interoperability Protocol) to custom testnets and local chains.

## 🚀 Quick Start

### 1. Setup
```bash
cd ccip-deployer
go mod tidy
```

### 2. Configure Your Deployment
Edit `config.yaml` with your testnet details:

```yaml
home_chain:
  chain_id: 31337
  name: "my-testnet-1"
  rpc_url: "http://localhost:8545"
  private_key: "0x..."

feed_chain:
  chain_id: 31338
  name: "my-testnet-2"  
  rpc_url: "http://localhost:8546"
  private_key: "0x..."

chains:
  - chain_id: 31339
    name: "my-testnet-3"
    rpc_url: "http://localhost:8547"
    private_key: "0x..."

deployment:
  rmn_enabled: true
  fresh_deployment: true
```

### 3. Deploy CCIP

**Option A: Full Automatic Deployment**
```bash
go run main.go config.yaml
```

**Option B: Step-by-Step Interactive Deployment**
```bash
go run deploy_step_by_step.go interactive config.yaml
```

**Option C: Step-by-Step Automatic Deployment**
```bash
go run deploy_step_by_step.go auto config.yaml
```

## 📋 What Gets Deployed

### Home Chain Contracts
- **CapabilityRegistry**: Central registry for DONs and capabilities
- **MCMS**: Multi-chain multi-sig for governance
- **DON configurations**: For OCR and CCIP workflows

### CCIP Contracts (on all chains)
- **Router**: Main CCIP entry point for sending messages
- **OnRamp/OffRamp**: Lane-specific contracts for message processing  
- **TokenPool**: For cross-chain token transfers
- **PriceRegistry**: For gas price and token price feeds
- **CommitStore**: For merkle root commitments
- **ExecutionReportRouter**: For execution reports

### Supporting Infrastructure
- **LINK Token**: Native payment token (deployed fresh)
- **Price Feeds**: ETH/USD and LINK/ETH feeds
- **Fee Quoter**: For calculating cross-chain fees
- **NonceManager**: For replay protection

## 🔧 Configuration Options

### Chain Configuration
```yaml
home_chain:
  chain_id: 31337           # Your testnet chain ID
  name: "local-testnet-1"   # Human-readable name
  rpc_url: "http://..."     # RPC endpoint
  private_key: "0x..."      # Deployer private key
```

### Deployment Options
```yaml
deployment:
  rmn_enabled: true         # Enable Risk Management Network
  fresh_deployment: true    # Deploy all contracts fresh
```

### For Forked Networks (Optional)
```yaml
existing_contracts:
  link_token: "0x514910771AF9Ca656af840dff83E8264EcF986CA"
  link_eth_feed: "0x2c1d072e956AFFC0D435Cb7AC38EF18d24d9127c"
  # Add other existing contract addresses
```

## 🌐 Chain Selectors

The tool automatically calculates chain selectors using:
```
Chain Selector = Chain ID + 0x1000000000000000
```

For production, you should use the official [chain-selectors](https://github.com/smartcontractkit/chain-selectors) library.

## 📊 Usage Examples

### Local Anvil Testnets
```bash
# Terminal 1: Start first testnet
anvil --port 8545 --chain-id 31337

# Terminal 2: Start second testnet  
anvil --port 8546 --chain-id 31338

# Terminal 3: Deploy CCIP
go run main.go config.yaml
```

### Custom Testnet Deployment
```yaml
home_chain:
  chain_id: 421614  # Arbitrum Sepolia
  rpc_url: "https://sepolia-rollup.arbitrum.io/rpc"
  private_key: "0x..."

feed_chain:
  chain_id: 11155111  # Ethereum Sepolia
  rpc_url: "https://eth-sepolia.public.blastapi.io"
  private_key: "0x..."
```

## 🛠️ Advanced Usage

### Step-by-Step Deployment
For debugging or partial deployments:

```bash
# Interactive mode - pause between each step
go run deploy_step_by_step.go interactive config.yaml

# Automatic mode - run all steps without pausing
go run deploy_step_by_step.go auto config.yaml
```

### Deployment Steps
1. **Deploy Home Chain**: CapabilityRegistry, MCMS, DONs
2. **Deploy CCIP Chains**: CCIP contracts on all chains
3. **Connect Lanes**: Establish cross-chain connections
4. **Configure OCR**: Set up oracle networks
5. **Fund Transmitters**: Fund oracle nodes for gas

## 🔍 Troubleshooting

### Common Issues

**"Failed to deploy home chain contracts"**
- Check RPC URL is accessible
- Verify private key has sufficient funds
- Ensure chain ID matches config

**"CCIP chains deployment failed"**
- Verify all chains are running and accessible
- Check that home chain deployment completed successfully
- Ensure consistent private keys across chains

**"AddressBook deprecated warnings"**
- These are expected during the transition to DataStore
- Functionality remains intact despite warnings

### Debug Mode
Add verbose logging by modifying the logger configuration in your deployment script.

## 📚 Next Steps

After successful deployment:

1. **Test Cross-Chain Messaging**: Send test messages between chains
2. **Monitor Lanes**: Check OCR reports and message execution
3. **Add More Chains**: Extend to additional networks
4. **Configure Token Pools**: Set up cross-chain token transfers

## 🔗 Related Documentation

- [Chainlink CCIP Documentation](https://docs.chain.link/ccip)
- [CCIP Architecture Overview](https://docs.chain.link/ccip/architecture)
- [Chain Selectors Registry](https://github.com/smartcontractkit/chain-selectors)

---

**Note**: This tool is designed for development and testing. For production deployments, follow Chainlink's official deployment procedures and security guidelines. 