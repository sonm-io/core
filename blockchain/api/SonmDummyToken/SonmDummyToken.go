// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package SonmDummyToken

import (
	"math/big"
	"strings"

	"github.com/sonm-io/go-ethereum/accounts/abi"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"github.com/sonm-io/go-ethereum/common"
	"github.com/sonm-io/go-ethereum/core/types"
)

// SonmDummyTokenABI is the input ABI used to generate the binding from.
const SonmDummyTokenABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"mintingFinished\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"INITIAL_SUPPLY\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"finishMinting\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"remaining\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"inputs\":[{\"name\":\"initialAccount\",\"type\":\"address\"}],\"payable\":false,\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Mint\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"MintFinished\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"}]"

// SonmDummyTokenBin is the compiled bytecode used for deploying new contracts.
const SonmDummyTokenBin = `0x606060409081526003805460a060020a60ff02191690558051908101604052601081527f536f6e6d2044756d6d7920546f6b656e000000000000000000000000000000006020820152600490805161005b929160200190610114565b5060408051908101604052600381527f5344540000000000000000000000000000000000000000000000000000000000602082015260059080516100a3929160200190610114565b506012600655620f424060075534156100bb57600080fd5b604051602080610c6d833981016040528080519150505b5b60038054600160a060020a03191633600160a060020a03161790555b6007546000818155600160a060020a0383168152600160205260409020555b506101b4565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061015557805160ff1916838001178555610182565b82800160010185558215610182579182015b82811115610182578251825591602001919060010190610167565b5b5061018f929150610193565b5090565b6101b191905b8082111561018f5760008155600101610199565b5090565b90565b610aaa806101c36000396000f300606060405236156100d85763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166305d2035b81146100dd57806306fdde0314610104578063095ea7b31461018f57806318160ddd146101c557806323b872dd146101ea5780632ff2e9dc14610226578063313ce5671461024b57806340c10f191461027057806370a08231146102a65780637d64bcb4146102d75780638da5cb5b146102fe57806395d89b411461032d578063a9059cbb146103b8578063dd62ed3e146103ee578063f2fde38b14610425575b600080fd5b34156100e857600080fd5b6100f0610446565b604051901515815260200160405180910390f35b341561010f57600080fd5b610117610467565b60405160208082528190810183818151815260200191508051906020019080838360005b838110156101545780820151818401525b60200161013b565b50505050905090810190601f1680156101815780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b341561019a57600080fd5b6100f0600160a060020a0360043516602435610505565b604051901515815260200160405180910390f35b34156101d057600080fd5b6101d86105ac565b60405190815260200160405180910390f35b34156101f557600080fd5b6100f0600160a060020a03600435811690602435166044356105b2565b604051901515815260200160405180910390f35b341561023157600080fd5b6101d86106c7565b60405190815260200160405180910390f35b341561025657600080fd5b6101d86106cd565b60405190815260200160405180910390f35b341561027b57600080fd5b6100f0600160a060020a03600435166024356106d3565b604051901515815260200160405180910390f35b34156102b157600080fd5b6101d8600160a060020a03600435166107b5565b60405190815260200160405180910390f35b34156102e257600080fd5b6100f06107d4565b604051901515815260200160405180910390f35b341561030957600080fd5b61031161085b565b604051600160a060020a03909116815260200160405180910390f35b341561033857600080fd5b61011761086a565b60405160208082528190810183818151815260200191508051906020019080838360005b838110156101545780820151818401525b60200161013b565b50505050905090810190601f1680156101815780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34156103c357600080fd5b6100f0600160a060020a0360043516602435610908565b604051901515815260200160405180910390f35b34156103f957600080fd5b6101d8600160a060020a03600435811690602435166109c8565b60405190815260200160405180910390f35b341561043057600080fd5b610444600160a060020a03600435166109f5565b005b60035474010000000000000000000000000000000000000000900460ff1681565b60048054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156104fd5780601f106104d2576101008083540402835291602001916104fd565b820191906000526020600020905b8154815290600101906020018083116104e057829003601f168201915b505050505081565b60008115806105375750600160a060020a03338116600090815260026020908152604080832093871683529290522054155b151561054257600080fd5b600160a060020a03338116600081815260026020908152604080832094881680845294909152908190208590557f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9259085905190815260200160405180910390a35060015b92915050565b60005481565b600160a060020a0380841660009081526002602090815260408083203385168452825280832054938616835260019091528120549091906105f9908463ffffffff610a4d16565b600160a060020a03808616600090815260016020526040808220939093559087168152205461062e908463ffffffff610a6716565b600160a060020a038616600090815260016020526040902055610657818463ffffffff610a6716565b600160a060020a03808716600081815260026020908152604080832033861684529091529081902093909355908616917fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9086905190815260200160405180910390a3600191505b509392505050565b60075481565b60065481565b60035460009033600160a060020a039081169116146106f157600080fd5b60035474010000000000000000000000000000000000000000900460ff161561071957600080fd5b60005461072c908363ffffffff610a4d16565b6000908155600160a060020a038416815260016020526040902054610757908363ffffffff610a4d16565b600160a060020a0384166000818152600160205260409081902092909255907f0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d41213968859084905190815260200160405180910390a25060015b5b5b92915050565b600160a060020a0381166000908152600160205260409020545b919050565b60035460009033600160a060020a039081169116146107f257600080fd5b6003805474ff00000000000000000000000000000000000000001916740100000000000000000000000000000000000000001790557fae5184fba832cb2b1f702aca6117b8d265eaf03ad33eb133f19dde0f5920fa0860405160405180910390a15060015b5b90565b600354600160a060020a031681565b60058054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156104fd5780601f106104d2576101008083540402835291602001916104fd565b820191906000526020600020905b8154815290600101906020018083116104e057829003601f168201915b505050505081565b600160a060020a033316600090815260016020526040812054610931908363ffffffff610a6716565b600160a060020a033381166000908152600160205260408082209390935590851681522054610966908363ffffffff610a4d16565b600160a060020a0380851660008181526001602052604090819020939093559133909116907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9085905190815260200160405180910390a35060015b92915050565b600160a060020a038083166000908152600260209081526040808320938516835292905220545b92915050565b60035433600160a060020a03908116911614610a1057600080fd5b600160a060020a03811615610a48576003805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0383161790555b5b5b50565b600082820183811015610a5c57fe5b8091505b5092915050565b600082821115610a7357fe5b508082035b929150505600a165627a7a7230582028ae14e21c3b478295b86212eb32a0f6ba022552c4c04fde7d763d0ac3767abd0029`

