<p align="center"><img src="https://framerusercontent.com/images/9FedKxMYLZKR9fxBCYj90z78.png?scale-down-to=512&width=893&height=363" alt="SSV Network"></p>

<a href="https://discord.com/invite/ssvnetworkofficial"><img src="https://img.shields.io/badge/discord-%23ssvlabs-8A2BE2.svg" alt="Discord" /></a>

## ✨ Introduction

Local testnet is a CLI tool for managing local L1 and L2 Ethereum test networks. It provides a complete local development environment for testing Ethereum applications with multiple L2 rollups.

## 📋 Prerequisites

- Docker and Docker Compose
- Go 1.25+
- Kurtosis (for L1 network)
- Foundry/Forge (for L2 commands)
- [just](https://github.com/casey/just) (for L2 commands)
- jq (for L2 commands)

## ⚙️  How to Build

```bash
# Clone the repo
git clone https://github.com/compose-network/local-testnet.git

# Navigate
cd local-testnet

# Build the binary
make build
```

The binary will be available at `cmd/localnet/bin/localnet`.

## 🚀 Commands

The tool provides three main commands, each managing a different part of the local network:

### L1 Network (`localnet l1`)
Manages the Layer 1 Ethereum test network using Kurtosis. Deploys execution and consensus clients along with SSV nodes.

**📖 [Read L1 Documentation](internal/l1/README.md)**

### L2 Network (`localnet l2`)
Manages Layer 2 rollup networks. Orchestrates a three-phase deployment process for multiple OP Stack rollups.

**📖 [Read L2 Documentation](internal/l2/README.md)**

### Observability (`localnet observability`)
Manages the observability stack for monitoring and debugging. Deploys Grafana, Prometheus, Loki, Tempo, and Alloy.

**📖 [Read Observability Documentation](internal/observability/README.md)**

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

## 🐳 Docker Usage

The tool can be run in Docker, which is useful for CI/CD or environments where dependencies are difficult to install.

**📖 [Read Docker Documentation](build/DOCKER.md)**

## License

Repository is distributed under [GPL-3.0](LICENSE).
