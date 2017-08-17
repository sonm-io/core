// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package api

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// WhitelistABI is the input ABI used to generate the binding from.
const WhitelistABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"time\",\"type\":\"uint64\"},{\"name\":\"stakeShare\",\"type\":\"uint256\"}],\"name\":\"RegisterMin\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"time\",\"type\":\"uint64\"}],\"name\":\"RegisterHub\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"UnRegisterHub\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"RegistredHubs\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"RegistredMiners\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"UnRegisterMiner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"inputs\":[{\"name\":\"_factory\",\"type\":\"address\"}],\"payable\":false,\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"wallet\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"time\",\"type\":\"uint64\"}],\"name\":\"RegistredHub\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"wallet\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"time\",\"type\":\"uint64\"},{\"indexed\":true,\"name\":\"stake\",\"type\":\"uint256\"}],\"name\":\"RegistredMiner\",\"type\":\"event\"}]"

// WhitelistBin is the compiled bytecode used for deploying new contracts.
const WhitelistBin = `0x6060604052341561000f57600080fd5b604051602080610605833981016040528080519150505b60008054600160a060020a031916600160a060020a0383161790555b505b6105b2806100536000396000f3006060604052361561005c5763ffffffff60e060020a6000350416630941d2e681146100615780633f05852c146100a35780634a5de165146100e25780634f89b674146101155780636e61cce314610148578063aaf22fe41461017b575b600080fd5b341561006c57600080fd5b61008f600160a060020a03600435166001604060020a03602435166044356101ae565b604051901515815260200160405180910390f35b34156100ae57600080fd5b61008f600160a060020a03600435166001604060020a03602435166102c7565b604051901515815260200160405180910390f35b34156100ed57600080fd5b61008f600160a060020a03600435166103de565b604051901515815260200160405180910390f35b341561012057600080fd5b61008f600160a060020a036004351661049d565b604051901515815260200160405180910390f35b341561015357600080fd5b61008f600160a060020a03600435166104b2565b604051901515815260200160405180910390f35b341561018657600080fd5b61008f600160a060020a03600435166104c7565b604051901515815260200160405180910390f35b600080548190600160a060020a031663eca939d686836040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b151561020957600080fd5b6102c65a03f1151561021a57600080fd5b50505060405180519050905033600160a060020a031681600160a060020a031614151561024657600080fd5b600160a060020a0380821660009081526002602052604090819020805460ff1916600117905584916001604060020a03871691908816907f01efd282f1b6e0b9bf41a06d831c8b246f0a6a650a848076129723831b2aa00d90859051600160a060020a03909116815260200160405180910390a4600191505b509392505050565b600080548190600160a060020a0316634b72831c85836040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b151561032257600080fd5b6102c65a03f1151561033357600080fd5b50505060405180519050905033600160a060020a031681600160a060020a031614151561035f57600080fd5b600160a060020a03808216600090815260016020819052604091829020805460ff191690911790556001604060020a038516918616907f4b639bce86a9292c37187d0b099958f116ccc27417075d090b204104fa18b70190849051600160a060020a03909116815260200160405180910390a3600191505b5092915050565b600080548190600160a060020a0316634b72831c84836040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b151561043957600080fd5b6102c65a03f1151561044a57600080fd5b50505060405180519050905033600160a060020a031681600160a060020a031614151561047657600080fd5b600160a060020a0381166000908152600160205260409020805460ff191690555b50919050565b60016020526000908152604090205460ff1681565b60026020526000908152604090205460ff1681565b600080548190600160a060020a031663eca939d684836040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b151561052257600080fd5b6102c65a03f1151561053357600080fd5b50505060405180519050905033600160a060020a031681600160a060020a031614151561055f57600080fd5b600160a060020a0381166000908152600260205260409020805460ff191690555b509190505600a165627a7a72305820841252a4bde84ea607aa0b258b9462263429d666627a11926762496dc0f5c5bb0029`

