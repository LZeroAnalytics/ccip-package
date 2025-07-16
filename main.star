chainlink_pkg = import_module("./src/chainlink-node-package/main.star")
hardhat_package = import_module("./src/hardhat-package/main.star")
ocr = import_module("./src/chainlink-node-package/src/ocr/ocr.star")
DON_NODES_COUNT = 6
CHAINLINK_IMAGE = "fravlaca/chainlink:0.3.0"
CCIP_UI_IMAGE = "fravlaca/ccip-ui:0.1.0"

# Auto-incremented version from GitHub workflow - update this when you want to use a newer build
# Check https://hub.docker.com/r/fravlaca/hardhat-ccip-contracts/tags for latest versions
HARDHAT_IMAGE = "fravlaca/hardhat-ccip-contracts:1.0.1"

def run(plan, args = {}):
    config = args
    env = args["env"]
   
    # Setup hardhat environment for contracts deployment
    hardhat_package.run(plan, image=HARDHAT_IMAGE+"-"+env) #project_url="github.com/LZeroAnalytics/hardhat-ccip-contracts"#, image="fravlaca/hardhat-ccip-contracts:0.1.0")
    networks = {}
    for chain in config["chains"]:
        networks[chain["name"]] = {
            "rpc_url": chain["rpc_url"],
            "chain_id": chain["chain_id"],
            "private_key": chain["private_key"],
            "existing_contracts": chain["existing_contracts"]
        }
    hardhat_package.configure_networks(plan, networks)

    hardhat_package.compile(plan)

    home_chain_contracts = deploy_home_chain_contracts(plan, config)

    # Start CCIP DON with proper node count (typically 7-13 nodes for production)
    deployed_nodes = start_don(plan, config, home_chain_contracts.capReg)

    nodes_infos = []
    for i in range(DON_NODES_COUNT):
        node_name = "chainlink-ccip-node-" + str(i)
        eth_keys = {}
        for chain in config["chains"]:
            eth_keys[chain["chain_id"]] = chainlink_pkg.node_utils.get_eth_key(plan, node_name, chain["chain_id"])
        nodes_infos.append({
            "node_name": node_name,
            "ocr_key": chainlink_pkg.node_utils.get_ocr_key(plan, node_name),
            "ocr_key_bundle_id": chainlink_pkg.node_utils.get_ocr_key_bundle_id(plan, node_name),
            "p2p_peer_id": chainlink_pkg.node_utils.get_p2p_peer_id(plan, node_name),
            "eth_keys": eth_keys
        })

    # Deploy and configure CCIP infrastructure FIRST
    chains_contracts = deploy_ccip_contracts_on_chains(plan, config, nodes_infos, home_chain_contracts)
    # Create CCIP jobs AFTER everything is configured
    bootstrap_node = plan.get_service(name="chainlink-ccip-node-0")
    p2pBootstraperID = nodes_infos[0]["p2p_peer_id"] + "@" + bootstrap_node.ip_address + ":" + str(bootstrap_node.ports["p2p"].number)
    ccip_jobs_result = _create_ccip_jobs(plan, nodes_infos, p2pBootstraperID)
    ccip_lanes_result = configure_ccip_lanes(plan, config, chains_contracts)


    don_ids_per_chain = config_ocr(plan, config, home_chain_contracts, chains_contracts, nodes_infos)

    contracts_addresses = struct(
        home_chain_contracts = home_chain_contracts,
        chains_contracts = chains_contracts,
        don_ids = don_ids_per_chain
    )

    #spinup_ccip_ui(plan, contracts_addresses, config["chains"])

    return struct(
        contracts_addresses = struct(
            home_chain_contracts = home_chain_contracts,
            chains_contracts = chains_contracts
        ),
        chainlink_nodes = deployed_nodes
    )


