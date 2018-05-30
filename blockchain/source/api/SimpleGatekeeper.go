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

// SimpleGatekeeperABI is the input ABI used to generate the binding from.
const SimpleGatekeeperABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"paid\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"transactionAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"txNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"PayInTx\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"txNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"PayoutTx\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"block\",\"type\":\"uint256\"}],\"name\":\"Suicide\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"PayIn\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"},{\"name\":\"_txNumber\",\"type\":\"uint256\"}],\"name\":\"Payout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"kill\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// SimpleGatekeeperBin is the compiled bytecode used for deploying new contracts.
const SimpleGatekeeperBin = `0x6080604052600060025534801561001557600080fd5b506040516020806106b583398101604052516000805460018054600160a060020a03909416600160a060020a0319948516179055821633908117909216909117905561064f806100666000396000f3006080604052600436106100825763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166341c0e1b58114610087578063634235fc1461009e5780638da5cb5b146100c55780639fab56ac146100f6578063add89bb21461010e578063d942bffa1461013a578063f2fde38b14610161575b600080fd5b34801561009357600080fd5b5061009c610182565b005b3480156100aa57600080fd5b5061009c600160a060020a0360043516602435604435610319565b3480156100d157600080fd5b506100da61047f565b60408051600160a060020a039092168252519081900360200190f35b34801561010257600080fd5b5061009c60043561048e565b34801561011a57600080fd5b50610126600435610574565b604080519115158252519081900360200190f35b34801561014657600080fd5b5061014f610589565b60408051918252519081900360200190f35b34801561016d57600080fd5b5061009c600160a060020a036004351661058f565b60008054600160a060020a0316331461019a57600080fd5b600154604080517f70a082310000000000000000000000000000000000000000000000000000000081523060048201529051600160a060020a03909216916370a08231916024808201926020929091908290030181600087803b15801561020057600080fd5b505af1158015610214573d6000803e3d6000fd5b505050506040513d602081101561022a57600080fd5b505160015460008054604080517fa9059cbb000000000000000000000000000000000000000000000000000000008152600160a060020a03928316600482015260248101869052905194955092169263a9059cbb926044808201936020939283900390910190829087803b1580156102a157600080fd5b505af11580156102b5573d6000803e3d6000fd5b505050506040513d60208110156102cb57600080fd5b505115156102d857600080fd5b6040805142815290517fa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde9181900360200190a1600054600160a060020a0316ff5b60008054600160a060020a0316331461033157600080fd5b50604080516c01000000000000000000000000600160a060020a0386160281526014810183905260348101849052815190819003605401902060008181526003602052919091205460ff161561038657600080fd5b600154604080517fa9059cbb000000000000000000000000000000000000000000000000000000008152600160a060020a038781166004830152602482018790529151919092169163a9059cbb9160448083019260209291908290030181600087803b1580156103f557600080fd5b505af1158015610409573d6000803e3d6000fd5b505050506040513d602081101561041f57600080fd5b5051151561042c57600080fd5b600081815260036020526040808220805460ff191660011790555184918491600160a060020a038816917f731af16374848c2c73a6154fd410cb421138e7db45c5a904e5a475c756faa8d991a450505050565b600054600160a060020a031681565b600154604080517f23b872dd000000000000000000000000000000000000000000000000000000008152336004820152306024820152604481018490529051600160a060020a03909216916323b872dd916064808201926020929091908290030181600087803b15801561050157600080fd5b505af1158015610515573d6000803e3d6000fd5b505050506040513d602081101561052b57600080fd5b5051151561053857600080fd5b600280546001019081905560405182919033907f63768eabd21c026cb17439a3c6556436c1b0219c2046875297ad3f4b14e6700f90600090a450565b60036020526000908152604090205460ff1681565b60025481565b600054600160a060020a031633146105a657600080fd5b600160a060020a03811615156105bb57600080fd5b60008054604051600160a060020a03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a36000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a03929092169190911790555600a165627a7a723058208455575f9e51d88841569307171b552e8f84b5f482f06219a6cf960f384f47c80029`

// DeploySimpleGatekeeper deploys a new Ethereum contract, binding an instance of SimpleGatekeeper to it.
func DeploySimpleGatekeeper(auth *bind.TransactOpts, backend bind.ContractBackend, _token common.Address) (common.Address, *types.Transaction, *SimpleGatekeeper, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleGatekeeperABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SimpleGatekeeperBin), backend, _token)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SimpleGatekeeper{SimpleGatekeeperCaller: SimpleGatekeeperCaller{contract: contract}, SimpleGatekeeperTransactor: SimpleGatekeeperTransactor{contract: contract}, SimpleGatekeeperFilterer: SimpleGatekeeperFilterer{contract: contract}}, nil
}

