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

// TestnetFaucetABI is the input ABI used to generate the binding from.
const TestnetFaucetABI = "[{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"}],\"name\":\"OwnershipRenounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[],\"name\":\"getTokens\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"target\",\"type\":\"address\"},{\"name\":\"mintedAmount\",\"type\":\"uint256\"}],\"name\":\"mintToken\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getTokenAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// TestnetFaucetBin is the compiled bytecode used for deploying new contracts.
const TestnetFaucetBin = `0x608060405234801561001057600080fd5b5060008054600160a060020a031916331790553061002c6100f2565b600160a060020a03909116815260405190819003602001906000f080158015610059573d6000803e3d6000fd5b5060018054600160a060020a03928316600160a060020a031991821617918290556000805490911633178155604080517ff21cdf6f0000000000000000000000000000000000000000000000000000000081529051929093169263f21cdf6f926004808301939282900301818387803b1580156100d557600080fd5b505af11580156100e9573d6000803e3d6000fd5b50505050610102565b6040516109e5806104ed83390190565b6103dc806101116000396000f3006080604052600436106100775763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166310fe9ae88114610082578063715018a6146100b357806379c65068146100ca5780638da5cb5b14610102578063aa6ca80814610117578063f2fde38b1461012c575b61007f61014d565b50005b34801561008e57600080fd5b506100976101df565b60408051600160a060020a039092168252519081900360200190f35b3480156100bf57600080fd5b506100c86101ee565b005b3480156100d657600080fd5b506100ee600160a060020a036004351660243561025a565b604080519115158252519081900360200190f35b34801561010e57600080fd5b50610097610301565b34801561012357600080fd5b506100ee61014d565b34801561013857600080fd5b506100c8600160a060020a0360043516610310565b600154604080517f40c10f1900000000000000000000000000000000000000000000000000000000815233600482015268056bc75e2d6310000060248201529051600092600160a060020a0316916340c10f19916044808301928692919082900301818387803b1580156101c057600080fd5b505af11580156101d4573d6000803e3d6000fd5b505050506001905090565b600154600160a060020a031690565b600054600160a060020a0316331461020557600080fd5b60008054604051600160a060020a03909116917ff8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c6482091a26000805473ffffffffffffffffffffffffffffffffffffffff19169055565b60008054600160a060020a0316331461027257600080fd5b600154604080517f40c10f19000000000000000000000000000000000000000000000000000000008152600160a060020a03868116600483015260248201869052915191909216916340c10f1991604480830192600092919082900301818387803b1580156102e057600080fd5b505af11580156102f4573d6000803e3d6000fd5b5060019695505050505050565b600054600160a060020a031681565b600054600160a060020a0316331461032757600080fd5b61033081610333565b50565b600160a060020a038116151561034857600080fd5b60008054604051600160a060020a03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a36000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a03929092169190911790555600a165627a7a72305820670d109fbb28d9287e819333a04df2c08a2d5476c4f56b8bcaf86a398b656e5a002960c0604052600a60808190527f534f4e4d20546f6b656e0000000000000000000000000000000000000000000060a090815261003e91600391906100f2565b506040805180820190915260038082527f534e4d00000000000000000000000000000000000000000000000000000000006020909201918252610083916004916100f2565b5060126005556006805460a060020a60ff021916740100000000000000000000000000000000000000001790553480156100bc57600080fd5b506040516020806109e5833981016040525160068054600160a060020a031916600160a060020a0390921691909117905561018d565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061013357805160ff1916838001178555610160565b82800160010185558215610160579182015b82811115610160578251825591602001919060010190610145565b5061016c929150610170565b5090565b61018a91905b8082111561016c5760008155600101610176565b90565b6108498061019c6000396000f3006080604052600436106100c45763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306fdde0381146100c9578063095ea7b31461015357806318160ddd1461017957806323b872dd146101a0578063313ce567146101ca57806340c10f19146101df5780635d4522011461020357806370a082311461023457806395d89b4114610255578063a9059cbb1461026a578063ca67065f1461028e578063dd62ed3e146102b7578063f21cdf6f146102de575b600080fd5b3480156100d557600080fd5b506100de6102f3565b6040805160208082528351818301528351919283929083019185019080838360005b83811015610118578181015183820152602001610100565b50505050905090810190601f1680156101455780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561015f57600080fd5b50610177600160a060020a0360043516602435610381565b005b34801561018557600080fd5b5061018e6103b7565b60408051918252519081900360200190f35b3480156101ac57600080fd5b50610177600160a060020a03600435811690602435166044356103bd565b3480156101d657600080fd5b5061018e6103f5565b3480156101eb57600080fd5b50610177600160a060020a03600435166024356103fb565b34801561020f57600080fd5b50610218610499565b60408051600160a060020a039092168252519081900360200190f35b34801561024057600080fd5b5061018e600160a060020a03600435166104a8565b34801561026157600080fd5b506100de6104c3565b34801561027657600080fd5b50610177600160a060020a036004351660243561051e565b34801561029a57600080fd5b506102a3610550565b604080519115158252519081900360200190f35b3480156102c357600080fd5b5061018e600160a060020a0360043581169060243516610571565b3480156102ea57600080fd5b5061017761059c565b6003805460408051602060026001851615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156103795780601f1061034e57610100808354040283529160200191610379565b820191906000526020600020905b81548152906001019060200180831161035c57829003601f168201915b505050505081565b60065474010000000000000000000000000000000000000000900460ff16156103a957600080fd5b6103b382826105d3565b5050565b60005481565b60065474010000000000000000000000000000000000000000900460ff16156103e557600080fd5b6103f0838383610635565b505050565b60055481565b600654600160a060020a0316331461041257600080fd5b80151561041e57600080fd5b6000546b016f44a83aab6c233c000000908201111561043c57600080fd5b600160a060020a0382166000818152600160209081526040808320805486019055825485018355805185815290517fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef929181900390910190a35050565b600654600160a060020a031681565b600160a060020a031660009081526001602052604090205490565b6004805460408051602060026001851615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156103795780601f1061034e57610100808354040283529160200191610379565b60065474010000000000000000000000000000000000000000900460ff161561054657600080fd5b6103b3828261073c565b60065474010000000000000000000000000000000000000000900460ff1681565b600160a060020a03918216600090815260026020908152604080832093909416825291909152205490565b600654600160a060020a031633146105b357600080fd5b6006805474ff000000000000000000000000000000000000000019169055565b336000818152600260209081526040808320600160a060020a03871680855290835292819020859055805185815290519293927f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925929181900390910190a35050565b600160a060020a03808416600090815260026020908152604080832033845282528083205493861683526001909152902054610677908363ffffffff6107f816565b600160a060020a0380851660009081526001602052604080822093909355908616815220546106ac908363ffffffff61080b16565b600160a060020a0385166000908152600160205260409020556106d5818363ffffffff61080b16565b600160a060020a03808616600081815260026020908152604080832033845282529182902094909455805186815290519287169391927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef929181900390910190a350505050565b6040604436101561074c57600080fd5b3360009081526001602052604090205461076c908363ffffffff61080b16565b3360009081526001602052604080822092909255600160a060020a0385168152205461079e908363ffffffff6107f816565b600160a060020a0384166000818152600160209081526040918290209390935580518581529051919233927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9281900390910190a3505050565b8181018281101561080557fe5b92915050565b60008282111561081757fe5b509003905600a165627a7a723058204500c0d8f5bbdf8d7c79234cb49c5740cc4a9114985c151434c7210e5156ef470029`

