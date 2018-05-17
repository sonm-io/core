// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package api

import (
	"math/big"
	"strings"

	ethereum "github.com/sonm-io/go-ethereum"
	"github.com/sonm-io/go-ethereum/accounts/abi"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"github.com/sonm-io/go-ethereum/common"
	"github.com/sonm-io/go-ethereum/core/types"
	"github.com/sonm-io/go-ethereum/event"
)

// GatekeeperABI is the input ABI used to generate the binding from.
const GatekeeperABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"txNumber\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"PayInTx\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"BlockEmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"txNumber\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"FreezeInTrx\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"txNumber\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"FreezeOutTrx\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"PayIn\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_root\",\"type\":\"bytes32\"}],\"name\":\"VotePayout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_proof\",\"type\":\"bytes\"},{\"name\":\"_root\",\"type\":\"uint256\"},{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_txNumber\",\"type\":\"uint256\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"PayoutOne\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// GatekeeperBin is the compiled bytecode used for deploying new contracts.
const GatekeeperBin = `0x60806040526000600281905560048190556007805460ff19169055600b5534801561002957600080fd5b506040516020806106a383398101604052516000805460018054600160a060020a03948516600160a060020a0319918216179091553393909316908316811790921690911790556106248061007f6000396000f30060806040526004361061006c5763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416638da5cb5b81146100715780639fab56ac146100a2578063ad7397c5146100bc578063f2fde38b1461012f578063f6e3a45714610150575b600080fd5b34801561007d57600080fd5b50610086610168565b60408051600160a060020a039092168252519081900360200190f35b3480156100ae57600080fd5b506100ba600435610177565b005b3480156100c857600080fd5b506040805160206004803580820135601f81018490048402850184019095528484526100ba94369492936024939284019190819084018382808284375094975050843595505050506020820135600160a060020a03169160408101359150606001356102bc565b34801561013b57600080fd5b506100ba600160a060020a0360043516610425565b34801561015c57600080fd5b506100ba6004356104bd565b600054600160a060020a031681565b600154604080517f23b872dd000000000000000000000000000000000000000000000000000000008152600160a060020a033381166004830152308116602483015260448201859052915191909216916323b872dd9160648083019260209291908290030181600087803b1580156101ee57600080fd5b505af1158015610202573d6000803e3d6000fd5b505050506040513d602081101561021857600080fd5b5051151561022557600080fd5b600280546001019081905560408051918252602082018390528051600160a060020a033316927f63768eabd21c026cb17439a3c6556436c1b0219c2046875297ad3f4b14e6700f92908290030190a260025461020014156102b95760006002556040805143815290517fefc6c60ac095a7bb2aa44f6dba3421076ce7baccf540ff61f4d6150d1f8440d79181900360200190a15b50565b604080516c01000000000000000000000000600160a060020a038616028152601481018490526034810183905281519081900360540190206000868152600a6020908152838220548252600c81528382208383529052919091205460ff161561032457600080fd5b6000858152600a602052604090205461033f90879083610551565b151561034a57600080fd5b600154604080517fa9059cbb000000000000000000000000000000000000000000000000000000008152600160a060020a038781166004830152602482018690529151919092169163a9059cbb9160448083019260209291908290030181600087803b1580156103b957600080fd5b505af11580156103cd573d6000803e3d6000fd5b505050506040513d60208110156103e357600080fd5b505115156103f057600080fd5b6000948552600a60209081526040808720548752600c8252808720928752919052909320805460ff1916600117905550505050565b60005433600160a060020a0390811691161461044057600080fd5b600160a060020a038116151561045557600080fd5b60008054604051600160a060020a03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a36000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0392909216919091179055565b600160a060020a0333166000908152600560209081526040808320548484526009909252909120546104f49163ffffffff6105e516565b60008281526009602052604090205560075460ff1680610530575060075460ff16158015610530575060005433600160a060020a039081169116145b156102b957600b8054600101908190556000908152600a6020526040902055565b6000806000806020875181151561056457fe5b061561057357600093506105db565b5083905060205b865161ffff8216116105d557868101519250828210156105b257604080519283526020830184905280519283900301909120906105cd565b60408051848152602081019390935280519283900301909120905b60200161057a565b81861493505b5050509392505050565b818101828110156105f257fe5b929150505600a165627a7a72305820b04d1f47907354ece686a9c61aa0c2de27e40b7f4af8bdc32101d5c1f71889f50029`

