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

// HubWalletABI is the input ABI used to generate the binding from.
const HubWalletABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"currentPhase\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"freezePeriod\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"gulag\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"suspect\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"sharesTokenAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"genesisTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint64\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"PayDay\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"freezeQuote\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"frozenTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint64\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"frozenFunds\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lockPercent\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"DAO\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lockedFunds\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"rehub\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"Factory\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"Registration\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"inputs\":[{\"name\":\"_hubowner\",\"type\":\"address\"},{\"name\":\"_dao\",\"type\":\"address\"},{\"name\":\"_whitelist\",\"type\":\"address\"},{\"name\":\"sharesAddress\",\"type\":\"address\"}],\"payable\":false,\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"newPhase\",\"type\":\"uint8\"}],\"name\":\"LogPhaseSwitch\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"val\",\"type\":\"uint256\"}],\"name\":\"ToVal\",\"type\":\"event\"}]"

// HubWalletBin is the compiled bytecode used for deploying new contracts.
const HubWalletBin = `0x60606040526000600855341561001457600080fd5b604051608080610f78833981016040528080519190602001805191906020018051919060200180519150505b5b60008054600160a060020a03191633600160a060020a03161790555b60008054600160a060020a0319908116600160a060020a0387811691909117909255600180548216868416178155600380548316868516179055600280548316338516178155600b805467ffffffffffffffff19164267ffffffffffffffff161790556004805490931693851693909317909155670de0b6b3a76400006005908155601e600755600c55620d2f00600a55600e8054909160ff1990911690835b02179055505b505050505b610e61806101176000396000f300606060405236156100f65763ffffffff60e060020a600035041663055ad42e81146100fb5780630a3cb663146101325780630b3eeac8146101575780631e1683af1461016c57806327ebcf0e146101815780633ccfd60b146101b057806342c6498a146101c55780634d78511c146101f4578063565f6c49146102095780638da5cb5b1461022e578063906db9ff1461025d57806390a74e2c1461028c57806391030cb6146102b157806398fabd3a146102d6578063a9059cbb14610305578063b8afaa4814610329578063bd73820d1461034e578063c83dd23114610363578063e8a3791914610392578063f2fde38b146103b9575b600080fd5b341561010657600080fd5b61010e6103da565b6040518082600481111561011e57fe5b60ff16815260200191505060405180910390f35b341561013d57600080fd5b6101456103e3565b60405190815260200160405180910390f35b341561016257600080fd5b61016a6103e9565b005b341561017757600080fd5b61016a610519565b005b341561018c57600080fd5b6101946106b0565b604051600160a060020a03909116815260200160405180910390f35b34156101bb57600080fd5b61016a6106bf565b005b34156101d057600080fd5b6101d86107f8565b6040516001604060020a03909116815260200160405180910390f35b34156101ff57600080fd5b61016a610807565b005b341561021457600080fd5b6101456109c8565b60405190815260200160405180910390f35b341561023957600080fd5b6101946109ce565b604051600160a060020a03909116815260200160405180910390f35b341561026857600080fd5b6101d86109dd565b6040516001604060020a03909116815260200160405180910390f35b341561029757600080fd5b6101456109ec565b60405190815260200160405180910390f35b34156102bc57600080fd5b6101456109f2565b60405190815260200160405180910390f35b34156102e157600080fd5b6101946109f8565b604051600160a060020a03909116815260200160405180910390f35b341561031057600080fd5b61016a600160a060020a0360043516602435610a07565b005b341561033457600080fd5b610145610b77565b60405190815260200160405180910390f35b341561035957600080fd5b61016a610b7d565b005b341561036e57600080fd5b610194610c16565b604051600160a060020a03909116815260200160405180910390f35b341561039d57600080fd5b6103a5610c25565b604051901515815260200160405180910390f35b34156103c457600080fd5b61016a600160a060020a0360043516610dca565b005b600e5460ff1681565b600a5481565b60015460009033600160a060020a0390811691161461040757600080fd5b60035b600e5460ff16600481111561041b57fe5b1461042557600080fd5b600a546009546001604060020a03160142101561044157600080fd5b50600854600454600154600160a060020a039182169163a9059cbb91168360006040516020015260405160e060020a63ffffffff8516028152600160a060020a0390921660048301526024820152604401602060405180830381600087803b15156104ab57600080fd5b6102c65a03f115156104bc57600080fd5b50505060405180515050600e80546004919060ff19166001835b0217905550600e54600080516020610e168339815191529060ff166040518082600481111561050157fe5b60ff16815260200191505060405180910390a15b5b50565b60015433600160a060020a0390811691161461053457600080fd5b60015b600e5460ff16600481111561054857fe5b1461055257600080fd5b600454600160a060020a03166370a082313060006040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b15156105ab57600080fd5b6102c65a03f115156105bc57600080fd5b505050604051805160085550600980546001604060020a031916426001604060020a0316179055600e80546003919060ff19166001835b0217905550600e54600080516020610e168339815191529060ff166040518082600481111561061e57fe5b60ff16815260200191505060405180910390a1629e3400600a55600354600160a060020a0316634a5de1653060006040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b151561069157600080fd5b6102c65a03f115156106a257600080fd5b505050604051805150505b5b565b600454600160a060020a031681565b6000805433600160a060020a039081169116146106db57600080fd5b60025b600e5460ff1660048111156106ef57fe5b146106f957600080fd5b600454600160a060020a03166370a082313060006040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b151561075257600080fd5b6102c65a03f1151561076357600080fd5b505050604051805160045460008054929450600160a060020a03918216935063a9059cbb929091169084906040516020015260405160e060020a63ffffffff8516028152600160a060020a0390921660048301526024820152604401602060405180830381600087803b15156107d857600080fd5b6102c65a03f115156107e957600080fd5b505050604051805150505b5b50565b600b546001604060020a031681565b60005433600160a060020a0390811691161461082257600080fd5b60015b600e5460ff16600481111561083657fe5b1461084057600080fd5b600c546008546103e891025b6006805492909104909101600d55600090819055600855600a546009546001604060020a031601421061087e57600080fd5b600454600154600d54600160a060020a039283169263a9059cbb92169060006040516020015260405160e060020a63ffffffff8516028152600160a060020a0390921660048301526024820152604401602060405180830381600087803b15156108e757600080fd5b6102c65a03f115156108f857600080fd5b50505060405180515050600354600160a060020a0316634a5de1653060006040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b151561095b57600080fd5b6102c65a03f1151561096c57600080fd5b50505060405180515050600e80546002919060ff19166001835b0217905550600e54600080516020610e168339815191529060ff16604051808260048111156109b157fe5b60ff16815260200191505060405180910390a15b5b565b60055481565b600054600160a060020a031681565b6009546001604060020a031681565b60065481565b60075481565b600154600160a060020a031681565b6000805481908190819033600160a060020a03908116911614610a2957600080fd5b60015b600e5460ff166004811115610a3d57fe5b14610a4757600080fd5b60075460649086025b6008546006546004549390920496508601945084019250848603915081830190600160a060020a03166370a082313360006040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b1515610ac657600080fd5b6102c65a03f11515610ad757600080fd5b5050506040518051905010151515610aee57600080fd5b6008839055600454600160a060020a031663095ea7b3878360006040516020015260405160e060020a63ffffffff8516028152600160a060020a0390921660048301526024820152604401602060405180830381600087803b1515610b5257600080fd5b6102c65a03f11515610b6357600080fd5b505050604051805150505b5b505050505050565b60085481565b60015433600160a060020a03908116911614610b9857600080fd5b60035b600e5460ff166004811115610bac57fe5b14610bb657600080fd5b60006008819055600655600e80546002919060ff1916600183610986565b0217905550600e54600080516020610e168339815191529060ff16604051808260048111156109b157fe5b60ff16815260200191505060405180910390a15b5b565b600254600160a060020a031681565b600060025b600e5460ff166004811115610c3b57fe5b14610c4557600080fd5b600554600454600160a060020a03166370a082313060006040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b1515610ca157600080fd5b6102c65a03f11515610cb257600080fd5b50505060405180519050111515610cc857600080fd5b600554600655600980546001604060020a031916426001604060020a039081169190911791829055600354600160a060020a031691633f05852c9130911660006040516020015260405160e060020a63ffffffff8516028152600160a060020a0390921660048301526001604060020a03166024820152604401602060405180830381600087803b1515610d5b57600080fd5b6102c65a03f11515610d6c57600080fd5b50505060405180515050600e80546001919060ff191682805b0217905550600e54600080516020610e168339815191529060ff1660405180826004811115610db057fe5b60ff16815260200191505060405180910390a15060015b90565b60005433600160a060020a03908116911614610de557600080fd5b600160a060020a038116156105155760008054600160a060020a031916600160a060020a0383161790555b5b5b5056008d9efa3fab1bd6476defa44f520afbf9337886a4947021fd7f2775e0efaf4571a165627a7a72305820c80a85c0d933a161dbc79e43a2cef2b39c5ffc6a4d6203d9fc4ba250a33f4b1d0029`

