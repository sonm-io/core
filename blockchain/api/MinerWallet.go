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

// MinerWalletABI is the input ABI used to generate the binding from.
const MinerWalletABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"currentPhase\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"freezePeriod\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"gulag\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"suspect\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"stakeShare\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"sharesTokenAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"genesisTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint64\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"PayDay\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"freezeQuote\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hubwallet\",\"type\":\"address\"}],\"name\":\"pullMoney\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"frozenTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint64\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"frozenFunds\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"DAO\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lockedFunds\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"rehub\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"Factory\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"stake\",\"type\":\"uint256\"}],\"name\":\"Registration\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"inputs\":[{\"name\":\"_minowner\",\"type\":\"address\"},{\"name\":\"_dao\",\"type\":\"address\"},{\"name\":\"_whitelist\",\"type\":\"address\"},{\"name\":\"_sharesAddress\",\"type\":\"address\"}],\"payable\":false,\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"newPhase\",\"type\":\"uint8\"}],\"name\":\"LogPhaseSwitch\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"hub\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"PulledMoney\",\"type\":\"event\"}]"

// MinerWalletBin is the compiled bytecode used for deploying new contracts.
const MinerWalletBin = `0x60606040526000600855600d80546002919060ff19166001835b0217905550341561002957600080fd5b604051608080611007833981016040528080519190602001805191906020018051919060200180519150505b5b60008054600160a060020a03191633600160a060020a03161790555b60008054600160a060020a0319908116600160a060020a0387811691909117909255600180548216868416179055600380548216858416179055600280548216338416179055600b805467ffffffffffffffff19164267ffffffffffffffff1617905560048054909116918316919091179055670de0b6b3a76400006005908155600c5562069780600a555b505050505b610ef5806101126000396000f300606060405236156100f65763ffffffff60e060020a600035041663055ad42e81146100fb5780630a3cb663146101325780630b3eeac8146101575780631e1683af1461016c5780631ea41c2c1461018157806327ebcf0e146101a65780633ccfd60b146101d557806342c6498a146101ea5780634d78511c14610219578063565f6c491461022e5780635caf77d9146102535780638da5cb5b14610274578063906db9ff146102a357806390a74e2c146102d257806398fabd3a146102f7578063b8afaa4814610326578063bd73820d1461034b578063c83dd23114610360578063dd1dcd9f1461038f578063f2fde38b146103b9575b600080fd5b341561010657600080fd5b61010e6103da565b6040518082600481111561011e57fe5b60ff16815260200191505060405180910390f35b341561013d57600080fd5b6101456103e3565b60405190815260200160405180910390f35b341561016257600080fd5b61016a6103e9565b005b341561017757600080fd5b61016a610519565b005b341561018c57600080fd5b6101456106b0565b60405190815260200160405180910390f35b34156101b157600080fd5b6101b96106b6565b604051600160a060020a03909116815260200160405180910390f35b34156101e057600080fd5b61016a6106c5565b005b34156101f557600080fd5b6101fd610869565b6040516001604060020a03909116815260200160405180910390f35b341561022457600080fd5b61016a610878565b005b341561023957600080fd5b610145610a38565b60405190815260200160405180910390f35b341561025e57600080fd5b61016a600160a060020a0360043516610a3e565b005b341561027f57600080fd5b6101b9610bab565b604051600160a060020a03909116815260200160405180910390f35b34156102ae57600080fd5b6101fd610bba565b6040516001604060020a03909116815260200160405180910390f35b34156102dd57600080fd5b610145610bc9565b60405190815260200160405180910390f35b341561030257600080fd5b6101b9610bcf565b604051600160a060020a03909116815260200160405180910390f35b341561033157600080fd5b610145610bde565b60405190815260200160405180910390f35b341561035657600080fd5b61016a610be4565b005b341561036b57600080fd5b6101b9610c7e565b604051600160a060020a03909116815260200160405180910390f35b341561039a57600080fd5b6103a5600435610c8d565b604051901515815260200160405180910390f35b34156103c457600080fd5b61016a600160a060020a0360043516610e5e565b005b600d5460ff1681565b600a5481565b60015460009033600160a060020a0390811691161461040757600080fd5b60035b600d5460ff16600481111561041b57fe5b1461042557600080fd5b600a546009546001604060020a03160142101561044157600080fd5b50600854600454600154600160a060020a039182169163a9059cbb91168360006040516020015260405160e060020a63ffffffff8516028152600160a060020a0390921660048301526024820152604401602060405180830381600087803b15156104ab57600080fd5b6102c65a03f115156104bc57600080fd5b50505060405180515050600d80546004919060ff19166001835b0217905550600d54600080516020610eaa8339815191529060ff166040518082600481111561050157fe5b60ff16815260200191505060405180910390a15b5b50565b60015433600160a060020a0390811691161461053457600080fd5b60015b600d5460ff16600481111561054857fe5b1461055257600080fd5b600454600160a060020a03166370a082313060006040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b15156105ab57600080fd5b6102c65a03f115156105bc57600080fd5b505050604051805160085550600980546001604060020a031916426001604060020a0316179055600d80546003919060ff19166001835b0217905550600d54600080516020610eaa8339815191529060ff166040518082600481111561061e57fe5b60ff16815260200191505060405180910390a1629e3400600a55600354600160a060020a031663aaf22fe43060006040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b151561069157600080fd5b6102c65a03f115156106a257600080fd5b505050604051805150505b5b565b60075481565b600454600160a060020a031681565b6000805433600160a060020a039081169116146106e157600080fd5b600754600454600160a060020a03166370a082313360006040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b151561073d57600080fd5b6102c65a03f1151561074e57600080fd5b505050604051805190501015151561076557600080fd5b600454600160a060020a03166370a082313060006040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b15156107be57600080fd5b6102c65a03f115156107cf57600080fd5b505050604051805160075460045460008054929093039450600160a060020a03908116935063a9059cbb9291169084906040516020015260405160e060020a63ffffffff8516028152600160a060020a0390921660048301526024820152604401602060405180830381600087803b151561084957600080fd5b6102c65a03f1151561085a57600080fd5b505050604051805150505b5b50565b600b546001604060020a031681565b6000805433600160a060020a0390811691161461089457600080fd5b60015b600d5460ff1660048111156108a857fe5b146108b257600080fd5b600a546009546001604060020a0316014210156108ce57600080fd5b600c546006546103e891025b600060068190556007819055600454600154939092049350600160a060020a039182169263a9059cbb92169084906040516020015260405160e060020a63ffffffff8516028152600160a060020a0390921660048301526024820152604401602060405180830381600087803b151561095257600080fd5b6102c65a03f1151561096357600080fd5b50505060405180515050600354600160a060020a031663aaf22fe43060006040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b15156109c657600080fd5b6102c65a03f115156109d757600080fd5b50505060405180515050600d80546002919060ff19166001836104d6565b0217905550600d54600080516020610eaa8339815191529060ff166040518082600481111561050157fe5b60ff16815260200191505060405180910390a15b5b50565b60055481565b6000805433600160a060020a03908116911614610a5a57600080fd5b600454600160a060020a031663dd62ed3e833060006040516020015260405160e060020a63ffffffff8516028152600160a060020a03928316600482015291166024820152604401602060405180830381600087803b1515610abb57600080fd5b6102c65a03f11515610acc57600080fd5b5050506040518051600454909250600160a060020a031690506323b872dd83308460006040516020015260405160e060020a63ffffffff8616028152600160a060020a0393841660048201529190921660248201526044810191909152606401602060405180830381600087803b1515610b4557600080fd5b6102c65a03f11515610b5657600080fd5b50505060405180519050507f419b5dbcc6505f17449a288ecca83a531350bf94db30fcd124c25b18792fcf938282604051600160a060020a03909216825260208201526040908101905180910390a15b5b5050565b600054600160a060020a031681565b6009546001604060020a031681565b60065481565b600154600160a060020a031681565b60085481565b60015433600160a060020a03908116911614610bff57600080fd5b60035b600d5460ff166004811115610c1357fe5b14610c1d57600080fd5b600060088190556006819055600755600d80546002919060ff19166001835b0217905550600d54600080516020610eaa8339815191529060ff1660405180826004811115610c6757fe5b60ff16815260200191505060405180910390a15b5b565b600254600160a060020a031681565b6000805433600160a060020a03908116911614610ca957600080fd5b60025b600d5460ff166004811115610cbd57fe5b14610cc757600080fd5b600554600454600160a060020a03166370a082313060006040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b1515610d2357600080fd5b6102c65a03f11515610d3457600080fd5b50505060405180519050111515610d4a57600080fd5b60078290556005548201600655600980546001604060020a031916426001604060020a039081169190911791829055600354600160a060020a031691630941d2e6913091168560006040516020015260405160e060020a63ffffffff8616028152600160a060020a0390931660048401526001604060020a0390911660248301526044820152606401602060405180830381600087803b1515610dec57600080fd5b6102c65a03f11515610dfd57600080fd5b50505060405180515050600d80546001919060ff191682805b0217905550600d54600080516020610eaa8339815191529060ff1660405180826004811115610e4157fe5b60ff16815260200191505060405180910390a15060015b5b919050565b60005433600160a060020a03908116911614610e7957600080fd5b600160a060020a038116156105155760008054600160a060020a031916600160a060020a0383161790555b5b5b5056008d9efa3fab1bd6476defa44f520afbf9337886a4947021fd7f2775e0efaf4571a165627a7a723058202a6737669838233e7e484e0ddb32d3fec357a76e1801134aea89dc03cad4c3e80029`

