// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package HubWallet

import (
	"strings"

	"github.com/sonm-io/go-ethereum/accounts/abi"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"github.com/sonm-io/go-ethereum/common"
	"github.com/sonm-io/go-ethereum/core/types"
)

// HubWalletABI is the input ABI used to generate the binding from.
const HubWalletABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_whitelist\",\"type\":\"address\"}],\"payable\":false,\"type\":\"constructor\"}]"

// HubWalletBin is the compiled bytecode used for deploying new contracts.
const HubWalletBin = `0x6060604052341561000f57600080fd5b6040516040806101ab83398101604052808051919060200180519150505b5b60008054600160a060020a03191633600160a060020a03161790555b60008054600160a060020a0319808216600160a060020a039283161790925560018054918416919092161790555b50505b6101218061008a6000396000f300606060405263ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416638da5cb5b81146046578063f2fde38b146072575b600080fd5b3415605057600080fd5b60566090565b604051600160a060020a03909116815260200160405180910390f35b3415607c57600080fd5b608e600160a060020a0360043516609f565b005b600054600160a060020a031681565b60005433600160a060020a0390811691161460b957600080fd5b600160a060020a0381161560f0576000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0383161790555b5b5b505600a165627a7a723058206d088466073b6a0d2d0860abc1f019f94097d74c83f4db4f732c1f76a8793dad0029`

// DeployHubWallet deploys a new Ethereum contract, binding an instance of HubWallet to it.
func DeployHubWallet(auth *bind.TransactOpts, backend bind.ContractBackend, _owner common.Address, _whitelist common.Address) (common.Address, *types.Transaction, *HubWallet, error) {
	parsed, err := abi.JSON(strings.NewReader(HubWalletABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(HubWalletBin), backend, _owner, _whitelist)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &HubWallet{HubWalletCaller: HubWalletCaller{contract: contract}, HubWalletTransactor: HubWalletTransactor{contract: contract}}, nil
}

// HubWallet is an auto generated Go binding around an Ethereum contract.
type HubWallet struct {
	HubWalletCaller     // Read-only binding to the contract
	HubWalletTransactor // Write-only binding to the contract
}

// HubWalletCaller is an auto generated read-only Go binding around an Ethereum contract.
type HubWalletCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HubWalletTransactor is an auto generated write-only Go binding around an Ethereum contract.
type HubWalletTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HubWalletSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type HubWalletSession struct {
	Contract     *HubWallet        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HubWalletCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type HubWalletCallerSession struct {
	Contract *HubWalletCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// HubWalletTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type HubWalletTransactorSession struct {
	Contract     *HubWalletTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// HubWalletRaw is an auto generated low-level Go binding around an Ethereum contract.
type HubWalletRaw struct {
	Contract *HubWallet // Generic contract binding to access the raw methods on
}

// HubWalletCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type HubWalletCallerRaw struct {
	Contract *HubWalletCaller // Generic read-only contract binding to access the raw methods on
}

// HubWalletTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type HubWalletTransactorRaw struct {
	Contract *HubWalletTransactor // Generic write-only contract binding to access the raw methods on
}

// NewHubWallet creates a new instance of HubWallet, bound to a specific deployed contract.
func NewHubWallet(address common.Address, backend bind.ContractBackend) (*HubWallet, error) {
	contract, err := bindHubWallet(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &HubWallet{HubWalletCaller: HubWalletCaller{contract: contract}, HubWalletTransactor: HubWalletTransactor{contract: contract}}, nil
}

// NewHubWalletCaller creates a new read-only instance of HubWallet, bound to a specific deployed contract.
func NewHubWalletCaller(address common.Address, caller bind.ContractCaller) (*HubWalletCaller, error) {
	contract, err := bindHubWallet(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &HubWalletCaller{contract: contract}, nil
}

// NewHubWalletTransactor creates a new write-only instance of HubWallet, bound to a specific deployed contract.
func NewHubWalletTransactor(address common.Address, transactor bind.ContractTransactor) (*HubWalletTransactor, error) {
	contract, err := bindHubWallet(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &HubWalletTransactor{contract: contract}, nil
}

// bindHubWallet binds a generic wrapper to an already deployed contract.
func bindHubWallet(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(HubWalletABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HubWallet *HubWalletRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _HubWallet.Contract.HubWalletCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HubWallet *HubWalletRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.Contract.HubWalletTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HubWallet *HubWalletRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HubWallet.Contract.HubWalletTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HubWallet *HubWalletCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _HubWallet.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HubWallet *HubWalletTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HubWallet *HubWalletTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HubWallet.Contract.contract.Transact(opts, method, params...)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_HubWallet *HubWalletCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_HubWallet *HubWalletSession) Owner() (common.Address, error) {
	return _HubWallet.Contract.Owner(&_HubWallet.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_HubWallet *HubWalletCallerSession) Owner() (common.Address, error) {
	return _HubWallet.Contract.Owner(&_HubWallet.CallOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_HubWallet *HubWalletTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_HubWallet *HubWalletSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _HubWallet.Contract.TransferOwnership(&_HubWallet.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_HubWallet *HubWalletTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _HubWallet.Contract.TransferOwnership(&_HubWallet.TransactOpts, newOwner)
}