// DeployHubWallet deploys a new Ethereum contract, binding an instance of HubWallet to it.
func DeployHubWallet(auth *bind.TransactOpts, backend bind.ContractBackend, _hubowner common.Address, _dao common.Address, _whitelist common.Address, sharesAddress common.Address) (common.Address, *types.Transaction, *HubWallet, error) {
	parsed, err := abi.JSON(strings.NewReader(HubWalletABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(HubWalletBin), backend, _hubowner, _dao, _whitelist, sharesAddress)
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

// DAO is a free data retrieval call binding the contract method 0x98fabd3a.
//
// Solidity: function DAO() constant returns(address)
func (_HubWallet *HubWalletCaller) DAO(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "DAO")
	return *ret0, err
}

// DAO is a free data retrieval call binding the contract method 0x98fabd3a.
//
// Solidity: function DAO() constant returns(address)
func (_HubWallet *HubWalletSession) DAO() (common.Address, error) {
	return _HubWallet.Contract.DAO(&_HubWallet.CallOpts)
}

// DAO is a free data retrieval call binding the contract method 0x98fabd3a.
//
// Solidity: function DAO() constant returns(address)
func (_HubWallet *HubWalletCallerSession) DAO() (common.Address, error) {
	return _HubWallet.Contract.DAO(&_HubWallet.CallOpts)
}

// Factory is a free data retrieval call binding the contract method 0xc83dd231.
//
// Solidity: function Factory() constant returns(address)
func (_HubWallet *HubWalletCaller) Factory(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "Factory")
	return *ret0, err
}

