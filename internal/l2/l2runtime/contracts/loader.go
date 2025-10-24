package contracts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func LoadCompiledContracts(contractsDir string) (map[ContractName]CompiledContract, error) {
	compiledPath := filepath.Join(contractsDir, contractsFileName)
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

	loadedContracts := make(map[ContractName]CompiledContract)

	for name, contract := range result {
		parsedABI, err := abi.JSON(strings.NewReader(string(contract.ABI)))
		if err != nil {
			return nil, fmt.Errorf("failed to parse ABI for %s: %w", name, err)
		}

		bytecodeHex := strings.TrimPrefix(contract.Bytecode, "0x")
		bytecode := common.Hex2Bytes(bytecodeHex)

		if _, ok := Contracts[ContractName(name)]; ok {
			loadedContracts[ContractName(name)] = CompiledContract{
				ABI:      parsedABI,
				RawABI:   string(contract.ABI),
				Bytecode: bytecode,
			}
		}
	}

	return loadedContracts, nil
}