def start_don(plan, config, capReg):
    # Create database and configs for all chainlink nodes in parallel
    chainlink_node_configs = []
    for i in range(DON_NODES_COUNT):
        chainlink_node_configs.append({
            "node_name": "chainlink-ccip-node-" + str(i),
            "image": CHAINLINK_IMAGE
        })
    
    chains = []
    for chain in config["chains"]:
        chains.append({
            "rpc": chain["rpc_url"],
            "ws": chain["ws_url"],
            "chain_id": chain["chain_id"], 
            "existing_contracts": chain["existing_contracts"]
        })
    
    # Pass the chainlink nodes configuration to the chainlink package
    result = chainlink_pkg.deployment.deploy_nodes(plan, args = { 
        "chains": chains,
        "chainlink_nodes": chainlink_node_configs
    }, capabilitiesRegistry=capReg)
    
    # Fund nodes if faucet provided
    for chain in config["chains"]:
        faucet_url = chain["faucet"]
        for node_name in result.services.keys():
            eth_key = chainlink_pkg.node_utils.get_eth_key(plan, node_name, chain["chain_id"])
            chainlink_pkg.node_utils.fund_eth_key(plan, eth_key, faucet_url)

    return result.services


def deploy_home_chain_contracts(plan, config):    
    # Build deployment parameters and extra cmds
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
        extraCmds = output_grep_cmd
    )

    return struct(
        capReg = deployment_homechain_result["extract.capabilitiesRegistry"],
        ccipHome = deployment_homechain_result["extract.ccipHome"],
        rmnHome = deployment_homechain_result["extract.rmnHome"]
    )

