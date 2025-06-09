chainlink_pkg = import_module("./src/chainlink-node-package/main.star")
ccip_deployer_pkg = import_module("./src/ccip-deployer/main.star")

def run(plan, args = {}):
    config = args
   
    if not config.get("chainlink") or not config.get("chainlink", {}).get("nodes") or len(config.get("chainlink", {}).get("nodes", [])) < 6:
        result = start_don(plan, config)
        config["chainlink"] = {
            "nodes": result.nodes
        }
    
    start_ccip(plan, config)

    return struct(
        contracts_addresses = "TODO: return actual addresses",
        chainlink_nodes = config["chainlink"]["nodes"]
    )

def start_don(plan, config):
    # Create database and configs for all chainlink nodes in parallel
    chainlink_node_configs = []
    for i in range(6):
        chainlink_node_configs.append({
            "node_name": "chainlink-node-" + str(i),
        })
    
    chains = [{
        "rpc": config["home_chain"]["rpc_url"],
        "ws": replace_http_with_ws(config["home_chain"]["rpc_url"]),
        "chain_id": config["home_chain"]["chain_id"]
    }]
    if config["feed_chain"]["chain_id"] != config["home_chain"]["chain_id"]:
        chains.append({
            "rpc": config["feed_chain"]["rpc_url"],
            "ws": replace_http_with_ws(config["feed_chain"]["rpc_url"]),
            "chain_id": config["feed_chain"]["chain_id"]
        })
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

    return struct(
        nodes = create_node_configs_for_ccip_deployer(plan, result),
        all_nodes = result.services
    )

def start_ccip(plan, config):
    job_distributor = start_job_distributor(plan)

    ccip_deployer_pkg.run(plan, config, jd_url = job_distributor.ip_address + ":" + str(job_distributor.ports["grpc"].number))

def start_job_distributor(plan):
    job_distributor = plan.add_service(
        name = "job-distributor",
        config = ServiceConfig(
            image = "fravlaca/job-distributor:latest",
            ports = {"grpc": PortSpec(50051, "TCP")},
            env_vars = {"JD_PORT": str(50051)}
        ),
    )
    return job_distributor

def create_node_configs_for_ccip_deployer(plan, result):
    node_configs = []

    i = 0
    for result_config in result.nodes_configs:
        node_service = result.services[result_config.node_name]
        p2p_peer_id = chainlink_pkg.node_utils.get_p2p_peer_id(plan, result_config.node_name)
        multi_addr = p2p_peer_id+"@"+node_service.ip_address+":"+str(node_service.ports["p2p"].number)
        
        if i == 0: 
            node_type = "bootstrap" 
        else: 
            node_type = "plugin"

        node_configs.append(struct(
            name = result_config.node_name,
            chainlink_config = struct(
                url = "http://" + node_service.ip_address + ":" + str(node_service.ports["http"].number),
                email = result_config.api_user,
                password = result_config.api_password,
            ),
            p2p_port = node_service.ports["p2p"].number,
            is_bootstrap = i == 0,
            admin_addr = chainlink_pkg.node_utils.get_eth_key(plan, result_config.node_name),
            multi_addr = multi_addr,
            container_name = node_service.hostname,
            labels = struct(
                type = node_type,
                environment = "devnet",
                product = "ccip"
            )
        ))
        i += 1

    return node_configs

def replace_http_with_ws(rpc_url):
    if rpc_url.startswith("https://"):
        return rpc_url.replace("https://", "wss://")
    elif rpc_url.startswith("http://"):
        return rpc_url.replace("http://", "ws://")
    else:
        return "ws://" + rpc_url