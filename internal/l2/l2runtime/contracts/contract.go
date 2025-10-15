package contracts

type contractName string

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