// DeployWhitelist deploys a new Ethereum contract, binding an instance of Whitelist to it.
func DeployWhitelist(auth *bind.TransactOpts, backend bind.ContractBackend, _factory common.Address) (common.Address, *types.Transaction, *Whitelist, error) {
	parsed, err := abi.JSON(strings.NewReader(WhitelistABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(WhitelistBin), backend, _factory)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Whitelist{WhitelistCaller: WhitelistCaller{contract: contract}, WhitelistTransactor: WhitelistTransactor{contract: contract}}, nil
}

// Whitelist is an auto generated Go binding around an Ethereum contract.
type Whitelist struct {
	WhitelistCaller     // Read-only binding to the contract
	WhitelistTransactor // Write-only binding to the contract
}

// WhitelistCaller is an auto generated read-only Go binding around an Ethereum contract.
type WhitelistCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WhitelistTransactor is an auto generated write-only Go binding around an Ethereum contract.
type WhitelistTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WhitelistSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type WhitelistSession struct {
	Contract     *Whitelist        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// WhitelistCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type WhitelistCallerSession struct {
	Contract *WhitelistCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// WhitelistTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type WhitelistTransactorSession struct {
	Contract     *WhitelistTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// WhitelistRaw is an auto generated low-level Go binding around an Ethereum contract.
type WhitelistRaw struct {
	Contract *Whitelist // Generic contract binding to access the raw methods on
}

// WhitelistCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type WhitelistCallerRaw struct {
	Contract *WhitelistCaller // Generic read-only contract binding to access the raw methods on
}

// WhitelistTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type WhitelistTransactorRaw struct {
	Contract *WhitelistTransactor // Generic write-only contract binding to access the raw methods on
}

// NewWhitelist creates a new instance of Whitelist, bound to a specific deployed contract.
func NewWhitelist(address common.Address, backend bind.ContractBackend) (*Whitelist, error) {
	contract, err := bindWhitelist(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Whitelist{WhitelistCaller: WhitelistCaller{contract: contract}, WhitelistTransactor: WhitelistTransactor{contract: contract}}, nil
}

// NewWhitelistCaller creates a new read-only instance of Whitelist, bound to a specific deployed contract.
func NewWhitelistCaller(address common.Address, caller bind.ContractCaller) (*WhitelistCaller, error) {
	contract, err := bindWhitelist(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &WhitelistCaller{contract: contract}, nil
}

// NewWhitelistTransactor creates a new write-only instance of Whitelist, bound to a specific deployed contract.
func NewWhitelistTransactor(address common.Address, transactor bind.ContractTransactor) (*WhitelistTransactor, error) {
	contract, err := bindWhitelist(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &WhitelistTransactor{contract: contract}, nil
}

// bindWhitelist binds a generic wrapper to an already deployed contract.
func bindWhitelist(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(WhitelistABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Whitelist *WhitelistRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Whitelist.Contract.WhitelistCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Whitelist *WhitelistRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Whitelist.Contract.WhitelistTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Whitelist *WhitelistRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Whitelist.Contract.WhitelistTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Whitelist *WhitelistCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Whitelist.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Whitelist *WhitelistTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Whitelist.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Whitelist *WhitelistTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Whitelist.Contract.contract.Transact(opts, method, params...)
}

// RegistredHubs is a free data retrieval call binding the contract method 0x4f89b674.
//
// Solidity: function RegistredHubs( address) constant returns(bool)
func (_Whitelist *WhitelistCaller) RegistredHubs(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Whitelist.contract.Call(opts, out, "RegistredHubs", arg0)
	return *ret0, err
}

// RegistredHubs is a free data retrieval call binding the contract method 0x4f89b674.
//
// Solidity: function RegistredHubs( address) constant returns(bool)
func (_Whitelist *WhitelistSession) RegistredHubs(arg0 common.Address) (bool, error) {
	return _Whitelist.Contract.RegistredHubs(&_Whitelist.CallOpts, arg0)
}

// RegistredHubs is a free data retrieval call binding the contract method 0x4f89b674.
//
// Solidity: function RegistredHubs( address) constant returns(bool)
func (_Whitelist *WhitelistCallerSession) RegistredHubs(arg0 common.Address) (bool, error) {
	return _Whitelist.Contract.RegistredHubs(&_Whitelist.CallOpts, arg0)
}

// RegistredMiners is a free data retrieval call binding the contract method 0x6e61cce3.
//
// Solidity: function RegistredMiners( address) constant returns(bool)
func (_Whitelist *WhitelistCaller) RegistredMiners(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Whitelist.contract.Call(opts, out, "RegistredMiners", arg0)
	return *ret0, err
}

// RegistredMiners is a free data retrieval call binding the contract method 0x6e61cce3.
//
// Solidity: function RegistredMiners( address) constant returns(bool)
func (_Whitelist *WhitelistSession) RegistredMiners(arg0 common.Address) (bool, error) {
	return _Whitelist.Contract.RegistredMiners(&_Whitelist.CallOpts, arg0)
}

// RegistredMiners is a free data retrieval call binding the contract method 0x6e61cce3.
//
// Solidity: function RegistredMiners( address) constant returns(bool)
func (_Whitelist *WhitelistCallerSession) RegistredMiners(arg0 common.Address) (bool, error) {
	return _Whitelist.Contract.RegistredMiners(&_Whitelist.CallOpts, arg0)
}

// RegisterHub is a paid mutator transaction binding the contract method 0x3f05852c.
//
// Solidity: function RegisterHub(_owner address, time uint64) returns(bool)
func (_Whitelist *WhitelistTransactor) RegisterHub(opts *bind.TransactOpts, _owner common.Address, time uint64) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "RegisterHub", _owner, time)
}

// RegisterHub is a paid mutator transaction binding the contract method 0x3f05852c.
//
// Solidity: function RegisterHub(_owner address, time uint64) returns(bool)
func (_Whitelist *WhitelistSession) RegisterHub(_owner common.Address, time uint64) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterHub(&_Whitelist.TransactOpts, _owner, time)
}

// RegisterHub is a paid mutator transaction binding the contract method 0x3f05852c.
//
// Solidity: function RegisterHub(_owner address, time uint64) returns(bool)
func (_Whitelist *WhitelistTransactorSession) RegisterHub(_owner common.Address, time uint64) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterHub(&_Whitelist.TransactOpts, _owner, time)
}

