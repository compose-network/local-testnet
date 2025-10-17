package mailbox

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"unicode"

	"github.com/ethereum/go-ethereum/common"
)

// DecodeMailboxCall decodes mailbox function call input data
func DecodeMailboxCall(input string) (*MailboxCall, error) {
	if input == "" || input == "0x" {
		return nil, fmt.Errorf("empty input")
	}

	payload, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex: %w", err)
	}

	if len(payload) < 4 {
		return &MailboxCall{Error: "input too short"}, nil
	}

	selector := hex.EncodeToString(payload[:4])
	body := payload[4:]

	call := &MailboxCall{
		Selector: selector,
		Function: SelectorMap[selector],
	}

	if call.Function == "" {
		call.Function = "unknown"
	}

	// Decode based on selector
	switch selector {
	case "308508ff": // write(uint256,address,uint256,bytes,bytes)
		return decodeWrite308508ff(call, body)
	case "9b8b9a26": // write(uint256,uint256,address,address,uint256,bytes)
		return decodeWrite9b8b9a26(call, body)
	case "a19ad3c7": // write(uint256,uint256,address,uint256,bytes,bytes)
		return decodeWriteA19ad3c7(call, body)
	case "222a7194", "45f303fe": // putInbox variants
		return decodePutInbox222a7194(call, body)
	case "8c29401f": // putInbox(uint256,uint256,address,uint256,bytes,bytes)
		return decodePutInbox8c29401f(call, body)
	case "fa67378b": // read(uint256,address,address,uint256,bytes)
		return decodeReadFa67378b(call, body)
	case "bd8b74e8": // read(uint256,uint256,address,uint256,bytes)
		return decodeReadBd8b74e8(call, body)
	case "52efea6e": // clear()
		call.CallType = "clear"
		return call, nil
	default:
		call.Error = "unknown selector"
		return call, nil
	}
}

func decodeWrite308508ff(call *MailboxCall, body []byte) (*MailboxCall, error) {
	call.CallType = "write"
	words := chunk32(body)
	if len(words) < 5 {
		call.Error = "insufficient data"
		return call, nil
	}

	chainDest := new(big.Int).SetBytes(words[0]).Uint64()
	call.ChainDest = &chainDest

	receiver := common.BytesToAddress(words[1][12:])
	call.Receiver = &receiver

	sessionID := new(big.Int).SetBytes(words[2]).Uint64()
	call.SessionID = &sessionID

	labelOffset := new(big.Int).SetBytes(words[3]).Uint64()
	dataOffset := new(big.Int).SetBytes(words[4]).Uint64()

	call.Label = formatBytes(decodeBytes(body, int(labelOffset)))
	call.Data = formatBytes(decodeBytes(body, int(dataOffset)))

	return call, nil
}

func decodeWrite9b8b9a26(call *MailboxCall, body []byte) (*MailboxCall, error) {
	call.CallType = "write"
	words := chunk32(body)
	if len(words) < 6 {
		call.Error = "insufficient data"
		return call, nil
	}

	chainSrc := new(big.Int).SetBytes(words[0]).Uint64()
	call.ChainSrc = &chainSrc

	chainDest := new(big.Int).SetBytes(words[1]).Uint64()
	call.ChainDest = &chainDest

	sender := common.BytesToAddress(words[2][12:])
	call.Sender = &sender

	receiver := common.BytesToAddress(words[3][12:])
	call.Receiver = &receiver

	sessionID := new(big.Int).SetBytes(words[4]).Uint64()
	call.SessionID = &sessionID

	dataOffset := new(big.Int).SetBytes(words[5]).Uint64()
	call.Data = formatBytes(decodeBytes(body, int(dataOffset)))
	call.Label = "LEGACY"

	return call, nil
}

func decodeWriteA19ad3c7(call *MailboxCall, body []byte) (*MailboxCall, error) {
	call.CallType = "write"
	words := chunk32(body)
	if len(words) < 6 {
		call.Error = "insufficient data"
		return call, nil
	}

	chainSrc := new(big.Int).SetBytes(words[0]).Uint64()
	call.ChainSrc = &chainSrc

	chainDest := new(big.Int).SetBytes(words[1]).Uint64()
	call.ChainDest = &chainDest

	receiver := common.BytesToAddress(words[2][12:])
	call.Receiver = &receiver

	sessionID := new(big.Int).SetBytes(words[3]).Uint64()
	call.SessionID = &sessionID

	dataOffset := new(big.Int).SetBytes(words[4]).Uint64()
	labelOffset := new(big.Int).SetBytes(words[5]).Uint64()

	call.Data = formatBytes(decodeBytes(body, int(dataOffset)))
	call.Label = formatBytes(decodeBytes(body, int(labelOffset)))

	return call, nil
}

