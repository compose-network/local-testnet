package bridge

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/compose-network/localnet-control-plane/configs"
	fsjson "github.com/compose-network/localnet-control-plane/internal/l2/infra/filesystem/json"
	"github.com/compose-network/localnet-control-plane/internal/l2/l2runtime/contracts/bindings"
	rollupv1 "github.com/compose-network/localnet-control-plane/internal/l2/proto/rollup/v1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"google.golang.org/protobuf/proto"
)

const (
	sendTxRPCMethod = "eth_sendXTransaction"
)

type RollupConfig struct {
	ChainID *big.Int
	RPC     string
	Bridge  common.Address
	Token   common.Address
	Key     *ecdsa.PrivateKey
	Address common.Address
}

func ExecuteBridge(ctx context.Context, cfg configs.L2, amountStr string, sessionIDStr string, waitSeconds int) error {
	slog.Info("Starting cross-chain bridge transaction")

	amount := new(big.Int)
	if _, ok := amount.SetString(amountStr, 10); !ok {
		return fmt.Errorf("invalid amount: %s", amountStr)
	}

	var sessionID *big.Int
	if sessionIDStr == "" {
		var err error
		sessionID, err = generateRandomSessionID()
		if err != nil {
			return fmt.Errorf("failed to generate session ID: %w", err)
		}
		slog.Info("Generated random session ID", "session_id", sessionID)
	} else {
		sessionID = new(big.Int)
		if _, ok := sessionID.SetString(sessionIDStr, 10); !ok {
			return fmt.Errorf("invalid session ID: %s", sessionIDStr)
		}
	}

	rollupA, err := loadRollupConfig(cfg, configs.L2ChainNameRollupA)
	if err != nil {
		return fmt.Errorf("failed to load rollup A config: %w", err)
	}

	rollupB, err := loadRollupConfig(cfg, configs.L2ChainNameRollupB)
	if err != nil {
		return fmt.Errorf("failed to load rollup B config: %w", err)
	}

	slog.Info("Loaded rollup configurations",
		"rollup_a_chain_id", rollupA.ChainID,
		"rollup_b_chain_id", rollupB.ChainID,
		"rollup_a_bridge", rollupA.Bridge.Hex(),
		"rollup_b_bridge", rollupB.Bridge.Hex(),
	)

	sendTx, receiveTx, err := createBridgeTransactions(rollupA, rollupB, amount, sessionID)
	if err != nil {
		return fmt.Errorf("failed to create transactions: %w", err)
	}

	xtRequest, err := packageXTRequest(rollupA.ChainID, rollupB.ChainID, sendTx, receiveTx)
	if err != nil {
		return fmt.Errorf("failed to package XTRequest: %w", err)
	}

	slog.Info("Created cross-chain transaction package",
		"amount", amount,
		"session_id", sessionID,
		"sender", rollupA.Address.Hex(),
		"receiver", rollupB.Address.Hex(),
	)

	txHashes, err := submitXTransaction(rollupA.RPC, xtRequest)
	if err != nil {
		return fmt.Errorf("failed to submit transaction: %w", err)
	}

	slog.Info("Submitted cross-chain transaction",
		"tx_count", len(txHashes),
		"tx_hashes", txHashes,
	)

	if waitSeconds > 0 {
		slog.Info("Waiting for transaction confirmation", "seconds", waitSeconds)
		time.Sleep(time.Duration(waitSeconds) * time.Second)

		if err := checkTransactionStatus(ctx, rollupA.RPC, rollupB.RPC, txHashes); err != nil {
			slog.Warn("Failed to check transaction status", "error", err)
		}
	}

	slog.Info("Bridge transaction completed successfully")
	return nil
}

