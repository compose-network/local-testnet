package balance

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/compose-network/localnet-control-plane/internal/logger"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Checker checks ETH and token balances
type Checker struct {
	logger *slog.Logger
}

// NewChecker creates a new balance checker
func NewChecker() *Checker {
	return &Checker{
		logger: logger.Named("balance_checker"),
	}
}

// BalanceInfo contains balance information
type BalanceInfo struct {
	RollupName string
	Address    string
	ETHBalance *big.Int
	ETHError   error
}

// TokenBalanceInfo contains token balance information
type TokenBalanceInfo struct {
	RollupName    string
	TokenAddress  string
	WalletAddress string
	Balance       *big.Int
	Error         error
}

// GetETHBalance gets ETH balance for an address
func (c *Checker) GetETHBalance(ctx context.Context, rpcURL string, address common.Address) (*big.Int, error) {
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC: %w", err)
	}
	defer client.Close()

	balance, err := client.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

// GetTokenBalance gets ERC20 token balance for an address
func (c *Checker) GetTokenBalance(ctx context.Context, rpcURL string, tokenAddr, accountAddr common.Address) (*big.Int, error) {
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC: %w", err)
	}
	defer client.Close()

	// Encode balanceOf(address) call
	methodID := crypto.Keccak256([]byte("balanceOf(address)"))[:4]
	paddedAddress := common.LeftPadBytes(accountAddr.Bytes(), 32)
	data := append(methodID, paddedAddress...)

	msg := ethereum.CallMsg{
		To:   &tokenAddr,
		Data: data,
	}

	result, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	balance := new(big.Int).SetBytes(result)
	return balance, nil
}

// FormatETHBalance formats ETH balance for display
func FormatETHBalance(rollupName string, address string, balance *big.Int, err error) string {
	if err != nil {
		return fmt.Sprintf("%s: balance query failed (%v)", rollupName, err)
	}

	eth := new(big.Float).Quo(
		new(big.Float).SetInt(balance),
		new(big.Float).SetInt(big.NewInt(1e18)),
	)

	return fmt.Sprintf("%s: balance %.4f ETH (%s wei)", rollupName, eth, balance.String())
}

// FormatTokenBalance formats token balance for display
func FormatTokenBalance(rollupName string, tokenAddr string, walletAddr string, balance *big.Int, err error) string {
	if err != nil {
		return fmt.Sprintf("%s: token balance query failed for %s (%v)", rollupName, tokenAddr, err)
	}

	if balance == nil {
		return fmt.Sprintf("%s: token balance unavailable", rollupName)
	}

	tokens := new(big.Float).Quo(
		new(big.Float).SetInt(balance),
		new(big.Float).SetInt(big.NewInt(1e18)),
	)

	return fmt.Sprintf("%s: token balance %.4f (%s raw) [%s]", rollupName, tokens, balance.String(), tokenAddr)
}
