def run(plan, args):
    # Parse configuration from args with defaults
    chains = args.get("chains", {})
    private_key = args.get("private_key", "")
    home_chain = args.get("home_chain", "")
    
    # Create preexisting contracts with hardcoded Link token and price feed addresses
    # These are the addresses provided in config.yaml
    preexisting_contracts = {}
    
    # Hardcoded addresses from config.yaml
    link_token_address = "0x514910771AF9Ca656af840dff83E8264EcF986CA"
    link_native_token_feed_address = "0xdc530d9457755926550b59e8eccdae7624181557"
    
    # Add Link token and price feed for each chain
    for chain_name, chain_config in chains.items():
        # Add Link token contract
        preexisting_contracts["link_token_" + chain_name] = {
            "address": link_token_address,
            "chain": chain_name,
            "type": "LinkToken"
        }
        
        # Add Link native token feed contract
        preexisting_contracts["price_feed_" + chain_name] = {
            "address": link_native_token_feed_address,
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
    
    # Create a files artifact with the Go code
    code_artifact = plan.upload_files(
        src = ".",
        name = "ccip-package-code"
    )
    
    # Run the Go code with the deployment configuration
    result = plan.run_sh(
        run = "mkdir -p /app && cp -r /files/* /app/ && echo '" + deployment_json + "' > /tmp/deployment.json && cd /app/src && go mod tidy && go build -o /tmp/deployer ./cmd/deployer/main.go && CONFIG_PATH=/tmp/deployment.json /tmp/deployer",
        image = "golang:1.24",
        files = {
            "/files": code_artifact
        }
    )
    
    # Return the result
    return result
