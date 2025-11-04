# Docker Usage

## Building the Image

```bash
# Build with default tag
make docker-build

# Build with custom tag
make docker-build DOCKER_IMAGE_TAG=v1.0.0

# Build with custom name and tag
make docker-build DOCKER_IMAGE_NAME=myorg/localnet DOCKER_IMAGE_TAG=v1.0.0
```

## Running the Container

### Prerequisites

The container requires access to the **host Docker daemon** to run docker compose commands for L2 services.

**Important:** Mount the host Docker socket:
```bash
-v /var/run/docker.sock:/var/run/docker.sock
```

### Basic Usage

```bash
# Show help
make docker-run

# Run L2 command with flags
make docker-run-l2 ARGS="--l1-chain-id 1 --l1-el-url http://host.docker.internal:8545"

# Or use docker directly
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(PWD):/workspace \
  -w /workspace \
  compose-network/local-testnet:latest \
  l2 --help
```

### Accessing Host Services from Container

When running in Docker and connecting to services on your host machine (e.g., L1 node), use:

**macOS/Windows:**
```bash
--l1-el-url http://host.docker.internal:8545
```

**Linux:**
```bash
--network host
# Or add host.docker.internal to extra_hosts
```

## Complete Example

```bash
# Build the image
make docker-build

# Run L2 deployment
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(PWD):/workspace \
  -w /workspace \
  compose-network/local-testnet:latest \
  l2 \
  --l1-chain-id 1 \
  --l1-el-url http://host.docker.internal:8545 \
  --l1-cl-url http://host.docker.internal:5052 \
  --private-key 0x... \
  --wallet-address 0x... \
  --coordinator-private-key 0x... \
  --dispute-network-name testnet \
  --dispute-explorer-url http://explorer.test \
  --dispute-explorer-api-url http://explorer.test/api \
  --dispute-verifier-address 0x... \
  --dispute-owner-address 0x... \
  --dispute-proposer-address 0x... \
  --dispute-aggregation-vkey 0x... \
  --dispute-admin-address 0x...
```

## Publishing to Registry

```bash
# Build for your registry
make docker-build DOCKER_IMAGE_NAME=ghcr.io/compose-network/local-testnet DOCKER_IMAGE_TAG=v1.0.0

# Push to registry
docker push ghcr.io/compose-network/local-testnet:v1.0.0
```

## Volumes

- **Docker socket:** `/var/run/docker.sock` - Required for docker compose operations
- **Working directory:** `/workspace` - Mount your project directory here so nested containers can access `.localnet/` subdirectories

## Installed Tools

The image includes:
- **localnet** binary at `/usr/local/bin/localnet`
- **Docker CLI** with compose plugin
- **Git** for repository cloning
- **Foundry** (forge, cast, anvil) for Solidity compilation
- **just** - Command runner for contract setup scripts
- **jq** - JSON processor for contract deployment scripts

## Troubleshooting

**Issue:** `Cannot connect to the Docker daemon`
- **Solution:** Ensure you mount the Docker socket: `-v /var/run/docker.sock:/var/run/docker.sock`

**Issue:** `Cannot connect to L1 RPC`
- **Solution:** Use `host.docker.internal` instead of `localhost` (macOS/Windows) or `--network host` (Linux)

**Issue:** Permission denied on Docker socket
- **Solution:** Ensure your user has Docker permissions or run with appropriate privileges
