<p align="center"><img src="https://framerusercontent.com/images/9FedKxMYLZKR9fxBCYj90z78.png?scale-down-to=512&width=893&height=363" alt="SSV Network"></p>

<a href="https://discord.com/invite/ssvnetworkofficial"><img src="https://img.shields.io/badge/discord-%23ssvlabs-8A2BE2.svg" alt="Discord" /></a>

## ✨ Introduction

Localnet Control Plane is a CLI tool for managing local L1 and L2 Ethereum test networks. It provides a complete local development environment for testing Ethereum applications with multiple L2 rollups.

## 📋 Prerequisites

- Docker and Docker Compose
- Go 1.25+
- Kurtosis (for L1 network)
- Foundry/Forge (for L2 contract compilation)

## ⚙️  How to Build

```bash
# Clone the repo
git clone https://github.com/compose-network/localnet-control-plane.git

# Navigate
cd localnet-control-plane

# Build the binary
make build
```

The binary will be available at `cmd/localnet/bin/localnet`.

## 🚀 Entry Points

The tool provides three main entry points, each managing a different part of the local network:

### L1 Network (`localnet l1`)
[Architecture](https://github.com/compose-network/local-testnet/blob/main/docs/l1-architecture.png)

Manages the Layer 1 Ethereum test network using Kurtosis. Deploys execution and consensus clients along with SSV nodes via the `github.com/ssvlabs/ssv-mini` package

### L2 Network (`localnet l2`)
[Architecture](https://github.com/compose-network/local-testnet/blob/main/docs/l2-architecture.png)

Manages Layer 2 rollup networks. Orchestrates a three-phase deployment:
1. **Phase 1**: Deploys L1 contracts using op-deployer
2. **Phase 2**: Generates L2 configuration files (genesis, rollup config, secrets)
3. **Phase 3**: Starts L2 runtime services (op-geth, op-node, batcher, proposer) and deploys L2 contracts

Supports multiple L2 chains (rollup-a, rollup-b) with configurable chain IDs and RPC ports.

#### L2 Contract Compilation (`localnet l2 compile`)

Compiles Solidity contracts from the publisher repository. This command generates/updates `contracts.json` in `internal/l2/l2runtime/contracts/compiled/`

**Note:** Requires Foundry/Forge to be installed locally.

### Observability (`localnet observability`)
Manages the observability stack for monitoring and debugging. Deploys a Docker-based infrastructure including Grafana (dashboards), Prometheus (metrics), Loki (logs), Tempo (traces), and Alloy (data collection). Provides real-time visibility into network behavior and performance.

## 🔧 Usage

```bash
# Build and run (based on what's enabled in config.yaml)
make run

# Or run specific components:
make run-l1              # Start L1 network
make run-l2              # Start L2 networks
make run-l2-compile      # Compile L2 contracts from publisher repo
make run-observability   # Start observability stack

# Inspect running services:
make show-l1             # Show Kurtosis enclave
make show-l2             # Show L2 docker containers
make show-observability  # Show observability containers

# Clean up:
make clean               # Clean all components
make clean-l1            # Clean L1 (Kurtosis)
make clean-l2            # Clean L2 (docker containers + generated files)
make clean-observability # Clean observability stack
```

Configuration is managed via `configs/config.yaml`.

## License

Repository is distributed under [GPL-3.0](LICENSE).
