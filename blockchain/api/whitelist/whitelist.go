// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package whitelist

import (
	"math/big"
	"strings"

	"github.com/sonm-io/go-ethereum/accounts/abi"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"github.com/sonm-io/go-ethereum/common"
	"github.com/sonm-io/go-ethereum/core/types"
)

// WhitelistABI is the input ABI used to generate the binding from.
const WhitelistABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"wallet\",\"type\":\"address\"}],\"name\":\"UnRegisterHub\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"wallet\",\"type\":\"address\"},{\"name\":\"time\",\"type\":\"uint64\"},{\"name\":\"stakeShare\",\"type\":\"uint256\"}],\"name\":\"RegisterMin\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"wallet\",\"type\":\"address\"}],\"name\":\"UnRegisterMiner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"wallet\",\"type\":\"address\"},{\"name\":\"time\",\"type\":\"uint64\"}],\"name\":\"RegisterHub\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"}]"

// WhitelistBin is the compiled bytecode used for deploying new contracts.
const WhitelistBin = `0x`

// DeployWhitelist deploys a new Ethereum contract, binding an instance of Whitelist to it.
func DeployWhitelist(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Whitelist, error) {
	parsed, err := abi.JSON(strings.NewReader(WhitelistABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(WhitelistBin), backend)
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

// RegisterHub is a paid mutator transaction binding the contract method 0x94baf301.
//
// Solidity: function RegisterHub(_owner address, wallet address, time uint64) returns(bool)
func (_Whitelist *WhitelistTransactor) RegisterHub(opts *bind.TransactOpts, _owner common.Address, wallet common.Address, time uint64) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "RegisterHub", _owner, wallet, time)
}

// RegisterHub is a paid mutator transaction binding the contract method 0x94baf301.
//
// Solidity: function RegisterHub(_owner address, wallet address, time uint64) returns(bool)
func (_Whitelist *WhitelistSession) RegisterHub(_owner common.Address, wallet common.Address, time uint64) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterHub(&_Whitelist.TransactOpts, _owner, wallet, time)
}

// RegisterHub is a paid mutator transaction binding the contract method 0x94baf301.
//
// Solidity: function RegisterHub(_owner address, wallet address, time uint64) returns(bool)
func (_Whitelist *WhitelistTransactorSession) RegisterHub(_owner common.Address, wallet common.Address, time uint64) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterHub(&_Whitelist.TransactOpts, _owner, wallet, time)
}

// RegisterMin is a paid mutator transaction binding the contract method 0x2edbb4fa.
//
// Solidity: function RegisterMin(_owner address, wallet address, time uint64, stakeShare uint256) returns(bool)
func (_Whitelist *WhitelistTransactor) RegisterMin(opts *bind.TransactOpts, _owner common.Address, wallet common.Address, time uint64, stakeShare *big.Int) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "RegisterMin", _owner, wallet, time, stakeShare)
}

// RegisterMin is a paid mutator transaction binding the contract method 0x2edbb4fa.
//
// Solidity: function RegisterMin(_owner address, wallet address, time uint64, stakeShare uint256) returns(bool)
func (_Whitelist *WhitelistSession) RegisterMin(_owner common.Address, wallet common.Address, time uint64, stakeShare *big.Int) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterMin(&_Whitelist.TransactOpts, _owner, wallet, time, stakeShare)
}

// RegisterMin is a paid mutator transaction binding the contract method 0x2edbb4fa.
//
// Solidity: function RegisterMin(_owner address, wallet address, time uint64, stakeShare uint256) returns(bool)
func (_Whitelist *WhitelistTransactorSession) RegisterMin(_owner common.Address, wallet common.Address, time uint64, stakeShare *big.Int) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterMin(&_Whitelist.TransactOpts, _owner, wallet, time, stakeShare)
}

// UnRegisterHub is a paid mutator transaction binding the contract method 0x0bdf0962.
//
// Solidity: function UnRegisterHub(_owner address, wallet address) returns(bool)
func (_Whitelist *WhitelistTransactor) UnRegisterHub(opts *bind.TransactOpts, _owner common.Address, wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "UnRegisterHub", _owner, wallet)
}

// UnRegisterHub is a paid mutator transaction binding the contract method 0x0bdf0962.
//
// Solidity: function UnRegisterHub(_owner address, wallet address) returns(bool)
func (_Whitelist *WhitelistSession) UnRegisterHub(_owner common.Address, wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterHub(&_Whitelist.TransactOpts, _owner, wallet)
}

// UnRegisterHub is a paid mutator transaction binding the contract method 0x0bdf0962.
//
// Solidity: function UnRegisterHub(_owner address, wallet address) returns(bool)
func (_Whitelist *WhitelistTransactorSession) UnRegisterHub(_owner common.Address, wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterHub(&_Whitelist.TransactOpts, _owner, wallet)
}

// UnRegisterMiner is a paid mutator transaction binding the contract method 0x7efad052.
//
// Solidity: function UnRegisterMiner(_owner address, wallet address) returns(bool)
func (_Whitelist *WhitelistTransactor) UnRegisterMiner(opts *bind.TransactOpts, _owner common.Address, wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "UnRegisterMiner", _owner, wallet)
}

// UnRegisterMiner is a paid mutator transaction binding the contract method 0x7efad052.
//
// Solidity: function UnRegisterMiner(_owner address, wallet address) returns(bool)
func (_Whitelist *WhitelistSession) UnRegisterMiner(_owner common.Address, wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterMiner(&_Whitelist.TransactOpts, _owner, wallet)
}

// UnRegisterMiner is a paid mutator transaction binding the contract method 0x7efad052.
//
// Solidity: function UnRegisterMiner(_owner address, wallet address) returns(bool)
func (_Whitelist *WhitelistTransactorSession) UnRegisterMiner(_owner common.Address, wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterMiner(&_Whitelist.TransactOpts, _owner, wallet)
}