// DeploySonmDummyToken deploys a new Ethereum contract, binding an instance of SonmDummyToken to it.
func DeploySonmDummyToken(auth *bind.TransactOpts, backend bind.ContractBackend, initialAccount common.Address) (common.Address, *types.Transaction, *SonmDummyToken, error) {
	parsed, err := abi.JSON(strings.NewReader(SonmDummyTokenABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SonmDummyTokenBin), backend, initialAccount)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SonmDummyToken{SonmDummyTokenCaller: SonmDummyTokenCaller{contract: contract}, SonmDummyTokenTransactor: SonmDummyTokenTransactor{contract: contract}}, nil
}

// SonmDummyToken is an auto generated Go binding around an Ethereum contract.
type SonmDummyToken struct {
	SonmDummyTokenCaller     // Read-only binding to the contract
	SonmDummyTokenTransactor // Write-only binding to the contract
}

// SonmDummyTokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type SonmDummyTokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SonmDummyTokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SonmDummyTokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SonmDummyTokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SonmDummyTokenSession struct {
	Contract     *SonmDummyToken   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SonmDummyTokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SonmDummyTokenCallerSession struct {
	Contract *SonmDummyTokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// SonmDummyTokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SonmDummyTokenTransactorSession struct {
	Contract     *SonmDummyTokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// SonmDummyTokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type SonmDummyTokenRaw struct {
	Contract *SonmDummyToken // Generic contract binding to access the raw methods on
}

// SonmDummyTokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SonmDummyTokenCallerRaw struct {
	Contract *SonmDummyTokenCaller // Generic read-only contract binding to access the raw methods on
}

// SonmDummyTokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SonmDummyTokenTransactorRaw struct {
	Contract *SonmDummyTokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSonmDummyToken creates a new instance of SonmDummyToken, bound to a specific deployed contract.
func NewSonmDummyToken(address common.Address, backend bind.ContractBackend) (*SonmDummyToken, error) {
	contract, err := bindSonmDummyToken(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SonmDummyToken{SonmDummyTokenCaller: SonmDummyTokenCaller{contract: contract}, SonmDummyTokenTransactor: SonmDummyTokenTransactor{contract: contract}}, nil
}

// NewSonmDummyTokenCaller creates a new read-only instance of SonmDummyToken, bound to a specific deployed contract.
func NewSonmDummyTokenCaller(address common.Address, caller bind.ContractCaller) (*SonmDummyTokenCaller, error) {
	contract, err := bindSonmDummyToken(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &SonmDummyTokenCaller{contract: contract}, nil
}

// NewSonmDummyTokenTransactor creates a new write-only instance of SonmDummyToken, bound to a specific deployed contract.
func NewSonmDummyTokenTransactor(address common.Address, transactor bind.ContractTransactor) (*SonmDummyTokenTransactor, error) {
	contract, err := bindSonmDummyToken(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &SonmDummyTokenTransactor{contract: contract}, nil
}

// bindSonmDummyToken binds a generic wrapper to an already deployed contract.
func bindSonmDummyToken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SonmDummyTokenABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SonmDummyToken *SonmDummyTokenRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SonmDummyToken.Contract.SonmDummyTokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SonmDummyToken *SonmDummyTokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SonmDummyToken.Contract.SonmDummyTokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SonmDummyToken *SonmDummyTokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SonmDummyToken.Contract.SonmDummyTokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SonmDummyToken *SonmDummyTokenCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SonmDummyToken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SonmDummyToken *SonmDummyTokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SonmDummyToken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SonmDummyToken *SonmDummyTokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SonmDummyToken.Contract.contract.Transact(opts, method, params...)
}

// INITIAL_SUPPLY is a free data retrieval call binding the contract method 0x2ff2e9dc.
//
// Solidity: function INITIAL_SUPPLY() constant returns(uint256)
func (_SonmDummyToken *SonmDummyTokenCaller) INITIAL_SUPPLY(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SonmDummyToken.contract.Call(opts, out, "INITIAL_SUPPLY")
	return *ret0, err
}

// INITIAL_SUPPLY is a free data retrieval call binding the contract method 0x2ff2e9dc.
//
// Solidity: function INITIAL_SUPPLY() constant returns(uint256)
func (_SonmDummyToken *SonmDummyTokenSession) INITIAL_SUPPLY() (*big.Int, error) {
	return _SonmDummyToken.Contract.INITIAL_SUPPLY(&_SonmDummyToken.CallOpts)
}

// INITIAL_SUPPLY is a free data retrieval call binding the contract method 0x2ff2e9dc.
//
// Solidity: function INITIAL_SUPPLY() constant returns(uint256)
func (_SonmDummyToken *SonmDummyTokenCallerSession) INITIAL_SUPPLY() (*big.Int, error) {
	return _SonmDummyToken.Contract.INITIAL_SUPPLY(&_SonmDummyToken.CallOpts)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_SonmDummyToken *SonmDummyTokenCaller) Allowance(opts *bind.CallOpts, _owner common.Address, _spender common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SonmDummyToken.contract.Call(opts, out, "allowance", _owner, _spender)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_SonmDummyToken *SonmDummyTokenSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _SonmDummyToken.Contract.Allowance(&_SonmDummyToken.CallOpts, _owner, _spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_SonmDummyToken *SonmDummyTokenCallerSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _SonmDummyToken.Contract.Allowance(&_SonmDummyToken.CallOpts, _owner, _spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_SonmDummyToken *SonmDummyTokenCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SonmDummyToken.contract.Call(opts, out, "balanceOf", _owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_SonmDummyToken *SonmDummyTokenSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _SonmDummyToken.Contract.BalanceOf(&_SonmDummyToken.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_SonmDummyToken *SonmDummyTokenCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _SonmDummyToken.Contract.BalanceOf(&_SonmDummyToken.CallOpts, _owner)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_SonmDummyToken *SonmDummyTokenCaller) Decimals(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SonmDummyToken.contract.Call(opts, out, "decimals")
	return *ret0, err
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_SonmDummyToken *SonmDummyTokenSession) Decimals() (*big.Int, error) {
	return _SonmDummyToken.Contract.Decimals(&_SonmDummyToken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_SonmDummyToken *SonmDummyTokenCallerSession) Decimals() (*big.Int, error) {
	return _SonmDummyToken.Contract.Decimals(&_SonmDummyToken.CallOpts)
}

// MintingFinished is a free data retrieval call binding the contract method 0x05d2035b.
//
// Solidity: function mintingFinished() constant returns(bool)
func (_SonmDummyToken *SonmDummyTokenCaller) MintingFinished(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _SonmDummyToken.contract.Call(opts, out, "mintingFinished")
	return *ret0, err
}

// MintingFinished is a free data retrieval call binding the contract method 0x05d2035b.
//
// Solidity: function mintingFinished() constant returns(bool)
func (_SonmDummyToken *SonmDummyTokenSession) MintingFinished() (bool, error) {
	return _SonmDummyToken.Contract.MintingFinished(&_SonmDummyToken.CallOpts)
}

// MintingFinished is a free data retrieval call binding the contract method 0x05d2035b.
//
// Solidity: function mintingFinished() constant returns(bool)
func (_SonmDummyToken *SonmDummyTokenCallerSession) MintingFinished() (bool, error) {
	return _SonmDummyToken.Contract.MintingFinished(&_SonmDummyToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_SonmDummyToken *SonmDummyTokenCaller) Name(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _SonmDummyToken.contract.Call(opts, out, "name")
	return *ret0, err
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_SonmDummyToken *SonmDummyTokenSession) Name() (string, error) {
	return _SonmDummyToken.Contract.Name(&_SonmDummyToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_SonmDummyToken *SonmDummyTokenCallerSession) Name() (string, error) {
	return _SonmDummyToken.Contract.Name(&_SonmDummyToken.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SonmDummyToken *SonmDummyTokenCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _SonmDummyToken.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SonmDummyToken *SonmDummyTokenSession) Owner() (common.Address, error) {
	return _SonmDummyToken.Contract.Owner(&_SonmDummyToken.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SonmDummyToken *SonmDummyTokenCallerSession) Owner() (common.Address, error) {
	return _SonmDummyToken.Contract.Owner(&_SonmDummyToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_SonmDummyToken *SonmDummyTokenCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _SonmDummyToken.contract.Call(opts, out, "symbol")
	return *ret0, err
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_SonmDummyToken *SonmDummyTokenSession) Symbol() (string, error) {
	return _SonmDummyToken.Contract.Symbol(&_SonmDummyToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_SonmDummyToken *SonmDummyTokenCallerSession) Symbol() (string, error) {
	return _SonmDummyToken.Contract.Symbol(&_SonmDummyToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_SonmDummyToken *SonmDummyTokenCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SonmDummyToken.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_SonmDummyToken *SonmDummyTokenSession) TotalSupply() (*big.Int, error) {
	return _SonmDummyToken.Contract.TotalSupply(&_SonmDummyToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_SonmDummyToken *SonmDummyTokenCallerSession) TotalSupply() (*big.Int, error) {
	return _SonmDummyToken.Contract.TotalSupply(&_SonmDummyToken.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(bool)
func (_SonmDummyToken *SonmDummyTokenTransactor) Approve(opts *bind.TransactOpts, _spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SonmDummyToken.contract.Transact(opts, "approve", _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(bool)
func (_SonmDummyToken *SonmDummyTokenSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SonmDummyToken.Contract.Approve(&_SonmDummyToken.TransactOpts, _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(bool)
func (_SonmDummyToken *SonmDummyTokenTransactorSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SonmDummyToken.Contract.Approve(&_SonmDummyToken.TransactOpts, _spender, _value)
}

// FinishMinting is a paid mutator transaction binding the contract method 0x7d64bcb4.
//
// Solidity: function finishMinting() returns(bool)
func (_SonmDummyToken *SonmDummyTokenTransactor) FinishMinting(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SonmDummyToken.contract.Transact(opts, "finishMinting")
}

// FinishMinting is a paid mutator transaction binding the contract method 0x7d64bcb4.
//
// Solidity: function finishMinting() returns(bool)
func (_SonmDummyToken *SonmDummyTokenSession) FinishMinting() (*types.Transaction, error) {
	return _SonmDummyToken.Contract.FinishMinting(&_SonmDummyToken.TransactOpts)
}

// FinishMinting is a paid mutator transaction binding the contract method 0x7d64bcb4.
//
// Solidity: function finishMinting() returns(bool)
func (_SonmDummyToken *SonmDummyTokenTransactorSession) FinishMinting() (*types.Transaction, error) {
	return _SonmDummyToken.Contract.FinishMinting(&_SonmDummyToken.TransactOpts)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(_to address, _amount uint256) returns(bool)
func (_SonmDummyToken *SonmDummyTokenTransactor) Mint(opts *bind.TransactOpts, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SonmDummyToken.contract.Transact(opts, "mint", _to, _amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(_to address, _amount uint256) returns(bool)
func (_SonmDummyToken *SonmDummyTokenSession) Mint(_to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SonmDummyToken.Contract.Mint(&_SonmDummyToken.TransactOpts, _to, _amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(_to address, _amount uint256) returns(bool)
func (_SonmDummyToken *SonmDummyTokenTransactorSession) Mint(_to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SonmDummyToken.Contract.Mint(&_SonmDummyToken.TransactOpts, _to, _amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(bool)
func (_SonmDummyToken *SonmDummyTokenTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SonmDummyToken.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(bool)
func (_SonmDummyToken *SonmDummyTokenSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SonmDummyToken.Contract.Transfer(&_SonmDummyToken.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(bool)
func (_SonmDummyToken *SonmDummyTokenTransactorSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SonmDummyToken.Contract.Transfer(&_SonmDummyToken.TransactOpts, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(bool)
func (_SonmDummyToken *SonmDummyTokenTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SonmDummyToken.contract.Transact(opts, "transferFrom", _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(bool)
func (_SonmDummyToken *SonmDummyTokenSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SonmDummyToken.Contract.TransferFrom(&_SonmDummyToken.TransactOpts, _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(bool)
func (_SonmDummyToken *SonmDummyTokenTransactorSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SonmDummyToken.Contract.TransferFrom(&_SonmDummyToken.TransactOpts, _from, _to, _value)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_SonmDummyToken *SonmDummyTokenTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SonmDummyToken.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_SonmDummyToken *SonmDummyTokenSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SonmDummyToken.Contract.TransferOwnership(&_SonmDummyToken.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_SonmDummyToken *SonmDummyTokenTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SonmDummyToken.Contract.TransferOwnership(&_SonmDummyToken.TransactOpts, newOwner)
}
