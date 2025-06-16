chainlink_pkg = import_module("./src/chainlink-node-package/main.star")
hardhat_package = import_module("github.com/LZeroAnalytics/hardhat-package/main.star")
ocr = import_module("./src/chainlink-node-package/src/ocr/ocr.star")
DON_NODES_COUNT = 6

def run(plan, args = {}):
    config = args
   
    # Start CCIP DON with proper node count (typically 7-13 nodes for production)
    deployed_nodes = start_don(plan, config)

    # Setup hardhat environment for contracts deployment
    hardhat_package.run(plan, "github.com/LZeroAnalytics/hardhat-ccip-contracts")
    networks = {}
    for chain in config["chains"]:
        networks[chain["name"]] = {
            "url": chain["rpc_url"],
            "chainId": chain["chain_id"],
            "privateKey": chain["private_key"]
        }
    hardhat_package.configure_networks(plan, networks)

    nodes_infos = []
    for i in range(DON_NODES_COUNT):
        node_name = "chainlink-ccip-node-" + str(i)
        nodes_infos.append({
            "node_name": node_name,
            "ocr_key": chainlink_pkg.node_utils.get_ocr_key(plan, node_name),
            "ocr_key_bundle_id": chainlink_pkg.node_utils.get_ocr_key_bundle_id(plan, node_name),
            "p2p_peer_id": chainlink_pkg.node_utils.get_p2p_peer_id(plan, node_name),
            "eth_key": chainlink_pkg.node_utils.get_eth_key(plan, node_name)
        })
    # Deploy and configure CCIP infrastructure
    home_chain_contracts = deploy_home_chain_contracts(plan, config, deployed_nodes, nodes_infos)
    chains_contracts = deploy_ccip_contracts_on_chains(plan, config, nodes_infos)
    p2pBootstraperID = nodes_infos[0]["p2p_peer_id"] + "@" + deployed_nodes["chainlink-ccip-node-0"].ip_address + ":" + str(deployed_nodes["chainlink-ccip-node-0"].ports["p2p"].number)
    ccip_jobs_result = _create_ccip_jobs(plan, nodes_infos, p2pBootstraperID) 
    ccip_lanes_result = configure_ccip_lanes(plan, config, chains_contracts)
    config_ocr(plan, config, home_chain_contracts, chains_contracts, nodes_infos)

    return struct(
        contracts_addresses = struct(
            home_chain_contracts = home_chain_contracts,
            chains_contracts = chains_contracts
        ),
        chainlink_nodes = config["chainlink"]["nodes"]
    )


def start_don(plan, config):
    # Create database and configs for all chainlink nodes in parallel
    chainlink_node_configs = []
    for i in range(DON_NODES_COUNT):
        chainlink_node_configs.append({
            "node_name": "chainlink-ccip-node-" + str(i),
        })
    
    chains = []
    for chain in config["chains"]:
        chains.append({
            "rpc": chain["rpc_url"],
            "ws": replace_http_with_ws(chain["rpc_url"]),
            "chain_id": chain["chain_id"]
        })
    
    # Pass the chainlink nodes configuration to the chainlink package
    result = chainlink_pkg.run(plan, args = { 
        "chains": chains,
        "chainlink_nodes": chainlink_node_configs
    })
    
    # Fund nodes if faucet provided
    for chain in config["chains"]:
        faucet_url = chain["faucet"]
        for node_name in result.services.keys():
            eth_key = chainlink_pkg.node_utils.get_eth_key(plan, node_name)
            chainlink_pkg.node_utils.fund_eth_key(plan, eth_key, faucet_url)

    return result.services


def deploy_home_chain_contracts(plan, config, nodes_result, nodes_infos):    
    # 2. Collect comprehensive node information
    node_operator_id = "CCIPOperator1"
    nodes_params = []
    for i in range(DON_NODES_COUNT):        
        # Build node params to add to CCIP Home contract in home chain
        ccip_node_param = {
            "nodeOperatorId": node_operator_id,
            "p2pId": nodes_infos[i]["p2p_peer_id"],    # Ensure 0x prefix
            "signer": nodes_infos[i]["ocr_key"].off_chain_key,             # OCR off-chain signing key (Ed25519, for consensus)
            "encryptionPublicKey": nodes_infos[i]["ocr_key"].config_key  # OCR config encryption key (X25519)
            # hashedCapabilityIds will be calculated in TypeScript (set to default for ccip v1.00)
        }
        nodes_params.append(ccip_node_param)
  
    # Build deployment parameters and extra cmds
    deployment_params = [node_operator_id, nodes_params]
    args = " -- "+ "'" + node_operator_id + "'" + " '" + json.encode(nodes_params) + "'"
    output_grep_cmd = " | grep -A 100 DEPLOYMENT_JSON_BEGIN | grep -B 100 DEPLOYMENT_JSON_END | sed '/DEPLOYMENT_JSON_BEGIN/d' | sed '/DEPLOYMENT_JSON_END/d'"

    # 5. Deploy CCIP home chain contracts
    deployment_homechain_result = hardhat_package.script(
        plan, 
        "scripts/deploy/ccip-v1_6/02-deploy-home-chain.ts", 
        network = config["chains"][0]["name"],
        return_keys = {
            "capabilitiesRegistry": "capabilitiesRegistry", 
            "ccipHome": "ccipHome",
            "rmnHome": "rmnHome"
        },
        extraCmds = args + output_grep_cmd
    )

    return struct(
        capReg = deployment_homechain_result["extract.capabilitiesRegistry"],
        ccipHome = deployment_homechain_result["extract.ccipHome"],
        rmnHome = deployment_homechain_result["extract.rmnHome"]
    )

