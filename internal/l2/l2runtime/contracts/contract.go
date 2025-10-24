package contracts

import "github.com/ethereum/go-ethereum/accounts/abi"

const contractsFileName = "contracts.json"

type (
	ContractName     string
	CompiledContract struct {
		ABI      abi.ABI
		RawABI   string
		Bytecode []byte
	}
)

const (
	ContractNameBridge          = "Bridge"
	ContractNameMailbox         = "Mailbox"
	ContractNamePingPong        = "PingPong"
	ContractNameBridgeableToken = "BridgeableToken"
	ContractNameStagedMailbox   = "StagedMailbox"
)

var Contracts = map[ContractName]struct{}{
	ContractNameBridge:          {},
	ContractNameMailbox:         {},
	ContractNamePingPong:        {},
	ContractNameBridgeableToken: {},
	ContractNameStagedMailbox:   {},
}