// SimpleGatekeeper is an auto generated Go binding around an Ethereum contract.
type SimpleGatekeeper struct {
	SimpleGatekeeperCaller     // Read-only binding to the contract
	SimpleGatekeeperTransactor // Write-only binding to the contract
	SimpleGatekeeperFilterer   // Log filterer for contract events
}

// SimpleGatekeeperCaller is an auto generated read-only Go binding around an Ethereum contract.
type SimpleGatekeeperCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleGatekeeperTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SimpleGatekeeperTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleGatekeeperFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SimpleGatekeeperFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleGatekeeperSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SimpleGatekeeperSession struct {
	Contract     *SimpleGatekeeper // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SimpleGatekeeperCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SimpleGatekeeperCallerSession struct {
	Contract *SimpleGatekeeperCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// SimpleGatekeeperTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SimpleGatekeeperTransactorSession struct {
	Contract     *SimpleGatekeeperTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// SimpleGatekeeperRaw is an auto generated low-level Go binding around an Ethereum contract.
type SimpleGatekeeperRaw struct {
	Contract *SimpleGatekeeper // Generic contract binding to access the raw methods on
}

// SimpleGatekeeperCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SimpleGatekeeperCallerRaw struct {
	Contract *SimpleGatekeeperCaller // Generic read-only contract binding to access the raw methods on
}

// SimpleGatekeeperTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SimpleGatekeeperTransactorRaw struct {
	Contract *SimpleGatekeeperTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSimpleGatekeeper creates a new instance of SimpleGatekeeper, bound to a specific deployed contract.