def deploy_ccip_contracts_on_chains(plan, config, nodes_infos, home_chain_contracts):
    # Build chain configurations
    chain_selectors = []
    readers = []
    fChain = (DON_NODES_COUNT-1)/3
    for chain in config["chains"]:
        chain_selectors.append(str(chain["chain_selector"]))  # Use chain_id as selector for now
    for i in range(DON_NODES_COUNT):
        readers.append(nodes_infos[i]["p2p_peer_id"])  # Use first node's address as reader for now
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
    #CCIP HOME - updateChainConfig - add new chains (selector, readers, f, capId) adn add nodes to don capreg
    hardhat_package.script(
        plan, 
        "scripts/deploy/ccip-v1_6/03-01-update-home-chain-configs.ts", 
        network = config["chains"][0]["name"],
        params = {
            "READERS": json.encode(readers),
            "F_CHAIN": fChain,
            "CHAIN_SELECTORS": json.encode(chain_selectors),
            "CCIP_HOME_ADDRESS": home_chain_contracts.ccipHome,
            "CAPABILITIES_REGISTRY": home_chain_contracts.capReg,
            "NODE_OPERATOR": node_operator_id,
            "NODES_PARAMS": json.encode(nodes_params)
        },
        extraCmds = " | grep -A 100 DEPLOYMENT_JSON_BEGIN | grep -B 100 DEPLOYMENT_JSON_END | sed '/DEPLOYMENT_JSON_BEGIN/d' | sed '/DEPLOYMENT_JSON_END/d'"
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
                "weth9": "weth9",
                "tokenPoolFactory": "tokenPoolFactory"
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
            params = {
                "PARAMS": json.encode(args),
                "LINK_TOKEN": chain["existing_contracts"]["link_token"],
                "WETH9": chain["existing_contracts"]["weth9"]
            },
            extraCmds= " | grep -A 100 DEPLOYMENT_JSON_BEGIN | grep -B 100 DEPLOYMENT_JSON_END | sed '/DEPLOYMENT_JSON_BEGIN/d' | sed '/DEPLOYMENT_JSON_END/d'"
        )

        chain_contracts[chain["name"]] = struct(
            rmnProxy = prereq_result["extract.rmnProxy"],
            tokenAdminRegistry = prereq_result["extract.tokenAdminRegistry"],
            registryModule = prereq_result["extract.registryModule"],
            router = prereq_result["extract.router"],
            tokenPoolFactory = prereq_result["extract.tokenPoolFactory"],
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
    plan.print("lanes_config: ")
    plan.print(lanes_config)
    hardhat_package.script(
        plan, 
        "scripts/deploy/ccip-v1_6/05-configure-lanes.ts", 
        network = config["chains"][0]["name"],
        params = {
            "CONFIG": json.encode(lanes_config)
        },
        extraCmds = " | grep -A 100 DEPLOYMENT_JSON_BEGIN | grep -B 100 DEPLOYMENT_JSON_END | sed '/DEPLOYMENT_JSON_BEGIN/d' | sed '/DEPLOYMENT_JSON_END/d'"
    )

def _create_ccip_jobs(plan, nodes_infos, p2pBootstraperID):
    """Create and sets up CCIP jobs on all nodes"""
    for i in range(DON_NODES_COUNT):
        if i == 0:
            # Bootstrap job for CCIP
            chainlink_pkg.node_utils.create_job(plan, nodes_infos[i]["node_name"], "ccip-bootstrap-job-template.toml", {
                "P2P_KEY_ID": nodes_infos[i]["p2p_peer_id"]
            })
        else:
            # Bootstrap job for CCIP
            chainlink_pkg.node_utils.create_job(plan, nodes_infos[i]["node_name"], "ccip-job-template.toml", {
                "P2P_V2_BOOTSTRAPPER": p2pBootstraperID,
                "P2P_KEY_ID": nodes_infos[i]["p2p_peer_id"],
                "OCR2_KEY_BUNDLE": nodes_infos[i]["ocr_key_bundle_id"]
            })

def config_ocr(plan, config, home_chain_contracts, chains_contracts, nodes_infos):
    """Configure OCR3 for CCIP commit and exec plugins on all chains."""
    
    # Extract required information
    home_chain_selector = config["chains"][0]["chain_selector"]
    feed_chain_selector = config["chains"][0]["chain_selector"]  # Using home as feed for now
    
    ocr.init_ocr3_service(plan)

    don_ids_per_chain = {}
    # Setup OCR3 for each chain
    for chain_config in config["chains"]:
        nodes_data = []
        for node_infos in nodes_infos:
            node_info = {
                "onchainKey": node_infos["ocr_key"].on_chain_key,
                "offchainKey": node_infos["ocr_key"].off_chain_key,
                "configKey": node_infos["ocr_key"].config_key,
                "peerID": node_infos["p2p_peer_id"],
                "transmitter": node_infos["eth_keys"][chain_config["chain_id"]]
            }
            nodes_data.append(node_info)
        chain_selector = chain_config["chain_selector"]
        chain_contracts = chains_contracts[chain_config["name"]]
        
        # Setup both commit and exec plugins
        # Generate OCR3 config using the existing generator
        commit_ocr3_input = {
            "nodes": nodes_data,
            "pluginType": "commit",
            "chainSelector": str(chain_selector),
            "feedChainSelector": str(feed_chain_selector)
        }
        
        commit_ocr3_result = ocr.generate_ocr3config(plan, commit_ocr3_input)

        commit_ocr3_input["pluginType"] = "exec"
        exec_ocr3_result = ocr.generate_ocr3config(plan, commit_ocr3_input)
        
        signers = []
        transmitters = []
        ocr3_nodes = []
        for i, node in enumerate(nodes_data):
            ocr3_nodes.append({
                "p2pId": node["peerID"],
                "signerKey": commit_ocr3_result["extract.signer_{}".format(i)],
                "transmitterKey": commit_ocr3_result["extract.transmitter_{}".format(i)]
            })
            signers.append(commit_ocr3_result["extract.signer_{}".format(i)])
            transmitters.append(commit_ocr3_result["extract.transmitter_{}".format(i)])

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
                "signers": signers,           # Use direct array
                "transmitters": transmitters, # Use direct array
                "f": commit_ocr3_result["extract.f"],
                "offchainConfigVersion": commit_ocr3_result["extract.offchain_cfg_ver"],
                "offchainConfig": commit_ocr3_result["extract.offchain_cfg"],
                "nodes": ocr3_nodes
            },
            "execOCR3Config": {
                "signers": signers,             # Use direct array
                "transmitters": transmitters,   # Use direct array
                "f": exec_ocr3_result["extract.f"],
                "offchainConfigVersion": exec_ocr3_result["extract.offchain_cfg_ver"],
                "offchainConfig": exec_ocr3_result["extract.offchain_cfg"],
                "nodes": ocr3_nodes
            }
        }

        plan.print("Setting up OCR3 ccip commit and exec plugin for chain {}".format(chain_config["name"]))

        existing_don_id = don_ids_per_chain.get(chain_config["name"])
        if existing_don_id:
            setup_params["existingDonId"] = existing_don_id

        plan.print("setup_params: ")
        plan.print(setup_params)
        
        # Run the TypeScript setup script
        result = hardhat_package.script(
            plan,
            "scripts/deploy/ccip-v1_6/06-setup-ocr.ts",
            network = config["chains"][0]["name"],
            return_keys= {
                "donID": "donID"
            },
            params = {
                "SETUP_PARAMS": json.encode(setup_params)
            },
            extraCmds = " | grep -A 100 DEPLOYMENT_JSON_BEGIN | grep -B 100 DEPLOYMENT_JSON_END | sed '/DEPLOYMENT_JSON_BEGIN/d' | sed '/DEPLOYMENT_JSON_END/d'"
        )

        don_id = result["extract.donID"]
        don_ids_per_chain[chain_config["name"]] = don_id
    
    return don_ids_per_chain

