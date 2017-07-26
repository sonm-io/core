// This file is an automatically generated Go binding. Do not modify as any
// change will likely be lost upon the next re-generation!

package Whitelist

import (
	"math/big"
	"strings"

	"github.com/sonm-io/go-ethereum/accounts/abi"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"github.com/sonm-io/go-ethereum/common"
	"github.com/sonm-io/go-ethereum/core/types"
)

// WhitelistABI is the input ABI used to generate the binding from.
const WhitelistABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_wallet\",\"type\":\"address\"}],\"name\":\"UnRegisterHub\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_wallet\",\"type\":\"address\"},{\"name\":\"time\",\"type\":\"uint64\"},{\"name\":\"stakeShare\",\"type\":\"uint256\"}],\"name\":\"RegisterMin\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"RegistredHubs\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"RegistredMiners\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_wallet\",\"type\":\"address\"}],\"name\":\"UnRegisterMiner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_wallet\",\"type\":\"address\"},{\"name\":\"time\",\"type\":\"uint64\"}],\"name\":\"RegisterHub\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"inputs\":[{\"name\":\"Factory\",\"type\":\"address\"}],\"payable\":false,\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"wallet\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"time\",\"type\":\"uint64\"}],\"name\":\"RegistredHub\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"wallet\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"time\",\"type\":\"uint64\"},{\"indexed\":true,\"name\":\"stake\",\"type\":\"uint256\"}],\"name\":\"RegistredMiner\",\"type\":\"event\"}]"

// WhitelistBin is the compiled bytecode used for deploying new contracts.
const WhitelistBin = `0x6060604052341561000c57fe5b60405160208061062683398101604052515b60008054600160a060020a031916600160a060020a0383161790555b505b6105db8061004b6000396000f300606060405236156100755763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416630bdf096281146100775780632edbb4fa146100ad5780634f89b674146100f35780636e61cce3146101235780637efad0521461015357806394baf30114610189575bfe5b341561007f57fe5b610099600160a060020a03600435811690602435166101cc565b604080519115158252519081900360200190f35b34156100b557fe5b610099600160a060020a036004358116906024351667ffffffffffffffff60443516606435610293565b604080519115158252519081900360200190f35b34156100fb57fe5b610099600160a060020a03600435166103aa565b604080519115158252519081900360200190f35b341561012b57fe5b610099600160a060020a03600435166103bf565b604080519115158252519081900360200190f35b341561015b57fe5b610099600160a060020a03600435811690602435166103d4565b604080519115158252519081900360200190f35b341561019157fe5b610099600160a060020a036004358116906024351667ffffffffffffffff6044351661049b565b604080519115158252519081900360200190f35b6000805460408051602090810184905281517f4b72831c000000000000000000000000000000000000000000000000000000008152600160a060020a038781166004830152925185949390931692634b72831c9260248084019391929182900301818787803b151561023a57fe5b6102c65a03f1151561024857fe5b50506040515191505033600160a060020a039081169082161461026b5760006000fd5b600160a060020a0381166000908152600160205260409020805460ff191690555b5092915050565b6000805460408051602090810184905281517feca939d6000000000000000000000000000000000000000000000000000000008152600160a060020a03898116600483015292518594939093169263eca939d69260248084019391929182900301818787803b151561030157fe5b6102c65a03f1151561030f57fe5b50506040515191505033600160a060020a03908116908216146103325760006000fd5b600160a060020a03808216600081815260026020908152604091829020805460ff1916600117905581519283529051869367ffffffffffffffff891693908b16927f01efd282f1b6e0b9bf41a06d831c8b246f0a6a650a848076129723831b2aa00d92918290030190a4600191505b50949350505050565b60016020526000908152604090205460ff1681565b60026020526000908152604090205460ff1681565b6000805460408051602090810184905281517feca939d6000000000000000000000000000000000000000000000000000000008152600160a060020a03878116600483015292518594939093169263eca939d69260248084019391929182900301818787803b151561044257fe5b6102c65a03f1151561045057fe5b50506040515191505033600160a060020a03908116908216146104735760006000fd5b600160a060020a0381166000908152600260205260409020805460ff191690555b5092915050565b6000805460408051602090810184905281517f4b72831c000000000000000000000000000000000000000000000000000000008152600160a060020a038881166004830152925185949390931692634b72831c9260248084019391929182900301818787803b151561050957fe5b6102c65a03f1151561051757fe5b50506040515191505033600160a060020a039081169082161461053a5760006000fd5b600160a060020a03808216600081815260016020818152604092839020805460ff19169092179091558151928352905167ffffffffffffffff8716938916927f4b639bce86a9292c37187d0b099958f116ccc27417075d090b204104fa18b70192908290030190a3600191505b5093925050505600a165627a7a7230582014b1134525b821d61b8e06f8ff543ec560795880fd5e50f74bd689718af9fefe0029`

