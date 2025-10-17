# Debug-Bridge Command

Go implementation of the Python `debug-bridge` diagnostics tool for Compose rollup networks.

## Overview

The `debug-bridge` command provides comprehensive diagnostics for cross-rollup bridge activity, including:
- Mailbox transaction scanning and decoding
- ETH and ERC20 token balance checks
- Shared publisher statistics
- Docker container logs

## Architecture

```
internal/l2/debug/
├── cmd.go              # Cobra command definition
├── service.go          # Main orchestrator
├── mailbox/
│   ├── types.go        # Data structures
│   ├── decoder.go      # Calldata decoding for mailbox functions
│   └── scanner.go      # Block scanning + debug_traceTransaction
├── balance/
│   └── checker.go      # ETH + ERC20 balance queries
├── publisher/
│   └── stats.go        # HTTP client for publisher stats
└── logs/
    └── collector.go    # Docker logs filtering
```

## Usage

### Debug Mode (Full Diagnostics)
```bash
./cmd/localnet/bin/localnet l2 debug-bridge --mode=debug --blocks=20
./cmd/localnet/bin/localnet l2 debug-bridge --mode=debug --session=12345 --blocks=50
```

Outputs:
- Mailbox activity for last N blocks (write, putInbox, read calls)
- Publisher stats JSON
- ETH and token balances on both rollups
- Filtered Docker logs from op-geth containers

### Check Mode (Quick Health)
```bash
./cmd/localnet/bin/localnet l2 debug-bridge --mode=check
```

Outputs:
- ETH and token balances
- Publisher connection status
- Latest block numbers

## Configuration

**Required configuration** in `configs/config.yaml`:
```yaml
l2:
  debug-bridge:
    publisher-stats-url: http://127.0.0.1:18081/stats  # Required
    default-blocks: 12                                  # Required, must be > 0
    default-log-window: 120s                            # Required
```

The command will validate configuration at startup and error if any required fields are missing or invalid.

## Implementation Details

### Mailbox Decoding
- Manual ABI decoding using 4-byte selectors (matching Python implementation)
- Supports 9 mailbox function variants (write, putInbox, read, clear)
- Handles dynamic bytes offsets per Solidity ABI encoding rules

### Block Scanning
- Uses `debug_traceTransaction` with `callTracer` to get full call traces
- Recursively extracts mailbox calls from transaction traces
- Filters by session ID when specified

### Balance Checking
- ETH: `client.BalanceAt()`
- ERC20: Manual `eth_call` encoding for `balanceOf(address)`

### Docker Logs
- Executes `docker logs --since=<window>` via `os/exec`
- Filters for keywords: "mailbox", "SSV", "Send CIRC", "putInbox", "Tracer captured"
- Returns last 20 matching lines

## Key Differences from Python Version

| Feature | Python | Go |
|---------|--------|---|
| RPC Client | urllib + manual JSON | ethclient + go-ethereum |
| ABI Decoding | Manual byte chunking | Manual byte chunking (same logic) |
| Config | .env files | Viper YAML |
| Logs | subprocess | os/exec |
| Output | Print statements | fmt.Printf |

## Testing

1. Start local-testnet-legacy Python stack:
   ```bash
   cd ../../../  # Back to local-testnet-legacy root
   ./compose up --fresh
   ```

2. Run debug-bridge:
   ```bash
   cd local-testnet
   ./cmd/localnet/bin/localnet l2 debug-bridge --mode=check
   ```

3. Trigger bridge activity:
   ```bash
   cd ../
   ./toolkit.sh bridge-once --mint-if-needed
   ```

4. Scan for mailbox calls:
   ```bash
   cd local-testnet
   ./cmd/localnet/bin/localnet l2 debug-bridge --mode=debug --blocks=20
   ```

## Future Enhancements

- [ ] Add proper ABI-based decoding using compiled Mailbox contract
- [ ] Support filtering by call type (write/read/putInbox)
- [ ] Add JSON output mode for machine parsing
- [ ] Integrate publisher stats URL from config (currently hardcoded)
- [ ] Add block number retrieval in check mode
- [ ] Support custom RPC URLs via flags
