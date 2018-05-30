// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package api

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// SNMTTokenABI is the input ABI used to generate the binding from.
const SNMTTokenABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"remaining\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"whom\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"GiveAway\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"target\",\"type\":\"address\"},{\"name\":\"mintedAmount\",\"type\":\"uint256\"}],\"name\":\"mintToken\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"getTokens\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// SNMTTokenBin is the compiled bytecode used for deploying new contracts.
const SNMTTokenBin = `0x60c0604052600f60808190527f534f4e4d207465737420746f6b656e000000000000000000000000000000000060a090815261003e91600491906100b4565b506040805180820190915260048082527f534e4d54000000000000000000000000000000000000000000000000000000006020909201918252610083916005916100b4565b50601260065534801561009557600080fd5b506003805433600160a060020a0319918216811790911617905561014f565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106100f557805160ff1916838001178555610122565b82800160010185558215610122579182015b82811115610122578251825591602001919060010190610107565b5061012e929150610132565b5090565b61014c91905b8082111561012e5760008155600101610138565b90565b6108a88061015e6000396000f3006080604052600436106100c45763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306fdde0381146100cf578063095ea7b31461015957806318160ddd1461019157806323b872dd146101b8578063313ce567146101e257806370a08231146101f757806379c65068146102185780638da5cb5b1461023c57806395d89b411461026d578063a9059cbb14610282578063aa6ca808146102a6578063dd62ed3e146102bb578063f2fde38b146102e2575b6100cc610305565b50005b3480156100db57600080fd5b506100e461039a565b6040805160208082528351818301528351919283929083019185019080838360005b8381101561011e578181015183820152602001610106565b50505050905090810190601f16801561014b5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561016557600080fd5b5061017d600160a060020a0360043516602435610428565b604080519115158252519081900360200190f35b34801561019d57600080fd5b506101a66104ca565b60408051918252519081900360200190f35b3480156101c457600080fd5b5061017d600160a060020a03600435811690602435166044356104d0565b3480156101ee57600080fd5b506101a66105df565b34801561020357600080fd5b506101a6600160a060020a03600435166105e5565b34801561022457600080fd5b5061017d600160a060020a0360043516602435610600565b34801561024857600080fd5b506102516106bd565b60408051600160a060020a039092168252519081900360200190f35b34801561027957600080fd5b506100e46106cc565b34801561028e57600080fd5b5061017d600160a060020a0360043516602435610727565b3480156102b257600080fd5b5061017d610305565b3480156102c757600080fd5b506101a6600160a060020a03600435811690602435166107d7565b3480156102ee57600080fd5b50610303600160a060020a0360043516610802565b005b3360009081526001602052604081205468056bc75e2d6310000090610330908263ffffffff61085416565b3360009081526001602052604081209190915554610354908263ffffffff61085416565b600055604080513381526020810183905281517fe08e9d066634006283658128ec91f58d444719d7a07d49f72924da4352ff94ad929181900390910190a1600191505090565b6004805460408051602060026001851615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156104205780601f106103f557610100808354040283529160200191610420565b820191906000526020600020905b81548152906001019060200180831161040357829003601f168201915b505050505081565b60008115806104585750336000908152600260209081526040808320600160a060020a0387168452909152902054155b151561046357600080fd5b336000818152600260209081526040808320600160a060020a03881680855290835292819020869055805186815290519293927f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925929181900390910190a350600192915050565b60005481565b600160a060020a03808416600090815260026020908152604080832033845282528083205493861683526001909152812054909190610515908463ffffffff61085416565b600160a060020a03808616600090815260016020526040808220939093559087168152205461054a908463ffffffff61086a16565b600160a060020a038616600090815260016020526040902055610573818463ffffffff61086a16565b600160a060020a03808716600081815260026020908152604080832033845282529182902094909455805187815290519288169391927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef929181900390910190a3506001949350505050565b60065481565b600160a060020a031660009081526001602052604090205490565b600354600090600160a060020a0316331461061a57600080fd5b600160a060020a038316600090815260016020526040902054610643908363ffffffff61085416565b600160a060020a03841660009081526001602052604081209190915554610670908363ffffffff61085416565b60005560408051600160a060020a03851681526020810184905281517fe08e9d066634006283658128ec91f58d444719d7a07d49f72924da4352ff94ad929181900390910190a192915050565b600354600160a060020a031681565b6005805460408051602060026001851615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156104205780601f106103f557610100808354040283529160200191610420565b33600090815260016020526040812054610747908363ffffffff61086a16565b3360009081526001602052604080822092909255600160a060020a03851681522054610779908363ffffffff61085416565b600160a060020a0384166000818152600160209081526040918290209390935580518581529051919233927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9281900390910190a350600192915050565b600160a060020a03918216600090815260026020908152604080832093909416825291909152205490565b600354600160a060020a0316331461081957600080fd5b600160a060020a03811615610851576003805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0383161790555b50565b60008282018381101561086357fe5b9392505050565b60008282111561087657fe5b509003905600a165627a7a72305820ed7a88d1277618494542996bc5eb50a8c71e13b00daf16bfed242817a06595330029`

