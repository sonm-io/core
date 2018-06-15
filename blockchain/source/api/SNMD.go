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

// SNMDABI is the input ABI used to generate the binding from.
const SNMDABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseApproval\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseApproval\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newMarket\",\"type\":\"address\"}],\"name\":\"AddMarket\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// SNMDBin is the compiled bytecode used for deploying new contracts.
const SNMDBin = `0x60c0604052601660808190527f534f4e4d20446576656c6f706d656e7420746f6b656e0000000000000000000060a090815261003e91600491906100d9565b506040805180820190915260048082527f534e4d44000000000000000000000000000000000000000000000000000000006020909201918252610083916005916100d9565b5060126006556b016f44a83aab6c233c0000006008553480156100a557600080fd5b506003805433600160a060020a03199182168117909116811790915560085460009182526020829052604090912055610174565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061011a57805160ff1916838001178555610147565b82800160010185558215610147579182015b8281111561014757825182559160200191906001019061012c565b50610153929150610157565b5090565b61017191905b80821115610153576000815560010161015d565b90565b610b27806101836000396000f3006080604052600436106100cf5763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306fdde0381146100d4578063095ea7b31461015e57806318160ddd1461019657806323b872dd146101bd578063313ce567146101e757806366188463146101fc57806370a08231146102205780638da5cb5b1461024157806395d89b4114610272578063a9059cbb14610287578063c3dfb88e146102ab578063d73dd623146102ce578063dd62ed3e146102f2578063f2fde38b14610319575b600080fd5b3480156100e057600080fd5b506100e961033a565b6040805160208082528351818301528351919283929083019185019080838360005b8381101561012357818101518382015260200161010b565b50505050905090810190601f1680156101505780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561016a57600080fd5b50610182600160a060020a03600435166024356103c8565b604080519115158252519081900360200190f35b3480156101a257600080fd5b506101ab61042e565b60408051918252519081900360200190f35b3480156101c957600080fd5b50610182600160a060020a0360043581169060243516604435610434565b3480156101f357600080fd5b506101ab6106e6565b34801561020857600080fd5b50610182600160a060020a03600435166024356106ec565b34801561022c57600080fd5b506101ab600160a060020a03600435166107dc565b34801561024d57600080fd5b506102566107f7565b60408051600160a060020a039092168252519081900360200190f35b34801561027e57600080fd5b506100e9610806565b34801561029357600080fd5b50610182600160a060020a0360043516602435610861565b3480156102b757600080fd5b506102cc600160a060020a0360043516610942565b005b3480156102da57600080fd5b50610182600160a060020a036004351660243561097d565b3480156102fe57600080fd5b506101ab600160a060020a0360043581169060243516610a16565b34801561032557600080fd5b506102cc600160a060020a0360043516610a41565b6004805460408051602060026001851615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156103c05780601f10610395576101008083540402835291602001916103c0565b820191906000526020600020905b8154815290600101906020018083116103a357829003601f168201915b505050505081565b336000818152600260209081526040808320600160a060020a038716808552908352818420869055815186815291519394909390927f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925928290030190a350600192915050565b60015490565b600160a060020a03821660009081526007602052604081205460ff1615156001141561057057600160a060020a038316151561046f57600080fd5b3360009081526007602052604090205460ff16151560011461049057600080fd5b600160a060020a0384166000908152602081905260409020548211156104b557600080fd5b600160a060020a0384166000908152602081905260409020546104de908363ffffffff610ad616565b600160a060020a038086166000908152602081905260408082209390935590851681522054610513908363ffffffff610ae816565b600160a060020a038085166000818152602081815260409182902094909455805186815290519193928816927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef92918290030190a35060016106df565b600160a060020a038316151561058557600080fd5b600160a060020a0384166000908152602081905260409020548211156105aa57600080fd5b600160a060020a03841660009081526002602090815260408083203384529091529020548211156105da57600080fd5b600160a060020a038416600090815260208190526040902054610603908363ffffffff610ad616565b600160a060020a038086166000908152602081905260408082209390935590851681522054610638908363ffffffff610ae816565b600160a060020a0380851660009081526020818152604080832094909455918716815260028252828120338252909152205461067a908363ffffffff610ad616565b600160a060020a03808616600081815260026020908152604080832033845282529182902094909455805186815290519287169391927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef929181900390910190a35060015b9392505050565b60065481565b336000908152600260209081526040808320600160a060020a03861684529091528120548083111561074157336000908152600260209081526040808320600160a060020a0388168452909152812055610776565b610751818463ffffffff610ad616565b336000908152600260209081526040808320600160a060020a03891684529091529020555b336000818152600260209081526040808320600160a060020a0389168085529083529281902054815190815290519293927f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925929181900390910190a35060019392505050565b600160a060020a031660009081526020819052604090205490565b600354600160a060020a031681565b6005805460408051602060026001851615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156103c05780601f10610395576101008083540402835291602001916103c0565b6000600160a060020a038316151561087857600080fd5b3360009081526020819052604090205482111561089457600080fd5b336000908152602081905260409020546108b4908363ffffffff610ad616565b3360009081526020819052604080822092909255600160a060020a038516815220546108e6908363ffffffff610ae816565b600160a060020a038416600081815260208181526040918290209390935580518581529051919233927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9281900390910190a350600192915050565b600354600160a060020a0316331461095957600080fd5b600160a060020a03166000908152600760205260409020805460ff19166001179055565b336000908152600260209081526040808320600160a060020a03861684529091528120546109b1908363ffffffff610ae816565b336000818152600260209081526040808320600160a060020a0389168085529083529281902085905580519485525191937f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925929081900390910190a350600192915050565b600160a060020a03918216600090815260026020908152604080832093909416825291909152205490565b600354600160a060020a03163314610a5857600080fd5b600160a060020a0381161515610a6d57600080fd5b600354604051600160a060020a038084169216907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a36003805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0392909216919091179055565b600082821115610ae257fe5b50900390565b81810182811015610af557fe5b929150505600a165627a7a72305820636b2a1ece2a7fc9da3d25c9846962d50fb0a3c6f739527d1cbdc8bfccd8bbf20029`