func loadRollupConfig(cfg configs.L2, rollupName configs.L2ChainName) (*RollupConfig, error) {
	// Get chain config
	chainConfig, exists := cfg.ChainConfigs[rollupName]
	if !exists {
		return nil, fmt.Errorf("chain config for %s not found", rollupName)
	}

	rpc := fmt.Sprintf("http://127.0.0.1:%d", chainConfig.RPCPort)

	contractsPath := fmt.Sprintf("networks/%s/contracts.json", rollupName)
	reader := fsjson.NewReader()
	var data map[string]any
	err := reader.ReadJSON(contractsPath, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", contractsPath, err)
	}

	bridgeAddr, ok := data["Bridge"].(string)
	if !ok {
		return nil, fmt.Errorf("bridge address not found in %s", contractsPath)
	}

	tokenAddr, ok := data["MyToken"].(string)
	if !ok {
		return nil, fmt.Errorf("MyToken address not found in %s", contractsPath)
	}

	privateKey, err := crypto.HexToECDSA(cfg.Wallet.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &RollupConfig{
		ChainID: big.NewInt(int64(chainConfig.ID)),
		RPC:     rpc,
		Bridge:  common.HexToAddress(bridgeAddr),
		Token:   common.HexToAddress(tokenAddr),
		Key:     privateKey,
		Address: address,
	}, nil
}

func createBridgeTransactions(rollupA, rollupB *RollupConfig, amount, sessionID *big.Int) (*types.Transaction, *types.Transaction, error) {
	nonceA, err := getNonce(rollupA.RPC, rollupA.Address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get nonce for rollup A: %w", err)
	}

	nonceB, err := getNonce(rollupB.RPC, rollupB.Address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get nonce for rollup B: %w", err)
	}

	sendTx, err := createSendTransaction(
		rollupA.ChainID,
		rollupB.ChainID,
		rollupA.Bridge,
		rollupB.Bridge,
		rollupA.Token,
		rollupA.Address,
		rollupB.Address,
		amount,
		sessionID,
		nonceA,
		rollupA.Key,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create send transaction: %w", err)
	}

	receiveTx, err := createReceiveTransaction(
		rollupA.ChainID,
		rollupB.ChainID,
		rollupA.Bridge,
		rollupB.Bridge,
		rollupA.Address,
		rollupB.Address,
		sessionID,
		nonceB,
		rollupB.Key,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create receive transaction: %w", err)
	}

	return sendTx, receiveTx, nil
}

func createSendTransaction(
	chainSrc, chainDest *big.Int,
	bridgeSrc, bridgeDest common.Address,
	token common.Address,
	sender, receiver common.Address,
	amount, sessionID *big.Int,
	nonce uint64,
	privateKey *ecdsa.PrivateKey,
) (*types.Transaction, error) {
	abi, err := bindings.BridgeMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get Bridge ABI: %w", err)
	}

	calldata, err := abi.Pack("send", chainSrc, chainDest, token, sender, receiver, amount, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to pack send call: %w", err)
	}

	txData := &types.DynamicFeeTx{
		ChainID:   chainSrc,
		Nonce:     nonce,
		GasTipCap: big.NewInt(1000000000),  // 1 gwei
		GasFeeCap: big.NewInt(20000000000), // 20 gwei
		Gas:       900000,
		To:        &bridgeSrc,
		Value:     big.NewInt(0),
		Data:      calldata,
	}

	tx := types.NewTx(txData)
	return types.SignTx(tx, types.NewLondonSigner(chainSrc), privateKey)
}

func createReceiveTransaction(
	chainSrc, chainDest *big.Int,
	bridgeSrc, bridgeDest common.Address,
	sender, receiver common.Address,
	sessionID *big.Int,
	nonce uint64,
	privateKey *ecdsa.PrivateKey,
) (*types.Transaction, error) {
	abi, err := bindings.BridgeMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get Bridge ABI: %w", err)
	}

	calldata, err := abi.Pack("receiveTokens", chainSrc, chainDest, sender, receiver, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to pack receiveTokens call: %w", err)
	}

	txData := &types.DynamicFeeTx{
		ChainID:   chainDest,
		Nonce:     nonce,
		GasTipCap: big.NewInt(1_000_000_000),  // 1 gwei
		GasFeeCap: big.NewInt(20_000_000_000), // 20 gwei
		Gas:       900000,
		To:        &bridgeDest,
		Value:     big.NewInt(0),
		Data:      calldata,
	}

	tx := types.NewTx(txData)
	return types.SignTx(tx, types.NewLondonSigner(chainDest), privateKey)
}