// Factory is a free data retrieval call binding the contract method 0xc83dd231.
//
// Solidity: function Factory() constant returns(address)
func (_HubWallet *HubWalletSession) Factory() (common.Address, error) {
	return _HubWallet.Contract.Factory(&_HubWallet.CallOpts)
}

// Factory is a free data retrieval call binding the contract method 0xc83dd231.
//
// Solidity: function Factory() constant returns(address)
func (_HubWallet *HubWalletCallerSession) Factory() (common.Address, error) {
	return _HubWallet.Contract.Factory(&_HubWallet.CallOpts)
}

// CurrentPhase is a free data retrieval call binding the contract method 0x055ad42e.
//
// Solidity: function currentPhase() constant returns(uint8)
func (_HubWallet *HubWalletCaller) CurrentPhase(opts *bind.CallOpts) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "currentPhase")
	return *ret0, err
}

// CurrentPhase is a free data retrieval call binding the contract method 0x055ad42e.
//
// Solidity: function currentPhase() constant returns(uint8)
func (_HubWallet *HubWalletSession) CurrentPhase() (uint8, error) {
	return _HubWallet.Contract.CurrentPhase(&_HubWallet.CallOpts)
}

// CurrentPhase is a free data retrieval call binding the contract method 0x055ad42e.
//
// Solidity: function currentPhase() constant returns(uint8)
func (_HubWallet *HubWalletCallerSession) CurrentPhase() (uint8, error) {
	return _HubWallet.Contract.CurrentPhase(&_HubWallet.CallOpts)
}

