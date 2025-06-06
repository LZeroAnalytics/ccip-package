# Job Distributor Service

A standalone Job Distributor service for CCIP deployment with **dynamic node registration** and auto-approval functionality.

## 🎯 **Architecture Overview**

The Job Distributor follows the **"Empty Start, Dynamic Registration"** pattern:

1. **🏁 JD starts empty** - No pre-configured nodes
2. **🚀 Nodes deploy separately** - Via Docker, Kubernetes, etc.
3. **📝 Nodes register themselves** - With the JD using CSA keys
4. **✅ Auto-approval** - Jobs are automatically approved for CCIP deployment

## 🚀 **Integration with CCIP Deployer**

### **Step 1: Deploy Infrastructure First**
```bash
# Deploy your Chainlink nodes first (Docker example)
docker-compose up chainlink-nodes

# Start the Job Distributor service
docker-compose up job-distributor
```

### **Step 2: Create NodeInfo Configuration**
Create `nodes.yaml` with your actual node details:
```yaml
nodes:
  - name: "chainlink-bootstrap-1"
    chainlink_url: "http://chainlink-bootstrap:6688"
    chainlink_email: "admin@bootstrap.com"  
    chainlink_password: "bootstrap_password"
    p2p_port: "6690"
    is_bootstrap: true
    admin_addr: "0x1000000000000000000000000000000000000000"
    multi_addr: "127.0.0.1:6690"
    labels:
      role: "bootstrap"
      network: "testnet"

  - name: "chainlink-node-1"
    chainlink_url: "http://chainlink-node-1:6688"
    chainlink_email: "admin@node1.com"
    chainlink_password: "node1_password"
    p2p_port: "6691"
    is_bootstrap: false
    admin_addr: "0x2000000000000000000000000000000000000000"
    multi_addr: "127.0.0.1:6691"
    labels:
      role: "ccip"
      network: "testnet"
```

### **Step 3: Run CCIP Deployer**
```bash
# Set the node configuration file
export NODE_CONFIG_FILE=nodes.yaml

# Run the CCIP deployer
cd ccip-deployer
go run main.go
```

## 🔧 **How It Works**

```mermaid
flowchart TD
    A[🏁 JD Service Starts Empty] --> B[🚀 CCIP Deployer Loads nodes.yaml]
    B --> C[📞 devenv.NewEnvironment(NodeInfos)]
    C --> D[🏗️ NewRegisteredDON()]
    D --> E[🔄 For Each Node]
    E --> F[📝 RegisterNodeToJobDistributor()]
    F --> G[🔌 CreateJobDistributor()]
    G --> H[✅ Bidirectional Connection]
    H --> I[🎯 Ready for CCIP Jobs]
```

### **Key Points:**
- ✅ **JD doesn't pre-configure nodes** - Starts completely empty
- ✅ **NodeInfos passed to devenv** - Not to JD directly  
- ✅ **Nodes register themselves** - Via `RegisterNodeToJobDistributor()`
- ✅ **Dynamic client creation** - JD creates clients as nodes register
- ✅ **Auto-approval** - All jobs automatically approved

## 🐳 **Docker Deployment**

The JD service is now simplified without pre-configured nodes:

```yaml
# docker-compose.yml
services:
  job-distributor:
    build: .
    ports:
      - "50051:50051"
    environment:
      - JD_PORT=50051
      - JD_LOG_LEVEL=info
    # ✅ No node configuration needed!
```

## 🔧 **Configuration**

### **Environment Variables**
- `JD_PORT` - gRPC server port (default: 50051)
- `JD_LOG_LEVEL` - Log level (default: info)
- `JD_ENABLE_METRICS` - Enable metrics (default: false)
- `JD_METRICS_PORT` - Metrics port (default: 8080)

### **No Node Configuration Required**
The JD service no longer requires pre-configured nodes. All node information is provided via the CCIP deployer's `nodes.yaml` file.

## 📊 **Service APIs**