// DeployWhitelist deploys a new Ethereum contract, binding an instance of Whitelist to it.
func DeployWhitelist(auth *bind.TransactOpts, backend bind.ContractBackend, Factory common.Address) (common.Address, *types.Transaction, *Whitelist, error) {
	parsed, err := abi.JSON(strings.NewReader(WhitelistABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(WhitelistBin), backend, Factory)
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

// RegisterHub is a paid mutator transaction binding the contract method 0x94baf301.
//
// Solidity: function RegisterHub(_owner address, _wallet address, time uint64) returns(bool)
func (_Whitelist *WhitelistTransactor) RegisterHub(opts *bind.TransactOpts, _owner common.Address, _wallet common.Address, time uint64) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "RegisterHub", _owner, _wallet, time)
}

// RegisterHub is a paid mutator transaction binding the contract method 0x94baf301.
//
// Solidity: function RegisterHub(_owner address, _wallet address, time uint64) returns(bool)
func (_Whitelist *WhitelistSession) RegisterHub(_owner common.Address, _wallet common.Address, time uint64) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterHub(&_Whitelist.TransactOpts, _owner, _wallet, time)
}

// RegisterHub is a paid mutator transaction binding the contract method 0x94baf301.
//
// Solidity: function RegisterHub(_owner address, _wallet address, time uint64) returns(bool)
func (_Whitelist *WhitelistTransactorSession) RegisterHub(_owner common.Address, _wallet common.Address, time uint64) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterHub(&_Whitelist.TransactOpts, _owner, _wallet, time)
}

// RegisterMin is a paid mutator transaction binding the contract method 0x2edbb4fa.
//
// Solidity: function RegisterMin(_owner address, _wallet address, time uint64, stakeShare uint256) returns(bool)
func (_Whitelist *WhitelistTransactor) RegisterMin(opts *bind.TransactOpts, _owner common.Address, _wallet common.Address, time uint64, stakeShare *big.Int) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "RegisterMin", _owner, _wallet, time, stakeShare)
}

// RegisterMin is a paid mutator transaction binding the contract method 0x2edbb4fa.
//
// Solidity: function RegisterMin(_owner address, _wallet address, time uint64, stakeShare uint256) returns(bool)
func (_Whitelist *WhitelistSession) RegisterMin(_owner common.Address, _wallet common.Address, time uint64, stakeShare *big.Int) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterMin(&_Whitelist.TransactOpts, _owner, _wallet, time, stakeShare)
}

// RegisterMin is a paid mutator transaction binding the contract method 0x2edbb4fa.
//
// Solidity: function RegisterMin(_owner address, _wallet address, time uint64, stakeShare uint256) returns(bool)
func (_Whitelist *WhitelistTransactorSession) RegisterMin(_owner common.Address, _wallet common.Address, time uint64, stakeShare *big.Int) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterMin(&_Whitelist.TransactOpts, _owner, _wallet, time, stakeShare)
}

// UnRegisterHub is a paid mutator transaction binding the contract method 0x0bdf0962.
//
// Solidity: function UnRegisterHub(_owner address, _wallet address) returns(bool)
func (_Whitelist *WhitelistTransactor) UnRegisterHub(opts *bind.TransactOpts, _owner common.Address, _wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "UnRegisterHub", _owner, _wallet)
}

// UnRegisterHub is a paid mutator transaction binding the contract method 0x0bdf0962.
//
// Solidity: function UnRegisterHub(_owner address, _wallet address) returns(bool)
func (_Whitelist *WhitelistSession) UnRegisterHub(_owner common.Address, _wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterHub(&_Whitelist.TransactOpts, _owner, _wallet)
}