// RegisterMin is a paid mutator transaction binding the contract method 0x0941d2e6.
//
// Solidity: function RegisterMin(_owner address, time uint64, stakeShare uint256) returns(bool)
func (_Whitelist *WhitelistTransactor) RegisterMin(opts *bind.TransactOpts, _owner common.Address, time uint64, stakeShare *big.Int) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "RegisterMin", _owner, time, stakeShare)
}

// RegisterMin is a paid mutator transaction binding the contract method 0x0941d2e6.
//
// Solidity: function RegisterMin(_owner address, time uint64, stakeShare uint256) returns(bool)
func (_Whitelist *WhitelistSession) RegisterMin(_owner common.Address, time uint64, stakeShare *big.Int) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterMin(&_Whitelist.TransactOpts, _owner, time, stakeShare)
}

// RegisterMin is a paid mutator transaction binding the contract method 0x0941d2e6.
//
// Solidity: function RegisterMin(_owner address, time uint64, stakeShare uint256) returns(bool)
func (_Whitelist *WhitelistTransactorSession) RegisterMin(_owner common.Address, time uint64, stakeShare *big.Int) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterMin(&_Whitelist.TransactOpts, _owner, time, stakeShare)
}

// UnRegisterHub is a paid mutator transaction binding the contract method 0x4a5de165.
//
// Solidity: function UnRegisterHub(_owner address) returns(bool)
func (_Whitelist *WhitelistTransactor) UnRegisterHub(opts *bind.TransactOpts, _owner common.Address) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "UnRegisterHub", _owner)
}

// UnRegisterHub is a paid mutator transaction binding the contract method 0x4a5de165.
//
// Solidity: function UnRegisterHub(_owner address) returns(bool)
func (_Whitelist *WhitelistSession) UnRegisterHub(_owner common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterHub(&_Whitelist.TransactOpts, _owner)
}

// UnRegisterHub is a paid mutator transaction binding the contract method 0x4a5de165.
//
// Solidity: function UnRegisterHub(_owner address) returns(bool)
func (_Whitelist *WhitelistTransactorSession) UnRegisterHub(_owner common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterHub(&_Whitelist.TransactOpts, _owner)
}

// UnRegisterMiner is a paid mutator transaction binding the contract method 0xaaf22fe4.
//
// Solidity: function UnRegisterMiner(_owner address) returns(bool)
func (_Whitelist *WhitelistTransactor) UnRegisterMiner(opts *bind.TransactOpts, _owner common.Address) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "UnRegisterMiner", _owner)
}

// UnRegisterMiner is a paid mutator transaction binding the contract method 0xaaf22fe4.
//
// Solidity: function UnRegisterMiner(_owner address) returns(bool)
func (_Whitelist *WhitelistSession) UnRegisterMiner(_owner common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterMiner(&_Whitelist.TransactOpts, _owner)
}

// UnRegisterMiner is a paid mutator transaction binding the contract method 0xaaf22fe4.
//
// Solidity: function UnRegisterMiner(_owner address) returns(bool)
func (_Whitelist *WhitelistTransactorSession) UnRegisterMiner(_owner common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterMiner(&_Whitelist.TransactOpts, _owner)
}
