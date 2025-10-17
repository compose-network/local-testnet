// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// BridgeMetaData contains all meta data concerning the Bridge contract.
var BridgeMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_mailbox\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"mailbox\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIMailbox\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"receiveTokens\",\"inputs\":[{\"name\":\"chainSrc\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"chainDest\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"receiver\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"sessionId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"send\",\"inputs\":[{\"name\":\"chainSrc\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"chainDest\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"receiver\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"sessionId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"DataWritten\",\"inputs\":[{\"name\":\"data\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EmptyEvent\",\"inputs\":[],\"anonymous\":false}]",
	Bin: "0x6080604052348015600e575f5ffd5b50604051610be0380380610be0833981016040819052602b91604e565b5f80546001600160a01b0319166001600160a01b03929092169190911790556079565b5f60208284031215605d575f5ffd5b81516001600160a01b03811681146072575f5ffd5b9392505050565b610b5a806100865f395ff3fe608060405234801561000f575f5ffd5b506004361061003f575f3560e01c80633cd28eb9146100435780639489e65e14610087578063d5438eae1461009c575b5f5ffd5b610056610051366004610703565b6100e0565b6040805173ffffffffffffffffffffffffffffffffffffffff90931683526020830191909152015b60405180910390f35b61009a610095366004610751565b610433565b005b5f546100bb9073ffffffffffffffffffffffffffffffffffffffff1681565b60405173ffffffffffffffffffffffffffffffffffffffff909116815260200161007e565b5f80546040517f93ab0ed40000000000000000000000000000000000000000000000000000000081528291829173ffffffffffffffffffffffffffffffffffffffff909116906393ab0ed490610140908b908a908a908a906004016107bd565b5f60405180830381865afa15801561015a573d5f5f3e3d5ffd5b505050506040513d5f823e601f3d908101601f191682016040526101819190810190610874565b905080515f0361018f575f5ffd5b5f5f828060200190518101906101a59190610928565b9097509550909250905073ffffffffffffffffffffffffffffffffffffffff80831690891614610236576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601760248201527f5468652073656e6465722073686f756c64206d6174636800000000000000000060448201526064015b60405180910390fd5b8673ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16146102cb576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601960248201527f5468652072656365697665722073686f756c64206d6174636800000000000000604482015260640161022d565b6040517f40c10f1900000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8881166004830152602482018690528616906340c10f19906044015f604051808303815f87803b158015610338575f5ffd5b505af115801561034a573d5f5f3e3d5ffd5b5050505060405160200161038f9060208082526002908201527f4f4b000000000000000000000000000000000000000000000000000000000000604082015260600190565b60408051601f19818403018152908290525f547fcf80ca9a00000000000000000000000000000000000000000000000000000000835290945073ffffffffffffffffffffffffffffffffffffffff169063cf80ca9a906103f9908c908b908b9089906004016109a6565b5f604051808303815f87803b158015610410575f5ffd5b505af1158015610422573d5f5f3e3d5ffd5b505050505050509550959350505050565b73ffffffffffffffffffffffffffffffffffffffff841633146104b2576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601960248201527f53686f756c6420626520746865207265616c2073656e64657200000000000000604482015260640161022d565b6040517f9dc29fac00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff858116600483015260248201849052861690639dc29fac906044015f604051808303815f87803b15801561051f575f5ffd5b505af1158015610531573d5f5f3e3d5ffd5b50506040805173ffffffffffffffffffffffffffffffffffffffff888116602083015287811682840152898116606083015260808083018890528351808403909101815260a08301938490525f547fcf80ca9a00000000000000000000000000000000000000000000000000000000909452945091909116915063cf80ca9a906105c5908b9088908790879060a401610a27565b5f604051808303815f87803b1580156105dc575f5ffd5b505af11580156105ee573d5f5f3e3d5ffd5b505050507f56217dea5af710bf0dc0a6c43e858f0eebf2aab53e7a16a2c364da75c405f5f3816040516106219190610a8b565b60405180910390a15f80546040517f93ab0ed400000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff909116906393ab0ed490610685908b908a908a908990600401610aa4565b5f60405180830381865afa15801561069f573d5f5f3e3d5ffd5b505050506040513d5f823e601f3d908101601f191682016040526106c69190810190610874565b905080515f036106d4575f5ffd5b505050505050505050565b73ffffffffffffffffffffffffffffffffffffffff81168114610700575f5ffd5b50565b5f5f5f5f5f60a08688031215610717575f5ffd5b85359450602086013593506040860135610730816106df565b92506060860135610740816106df565b949793965091946080013592915050565b5f5f5f5f5f5f5f60e0888a031215610767575f5ffd5b87359650602088013595506040880135610780816106df565b94506060880135610790816106df565b935060808801356107a0816106df565b9699959850939692959460a0840135945060c09093013592915050565b84815273ffffffffffffffffffffffffffffffffffffffff8416602082015273ffffffffffffffffffffffffffffffffffffffff8316604082015281606082015260a060808201525f61083d60a08301600481527f53454e4400000000000000000000000000000000000000000000000000000000602082015260400190565b9695505050505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b5f60208284031215610884575f5ffd5b815167ffffffffffffffff81111561089a575f5ffd5b8201601f810184136108aa575f5ffd5b805167ffffffffffffffff8111156108c4576108c4610847565b604051601f19603f601f19601f8501160116810181811067ffffffffffffffff821117156108f4576108f4610847565b60405281815282820160200186101561090b575f5ffd5b8160208401602083015e5f91810160200191909152949350505050565b5f5f5f5f6080858703121561093b575f5ffd5b8451610946816106df565b6020860151909450610957816106df565b6040860151909350610968816106df565b6060959095015193969295505050565b5f81518084528060208401602086015e5f602082860101526020601f19601f83011685010191505092915050565b84815273ffffffffffffffffffffffffffffffffffffffff8416602082015282604082015260a060608201525f610a0a60a08301600881527f41434b2053454e44000000000000000000000000000000000000000000000000602082015260400190565b8281036080840152610a1c8185610978565b979650505050505050565b84815273ffffffffffffffffffffffffffffffffffffffff8416602082015282604082015260a060608201525f610a0a60a08301600481527f53454e4400000000000000000000000000000000000000000000000000000000602082015260400190565b602081525f610a9d6020830184610978565b9392505050565b84815273ffffffffffffffffffffffffffffffffffffffff8416602082015273ffffffffffffffffffffffffffffffffffffffff8316604082015281606082015260a060808201525f61083d60a08301600881527f41434b2053454e4400000000000000000000000000000000000000000000000060208201526040019056fea26469706673582212204d3a8a55a76f6b361ba714daf6f45c3b2e745f1caaf6c496f19fafde92ac105464736f6c634300081e0033",
}

