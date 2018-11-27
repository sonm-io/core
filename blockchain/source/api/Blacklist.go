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

// BlacklistABI is the input ABI used to generate the binding from.
const BlacklistABI = "[{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"market\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"adder\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"addee\",\"type\":\"address\"}],\"name\":\"AddedToBlacklist\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"remover\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"removee\",\"type\":\"address\"}],\"name\":\"RemovedFromBlacklist\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"}],\"name\":\"OwnershipRenounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"name\":\"_who\",\"type\":\"address\"},{\"name\":\"_whom\",\"type\":\"address\"}],\"name\":\"Check\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_who\",\"type\":\"address\"},{\"name\":\"_whom\",\"type\":\"address\"}],\"name\":\"Add\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_whom\",\"type\":\"address\"}],\"name\":\"Remove\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_root\",\"type\":\"address\"}],\"name\":\"AddMaster\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_root\",\"type\":\"address\"}],\"name\":\"RemoveMaster\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_market\",\"type\":\"address\"}],\"name\":\"SetMarketAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// BlacklistBin is the compiled bytecode used for deploying new contracts.
const BlacklistBin = `0x608060405260038054600160a060020a031916905534801561002057600080fd5b5060008054600160a060020a0319908116339081179091161790556105e68061004a6000396000f3006080604052600436106100a35763ffffffff7c0100000000000000000000000000000000000000000000000000000000600035041663473b736f81146100a8578063584720f5146100e3578063715018a61461010457806377d58a151461011b57806380f556051461013c5780638a1051f81461016d5780638da5cb5b1461018e578063968f600c146101a3578063be7c7ac3146101ca578063f2fde38b146101eb575b600080fd5b3480156100b457600080fd5b506100cf600160a060020a036004358116906024351661020c565b604080519115158252519081900360200190f35b3480156100ef57600080fd5b506100cf600160a060020a03600435166102bc565b34801561011057600080fd5b50610119610323565b005b34801561012757600080fd5b506100cf600160a060020a036004351661038f565b34801561014857600080fd5b506101516103d9565b60408051600160a060020a039092168252519081900360200190f35b34801561017957600080fd5b506100cf600160a060020a03600435166103e8565b34801561019a57600080fd5b5061015161044f565b3480156101af57600080fd5b506100cf600160a060020a036004358116906024351661045e565b3480156101d657600080fd5b506100cf600160a060020a036004351661048c565b3480156101f757600080fd5b50610119600160a060020a036004351661051a565b600354600090600160a060020a0316151561022657600080fd5b600354600160a060020a031633148061024e57503360009081526002602052604090205460ff165b151561025957600080fd5b600160a060020a03808416600081815260016020818152604080842095881680855295909152808320805460ff1916909217909155517f708802ac7da0a63d9f6b2df693b53345ad263e42d74c245110e1ec1e03a1567e9190a350600192915050565b60008054600160a060020a031633146102d457600080fd5b600160a060020a03821660009081526002602052604090205460ff1615156001146102fe57600080fd5b50600160a060020a03166000908152600260205260409020805460ff19169055600190565b600054600160a060020a0316331461033a57600080fd5b60008054604051600160a060020a03909116917ff8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c6482091a26000805473ffffffffffffffffffffffffffffffffffffffff19169055565b60008054600160a060020a031633146103a757600080fd5b5060038054600160a060020a03831673ffffffffffffffffffffffffffffffffffffffff199091161790556001919050565b600354600160a060020a031681565b60008054600160a060020a0316331461040057600080fd5b600160a060020a03821660009081526002602052604090205460ff161561042657600080fd5b50600160a060020a03166000908152600260205260409020805460ff1916600190811790915590565b600054600160a060020a031681565b600160a060020a03918216600090815260016020908152604080832093909416825291909152205460ff1690565b336000908152600160208181526040808420600160a060020a038616855290915282205460ff161515146104bf57600080fd5b336000818152600160209081526040808320600160a060020a0387168085529252808320805460ff19169055519092917f576a9aef294e1b4baf3617fde4cbc80ba5344d5eb508222f29e558981704a45791a3506001919050565b600054600160a060020a0316331461053157600080fd5b61053a8161053d565b50565b600160a060020a038116151561055257600080fd5b60008054604051600160a060020a03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a36000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a03929092169190911790555600a165627a7a72305820464d7612eb697c84e91586dcfd1cadb3322d3aa9c03146de08040a490ef5fd2b0029`

