# Docker Usage

## Building the Image

```bash
# Build with default tag
make docker-build
```

## Running the Container (L2 Hoodi)

### Quick Start

The easiest way to run the container is using the provided example script:

```bash
# 1. Copy the example script
cp build/docker-run-l2-hoodi-example.sh docker-run-l2-hoodi.sh

# 2. Edit the script and replace required values

# 3. Run the script
chmod +x docker-run-l2-hoodi.sh
./docker-run-l2-hoodi.sh
```

**Required values to replace:**
- `--l1-el-url` - Your L1 execution client RPC URL
- `--l1-cl-url` - Your L1 consensus client REST URL
- `--wallet-private-key` - Private key for deployment transactions
- `--wallet-address` - Address corresponding to the private key
- `--coordinator-private-key` - Private key for coordinator operations
- `--dispute-verifier-address` - Address of the dispute verifier contract
- `--dispute-owner-address` - Owner address for dispute contracts (can be same as wallet-address)
- `--dispute-proposer-address` - Proposer address for dispute game (can be same as wallet-address)
- `--dispute-aggregation-vkey` - Verification key for aggregation
- `--dispute-admin-address` - Admin address for dispute contracts (can be same as wallet-address)
- `--dispute-explorer-url` - Block explorer URL (e.g., Etherscan)
- `--dispute-explorer-api-url` - Block explorer API URL

**Note:** For testing/development, `--dispute-owner-address`, `--dispute-proposer-address`, and `--dispute-admin-address` can all be set to the same value as `--wallet-address`.

**Optional values to customize:**
- Repository URLs and branches (if using forks)
- OP Stack component versions (deployer, node, proposer, batcher tags)
- Chain IDs and RPC ports for rollups
- Genesis balance


## Installed Tools

The image includes:
- **localnet** binary at `/usr/local/bin/localnet`
- **Docker CLI** with compose plugin
- **Git** for repository cloning
- **Foundry** (forge, cast, anvil) for Solidity compilation
- **just** - Command runner for contract setup scripts
- **jq** - JSON processor for contract deployment scripts