### **Job Service**
- `ProposeJob()` - Auto-approves and submits jobs to Chainlink nodes
- `GetJob()` - Retrieves job details
- `ListJobs()` - Lists all jobs with filtering
- `DeleteJob()` - Removes jobs from nodes
- `UpdateJob()` - Updates existing jobs

### **Node Service**  
- `RegisterNode()` - Registers new nodes dynamically
- `ListNodes()` - Lists registered nodes
- `GetNode()` - Gets node details

### **CSA Service**
- `ListCSAKeys()` - Lists CSA keys from registered nodes
- `GetCSAKey()` - Gets specific CSA key

## 🎯 **Usage Pattern**

1. **Deploy your Chainlink nodes** (Docker/K8s/Manual)
2. **Start the JD service** (empty, no config needed)
3. **Create `nodes.yaml`** with your node URLs and credentials  
4. **Run CCIP deployer** with `NODE_CONFIG_FILE=nodes.yaml`
5. **Nodes auto-register** with JD during deployment
6. **Jobs auto-approved** and submitted to nodes
7. **CCIP deployment completes** successfully

This architecture ensures clean separation of concerns and follows the established CCIP deployment patterns.

## 🚀 Quick Start

### 1. Using Docker Compose (Recommended)

```bash
cd job-distributor
docker-compose up -d
```

The service will be available at `localhost:8080`.

### 2. Using Docker

```bash
cd job-distributor
docker build -t job-distributor .
docker run -p 8080:8080 -e JD_NODE_COUNT=4 job-distributor
```

### 3. Local Development

```bash
cd job-distributor
go mod tidy
go run main.go
```

## 📝 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `JD_PORT` | `8080` | gRPC server port |
| `JD_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `JD_NODE_COUNT` | `4` | Number of mock Chainlink nodes to simulate |
| `JD_CONFIG_FILE` | `""` | Path to JSON config file (optional) |

### Config File Format

You can provide a JSON config file to customize node configurations:

```json
{
  "port": "8080",
  "log_level": "info",
  "nodes": [
    {
      "id": "1",
      "name": "chainlink-node-1",
      "csa_public_key": "your_csa_key_here",
      "workflow_key": "your_workflow_key_here", 
      "p2p_peer_id": "your_p2p_peer_id_here",
      "is_bootstrap": true,
      "admin_addr": "0x1000000000000000000000000000000000000000",
      "multi_addr": "127.0.0.1:6680",
      "chainlink_api_url": "http://chainlink-node-1:6688",
      "chainlink_api_key": "admin@example.com",
      "chainlink_password": "password",
      "labels": {
        "role": "ccip",
        "network": "testnet",
        "type": "bootstrap"
      }
    }
  ],
  "csa_keys": [
    {
      "node_id": "1",
      "public_key": "your_csa_key_here"
    }
  ]
}
```

## 🔧 Integration with CCIP Deployer

### 1. Update your CCIP deployer environment

Set these environment variables before running the CCIP deployer:

```bash
export JD_GRPC_URL="localhost:8080"
export JD_WSRPC_URL="ws://localhost:8080"
```

### 2. Run the Job Distributor first

```bash
cd job-distributor
docker-compose up -d
```

### 3. Then run your CCIP deployer

```bash
cd ../ccip-deployer
go run main.go config.yaml
```

## 🔌 Supported gRPC Services

The Job Distributor implements these Chainlink protobuf services:

### JobService
- `ProposeJob` - **Main method called by CCIP deployer**
- `BatchProposeJob` - Batch job proposals
- `ListJobs` - List all jobs
- `GetJob` - Get specific job
- `DeleteJob` - Delete a job
- `UpdateJob` - Update a job
- `ListProposals` - List job proposals  
- `GetProposal` - Get specific proposal
- `RevokeJob` - Revoke a job proposal

### NodeService  
- `ListNodes` - List all nodes
- `GetNode` - Get specific node
- `RegisterNode` - Register new node
- `EnableNode` - Enable a node
- `DisableNode` - Disable a node
- `UpdateNode` - Update node config

### CSAService
- `ListKeypairs` - List CSA keypairs
- `GetKeypair` - Get specific keypair

## 🔄 Auto-Approval Flow

The Job Distributor implements **auto-approval** for job proposals:

1. **CCIP Deployer** calls `ProposeJob(nodeID, spec)`
2. **Job Distributor** receives the request
3. **Auto-approve**: Status is immediately set to `APPROVED`
4. **Job Created**: Job record is stored with proposal
5. **Response**: Success response returned to deployer

This eliminates the manual approval step that would normally be required.

## 🛠 Extension Points

### Adding Real Chainlink Integration

To connect to actual Chainlink nodes, modify:

1. **`internal/chainlink/client.go`** - Implement real HTTP API calls
2. **`internal/server/server.go`** - Update `submitJobToNode()` and `deleteJobFromNode()`
3. **Configuration** - Add real node API endpoints and credentials

### Example Real Implementation:

```go
func (c *Client) CreateJob(ctx context.Context, spec string) error {
    // Real implementation example
    client := &http.Client{}
    
    req, _ := http.NewRequest("POST", c.Config.ChainlinkAPIURL+"/v2/jobs", 
        strings.NewReader(spec))
    req.Header.Set("Content-Type", "application/toml")
    req.Header.Set("X-Chainlink-User", c.Config.ChainlinkAPIKey) 
    req.Header.Set("X-Chainlink-Password", c.Config.ChainlinkPassword)
    
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return fmt.Errorf("failed to create job: %d", resp.StatusCode)
    }
    
    return nil
}
```

## 🧪 Testing

You can test the gRPC endpoints using `grpcurl`:

```bash
# List nodes
grpcurl -plaintext localhost:8080 node.v1.NodeService/ListNodes

