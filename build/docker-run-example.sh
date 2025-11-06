#!/bin/bash
# Example script for running local-testnet L2 command in Docker
# This provides all required flags for a complete L2 deployment

docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(PWD)/.localnet:/workspace/.localnet \
  compose-network/local-testnet:latest \
  l2 \
  --l1-chain-id 1 \
  --l1-el-url http://host.docker.internal:8545 \
  --l1-cl-url http://host.docker.internal:5052 \
  --compose-network-name testnet \
  --wallet-private-key 0x1234567890123456789012345678901234567890123456789012345678901234 \
  --wallet-address 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1 \
  --coordinator-private-key 0x1234567890123456789012345678901234567890123456789012345678901234 \
  --dispute-network-name testnet \
  --dispute-explorer-url http://explorer.test.com \
  --dispute-explorer-api-url http://explorer.test.com/api \
  --dispute-verifier-address 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1 \
  --dispute-owner-address 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1 \
  --dispute-proposer-address 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1 \
  --dispute-aggregation-vkey 0x1234567890123456789012345678901234567890123456789012345678901234 \
  --dispute-admin-address 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1c

# Notes:
# - Replace wallet addresses and private keys with your actual values
# - Use host.docker.internal for accessing services on your host (macOS/Windows)
# - On Linux, you may need to use --network host or the actual host IP
# - The .localnet directory will be created on your host to persist data
# - All other flags (repositories, images, chain configs) use defaults
