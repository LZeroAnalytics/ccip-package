GO_SERVICE_NAME = "ccip-deployer"
GO_IMAGE = "golang:1.24.3-alpine"  # Base Go image

def run(plan, config_struct):
    """Deploy CCIP using proper render_templates with YAML template"""
    
    # Upload source code directly and config yaml
    go_source = plan.upload_files(".")
    config_artifact = struct_to_yaml_template(config_struct)
    
    # Create service with base Go image
    go_service = plan.add_service(
        name = GO_SERVICE_NAME,
        config = ServiceConfig(
            image = GO_IMAGE,
            entrypoint = ["tail", "-f", "/dev/null"],  # Keep container running
            files = {
                "/app": go_source,  # Mount source code
                "/config": config_artifact,  # Mount generated config
            },
        )
    )
    
    # Build the Go application inside the container
    plan.exec(
        service_name = GO_SERVICE_NAME,
        recipe = ExecRecipe(
            command = ["sh", "-c", "cd /app && go build -o /usr/local/bin/ccip-deployer ."]
        )
    )
    
    # Deploy with config file
    result = plan.exec(
        service_name = GO_SERVICE_NAME,
        recipe = ExecRecipe(
            command = ["/usr/local/bin/ccip-deployer", "deploy", "/config/deploy.yaml"]
        )
    )
    
    return result

def struct_to_yaml_template(config_struct):
        yaml_template = """home_chain:
  chain_id: {{.home_chain.chain_id}}
  name: "{{.home_chain.name}}"
  rpc_url: "{{.home_chain.rpc_url}}"
  private_key: "{{.home_chain.private_key}}"

feed_chain:
  chain_id: {{.feed_chain.chain_id}}
  name: "{{.feed_chain.name}}"
  rpc_url: "{{.feed_chain.rpc_url}}"
  private_key: "{{.feed_chain.private_key}}"

chains:
{{- range .chains}}
  - chain_id: {{.chain_id}}
    name: "{{.name}}"
    rpc_url: "{{.rpc_url}}"
    private_key: "{{.private_key}}"
{{- end}}

deployment:
  rmn_enabled: {{.deployment.rmn_enabled}}

existing_contracts:
  link_token: "{{.existing_contracts.link_token}}"
  link_eth_feed: "{{.existing_contracts.link_eth_feed}}"
  eth_usd_feed: "{{.existing_contracts.eth_usd_feed}}"

chainlink:
  nodes:
{{- range .chainlink.nodes}}
    - name: "{{.name}}"
      chainlink_config:
        url: "{{.chainlink_config.url}}"
        email: "{{.chainlink_config.email}}"
        password: "{{.chainlink_config.password}}"
{{- if .chainlink_config.internal_ip}}
        internal_ip: "{{.chainlink_config.internal_ip}}"
{{- end}}
{{- if .chainlink_config.headers}}
        headers:
{{- range $key, $value := .chainlink_config.headers}}
          "{{$key}}": "{{$value}}"
{{- end}}
{{- end}}
      p2p_port: "{{.p2p_port}}"
      is_bootstrap: {{.is_bootstrap}}
      admin_addr: "{{.admin_addr}}"
      multi_addr: "{{.multi_addr}}"
      container_name: "{{.container_name}}"
      labels:
        type: "{{.labels.type}}"
        environment: "{{.labels.environment}}"
        product: "{{.labels.product}}"
{{- end}}"""
    
    # Create config file using proper render_templates
    config_artifact = plan.render_templates(
        config = {
            "/deploy.yaml": struct(
                template = yaml_template,
                data = config_struct
            )
        },
        name = "ccip-deployer-config",
        description = "CCIP deployer configuration file"
    )

    return config_artifact

def run_command(plan, args):
    result = plan.exec(
        service_name = GO_SERVICE_NAME,
        recipe = ExecRecipe(
            command = ["/usr/local/bin/ccip-deployer"] + args
        )
    )
    return result