def deploy_ccip_contracts_on_chains(plan, config, nodes_infos):
    # Build chain configurations
    chain_selectors = []
    readers = []
    fChain = (DON_NODES_COUNT-1)/3
    for chain in config["chains"]:
        chain_selectors.append(str(chain["chain_selector"]))  # Use chain_id as selector for now
    for i in range(DON_NODES_COUNT):
        readers.append(nodes_infos[i]["p2p_peer_id"])  # Use first node's address as reader for now
    #CCIP HOME - updateChainConfig - add new chains (selector, readers, f, capId)
    hardhat_package.script(
        plan, 
        "scripts/deploy/ccip-v1_6/03-01-update-home-chain-configs.ts", 
        network = config["chains"][0]["name"],
        extraCmds = " -- "+ "'" + json.encode(readers)+ "'" + " '" + str(fChain) + "'" + " '" + json.encode(chain_selectors)+ "'"
    )

    chain_contracts = {}

    for chain in config["chains"]:
        prereq_result = hardhat_package.script(
            plan, 
            "scripts/deploy/ccip-v1_6/03-deploy-chain-pre-req.ts", 
            network = chain["name"],
            return_keys = {
                "rmnProxy": "rmnProxy",
                "tokenAdminRegistry": "tokenAdminRegistry",
                "registryModule": "registryModule",
                "router": "router",
                "linkToken": "linkToken",
                "weth9": "weth9"
            },
            params = {
                "LINK_TOKEN": chain["existing_contracts"]["link_token"],
                "WETH9": chain["existing_contracts"]["weth9"]
            },
            extraCmds = " | grep -A 100 DEPLOYMENT_JSON_BEGIN | grep -B 100 DEPLOYMENT_JSON_END | sed '/DEPLOYMENT_JSON_BEGIN/d' | sed '/DEPLOYMENT_JSON_END/d'"
        )

        # Deploy core ccip contract for each chain
        args = struct(
            chainSelector = chain["chain_selector"],
            rmnProxy = prereq_result["extract.rmnProxy"],
            tokenAdminRegistry = prereq_result["extract.tokenAdminRegistry"]
        )
        homechain_contracts = hardhat_package.script(
            plan, 
            "scripts/deploy/ccip-v1_6/04-deploy-chain-contracts.ts", 
            network = chain["name"],
            return_keys= {
                "nonceManager": "nonceManager",
                "feeQuoter": "feeQuoter",
                "onRamp": "onRamp",
                "offRamp": "offRamp"
            },
            extraCmds= " -- "+ "'" + json.encode(args)+ "'" + " | grep -A 100 DEPLOYMENT_JSON_BEGIN | grep -B 100 DEPLOYMENT_JSON_END | sed '/DEPLOYMENT_JSON_BEGIN/d' | sed '/DEPLOYMENT_JSON_END/d'"
        )

        chain_contracts[chain["name"]] = struct(
            rmnProxy = prereq_result["extract.rmnProxy"],
            tokenAdminRegistry = prereq_result["extract.tokenAdminRegistry"],
            registryModule = prereq_result["extract.registryModule"],
            router = prereq_result["extract.router"],
            linkToken = prereq_result["extract.linkToken"],
            weth9 = prereq_result["extract.weth9"],
            nonceManager = homechain_contracts["extract.nonceManager"],
            feeQuoter = homechain_contracts["extract.feeQuoter"],
            onRamp = homechain_contracts["extract.onRamp"],
            offRamp = homechain_contracts["extract.offRamp"]
        )

    return chain_contracts



def configure_ccip_lanes(plan, config, ccip_chains_result):
    # Build bidirectional lanes configuration for all chain pairs
    lanes = []
    for chain in config["chains"]:
        chain_contracts = ccip_chains_result[chain["name"]]
        lanes.append({
            "sourceChainSelector": str(chain["chain_selector"]),
            "chainName": str(chain["name"]),
            "onRamp": chain_contracts.onRamp,
            "offRamp": chain_contracts.offRamp,
            "feeQuoter": chain_contracts.feeQuoter,
            "router": chain_contracts.router,
            "linkToken": chain_contracts.linkToken,
            "weth9": chain_contracts.weth9
        })
    
    lanes_config = {"lanes": lanes}
    
    # Configure lanes on first chain (script handles all chains)
    hardhat_package.script(
        plan, 
        "scripts/deploy/ccip-v1_6/05-configure-lanes.ts", 
        network = config["chains"][0]["name"],
        extraCmds = " -- '" + json.encode(lanes_config) + "'"
    )