# Propose a job
grpcurl -plaintext -d '{"node_id":"1","spec":"type=\"webhook\"\nschemaVersion=1\n"}' \
  localhost:8080 job.v1.JobService/ProposeJob

# List jobs
grpcurl -plaintext localhost:8080 job.v1.JobService/ListJobs
```

## 📚 Architecture

```
CCIP Deployer                    Job Distributor                Chainlink Nodes
┌─────────────┐                 ┌──────────────────┐            ┌─────────────┐
│             │  gRPC calls     │                  │ HTTP API   │             │
│ ProposeJob()├────────────────►│  Auto-approve    ├───────────►│ Node 1      │
│ ListJobs()  │                 │  Job Storage     │            │             │
│ ListNodes() │                 │  Node Registry   │            ├─────────────┤
│             │                 │                  │            │ Node 2      │
└─────────────┘                 └──────────────────┘            │             │
                                                                 ├─────────────┤
                                                                 │ Node 3      │
                                                                 │             │
                                                                 ├─────────────┤
                                                                 │ Node 4      │
                                                                 │             │
                                                                 └─────────────┘
```

## 🔍 Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   lsof -i :8080
   # Kill the process and restart
   ```

2. **gRPC connection refused**
   - Ensure Job Distributor is running
   - Check `JD_GRPC_URL` environment variable
   - Verify firewall/network settings

3. **CCIP deployer can't find JD**
   - Set correct environment variables:
     ```bash
     export JD_GRPC_URL="localhost:8080" 
     export JD_WSRPC_URL="ws://localhost:8080"
     ```

### Debug Mode

Run with debug logging:
```bash
docker run -p 8080:8080 -e JD_LOG_LEVEL=debug job-distributor
```

## 🎯 Next Steps

1. **Start the Job Distributor**: `docker-compose up -d`
2. **Configure CCIP Deployer**: Set JD environment variables
3. **Run CCIP Deployment**: Use your existing config.yaml
4. **Monitor Logs**: Check both services for successful integration
5. **Extend**: Add real Chainlink node integration as needed 