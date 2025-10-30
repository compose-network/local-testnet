package contracts

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

//go:embed compiled/contracts.json
var compiledContractsFS embed.FS

// LoadCompiledContracts loads compiled contracts.
func LoadCompiledContracts() (map[ContractName]CompiledContract, error) {
	data, err := compiledContractsFS.ReadFile("compiled/contracts.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded contracts: %w", err)
	}

	return parseContracts(data)
}

// parseContracts parses contract JSON data into CompiledContract map
func parseContracts(data []byte) (map[ContractName]CompiledContract, error) {
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
