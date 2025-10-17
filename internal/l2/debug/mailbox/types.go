package mailbox

import "github.com/ethereum/go-ethereum/common"

// MailboxCall represents a decoded mailbox function call
type MailboxCall struct {
	BlockNumber uint64
	TxHash      string
	From        string
	To          string
	Value       string
	Selector    string
	Function    string
	CallType    string
	ChainSrc    *uint64
	ChainDest   *uint64
	Sender      *common.Address
	Receiver    *common.Address
	SessionID   *uint64
	Label       string
	Data        string
	Error       string
}

// CallTrace represents the result of debug_traceTransaction with callTracer
type CallTrace struct {
	Type    string      `json:"type"`
	From    string      `json:"from"`
	To      string      `json:"to"`
	Value   string      `json:"value"`
	Gas     string      `json:"gas"`
	GasUsed string      `json:"gasUsed"`
	Input   string      `json:"input"`
	Output  string      `json:"output"`
	Error   string      `json:"error"`
	Calls   []CallTrace `json:"calls"`
}

// SelectorMap maps 4-byte function selectors to their signatures
var SelectorMap = map[string]string{
	"308508ff": "write(uint256,address,uint256,bytes,bytes)",
	"9b8b9a26": "write(uint256,uint256,address,address,uint256,bytes)",
	"222a7194": "putInbox(uint256,address,address,uint256,bytes,bytes)",
	"45f303fe": "putInbox(uint256,uint256,address,address,uint256,bytes)",
	"fa67378b": "read(uint256,address,address,uint256,bytes)",
	"a19ad3c7": "write(uint256,uint256,address,uint256,bytes,bytes)",
	"8c29401f": "putInbox(uint256,uint256,address,uint256,bytes,bytes)",
	"bd8b74e8": "read(uint256,uint256,address,uint256,bytes)",
	"52efea6e": "clear()",
}