// BridgeABI is the input ABI used to generate the binding from.
// Deprecated: Use BridgeMetaData.ABI instead.
var BridgeABI = BridgeMetaData.ABI

// BridgeBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use BridgeMetaData.Bin instead.
var BridgeBin = BridgeMetaData.Bin

// DeployBridge deploys a new Ethereum contract, binding an instance of Bridge to it.
func DeployBridge(auth *bind.TransactOpts, backend bind.ContractBackend, _mailbox common.Address) (common.Address, *types.Transaction, *Bridge, error) {
	parsed, err := BridgeMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(BridgeBin), backend, _mailbox)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Bridge{BridgeCaller: BridgeCaller{contract: contract}, BridgeTransactor: BridgeTransactor{contract: contract}, BridgeFilterer: BridgeFilterer{contract: contract}}, nil
}

// Bridge is an auto generated Go binding around an Ethereum contract.
type Bridge struct {
	BridgeCaller     // Read-only binding to the contract
	BridgeTransactor // Write-only binding to the contract
	BridgeFilterer   // Log filterer for contract events
}

// BridgeCaller is an auto generated read-only Go binding around an Ethereum contract.
type BridgeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BridgeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BridgeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BridgeSession struct {
	Contract     *Bridge           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BridgeCallerSession struct {
	Contract *BridgeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// BridgeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BridgeTransactorSession struct {
	Contract     *BridgeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeRaw is an auto generated low-level Go binding around an Ethereum contract.
type BridgeRaw struct {
	Contract *Bridge // Generic contract binding to access the raw methods on
}

// BridgeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BridgeCallerRaw struct {
	Contract *BridgeCaller // Generic read-only contract binding to access the raw methods on
}

// BridgeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BridgeTransactorRaw struct {
	Contract *BridgeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBridge creates a new instance of Bridge, bound to a specific deployed contract.
func NewBridge(address common.Address, backend bind.ContractBackend) (*Bridge, error) {
	contract, err := bindBridge(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Bridge{BridgeCaller: BridgeCaller{contract: contract}, BridgeTransactor: BridgeTransactor{contract: contract}, BridgeFilterer: BridgeFilterer{contract: contract}}, nil
}

// NewBridgeCaller creates a new read-only instance of Bridge, bound to a specific deployed contract.
func NewBridgeCaller(address common.Address, caller bind.ContractCaller) (*BridgeCaller, error) {
	contract, err := bindBridge(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeCaller{contract: contract}, nil
}

// NewBridgeTransactor creates a new write-only instance of Bridge, bound to a specific deployed contract.
func NewBridgeTransactor(address common.Address, transactor bind.ContractTransactor) (*BridgeTransactor, error) {
	contract, err := bindBridge(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeTransactor{contract: contract}, nil
}

// NewBridgeFilterer creates a new log filterer instance of Bridge, bound to a specific deployed contract.
func NewBridgeFilterer(address common.Address, filterer bind.ContractFilterer) (*BridgeFilterer, error) {
	contract, err := bindBridge(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BridgeFilterer{contract: contract}, nil
}

// bindBridge binds a generic wrapper to an already deployed contract.
func bindBridge(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BridgeMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bridge *BridgeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bridge.Contract.BridgeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bridge *BridgeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bridge *BridgeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bridge *BridgeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bridge.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bridge *BridgeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bridge *BridgeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bridge.Contract.contract.Transact(opts, method, params...)
}

// Mailbox is a free data retrieval call binding the contract method 0xd5438eae.
//
// Solidity: function mailbox() view returns(address)
func (_Bridge *BridgeCaller) Mailbox(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "mailbox")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Mailbox is a free data retrieval call binding the contract method 0xd5438eae.
//
// Solidity: function mailbox() view returns(address)
func (_Bridge *BridgeSession) Mailbox() (common.Address, error) {
	return _Bridge.Contract.Mailbox(&_Bridge.CallOpts)
}

// Mailbox is a free data retrieval call binding the contract method 0xd5438eae.
//
// Solidity: function mailbox() view returns(address)
func (_Bridge *BridgeCallerSession) Mailbox() (common.Address, error) {
	return _Bridge.Contract.Mailbox(&_Bridge.CallOpts)
}

// ReceiveTokens is a paid mutator transaction binding the contract method 0x3cd28eb9.
//
// Solidity: function receiveTokens(uint256 chainSrc, uint256 chainDest, address sender, address receiver, uint256 sessionId) returns(address token, uint256 amount)
func (_Bridge *BridgeTransactor) ReceiveTokens(opts *bind.TransactOpts, chainSrc *big.Int, chainDest *big.Int, sender common.Address, receiver common.Address, sessionId *big.Int) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "receiveTokens", chainSrc, chainDest, sender, receiver, sessionId)
}

// ReceiveTokens is a paid mutator transaction binding the contract method 0x3cd28eb9.
//
// Solidity: function receiveTokens(uint256 chainSrc, uint256 chainDest, address sender, address receiver, uint256 sessionId) returns(address token, uint256 amount)
func (_Bridge *BridgeSession) ReceiveTokens(chainSrc *big.Int, chainDest *big.Int, sender common.Address, receiver common.Address, sessionId *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.ReceiveTokens(&_Bridge.TransactOpts, chainSrc, chainDest, sender, receiver, sessionId)
}

// ReceiveTokens is a paid mutator transaction binding the contract method 0x3cd28eb9.
//
// Solidity: function receiveTokens(uint256 chainSrc, uint256 chainDest, address sender, address receiver, uint256 sessionId) returns(address token, uint256 amount)
func (_Bridge *BridgeTransactorSession) ReceiveTokens(chainSrc *big.Int, chainDest *big.Int, sender common.Address, receiver common.Address, sessionId *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.ReceiveTokens(&_Bridge.TransactOpts, chainSrc, chainDest, sender, receiver, sessionId)
}

// Send is a paid mutator transaction binding the contract method 0x9489e65e.
//
// Solidity: function send(uint256 chainSrc, uint256 chainDest, address token, address sender, address receiver, uint256 amount, uint256 sessionId) returns()
func (_Bridge *BridgeTransactor) Send(opts *bind.TransactOpts, chainSrc *big.Int, chainDest *big.Int, token common.Address, sender common.Address, receiver common.Address, amount *big.Int, sessionId *big.Int) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "send", chainSrc, chainDest, token, sender, receiver, amount, sessionId)
}

// Send is a paid mutator transaction binding the contract method 0x9489e65e.
//
// Solidity: function send(uint256 chainSrc, uint256 chainDest, address token, address sender, address receiver, uint256 amount, uint256 sessionId) returns()
func (_Bridge *BridgeSession) Send(chainSrc *big.Int, chainDest *big.Int, token common.Address, sender common.Address, receiver common.Address, amount *big.Int, sessionId *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.Send(&_Bridge.TransactOpts, chainSrc, chainDest, token, sender, receiver, amount, sessionId)
}

// Send is a paid mutator transaction binding the contract method 0x9489e65e.
//
// Solidity: function send(uint256 chainSrc, uint256 chainDest, address token, address sender, address receiver, uint256 amount, uint256 sessionId) returns()
func (_Bridge *BridgeTransactorSession) Send(chainSrc *big.Int, chainDest *big.Int, token common.Address, sender common.Address, receiver common.Address, amount *big.Int, sessionId *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.Send(&_Bridge.TransactOpts, chainSrc, chainDest, token, sender, receiver, amount, sessionId)
}

// BridgeDataWrittenIterator is returned from FilterDataWritten and is used to iterate over the raw logs and unpacked data for DataWritten events raised by the Bridge contract.
type BridgeDataWrittenIterator struct {
	Event *BridgeDataWritten // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeDataWrittenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeDataWritten)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeDataWritten)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeDataWrittenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeDataWrittenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeDataWritten represents a DataWritten event raised by the Bridge contract.
type BridgeDataWritten struct {
	Data []byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterDataWritten is a free log retrieval operation binding the contract event 0x56217dea5af710bf0dc0a6c43e858f0eebf2aab53e7a16a2c364da75c405f5f3.
//
// Solidity: event DataWritten(bytes data)
func (_Bridge *BridgeFilterer) FilterDataWritten(opts *bind.FilterOpts) (*BridgeDataWrittenIterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "DataWritten")
	if err != nil {
		return nil, err
	}
	return &BridgeDataWrittenIterator{contract: _Bridge.contract, event: "DataWritten", logs: logs, sub: sub}, nil
}

// WatchDataWritten is a free log subscription operation binding the contract event 0x56217dea5af710bf0dc0a6c43e858f0eebf2aab53e7a16a2c364da75c405f5f3.
//
// Solidity: event DataWritten(bytes data)
func (_Bridge *BridgeFilterer) WatchDataWritten(opts *bind.WatchOpts, sink chan<- *BridgeDataWritten) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "DataWritten")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeDataWritten)
				if err := _Bridge.contract.UnpackLog(event, "DataWritten", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDataWritten is a log parse operation binding the contract event 0x56217dea5af710bf0dc0a6c43e858f0eebf2aab53e7a16a2c364da75c405f5f3.
//
// Solidity: event DataWritten(bytes data)
func (_Bridge *BridgeFilterer) ParseDataWritten(log types.Log) (*BridgeDataWritten, error) {
	event := new(BridgeDataWritten)
	if err := _Bridge.contract.UnpackLog(event, "DataWritten", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeEmptyEventIterator is returned from FilterEmptyEvent and is used to iterate over the raw logs and unpacked data for EmptyEvent events raised by the Bridge contract.
type BridgeEmptyEventIterator struct {
	Event *BridgeEmptyEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeEmptyEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeEmptyEvent)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeEmptyEvent)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeEmptyEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeEmptyEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeEmptyEvent represents a EmptyEvent event raised by the Bridge contract.
type BridgeEmptyEvent struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEmptyEvent is a free log retrieval operation binding the contract event 0xcf16a92280c1bbb43f72d31126b724d508df2877835849e8744017ab36a9b47f.
//
// Solidity: event EmptyEvent()
func (_Bridge *BridgeFilterer) FilterEmptyEvent(opts *bind.FilterOpts) (*BridgeEmptyEventIterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "EmptyEvent")
	if err != nil {
		return nil, err
	}
	return &BridgeEmptyEventIterator{contract: _Bridge.contract, event: "EmptyEvent", logs: logs, sub: sub}, nil
}

// WatchEmptyEvent is a free log subscription operation binding the contract event 0xcf16a92280c1bbb43f72d31126b724d508df2877835849e8744017ab36a9b47f.
//
// Solidity: event EmptyEvent()
func (_Bridge *BridgeFilterer) WatchEmptyEvent(opts *bind.WatchOpts, sink chan<- *BridgeEmptyEvent) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "EmptyEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeEmptyEvent)
				if err := _Bridge.contract.UnpackLog(event, "EmptyEvent", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEmptyEvent is a log parse operation binding the contract event 0xcf16a92280c1bbb43f72d31126b724d508df2877835849e8744017ab36a9b47f.
//
// Solidity: event EmptyEvent()
func (_Bridge *BridgeFilterer) ParseEmptyEvent(log types.Log) (*BridgeEmptyEvent, error) {
	event := new(BridgeEmptyEvent)
	if err := _Bridge.contract.UnpackLog(event, "EmptyEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