func NewSimpleGatekeeper(address common.Address, backend bind.ContractBackend) (*SimpleGatekeeper, error) {
	contract, err := bindSimpleGatekeeper(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeper{SimpleGatekeeperCaller: SimpleGatekeeperCaller{contract: contract}, SimpleGatekeeperTransactor: SimpleGatekeeperTransactor{contract: contract}, SimpleGatekeeperFilterer: SimpleGatekeeperFilterer{contract: contract}}, nil
}

// NewSimpleGatekeeperCaller creates a new read-only instance of SimpleGatekeeper, bound to a specific deployed contract.
func NewSimpleGatekeeperCaller(address common.Address, caller bind.ContractCaller) (*SimpleGatekeeperCaller, error) {
	contract, err := bindSimpleGatekeeper(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperCaller{contract: contract}, nil
}

// NewSimpleGatekeeperTransactor creates a new write-only instance of SimpleGatekeeper, bound to a specific deployed contract.
func NewSimpleGatekeeperTransactor(address common.Address, transactor bind.ContractTransactor) (*SimpleGatekeeperTransactor, error) {
	contract, err := bindSimpleGatekeeper(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperTransactor{contract: contract}, nil
}

// NewSimpleGatekeeperFilterer creates a new log filterer instance of SimpleGatekeeper, bound to a specific deployed contract.
func NewSimpleGatekeeperFilterer(address common.Address, filterer bind.ContractFilterer) (*SimpleGatekeeperFilterer, error) {
	contract, err := bindSimpleGatekeeper(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperFilterer{contract: contract}, nil
}

// bindSimpleGatekeeper binds a generic wrapper to an already deployed contract.
func bindSimpleGatekeeper(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleGatekeeperABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleGatekeeper *SimpleGatekeeperRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleGatekeeper.Contract.SimpleGatekeeperCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleGatekeeper *SimpleGatekeeperRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleGatekeeper.Contract.SimpleGatekeeperTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleGatekeeper *SimpleGatekeeperRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleGatekeeper.Contract.SimpleGatekeeperTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleGatekeeper *SimpleGatekeeperCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleGatekeeper.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleGatekeeper *SimpleGatekeeperTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleGatekeeper.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleGatekeeper *SimpleGatekeeperTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleGatekeeper.Contract.contract.Transact(opts, method, params...)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SimpleGatekeeper *SimpleGatekeeperCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _SimpleGatekeeper.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SimpleGatekeeper *SimpleGatekeeperSession) Owner() (common.Address, error) {
	return _SimpleGatekeeper.Contract.Owner(&_SimpleGatekeeper.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SimpleGatekeeper *SimpleGatekeeperCallerSession) Owner() (common.Address, error) {
	return _SimpleGatekeeper.Contract.Owner(&_SimpleGatekeeper.CallOpts)
}

// Paid is a free data retrieval call binding the contract method 0xadd89bb2.
//
// Solidity: function paid( bytes32) constant returns(bool)
func (_SimpleGatekeeper *SimpleGatekeeperCaller) Paid(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _SimpleGatekeeper.contract.Call(opts, out, "paid", arg0)
	return *ret0, err
}

// Paid is a free data retrieval call binding the contract method 0xadd89bb2.
//
// Solidity: function paid( bytes32) constant returns(bool)
func (_SimpleGatekeeper *SimpleGatekeeperSession) Paid(arg0 [32]byte) (bool, error) {
	return _SimpleGatekeeper.Contract.Paid(&_SimpleGatekeeper.CallOpts, arg0)
}

// Paid is a free data retrieval call binding the contract method 0xadd89bb2.
//
// Solidity: function paid( bytes32) constant returns(bool)
func (_SimpleGatekeeper *SimpleGatekeeperCallerSession) Paid(arg0 [32]byte) (bool, error) {
	return _SimpleGatekeeper.Contract.Paid(&_SimpleGatekeeper.CallOpts, arg0)
}

// TransactionAmount is a free data retrieval call binding the contract method 0xd942bffa.
//
// Solidity: function transactionAmount() constant returns(uint256)
func (_SimpleGatekeeper *SimpleGatekeeperCaller) TransactionAmount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleGatekeeper.contract.Call(opts, out, "transactionAmount")
	return *ret0, err
}

// TransactionAmount is a free data retrieval call binding the contract method 0xd942bffa.
//
// Solidity: function transactionAmount() constant returns(uint256)
func (_SimpleGatekeeper *SimpleGatekeeperSession) TransactionAmount() (*big.Int, error) {
	return _SimpleGatekeeper.Contract.TransactionAmount(&_SimpleGatekeeper.CallOpts)
}

// TransactionAmount is a free data retrieval call binding the contract method 0xd942bffa.
//
// Solidity: function transactionAmount() constant returns(uint256)
func (_SimpleGatekeeper *SimpleGatekeeperCallerSession) TransactionAmount() (*big.Int, error) {
	return _SimpleGatekeeper.Contract.TransactionAmount(&_SimpleGatekeeper.CallOpts)
}

// PayIn is a paid mutator transaction binding the contract method 0x9fab56ac.
//
// Solidity: function PayIn(_value uint256) returns()
func (_SimpleGatekeeper *SimpleGatekeeperTransactor) PayIn(opts *bind.TransactOpts, _value *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeper.contract.Transact(opts, "PayIn", _value)
}

// PayIn is a paid mutator transaction binding the contract method 0x9fab56ac.
//
// Solidity: function PayIn(_value uint256) returns()
func (_SimpleGatekeeper *SimpleGatekeeperSession) PayIn(_value *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeper.Contract.PayIn(&_SimpleGatekeeper.TransactOpts, _value)
}

// PayIn is a paid mutator transaction binding the contract method 0x9fab56ac.
//
// Solidity: function PayIn(_value uint256) returns()
func (_SimpleGatekeeper *SimpleGatekeeperTransactorSession) PayIn(_value *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeper.Contract.PayIn(&_SimpleGatekeeper.TransactOpts, _value)
}

// Payout is a paid mutator transaction binding the contract method 0x634235fc.
//
// Solidity: function Payout(_to address, _value uint256, _txNumber uint256) returns()
func (_SimpleGatekeeper *SimpleGatekeeperTransactor) Payout(opts *bind.TransactOpts, _to common.Address, _value *big.Int, _txNumber *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeper.contract.Transact(opts, "Payout", _to, _value, _txNumber)
}

// Payout is a paid mutator transaction binding the contract method 0x634235fc.
//
// Solidity: function Payout(_to address, _value uint256, _txNumber uint256) returns()
func (_SimpleGatekeeper *SimpleGatekeeperSession) Payout(_to common.Address, _value *big.Int, _txNumber *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeper.Contract.Payout(&_SimpleGatekeeper.TransactOpts, _to, _value, _txNumber)
}

// Payout is a paid mutator transaction binding the contract method 0x634235fc.
//
// Solidity: function Payout(_to address, _value uint256, _txNumber uint256) returns()
func (_SimpleGatekeeper *SimpleGatekeeperTransactorSession) Payout(_to common.Address, _value *big.Int, _txNumber *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeper.Contract.Payout(&_SimpleGatekeeper.TransactOpts, _to, _value, _txNumber)
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_SimpleGatekeeper *SimpleGatekeeperTransactor) Kill(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleGatekeeper.contract.Transact(opts, "kill")
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_SimpleGatekeeper *SimpleGatekeeperSession) Kill() (*types.Transaction, error) {
	return _SimpleGatekeeper.Contract.Kill(&_SimpleGatekeeper.TransactOpts)
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_SimpleGatekeeper *SimpleGatekeeperTransactorSession) Kill() (*types.Transaction, error) {
	return _SimpleGatekeeper.Contract.Kill(&_SimpleGatekeeper.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_SimpleGatekeeper *SimpleGatekeeperTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeper.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_SimpleGatekeeper *SimpleGatekeeperSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeper.Contract.TransferOwnership(&_SimpleGatekeeper.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_SimpleGatekeeper *SimpleGatekeeperTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeper.Contract.TransferOwnership(&_SimpleGatekeeper.TransactOpts, newOwner)
}

// SimpleGatekeeperOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SimpleGatekeeper contract.
type SimpleGatekeeperOwnershipTransferredIterator struct {
	Event *SimpleGatekeeperOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperOwnershipTransferred)
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
		it.Event = new(SimpleGatekeeperOwnershipTransferred)
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
func (it *SimpleGatekeeperOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperOwnershipTransferred represents a OwnershipTransferred event raised by the SimpleGatekeeper contract.
type SimpleGatekeeperOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_SimpleGatekeeper *SimpleGatekeeperFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SimpleGatekeeperOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SimpleGatekeeper.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperOwnershipTransferredIterator{contract: _SimpleGatekeeper.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_SimpleGatekeeper *SimpleGatekeeperFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SimpleGatekeeper.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperOwnershipTransferred)
				if err := _SimpleGatekeeper.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// SimpleGatekeeperPayInTxIterator is returned from FilterPayInTx and is used to iterate over the raw logs and unpacked data for PayInTx events raised by the SimpleGatekeeper contract.
type SimpleGatekeeperPayInTxIterator struct {
	Event *SimpleGatekeeperPayInTx // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperPayInTxIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperPayInTx)
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
		it.Event = new(SimpleGatekeeperPayInTx)
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
func (it *SimpleGatekeeperPayInTxIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperPayInTxIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperPayInTx represents a PayInTx event raised by the SimpleGatekeeper contract.
type SimpleGatekeeperPayInTx struct {
	From     common.Address
	TxNumber *big.Int
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPayInTx is a free log retrieval operation binding the contract event 0x63768eabd21c026cb17439a3c6556436c1b0219c2046875297ad3f4b14e6700f.
//
// Solidity: event PayInTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_SimpleGatekeeper *SimpleGatekeeperFilterer) FilterPayInTx(opts *bind.FilterOpts, from []common.Address, txNumber []*big.Int, value []*big.Int) (*SimpleGatekeeperPayInTxIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var txNumberRule []interface{}
	for _, txNumberItem := range txNumber {
		txNumberRule = append(txNumberRule, txNumberItem)
	}
	var valueRule []interface{}
	for _, valueItem := range value {
		valueRule = append(valueRule, valueItem)
	}

	logs, sub, err := _SimpleGatekeeper.contract.FilterLogs(opts, "PayInTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperPayInTxIterator{contract: _SimpleGatekeeper.contract, event: "PayInTx", logs: logs, sub: sub}, nil
}

// WatchPayInTx is a free log subscription operation binding the contract event 0x63768eabd21c026cb17439a3c6556436c1b0219c2046875297ad3f4b14e6700f.
//
// Solidity: event PayInTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_SimpleGatekeeper *SimpleGatekeeperFilterer) WatchPayInTx(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperPayInTx, from []common.Address, txNumber []*big.Int, value []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var txNumberRule []interface{}
	for _, txNumberItem := range txNumber {
		txNumberRule = append(txNumberRule, txNumberItem)
	}
	var valueRule []interface{}
	for _, valueItem := range value {
		valueRule = append(valueRule, valueItem)
	}

	logs, sub, err := _SimpleGatekeeper.contract.WatchLogs(opts, "PayInTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperPayInTx)
				if err := _SimpleGatekeeper.contract.UnpackLog(event, "PayInTx", log); err != nil {
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

// SimpleGatekeeperPayoutTxIterator is returned from FilterPayoutTx and is used to iterate over the raw logs and unpacked data for PayoutTx events raised by the SimpleGatekeeper contract.
type SimpleGatekeeperPayoutTxIterator struct {
	Event *SimpleGatekeeperPayoutTx // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperPayoutTxIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperPayoutTx)
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
		it.Event = new(SimpleGatekeeperPayoutTx)
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
func (it *SimpleGatekeeperPayoutTxIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperPayoutTxIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperPayoutTx represents a PayoutTx event raised by the SimpleGatekeeper contract.
type SimpleGatekeeperPayoutTx struct {
	From     common.Address
	TxNumber *big.Int
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPayoutTx is a free log retrieval operation binding the contract event 0x731af16374848c2c73a6154fd410cb421138e7db45c5a904e5a475c756faa8d9.
//
// Solidity: event PayoutTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_SimpleGatekeeper *SimpleGatekeeperFilterer) FilterPayoutTx(opts *bind.FilterOpts, from []common.Address, txNumber []*big.Int, value []*big.Int) (*SimpleGatekeeperPayoutTxIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var txNumberRule []interface{}
	for _, txNumberItem := range txNumber {
		txNumberRule = append(txNumberRule, txNumberItem)
	}
	var valueRule []interface{}
	for _, valueItem := range value {
		valueRule = append(valueRule, valueItem)
	}

	logs, sub, err := _SimpleGatekeeper.contract.FilterLogs(opts, "PayoutTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperPayoutTxIterator{contract: _SimpleGatekeeper.contract, event: "PayoutTx", logs: logs, sub: sub}, nil
}

// WatchPayoutTx is a free log subscription operation binding the contract event 0x731af16374848c2c73a6154fd410cb421138e7db45c5a904e5a475c756faa8d9.
//
// Solidity: event PayoutTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_SimpleGatekeeper *SimpleGatekeeperFilterer) WatchPayoutTx(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperPayoutTx, from []common.Address, txNumber []*big.Int, value []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var txNumberRule []interface{}
	for _, txNumberItem := range txNumber {
		txNumberRule = append(txNumberRule, txNumberItem)
	}
	var valueRule []interface{}
	for _, valueItem := range value {
		valueRule = append(valueRule, valueItem)
	}

	logs, sub, err := _SimpleGatekeeper.contract.WatchLogs(opts, "PayoutTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperPayoutTx)
				if err := _SimpleGatekeeper.contract.UnpackLog(event, "PayoutTx", log); err != nil {
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

// SimpleGatekeeperSuicideIterator is returned from FilterSuicide and is used to iterate over the raw logs and unpacked data for Suicide events raised by the SimpleGatekeeper contract.
type SimpleGatekeeperSuicideIterator struct {
	Event *SimpleGatekeeperSuicide // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperSuicideIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperSuicide)
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
		it.Event = new(SimpleGatekeeperSuicide)
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
func (it *SimpleGatekeeperSuicideIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperSuicideIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperSuicide represents a Suicide event raised by the SimpleGatekeeper contract.
type SimpleGatekeeperSuicide struct {
	Block *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterSuicide is a free log retrieval operation binding the contract event 0xa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde.
//
// Solidity: event Suicide(block uint256)
func (_SimpleGatekeeper *SimpleGatekeeperFilterer) FilterSuicide(opts *bind.FilterOpts) (*SimpleGatekeeperSuicideIterator, error) {

	logs, sub, err := _SimpleGatekeeper.contract.FilterLogs(opts, "Suicide")
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperSuicideIterator{contract: _SimpleGatekeeper.contract, event: "Suicide", logs: logs, sub: sub}, nil
}

// WatchSuicide is a free log subscription operation binding the contract event 0xa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde.
//
// Solidity: event Suicide(block uint256)
func (_SimpleGatekeeper *SimpleGatekeeperFilterer) WatchSuicide(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperSuicide) (event.Subscription, error) {

	logs, sub, err := _SimpleGatekeeper.contract.WatchLogs(opts, "Suicide")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperSuicide)
				if err := _SimpleGatekeeper.contract.UnpackLog(event, "Suicide", log); err != nil {
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
