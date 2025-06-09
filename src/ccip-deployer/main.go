package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"

	chain_selectors "github.com/smartcontractkit/chain-selectors"
	"github.com/smartcontractkit/chainlink/deployment"
	"github.com/smartcontractkit/chainlink/deployment/environment/crib"
	"github.com/smartcontractkit/chainlink/deployment/environment/devenv"
	"github.com/smartcontractkit/chainlink/v2/core/logger"

	clclient "github.com/smartcontractkit/chainlink/deployment/environment/nodeclient"
)

// Config represents the deployment configuration
type Config struct {
	HomeChain struct {
		ChainID    int64  `yaml:"chain_id"`
		Name       string `yaml:"name"`
		RPCURL     string `yaml:"rpc_url"`
		PrivateKey string `yaml:"private_key"`
	} `yaml:"home_chain"`

	FeedChain struct {
		ChainID    int64  `yaml:"chain_id"`
		Name       string `yaml:"name"`
		RPCURL     string `yaml:"rpc_url"`
		PrivateKey string `yaml:"private_key"`
	} `yaml:"feed_chain"`

	Chains []struct {
		ChainID    int64  `yaml:"chain_id"`
		Name       string `yaml:"name"`
		RPCURL     string `yaml:"rpc_url"`
		PrivateKey string `yaml:"private_key"`
	} `yaml:"chains"`

	Deployment struct {
		RMNEnabled bool `yaml:"rmn_enabled"`
	} `yaml:"deployment"`

	ExistingContracts map[string]string `yaml:"existing_contracts"`

	Chainlink struct {
		Nodes []NodeInfoYAML `yaml:"nodes"`
	} `yaml:"chainlink"`
}

// NodeInfoConfig represents node configuration that can be loaded from YAML
type NodeInfoConfig struct {
	Nodes []NodeInfoYAML `yaml:"nodes"`
}

// NodeInfoYAML maps directly to devenv.NodeInfo with user-friendly YAML field names
type NodeInfoYAML struct {
	Name            string              `yaml:"name"`
	ChainlinkConfig ChainlinkConfigYAML `yaml:"chainlink_config"`
	P2PPort         string              `yaml:"p2p_port"`
	IsBootstrap     bool                `yaml:"is_bootstrap"`
	AdminAddr       string              `yaml:"admin_addr"`
	MultiAddr       string              `yaml:"multi_addr"`
	Labels          map[string]string   `yaml:"labels"`
	ContainerName   string              `yaml:"container_name,omitempty"`
}

// ChainlinkConfigYAML maps to clclient.ChainlinkConfig with YAML tags
type ChainlinkConfigYAML struct {
	URL        string            `yaml:"url"`
	Email      string            `yaml:"email"`
	Password   string            `yaml:"password"`
	InternalIP string            `yaml:"internal_ip,omitempty"`
	Headers    map[string]string `yaml:"headers,omitempty"`
}

// ToDevenvNodeInfo converts NodeInfoYAML to devenv.NodeInfo
func (n NodeInfoYAML) ToDevenvNodeInfo() devenv.NodeInfo {
	return devenv.NodeInfo{
		Name: n.Name,
		CLConfig: clclient.ChainlinkConfig{
			URL:        n.ChainlinkConfig.URL,
			Email:      n.ChainlinkConfig.Email,
			Password:   n.ChainlinkConfig.Password,
			InternalIP: n.ChainlinkConfig.InternalIP,
			Headers:    n.ChainlinkConfig.Headers,
		},
		P2PPort:       n.P2PPort,
		IsBootstrap:   n.IsBootstrap,
		AdminAddr:     n.AdminAddr,
		MultiAddr:     n.MultiAddr,
		Labels:        n.Labels,
		ContainerName: n.ContainerName,
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <config.yaml>")
		fmt.Println("       go run main.go deploy")
		os.Exit(1)
	}

	configPath := "config.yaml"
	if len(os.Args) > 1 && os.Args[1] != "deploy" {
		configPath = os.Args[1]
	}

	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()
	lggr, closeLogger := logger.NewLogger()
	defer closeLogger()

	fmt.Println("🚀 Starting CCIP Deployment...")
	fmt.Printf("📍 Home Chain: %s (ID: %d)\n", config.HomeChain.Name, config.HomeChain.ChainID)
	fmt.Printf("📊 Feed Chain: %s (ID: %d)\n", config.FeedChain.Name, config.FeedChain.ChainID)
	fmt.Printf("🔗 Additional Chains: %d\n", len(config.Chains))

	nodeInfos := make([]devenv.NodeInfo, len(config.Chainlink.Nodes))
	for i, node := range config.Chainlink.Nodes {
		nodeInfos[i] = node.ToDevenvNodeInfo()
	}

	log.Printf("🎯 Loaded %d nodes from config", len(nodeInfos))

	// Build environment config with NodeInfo
	envConfig := buildEnvironmentConfig(config, nodeInfos)

	// Calculate chain selectors
	homeChainSel, err := chain_selectors.SelectorFromChainId(uint64(config.HomeChain.ChainID))
	if err != nil {
		log.Fatalf("Failed to get home chain selector: %v", err)
	}
	feedChainSel, err := chain_selectors.SelectorFromChainId(uint64(config.FeedChain.ChainID))
	if err != nil {
		log.Fatalf("Failed to get feed chain selector: %v", err)
	}

	fmt.Printf("🆔 Home Chain Selector: %d\n", homeChainSel)
	fmt.Printf("🆔 Feed Chain Selector: %d\n", feedChainSel)

	// Step 1: Deploy Home Chain Contracts
	fmt.Println("\n📦 Step 1: Deploying Home Chain Contracts...")
	capRegistry, addressBook, err := crib.DeployHomeChainContracts(
		ctx, lggr, envConfig, homeChainSel, feedChainSel,
	)
	if err != nil {
		log.Fatalf("Failed to deploy home chain contracts: %v", err)
	}
	fmt.Println("✅ Home chain contracts deployed successfully!")

	// Step 2: Deploy CCIP on all chains and add lanes
	fmt.Println("\n🌐 Step 2: Deploying CCIP and Adding Lanes...")
	output, err := crib.DeployCCIPAndAddLanes(
		ctx, lggr, envConfig, homeChainSel, feedChainSel, addressBook, config.Deployment.RMNEnabled,
	)
	if err != nil {
		log.Fatalf("Failed to deploy CCIP and add lanes: %v", err)
	}
	fmt.Println("✅ CCIP deployed and lanes added successfully!")

	// Print deployment summary
	printDeploymentSummary(output, capRegistry)

	fmt.Println("\n🎉 CCIP Deployment Complete!")
	fmt.Println("💡 Next steps:")
	fmt.Println("   - Test cross-chain messaging")
	fmt.Println("   - Monitor the lanes")
	fmt.Println("   - Add more chains if needed")
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &config, nil
}