// UnRegisterHub is a paid mutator transaction binding the contract method 0x0bdf0962.
//
// Solidity: function UnRegisterHub(_owner address, _wallet address) returns(bool)
func (_Whitelist *WhitelistTransactorSession) UnRegisterHub(_owner common.Address, _wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterHub(&_Whitelist.TransactOpts, _owner, _wallet)
}

// UnRegisterMiner is a paid mutator transaction binding the contract method 0x7efad052.
//
// Solidity: function UnRegisterMiner(_owner address, _wallet address) returns(bool)
func (_Whitelist *WhitelistTransactor) UnRegisterMiner(opts *bind.TransactOpts, _owner common.Address, _wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "UnRegisterMiner", _owner, _wallet)
}

// UnRegisterMiner is a paid mutator transaction binding the contract method 0x7efad052.
//
// Solidity: function UnRegisterMiner(_owner address, _wallet address) returns(bool)
func (_Whitelist *WhitelistSession) UnRegisterMiner(_owner common.Address, _wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterMiner(&_Whitelist.TransactOpts, _owner, _wallet)
}

// UnRegisterMiner is a paid mutator transaction binding the contract method 0x7efad052.
//
// Solidity: function UnRegisterMiner(_owner address, _wallet address) returns(bool)
func (_Whitelist *WhitelistTransactorSession) UnRegisterMiner(_owner common.Address, _wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterMiner(&_Whitelist.TransactOpts, _owner, _wallet)
}

// FactoryABI is the input ABI used to generate the binding from.
const FactoryABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"HubOf\",\"outputs\":[{\"name\":\"_wallet\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"miners\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"hubs\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"MinerOf\",\"outputs\":[{\"name\":\"_wallet\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"}]"

// FactoryBin is the compiled bytecode used for deploying new contracts.
const FactoryBin = `0x`