// FreezePeriod is a free data retrieval call binding the contract method 0x0a3cb663.
//
// Solidity: function freezePeriod() constant returns(uint256)
func (_HubWallet *HubWalletCaller) FreezePeriod(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "freezePeriod")
	return *ret0, err
}

// FreezePeriod is a free data retrieval call binding the contract method 0x0a3cb663.
//
// Solidity: function freezePeriod() constant returns(uint256)
func (_HubWallet *HubWalletSession) FreezePeriod() (*big.Int, error) {
	return _HubWallet.Contract.FreezePeriod(&_HubWallet.CallOpts)
}

// FreezePeriod is a free data retrieval call binding the contract method 0x0a3cb663.
//
// Solidity: function freezePeriod() constant returns(uint256)
func (_HubWallet *HubWalletCallerSession) FreezePeriod() (*big.Int, error) {
	return _HubWallet.Contract.FreezePeriod(&_HubWallet.CallOpts)
}

// FreezeQuote is a free data retrieval call binding the contract method 0x565f6c49.
//
// Solidity: function freezeQuote() constant returns(uint256)
func (_HubWallet *HubWalletCaller) FreezeQuote(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "freezeQuote")
	return *ret0, err
}

// FreezeQuote is a free data retrieval call binding the contract method 0x565f6c49.
//
// Solidity: function freezeQuote() constant returns(uint256)
func (_HubWallet *HubWalletSession) FreezeQuote() (*big.Int, error) {
	return _HubWallet.Contract.FreezeQuote(&_HubWallet.CallOpts)
}

// FreezeQuote is a free data retrieval call binding the contract method 0x565f6c49.
//
// Solidity: function freezeQuote() constant returns(uint256)
func (_HubWallet *HubWalletCallerSession) FreezeQuote() (*big.Int, error) {
	return _HubWallet.Contract.FreezeQuote(&_HubWallet.CallOpts)
}

// FrozenFunds is a free data retrieval call binding the contract method 0x90a74e2c.
//
// Solidity: function frozenFunds() constant returns(uint256)
func (_HubWallet *HubWalletCaller) FrozenFunds(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "frozenFunds")
	return *ret0, err
}

// FrozenFunds is a free data retrieval call binding the contract method 0x90a74e2c.
//
// Solidity: function frozenFunds() constant returns(uint256)
func (_HubWallet *HubWalletSession) FrozenFunds() (*big.Int, error) {
	return _HubWallet.Contract.FrozenFunds(&_HubWallet.CallOpts)
}

// FrozenFunds is a free data retrieval call binding the contract method 0x90a74e2c.
//
// Solidity: function frozenFunds() constant returns(uint256)
func (_HubWallet *HubWalletCallerSession) FrozenFunds() (*big.Int, error) {
	return _HubWallet.Contract.FrozenFunds(&_HubWallet.CallOpts)
}

// FrozenTime is a free data retrieval call binding the contract method 0x906db9ff.
//
// Solidity: function frozenTime() constant returns(uint64)
func (_HubWallet *HubWalletCaller) FrozenTime(opts *bind.CallOpts) (uint64, error) {
	var (
		ret0 = new(uint64)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "frozenTime")
	return *ret0, err
}

// FrozenTime is a free data retrieval call binding the contract method 0x906db9ff.
//
// Solidity: function frozenTime() constant returns(uint64)
func (_HubWallet *HubWalletSession) FrozenTime() (uint64, error) {
	return _HubWallet.Contract.FrozenTime(&_HubWallet.CallOpts)
}

// FrozenTime is a free data retrieval call binding the contract method 0x906db9ff.
//
// Solidity: function frozenTime() constant returns(uint64)
func (_HubWallet *HubWalletCallerSession) FrozenTime() (uint64, error) {
	return _HubWallet.Contract.FrozenTime(&_HubWallet.CallOpts)
}

// GenesisTime is a free data retrieval call binding the contract method 0x42c6498a.
//
// Solidity: function genesisTime() constant returns(uint64)
func (_HubWallet *HubWalletCaller) GenesisTime(opts *bind.CallOpts) (uint64, error) {
	var (
		ret0 = new(uint64)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "genesisTime")
	return *ret0, err
}

