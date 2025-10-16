package contracts

import "github.com/ethereum/go-ethereum/accounts/abi"

type (
	contractName     string
	compiledContract struct {
		ABI      abi.ABI
		Bytecode []byte
	}
)

const (
	contractNameBridge   = "Bridge"
	contractNameMailbox  = "Mailbox"
	contractNamePingPong = "PingPong"
	contractNameMyToken  = "MyToken"
)

var contracts = map[contractName]struct{}{
	contractNameBridge:   {},
	contractNameMailbox:  {},
	contractNamePingPong: {},
	contractNameMyToken:  {},
}
