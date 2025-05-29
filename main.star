def run(plan, args):
    # This is a simplified version of the CCIP deployer
    # It demonstrates how to use preexisting contracts from config.yaml
    
    # Print a message about the preexisting contracts
    result = plan.run_sh(
        run = "echo 'CCIP Deployer Wrapper' && echo 'Using preexisting contracts:' && echo '- Link Token: 0x514910771AF9Ca656af840dff83E8264EcF986CA' && echo '- Price Feed: 0xdc530d9457755926550b59e8eccdae7624181557'",
        image = "golang:1.21"
    )
    
    return result
