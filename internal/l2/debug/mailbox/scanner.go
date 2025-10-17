package mailbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/compose-network/localnet-control-plane/internal/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// Scanner scans blocks for mailbox activity
type Scanner struct {
	logger *slog.Logger
}

// NewScanner creates a new mailbox scanner
func NewScanner() *Scanner {
	return &Scanner{
		logger: logger.Named("mailbox_scanner"),
	}
}

// ScanBlocks scans recent blocks for mailbox activity
func (s *Scanner) ScanBlocks(ctx context.Context, rpcURL string, mailboxAddr common.Address, blockWindow int, sessionFilter *uint64) ([]*MailboxCall, error) {
	s.logger.With("rpc_url", rpcURL).With("blocks", blockWindow).Info("scanning blocks for mailbox activity")

	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC: %w", err)
	}
	defer client.Close()

	rpcClient := client.Client()

	latest, err := client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	first := uint64(0)
	if latest >= uint64(blockWindow) {
		first = latest - uint64(blockWindow) + 1
	}

	var calls []*MailboxCall

	for num := latest; num >= first && num <= latest; num-- {
		block, err := client.BlockByNumber(ctx, big.NewInt(int64(num)))
		if err != nil {
			s.logger.With("block", num).With("err", err).Warn("failed to get block")
			continue
		}

		for _, tx := range block.Transactions() {
			txHash := tx.Hash().Hex()

			// Call debug_traceTransaction
			trace, err := s.traceTransaction(ctx, rpcClient, txHash)
			if err != nil {
				continue // Skip transactions that fail to trace
			}

			// Extract mailbox calls from trace
			mailboxCalls := s.extractMailboxCalls(trace, mailboxAddr, num, txHash)
			calls = append(calls, mailboxCalls...)
		}
	}

	// Filter by session if requested
	if sessionFilter != nil {
		filtered := make([]*MailboxCall, 0)
		for _, call := range calls {
			if call.SessionID != nil && *call.SessionID == *sessionFilter {
				filtered = append(filtered, call)
			}
		}
		calls = filtered
	}

	s.logger.With("calls_found", len(calls)).Info("scan completed")

	return calls, nil
}

// traceTransaction calls debug_traceTransaction with callTracer
func (s *Scanner) traceTransaction(ctx context.Context, client *rpc.Client, txHash string) (*CallTrace, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var trace CallTrace
	err := client.CallContext(ctx, &trace, "debug_traceTransaction", txHash, map[string]string{
		"tracer": "callTracer",
	})
	if err != nil {
		return nil, fmt.Errorf("debug_traceTransaction failed: %w", err)
	}

	return &trace, nil
}

// extractMailboxCalls recursively extracts mailbox calls from trace
func (s *Scanner) extractMailboxCalls(trace *CallTrace, mailboxAddr common.Address, blockNum uint64, txHash string) []*MailboxCall {
	var calls []*MailboxCall

	// Check if this call is to the mailbox
	if strings.EqualFold(trace.To, mailboxAddr.Hex()) {
		call, err := DecodeMailboxCall(trace.Input)
		if err == nil {
			call.BlockNumber = blockNum
			call.TxHash = txHash
			call.From = trace.From
			call.To = trace.To
			call.Value = trace.Value
			calls = append(calls, call)
		}
	}

	// Recursively check subcalls
	for _, subcall := range trace.Calls {
		subcalls := s.extractMailboxCalls(&subcall, mailboxAddr, blockNum, txHash)
		calls = append(calls, subcalls...)
	}

	return calls
}

// FormatCall formats a mailbox call for display
func FormatCall(call *MailboxCall) string {
	sessionID := "?"
	if call.SessionID != nil {
		sessionID = fmt.Sprintf("%d", *call.SessionID)
	}

	txShort := call.TxHash
	if len(txShort) > 10 {
		txShort = txShort[:10] + "…"
	}

	label := call.Label
	if label == "" {
		label = "?"
	}

	data := call.Data
	if len(data) > 20 {
		data = data[:20] + "…"
	}

	return fmt.Sprintf("block %d tx %s fn=%s session=%s label=%s data=%s",
		call.BlockNumber, txShort, call.Function, sessionID, label, data)
}

// DebugTraceResponse is used for raw RPC unmarshaling
type DebugTraceResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