// DeploySNMTToken deploys a new Ethereum contract, binding an instance of SNMTToken to it.
func DeploySNMTToken(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SNMTToken, error) {
	parsed, err := abi.JSON(strings.NewReader(SNMTTokenABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SNMTTokenBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SNMTToken{SNMTTokenCaller: SNMTTokenCaller{contract: contract}, SNMTTokenTransactor: SNMTTokenTransactor{contract: contract}, SNMTTokenFilterer: SNMTTokenFilterer{contract: contract}}, nil
}

// SNMTToken is an auto generated Go binding around an Ethereum contract.
type SNMTToken struct {
	SNMTTokenCaller     // Read-only binding to the contract
	SNMTTokenTransactor // Write-only binding to the contract
	SNMTTokenFilterer   // Log filterer for contract events
}

// SNMTTokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type SNMTTokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SNMTTokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SNMTTokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SNMTTokenFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SNMTTokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SNMTTokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SNMTTokenSession struct {
	Contract     *SNMTToken        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SNMTTokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SNMTTokenCallerSession struct {
	Contract *SNMTTokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// SNMTTokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SNMTTokenTransactorSession struct {
	Contract     *SNMTTokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// SNMTTokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type SNMTTokenRaw struct {
	Contract *SNMTToken // Generic contract binding to access the raw methods on
}

// SNMTTokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SNMTTokenCallerRaw struct {
	Contract *SNMTTokenCaller // Generic read-only contract binding to access the raw methods on
}

// SNMTTokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SNMTTokenTransactorRaw struct {
	Contract *SNMTTokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSNMTToken creates a new instance of SNMTToken, bound to a specific deployed contract.
func NewSNMTToken(address common.Address, backend bind.ContractBackend) (*SNMTToken, error) {
	contract, err := bindSNMTToken(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SNMTToken{SNMTTokenCaller: SNMTTokenCaller{contract: contract}, SNMTTokenTransactor: SNMTTokenTransactor{contract: contract}, SNMTTokenFilterer: SNMTTokenFilterer{contract: contract}}, nil
}

// NewSNMTTokenCaller creates a new read-only instance of SNMTToken, bound to a specific deployed contract.
func NewSNMTTokenCaller(address common.Address, caller bind.ContractCaller) (*SNMTTokenCaller, error) {
	contract, err := bindSNMTToken(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SNMTTokenCaller{contract: contract}, nil
}

// NewSNMTTokenTransactor creates a new write-only instance of SNMTToken, bound to a specific deployed contract.
func NewSNMTTokenTransactor(address common.Address, transactor bind.ContractTransactor) (*SNMTTokenTransactor, error) {
	contract, err := bindSNMTToken(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SNMTTokenTransactor{contract: contract}, nil
}

// NewSNMTTokenFilterer creates a new log filterer instance of SNMTToken, bound to a specific deployed contract.
func NewSNMTTokenFilterer(address common.Address, filterer bind.ContractFilterer) (*SNMTTokenFilterer, error) {
	contract, err := bindSNMTToken(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SNMTTokenFilterer{contract: contract}, nil
}

// bindSNMTToken binds a generic wrapper to an already deployed contract.
func bindSNMTToken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SNMTTokenABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SNMTToken *SNMTTokenRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SNMTToken.Contract.SNMTTokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SNMTToken *SNMTTokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SNMTToken.Contract.SNMTTokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SNMTToken *SNMTTokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SNMTToken.Contract.SNMTTokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SNMTToken *SNMTTokenCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SNMTToken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SNMTToken *SNMTTokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SNMTToken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SNMTToken *SNMTTokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SNMTToken.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_SNMTToken *SNMTTokenCaller) Allowance(opts *bind.CallOpts, _owner common.Address, _spender common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SNMTToken.contract.Call(opts, out, "allowance", _owner, _spender)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_SNMTToken *SNMTTokenSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _SNMTToken.Contract.Allowance(&_SNMTToken.CallOpts, _owner, _spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_SNMTToken *SNMTTokenCallerSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _SNMTToken.Contract.Allowance(&_SNMTToken.CallOpts, _owner, _spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_SNMTToken *SNMTTokenCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SNMTToken.contract.Call(opts, out, "balanceOf", _owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_SNMTToken *SNMTTokenSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _SNMTToken.Contract.BalanceOf(&_SNMTToken.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_SNMTToken *SNMTTokenCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _SNMTToken.Contract.BalanceOf(&_SNMTToken.CallOpts, _owner)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_SNMTToken *SNMTTokenCaller) Decimals(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SNMTToken.contract.Call(opts, out, "decimals")
	return *ret0, err
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_SNMTToken *SNMTTokenSession) Decimals() (*big.Int, error) {
	return _SNMTToken.Contract.Decimals(&_SNMTToken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_SNMTToken *SNMTTokenCallerSession) Decimals() (*big.Int, error) {
	return _SNMTToken.Contract.Decimals(&_SNMTToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_SNMTToken *SNMTTokenCaller) Name(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _SNMTToken.contract.Call(opts, out, "name")
	return *ret0, err
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_SNMTToken *SNMTTokenSession) Name() (string, error) {
	return _SNMTToken.Contract.Name(&_SNMTToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_SNMTToken *SNMTTokenCallerSession) Name() (string, error) {
	return _SNMTToken.Contract.Name(&_SNMTToken.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SNMTToken *SNMTTokenCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _SNMTToken.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SNMTToken *SNMTTokenSession) Owner() (common.Address, error) {
	return _SNMTToken.Contract.Owner(&_SNMTToken.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SNMTToken *SNMTTokenCallerSession) Owner() (common.Address, error) {
	return _SNMTToken.Contract.Owner(&_SNMTToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_SNMTToken *SNMTTokenCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _SNMTToken.contract.Call(opts, out, "symbol")
	return *ret0, err
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_SNMTToken *SNMTTokenSession) Symbol() (string, error) {
	return _SNMTToken.Contract.Symbol(&_SNMTToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_SNMTToken *SNMTTokenCallerSession) Symbol() (string, error) {
	return _SNMTToken.Contract.Symbol(&_SNMTToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_SNMTToken *SNMTTokenCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SNMTToken.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_SNMTToken *SNMTTokenSession) TotalSupply() (*big.Int, error) {
	return _SNMTToken.Contract.TotalSupply(&_SNMTToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_SNMTToken *SNMTTokenCallerSession) TotalSupply() (*big.Int, error) {
	return _SNMTToken.Contract.TotalSupply(&_SNMTToken.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(bool)
func (_SNMTToken *SNMTTokenTransactor) Approve(opts *bind.TransactOpts, _spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMTToken.contract.Transact(opts, "approve", _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(bool)
func (_SNMTToken *SNMTTokenSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMTToken.Contract.Approve(&_SNMTToken.TransactOpts, _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(bool)
func (_SNMTToken *SNMTTokenTransactorSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMTToken.Contract.Approve(&_SNMTToken.TransactOpts, _spender, _value)
}

// GetTokens is a paid mutator transaction binding the contract method 0xaa6ca808.
//
// Solidity: function getTokens() returns(bool)
func (_SNMTToken *SNMTTokenTransactor) GetTokens(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SNMTToken.contract.Transact(opts, "getTokens")
}

// GetTokens is a paid mutator transaction binding the contract method 0xaa6ca808.
//
// Solidity: function getTokens() returns(bool)
func (_SNMTToken *SNMTTokenSession) GetTokens() (*types.Transaction, error) {
	return _SNMTToken.Contract.GetTokens(&_SNMTToken.TransactOpts)
}

// GetTokens is a paid mutator transaction binding the contract method 0xaa6ca808.
//
// Solidity: function getTokens() returns(bool)
func (_SNMTToken *SNMTTokenTransactorSession) GetTokens() (*types.Transaction, error) {
	return _SNMTToken.Contract.GetTokens(&_SNMTToken.TransactOpts)
}

// MintToken is a paid mutator transaction binding the contract method 0x79c65068.
//
// Solidity: function mintToken(target address, mintedAmount uint256) returns(bool)
func (_SNMTToken *SNMTTokenTransactor) MintToken(opts *bind.TransactOpts, target common.Address, mintedAmount *big.Int) (*types.Transaction, error) {
	return _SNMTToken.contract.Transact(opts, "mintToken", target, mintedAmount)
}

// MintToken is a paid mutator transaction binding the contract method 0x79c65068.
//
// Solidity: function mintToken(target address, mintedAmount uint256) returns(bool)
func (_SNMTToken *SNMTTokenSession) MintToken(target common.Address, mintedAmount *big.Int) (*types.Transaction, error) {
	return _SNMTToken.Contract.MintToken(&_SNMTToken.TransactOpts, target, mintedAmount)
}

// MintToken is a paid mutator transaction binding the contract method 0x79c65068.
//
// Solidity: function mintToken(target address, mintedAmount uint256) returns(bool)
func (_SNMTToken *SNMTTokenTransactorSession) MintToken(target common.Address, mintedAmount *big.Int) (*types.Transaction, error) {
	return _SNMTToken.Contract.MintToken(&_SNMTToken.TransactOpts, target, mintedAmount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(bool)
func (_SNMTToken *SNMTTokenTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMTToken.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(bool)
func (_SNMTToken *SNMTTokenSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMTToken.Contract.Transfer(&_SNMTToken.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(bool)
func (_SNMTToken *SNMTTokenTransactorSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMTToken.Contract.Transfer(&_SNMTToken.TransactOpts, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(bool)
func (_SNMTToken *SNMTTokenTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMTToken.contract.Transact(opts, "transferFrom", _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(bool)
func (_SNMTToken *SNMTTokenSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMTToken.Contract.TransferFrom(&_SNMTToken.TransactOpts, _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(bool)
func (_SNMTToken *SNMTTokenTransactorSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMTToken.Contract.TransferFrom(&_SNMTToken.TransactOpts, _from, _to, _value)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_SNMTToken *SNMTTokenTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SNMTToken.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_SNMTToken *SNMTTokenSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SNMTToken.Contract.TransferOwnership(&_SNMTToken.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_SNMTToken *SNMTTokenTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SNMTToken.Contract.TransferOwnership(&_SNMTToken.TransactOpts, newOwner)
}

// SNMTTokenApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the SNMTToken contract.
type SNMTTokenApprovalIterator struct {
	Event *SNMTTokenApproval // Event containing the contract specifics and raw log

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
func (it *SNMTTokenApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SNMTTokenApproval)
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
		it.Event = new(SNMTTokenApproval)
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
func (it *SNMTTokenApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SNMTTokenApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SNMTTokenApproval represents a Approval event raised by the SNMTToken contract.
type SNMTTokenApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(owner indexed address, spender indexed address, value uint256)
func (_SNMTToken *SNMTTokenFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*SNMTTokenApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _SNMTToken.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &SNMTTokenApprovalIterator{contract: _SNMTToken.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(owner indexed address, spender indexed address, value uint256)
func (_SNMTToken *SNMTTokenFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *SNMTTokenApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _SNMTToken.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SNMTTokenApproval)
				if err := _SNMTToken.contract.UnpackLog(event, "Approval", log); err != nil {
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

// SNMTTokenGiveAwayIterator is returned from FilterGiveAway and is used to iterate over the raw logs and unpacked data for GiveAway events raised by the SNMTToken contract.
type SNMTTokenGiveAwayIterator struct {
	Event *SNMTTokenGiveAway // Event containing the contract specifics and raw log

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
func (it *SNMTTokenGiveAwayIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SNMTTokenGiveAway)
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
		it.Event = new(SNMTTokenGiveAway)
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
func (it *SNMTTokenGiveAwayIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SNMTTokenGiveAwayIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SNMTTokenGiveAway represents a GiveAway event raised by the SNMTToken contract.
type SNMTTokenGiveAway struct {
	Whom   common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterGiveAway is a free log retrieval operation binding the contract event 0xe08e9d066634006283658128ec91f58d444719d7a07d49f72924da4352ff94ad.
//
// Solidity: event GiveAway(whom address, amount uint256)
func (_SNMTToken *SNMTTokenFilterer) FilterGiveAway(opts *bind.FilterOpts) (*SNMTTokenGiveAwayIterator, error) {

	logs, sub, err := _SNMTToken.contract.FilterLogs(opts, "GiveAway")
	if err != nil {
		return nil, err
	}
	return &SNMTTokenGiveAwayIterator{contract: _SNMTToken.contract, event: "GiveAway", logs: logs, sub: sub}, nil
}

// WatchGiveAway is a free log subscription operation binding the contract event 0xe08e9d066634006283658128ec91f58d444719d7a07d49f72924da4352ff94ad.
//
// Solidity: event GiveAway(whom address, amount uint256)
func (_SNMTToken *SNMTTokenFilterer) WatchGiveAway(opts *bind.WatchOpts, sink chan<- *SNMTTokenGiveAway) (event.Subscription, error) {

	logs, sub, err := _SNMTToken.contract.WatchLogs(opts, "GiveAway")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SNMTTokenGiveAway)
				if err := _SNMTToken.contract.UnpackLog(event, "GiveAway", log); err != nil {
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

// SNMTTokenTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the SNMTToken contract.
type SNMTTokenTransferIterator struct {
	Event *SNMTTokenTransfer // Event containing the contract specifics and raw log

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
func (it *SNMTTokenTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SNMTTokenTransfer)
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
		it.Event = new(SNMTTokenTransfer)
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
func (it *SNMTTokenTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SNMTTokenTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SNMTTokenTransfer represents a Transfer event raised by the SNMTToken contract.
type SNMTTokenTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(from indexed address, to indexed address, value uint256)
func (_SNMTToken *SNMTTokenFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*SNMTTokenTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _SNMTToken.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &SNMTTokenTransferIterator{contract: _SNMTToken.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(from indexed address, to indexed address, value uint256)
func (_SNMTToken *SNMTTokenFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *SNMTTokenTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _SNMTToken.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SNMTTokenTransfer)
				if err := _SNMTToken.contract.UnpackLog(event, "Transfer", log); err != nil {
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
