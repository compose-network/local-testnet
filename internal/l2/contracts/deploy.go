package contracts

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/compose-network/localnet-control-plane/configs"
)

const (
	rollupARPC = "http://localhost:18545"
	rollupBRPC = "http://localhost:28545"
)

// DeployContracts deploys helper contracts to both rollups using go-ethereum.
func DeployContracts(ctx context.Context, contractsDir, networksDir string, cfg configs.L2) error {
	if _, err := os.Stat(contractsDir); os.IsNotExist(err) {
		slog.Warn("contracts directory not found; skipping helper deployment", "dir", contractsDir)
		return nil
	}

	slog.Info("waiting for rollup RPCs")
	if err := waitForRPC(ctx, rollupARPC, "Rollup A"); err != nil {
		return err
	}
	if err := waitForRPC(ctx, rollupBRPC, "Rollup B"); err != nil {
		return err
	}

	settleDelay := 20 * time.Second
	slog.Info("waiting for services to settle", "delay", settleDelay)
	time.Sleep(settleDelay)

	// Load precompiled contracts
	slog.Info("loading precompiled contracts")
	compiledContracts, err := loadCompiledContracts(contractsDir)
	if err != nil {
		return fmt.Errorf("failed to load compiled contracts: %w", err)
	}

	// Deploy to Rollup A
	slog.Info("deploying helper contracts to Rollup A")
	addressesA, err := deployToChain(ctx, rollupARPC, cfg, compiledContracts)
	if err != nil {
		return fmt.Errorf("failed to deploy to Rollup A: %w", err)
	}

	// Deploy to Rollup B
	slog.Info("deploying helper contracts to Rollup B")
	addressesB, err := deployToChain(ctx, rollupBRPC, cfg, compiledContracts)
	if err != nil {
		return fmt.Errorf("failed to deploy to Rollup B: %w", err)
	}

	// Verify addresses match (CREATE2 should make them deterministic)
	if !addressesMatch(addressesA, addressesB) {
		slog.Warn("helper contract addresses differ between rollups; not updating config")
		return nil
	}

	// Write config files
	rollupADir := filepath.Join(networksDir, "rollup-a")
	rollupBDir := filepath.Join(networksDir, "rollup-b")

	if err := writeContractJSON(filepath.Join(rollupADir, "contracts.json"), addressesA, uint64(cfg.ChainIDs.RollupA)); err != nil {
		return fmt.Errorf("failed to write contracts.json for Rollup A: %w", err)
	}

	if err := writeContractJSON(filepath.Join(rollupBDir, "contracts.json"), addressesB, uint64(cfg.ChainIDs.RollupB)); err != nil {
		return fmt.Errorf("failed to write contracts.json for Rollup B: %w", err)
	}

	opGethConfigPath := filepath.Join(contractsDir, "..", "services", "op-geth", "config.yml")
	if err := writeHelperConfig(opGethConfigPath, addressesA, cfg); err != nil {
		return fmt.Errorf("failed to write helper config: %w", err)
	}

	slog.Info("helper contracts deployed successfully")
	return nil
}

