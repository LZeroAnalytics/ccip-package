def run(plan, args):
    # Parse configuration from args with defaults
    chains = args.get("chains", {})
    private_key = args.get("private_key", "")
    home_chain = args.get("home_chain", "")
    
    # Create preexisting contracts with hardcoded Link token and price feed addresses
    # These are the addresses provided in config.yaml
    preexisting_contracts = {}
    
    # Add Link token and price feed for each chain
    for chain_name, chain_config in chains.items():
        chain_id = str(chain_config.get("chain_id", ""))
        if chain_id:
            # Add Link token contract
            preexisting_contracts["link_token_" + chain_name] = {
                "address": "0x514910771AF9Ca656af840dff83E8264EcF986CA",
                "chain": chain_name,
                "type": "LinkToken"
            }
            
            # Add Link native token feed contract
            preexisting_contracts["price_feed_" + chain_name] = {
                "address": "0xdc530d9457755926550b59e8eccdae7624181557",
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
    
    # Convert deployment config to JSON
    deployment_json = json.encode(deployment_config)
    
    # Use echo to create the deployment.json file in the container
    go_run_result = plan.run_sh(
        run = "echo '" + deployment_json + "' > /tmp/deployment.json && cd /app && go build -o deployer src/cmd/deployer/main.go && CONFIG_PATH=/tmp/deployment.json ./deployer",
        image = "golang:1.21"
    )
    
    return go_run_result