def _create_ccip_jobs(plan, nodes_infos, p2pBootstraperID):
    """Create and sets up CCIP jobs on all nodes"""
    for i in range(DON_NODES_COUNT):
        if i == 0:
            p2pBootID = None
            keyBundle = None
        else:
            p2pBootID = p2pBootstraperID
            keyBundle = nodes_infos[i]["ocr_key_bundle_id"]
        
        # Bootstrap job for CCIP
        chainlink_pkg.node_utils.create_job(plan, nodes_infos[i]["node_name"], "ccip-job-template.toml", {
            "P2P_V2_BOOTSTRAPPER": p2pBootID,
            "P2P_KEY_ID": nodes_infos[i]["p2p_peer_id"],
            "OCR2_KEY_BUNDLE": keyBundle
        })

def config_ocr(plan, config, home_chain_contracts, chains_contracts, nodes_infos):
    """Configure OCR3 for CCIP commit and exec plugins on all chains."""
    
    # Extract required information
    home_chain_selector = config["chains"][0]["chain_selector"]
    feed_chain_selector = config["chains"][0]["chain_selector"]  # Using home as feed for now
    
    # Collect node OCR3 information
    nodes_data = []
    for node_infos in nodes_infos:
        node_info = {
            "onchainKey": node_infos["ocr_key"].on_chain_key,
            "offchainKey": node_infos["ocr_key"].off_chain_key,
            "configKey": node_infos["ocr_key"].config_key,
            "peerID": node_infos["p2p_peer_id"],
            "transmitter": node_infos["eth_key"]
        }
        nodes_data.append(node_info)
        
    # Setup OCR3 for each chain
    for chain_config in config["chains"]:
        chain_selector = chain_config["chain_selector"]
        chain_contracts = chains_contracts[chain_config["name"]]
        
        # Setup both commit and exec plugins
        for plugin_type in ["commit", "exec"]:
            plan.print("Setting up OCR3 {} plugin for chain {}".format(plugin_type, chain_config["name"]))
            
            # Generate OCR3 config using the existing generator
            commit_ocr3_input = {
                "nodes": nodes_data,
                "pluginType": "commit",
                "chainSelector": str(chain_selector),
                "feedChainSelector": str(feed_chain_selector)
            }
            
            commit_ocr3_result = ocr.generate_ocr3config(plan, commit_ocr3_input)
            commit_ocr3_config = json.decode(commit_ocr3_result)

            commit_ocr3_input["pluginType"] = "exec"
            exec_ocr3_result = ocr.generate_ocr3config(plan, commit_ocr3_input)
            exec_ocr3_config = json.decode(exec_ocr3_result)
            
            # Build nodes array with p2pId, signerKey, transmitterKey
            ocr3_nodes = []
            for i, node in enumerate(nodes_data):
                ocr3_nodes.append({
                    "p2pId": node["peerID"],
                    "signerKey": commit_ocr3_config["signers"][i],
                    "transmitterKey": commit_ocr3_config["transmitters"][i]
                })
            
            # Prepare parameters for TypeScript script
            setup_params = {
                "homeChainSelector": str(home_chain_selector),
                "remoteChainSelector": str(chain_selector),
                "chainName": chain_config["name"],
                "feedChainSelector": str(feed_chain_selector),
                "ccipHome": home_chain_contracts.ccipHome,
                "capabilitiesRegistry": home_chain_contracts.capReg,
                "offRamp": chain_contracts.offRamp,
                "rmnHome": home_chain_contracts.rmnHome,
                "commitOCR3Config": {
                    "signers": commit_ocr3_config["signers"],
                    "transmitters": commit_ocr3_config["transmitters"],
                    "f": commit_ocr3_config["f"],
                    "offchainConfigVersion": commit_ocr3_config["offchainConfigVersion"],
                    "offchainConfig": commit_ocr3_config["offchainConfig"],
                    "nodes": ocr3_nodes
                },
                "execOCR3Config": {
                    "signers": exec_ocr3_config["signers"],
                    "transmitters": exec_ocr3_config["transmitters"],
                    "f": exec_ocr3_config["f"],
                    "offchainConfigVersion": exec_ocr3_config["offchainConfigVersion"],
                    "offchainConfig": exec_ocr3_config["offchainConfig"],
                    "nodes": ocr3_nodes
                }
            }
            
            # Run the TypeScript setup script
            hardhat_package.script(
                plan,
                "scripts/deploy/ccip-v1_6/06-setup-ocr.ts",
                network = config["chains"][0]["name"],
                params = json.encode(setup_params)
            )
            
            plan.print("✅ OCR3 {} plugin configured for chain {}".format(plugin_type, chain_config["name"]))
    
    plan.print("✅ OCR3 configuration completed for all chains")


def replace_http_with_ws(rpc_url):
    """Convert HTTP RPC URL to WebSocket URL"""
    return rpc_url.replace("http://", "ws://").replace("https://", "wss://")

    