// DeploySNMD deploys a new Ethereum contract, binding an instance of SNMD to it.
func DeploySNMD(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SNMD, error) {
	parsed, err := abi.JSON(strings.NewReader(SNMDABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SNMDBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SNMD{SNMDCaller: SNMDCaller{contract: contract}, SNMDTransactor: SNMDTransactor{contract: contract}, SNMDFilterer: SNMDFilterer{contract: contract}}, nil
}

// SNMD is an auto generated Go binding around an Ethereum contract.
type SNMD struct {
	SNMDCaller     // Read-only binding to the contract
	SNMDTransactor // Write-only binding to the contract
	SNMDFilterer   // Log filterer for contract events
}

// SNMDCaller is an auto generated read-only Go binding around an Ethereum contract.
type SNMDCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SNMDTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SNMDTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SNMDFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SNMDFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SNMDSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SNMDSession struct {
	Contract     *SNMD             // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SNMDCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SNMDCallerSession struct {
	Contract *SNMDCaller   // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// SNMDTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SNMDTransactorSession struct {
	Contract     *SNMDTransactor   // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SNMDRaw is an auto generated low-level Go binding around an Ethereum contract.
type SNMDRaw struct {
	Contract *SNMD // Generic contract binding to access the raw methods on
}

// SNMDCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SNMDCallerRaw struct {
	Contract *SNMDCaller // Generic read-only contract binding to access the raw methods on
}

// SNMDTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SNMDTransactorRaw struct {
	Contract *SNMDTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSNMD creates a new instance of SNMD, bound to a specific deployed contract.
func NewSNMD(address common.Address, backend bind.ContractBackend) (*SNMD, error) {
	contract, err := bindSNMD(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SNMD{SNMDCaller: SNMDCaller{contract: contract}, SNMDTransactor: SNMDTransactor{contract: contract}, SNMDFilterer: SNMDFilterer{contract: contract}}, nil
}

// NewSNMDCaller creates a new read-only instance of SNMD, bound to a specific deployed contract.
func NewSNMDCaller(address common.Address, caller bind.ContractCaller) (*SNMDCaller, error) {
	contract, err := bindSNMD(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SNMDCaller{contract: contract}, nil
}

// NewSNMDTransactor creates a new write-only instance of SNMD, bound to a specific deployed contract.
func NewSNMDTransactor(address common.Address, transactor bind.ContractTransactor) (*SNMDTransactor, error) {
	contract, err := bindSNMD(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SNMDTransactor{contract: contract}, nil
}

// NewSNMDFilterer creates a new log filterer instance of SNMD, bound to a specific deployed contract.
func NewSNMDFilterer(address common.Address, filterer bind.ContractFilterer) (*SNMDFilterer, error) {
	contract, err := bindSNMD(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SNMDFilterer{contract: contract}, nil
}

// bindSNMD binds a generic wrapper to an already deployed contract.
func bindSNMD(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SNMDABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SNMD *SNMDRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SNMD.Contract.SNMDCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SNMD *SNMDRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SNMD.Contract.SNMDTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SNMD *SNMDRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SNMD.Contract.SNMDTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SNMD *SNMDCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SNMD.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SNMD *SNMDTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SNMD.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SNMD *SNMDTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SNMD.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(uint256)
func (_SNMD *SNMDCaller) Allowance(opts *bind.CallOpts, _owner common.Address, _spender common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SNMD.contract.Call(opts, out, "allowance", _owner, _spender)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(uint256)
func (_SNMD *SNMDSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _SNMD.Contract.Allowance(&_SNMD.CallOpts, _owner, _spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(uint256)
func (_SNMD *SNMDCallerSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _SNMD.Contract.Allowance(&_SNMD.CallOpts, _owner, _spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(uint256)
func (_SNMD *SNMDCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SNMD.contract.Call(opts, out, "balanceOf", _owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(uint256)
func (_SNMD *SNMDSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _SNMD.Contract.BalanceOf(&_SNMD.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(uint256)
func (_SNMD *SNMDCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _SNMD.Contract.BalanceOf(&_SNMD.CallOpts, _owner)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_SNMD *SNMDCaller) Decimals(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SNMD.contract.Call(opts, out, "decimals")
	return *ret0, err
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_SNMD *SNMDSession) Decimals() (*big.Int, error) {
	return _SNMD.Contract.Decimals(&_SNMD.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_SNMD *SNMDCallerSession) Decimals() (*big.Int, error) {
	return _SNMD.Contract.Decimals(&_SNMD.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_SNMD *SNMDCaller) Name(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _SNMD.contract.Call(opts, out, "name")
	return *ret0, err
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_SNMD *SNMDSession) Name() (string, error) {
	return _SNMD.Contract.Name(&_SNMD.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_SNMD *SNMDCallerSession) Name() (string, error) {
	return _SNMD.Contract.Name(&_SNMD.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SNMD *SNMDCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _SNMD.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SNMD *SNMDSession) Owner() (common.Address, error) {
	return _SNMD.Contract.Owner(&_SNMD.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SNMD *SNMDCallerSession) Owner() (common.Address, error) {
	return _SNMD.Contract.Owner(&_SNMD.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_SNMD *SNMDCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _SNMD.contract.Call(opts, out, "symbol")
	return *ret0, err
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_SNMD *SNMDSession) Symbol() (string, error) {
	return _SNMD.Contract.Symbol(&_SNMD.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_SNMD *SNMDCallerSession) Symbol() (string, error) {
	return _SNMD.Contract.Symbol(&_SNMD.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_SNMD *SNMDCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SNMD.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_SNMD *SNMDSession) TotalSupply() (*big.Int, error) {
	return _SNMD.Contract.TotalSupply(&_SNMD.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_SNMD *SNMDCallerSession) TotalSupply() (*big.Int, error) {
	return _SNMD.Contract.TotalSupply(&_SNMD.CallOpts)
}

// AddMarket is a paid mutator transaction binding the contract method 0xc3dfb88e.
//
// Solidity: function AddMarket(_newMarket address) returns()
func (_SNMD *SNMDTransactor) AddMarket(opts *bind.TransactOpts, _newMarket common.Address) (*types.Transaction, error) {
	return _SNMD.contract.Transact(opts, "AddMarket", _newMarket)
}

// AddMarket is a paid mutator transaction binding the contract method 0xc3dfb88e.
//
// Solidity: function AddMarket(_newMarket address) returns()
func (_SNMD *SNMDSession) AddMarket(_newMarket common.Address) (*types.Transaction, error) {
	return _SNMD.Contract.AddMarket(&_SNMD.TransactOpts, _newMarket)
}

// AddMarket is a paid mutator transaction binding the contract method 0xc3dfb88e.
//
// Solidity: function AddMarket(_newMarket address) returns()
func (_SNMD *SNMDTransactorSession) AddMarket(_newMarket common.Address) (*types.Transaction, error) {
	return _SNMD.Contract.AddMarket(&_SNMD.TransactOpts, _newMarket)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(bool)
func (_SNMD *SNMDTransactor) Approve(opts *bind.TransactOpts, _spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMD.contract.Transact(opts, "approve", _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(bool)
func (_SNMD *SNMDSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMD.Contract.Approve(&_SNMD.TransactOpts, _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(bool)
func (_SNMD *SNMDTransactorSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMD.Contract.Approve(&_SNMD.TransactOpts, _spender, _value)
}

// DecreaseApproval is a paid mutator transaction binding the contract method 0x66188463.
//
// Solidity: function decreaseApproval(_spender address, _subtractedValue uint256) returns(bool)
func (_SNMD *SNMDTransactor) DecreaseApproval(opts *bind.TransactOpts, _spender common.Address, _subtractedValue *big.Int) (*types.Transaction, error) {
	return _SNMD.contract.Transact(opts, "decreaseApproval", _spender, _subtractedValue)
}

// DecreaseApproval is a paid mutator transaction binding the contract method 0x66188463.
//
// Solidity: function decreaseApproval(_spender address, _subtractedValue uint256) returns(bool)
func (_SNMD *SNMDSession) DecreaseApproval(_spender common.Address, _subtractedValue *big.Int) (*types.Transaction, error) {
	return _SNMD.Contract.DecreaseApproval(&_SNMD.TransactOpts, _spender, _subtractedValue)
}

// DecreaseApproval is a paid mutator transaction binding the contract method 0x66188463.
//
// Solidity: function decreaseApproval(_spender address, _subtractedValue uint256) returns(bool)
func (_SNMD *SNMDTransactorSession) DecreaseApproval(_spender common.Address, _subtractedValue *big.Int) (*types.Transaction, error) {
	return _SNMD.Contract.DecreaseApproval(&_SNMD.TransactOpts, _spender, _subtractedValue)
}

// IncreaseApproval is a paid mutator transaction binding the contract method 0xd73dd623.
//
// Solidity: function increaseApproval(_spender address, _addedValue uint256) returns(bool)
func (_SNMD *SNMDTransactor) IncreaseApproval(opts *bind.TransactOpts, _spender common.Address, _addedValue *big.Int) (*types.Transaction, error) {
	return _SNMD.contract.Transact(opts, "increaseApproval", _spender, _addedValue)
}

// IncreaseApproval is a paid mutator transaction binding the contract method 0xd73dd623.
//
// Solidity: function increaseApproval(_spender address, _addedValue uint256) returns(bool)
func (_SNMD *SNMDSession) IncreaseApproval(_spender common.Address, _addedValue *big.Int) (*types.Transaction, error) {
	return _SNMD.Contract.IncreaseApproval(&_SNMD.TransactOpts, _spender, _addedValue)
}

// IncreaseApproval is a paid mutator transaction binding the contract method 0xd73dd623.
//
// Solidity: function increaseApproval(_spender address, _addedValue uint256) returns(bool)
func (_SNMD *SNMDTransactorSession) IncreaseApproval(_spender common.Address, _addedValue *big.Int) (*types.Transaction, error) {
	return _SNMD.Contract.IncreaseApproval(&_SNMD.TransactOpts, _spender, _addedValue)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(bool)
func (_SNMD *SNMDTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMD.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(bool)
func (_SNMD *SNMDSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMD.Contract.Transfer(&_SNMD.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(bool)
func (_SNMD *SNMDTransactorSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMD.Contract.Transfer(&_SNMD.TransactOpts, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(bool)
func (_SNMD *SNMDTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMD.contract.Transact(opts, "transferFrom", _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(bool)
func (_SNMD *SNMDSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMD.Contract.TransferFrom(&_SNMD.TransactOpts, _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(bool)
func (_SNMD *SNMDTransactorSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMD.Contract.TransferFrom(&_SNMD.TransactOpts, _from, _to, _value)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_SNMD *SNMDTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SNMD.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_SNMD *SNMDSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SNMD.Contract.TransferOwnership(&_SNMD.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_SNMD *SNMDTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SNMD.Contract.TransferOwnership(&_SNMD.TransactOpts, newOwner)
}

// SNMDApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the SNMD contract.
type SNMDApprovalIterator struct {
	Event *SNMDApproval // Event containing the contract specifics and raw log

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
func (it *SNMDApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SNMDApproval)
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
		it.Event = new(SNMDApproval)
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
func (it *SNMDApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SNMDApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SNMDApproval represents a Approval event raised by the SNMD contract.
type SNMDApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(owner indexed address, spender indexed address, value uint256)
func (_SNMD *SNMDFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*SNMDApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _SNMD.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &SNMDApprovalIterator{contract: _SNMD.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(owner indexed address, spender indexed address, value uint256)
func (_SNMD *SNMDFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *SNMDApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _SNMD.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SNMDApproval)
				if err := _SNMD.contract.UnpackLog(event, "Approval", log); err != nil {
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

// SNMDOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SNMD contract.
type SNMDOwnershipTransferredIterator struct {
	Event *SNMDOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *SNMDOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SNMDOwnershipTransferred)
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
		it.Event = new(SNMDOwnershipTransferred)
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
func (it *SNMDOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SNMDOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SNMDOwnershipTransferred represents a OwnershipTransferred event raised by the SNMD contract.
type SNMDOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_SNMD *SNMDFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SNMDOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SNMD.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SNMDOwnershipTransferredIterator{contract: _SNMD.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_SNMD *SNMDFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SNMDOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SNMD.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SNMDOwnershipTransferred)
				if err := _SNMD.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// SNMDTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the SNMD contract.
type SNMDTransferIterator struct {
	Event *SNMDTransfer // Event containing the contract specifics and raw log

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
func (it *SNMDTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SNMDTransfer)
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
		it.Event = new(SNMDTransfer)
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
func (it *SNMDTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SNMDTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SNMDTransfer represents a Transfer event raised by the SNMD contract.
type SNMDTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(from indexed address, to indexed address, value uint256)
func (_SNMD *SNMDFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*SNMDTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _SNMD.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &SNMDTransferIterator{contract: _SNMD.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(from indexed address, to indexed address, value uint256)
func (_SNMD *SNMDFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *SNMDTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _SNMD.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SNMDTransfer)
				if err := _SNMD.contract.UnpackLog(event, "Transfer", log); err != nil {
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