func packageXTRequest(chainA, chainB *big.Int, sendTx, receiveTx *types.Transaction) (*rollupv1.XTRequest, error) {
	rlpSendTx, err := sendTx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal send transaction: %w", err)
	}

	rlpReceiveTx, err := receiveTx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal receive transaction: %w", err)
	}

	return &rollupv1.XTRequest{
		Transactions: []*rollupv1.TransactionRequest{
			{
				ChainId:     chainA.Bytes(),
				Transaction: [][]byte{rlpSendTx},
			},
			{
				ChainId:     chainB.Bytes(),
				Transaction: [][]byte{rlpReceiveTx},
			},
		},
	}, nil
}

func submitXTransaction(rpcURL string, xtRequest *rollupv1.XTRequest) ([]common.Hash, error) {
	spMsg := &rollupv1.Message{
		SenderId: "localnet-bridge-cli",
		Payload: &rollupv1.Message_XtRequest{
			XtRequest: xtRequest,
		},
	}

	encodedPayload, err := proto.Marshal(spMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal XTRequest: %w", err)
	}

	slog.Debug("Encoded XTRequest", "size_bytes", len(encodedPayload))

	client, err := rpc.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}
	defer client.Close()

	var resultHashes []common.Hash
	err = client.CallContext(context.Background(), &resultHashes, sendTxRPCMethod, hexutil.Encode(encodedPayload))
	if err != nil {
		return nil, fmt.Errorf("RPC call failed: %w", err)
	}

	return resultHashes, nil
}

func getNonce(rpcURL string, address common.Address) (uint64, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to RPC: %w", err)
	}
	defer client.Close()

	nonce, err := client.PendingNonceAt(context.Background(), address)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending nonce: %w", err)
	}

	return nonce, nil
}

func checkTransactionStatus(ctx context.Context, rpcA, rpcB string, txHashes []common.Hash) error {
	if len(txHashes) < 2 {
		return fmt.Errorf("expected at least 2 transaction hashes, got %d", len(txHashes))
	}

	clientA, err := ethclient.Dial(rpcA)
	if err != nil {
		return fmt.Errorf("failed to connect to rollup A RPC: %w", err)
	}
	defer clientA.Close()

	clientB, err := ethclient.Dial(rpcB)
	if err != nil {
		return fmt.Errorf("failed to connect to rollup B RPC: %w", err)
	}
	defer clientB.Close()

	receiptA, err := clientA.TransactionReceipt(ctx, txHashes[0])
	if err != nil {
		slog.Warn("Transaction receipt not found on rollup A", "tx_hash", txHashes[0], "error", err)
	} else {
		slog.Info("Transaction confirmed on rollup A",
			"tx_hash", txHashes[0],
			"block_number", receiptA.BlockNumber,
			"status", receiptA.Status,
		)
	}

	receiptB, err := clientB.TransactionReceipt(ctx, txHashes[1])
	if err != nil {
		slog.Warn("Transaction receipt not found on rollup B", "tx_hash", txHashes[1], "error", err)
	} else {
		slog.Info("Transaction confirmed on rollup B",
			"tx_hash", txHashes[1],
			"block_number", receiptB.BlockNumber,
			"status", receiptB.Status,
		)
	}

	return nil
}

func generateRandomSessionID() (*big.Int, error) {
	max := new(big.Int).Lsh(big.NewInt(1), 63)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random number: %w", err)
	}
	return n, nil
}