def spinup_ccip_ui(plan, contracts_addresses, chains_config):
    """Spins up the CCIP UI with dynamically generated network configuration"""
    
    # Prepare template data
    networks = []
    cct_tokens_input = [] # TODO: deploy CCT tokens (BnM and LnM examples)
    cct_tokens = []
    
    for chain in chains_config:
        chain_contracts = contracts_addresses.chains_contracts.get(chain["name"], {})
        
        # Build network
        network = {
            "ChainID": chain["chain_id"],
            "Key": chain["name"],
            "Name": chain.get("display_name", chain["name"].title()),
            "NativeCurrency": {
                "Name": chain.get("native_currency_name", "ETH"),
                "Symbol": chain.get("native_currency_symbol", "ETH"),
                "Decimals": chain.get("native_currency_decimals", 18)
            },
            "RpcUrl": chain["rpc_url"],
            "Explorer": {
                "Name": chain.get("explorer_name", "Explorer"),
                "Url": chain.get("explorer_url", "")
            },
            "LogoURL": _get_chain_logo(chain["name"]),
            "Testnet": True,
            "ChainSelector": chain["chain_selector"],
            "LinkContract": chain_contracts.linkToken or chain["existing_contracts"]["link_token"],
            "RouterAddress": chain_contracts.router
        }
        networks.append(network)

    
    # Build tokens
    token_configs = {
        "CCIP-BnM": {"logo": "ccip-bnm", "tags": ["chainlink", "default"]},
        "LINK": {"logo": "link", "tags": ["chainlink", "default"]},
        "WETH": {"logo": "weth", "tags": ["wrapped", "default"]}
    }
    
    
    for cct_token in cct_tokens:
        cct_tokens.append({
            "Symbol": cct_token.symbol,
            "LogoURL": cct_token.logo_url,
            "Tags": json.encode(cct_token.tags),
            "Addresses": cct_token.addresses # TODO: map chain id per corresponding address for token in that chain
        })
    
    # Generate config artifact
    config_artifact = plan.render_templates(
        name = "ccip-ui-config",
        config = {
            "network-config.yaml": struct(
                template = read_file("./ccip-ui-template.yaml"),
                data = {"Networks": networks, "Tokens": cct_tokens}
            )
        }
    )
    
    # Start service
    ccip_ui = plan.add_service(
        name = "ccip-ui",
        config = ServiceConfig(
            image = CCIP_UI_IMAGE,
            ports = {"http": PortSpec(number = 3000, transport_protocol = "TCP")},
            files = {"/app/public": config_artifact},
            env_vars = {"NEXT_PUBLIC_CCIP_CONFIG_FILE": "/network-config.yaml"}
        )
    )
    
    plan.print("CCIP UI started at: http://{}:{}".format(ccip_ui.ip_address, ccip_ui.ports["http"].number))
    return ccip_ui

def _get_chain_logo(chain_name):
    """Returns logo URL for known chains"""
    logos = {"ethereum": "ethereum", "sepolia": "ethereum", "arbitrum": "arbitrum", 
             "optimism": "optimism", "polygon": "polygon", "avalanche": "avalanche", 
             "bsc": "bsc", "base": "base"}
    
    for key, logo in logos.items():
        if key in chain_name.lower():
            return "https://d2f70xi62kby8n.cloudfront.net/bridge/icons/networks/{}.svg?auto=compress%2Cformat".format(logo)
    return "https://d2f70xi62kby8n.cloudfront.net/bridge/icons/networks/ethereum.svg?auto=compress%2Cformat"

def replace_http_with_ws(rpc_url):
    """Convert HTTP RPC URL to WebSocket URL"""
    return rpc_url.replace("http://", "ws://").replace("https://", "wss://")

    