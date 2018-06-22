// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package api

import (
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// AddressHashMapABI is the input ABI used to generate the binding from.
const AddressHashMapABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_key\",\"type\":\"bytes32\"},{\"name\":\"_value\",\"type\":\"address\"}],\"name\":\"write\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_key\",\"type\":\"bytes32\"}],\"name\":\"read\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// AddressHashMapBin is the compiled bytecode used for deploying new contracts.
const AddressHashMapBin = `0x608060405234801561001057600080fd5b5060008054600160a060020a0319908116339081179091161790556102328061003a6000396000f3006080604052600436106100615763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166361da14398114610066578063853eadce1461009a5780638da5cb5b146100c0578063f2fde38b146100d5575b600080fd5b34801561007257600080fd5b5061007e6004356100f6565b60408051600160a060020a039092168252519081900360200190f35b3480156100a657600080fd5b506100be600435600160a060020a0360243516610111565b005b3480156100cc57600080fd5b5061007e610163565b3480156100e157600080fd5b506100be600160a060020a0360043516610172565b600090815260016020526040902054600160a060020a031690565b600054600160a060020a0316331461012857600080fd5b600091825260016020526040909120805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a03909216919091179055565b600054600160a060020a031681565b600054600160a060020a0316331461018957600080fd5b600160a060020a038116151561019e57600080fd5b60008054604051600160a060020a03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a36000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a03929092169190911790555600a165627a7a72305820fe07c062f1a544d9593b25caaee4c6036f7fb568dd5e322e4e8bc8994e47f5340029`

// DeployAddressHashMap deploys a new Ethereum contract, binding an instance of AddressHashMap to it.
func DeployAddressHashMap(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *AddressHashMap, error) {
	parsed, err := abi.JSON(strings.NewReader(AddressHashMapABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(AddressHashMapBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &AddressHashMap{AddressHashMapCaller: AddressHashMapCaller{contract: contract}, AddressHashMapTransactor: AddressHashMapTransactor{contract: contract}, AddressHashMapFilterer: AddressHashMapFilterer{contract: contract}}, nil
}

// AddressHashMap is an auto generated Go binding around an Ethereum contract.
type AddressHashMap struct {
	AddressHashMapCaller     // Read-only binding to the contract
	AddressHashMapTransactor // Write-only binding to the contract
	AddressHashMapFilterer   // Log filterer for contract events
}

// AddressHashMapCaller is an auto generated read-only Go binding around an Ethereum contract.
type AddressHashMapCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AddressHashMapTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AddressHashMapTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AddressHashMapFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AddressHashMapFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AddressHashMapSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AddressHashMapSession struct {
	Contract     *AddressHashMap   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AddressHashMapCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AddressHashMapCallerSession struct {
	Contract *AddressHashMapCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// AddressHashMapTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AddressHashMapTransactorSession struct {
	Contract     *AddressHashMapTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// AddressHashMapRaw is an auto generated low-level Go binding around an Ethereum contract.
type AddressHashMapRaw struct {
	Contract *AddressHashMap // Generic contract binding to access the raw methods on
}

// AddressHashMapCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AddressHashMapCallerRaw struct {
	Contract *AddressHashMapCaller // Generic read-only contract binding to access the raw methods on
}

// AddressHashMapTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AddressHashMapTransactorRaw struct {
	Contract *AddressHashMapTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAddressHashMap creates a new instance of AddressHashMap, bound to a specific deployed contract.
func NewAddressHashMap(address common.Address, backend bind.ContractBackend) (*AddressHashMap, error) {
	contract, err := bindAddressHashMap(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AddressHashMap{AddressHashMapCaller: AddressHashMapCaller{contract: contract}, AddressHashMapTransactor: AddressHashMapTransactor{contract: contract}, AddressHashMapFilterer: AddressHashMapFilterer{contract: contract}}, nil
}

// NewAddressHashMapCaller creates a new read-only instance of AddressHashMap, bound to a specific deployed contract.
func NewAddressHashMapCaller(address common.Address, caller bind.ContractCaller) (*AddressHashMapCaller, error) {
	contract, err := bindAddressHashMap(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AddressHashMapCaller{contract: contract}, nil
}

// NewAddressHashMapTransactor creates a new write-only instance of AddressHashMap, bound to a specific deployed contract.
func NewAddressHashMapTransactor(address common.Address, transactor bind.ContractTransactor) (*AddressHashMapTransactor, error) {
	contract, err := bindAddressHashMap(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AddressHashMapTransactor{contract: contract}, nil
}

// NewAddressHashMapFilterer creates a new log filterer instance of AddressHashMap, bound to a specific deployed contract.
func NewAddressHashMapFilterer(address common.Address, filterer bind.ContractFilterer) (*AddressHashMapFilterer, error) {
	contract, err := bindAddressHashMap(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AddressHashMapFilterer{contract: contract}, nil
}

// bindAddressHashMap binds a generic wrapper to an already deployed contract.
func bindAddressHashMap(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AddressHashMapABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AddressHashMap *AddressHashMapRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _AddressHashMap.Contract.AddressHashMapCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AddressHashMap *AddressHashMapRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AddressHashMap.Contract.AddressHashMapTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AddressHashMap *AddressHashMapRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AddressHashMap.Contract.AddressHashMapTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AddressHashMap *AddressHashMapCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _AddressHashMap.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AddressHashMap *AddressHashMapTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AddressHashMap.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AddressHashMap *AddressHashMapTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AddressHashMap.Contract.contract.Transact(opts, method, params...)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_AddressHashMap *AddressHashMapCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _AddressHashMap.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_AddressHashMap *AddressHashMapSession) Owner() (common.Address, error) {
	return _AddressHashMap.Contract.Owner(&_AddressHashMap.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_AddressHashMap *AddressHashMapCallerSession) Owner() (common.Address, error) {
	return _AddressHashMap.Contract.Owner(&_AddressHashMap.CallOpts)
}

// Read is a free data retrieval call binding the contract method 0x61da1439.
//
// Solidity: function read(_key bytes32) constant returns(address)
func (_AddressHashMap *AddressHashMapCaller) Read(opts *bind.CallOpts, _key [32]byte) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _AddressHashMap.contract.Call(opts, out, "read", _key)
	return *ret0, err
}

// Read is a free data retrieval call binding the contract method 0x61da1439.
//
// Solidity: function read(_key bytes32) constant returns(address)
func (_AddressHashMap *AddressHashMapSession) Read(_key [32]byte) (common.Address, error) {
	return _AddressHashMap.Contract.Read(&_AddressHashMap.CallOpts, _key)
}

// Read is a free data retrieval call binding the contract method 0x61da1439.
//
// Solidity: function read(_key bytes32) constant returns(address)
func (_AddressHashMap *AddressHashMapCallerSession) Read(_key [32]byte) (common.Address, error) {
	return _AddressHashMap.Contract.Read(&_AddressHashMap.CallOpts, _key)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_AddressHashMap *AddressHashMapTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _AddressHashMap.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_AddressHashMap *AddressHashMapSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AddressHashMap.Contract.TransferOwnership(&_AddressHashMap.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_AddressHashMap *AddressHashMapTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AddressHashMap.Contract.TransferOwnership(&_AddressHashMap.TransactOpts, newOwner)
}

// Write is a paid mutator transaction binding the contract method 0x853eadce.
//
// Solidity: function write(_key bytes32, _value address) returns()
func (_AddressHashMap *AddressHashMapTransactor) Write(opts *bind.TransactOpts, _key [32]byte, _value common.Address) (*types.Transaction, error) {
	return _AddressHashMap.contract.Transact(opts, "write", _key, _value)
}

// Write is a paid mutator transaction binding the contract method 0x853eadce.
//
// Solidity: function write(_key bytes32, _value address) returns()
func (_AddressHashMap *AddressHashMapSession) Write(_key [32]byte, _value common.Address) (*types.Transaction, error) {
	return _AddressHashMap.Contract.Write(&_AddressHashMap.TransactOpts, _key, _value)
}

// Write is a paid mutator transaction binding the contract method 0x853eadce.
//
// Solidity: function write(_key bytes32, _value address) returns()
func (_AddressHashMap *AddressHashMapTransactorSession) Write(_key [32]byte, _value common.Address) (*types.Transaction, error) {
	return _AddressHashMap.Contract.Write(&_AddressHashMap.TransactOpts, _key, _value)
}

// AddressHashMapOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the AddressHashMap contract.
type AddressHashMapOwnershipTransferredIterator struct {
	Event *AddressHashMapOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *AddressHashMapOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AddressHashMapOwnershipTransferred)
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
		it.Event = new(AddressHashMapOwnershipTransferred)
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
func (it *AddressHashMapOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AddressHashMapOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AddressHashMapOwnershipTransferred represents a OwnershipTransferred event raised by the AddressHashMap contract.
type AddressHashMapOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_AddressHashMap *AddressHashMapFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*AddressHashMapOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AddressHashMap.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &AddressHashMapOwnershipTransferredIterator{contract: _AddressHashMap.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_AddressHashMap *AddressHashMapFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *AddressHashMapOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AddressHashMap.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AddressHashMapOwnershipTransferred)
				if err := _AddressHashMap.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
