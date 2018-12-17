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

// SNMMasterchainABI is the input ABI used to generate the binding from.
const SNMMasterchainABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ico\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"tokensAreFrozen\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"remaining\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_ico\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_holder\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"defrost\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// SNMMasterchainBin is the compiled bytecode used for deploying new contracts.
const SNMMasterchainBin = `0x60c0604052600a60808190527f534f4e4d20546f6b656e0000000000000000000000000000000000000000000060a090815261003e91600391906100f2565b506040805180820190915260038082527f534e4d00000000000000000000000000000000000000000000000000000000006020909201918252610083916004916100f2565b5060126005556006805460a060020a60ff021916740100000000000000000000000000000000000000001790553480156100bc57600080fd5b506040516020806109e5833981016040525160068054600160a060020a031916600160a060020a0390921691909117905561018d565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061013357805160ff1916838001178555610160565b82800160010185558215610160579182015b82811115610160578251825591602001919060010190610145565b5061016c929150610170565b5090565b61018a91905b8082111561016c5760008155600101610176565b90565b6108498061019c6000396000f3006080604052600436106100c45763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306fdde0381146100c9578063095ea7b31461015357806318160ddd1461017957806323b872dd146101a0578063313ce567146101ca57806340c10f19146101df5780635d4522011461020357806370a082311461023457806395d89b4114610255578063a9059cbb1461026a578063ca67065f1461028e578063dd62ed3e146102b7578063f21cdf6f146102de575b600080fd5b3480156100d557600080fd5b506100de6102f3565b6040805160208082528351818301528351919283929083019185019080838360005b83811015610118578181015183820152602001610100565b50505050905090810190601f1680156101455780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561015f57600080fd5b50610177600160a060020a0360043516602435610381565b005b34801561018557600080fd5b5061018e6103b7565b60408051918252519081900360200190f35b3480156101ac57600080fd5b50610177600160a060020a03600435811690602435166044356103bd565b3480156101d657600080fd5b5061018e6103f5565b3480156101eb57600080fd5b50610177600160a060020a03600435166024356103fb565b34801561020f57600080fd5b50610218610499565b60408051600160a060020a039092168252519081900360200190f35b34801561024057600080fd5b5061018e600160a060020a03600435166104a8565b34801561026157600080fd5b506100de6104c3565b34801561027657600080fd5b50610177600160a060020a036004351660243561051e565b34801561029a57600080fd5b506102a3610550565b604080519115158252519081900360200190f35b3480156102c357600080fd5b5061018e600160a060020a0360043581169060243516610571565b3480156102ea57600080fd5b5061017761059c565b6003805460408051602060026001851615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156103795780601f1061034e57610100808354040283529160200191610379565b820191906000526020600020905b81548152906001019060200180831161035c57829003601f168201915b505050505081565b60065474010000000000000000000000000000000000000000900460ff16156103a957600080fd5b6103b382826105d3565b5050565b60005481565b60065474010000000000000000000000000000000000000000900460ff16156103e557600080fd5b6103f0838383610635565b505050565b60055481565b600654600160a060020a0316331461041257600080fd5b80151561041e57600080fd5b6000546b016f44a83aab6c233c000000908201111561043c57600080fd5b600160a060020a0382166000818152600160209081526040808320805486019055825485018355805185815290517fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef929181900390910190a35050565b600654600160a060020a031681565b600160a060020a031660009081526001602052604090205490565b6004805460408051602060026001851615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156103795780601f1061034e57610100808354040283529160200191610379565b60065474010000000000000000000000000000000000000000900460ff161561054657600080fd5b6103b3828261073c565b60065474010000000000000000000000000000000000000000900460ff1681565b600160a060020a03918216600090815260026020908152604080832093909416825291909152205490565b600654600160a060020a031633146105b357600080fd5b6006805474ff000000000000000000000000000000000000000019169055565b336000818152600260209081526040808320600160a060020a03871680855290835292819020859055805185815290519293927f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925929181900390910190a35050565b600160a060020a03808416600090815260026020908152604080832033845282528083205493861683526001909152902054610677908363ffffffff6107f816565b600160a060020a0380851660009081526001602052604080822093909355908616815220546106ac908363ffffffff61080b16565b600160a060020a0385166000908152600160205260409020556106d5818363ffffffff61080b16565b600160a060020a03808616600081815260026020908152604080832033845282529182902094909455805186815290519287169391927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef929181900390910190a350505050565b6040604436101561074c57600080fd5b3360009081526001602052604090205461076c908363ffffffff61080b16565b3360009081526001602052604080822092909255600160a060020a0385168152205461079e908363ffffffff6107f816565b600160a060020a0384166000818152600160209081526040918290209390935580518581529051919233927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9281900390910190a3505050565b8181018281101561080557fe5b92915050565b60008282111561081757fe5b509003905600a165627a7a723058204500c0d8f5bbdf8d7c79234cb49c5740cc4a9114985c151434c7210e5156ef470029`