// DeployMinerWallet deploys a new Ethereum contract, binding an instance of MinerWallet to it.
func DeployMinerWallet(auth *bind.TransactOpts, backend bind.ContractBackend, _minowner common.Address, _dao common.Address, _whitelist common.Address, _sharesAddress common.Address) (common.Address, *types.Transaction, *MinerWallet, error) {
	parsed, err := abi.JSON(strings.NewReader(MinerWalletABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(MinerWalletBin), backend, _minowner, _dao, _whitelist, _sharesAddress)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &MinerWallet{MinerWalletCaller: MinerWalletCaller{contract: contract}, MinerWalletTransactor: MinerWalletTransactor{contract: contract}}, nil
}

// MinerWallet is an auto generated Go binding around an Ethereum contract.
type MinerWallet struct {
	MinerWalletCaller     // Read-only binding to the contract
	MinerWalletTransactor // Write-only binding to the contract
}

// MinerWalletCaller is an auto generated read-only Go binding around an Ethereum contract.
type MinerWalletCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MinerWalletTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MinerWalletTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MinerWalletSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MinerWalletSession struct {
	Contract     *MinerWallet      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MinerWalletCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MinerWalletCallerSession struct {
	Contract *MinerWalletCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// MinerWalletTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MinerWalletTransactorSession struct {
	Contract     *MinerWalletTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// MinerWalletRaw is an auto generated low-level Go binding around an Ethereum contract.
type MinerWalletRaw struct {
	Contract *MinerWallet // Generic contract binding to access the raw methods on
}

// MinerWalletCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MinerWalletCallerRaw struct {
	Contract *MinerWalletCaller // Generic read-only contract binding to access the raw methods on
}

// MinerWalletTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MinerWalletTransactorRaw struct {
	Contract *MinerWalletTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMinerWallet creates a new instance of MinerWallet, bound to a specific deployed contract.
func NewMinerWallet(address common.Address, backend bind.ContractBackend) (*MinerWallet, error) {
	contract, err := bindMinerWallet(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MinerWallet{MinerWalletCaller: MinerWalletCaller{contract: contract}, MinerWalletTransactor: MinerWalletTransactor{contract: contract}}, nil
}

// NewMinerWalletCaller creates a new read-only instance of MinerWallet, bound to a specific deployed contract.
func NewMinerWalletCaller(address common.Address, caller bind.ContractCaller) (*MinerWalletCaller, error) {
	contract, err := bindMinerWallet(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &MinerWalletCaller{contract: contract}, nil
}

// NewMinerWalletTransactor creates a new write-only instance of MinerWallet, bound to a specific deployed contract.
func NewMinerWalletTransactor(address common.Address, transactor bind.ContractTransactor) (*MinerWalletTransactor, error) {
	contract, err := bindMinerWallet(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &MinerWalletTransactor{contract: contract}, nil
}

// bindMinerWallet binds a generic wrapper to an already deployed contract.
func bindMinerWallet(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MinerWalletABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MinerWallet *MinerWalletRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MinerWallet.Contract.MinerWalletCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MinerWallet *MinerWalletRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MinerWallet.Contract.MinerWalletTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MinerWallet *MinerWalletRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MinerWallet.Contract.MinerWalletTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MinerWallet *MinerWalletCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MinerWallet.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MinerWallet *MinerWalletTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MinerWallet.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MinerWallet *MinerWalletTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MinerWallet.Contract.contract.Transact(opts, method, params...)
}

// DAO is a free data retrieval call binding the contract method 0x98fabd3a.
//
// Solidity: function DAO() constant returns(address)
func (_MinerWallet *MinerWalletCaller) DAO(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _MinerWallet.contract.Call(opts, out, "DAO")
	return *ret0, err
}

// DAO is a free data retrieval call binding the contract method 0x98fabd3a.
//
// Solidity: function DAO() constant returns(address)
func (_MinerWallet *MinerWalletSession) DAO() (common.Address, error) {
	return _MinerWallet.Contract.DAO(&_MinerWallet.CallOpts)
}

// DAO is a free data retrieval call binding the contract method 0x98fabd3a.
//
// Solidity: function DAO() constant returns(address)
func (_MinerWallet *MinerWalletCallerSession) DAO() (common.Address, error) {
	return _MinerWallet.Contract.DAO(&_MinerWallet.CallOpts)
}

// Factory is a free data retrieval call binding the contract method 0xc83dd231.
//
// Solidity: function Factory() constant returns(address)
func (_MinerWallet *MinerWalletCaller) Factory(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _MinerWallet.contract.Call(opts, out, "Factory")
	return *ret0, err
}

// Factory is a free data retrieval call binding the contract method 0xc83dd231.
//
// Solidity: function Factory() constant returns(address)
func (_MinerWallet *MinerWalletSession) Factory() (common.Address, error) {
	return _MinerWallet.Contract.Factory(&_MinerWallet.CallOpts)
}

// Factory is a free data retrieval call binding the contract method 0xc83dd231.
//
// Solidity: function Factory() constant returns(address)
func (_MinerWallet *MinerWalletCallerSession) Factory() (common.Address, error) {
	return _MinerWallet.Contract.Factory(&_MinerWallet.CallOpts)
}

// CurrentPhase is a free data retrieval call binding the contract method 0x055ad42e.
//
// Solidity: function currentPhase() constant returns(uint8)
func (_MinerWallet *MinerWalletCaller) CurrentPhase(opts *bind.CallOpts) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _MinerWallet.contract.Call(opts, out, "currentPhase")
	return *ret0, err
}

// CurrentPhase is a free data retrieval call binding the contract method 0x055ad42e.
//
// Solidity: function currentPhase() constant returns(uint8)
func (_MinerWallet *MinerWalletSession) CurrentPhase() (uint8, error) {
	return _MinerWallet.Contract.CurrentPhase(&_MinerWallet.CallOpts)
}

// CurrentPhase is a free data retrieval call binding the contract method 0x055ad42e.
//
// Solidity: function currentPhase() constant returns(uint8)
func (_MinerWallet *MinerWalletCallerSession) CurrentPhase() (uint8, error) {
	return _MinerWallet.Contract.CurrentPhase(&_MinerWallet.CallOpts)
}

// FreezePeriod is a free data retrieval call binding the contract method 0x0a3cb663.
//
// Solidity: function freezePeriod() constant returns(uint256)
func (_MinerWallet *MinerWalletCaller) FreezePeriod(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MinerWallet.contract.Call(opts, out, "freezePeriod")
	return *ret0, err
}

// FreezePeriod is a free data retrieval call binding the contract method 0x0a3cb663.
//
// Solidity: function freezePeriod() constant returns(uint256)
func (_MinerWallet *MinerWalletSession) FreezePeriod() (*big.Int, error) {
	return _MinerWallet.Contract.FreezePeriod(&_MinerWallet.CallOpts)
}

// FreezePeriod is a free data retrieval call binding the contract method 0x0a3cb663.
//
// Solidity: function freezePeriod() constant returns(uint256)
func (_MinerWallet *MinerWalletCallerSession) FreezePeriod() (*big.Int, error) {
	return _MinerWallet.Contract.FreezePeriod(&_MinerWallet.CallOpts)
}

// FreezeQuote is a free data retrieval call binding the contract method 0x565f6c49.
//
// Solidity: function freezeQuote() constant returns(uint256)
func (_MinerWallet *MinerWalletCaller) FreezeQuote(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MinerWallet.contract.Call(opts, out, "freezeQuote")
	return *ret0, err
}

// FreezeQuote is a free data retrieval call binding the contract method 0x565f6c49.
//
// Solidity: function freezeQuote() constant returns(uint256)
func (_MinerWallet *MinerWalletSession) FreezeQuote() (*big.Int, error) {
	return _MinerWallet.Contract.FreezeQuote(&_MinerWallet.CallOpts)
}

// FreezeQuote is a free data retrieval call binding the contract method 0x565f6c49.
//
// Solidity: function freezeQuote() constant returns(uint256)
func (_MinerWallet *MinerWalletCallerSession) FreezeQuote() (*big.Int, error) {
	return _MinerWallet.Contract.FreezeQuote(&_MinerWallet.CallOpts)
}

// FrozenFunds is a free data retrieval call binding the contract method 0x90a74e2c.
//
// Solidity: function frozenFunds() constant returns(uint256)
func (_MinerWallet *MinerWalletCaller) FrozenFunds(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MinerWallet.contract.Call(opts, out, "frozenFunds")
	return *ret0, err
}

// FrozenFunds is a free data retrieval call binding the contract method 0x90a74e2c.
//
// Solidity: function frozenFunds() constant returns(uint256)
func (_MinerWallet *MinerWalletSession) FrozenFunds() (*big.Int, error) {
	return _MinerWallet.Contract.FrozenFunds(&_MinerWallet.CallOpts)
}

// FrozenFunds is a free data retrieval call binding the contract method 0x90a74e2c.
//
// Solidity: function frozenFunds() constant returns(uint256)
func (_MinerWallet *MinerWalletCallerSession) FrozenFunds() (*big.Int, error) {
	return _MinerWallet.Contract.FrozenFunds(&_MinerWallet.CallOpts)
}

// FrozenTime is a free data retrieval call binding the contract method 0x906db9ff.
//
// Solidity: function frozenTime() constant returns(uint64)
func (_MinerWallet *MinerWalletCaller) FrozenTime(opts *bind.CallOpts) (uint64, error) {
	var (
		ret0 = new(uint64)
	)
	out := ret0
	err := _MinerWallet.contract.Call(opts, out, "frozenTime")
	return *ret0, err
}

// FrozenTime is a free data retrieval call binding the contract method 0x906db9ff.
//
// Solidity: function frozenTime() constant returns(uint64)
func (_MinerWallet *MinerWalletSession) FrozenTime() (uint64, error) {
	return _MinerWallet.Contract.FrozenTime(&_MinerWallet.CallOpts)
}

// FrozenTime is a free data retrieval call binding the contract method 0x906db9ff.
//
// Solidity: function frozenTime() constant returns(uint64)
func (_MinerWallet *MinerWalletCallerSession) FrozenTime() (uint64, error) {
	return _MinerWallet.Contract.FrozenTime(&_MinerWallet.CallOpts)
}

// GenesisTime is a free data retrieval call binding the contract method 0x42c6498a.
//
// Solidity: function genesisTime() constant returns(uint64)
func (_MinerWallet *MinerWalletCaller) GenesisTime(opts *bind.CallOpts) (uint64, error) {
	var (
		ret0 = new(uint64)
	)
	out := ret0
	err := _MinerWallet.contract.Call(opts, out, "genesisTime")
	return *ret0, err
}

// GenesisTime is a free data retrieval call binding the contract method 0x42c6498a.
//
// Solidity: function genesisTime() constant returns(uint64)
func (_MinerWallet *MinerWalletSession) GenesisTime() (uint64, error) {
	return _MinerWallet.Contract.GenesisTime(&_MinerWallet.CallOpts)
}

// GenesisTime is a free data retrieval call binding the contract method 0x42c6498a.
//
// Solidity: function genesisTime() constant returns(uint64)
func (_MinerWallet *MinerWalletCallerSession) GenesisTime() (uint64, error) {
	return _MinerWallet.Contract.GenesisTime(&_MinerWallet.CallOpts)
}

// LockedFunds is a free data retrieval call binding the contract method 0xb8afaa48.
//
// Solidity: function lockedFunds() constant returns(uint256)
func (_MinerWallet *MinerWalletCaller) LockedFunds(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MinerWallet.contract.Call(opts, out, "lockedFunds")
	return *ret0, err
}

// LockedFunds is a free data retrieval call binding the contract method 0xb8afaa48.
//
// Solidity: function lockedFunds() constant returns(uint256)
func (_MinerWallet *MinerWalletSession) LockedFunds() (*big.Int, error) {
	return _MinerWallet.Contract.LockedFunds(&_MinerWallet.CallOpts)
}

// LockedFunds is a free data retrieval call binding the contract method 0xb8afaa48.
//
// Solidity: function lockedFunds() constant returns(uint256)
func (_MinerWallet *MinerWalletCallerSession) LockedFunds() (*big.Int, error) {
	return _MinerWallet.Contract.LockedFunds(&_MinerWallet.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_MinerWallet *MinerWalletCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _MinerWallet.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_MinerWallet *MinerWalletSession) Owner() (common.Address, error) {
	return _MinerWallet.Contract.Owner(&_MinerWallet.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_MinerWallet *MinerWalletCallerSession) Owner() (common.Address, error) {
	return _MinerWallet.Contract.Owner(&_MinerWallet.CallOpts)
}

// SharesTokenAddress is a free data retrieval call binding the contract method 0x27ebcf0e.
//
// Solidity: function sharesTokenAddress() constant returns(address)
func (_MinerWallet *MinerWalletCaller) SharesTokenAddress(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _MinerWallet.contract.Call(opts, out, "sharesTokenAddress")
	return *ret0, err
}

// SharesTokenAddress is a free data retrieval call binding the contract method 0x27ebcf0e.
//
// Solidity: function sharesTokenAddress() constant returns(address)
func (_MinerWallet *MinerWalletSession) SharesTokenAddress() (common.Address, error) {
	return _MinerWallet.Contract.SharesTokenAddress(&_MinerWallet.CallOpts)
}

// SharesTokenAddress is a free data retrieval call binding the contract method 0x27ebcf0e.
//
// Solidity: function sharesTokenAddress() constant returns(address)
func (_MinerWallet *MinerWalletCallerSession) SharesTokenAddress() (common.Address, error) {
	return _MinerWallet.Contract.SharesTokenAddress(&_MinerWallet.CallOpts)
}

// StakeShare is a free data retrieval call binding the contract method 0x1ea41c2c.
//
// Solidity: function stakeShare() constant returns(uint256)
func (_MinerWallet *MinerWalletCaller) StakeShare(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MinerWallet.contract.Call(opts, out, "stakeShare")
	return *ret0, err
}

// StakeShare is a free data retrieval call binding the contract method 0x1ea41c2c.
//
// Solidity: function stakeShare() constant returns(uint256)
func (_MinerWallet *MinerWalletSession) StakeShare() (*big.Int, error) {
	return _MinerWallet.Contract.StakeShare(&_MinerWallet.CallOpts)
}

// StakeShare is a free data retrieval call binding the contract method 0x1ea41c2c.
//
// Solidity: function stakeShare() constant returns(uint256)
func (_MinerWallet *MinerWalletCallerSession) StakeShare() (*big.Int, error) {
	return _MinerWallet.Contract.StakeShare(&_MinerWallet.CallOpts)
}

// PayDay is a paid mutator transaction binding the contract method 0x4d78511c.
//
// Solidity: function PayDay() returns()
func (_MinerWallet *MinerWalletTransactor) PayDay(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MinerWallet.contract.Transact(opts, "PayDay")
}

// PayDay is a paid mutator transaction binding the contract method 0x4d78511c.
//
// Solidity: function PayDay() returns()
func (_MinerWallet *MinerWalletSession) PayDay() (*types.Transaction, error) {
	return _MinerWallet.Contract.PayDay(&_MinerWallet.TransactOpts)
}

// PayDay is a paid mutator transaction binding the contract method 0x4d78511c.
//
// Solidity: function PayDay() returns()
func (_MinerWallet *MinerWalletTransactorSession) PayDay() (*types.Transaction, error) {
	return _MinerWallet.Contract.PayDay(&_MinerWallet.TransactOpts)
}

// Registration is a paid mutator transaction binding the contract method 0xdd1dcd9f.
//
// Solidity: function Registration(stake uint256) returns(success bool)
func (_MinerWallet *MinerWalletTransactor) Registration(opts *bind.TransactOpts, stake *big.Int) (*types.Transaction, error) {
	return _MinerWallet.contract.Transact(opts, "Registration", stake)
}

// Registration is a paid mutator transaction binding the contract method 0xdd1dcd9f.
//
// Solidity: function Registration(stake uint256) returns(success bool)
func (_MinerWallet *MinerWalletSession) Registration(stake *big.Int) (*types.Transaction, error) {
	return _MinerWallet.Contract.Registration(&_MinerWallet.TransactOpts, stake)
}

// Registration is a paid mutator transaction binding the contract method 0xdd1dcd9f.
//
// Solidity: function Registration(stake uint256) returns(success bool)
func (_MinerWallet *MinerWalletTransactorSession) Registration(stake *big.Int) (*types.Transaction, error) {
	return _MinerWallet.Contract.Registration(&_MinerWallet.TransactOpts, stake)
}

// Gulag is a paid mutator transaction binding the contract method 0x0b3eeac8.
//
// Solidity: function gulag() returns()
func (_MinerWallet *MinerWalletTransactor) Gulag(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MinerWallet.contract.Transact(opts, "gulag")
}

// Gulag is a paid mutator transaction binding the contract method 0x0b3eeac8.
//
// Solidity: function gulag() returns()
func (_MinerWallet *MinerWalletSession) Gulag() (*types.Transaction, error) {
	return _MinerWallet.Contract.Gulag(&_MinerWallet.TransactOpts)
}

// Gulag is a paid mutator transaction binding the contract method 0x0b3eeac8.
//
// Solidity: function gulag() returns()
func (_MinerWallet *MinerWalletTransactorSession) Gulag() (*types.Transaction, error) {
	return _MinerWallet.Contract.Gulag(&_MinerWallet.TransactOpts)
}

// PullMoney is a paid mutator transaction binding the contract method 0x5caf77d9.
//
// Solidity: function pullMoney(hubwallet address) returns()
func (_MinerWallet *MinerWalletTransactor) PullMoney(opts *bind.TransactOpts, hubwallet common.Address) (*types.Transaction, error) {
	return _MinerWallet.contract.Transact(opts, "pullMoney", hubwallet)
}

// PullMoney is a paid mutator transaction binding the contract method 0x5caf77d9.
//
// Solidity: function pullMoney(hubwallet address) returns()
func (_MinerWallet *MinerWalletSession) PullMoney(hubwallet common.Address) (*types.Transaction, error) {
	return _MinerWallet.Contract.PullMoney(&_MinerWallet.TransactOpts, hubwallet)
}

// PullMoney is a paid mutator transaction binding the contract method 0x5caf77d9.
//
// Solidity: function pullMoney(hubwallet address) returns()
func (_MinerWallet *MinerWalletTransactorSession) PullMoney(hubwallet common.Address) (*types.Transaction, error) {
	return _MinerWallet.Contract.PullMoney(&_MinerWallet.TransactOpts, hubwallet)
}

// Rehub is a paid mutator transaction binding the contract method 0xbd73820d.
//
// Solidity: function rehub() returns()
func (_MinerWallet *MinerWalletTransactor) Rehub(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MinerWallet.contract.Transact(opts, "rehub")
}

// Rehub is a paid mutator transaction binding the contract method 0xbd73820d.
//
// Solidity: function rehub() returns()
func (_MinerWallet *MinerWalletSession) Rehub() (*types.Transaction, error) {
	return _MinerWallet.Contract.Rehub(&_MinerWallet.TransactOpts)
}

// Rehub is a paid mutator transaction binding the contract method 0xbd73820d.
//
// Solidity: function rehub() returns()
func (_MinerWallet *MinerWalletTransactorSession) Rehub() (*types.Transaction, error) {
	return _MinerWallet.Contract.Rehub(&_MinerWallet.TransactOpts)
}

// Suspect is a paid mutator transaction binding the contract method 0x1e1683af.
//
// Solidity: function suspect() returns()
func (_MinerWallet *MinerWalletTransactor) Suspect(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MinerWallet.contract.Transact(opts, "suspect")
}

// Suspect is a paid mutator transaction binding the contract method 0x1e1683af.
//
// Solidity: function suspect() returns()
func (_MinerWallet *MinerWalletSession) Suspect() (*types.Transaction, error) {
	return _MinerWallet.Contract.Suspect(&_MinerWallet.TransactOpts)
}

// Suspect is a paid mutator transaction binding the contract method 0x1e1683af.
//
// Solidity: function suspect() returns()
func (_MinerWallet *MinerWalletTransactorSession) Suspect() (*types.Transaction, error) {
	return _MinerWallet.Contract.Suspect(&_MinerWallet.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_MinerWallet *MinerWalletTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _MinerWallet.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_MinerWallet *MinerWalletSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _MinerWallet.Contract.TransferOwnership(&_MinerWallet.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_MinerWallet *MinerWalletTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _MinerWallet.Contract.TransferOwnership(&_MinerWallet.TransactOpts, newOwner)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_MinerWallet *MinerWalletTransactor) Withdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MinerWallet.contract.Transact(opts, "withdraw")
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_MinerWallet *MinerWalletSession) Withdraw() (*types.Transaction, error) {
	return _MinerWallet.Contract.Withdraw(&_MinerWallet.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_MinerWallet *MinerWalletTransactorSession) Withdraw() (*types.Transaction, error) {
	return _MinerWallet.Contract.Withdraw(&_MinerWallet.TransactOpts)
}
