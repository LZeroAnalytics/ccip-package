#!/bin/bash
set -e

echo "🧪 Testing CCIP Deployer with preexisting contracts"

cat > /tmp/test_deployment.json << EOF
{
  "chains": {
    "chain_9250445": {
      "chain_selector": 4949039107694359620,
      "rpc_endpoint": "https://ec82cfa994764d8285fa0d42ba974cb4-rpc.network.bloctopus.io",
      "chain_id": 9250445,
      "name": "chain_9250445"
    },
    "chain_9388201": {
      "chain_selector": 3478487238524512106,
      "rpc_endpoint": "https://d25028740f3a45359c410a2303a34d34-rpc.network.bloctopus.io",
      "chain_id": 9388201,
      "name": "chain_9388201"
    }
  },
  "private_key": "3a23daa1250597152769c50729081a957271a32fee151e478356d1f75867a527",
  "home_chain": "chain_9250445",
  "num_nodes": 4,
  "num_bootstraps": 1,
  "enable_mercury": false,
  "enable_log_triggers": false,
  "preexisting_contracts": {
    "link_token_chain_9250445": {
      "address": "0x514910771AF9Ca656af840dff83E8264EcF986CA",
      "chain": "chain_9250445",
      "type": "LinkToken"
    },
    "link_token_chain_9388201": {
      "address": "0x514910771AF9Ca656af840dff83E8264EcF986CA",
      "chain": "chain_9388201",
      "type": "LinkToken"
    },
    "price_feed_chain_9250445": {
      "address": "0xdc530d9457755926550b59e8eccdae7624181557",
      "chain": "chain_9250445",
      "type": "PriceFeed"
    },
    "price_feed_chain_9388201": {
      "address": "0xdc530d9457755926550b59e8eccdae7624181557",
      "chain": "chain_9388201",
      "type": "PriceFeed"
    }
  }
}
EOF

cd src
echo "🔨 Building simplified deployer..."
go build -o /tmp/deployer cmd/deployer/simplified_main.go

echo "🚀 Running deployer with preexisting contracts..."
CONFIG_PATH=/tmp/test_deployment.json /tmp/deployer

echo "✅ Test completed"
