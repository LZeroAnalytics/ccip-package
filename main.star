def run(plan, args):
    # Read config.yaml to get network configurations
    config_yaml = plan.read_file("config.yaml")
    
    # Parse configuration from args with defaults
    chains = args.get("chains", {})
    private_key = args.get("private_key", "")
    home_chain = args.get("home_chain", "")
    
    # Extract preexisting contracts from config.yaml
    # Since Starlark doesn't have a built-in YAML parser, we'll parse the relevant sections manually
    preexisting_contracts = {}
    
    # Simple parser for the link_addresses section in config.yaml
    networks = []
    in_network = False
    current_network = {}
    
    for line in config_yaml.split("\n"):
        line = line.strip()
        
        if line.startswith("- enclaveId:"):
            if current_network:
                networks.append(current_network)
            current_network = {}
            in_network = True
        
        if in_network:
            if line.startswith("chain_id:"):
                current_network["chain_id"] = line.split(":")[1].strip().strip('"')
            elif line.startswith("link_token_address:"):
                if "link_addresses" not in current_network:
                    current_network["link_addresses"] = {}
                current_network["link_addresses"]["link_token_address"] = line.split(":")[1].strip().strip('"')
            elif line.startswith("link_native_token_feed_address:"):
                if "link_addresses" not in current_network:
                    current_network["link_addresses"] = {}
                current_network["link_addresses"]["link_native_token_feed_address"] = line.split(":")[1].strip().strip('"')
    
    # Add the last network if it exists
    if current_network:
        networks.append(current_network)
    
    # Create preexisting contracts from parsed networks
    for network in networks:
        if "chain_id" in network and "link_addresses" in network:
            chain_name = f"chain_{network['chain_id']}"
            
            # Add Link token contract
            if "link_token_address" in network["link_addresses"]:
                preexisting_contracts[f"link_token_{chain_name}"] = {
                    "address": network["link_addresses"]["link_token_address"],
                    "chain": chain_name,
                    "type": "LinkToken"
                }
            
            # Add Link native token feed contract
            if "link_native_token_feed_address" in network["link_addresses"]:
                preexisting_contracts[f"price_feed_{chain_name}"] = {
                    "address": network["link_addresses"]["link_native_token_feed_address"],
                    "chain": chain_name,
                    "type": "PriceFeed"
                }
    
    # Merge with any additional preexisting contracts from args
    args_preexisting = args.get("preexisting_contracts", {})
    for key, value in args_preexisting.items():
        preexisting_contracts[key] = value
    
    # Create deployment config
    deployment_config = {
        "chains": chains,
        "private_key": private_key,
        "home_chain": home_chain,
        "num_nodes": args.get("num_nodes", 4),
        "num_bootstraps": args.get("num_bootstraps", 1),
        "enable_mercury": args.get("enable_mercury", False),
        "enable_log_triggers": args.get("enable_log_triggers", False),
        "preexisting_contracts": preexisting_contracts
    }
    
    # Add config file to the deployment
    config_file = plan.render_templates(
        config = {
            "deployment.json": struct(
                template = json.encode(deployment_config),
                data = {}
            )
        }
    )
    
    # Build and run the Go deployer
    go_run_result = plan.run_sh(
        run = "cd /app && go build -o deployer src/cmd/deployer/main.go && CONFIG_PATH=/tmp/deployment.json ./deployer",
        files = {
            "/tmp/deployment.json": config_file.files["deployment.json"],
            "/app/config.yaml": config_yaml
        },
        image = "golang:1.21"
    )
    
    return go_run_result