// DeployBlacklist deploys a new Ethereum contract, binding an instance of Blacklist to it.
func DeployBlacklist(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Blacklist, error) {
	parsed, err := abi.JSON(strings.NewReader(BlacklistABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(BlacklistBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Blacklist{BlacklistCaller: BlacklistCaller{contract: contract}, BlacklistTransactor: BlacklistTransactor{contract: contract}, BlacklistFilterer: BlacklistFilterer{contract: contract}}, nil
}

// Blacklist is an auto generated Go binding around an Ethereum contract.
type Blacklist struct {
	BlacklistCaller     // Read-only binding to the contract
	BlacklistTransactor // Write-only binding to the contract
	BlacklistFilterer   // Log filterer for contract events
}

// BlacklistCaller is an auto generated read-only Go binding around an Ethereum contract.
type BlacklistCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BlacklistTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BlacklistTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BlacklistFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BlacklistFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BlacklistSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BlacklistSession struct {
	Contract     *Blacklist        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BlacklistCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BlacklistCallerSession struct {
	Contract *BlacklistCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// BlacklistTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BlacklistTransactorSession struct {
	Contract     *BlacklistTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// BlacklistRaw is an auto generated low-level Go binding around an Ethereum contract.
type BlacklistRaw struct {
	Contract *Blacklist // Generic contract binding to access the raw methods on
}

// BlacklistCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BlacklistCallerRaw struct {
	Contract *BlacklistCaller // Generic read-only contract binding to access the raw methods on
}

// BlacklistTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BlacklistTransactorRaw struct {
	Contract *BlacklistTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBlacklist creates a new instance of Blacklist, bound to a specific deployed contract.
func NewBlacklist(address common.Address, backend bind.ContractBackend) (*Blacklist, error) {
	contract, err := bindBlacklist(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Blacklist{BlacklistCaller: BlacklistCaller{contract: contract}, BlacklistTransactor: BlacklistTransactor{contract: contract}, BlacklistFilterer: BlacklistFilterer{contract: contract}}, nil
}

// NewBlacklistCaller creates a new read-only instance of Blacklist, bound to a specific deployed contract.
func NewBlacklistCaller(address common.Address, caller bind.ContractCaller) (*BlacklistCaller, error) {
	contract, err := bindBlacklist(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BlacklistCaller{contract: contract}, nil
}

// NewBlacklistTransactor creates a new write-only instance of Blacklist, bound to a specific deployed contract.
func NewBlacklistTransactor(address common.Address, transactor bind.ContractTransactor) (*BlacklistTransactor, error) {
	contract, err := bindBlacklist(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BlacklistTransactor{contract: contract}, nil
}

// NewBlacklistFilterer creates a new log filterer instance of Blacklist, bound to a specific deployed contract.
func NewBlacklistFilterer(address common.Address, filterer bind.ContractFilterer) (*BlacklistFilterer, error) {
	contract, err := bindBlacklist(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BlacklistFilterer{contract: contract}, nil
}

// bindBlacklist binds a generic wrapper to an already deployed contract.
func bindBlacklist(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BlacklistABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Blacklist *BlacklistRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Blacklist.Contract.BlacklistCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Blacklist *BlacklistRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Blacklist.Contract.BlacklistTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Blacklist *BlacklistRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Blacklist.Contract.BlacklistTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Blacklist *BlacklistCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Blacklist.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Blacklist *BlacklistTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Blacklist.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Blacklist *BlacklistTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Blacklist.Contract.contract.Transact(opts, method, params...)
}

// Check is a free data retrieval call binding the contract method 0x968f600c.
//
// Solidity: function Check(_who address, _whom address) constant returns(bool)
func (_Blacklist *BlacklistCaller) Check(opts *bind.CallOpts, _who common.Address, _whom common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Blacklist.contract.Call(opts, out, "Check", _who, _whom)
	return *ret0, err
}

// Check is a free data retrieval call binding the contract method 0x968f600c.
//
// Solidity: function Check(_who address, _whom address) constant returns(bool)
func (_Blacklist *BlacklistSession) Check(_who common.Address, _whom common.Address) (bool, error) {
	return _Blacklist.Contract.Check(&_Blacklist.CallOpts, _who, _whom)
}

// Check is a free data retrieval call binding the contract method 0x968f600c.
//
// Solidity: function Check(_who address, _whom address) constant returns(bool)
func (_Blacklist *BlacklistCallerSession) Check(_who common.Address, _whom common.Address) (bool, error) {
	return _Blacklist.Contract.Check(&_Blacklist.CallOpts, _who, _whom)
}

// Market is a free data retrieval call binding the contract method 0x80f55605.
//
// Solidity: function market() constant returns(address)
func (_Blacklist *BlacklistCaller) Market(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Blacklist.contract.Call(opts, out, "market")
	return *ret0, err
}

// Market is a free data retrieval call binding the contract method 0x80f55605.
//
// Solidity: function market() constant returns(address)
func (_Blacklist *BlacklistSession) Market() (common.Address, error) {
	return _Blacklist.Contract.Market(&_Blacklist.CallOpts)
}

// Market is a free data retrieval call binding the contract method 0x80f55605.
//
// Solidity: function market() constant returns(address)
func (_Blacklist *BlacklistCallerSession) Market() (common.Address, error) {
	return _Blacklist.Contract.Market(&_Blacklist.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Blacklist *BlacklistCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Blacklist.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Blacklist *BlacklistSession) Owner() (common.Address, error) {
	return _Blacklist.Contract.Owner(&_Blacklist.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Blacklist *BlacklistCallerSession) Owner() (common.Address, error) {
	return _Blacklist.Contract.Owner(&_Blacklist.CallOpts)
}

// Add is a paid mutator transaction binding the contract method 0x473b736f.
//
// Solidity: function Add(_who address, _whom address) returns(bool)
func (_Blacklist *BlacklistTransactor) Add(opts *bind.TransactOpts, _who common.Address, _whom common.Address) (*types.Transaction, error) {
	return _Blacklist.contract.Transact(opts, "Add", _who, _whom)
}

// Add is a paid mutator transaction binding the contract method 0x473b736f.
//
// Solidity: function Add(_who address, _whom address) returns(bool)
func (_Blacklist *BlacklistSession) Add(_who common.Address, _whom common.Address) (*types.Transaction, error) {
	return _Blacklist.Contract.Add(&_Blacklist.TransactOpts, _who, _whom)
}

// Add is a paid mutator transaction binding the contract method 0x473b736f.
//
// Solidity: function Add(_who address, _whom address) returns(bool)
func (_Blacklist *BlacklistTransactorSession) Add(_who common.Address, _whom common.Address) (*types.Transaction, error) {
	return _Blacklist.Contract.Add(&_Blacklist.TransactOpts, _who, _whom)
}

// AddMaster is a paid mutator transaction binding the contract method 0x8a1051f8.
//
// Solidity: function AddMaster(_root address) returns(bool)
func (_Blacklist *BlacklistTransactor) AddMaster(opts *bind.TransactOpts, _root common.Address) (*types.Transaction, error) {
	return _Blacklist.contract.Transact(opts, "AddMaster", _root)
}

// AddMaster is a paid mutator transaction binding the contract method 0x8a1051f8.
//
// Solidity: function AddMaster(_root address) returns(bool)
func (_Blacklist *BlacklistSession) AddMaster(_root common.Address) (*types.Transaction, error) {
	return _Blacklist.Contract.AddMaster(&_Blacklist.TransactOpts, _root)
}

// AddMaster is a paid mutator transaction binding the contract method 0x8a1051f8.
//
// Solidity: function AddMaster(_root address) returns(bool)
func (_Blacklist *BlacklistTransactorSession) AddMaster(_root common.Address) (*types.Transaction, error) {
	return _Blacklist.Contract.AddMaster(&_Blacklist.TransactOpts, _root)
}

// Remove is a paid mutator transaction binding the contract method 0xbe7c7ac3.
//
// Solidity: function Remove(_whom address) returns(bool)
func (_Blacklist *BlacklistTransactor) Remove(opts *bind.TransactOpts, _whom common.Address) (*types.Transaction, error) {
	return _Blacklist.contract.Transact(opts, "Remove", _whom)
}

// Remove is a paid mutator transaction binding the contract method 0xbe7c7ac3.
//
// Solidity: function Remove(_whom address) returns(bool)
func (_Blacklist *BlacklistSession) Remove(_whom common.Address) (*types.Transaction, error) {
	return _Blacklist.Contract.Remove(&_Blacklist.TransactOpts, _whom)
}

// Remove is a paid mutator transaction binding the contract method 0xbe7c7ac3.
//
// Solidity: function Remove(_whom address) returns(bool)
func (_Blacklist *BlacklistTransactorSession) Remove(_whom common.Address) (*types.Transaction, error) {
	return _Blacklist.Contract.Remove(&_Blacklist.TransactOpts, _whom)
}

// RemoveMaster is a paid mutator transaction binding the contract method 0x584720f5.
//
// Solidity: function RemoveMaster(_root address) returns(bool)
func (_Blacklist *BlacklistTransactor) RemoveMaster(opts *bind.TransactOpts, _root common.Address) (*types.Transaction, error) {
	return _Blacklist.contract.Transact(opts, "RemoveMaster", _root)
}

// RemoveMaster is a paid mutator transaction binding the contract method 0x584720f5.
//
// Solidity: function RemoveMaster(_root address) returns(bool)
func (_Blacklist *BlacklistSession) RemoveMaster(_root common.Address) (*types.Transaction, error) {
	return _Blacklist.Contract.RemoveMaster(&_Blacklist.TransactOpts, _root)
}

// RemoveMaster is a paid mutator transaction binding the contract method 0x584720f5.
//
// Solidity: function RemoveMaster(_root address) returns(bool)
func (_Blacklist *BlacklistTransactorSession) RemoveMaster(_root common.Address) (*types.Transaction, error) {
	return _Blacklist.Contract.RemoveMaster(&_Blacklist.TransactOpts, _root)
}

// SetMarketAddress is a paid mutator transaction binding the contract method 0x77d58a15.
//
// Solidity: function SetMarketAddress(_market address) returns(bool)
func (_Blacklist *BlacklistTransactor) SetMarketAddress(opts *bind.TransactOpts, _market common.Address) (*types.Transaction, error) {
	return _Blacklist.contract.Transact(opts, "SetMarketAddress", _market)
}

// SetMarketAddress is a paid mutator transaction binding the contract method 0x77d58a15.
//
// Solidity: function SetMarketAddress(_market address) returns(bool)
func (_Blacklist *BlacklistSession) SetMarketAddress(_market common.Address) (*types.Transaction, error) {
	return _Blacklist.Contract.SetMarketAddress(&_Blacklist.TransactOpts, _market)
}

// SetMarketAddress is a paid mutator transaction binding the contract method 0x77d58a15.
//
// Solidity: function SetMarketAddress(_market address) returns(bool)
func (_Blacklist *BlacklistTransactorSession) SetMarketAddress(_market common.Address) (*types.Transaction, error) {
	return _Blacklist.Contract.SetMarketAddress(&_Blacklist.TransactOpts, _market)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Blacklist *BlacklistTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Blacklist.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Blacklist *BlacklistSession) RenounceOwnership() (*types.Transaction, error) {
	return _Blacklist.Contract.RenounceOwnership(&_Blacklist.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Blacklist *BlacklistTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Blacklist.Contract.RenounceOwnership(&_Blacklist.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_Blacklist *BlacklistTransactor) TransferOwnership(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _Blacklist.contract.Transact(opts, "transferOwnership", _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_Blacklist *BlacklistSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _Blacklist.Contract.TransferOwnership(&_Blacklist.TransactOpts, _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_Blacklist *BlacklistTransactorSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _Blacklist.Contract.TransferOwnership(&_Blacklist.TransactOpts, _newOwner)
}

// BlacklistAddedToBlacklistIterator is returned from FilterAddedToBlacklist and is used to iterate over the raw logs and unpacked data for AddedToBlacklist events raised by the Blacklist contract.
type BlacklistAddedToBlacklistIterator struct {
	Event *BlacklistAddedToBlacklist // Event containing the contract specifics and raw log

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
func (it *BlacklistAddedToBlacklistIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BlacklistAddedToBlacklist)
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
		it.Event = new(BlacklistAddedToBlacklist)
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
func (it *BlacklistAddedToBlacklistIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BlacklistAddedToBlacklistIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BlacklistAddedToBlacklist represents a AddedToBlacklist event raised by the Blacklist contract.
type BlacklistAddedToBlacklist struct {
	Adder common.Address
	Addee common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterAddedToBlacklist is a free log retrieval operation binding the contract event 0x708802ac7da0a63d9f6b2df693b53345ad263e42d74c245110e1ec1e03a1567e.
//
// Solidity: e AddedToBlacklist(adder indexed address, addee indexed address)
func (_Blacklist *BlacklistFilterer) FilterAddedToBlacklist(opts *bind.FilterOpts, adder []common.Address, addee []common.Address) (*BlacklistAddedToBlacklistIterator, error) {

	var adderRule []interface{}
	for _, adderItem := range adder {
		adderRule = append(adderRule, adderItem)
	}
	var addeeRule []interface{}
	for _, addeeItem := range addee {
		addeeRule = append(addeeRule, addeeItem)
	}

	logs, sub, err := _Blacklist.contract.FilterLogs(opts, "AddedToBlacklist", adderRule, addeeRule)
	if err != nil {
		return nil, err
	}
	return &BlacklistAddedToBlacklistIterator{contract: _Blacklist.contract, event: "AddedToBlacklist", logs: logs, sub: sub}, nil
}

// WatchAddedToBlacklist is a free log subscription operation binding the contract event 0x708802ac7da0a63d9f6b2df693b53345ad263e42d74c245110e1ec1e03a1567e.
//
// Solidity: e AddedToBlacklist(adder indexed address, addee indexed address)
func (_Blacklist *BlacklistFilterer) WatchAddedToBlacklist(opts *bind.WatchOpts, sink chan<- *BlacklistAddedToBlacklist, adder []common.Address, addee []common.Address) (event.Subscription, error) {

	var adderRule []interface{}
	for _, adderItem := range adder {
		adderRule = append(adderRule, adderItem)
	}
	var addeeRule []interface{}
	for _, addeeItem := range addee {
		addeeRule = append(addeeRule, addeeItem)
	}

	logs, sub, err := _Blacklist.contract.WatchLogs(opts, "AddedToBlacklist", adderRule, addeeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BlacklistAddedToBlacklist)
				if err := _Blacklist.contract.UnpackLog(event, "AddedToBlacklist", log); err != nil {
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

// BlacklistOwnershipRenouncedIterator is returned from FilterOwnershipRenounced and is used to iterate over the raw logs and unpacked data for OwnershipRenounced events raised by the Blacklist contract.
type BlacklistOwnershipRenouncedIterator struct {
	Event *BlacklistOwnershipRenounced // Event containing the contract specifics and raw log

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
func (it *BlacklistOwnershipRenouncedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BlacklistOwnershipRenounced)
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
		it.Event = new(BlacklistOwnershipRenounced)
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
func (it *BlacklistOwnershipRenouncedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BlacklistOwnershipRenouncedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BlacklistOwnershipRenounced represents a OwnershipRenounced event raised by the Blacklist contract.
type BlacklistOwnershipRenounced struct {
	PreviousOwner common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipRenounced is a free log retrieval operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_Blacklist *BlacklistFilterer) FilterOwnershipRenounced(opts *bind.FilterOpts, previousOwner []common.Address) (*BlacklistOwnershipRenouncedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _Blacklist.contract.FilterLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BlacklistOwnershipRenouncedIterator{contract: _Blacklist.contract, event: "OwnershipRenounced", logs: logs, sub: sub}, nil
}

// WatchOwnershipRenounced is a free log subscription operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_Blacklist *BlacklistFilterer) WatchOwnershipRenounced(opts *bind.WatchOpts, sink chan<- *BlacklistOwnershipRenounced, previousOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _Blacklist.contract.WatchLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BlacklistOwnershipRenounced)
				if err := _Blacklist.contract.UnpackLog(event, "OwnershipRenounced", log); err != nil {
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

// BlacklistOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Blacklist contract.
type BlacklistOwnershipTransferredIterator struct {
	Event *BlacklistOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BlacklistOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BlacklistOwnershipTransferred)
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
		it.Event = new(BlacklistOwnershipTransferred)
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
func (it *BlacklistOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BlacklistOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BlacklistOwnershipTransferred represents a OwnershipTransferred event raised by the Blacklist contract.
type BlacklistOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_Blacklist *BlacklistFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BlacklistOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Blacklist.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BlacklistOwnershipTransferredIterator{contract: _Blacklist.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_Blacklist *BlacklistFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BlacklistOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Blacklist.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BlacklistOwnershipTransferred)
				if err := _Blacklist.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// BlacklistRemovedFromBlacklistIterator is returned from FilterRemovedFromBlacklist and is used to iterate over the raw logs and unpacked data for RemovedFromBlacklist events raised by the Blacklist contract.
type BlacklistRemovedFromBlacklistIterator struct {
	Event *BlacklistRemovedFromBlacklist // Event containing the contract specifics and raw log

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
func (it *BlacklistRemovedFromBlacklistIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BlacklistRemovedFromBlacklist)
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
		it.Event = new(BlacklistRemovedFromBlacklist)
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
func (it *BlacklistRemovedFromBlacklistIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BlacklistRemovedFromBlacklistIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BlacklistRemovedFromBlacklist represents a RemovedFromBlacklist event raised by the Blacklist contract.
type BlacklistRemovedFromBlacklist struct {
	Remover common.Address
	Removee common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRemovedFromBlacklist is a free log retrieval operation binding the contract event 0x576a9aef294e1b4baf3617fde4cbc80ba5344d5eb508222f29e558981704a457.
//
// Solidity: e RemovedFromBlacklist(remover indexed address, removee indexed address)
func (_Blacklist *BlacklistFilterer) FilterRemovedFromBlacklist(opts *bind.FilterOpts, remover []common.Address, removee []common.Address) (*BlacklistRemovedFromBlacklistIterator, error) {

	var removerRule []interface{}
	for _, removerItem := range remover {
		removerRule = append(removerRule, removerItem)
	}
	var removeeRule []interface{}
	for _, removeeItem := range removee {
		removeeRule = append(removeeRule, removeeItem)
	}

	logs, sub, err := _Blacklist.contract.FilterLogs(opts, "RemovedFromBlacklist", removerRule, removeeRule)
	if err != nil {
		return nil, err
	}
	return &BlacklistRemovedFromBlacklistIterator{contract: _Blacklist.contract, event: "RemovedFromBlacklist", logs: logs, sub: sub}, nil
}

// WatchRemovedFromBlacklist is a free log subscription operation binding the contract event 0x576a9aef294e1b4baf3617fde4cbc80ba5344d5eb508222f29e558981704a457.
//
// Solidity: e RemovedFromBlacklist(remover indexed address, removee indexed address)
func (_Blacklist *BlacklistFilterer) WatchRemovedFromBlacklist(opts *bind.WatchOpts, sink chan<- *BlacklistRemovedFromBlacklist, remover []common.Address, removee []common.Address) (event.Subscription, error) {

	var removerRule []interface{}
	for _, removerItem := range remover {
		removerRule = append(removerRule, removerItem)
	}
	var removeeRule []interface{}
	for _, removeeItem := range removee {
		removeeRule = append(removeeRule, removeeItem)
	}

	logs, sub, err := _Blacklist.contract.WatchLogs(opts, "RemovedFromBlacklist", removerRule, removeeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BlacklistRemovedFromBlacklist)
				if err := _Blacklist.contract.UnpackLog(event, "RemovedFromBlacklist", log); err != nil {
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