func buildEnvironmentConfig(config *Config, nodeInfos []devenv.NodeInfo) devenv.EnvironmentConfig {
	var chains []devenv.ChainConfig

	// Add home chain
	homeChain := devenv.ChainConfig{
		ChainID:   strconv.FormatInt(config.HomeChain.ChainID, 10),
		ChainName: config.HomeChain.Name,
		ChainType: "EVM", // Assuming EVM chains
		WSRPCs: []devenv.CribRPCs{{
			External: config.HomeChain.RPCURL,
		}},
		HTTPRPCs: []devenv.CribRPCs{{
			External: config.HomeChain.RPCURL,
		}},
	}
	// Set deployer key
	if err := homeChain.SetDeployerKey(&config.HomeChain.PrivateKey); err != nil {
		log.Fatalf("Failed to set deployer key for home chain: %v", err)
	}
	chains = append(chains, homeChain)

	// Add feed chain if different from home chain
	if config.FeedChain.ChainID != config.HomeChain.ChainID {
		feedChain := devenv.ChainConfig{
			ChainID:   strconv.FormatInt(config.FeedChain.ChainID, 10),
			ChainName: config.FeedChain.Name,
			ChainType: "EVM",
			WSRPCs: []devenv.CribRPCs{{
				External: config.FeedChain.RPCURL,
			}},
			HTTPRPCs: []devenv.CribRPCs{{
				External: config.FeedChain.RPCURL,
			}},
		}
		if err := feedChain.SetDeployerKey(&config.FeedChain.PrivateKey); err != nil {
			log.Fatalf("Failed to set deployer key for feed chain: %v", err)
		}
		chains = append(chains, feedChain)
	}

	// Add additional chains
	for _, chain := range config.Chains {
		if chain.ChainID != config.HomeChain.ChainID && chain.ChainID != config.FeedChain.ChainID {
			chainConfig := devenv.ChainConfig{
				ChainID:   strconv.FormatInt(chain.ChainID, 10),
				ChainName: chain.Name,
				ChainType: "EVM",
				WSRPCs: []devenv.CribRPCs{{
					External: chain.RPCURL,
				}},
				HTTPRPCs: []devenv.CribRPCs{{
					External: chain.RPCURL,
				}},
			}
			if err := chainConfig.SetDeployerKey(&chain.PrivateKey); err != nil {
				log.Fatalf("Failed to set deployer key for chain %s: %v", chain.Name, err)
			}
			chains = append(chains, chainConfig)
		}
	}

	// Get JD configuration from environment
	jdGRPC := os.Getenv("JD_GRPC_URL")
	if jdGRPC == "" {
		jdGRPC = "localhost:50051"
	}
	jdWSRPC := os.Getenv("JD_WSRPC_URL")
	if jdWSRPC == "" {
		jdWSRPC = "ws://localhost:50051"
	}

	return devenv.EnvironmentConfig{
		Chains: chains,
		JDConfig: devenv.JDConfig{
			GRPC:     jdGRPC,
			WSRPC:    jdWSRPC,
			Creds:    insecure.NewCredentials(),
			NodeInfo: nodeInfos,
		},
	}
}

func printDeploymentSummary(output crib.DeployCCIPOutput, capRegistry deployment.CapabilityRegistryConfig) {
	fmt.Println("\n📋 Deployment Summary:")
	fmt.Println(strings.Repeat("=", 50))

	if len(output.NodeIDs) > 0 {
		fmt.Printf(" Deployed Nodes: %d\n", len(output.NodeIDs))
		for i, nodeID := range output.NodeIDs {
			fmt.Printf("   - Node %d: %s\n", i+1, nodeID)
		}
	}

	fmt.Printf("🏠 Capability Registry Address: %s\n", capRegistry.Contract.Hex())
	fmt.Printf("⛓️  Capability Registry Chain: %d\n", capRegistry.EVMChainID)
	fmt.Println(strings.Repeat("=", 50))
}
