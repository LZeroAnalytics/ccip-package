def run(plan, args):
    # Parse configuration from args
    chains = args.get("chains", {})
    private_key = args.get("private_key", "")
    home_chain = args.get("home_chain", "")
    preexisting_contracts = args.get("preexisting_contracts", {})
    
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
            "/tmp/deployment.json": config_file.files["deployment.json"]
        },
        image = "golang:1.21"
    )
    
    return go_run_result