// DeployGatekeeper deploys a new Ethereum contract, binding an instance of Gatekeeper to it.
func DeployGatekeeper(auth *bind.TransactOpts, backend bind.ContractBackend, _token common.Address) (common.Address, *types.Transaction, *Gatekeeper, error) {
	parsed, err := abi.JSON(strings.NewReader(GatekeeperABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(GatekeeperBin), backend, _token)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Gatekeeper{GatekeeperCaller: GatekeeperCaller{contract: contract}, GatekeeperTransactor: GatekeeperTransactor{contract: contract}, GatekeeperFilterer: GatekeeperFilterer{contract: contract}}, nil
}

// Gatekeeper is an auto generated Go binding around an Ethereum contract.
type Gatekeeper struct {
	GatekeeperCaller     // Read-only binding to the contract
	GatekeeperTransactor // Write-only binding to the contract
	GatekeeperFilterer   // Log filterer for contract events
}

// GatekeeperCaller is an auto generated read-only Go binding around an Ethereum contract.
type GatekeeperCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GatekeeperTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GatekeeperTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GatekeeperFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GatekeeperFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GatekeeperSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GatekeeperSession struct {
	Contract     *Gatekeeper       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GatekeeperCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GatekeeperCallerSession struct {
	Contract *GatekeeperCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// GatekeeperTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GatekeeperTransactorSession struct {
	Contract     *GatekeeperTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// GatekeeperRaw is an auto generated low-level Go binding around an Ethereum contract.
type GatekeeperRaw struct {
	Contract *Gatekeeper // Generic contract binding to access the raw methods on
}

// GatekeeperCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GatekeeperCallerRaw struct {
	Contract *GatekeeperCaller // Generic read-only contract binding to access the raw methods on
}

// GatekeeperTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GatekeeperTransactorRaw struct {
	Contract *GatekeeperTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGatekeeper creates a new instance of Gatekeeper, bound to a specific deployed contract.
func NewGatekeeper(address common.Address, backend bind.ContractBackend) (*Gatekeeper, error) {
	contract, err := bindGatekeeper(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Gatekeeper{GatekeeperCaller: GatekeeperCaller{contract: contract}, GatekeeperTransactor: GatekeeperTransactor{contract: contract}, GatekeeperFilterer: GatekeeperFilterer{contract: contract}}, nil
}

// NewGatekeeperCaller creates a new read-only instance of Gatekeeper, bound to a specific deployed contract.
func NewGatekeeperCaller(address common.Address, caller bind.ContractCaller) (*GatekeeperCaller, error) {
	contract, err := bindGatekeeper(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GatekeeperCaller{contract: contract}, nil
}

// NewGatekeeperTransactor creates a new write-only instance of Gatekeeper, bound to a specific deployed contract.
func NewGatekeeperTransactor(address common.Address, transactor bind.ContractTransactor) (*GatekeeperTransactor, error) {
	contract, err := bindGatekeeper(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GatekeeperTransactor{contract: contract}, nil
}

// NewGatekeeperFilterer creates a new log filterer instance of Gatekeeper, bound to a specific deployed contract.
func NewGatekeeperFilterer(address common.Address, filterer bind.ContractFilterer) (*GatekeeperFilterer, error) {
	contract, err := bindGatekeeper(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GatekeeperFilterer{contract: contract}, nil
}

// bindGatekeeper binds a generic wrapper to an already deployed contract.
func bindGatekeeper(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(GatekeeperABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Gatekeeper *GatekeeperRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Gatekeeper.Contract.GatekeeperCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Gatekeeper *GatekeeperRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Gatekeeper.Contract.GatekeeperTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Gatekeeper *GatekeeperRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Gatekeeper.Contract.GatekeeperTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Gatekeeper *GatekeeperCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Gatekeeper.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Gatekeeper *GatekeeperTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Gatekeeper.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Gatekeeper *GatekeeperTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Gatekeeper.Contract.contract.Transact(opts, method, params...)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Gatekeeper *GatekeeperCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Gatekeeper.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Gatekeeper *GatekeeperSession) Owner() (common.Address, error) {
	return _Gatekeeper.Contract.Owner(&_Gatekeeper.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Gatekeeper *GatekeeperCallerSession) Owner() (common.Address, error) {
	return _Gatekeeper.Contract.Owner(&_Gatekeeper.CallOpts)
}

// PayIn is a paid mutator transaction binding the contract method 0x9fab56ac.
//
// Solidity: function PayIn(_value uint256) returns()
func (_Gatekeeper *GatekeeperTransactor) PayIn(opts *bind.TransactOpts, _value *big.Int) (*types.Transaction, error) {
	return _Gatekeeper.contract.Transact(opts, "PayIn", _value)
}

// PayIn is a paid mutator transaction binding the contract method 0x9fab56ac.
//
// Solidity: function PayIn(_value uint256) returns()
func (_Gatekeeper *GatekeeperSession) PayIn(_value *big.Int) (*types.Transaction, error) {
	return _Gatekeeper.Contract.PayIn(&_Gatekeeper.TransactOpts, _value)
}

// PayIn is a paid mutator transaction binding the contract method 0x9fab56ac.
//
// Solidity: function PayIn(_value uint256) returns()
func (_Gatekeeper *GatekeeperTransactorSession) PayIn(_value *big.Int) (*types.Transaction, error) {
	return _Gatekeeper.Contract.PayIn(&_Gatekeeper.TransactOpts, _value)
}

// PayoutOne is a paid mutator transaction binding the contract method 0xad7397c5.
//
// Solidity: function PayoutOne(_proof bytes, _root uint256, _from address, _txNumber uint256, _value uint256) returns()
func (_Gatekeeper *GatekeeperTransactor) PayoutOne(opts *bind.TransactOpts, _proof []byte, _root *big.Int, _from common.Address, _txNumber *big.Int, _value *big.Int) (*types.Transaction, error) {
	return _Gatekeeper.contract.Transact(opts, "PayoutOne", _proof, _root, _from, _txNumber, _value)
}

// PayoutOne is a paid mutator transaction binding the contract method 0xad7397c5.
//
// Solidity: function PayoutOne(_proof bytes, _root uint256, _from address, _txNumber uint256, _value uint256) returns()
func (_Gatekeeper *GatekeeperSession) PayoutOne(_proof []byte, _root *big.Int, _from common.Address, _txNumber *big.Int, _value *big.Int) (*types.Transaction, error) {
	return _Gatekeeper.Contract.PayoutOne(&_Gatekeeper.TransactOpts, _proof, _root, _from, _txNumber, _value)
}

// PayoutOne is a paid mutator transaction binding the contract method 0xad7397c5.
//
// Solidity: function PayoutOne(_proof bytes, _root uint256, _from address, _txNumber uint256, _value uint256) returns()
func (_Gatekeeper *GatekeeperTransactorSession) PayoutOne(_proof []byte, _root *big.Int, _from common.Address, _txNumber *big.Int, _value *big.Int) (*types.Transaction, error) {
	return _Gatekeeper.Contract.PayoutOne(&_Gatekeeper.TransactOpts, _proof, _root, _from, _txNumber, _value)
}

// VotePayout is a paid mutator transaction binding the contract method 0xf6e3a457.
//
// Solidity: function VotePayout(_root bytes32) returns()
func (_Gatekeeper *GatekeeperTransactor) VotePayout(opts *bind.TransactOpts, _root [32]byte) (*types.Transaction, error) {
	return _Gatekeeper.contract.Transact(opts, "VotePayout", _root)
}

// VotePayout is a paid mutator transaction binding the contract method 0xf6e3a457.
//
// Solidity: function VotePayout(_root bytes32) returns()
func (_Gatekeeper *GatekeeperSession) VotePayout(_root [32]byte) (*types.Transaction, error) {
	return _Gatekeeper.Contract.VotePayout(&_Gatekeeper.TransactOpts, _root)
}

// VotePayout is a paid mutator transaction binding the contract method 0xf6e3a457.
//
// Solidity: function VotePayout(_root bytes32) returns()
func (_Gatekeeper *GatekeeperTransactorSession) VotePayout(_root [32]byte) (*types.Transaction, error) {
	return _Gatekeeper.Contract.VotePayout(&_Gatekeeper.TransactOpts, _root)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_Gatekeeper *GatekeeperTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Gatekeeper.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_Gatekeeper *GatekeeperSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Gatekeeper.Contract.TransferOwnership(&_Gatekeeper.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_Gatekeeper *GatekeeperTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Gatekeeper.Contract.TransferOwnership(&_Gatekeeper.TransactOpts, newOwner)
}

// GatekeeperBlockEmittedIterator is returned from FilterBlockEmitted and is used to iterate over the raw logs and unpacked data for BlockEmitted events raised by the Gatekeeper contract.
type GatekeeperBlockEmittedIterator struct {
	Event *GatekeeperBlockEmitted // Event containing the contract specifics and raw log

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
func (it *GatekeeperBlockEmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatekeeperBlockEmitted)
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
		it.Event = new(GatekeeperBlockEmitted)
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
func (it *GatekeeperBlockEmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatekeeperBlockEmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatekeeperBlockEmitted represents a BlockEmitted event raised by the Gatekeeper contract.
type GatekeeperBlockEmitted struct {
	BlockNumber *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterBlockEmitted is a free log retrieval operation binding the contract event 0xefc6c60ac095a7bb2aa44f6dba3421076ce7baccf540ff61f4d6150d1f8440d7.
//
// Solidity: event BlockEmitted(blockNumber uint256)
func (_Gatekeeper *GatekeeperFilterer) FilterBlockEmitted(opts *bind.FilterOpts) (*GatekeeperBlockEmittedIterator, error) {

	logs, sub, err := _Gatekeeper.contract.FilterLogs(opts, "BlockEmitted")
	if err != nil {
		return nil, err
	}
	return &GatekeeperBlockEmittedIterator{contract: _Gatekeeper.contract, event: "BlockEmitted", logs: logs, sub: sub}, nil
}

// WatchBlockEmitted is a free log subscription operation binding the contract event 0xefc6c60ac095a7bb2aa44f6dba3421076ce7baccf540ff61f4d6150d1f8440d7.
//
// Solidity: event BlockEmitted(blockNumber uint256)
func (_Gatekeeper *GatekeeperFilterer) WatchBlockEmitted(opts *bind.WatchOpts, sink chan<- *GatekeeperBlockEmitted) (event.Subscription, error) {

	logs, sub, err := _Gatekeeper.contract.WatchLogs(opts, "BlockEmitted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatekeeperBlockEmitted)
				if err := _Gatekeeper.contract.UnpackLog(event, "BlockEmitted", log); err != nil {
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

// GatekeeperFreezeInTrxIterator is returned from FilterFreezeInTrx and is used to iterate over the raw logs and unpacked data for FreezeInTrx events raised by the Gatekeeper contract.
type GatekeeperFreezeInTrxIterator struct {
	Event *GatekeeperFreezeInTrx // Event containing the contract specifics and raw log

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
func (it *GatekeeperFreezeInTrxIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatekeeperFreezeInTrx)
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
		it.Event = new(GatekeeperFreezeInTrx)
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
func (it *GatekeeperFreezeInTrxIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatekeeperFreezeInTrxIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatekeeperFreezeInTrx represents a FreezeInTrx event raised by the Gatekeeper contract.
type GatekeeperFreezeInTrx struct {
	From     common.Address
	TxNumber *big.Int
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterFreezeInTrx is a free log retrieval operation binding the contract event 0x077e9d08fe7deaad81556312688953ddd62a070bb4878f9b2f31d9d74c5df34e.
//
// Solidity: event FreezeInTrx(from indexed address, txNumber uint256, value uint256)
func (_Gatekeeper *GatekeeperFilterer) FilterFreezeInTrx(opts *bind.FilterOpts, from []common.Address) (*GatekeeperFreezeInTrxIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _Gatekeeper.contract.FilterLogs(opts, "FreezeInTrx", fromRule)
	if err != nil {
		return nil, err
	}
	return &GatekeeperFreezeInTrxIterator{contract: _Gatekeeper.contract, event: "FreezeInTrx", logs: logs, sub: sub}, nil
}

// WatchFreezeInTrx is a free log subscription operation binding the contract event 0x077e9d08fe7deaad81556312688953ddd62a070bb4878f9b2f31d9d74c5df34e.
//
// Solidity: event FreezeInTrx(from indexed address, txNumber uint256, value uint256)
func (_Gatekeeper *GatekeeperFilterer) WatchFreezeInTrx(opts *bind.WatchOpts, sink chan<- *GatekeeperFreezeInTrx, from []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _Gatekeeper.contract.WatchLogs(opts, "FreezeInTrx", fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatekeeperFreezeInTrx)
				if err := _Gatekeeper.contract.UnpackLog(event, "FreezeInTrx", log); err != nil {
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

// GatekeeperFreezeOutTrxIterator is returned from FilterFreezeOutTrx and is used to iterate over the raw logs and unpacked data for FreezeOutTrx events raised by the Gatekeeper contract.
type GatekeeperFreezeOutTrxIterator struct {
	Event *GatekeeperFreezeOutTrx // Event containing the contract specifics and raw log

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
func (it *GatekeeperFreezeOutTrxIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatekeeperFreezeOutTrx)
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
		it.Event = new(GatekeeperFreezeOutTrx)
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
func (it *GatekeeperFreezeOutTrxIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatekeeperFreezeOutTrxIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatekeeperFreezeOutTrx represents a FreezeOutTrx event raised by the Gatekeeper contract.
type GatekeeperFreezeOutTrx struct {
	From     common.Address
	TxNumber *big.Int
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterFreezeOutTrx is a free log retrieval operation binding the contract event 0x6657b571950d206399bd609573e4103ec477fd9ed103a2b7b0032253ecc180fb.
//
// Solidity: event FreezeOutTrx(from indexed address, txNumber uint256, value uint256)
func (_Gatekeeper *GatekeeperFilterer) FilterFreezeOutTrx(opts *bind.FilterOpts, from []common.Address) (*GatekeeperFreezeOutTrxIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _Gatekeeper.contract.FilterLogs(opts, "FreezeOutTrx", fromRule)
	if err != nil {
		return nil, err
	}
	return &GatekeeperFreezeOutTrxIterator{contract: _Gatekeeper.contract, event: "FreezeOutTrx", logs: logs, sub: sub}, nil
}

// WatchFreezeOutTrx is a free log subscription operation binding the contract event 0x6657b571950d206399bd609573e4103ec477fd9ed103a2b7b0032253ecc180fb.
//
// Solidity: event FreezeOutTrx(from indexed address, txNumber uint256, value uint256)
func (_Gatekeeper *GatekeeperFilterer) WatchFreezeOutTrx(opts *bind.WatchOpts, sink chan<- *GatekeeperFreezeOutTrx, from []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _Gatekeeper.contract.WatchLogs(opts, "FreezeOutTrx", fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatekeeperFreezeOutTrx)
				if err := _Gatekeeper.contract.UnpackLog(event, "FreezeOutTrx", log); err != nil {
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

// GatekeeperOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Gatekeeper contract.
type GatekeeperOwnershipTransferredIterator struct {
	Event *GatekeeperOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *GatekeeperOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatekeeperOwnershipTransferred)
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
		it.Event = new(GatekeeperOwnershipTransferred)
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
func (it *GatekeeperOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatekeeperOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatekeeperOwnershipTransferred represents a OwnershipTransferred event raised by the Gatekeeper contract.
type GatekeeperOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_Gatekeeper *GatekeeperFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*GatekeeperOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Gatekeeper.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &GatekeeperOwnershipTransferredIterator{contract: _Gatekeeper.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_Gatekeeper *GatekeeperFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *GatekeeperOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Gatekeeper.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatekeeperOwnershipTransferred)
				if err := _Gatekeeper.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// GatekeeperPayInTxIterator is returned from FilterPayInTx and is used to iterate over the raw logs and unpacked data for PayInTx events raised by the Gatekeeper contract.
type GatekeeperPayInTxIterator struct {
	Event *GatekeeperPayInTx // Event containing the contract specifics and raw log

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
func (it *GatekeeperPayInTxIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatekeeperPayInTx)
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
		it.Event = new(GatekeeperPayInTx)
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
func (it *GatekeeperPayInTxIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatekeeperPayInTxIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatekeeperPayInTx represents a PayInTx event raised by the Gatekeeper contract.
type GatekeeperPayInTx struct {
	From     common.Address
	TxNumber *big.Int
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPayInTx is a free log retrieval operation binding the contract event 0x63768eabd21c026cb17439a3c6556436c1b0219c2046875297ad3f4b14e6700f.
//
// Solidity: event PayInTx(from indexed address, txNumber uint256, value uint256)
func (_Gatekeeper *GatekeeperFilterer) FilterPayInTx(opts *bind.FilterOpts, from []common.Address) (*GatekeeperPayInTxIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _Gatekeeper.contract.FilterLogs(opts, "PayInTx", fromRule)
	if err != nil {
		return nil, err
	}
	return &GatekeeperPayInTxIterator{contract: _Gatekeeper.contract, event: "PayInTx", logs: logs, sub: sub}, nil
}

// WatchPayInTx is a free log subscription operation binding the contract event 0x63768eabd21c026cb17439a3c6556436c1b0219c2046875297ad3f4b14e6700f.
//
// Solidity: event PayInTx(from indexed address, txNumber uint256, value uint256)
func (_Gatekeeper *GatekeeperFilterer) WatchPayInTx(opts *bind.WatchOpts, sink chan<- *GatekeeperPayInTx, from []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _Gatekeeper.contract.WatchLogs(opts, "PayInTx", fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatekeeperPayInTx)
				if err := _Gatekeeper.contract.UnpackLog(event, "PayInTx", log); err != nil {
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