// DeployTestnetFaucet deploys a new Ethereum contract, binding an instance of TestnetFaucet to it.
func DeployTestnetFaucet(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *TestnetFaucet, error) {
	parsed, err := abi.JSON(strings.NewReader(TestnetFaucetABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(TestnetFaucetBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &TestnetFaucet{TestnetFaucetCaller: TestnetFaucetCaller{contract: contract}, TestnetFaucetTransactor: TestnetFaucetTransactor{contract: contract}, TestnetFaucetFilterer: TestnetFaucetFilterer{contract: contract}}, nil
}

// TestnetFaucet is an auto generated Go binding around an Ethereum contract.
type TestnetFaucet struct {
	TestnetFaucetCaller     // Read-only binding to the contract
	TestnetFaucetTransactor // Write-only binding to the contract
	TestnetFaucetFilterer   // Log filterer for contract events
}

// TestnetFaucetCaller is an auto generated read-only Go binding around an Ethereum contract.
type TestnetFaucetCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestnetFaucetTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TestnetFaucetTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestnetFaucetFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TestnetFaucetFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestnetFaucetSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TestnetFaucetSession struct {
	Contract     *TestnetFaucet    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TestnetFaucetCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TestnetFaucetCallerSession struct {
	Contract *TestnetFaucetCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// TestnetFaucetTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TestnetFaucetTransactorSession struct {
	Contract     *TestnetFaucetTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// TestnetFaucetRaw is an auto generated low-level Go binding around an Ethereum contract.
type TestnetFaucetRaw struct {
	Contract *TestnetFaucet // Generic contract binding to access the raw methods on
}

// TestnetFaucetCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TestnetFaucetCallerRaw struct {
	Contract *TestnetFaucetCaller // Generic read-only contract binding to access the raw methods on
}

// TestnetFaucetTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TestnetFaucetTransactorRaw struct {
	Contract *TestnetFaucetTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTestnetFaucet creates a new instance of TestnetFaucet, bound to a specific deployed contract.
func NewTestnetFaucet(address common.Address, backend bind.ContractBackend) (*TestnetFaucet, error) {
	contract, err := bindTestnetFaucet(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TestnetFaucet{TestnetFaucetCaller: TestnetFaucetCaller{contract: contract}, TestnetFaucetTransactor: TestnetFaucetTransactor{contract: contract}, TestnetFaucetFilterer: TestnetFaucetFilterer{contract: contract}}, nil
}

// NewTestnetFaucetCaller creates a new read-only instance of TestnetFaucet, bound to a specific deployed contract.
func NewTestnetFaucetCaller(address common.Address, caller bind.ContractCaller) (*TestnetFaucetCaller, error) {
	contract, err := bindTestnetFaucet(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TestnetFaucetCaller{contract: contract}, nil
}

// NewTestnetFaucetTransactor creates a new write-only instance of TestnetFaucet, bound to a specific deployed contract.
func NewTestnetFaucetTransactor(address common.Address, transactor bind.ContractTransactor) (*TestnetFaucetTransactor, error) {
	contract, err := bindTestnetFaucet(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TestnetFaucetTransactor{contract: contract}, nil
}

// NewTestnetFaucetFilterer creates a new log filterer instance of TestnetFaucet, bound to a specific deployed contract.
func NewTestnetFaucetFilterer(address common.Address, filterer bind.ContractFilterer) (*TestnetFaucetFilterer, error) {
	contract, err := bindTestnetFaucet(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TestnetFaucetFilterer{contract: contract}, nil
}

// bindTestnetFaucet binds a generic wrapper to an already deployed contract.
func bindTestnetFaucet(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(TestnetFaucetABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestnetFaucet *TestnetFaucetRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _TestnetFaucet.Contract.TestnetFaucetCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestnetFaucet *TestnetFaucetRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestnetFaucet.Contract.TestnetFaucetTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestnetFaucet *TestnetFaucetRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestnetFaucet.Contract.TestnetFaucetTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestnetFaucet *TestnetFaucetCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _TestnetFaucet.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestnetFaucet *TestnetFaucetTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestnetFaucet.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestnetFaucet *TestnetFaucetTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestnetFaucet.Contract.contract.Transact(opts, method, params...)
}

// GetTokenAddress is a free data retrieval call binding the contract method 0x10fe9ae8.
//
// Solidity: function getTokenAddress() constant returns(address)
func (_TestnetFaucet *TestnetFaucetCaller) GetTokenAddress(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _TestnetFaucet.contract.Call(opts, out, "getTokenAddress")
	return *ret0, err
}

// GetTokenAddress is a free data retrieval call binding the contract method 0x10fe9ae8.
//
// Solidity: function getTokenAddress() constant returns(address)
func (_TestnetFaucet *TestnetFaucetSession) GetTokenAddress() (common.Address, error) {
	return _TestnetFaucet.Contract.GetTokenAddress(&_TestnetFaucet.CallOpts)
}

// GetTokenAddress is a free data retrieval call binding the contract method 0x10fe9ae8.
//
// Solidity: function getTokenAddress() constant returns(address)
func (_TestnetFaucet *TestnetFaucetCallerSession) GetTokenAddress() (common.Address, error) {
	return _TestnetFaucet.Contract.GetTokenAddress(&_TestnetFaucet.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_TestnetFaucet *TestnetFaucetCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _TestnetFaucet.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_TestnetFaucet *TestnetFaucetSession) Owner() (common.Address, error) {
	return _TestnetFaucet.Contract.Owner(&_TestnetFaucet.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_TestnetFaucet *TestnetFaucetCallerSession) Owner() (common.Address, error) {
	return _TestnetFaucet.Contract.Owner(&_TestnetFaucet.CallOpts)
}

// GetTokens is a paid mutator transaction binding the contract method 0xaa6ca808.
//
// Solidity: function getTokens() returns(bool)
func (_TestnetFaucet *TestnetFaucetTransactor) GetTokens(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestnetFaucet.contract.Transact(opts, "getTokens")
}

// GetTokens is a paid mutator transaction binding the contract method 0xaa6ca808.
//
// Solidity: function getTokens() returns(bool)
func (_TestnetFaucet *TestnetFaucetSession) GetTokens() (*types.Transaction, error) {
	return _TestnetFaucet.Contract.GetTokens(&_TestnetFaucet.TransactOpts)
}

// GetTokens is a paid mutator transaction binding the contract method 0xaa6ca808.
//
// Solidity: function getTokens() returns(bool)
func (_TestnetFaucet *TestnetFaucetTransactorSession) GetTokens() (*types.Transaction, error) {
	return _TestnetFaucet.Contract.GetTokens(&_TestnetFaucet.TransactOpts)
}

// MintToken is a paid mutator transaction binding the contract method 0x79c65068.
//
// Solidity: function mintToken(target address, mintedAmount uint256) returns(bool)
func (_TestnetFaucet *TestnetFaucetTransactor) MintToken(opts *bind.TransactOpts, target common.Address, mintedAmount *big.Int) (*types.Transaction, error) {
	return _TestnetFaucet.contract.Transact(opts, "mintToken", target, mintedAmount)
}

// MintToken is a paid mutator transaction binding the contract method 0x79c65068.
//
// Solidity: function mintToken(target address, mintedAmount uint256) returns(bool)
func (_TestnetFaucet *TestnetFaucetSession) MintToken(target common.Address, mintedAmount *big.Int) (*types.Transaction, error) {
	return _TestnetFaucet.Contract.MintToken(&_TestnetFaucet.TransactOpts, target, mintedAmount)
}

// MintToken is a paid mutator transaction binding the contract method 0x79c65068.
//
// Solidity: function mintToken(target address, mintedAmount uint256) returns(bool)
func (_TestnetFaucet *TestnetFaucetTransactorSession) MintToken(target common.Address, mintedAmount *big.Int) (*types.Transaction, error) {
	return _TestnetFaucet.Contract.MintToken(&_TestnetFaucet.TransactOpts, target, mintedAmount)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_TestnetFaucet *TestnetFaucetTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestnetFaucet.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_TestnetFaucet *TestnetFaucetSession) RenounceOwnership() (*types.Transaction, error) {
	return _TestnetFaucet.Contract.RenounceOwnership(&_TestnetFaucet.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_TestnetFaucet *TestnetFaucetTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _TestnetFaucet.Contract.RenounceOwnership(&_TestnetFaucet.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_TestnetFaucet *TestnetFaucetTransactor) TransferOwnership(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _TestnetFaucet.contract.Transact(opts, "transferOwnership", _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_TestnetFaucet *TestnetFaucetSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _TestnetFaucet.Contract.TransferOwnership(&_TestnetFaucet.TransactOpts, _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_TestnetFaucet *TestnetFaucetTransactorSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _TestnetFaucet.Contract.TransferOwnership(&_TestnetFaucet.TransactOpts, _newOwner)
}

// TestnetFaucetOwnershipRenouncedIterator is returned from FilterOwnershipRenounced and is used to iterate over the raw logs and unpacked data for OwnershipRenounced events raised by the TestnetFaucet contract.
type TestnetFaucetOwnershipRenouncedIterator struct {
	Event *TestnetFaucetOwnershipRenounced // Event containing the contract specifics and raw log

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
func (it *TestnetFaucetOwnershipRenouncedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TestnetFaucetOwnershipRenounced)
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
		it.Event = new(TestnetFaucetOwnershipRenounced)
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
func (it *TestnetFaucetOwnershipRenouncedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TestnetFaucetOwnershipRenouncedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TestnetFaucetOwnershipRenounced represents a OwnershipRenounced event raised by the TestnetFaucet contract.
type TestnetFaucetOwnershipRenounced struct {
	PreviousOwner common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipRenounced is a free log retrieval operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_TestnetFaucet *TestnetFaucetFilterer) FilterOwnershipRenounced(opts *bind.FilterOpts, previousOwner []common.Address) (*TestnetFaucetOwnershipRenouncedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _TestnetFaucet.contract.FilterLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return &TestnetFaucetOwnershipRenouncedIterator{contract: _TestnetFaucet.contract, event: "OwnershipRenounced", logs: logs, sub: sub}, nil
}

// WatchOwnershipRenounced is a free log subscription operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_TestnetFaucet *TestnetFaucetFilterer) WatchOwnershipRenounced(opts *bind.WatchOpts, sink chan<- *TestnetFaucetOwnershipRenounced, previousOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _TestnetFaucet.contract.WatchLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TestnetFaucetOwnershipRenounced)
				if err := _TestnetFaucet.contract.UnpackLog(event, "OwnershipRenounced", log); err != nil {
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

// TestnetFaucetOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the TestnetFaucet contract.
type TestnetFaucetOwnershipTransferredIterator struct {
	Event *TestnetFaucetOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *TestnetFaucetOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TestnetFaucetOwnershipTransferred)
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
		it.Event = new(TestnetFaucetOwnershipTransferred)
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
func (it *TestnetFaucetOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TestnetFaucetOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TestnetFaucetOwnershipTransferred represents a OwnershipTransferred event raised by the TestnetFaucet contract.
type TestnetFaucetOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_TestnetFaucet *TestnetFaucetFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*TestnetFaucetOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _TestnetFaucet.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &TestnetFaucetOwnershipTransferredIterator{contract: _TestnetFaucet.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_TestnetFaucet *TestnetFaucetFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *TestnetFaucetOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _TestnetFaucet.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TestnetFaucetOwnershipTransferred)
				if err := _TestnetFaucet.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