func waitForRPC(ctx context.Context, url, name string) error {
	client, err := ethclient.DialContext(ctx, url)
	if err == nil {
		defer client.Close()
	}

	for i := 0; i < 120; i++ {
		client, err := ethclient.DialContext(ctx, url)
		if err == nil {
			defer client.Close()
			_, err := client.BlockNumber(ctx)
			if err == nil {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timed out waiting for %s RPC at %s", name, url)
}

type CompiledContract struct {
	ABI      abi.ABI
	Bytecode []byte
}

func loadCompiledContracts(contractsDir string) (map[string]*CompiledContract, error) {
	compiledPath := filepath.Join(contractsDir, "compiled", "contracts.json")
	data, err := os.ReadFile(compiledPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compiled contracts: %w", err)
	}

	var result map[string]struct {
		ABI      json.RawMessage `json:"abi"`
		Bytecode string          `json:"bytecode"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse compiled contracts: %w", err)
	}

	contracts := make(map[string]*CompiledContract)
	for contractName, contract := range result {
		parsedABI, err := abi.JSON(strings.NewReader(string(contract.ABI)))
		if err != nil {
			return nil, fmt.Errorf("failed to parse ABI for %s: %w", contractName, err)
		}

		bytecodeHex := strings.TrimPrefix(contract.Bytecode, "0x")
		bytecode := common.Hex2Bytes(bytecodeHex)

		contracts[contractName] = &CompiledContract{
			ABI:      parsedABI,
			Bytecode: bytecode,
		}
	}

	return contracts, nil
}

func deployToChain(ctx context.Context, rpcURL string, cfg configs.L2, contracts map[string]*CompiledContract) (map[string]string, error) {
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", rpcURL, err)
	}
	defer client.Close()

	// Parse private key
	privateKeyHex := strings.TrimPrefix(cfg.CoordinatorPrivateKey, "0x")
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA")
	}

	coordinatorAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	addresses := make(map[string]string)

	// Deploy Mailbox
	mailboxAddr, err := deployContract(ctx, client, privateKey, chainID, contracts["Mailbox"], coordinatorAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Mailbox: %w", err)
	}
	addresses["Mailbox"] = mailboxAddr.Hex()
	slog.Info("deployed Mailbox", "address", mailboxAddr.Hex())

	// Deploy PingPong
	pingPongAddr, err := deployContract(ctx, client, privateKey, chainID, contracts["PingPong"], mailboxAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy PingPong: %w", err)
	}
	addresses["PingPong"] = pingPongAddr.Hex()
	slog.Info("deployed PingPong", "address", pingPongAddr.Hex())

	// Deploy Bridge
	bridgeAddr, err := deployContract(ctx, client, privateKey, chainID, contracts["Bridge"], mailboxAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Bridge: %w", err)
	}
	addresses["Bridge"] = bridgeAddr.Hex()
	slog.Info("deployed Bridge", "address", bridgeAddr.Hex())

	// Deploy MyToken
	myTokenAddr, err := deployContract(ctx, client, privateKey, chainID, contracts["MyToken"], mailboxAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy MyToken: %w", err)
	}
	addresses["MyToken"] = myTokenAddr.Hex()
	slog.Info("deployed MyToken", "address", myTokenAddr.Hex())

	// Deploy Coordinator
	coordinatorAddr, err := deployContract(ctx, client, privateKey, chainID, contracts["Coordinator"], coordinatorAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Coordinator: %w", err)
	}
	addresses["Coordinator"] = coordinatorAddr.Hex()
	slog.Info("deployed Coordinator", "address", coordinatorAddr.Hex())

	return addresses, nil
}

func deployContract(ctx context.Context, client *ethclient.Client, privateKey *ecdsa.PrivateKey, chainID *big.Int, contract *CompiledContract, constructorArgs ...interface{}) (common.Address, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.Context = ctx
	auth.GasLimit = uint64(5_000_000)

	address, tx, _, err := bind.DeployContract(auth, contract.ABI, contract.Bytecode, client, constructorArgs...)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to deploy contract: %w", err)
	}

	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to wait for transaction: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return common.Address{}, fmt.Errorf("contract deployment failed with status %d", receipt.Status)
	}

	return address, nil
}

func writeContractJSON(path string, addresses map[string]string, chainID uint64) error {
	payload := map[string]any{
		"chainInfo": map[string]any{
			"chainId": chainID,
		},
		"addresses": addresses,
	}

	if err := writeJSON(path, payload); err != nil {
		return err
	}

	return nil
}

func writeJSON(path string, data any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", path, err)
	}

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON for %s: %w", path, err)
	}

	if err := os.WriteFile(path, append(content, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	return nil
}

func addressesMatch(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, aValue := range a {
		if bValue, ok := b[key]; !ok || !strings.EqualFold(aValue, bValue) {
			return false
		}
	}
	return true
}

func writeHelperConfig(path string, addresses map[string]string, cfg configs.L2) error {
	type helperData struct {
		MailboxAddress string
		ChainID        int
		L1RPCURL       string
		PrivateKey     string
	}

	data := helperData{
		MailboxAddress: addresses["Mailbox"],
		ChainID:        cfg.ChainIDs.RollupA,
		L1RPCURL:       cfg.L1ElURL,
		PrivateKey:     cfg.CoordinatorPrivateKey,
	}

	tplPath := filepath.Join(filepath.Dir(path), "config.yml.tmpl")
	templateBytes, err := os.ReadFile(tplPath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("config").Parse(string(templateBytes))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
