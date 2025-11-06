# L2 Network Command

The `localnet l2` command manages Layer 2 rollup networks built on the OP Stack.

## Architecture

![L2 Architecture](../../docs/l2-architecture.png)

## What It Does

The L2 command orchestrates a complete rollup deployment in three phases:

### Phase 1: L1 Contract Deployment
Deploys OP Stack contracts to L1 using `op-deployer`:
- System Config
- L1 Standard Bridge
- L1 Cross Domain Messenger
- OptimismPortal
- DisputeGameFactory
- And other core contracts

### Phase 2: Configuration Generation
Generates configuration files for each L2 chain:
- `genesis.json` - Initial blockchain state
- `rollup.json` - Rollup configuration
- `jwt-secret.txt` - Authentication between services

### Phase 3: Runtime Deployment
Starts L2 services using Docker Compose:
- **op-geth**: Execution client for each rollup
- **op-node**: Consensus/derivation client
- **op-batcher**: Batches transactions to L1
- **op-proposer**: Proposes output roots to L1
- **Publisher**: Publishes superblocks to L1

Deploys Compose-specific contracts to L2:
- Dispute settlement contracts
- Verification contracts

## Prerequisites

- **L1 Network**: Must have L1 running (`make run-l1`) or specify external L1 URLs
- **Foundry/Forge**: For Solidity compilation
- **just**: Command runner for contract scripts
- **jq**: JSON processor for deployment scripts
- **Docker**: For running L2 services

## Configuration

All L2 settings are configured in `configs/config.yaml`. See [example config](../../configs/config.yaml) for all available options.

**Required settings:**
- L1 connection (chain ID, EL URL, CL URL)
- Wallet credentials (private key, address)
- Coordinator credentials
- Compose network name
- Dispute game settings (addresses, vkeys, explorer URLs)

## Usage

### Running L2 Networks

```bash
# Start L2 deployment (all phases)
make run-l2

# Or run directly
./cmd/localnet/bin/localnet l2

# Show running services
make show-l2

# Clean up
make clean-l2
```

### Compiling Contracts

```bash
# Compile contracts from compose-contracts repository
make run-l2-compile

# Or run directly
./cmd/localnet/bin/localnet l2 compile
```

This generates `contracts.json` in `.localnet/compiled-contracts/`. To embed in binary, copy to `internal/l2/l2runtime/contracts/compiled/` and commit.

### Docker Usage

For running in Docker, see the [Docker documentation](../../build/DOCKER.md) and example scripts:
- `build/docker-run-example.sh` - Generic template
- `build/docker-run-hoodi.sh` - Hoodi testnet example

## Multi-Chain Support

The L2 command supports deploying multiple rollup chains simultaneously:
- **Rollup A**: Default chain ID 77777, RPC port 18545
- **Rollup B**: Default chain ID 88888, RPC port 28545

Configure each chain independently via flags or config file.

## Implementation Details

### Directory Structure
- `l2config/` - Phase 2 configuration generation
- `l2deployer/` - Phase 1 L1 contract deployment
- `l2runtime/` - Phase 3 runtime orchestration
  - `contracts/` - L2 contract deployment
  - `publisher/` - Publisher configuration
- `infra/docker/` - Docker Compose templates
- `service.go` - Service management (restart, etc.)

### Generated Files
All generated files are stored in `.localnet/`:
- `genesis/chain-{id}/` - Genesis and rollup configs
- `services/op-geth-{chain}/` - Geth data directories
- `services/compose-contracts/` - Contract repositories
- `registry/` - Publisher registry configuration

## Troubleshooting

**Issue:** L1 connection failed
**Solution:** Ensure L1 is running and URLs are correct. Use `host.docker.internal` when running in Docker (macOS/Windows).

**Issue:** Docker Compose errors
**Solution:** Ensure Docker daemon is running and socket is accessible.

**Issue:** Contract compilation failed
**Solution:** Verify Foundry/Forge is installed: `forge --version`

**Issue:** Publisher crashes with "empty compose network name"
**Solution:** Ensure `--compose-network-name` flag is set or configured in `config.yaml`

For more details, see the [main README](../../README.md) or [Docker documentation](../../build/DOCKER.md).