func decodePutInbox222a7194(call *MailboxCall, body []byte) (*MailboxCall, error) {
	call.CallType = "putInbox"
	words := chunk32(body)
	if len(words) < 6 {
		call.Error = "insufficient data"
		return call, nil
	}

	chainSrc := new(big.Int).SetBytes(words[0]).Uint64()
	call.ChainSrc = &chainSrc

	receiver := common.BytesToAddress(words[2][12:])
	call.Receiver = &receiver

	sessionID := new(big.Int).SetBytes(words[3]).Uint64()
	call.SessionID = &sessionID

	labelOffset := new(big.Int).SetBytes(words[4]).Uint64()
	dataOffset := new(big.Int).SetBytes(words[5]).Uint64()

	call.Label = formatBytes(decodeBytes(body, int(labelOffset)))
	call.Data = formatBytes(decodeBytes(body, int(dataOffset)))

	return call, nil
}

func decodePutInbox8c29401f(call *MailboxCall, body []byte) (*MailboxCall, error) {
	call.CallType = "putInbox"
	words := chunk32(body)
	if len(words) < 6 {
		call.Error = "insufficient data"
		return call, nil
	}

	chainSrc := new(big.Int).SetBytes(words[0]).Uint64()
	call.ChainSrc = &chainSrc

	chainDest := new(big.Int).SetBytes(words[1]).Uint64()
	call.ChainDest = &chainDest

	receiver := common.BytesToAddress(words[2][12:])
	call.Receiver = &receiver

	sessionID := new(big.Int).SetBytes(words[3]).Uint64()
	call.SessionID = &sessionID

	dataOffset := new(big.Int).SetBytes(words[4]).Uint64()
	labelOffset := new(big.Int).SetBytes(words[5]).Uint64()

	call.Data = formatBytes(decodeBytes(body, int(dataOffset)))
	call.Label = formatBytes(decodeBytes(body, int(labelOffset)))

	return call, nil
}

func decodeReadFa67378b(call *MailboxCall, body []byte) (*MailboxCall, error) {
	call.CallType = "read"
	words := chunk32(body)
	if len(words) < 5 {
		call.Error = "insufficient data"
		return call, nil
	}

	chainSrc := new(big.Int).SetBytes(words[0]).Uint64()
	call.ChainSrc = &chainSrc

	sender := common.BytesToAddress(words[1][12:])
	call.Sender = &sender

	receiver := common.BytesToAddress(words[2][12:])
	call.Receiver = &receiver

	sessionID := new(big.Int).SetBytes(words[3]).Uint64()
	call.SessionID = &sessionID

	labelOffset := new(big.Int).SetBytes(words[4]).Uint64()
	call.Label = formatBytes(decodeBytes(body, int(labelOffset)))

	return call, nil
}

func decodeReadBd8b74e8(call *MailboxCall, body []byte) (*MailboxCall, error) {
	call.CallType = "read"
	words := chunk32(body)
	if len(words) < 5 {
		call.Error = "insufficient data"
		return call, nil
	}

	chainSrc := new(big.Int).SetBytes(words[0]).Uint64()
	call.ChainSrc = &chainSrc

	chainDest := new(big.Int).SetBytes(words[1]).Uint64()
	call.ChainDest = &chainDest

	receiver := common.BytesToAddress(words[2][12:])
	call.Receiver = &receiver

	sessionID := new(big.Int).SetBytes(words[3]).Uint64()
	call.SessionID = &sessionID

	labelOffset := new(big.Int).SetBytes(words[4]).Uint64()
	call.Label = formatBytes(decodeBytes(body, int(labelOffset)))

	return call, nil
}

// chunk32 splits bytes into 32-byte words
func chunk32(data []byte) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(data); i += 32 {
		end := i + 32
		if end > len(data) {
			end = len(data)
		}
		chunk := make([]byte, 32)
		copy(chunk, data[i:end])
		chunks = append(chunks, chunk)
	}
	return chunks
}

// decodeBytes decodes dynamic bytes at offset
func decodeBytes(data []byte, offset int) []byte {
	if offset < 0 || offset+32 > len(data) {
		return []byte{}
	}
	length := new(big.Int).SetBytes(data[offset : offset+32]).Uint64()
	start := offset + 32
	end := start + int(length)
	if end > len(data) {
		return []byte{}
	}
	return data[start:end]
}

// formatBytes formats bytes for display
func formatBytes(value []byte) string {
	if len(value) == 0 {
		return "0x"
	}

	// Try UTF-8 decoding
	str := string(value)
	printable := true
	for _, r := range str {
		if !unicode.IsPrint(r) {
			printable = false
			break
		}
	}

	if printable {
		return str
	}

	return "0x" + hex.EncodeToString(value)
}
