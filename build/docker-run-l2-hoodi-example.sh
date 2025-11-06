#!/bin/bash
# Example script for running local-testnet L2 command in Docker
# This provides all required flags for a complete L2 deployment

docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(PWD):/workspace \
  -w /workspace \
  -e HOST_PROJECT_PATH=$(PWD) \
  compose-network/local-testnet:latest \
  l2 \
  --l1-chain-id 560048 \
  --l1-el-url http://host.docker.internal:8545 \
  --l1-cl-url http://host.docker.internal:5052 \
  --compose-network-name hoodi \
  --wallet-private-key 0x1234567890123456789012345678901234567890123456789012345678901234 \
  --wallet-address 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1 \
  --coordinator-private-key 0x1234567890123456789012345678901234567890123456789012345678901234 \
  --deployment-target live \
  --genesis-balance-wei 100000000000000000000 \
  --op-geth-url https://github.com/compose-network/op-geth.git \
  --op-geth-branch stage \
  --publisher-url https://github.com/compose-network/publisher.git \
  --publisher-branch stage \
  --compose-contracts-url https://github.com/compose-network/contracts.git \
  --compose-contracts-branch develop \
  --op-deployer-tag v0.4.5 \
  --op-node-tag v1.13.6 \
  --op-proposer-tag v1.10.0 \
  --op-batcher-tag v1.14.0 \
  --rollup-a-id 77777 \
  --rollup-a-rpc-port 18545 \
  --rollup-b-id 88888 \
  --rollup-b-rpc-port 28545 \
  --dispute-network-name hoodi \
  --dispute-explorer-url https://hoodi.etherscan.io \
  --dispute-explorer-api-url https://api-hoodi.etherscan.io/api \
  --dispute-verifier-address 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1 \
  --dispute-owner-address 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1 \
  --dispute-proposer-address 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1 \
  --dispute-aggregation-vkey 0x1234567890123456789012345678901234567890123456789012345678901234 \
  --dispute-admin-address 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1 \
  --dispute-starting-superblock-number 0

# Notes:
# - Replace wallet addresses and private keys with your actual values
# - Use host.docker.internal for accessing services on your host (macOS/Windows)
# - On Linux, you may need to use --network host or the actual host IP
# - The entire current directory is mounted to /workspace in the container
# - HOST_PROJECT_PATH is required for Docker-in-Docker volume mounts
# - Repository URLs use HTTPS (public repos)
# - Adjust branch names (stage, develop) and version tags as needed
# - Chain IDs 77777 and 88888 are example values for rollup-a and rollup-b
# - RPC ports 18545 and 28545 avoid conflicts with standard 8545