// GenesisTime is a free data retrieval call binding the contract method 0x42c6498a.
//
// Solidity: function genesisTime() constant returns(uint64)
func (_HubWallet *HubWalletSession) GenesisTime() (uint64, error) {
	return _HubWallet.Contract.GenesisTime(&_HubWallet.CallOpts)
}

// GenesisTime is a free data retrieval call binding the contract method 0x42c6498a.
//
// Solidity: function genesisTime() constant returns(uint64)
func (_HubWallet *HubWalletCallerSession) GenesisTime() (uint64, error) {
	return _HubWallet.Contract.GenesisTime(&_HubWallet.CallOpts)
}

// LockPercent is a free data retrieval call binding the contract method 0x91030cb6.
//
// Solidity: function lockPercent() constant returns(uint256)
func (_HubWallet *HubWalletCaller) LockPercent(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "lockPercent")
	return *ret0, err
}

// LockPercent is a free data retrieval call binding the contract method 0x91030cb6.
//
// Solidity: function lockPercent() constant returns(uint256)
func (_HubWallet *HubWalletSession) LockPercent() (*big.Int, error) {
	return _HubWallet.Contract.LockPercent(&_HubWallet.CallOpts)
}

// LockPercent is a free data retrieval call binding the contract method 0x91030cb6.
//
// Solidity: function lockPercent() constant returns(uint256)
func (_HubWallet *HubWalletCallerSession) LockPercent() (*big.Int, error) {
	return _HubWallet.Contract.LockPercent(&_HubWallet.CallOpts)
}

// LockedFunds is a free data retrieval call binding the contract method 0xb8afaa48.
//
// Solidity: function lockedFunds() constant returns(uint256)
func (_HubWallet *HubWalletCaller) LockedFunds(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "lockedFunds")
	return *ret0, err
}

// LockedFunds is a free data retrieval call binding the contract method 0xb8afaa48.
//
// Solidity: function lockedFunds() constant returns(uint256)
func (_HubWallet *HubWalletSession) LockedFunds() (*big.Int, error) {
	return _HubWallet.Contract.LockedFunds(&_HubWallet.CallOpts)
}

// LockedFunds is a free data retrieval call binding the contract method 0xb8afaa48.
//
// Solidity: function lockedFunds() constant returns(uint256)
func (_HubWallet *HubWalletCallerSession) LockedFunds() (*big.Int, error) {
	return _HubWallet.Contract.LockedFunds(&_HubWallet.CallOpts)
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

// SharesTokenAddress is a free data retrieval call binding the contract method 0x27ebcf0e.
//
// Solidity: function sharesTokenAddress() constant returns(address)
func (_HubWallet *HubWalletCaller) SharesTokenAddress(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "sharesTokenAddress")
	return *ret0, err
}

// SharesTokenAddress is a free data retrieval call binding the contract method 0x27ebcf0e.
//
// Solidity: function sharesTokenAddress() constant returns(address)
func (_HubWallet *HubWalletSession) SharesTokenAddress() (common.Address, error) {
	return _HubWallet.Contract.SharesTokenAddress(&_HubWallet.CallOpts)
}

// SharesTokenAddress is a free data retrieval call binding the contract method 0x27ebcf0e.
//
// Solidity: function sharesTokenAddress() constant returns(address)
func (_HubWallet *HubWalletCallerSession) SharesTokenAddress() (common.Address, error) {
	return _HubWallet.Contract.SharesTokenAddress(&_HubWallet.CallOpts)
}

// PayDay is a paid mutator transaction binding the contract method 0x4d78511c.
//
// Solidity: function PayDay() returns()
func (_HubWallet *HubWalletTransactor) PayDay(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "PayDay")
}

// PayDay is a paid mutator transaction binding the contract method 0x4d78511c.
//
// Solidity: function PayDay() returns()
func (_HubWallet *HubWalletSession) PayDay() (*types.Transaction, error) {
	return _HubWallet.Contract.PayDay(&_HubWallet.TransactOpts)
}

// PayDay is a paid mutator transaction binding the contract method 0x4d78511c.
//
// Solidity: function PayDay() returns()
func (_HubWallet *HubWalletTransactorSession) PayDay() (*types.Transaction, error) {
	return _HubWallet.Contract.PayDay(&_HubWallet.TransactOpts)
}