// DeployFactory deploys a new Ethereum contract, binding an instance of Factory to it.
func DeployFactory(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Factory, error) {
	parsed, err := abi.JSON(strings.NewReader(FactoryABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(FactoryBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Factory{FactoryCaller: FactoryCaller{contract: contract}, FactoryTransactor: FactoryTransactor{contract: contract}}, nil
}

// Factory is an auto generated Go binding around an Ethereum contract.
type Factory struct {
	FactoryCaller     // Read-only binding to the contract
	FactoryTransactor // Write-only binding to the contract
}

// FactoryCaller is an auto generated read-only Go binding around an Ethereum contract.
type FactoryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FactoryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FactoryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FactorySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FactorySession struct {
	Contract     *Factory          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// FactoryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FactoryCallerSession struct {
	Contract *FactoryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// FactoryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FactoryTransactorSession struct {
	Contract     *FactoryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// FactoryRaw is an auto generated low-level Go binding around an Ethereum contract.
type FactoryRaw struct {
	Contract *Factory // Generic contract binding to access the raw methods on
}

// FactoryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FactoryCallerRaw struct {
	Contract *FactoryCaller // Generic read-only contract binding to access the raw methods on
}

// FactoryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FactoryTransactorRaw struct {
	Contract *FactoryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFactory creates a new instance of Factory, bound to a specific deployed contract.
func NewFactory(address common.Address, backend bind.ContractBackend) (*Factory, error) {
	contract, err := bindFactory(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Factory{FactoryCaller: FactoryCaller{contract: contract}, FactoryTransactor: FactoryTransactor{contract: contract}}, nil
}

// NewFactoryCaller creates a new read-only instance of Factory, bound to a specific deployed contract.
func NewFactoryCaller(address common.Address, caller bind.ContractCaller) (*FactoryCaller, error) {
	contract, err := bindFactory(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &FactoryCaller{contract: contract}, nil
}

// NewFactoryTransactor creates a new write-only instance of Factory, bound to a specific deployed contract.
func NewFactoryTransactor(address common.Address, transactor bind.ContractTransactor) (*FactoryTransactor, error) {
	contract, err := bindFactory(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &FactoryTransactor{contract: contract}, nil
}

// bindFactory binds a generic wrapper to an already deployed contract.
func bindFactory(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(FactoryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Factory *FactoryRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Factory.Contract.FactoryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Factory *FactoryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Factory.Contract.FactoryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Factory *FactoryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Factory.Contract.FactoryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Factory *FactoryCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Factory.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Factory *FactoryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Factory.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Factory *FactoryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Factory.Contract.contract.Transact(opts, method, params...)
}

// HubOf is a free data retrieval call binding the contract method 0x4b72831c.
//
// Solidity: function HubOf(_owner address) constant returns(_wallet address)
func (_Factory *FactoryCaller) HubOf(opts *bind.CallOpts, _owner common.Address) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Factory.contract.Call(opts, out, "HubOf", _owner)
	return *ret0, err
}

// HubOf is a free data retrieval call binding the contract method 0x4b72831c.
//
// Solidity: function HubOf(_owner address) constant returns(_wallet address)
func (_Factory *FactorySession) HubOf(_owner common.Address) (common.Address, error) {
	return _Factory.Contract.HubOf(&_Factory.CallOpts, _owner)
}

// HubOf is a free data retrieval call binding the contract method 0x4b72831c.
//
// Solidity: function HubOf(_owner address) constant returns(_wallet address)
func (_Factory *FactoryCallerSession) HubOf(_owner common.Address) (common.Address, error) {
	return _Factory.Contract.HubOf(&_Factory.CallOpts, _owner)
}

// MinerOf is a free data retrieval call binding the contract method 0xeca939d6.
//
// Solidity: function MinerOf(_owner address) constant returns(_wallet address)
func (_Factory *FactoryCaller) MinerOf(opts *bind.CallOpts, _owner common.Address) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Factory.contract.Call(opts, out, "MinerOf", _owner)
	return *ret0, err
}

// MinerOf is a free data retrieval call binding the contract method 0xeca939d6.
//
// Solidity: function MinerOf(_owner address) constant returns(_wallet address)
func (_Factory *FactorySession) MinerOf(_owner common.Address) (common.Address, error) {
	return _Factory.Contract.MinerOf(&_Factory.CallOpts, _owner)
}

// MinerOf is a free data retrieval call binding the contract method 0xeca939d6.
//
// Solidity: function MinerOf(_owner address) constant returns(_wallet address)
func (_Factory *FactoryCallerSession) MinerOf(_owner common.Address) (common.Address, error) {
	return _Factory.Contract.MinerOf(&_Factory.CallOpts, _owner)
}

// Hubs is a free data retrieval call binding the contract method 0xca3106ae.
//
// Solidity: function hubs( address) constant returns(address)
func (_Factory *FactoryCaller) Hubs(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Factory.contract.Call(opts, out, "hubs", arg0)
	return *ret0, err
}

// Hubs is a free data retrieval call binding the contract method 0xca3106ae.
//
// Solidity: function hubs( address) constant returns(address)
func (_Factory *FactorySession) Hubs(arg0 common.Address) (common.Address, error) {
	return _Factory.Contract.Hubs(&_Factory.CallOpts, arg0)
}

// Hubs is a free data retrieval call binding the contract method 0xca3106ae.
//
// Solidity: function hubs( address) constant returns(address)
func (_Factory *FactoryCallerSession) Hubs(arg0 common.Address) (common.Address, error) {
	return _Factory.Contract.Hubs(&_Factory.CallOpts, arg0)
}

// Miners is a free data retrieval call binding the contract method 0x648ec7b9.
//
// Solidity: function miners( address) constant returns(address)
func (_Factory *FactoryCaller) Miners(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Factory.contract.Call(opts, out, "miners", arg0)
	return *ret0, err
}

// Miners is a free data retrieval call binding the contract method 0x648ec7b9.
//
// Solidity: function miners( address) constant returns(address)
func (_Factory *FactorySession) Miners(arg0 common.Address) (common.Address, error) {
	return _Factory.Contract.Miners(&_Factory.CallOpts, arg0)
}

// Miners is a free data retrieval call binding the contract method 0x648ec7b9.
//
// Solidity: function miners( address) constant returns(address)
func (_Factory *FactoryCallerSession) Miners(arg0 common.Address) (common.Address, error) {
	return _Factory.Contract.Miners(&_Factory.CallOpts, arg0)
}
