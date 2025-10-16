package contracts

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/compose-network/localnet-control-plane/internal/logger"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Compiler compiles Solidity L2 contracts
type Compiler struct {
	contractsRootDir string
	outputDir        string
	logger           *slog.Logger
}

// NewCompiler creates a new contract compiler
func NewCompiler(contractsRootDir, outputDir string) *Compiler {
	return &Compiler{
		contractsRootDir: contractsRootDir,
		outputDir:        outputDir,
		logger:           logger.Named("contracts_compiler"),
	}
}

// Compile compiles Solidity contracts and persists the output
func (c *Compiler) Compile(ctx context.Context) error {
	c.logger.
		With("contracts_dir", c.contractsRootDir).
		Info("starting contract compilation")

	c.logger.Info("installing forge dependencies")
	if err := c.installDependencies(ctx); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	compiledContracts := make(map[contractName]compiledContract)
	for name := range contracts {
		c.logger.With("name", name).Info("compiling contract")

		contract, err := c.compileContract(ctx, name)
		if err != nil {
			return fmt.Errorf("failed to compile %s: %w", name, err)
		}

		compiledContracts[name] = contract
	}

	if err := c.writeContractsJSON(compiledContracts); err != nil {
		return fmt.Errorf("failed to write %s: %w", contractsFileName, err)
	}

	c.logger.Info("contracts compiled successfully")

	return nil
}

func (c *Compiler) installDependencies(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "forge", "install")
	cmd.Dir = c.contractsRootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("forge install failed: %w", err)
	}

	return nil
}

func (c *Compiler) compileContract(ctx context.Context, contractName contractName) (compiledContract, error) {
	abiCmd := exec.CommandContext(ctx, "forge", "inspect", string(contractName), "abi", "--json")
	// Forge automatically looks for contracts in src/ subdirectory relative to the working directory
	abiCmd.Dir = c.contractsRootDir

	abiOutput, err := abiCmd.Output()
	if err != nil {
		return compiledContract{}, fmt.Errorf("failed to get ABI for %s: %w", contractName, err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(abiOutput)))
	if err != nil {
		return compiledContract{}, fmt.Errorf("failed to parse ABI for %s: %w", contractName, err)
	}

	bytecodeCmd := exec.CommandContext(ctx, "forge", "inspect", string(contractName), "bytecode")
	bytecodeCmd.Dir = c.contractsRootDir

	bytecodeOutput, err := bytecodeCmd.Output()
	if err != nil {
		return compiledContract{}, fmt.Errorf("failed to get bytecode for %s: %w", contractName, err)
	}

	bytecodeStr := strings.TrimSpace(string(bytecodeOutput))
	bytecodeHex := strings.TrimPrefix(bytecodeStr, "0x")
	bytecode := common.Hex2Bytes(bytecodeHex)

	return compiledContract{
		ABI:      parsedABI,
		Bytecode: bytecode,
	}, nil
}

func (c *Compiler) writeContractsJSON(contracts map[contractName]compiledContract) error {
	if err := os.MkdirAll(c.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	outputPath := filepath.Join(c.outputDir, contractsFileName)

	data, err := json.MarshalIndent(contracts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal contracts: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", contractsFileName, err)
	}

	return nil
}