// DeploySNMMasterchain deploys a new Ethereum contract, binding an instance of SNMMasterchain to it.
func DeploySNMMasterchain(auth *bind.TransactOpts, backend bind.ContractBackend, _ico common.Address) (common.Address, *types.Transaction, *SNMMasterchain, error) {
	parsed, err := abi.JSON(strings.NewReader(SNMMasterchainABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SNMMasterchainBin), backend, _ico)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SNMMasterchain{SNMMasterchainCaller: SNMMasterchainCaller{contract: contract}, SNMMasterchainTransactor: SNMMasterchainTransactor{contract: contract}, SNMMasterchainFilterer: SNMMasterchainFilterer{contract: contract}}, nil
}

// SNMMasterchain is an auto generated Go binding around an Ethereum contract.
type SNMMasterchain struct {
	SNMMasterchainCaller     // Read-only binding to the contract
	SNMMasterchainTransactor // Write-only binding to the contract
	SNMMasterchainFilterer   // Log filterer for contract events
}

// SNMMasterchainCaller is an auto generated read-only Go binding around an Ethereum contract.
type SNMMasterchainCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SNMMasterchainTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SNMMasterchainTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SNMMasterchainFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SNMMasterchainFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SNMMasterchainSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SNMMasterchainSession struct {
	Contract     *SNMMasterchain   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SNMMasterchainCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SNMMasterchainCallerSession struct {
	Contract *SNMMasterchainCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// SNMMasterchainTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SNMMasterchainTransactorSession struct {
	Contract     *SNMMasterchainTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// SNMMasterchainRaw is an auto generated low-level Go binding around an Ethereum contract.
type SNMMasterchainRaw struct {
	Contract *SNMMasterchain // Generic contract binding to access the raw methods on
}

// SNMMasterchainCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SNMMasterchainCallerRaw struct {
	Contract *SNMMasterchainCaller // Generic read-only contract binding to access the raw methods on
}

// SNMMasterchainTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SNMMasterchainTransactorRaw struct {
	Contract *SNMMasterchainTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSNMMasterchain creates a new instance of SNMMasterchain, bound to a specific deployed contract.
func NewSNMMasterchain(address common.Address, backend bind.ContractBackend) (*SNMMasterchain, error) {
	contract, err := bindSNMMasterchain(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SNMMasterchain{SNMMasterchainCaller: SNMMasterchainCaller{contract: contract}, SNMMasterchainTransactor: SNMMasterchainTransactor{contract: contract}, SNMMasterchainFilterer: SNMMasterchainFilterer{contract: contract}}, nil
}

// NewSNMMasterchainCaller creates a new read-only instance of SNMMasterchain, bound to a specific deployed contract.
func NewSNMMasterchainCaller(address common.Address, caller bind.ContractCaller) (*SNMMasterchainCaller, error) {
	contract, err := bindSNMMasterchain(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SNMMasterchainCaller{contract: contract}, nil
}

// NewSNMMasterchainTransactor creates a new write-only instance of SNMMasterchain, bound to a specific deployed contract.
func NewSNMMasterchainTransactor(address common.Address, transactor bind.ContractTransactor) (*SNMMasterchainTransactor, error) {
	contract, err := bindSNMMasterchain(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SNMMasterchainTransactor{contract: contract}, nil
}

// NewSNMMasterchainFilterer creates a new log filterer instance of SNMMasterchain, bound to a specific deployed contract.
func NewSNMMasterchainFilterer(address common.Address, filterer bind.ContractFilterer) (*SNMMasterchainFilterer, error) {
	contract, err := bindSNMMasterchain(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SNMMasterchainFilterer{contract: contract}, nil
}

// bindSNMMasterchain binds a generic wrapper to an already deployed contract.
func bindSNMMasterchain(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SNMMasterchainABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SNMMasterchain *SNMMasterchainRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SNMMasterchain.Contract.SNMMasterchainCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SNMMasterchain *SNMMasterchainRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SNMMasterchain.Contract.SNMMasterchainTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SNMMasterchain *SNMMasterchainRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SNMMasterchain.Contract.SNMMasterchainTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SNMMasterchain *SNMMasterchainCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SNMMasterchain.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SNMMasterchain *SNMMasterchainTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SNMMasterchain.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SNMMasterchain *SNMMasterchainTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SNMMasterchain.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_SNMMasterchain *SNMMasterchainCaller) Allowance(opts *bind.CallOpts, _owner common.Address, _spender common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SNMMasterchain.contract.Call(opts, out, "allowance", _owner, _spender)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_SNMMasterchain *SNMMasterchainSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _SNMMasterchain.Contract.Allowance(&_SNMMasterchain.CallOpts, _owner, _spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_SNMMasterchain *SNMMasterchainCallerSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _SNMMasterchain.Contract.Allowance(&_SNMMasterchain.CallOpts, _owner, _spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_SNMMasterchain *SNMMasterchainCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SNMMasterchain.contract.Call(opts, out, "balanceOf", _owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_SNMMasterchain *SNMMasterchainSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _SNMMasterchain.Contract.BalanceOf(&_SNMMasterchain.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_SNMMasterchain *SNMMasterchainCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _SNMMasterchain.Contract.BalanceOf(&_SNMMasterchain.CallOpts, _owner)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_SNMMasterchain *SNMMasterchainCaller) Decimals(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SNMMasterchain.contract.Call(opts, out, "decimals")
	return *ret0, err
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_SNMMasterchain *SNMMasterchainSession) Decimals() (*big.Int, error) {
	return _SNMMasterchain.Contract.Decimals(&_SNMMasterchain.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_SNMMasterchain *SNMMasterchainCallerSession) Decimals() (*big.Int, error) {
	return _SNMMasterchain.Contract.Decimals(&_SNMMasterchain.CallOpts)
}

// Ico is a free data retrieval call binding the contract method 0x5d452201.
//
// Solidity: function ico() constant returns(address)
func (_SNMMasterchain *SNMMasterchainCaller) Ico(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _SNMMasterchain.contract.Call(opts, out, "ico")
	return *ret0, err
}

// Ico is a free data retrieval call binding the contract method 0x5d452201.
//
// Solidity: function ico() constant returns(address)
func (_SNMMasterchain *SNMMasterchainSession) Ico() (common.Address, error) {
	return _SNMMasterchain.Contract.Ico(&_SNMMasterchain.CallOpts)
}

// Ico is a free data retrieval call binding the contract method 0x5d452201.
//
// Solidity: function ico() constant returns(address)
func (_SNMMasterchain *SNMMasterchainCallerSession) Ico() (common.Address, error) {
	return _SNMMasterchain.Contract.Ico(&_SNMMasterchain.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_SNMMasterchain *SNMMasterchainCaller) Name(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _SNMMasterchain.contract.Call(opts, out, "name")
	return *ret0, err
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_SNMMasterchain *SNMMasterchainSession) Name() (string, error) {
	return _SNMMasterchain.Contract.Name(&_SNMMasterchain.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_SNMMasterchain *SNMMasterchainCallerSession) Name() (string, error) {
	return _SNMMasterchain.Contract.Name(&_SNMMasterchain.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_SNMMasterchain *SNMMasterchainCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _SNMMasterchain.contract.Call(opts, out, "symbol")
	return *ret0, err
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_SNMMasterchain *SNMMasterchainSession) Symbol() (string, error) {
	return _SNMMasterchain.Contract.Symbol(&_SNMMasterchain.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_SNMMasterchain *SNMMasterchainCallerSession) Symbol() (string, error) {
	return _SNMMasterchain.Contract.Symbol(&_SNMMasterchain.CallOpts)
}

// TokensAreFrozen is a free data retrieval call binding the contract method 0xca67065f.
//
// Solidity: function tokensAreFrozen() constant returns(bool)
func (_SNMMasterchain *SNMMasterchainCaller) TokensAreFrozen(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _SNMMasterchain.contract.Call(opts, out, "tokensAreFrozen")
	return *ret0, err
}

// TokensAreFrozen is a free data retrieval call binding the contract method 0xca67065f.
//
// Solidity: function tokensAreFrozen() constant returns(bool)
func (_SNMMasterchain *SNMMasterchainSession) TokensAreFrozen() (bool, error) {
	return _SNMMasterchain.Contract.TokensAreFrozen(&_SNMMasterchain.CallOpts)
}

// TokensAreFrozen is a free data retrieval call binding the contract method 0xca67065f.
//
// Solidity: function tokensAreFrozen() constant returns(bool)
func (_SNMMasterchain *SNMMasterchainCallerSession) TokensAreFrozen() (bool, error) {
	return _SNMMasterchain.Contract.TokensAreFrozen(&_SNMMasterchain.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_SNMMasterchain *SNMMasterchainCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SNMMasterchain.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_SNMMasterchain *SNMMasterchainSession) TotalSupply() (*big.Int, error) {
	return _SNMMasterchain.Contract.TotalSupply(&_SNMMasterchain.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_SNMMasterchain *SNMMasterchainCallerSession) TotalSupply() (*big.Int, error) {
	return _SNMMasterchain.Contract.TotalSupply(&_SNMMasterchain.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns()
func (_SNMMasterchain *SNMMasterchainTransactor) Approve(opts *bind.TransactOpts, _spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMMasterchain.contract.Transact(opts, "approve", _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns()
func (_SNMMasterchain *SNMMasterchainSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMMasterchain.Contract.Approve(&_SNMMasterchain.TransactOpts, _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns()
func (_SNMMasterchain *SNMMasterchainTransactorSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMMasterchain.Contract.Approve(&_SNMMasterchain.TransactOpts, _spender, _value)
}

// Defrost is a paid mutator transaction binding the contract method 0xf21cdf6f.
//
// Solidity: function defrost() returns()
func (_SNMMasterchain *SNMMasterchainTransactor) Defrost(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SNMMasterchain.contract.Transact(opts, "defrost")
}

// Defrost is a paid mutator transaction binding the contract method 0xf21cdf6f.
//
// Solidity: function defrost() returns()
func (_SNMMasterchain *SNMMasterchainSession) Defrost() (*types.Transaction, error) {
	return _SNMMasterchain.Contract.Defrost(&_SNMMasterchain.TransactOpts)
}

// Defrost is a paid mutator transaction binding the contract method 0xf21cdf6f.
//
// Solidity: function defrost() returns()
func (_SNMMasterchain *SNMMasterchainTransactorSession) Defrost() (*types.Transaction, error) {
	return _SNMMasterchain.Contract.Defrost(&_SNMMasterchain.TransactOpts)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(_holder address, _value uint256) returns()
func (_SNMMasterchain *SNMMasterchainTransactor) Mint(opts *bind.TransactOpts, _holder common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMMasterchain.contract.Transact(opts, "mint", _holder, _value)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(_holder address, _value uint256) returns()
func (_SNMMasterchain *SNMMasterchainSession) Mint(_holder common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMMasterchain.Contract.Mint(&_SNMMasterchain.TransactOpts, _holder, _value)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(_holder address, _value uint256) returns()
func (_SNMMasterchain *SNMMasterchainTransactorSession) Mint(_holder common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMMasterchain.Contract.Mint(&_SNMMasterchain.TransactOpts, _holder, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns()
func (_SNMMasterchain *SNMMasterchainTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMMasterchain.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns()
func (_SNMMasterchain *SNMMasterchainSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMMasterchain.Contract.Transfer(&_SNMMasterchain.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns()
func (_SNMMasterchain *SNMMasterchainTransactorSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMMasterchain.Contract.Transfer(&_SNMMasterchain.TransactOpts, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns()
func (_SNMMasterchain *SNMMasterchainTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMMasterchain.contract.Transact(opts, "transferFrom", _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns()
func (_SNMMasterchain *SNMMasterchainSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMMasterchain.Contract.TransferFrom(&_SNMMasterchain.TransactOpts, _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns()
func (_SNMMasterchain *SNMMasterchainTransactorSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _SNMMasterchain.Contract.TransferFrom(&_SNMMasterchain.TransactOpts, _from, _to, _value)
}

// SNMMasterchainApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the SNMMasterchain contract.
type SNMMasterchainApprovalIterator struct {
	Event *SNMMasterchainApproval // Event containing the contract specifics and raw log

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
func (it *SNMMasterchainApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SNMMasterchainApproval)
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
		it.Event = new(SNMMasterchainApproval)
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
func (it *SNMMasterchainApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SNMMasterchainApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SNMMasterchainApproval represents a Approval event raised by the SNMMasterchain contract.
type SNMMasterchainApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(owner indexed address, spender indexed address, value uint256)
func (_SNMMasterchain *SNMMasterchainFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*SNMMasterchainApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _SNMMasterchain.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &SNMMasterchainApprovalIterator{contract: _SNMMasterchain.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(owner indexed address, spender indexed address, value uint256)
func (_SNMMasterchain *SNMMasterchainFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *SNMMasterchainApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _SNMMasterchain.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SNMMasterchainApproval)
				if err := _SNMMasterchain.contract.UnpackLog(event, "Approval", log); err != nil {
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

// SNMMasterchainTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the SNMMasterchain contract.
type SNMMasterchainTransferIterator struct {
	Event *SNMMasterchainTransfer // Event containing the contract specifics and raw log

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
func (it *SNMMasterchainTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SNMMasterchainTransfer)
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
		it.Event = new(SNMMasterchainTransfer)
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
func (it *SNMMasterchainTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SNMMasterchainTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SNMMasterchainTransfer represents a Transfer event raised by the SNMMasterchain contract.
type SNMMasterchainTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(from indexed address, to indexed address, value uint256)
func (_SNMMasterchain *SNMMasterchainFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*SNMMasterchainTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _SNMMasterchain.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &SNMMasterchainTransferIterator{contract: _SNMMasterchain.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(from indexed address, to indexed address, value uint256)
func (_SNMMasterchain *SNMMasterchainFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *SNMMasterchainTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _SNMMasterchain.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SNMMasterchainTransfer)
				if err := _SNMMasterchain.contract.UnpackLog(event, "Transfer", log); err != nil {
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