// Registration is a paid mutator transaction binding the contract method 0xe8a37919.
//
// Solidity: function Registration() returns(success bool)
func (_HubWallet *HubWalletTransactor) Registration(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "Registration")
}

// Registration is a paid mutator transaction binding the contract method 0xe8a37919.
//
// Solidity: function Registration() returns(success bool)
func (_HubWallet *HubWalletSession) Registration() (*types.Transaction, error) {
	return _HubWallet.Contract.Registration(&_HubWallet.TransactOpts)
}

// Registration is a paid mutator transaction binding the contract method 0xe8a37919.
//
// Solidity: function Registration() returns(success bool)
func (_HubWallet *HubWalletTransactorSession) Registration() (*types.Transaction, error) {
	return _HubWallet.Contract.Registration(&_HubWallet.TransactOpts)
}

// Gulag is a paid mutator transaction binding the contract method 0x0b3eeac8.
//
// Solidity: function gulag() returns()
func (_HubWallet *HubWalletTransactor) Gulag(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "gulag")
}

// Gulag is a paid mutator transaction binding the contract method 0x0b3eeac8.
//
// Solidity: function gulag() returns()
func (_HubWallet *HubWalletSession) Gulag() (*types.Transaction, error) {
	return _HubWallet.Contract.Gulag(&_HubWallet.TransactOpts)
}

// Gulag is a paid mutator transaction binding the contract method 0x0b3eeac8.
//
// Solidity: function gulag() returns()
func (_HubWallet *HubWalletTransactorSession) Gulag() (*types.Transaction, error) {
	return _HubWallet.Contract.Gulag(&_HubWallet.TransactOpts)
}

// Rehub is a paid mutator transaction binding the contract method 0xbd73820d.
//
// Solidity: function rehub() returns()
func (_HubWallet *HubWalletTransactor) Rehub(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "rehub")
}

// Rehub is a paid mutator transaction binding the contract method 0xbd73820d.
//
// Solidity: function rehub() returns()
func (_HubWallet *HubWalletSession) Rehub() (*types.Transaction, error) {
	return _HubWallet.Contract.Rehub(&_HubWallet.TransactOpts)
}

// Rehub is a paid mutator transaction binding the contract method 0xbd73820d.
//
// Solidity: function rehub() returns()
func (_HubWallet *HubWalletTransactorSession) Rehub() (*types.Transaction, error) {
	return _HubWallet.Contract.Rehub(&_HubWallet.TransactOpts)
}

// Suspect is a paid mutator transaction binding the contract method 0x1e1683af.
//
// Solidity: function suspect() returns()
func (_HubWallet *HubWalletTransactor) Suspect(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "suspect")
}

// Suspect is a paid mutator transaction binding the contract method 0x1e1683af.
//
// Solidity: function suspect() returns()
func (_HubWallet *HubWalletSession) Suspect() (*types.Transaction, error) {
	return _HubWallet.Contract.Suspect(&_HubWallet.TransactOpts)
}

// Suspect is a paid mutator transaction binding the contract method 0x1e1683af.
//
// Solidity: function suspect() returns()
func (_HubWallet *HubWalletTransactorSession) Suspect() (*types.Transaction, error) {
	return _HubWallet.Contract.Suspect(&_HubWallet.TransactOpts)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns()
func (_HubWallet *HubWalletTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns()
func (_HubWallet *HubWalletSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _HubWallet.Contract.Transfer(&_HubWallet.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns()
func (_HubWallet *HubWalletTransactorSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _HubWallet.Contract.Transfer(&_HubWallet.TransactOpts, _to, _value)
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

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_HubWallet *HubWalletTransactor) Withdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "withdraw")
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_HubWallet *HubWalletSession) Withdraw() (*types.Transaction, error) {
	return _HubWallet.Contract.Withdraw(&_HubWallet.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_HubWallet *HubWalletTransactorSession) Withdraw() (*types.Transaction, error) {
	return _HubWallet.Contract.Withdraw(&_HubWallet.TransactOpts)
}
