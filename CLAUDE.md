# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Localnet Control Plane is a CLI tool for managing local L1 and L2 Ethereum test networks. It orchestrates:
- **L1 Network**: Uses Kurtosis to deploy Ethereum execution/consensus clients and SSV nodes via the `github.com/ssvlabs/ssv-mini` package
- **Observability Stack**: Manages a Docker-based monitoring infrastructure (Grafana, Prometheus, Loki, Tempo, Alloy)

## Build and Run Commands

### Basic Operations
- `make build` - Build the binary to `cmd/localnet/bin/localnet`
- `make run` - Build and run (default make target)
- `make test` - Run all tests
- `make lint` - Run golangci-lint

### L1 Network Management
- `make run-l1` - Run the L1 command (shows help)
- `make show-l1` - Inspect the Kurtosis enclave named `localnet`
- `make clean-l1` - Clean up Kurtosis enclaves with `kurtosis clean -a`
- `make restart-ssv-nodes` - Restart SSV node services (default count: 4, override with `SSV_NODE_COUNT=N`)

### Observability Management
- `make show-observability` - Show Docker containers with label `stack=localnet-observability`
- `make clean-observability` - Remove all observability containers
- `make clean` - Clean both observability and Kurtosis resources

### Running the Application
The built binary requires a `config.yaml` file in the same directory. The Makefile copies `configs/config.yaml` to the binary directory during build.

## Architecture

### Command Structure
- **Entry point**: `cmd/localnet/main.go` - Cobra-based CLI that loads `config.yaml` using Viper
- **L1 command**: `internal/l1/cmd.go` - Conditionally starts localnet and/or observability based on config flags

### Configuration System
- Config defined in `configs/config.go` and loaded into global `configs.Values`
- Config structure:
  ```yaml
  l1:
    enabled: bool
    observability:
      enabled: bool
  ```

### L1 Network (`internal/l1/localnet/`)
- Uses Kurtosis SDK to create an enclave and run the `github.com/ssvlabs/ssv-mini` Starlark package
- Configuration embedded from `params.yaml` which defines:
  - Network participants (EL: Geth, CL: Lighthouse)
  - MEV configuration (relay, builder, boost)
  - Network parameters (fork epochs, validator counts)
  - SSV and Anchor node counts

### Observability Stack (`internal/l1/observability/`)
- Each service (Alloy, Grafana, Loki, Prometheus, Tempo) has its own package under `observability/`
- Services are Docker containers started via Docker SDK
- All containers:
  - Share the `stack=localnet-observability` label
  - Connect to a shared Docker network managed by `observability/shared/`
  - Have configs in `configs/{service-name}/`
- Port bindings:
  - Grafana: 3000 (anonymous admin access enabled)

### Package Organization
- `internal/l1/` - L1-specific functionality (localnet and observability)
- `internal/logger/` - Logging utilities
- `configs/` - Application config + service-specific configs for observability components

## Key Implementation Details

### Kurtosis Integration
- Enclave name is hardcoded as `"localnet"`
- Starlark package: `github.com/ssvlabs/ssv-mini`
- Parameters passed as serialized YAML from embedded `params.yaml`
- Stream-based output processing with structured logging for progress, instructions, warnings, and errors
- Errors from Kurtosis package execution are returned and will stop the application

### Docker Service Pattern
Each observability service follows the same pattern:
1. Pull Docker image
2. Create container with service-specific config
3. Attach to shared network
4. Start container
5. Log progress and return errors up the chain

### Error Handling
- All errors are wrapped with context using `errors.Join(err, errors.New("context message"))`
- Kurtosis errors are surfaced through the output stream and logged before returning
