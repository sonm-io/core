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
	ethereum "github.com/sonm-io/go-ethereum"
	"github.com/sonm-io/go-ethereum/event"
)

// AutoPayoutABI is the input ABI used to generate the binding from.
const AutoPayoutABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"allowedPayouts\",\"outputs\":[{\"name\":\"lowLimit\",\"type\":\"uint256\"},{\"name\":\"target\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"},{\"name\":\"_gatekeeper\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"master\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"target\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"limit\",\"type\":\"uint256\"}],\"name\":\"AutoPayoutChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"master\",\"type\":\"address\"}],\"name\":\"AutoPayout\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"block\",\"type\":\"uint256\"}],\"name\":\"Suicide\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"}],\"name\":\"OwnershipRenounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_limit\",\"type\":\"uint256\"},{\"name\":\"_target\",\"type\":\"address\"}],\"name\":\"SetAutoPayout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_master\",\"type\":\"address\"}],\"name\":\"DoAutoPayout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"kill\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// AutoPayoutBin is the compiled bytecode used for deploying new contracts.
const AutoPayoutBin = `0x608060405234801561001057600080fd5b5060405160408061083b8339810160405280516020909101516000805460018054600160a060020a0319908116600160a060020a039687161790915560028054821695909416949094179092553391831682179092161781556107c290819061007990396000f3006080604052600436106100825763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166312fbd33c811461008757806341c0e1b5146100c9578063715018a6146100e05780638a12ae40146100f55780638da5cb5b14610119578063b5dfc1d51461014a578063f2fde38b1461016b575b600080fd5b34801561009357600080fd5b506100a8600160a060020a036004351661018c565b60408051928352600160a060020a0390911660208301528051918290030190f35b3480156100d557600080fd5b506100de6101ae565b005b3480156100ec57600080fd5b506100de610343565b34801561010157600080fd5b506100de600435600160a060020a03602435166103af565b34801561012557600080fd5b5061012e61041b565b60408051600160a060020a039092168252519081900360200190f35b34801561015657600080fd5b506100de600160a060020a036004351661042a565b34801561017757600080fd5b506100de600160a060020a03600435166106f6565b60036020526000908152604090208054600190910154600160a060020a031682565b600054600160a060020a031633146101c557600080fd5b60015460008054604080517f70a082310000000000000000000000000000000000000000000000000000000081523060048201529051600160a060020a039485169463a9059cbb9493169285926370a082319260248083019360209383900390910190829087803b15801561023957600080fd5b505af115801561024d573d6000803e3d6000fd5b505050506040513d602081101561026357600080fd5b5051604080517c010000000000000000000000000000000000000000000000000000000063ffffffff8616028152600160a060020a03909316600484015260248301919091525160448083019260209291908290030181600087803b1580156102cb57600080fd5b505af11580156102df573d6000803e3d6000fd5b505050506040513d60208110156102f557600080fd5b5051151561030257600080fd5b6040805142815290517fa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde9181900360200190a1600054600160a060020a0316ff5b600054600160a060020a0316331461035a57600080fd5b60008054604051600160a060020a03909116917ff8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c6482091a26000805473ffffffffffffffffffffffffffffffffffffffff19169055565b33600081815260036020526040808220858155600101805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0386169081179091559051859391927f1ee471395ea2cfa5c7eec94aabde2b3c330825cb942ef8d0fd1eeb7a0d3d275291a45050565b600054600160a060020a031681565b600154604080517f70a08231000000000000000000000000000000000000000000000000000000008152600160a060020a038481166004830152915160009392909216916370a082319160248082019260209290919082900301818787803b15801561049557600080fd5b505af11580156104a9573d6000803e3d6000fd5b505050506040513d60208110156104bf57600080fd5b5051600160a060020a0383166000908152600360205260409020549091508110156104e957600080fd5b600154604080517f23b872dd000000000000000000000000000000000000000000000000000000008152600160a060020a03858116600483015230602483015260448201859052915191909216916323b872dd9160648083019260209291908290030181600087803b15801561055e57600080fd5b505af1158015610572573d6000803e3d6000fd5b505050506040513d602081101561058857600080fd5b5050600154600254604080517f095ea7b3000000000000000000000000000000000000000000000000000000008152600160a060020a039283166004820152602481018590529051919092169163095ea7b39160448083019260209291908290030181600087803b1580156105fc57600080fd5b505af1158015610610573d6000803e3d6000fd5b505050506040513d602081101561062657600080fd5b5050600254600160a060020a038381166000908152600360205260408082206001015481517fe3fcd18e0000000000000000000000000000000000000000000000000000000081526004810187905290841660248201529051929093169263e3fcd18e926044808301939282900301818387803b1580156106a657600080fd5b505af11580156106ba573d6000803e3d6000fd5b5050604051600160a060020a03851692507f5a9b1e90057b7163b237d41fbf5ba76b7eaf01f482fe75255aa290ced89e91b29150600090a25050565b600054600160a060020a0316331461070d57600080fd5b61071681610719565b50565b600160a060020a038116151561072e57600080fd5b60008054604051600160a060020a03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a36000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a03929092169190911790555600a165627a7a723058206f97111d27c4489baef8fe0d155f4b1dd6fb9a9648c941e11aef66dcbe03a1700029`

// DeployAutoPayout deploys a new Ethereum contract, binding an instance of AutoPayout to it.
func DeployAutoPayout(auth *bind.TransactOpts, backend bind.ContractBackend, _token common.Address, _gatekeeper common.Address) (common.Address, *types.Transaction, *AutoPayout, error) {
	parsed, err := abi.JSON(strings.NewReader(AutoPayoutABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(AutoPayoutBin), backend, _token, _gatekeeper)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &AutoPayout{AutoPayoutCaller: AutoPayoutCaller{contract: contract}, AutoPayoutTransactor: AutoPayoutTransactor{contract: contract}, AutoPayoutFilterer: AutoPayoutFilterer{contract: contract}}, nil
}

// AutoPayout is an auto generated Go binding around an Ethereum contract.
type AutoPayout struct {
	AutoPayoutCaller     // Read-only binding to the contract
	AutoPayoutTransactor // Write-only binding to the contract
	AutoPayoutFilterer   // Log filterer for contract events
}

// AutoPayoutCaller is an auto generated read-only Go binding around an Ethereum contract.
type AutoPayoutCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AutoPayoutTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AutoPayoutTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AutoPayoutFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AutoPayoutFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AutoPayoutSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AutoPayoutSession struct {
	Contract     *AutoPayout       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AutoPayoutCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AutoPayoutCallerSession struct {
	Contract *AutoPayoutCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// AutoPayoutTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AutoPayoutTransactorSession struct {
	Contract     *AutoPayoutTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// AutoPayoutRaw is an auto generated low-level Go binding around an Ethereum contract.
type AutoPayoutRaw struct {
	Contract *AutoPayout // Generic contract binding to access the raw methods on
}

// AutoPayoutCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AutoPayoutCallerRaw struct {
	Contract *AutoPayoutCaller // Generic read-only contract binding to access the raw methods on
}

// AutoPayoutTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AutoPayoutTransactorRaw struct {
	Contract *AutoPayoutTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAutoPayout creates a new instance of AutoPayout, bound to a specific deployed contract.
func NewAutoPayout(address common.Address, backend bind.ContractBackend) (*AutoPayout, error) {
	contract, err := bindAutoPayout(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AutoPayout{AutoPayoutCaller: AutoPayoutCaller{contract: contract}, AutoPayoutTransactor: AutoPayoutTransactor{contract: contract}, AutoPayoutFilterer: AutoPayoutFilterer{contract: contract}}, nil
}

// NewAutoPayoutCaller creates a new read-only instance of AutoPayout, bound to a specific deployed contract.
func NewAutoPayoutCaller(address common.Address, caller bind.ContractCaller) (*AutoPayoutCaller, error) {
	contract, err := bindAutoPayout(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AutoPayoutCaller{contract: contract}, nil
}

// NewAutoPayoutTransactor creates a new write-only instance of AutoPayout, bound to a specific deployed contract.
func NewAutoPayoutTransactor(address common.Address, transactor bind.ContractTransactor) (*AutoPayoutTransactor, error) {
	contract, err := bindAutoPayout(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AutoPayoutTransactor{contract: contract}, nil
}

// NewAutoPayoutFilterer creates a new log filterer instance of AutoPayout, bound to a specific deployed contract.
func NewAutoPayoutFilterer(address common.Address, filterer bind.ContractFilterer) (*AutoPayoutFilterer, error) {
	contract, err := bindAutoPayout(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AutoPayoutFilterer{contract: contract}, nil
}

// bindAutoPayout binds a generic wrapper to an already deployed contract.
func bindAutoPayout(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AutoPayoutABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AutoPayout *AutoPayoutRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _AutoPayout.Contract.AutoPayoutCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AutoPayout *AutoPayoutRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AutoPayout.Contract.AutoPayoutTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AutoPayout *AutoPayoutRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AutoPayout.Contract.AutoPayoutTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AutoPayout *AutoPayoutCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _AutoPayout.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AutoPayout *AutoPayoutTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AutoPayout.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AutoPayout *AutoPayoutTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AutoPayout.Contract.contract.Transact(opts, method, params...)
}

// AllowedPayouts is a free data retrieval call binding the contract method 0x12fbd33c.
//
// Solidity: function allowedPayouts( address) constant returns(lowLimit uint256, target address)
func (_AutoPayout *AutoPayoutCaller) AllowedPayouts(opts *bind.CallOpts, arg0 common.Address) (struct {
	LowLimit *big.Int
	Target   common.Address
}, error) {
	ret := new(struct {
		LowLimit *big.Int
		Target   common.Address
	})
	out := ret
	err := _AutoPayout.contract.Call(opts, out, "allowedPayouts", arg0)
	return *ret, err
}

// AllowedPayouts is a free data retrieval call binding the contract method 0x12fbd33c.
//
// Solidity: function allowedPayouts( address) constant returns(lowLimit uint256, target address)
func (_AutoPayout *AutoPayoutSession) AllowedPayouts(arg0 common.Address) (struct {
	LowLimit *big.Int
	Target   common.Address
}, error) {
	return _AutoPayout.Contract.AllowedPayouts(&_AutoPayout.CallOpts, arg0)
}

// AllowedPayouts is a free data retrieval call binding the contract method 0x12fbd33c.
//
// Solidity: function allowedPayouts( address) constant returns(lowLimit uint256, target address)
func (_AutoPayout *AutoPayoutCallerSession) AllowedPayouts(arg0 common.Address) (struct {
	LowLimit *big.Int
	Target   common.Address
}, error) {
	return _AutoPayout.Contract.AllowedPayouts(&_AutoPayout.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_AutoPayout *AutoPayoutCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _AutoPayout.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_AutoPayout *AutoPayoutSession) Owner() (common.Address, error) {
	return _AutoPayout.Contract.Owner(&_AutoPayout.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_AutoPayout *AutoPayoutCallerSession) Owner() (common.Address, error) {
	return _AutoPayout.Contract.Owner(&_AutoPayout.CallOpts)
}

// DoAutoPayout is a paid mutator transaction binding the contract method 0xb5dfc1d5.
//
// Solidity: function DoAutoPayout(_master address) returns()
func (_AutoPayout *AutoPayoutTransactor) DoAutoPayout(opts *bind.TransactOpts, _master common.Address) (*types.Transaction, error) {
	return _AutoPayout.contract.Transact(opts, "DoAutoPayout", _master)
}

// DoAutoPayout is a paid mutator transaction binding the contract method 0xb5dfc1d5.
//
// Solidity: function DoAutoPayout(_master address) returns()
func (_AutoPayout *AutoPayoutSession) DoAutoPayout(_master common.Address) (*types.Transaction, error) {
	return _AutoPayout.Contract.DoAutoPayout(&_AutoPayout.TransactOpts, _master)
}

// DoAutoPayout is a paid mutator transaction binding the contract method 0xb5dfc1d5.
//
// Solidity: function DoAutoPayout(_master address) returns()
func (_AutoPayout *AutoPayoutTransactorSession) DoAutoPayout(_master common.Address) (*types.Transaction, error) {
	return _AutoPayout.Contract.DoAutoPayout(&_AutoPayout.TransactOpts, _master)
}

// SetAutoPayout is a paid mutator transaction binding the contract method 0x8a12ae40.
//
// Solidity: function SetAutoPayout(_limit uint256, _target address) returns()
func (_AutoPayout *AutoPayoutTransactor) SetAutoPayout(opts *bind.TransactOpts, _limit *big.Int, _target common.Address) (*types.Transaction, error) {
	return _AutoPayout.contract.Transact(opts, "SetAutoPayout", _limit, _target)
}

// SetAutoPayout is a paid mutator transaction binding the contract method 0x8a12ae40.
//
// Solidity: function SetAutoPayout(_limit uint256, _target address) returns()
func (_AutoPayout *AutoPayoutSession) SetAutoPayout(_limit *big.Int, _target common.Address) (*types.Transaction, error) {
	return _AutoPayout.Contract.SetAutoPayout(&_AutoPayout.TransactOpts, _limit, _target)
}

// SetAutoPayout is a paid mutator transaction binding the contract method 0x8a12ae40.
//
// Solidity: function SetAutoPayout(_limit uint256, _target address) returns()
func (_AutoPayout *AutoPayoutTransactorSession) SetAutoPayout(_limit *big.Int, _target common.Address) (*types.Transaction, error) {
	return _AutoPayout.Contract.SetAutoPayout(&_AutoPayout.TransactOpts, _limit, _target)
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_AutoPayout *AutoPayoutTransactor) Kill(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AutoPayout.contract.Transact(opts, "kill")
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_AutoPayout *AutoPayoutSession) Kill() (*types.Transaction, error) {
	return _AutoPayout.Contract.Kill(&_AutoPayout.TransactOpts)
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_AutoPayout *AutoPayoutTransactorSession) Kill() (*types.Transaction, error) {
	return _AutoPayout.Contract.Kill(&_AutoPayout.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AutoPayout *AutoPayoutTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AutoPayout.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AutoPayout *AutoPayoutSession) RenounceOwnership() (*types.Transaction, error) {
	return _AutoPayout.Contract.RenounceOwnership(&_AutoPayout.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AutoPayout *AutoPayoutTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _AutoPayout.Contract.RenounceOwnership(&_AutoPayout.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_AutoPayout *AutoPayoutTransactor) TransferOwnership(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _AutoPayout.contract.Transact(opts, "transferOwnership", _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_AutoPayout *AutoPayoutSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _AutoPayout.Contract.TransferOwnership(&_AutoPayout.TransactOpts, _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_AutoPayout *AutoPayoutTransactorSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _AutoPayout.Contract.TransferOwnership(&_AutoPayout.TransactOpts, _newOwner)
}

// AutoPayoutAutoPayoutIterator is returned from FilterAutoPayout and is used to iterate over the raw logs and unpacked data for AutoPayout events raised by the AutoPayout contract.
type AutoPayoutAutoPayoutIterator struct {
	Event *AutoPayoutAutoPayout // Event containing the contract specifics and raw log

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
func (it *AutoPayoutAutoPayoutIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AutoPayoutAutoPayout)
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
		it.Event = new(AutoPayoutAutoPayout)
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
func (it *AutoPayoutAutoPayoutIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AutoPayoutAutoPayoutIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AutoPayoutAutoPayout represents a AutoPayout event raised by the AutoPayout contract.
type AutoPayoutAutoPayout struct {
	Master common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterAutoPayout is a free log retrieval operation binding the contract event 0x5a9b1e90057b7163b237d41fbf5ba76b7eaf01f482fe75255aa290ced89e91b2.
//
// Solidity: e AutoPayout(master indexed address)
func (_AutoPayout *AutoPayoutFilterer) FilterAutoPayout(opts *bind.FilterOpts, master []common.Address) (*AutoPayoutAutoPayoutIterator, error) {

	var masterRule []interface{}
	for _, masterItem := range master {
		masterRule = append(masterRule, masterItem)
	}

	logs, sub, err := _AutoPayout.contract.FilterLogs(opts, "AutoPayout", masterRule)
	if err != nil {
		return nil, err
	}
	return &AutoPayoutAutoPayoutIterator{contract: _AutoPayout.contract, event: "AutoPayout", logs: logs, sub: sub}, nil
}

// WatchAutoPayout is a free log subscription operation binding the contract event 0x5a9b1e90057b7163b237d41fbf5ba76b7eaf01f482fe75255aa290ced89e91b2.
//
// Solidity: e AutoPayout(master indexed address)
func (_AutoPayout *AutoPayoutFilterer) WatchAutoPayout(opts *bind.WatchOpts, sink chan<- *AutoPayoutAutoPayout, master []common.Address) (event.Subscription, error) {

	var masterRule []interface{}
	for _, masterItem := range master {
		masterRule = append(masterRule, masterItem)
	}

	logs, sub, err := _AutoPayout.contract.WatchLogs(opts, "AutoPayout", masterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AutoPayoutAutoPayout)
				if err := _AutoPayout.contract.UnpackLog(event, "AutoPayout", log); err != nil {
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

// AutoPayoutAutoPayoutChangedIterator is returned from FilterAutoPayoutChanged and is used to iterate over the raw logs and unpacked data for AutoPayoutChanged events raised by the AutoPayout contract.
type AutoPayoutAutoPayoutChangedIterator struct {
	Event *AutoPayoutAutoPayoutChanged // Event containing the contract specifics and raw log

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
func (it *AutoPayoutAutoPayoutChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AutoPayoutAutoPayoutChanged)
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
		it.Event = new(AutoPayoutAutoPayoutChanged)
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
func (it *AutoPayoutAutoPayoutChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AutoPayoutAutoPayoutChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AutoPayoutAutoPayoutChanged represents a AutoPayoutChanged event raised by the AutoPayout contract.
type AutoPayoutAutoPayoutChanged struct {
	Master common.Address
	Target common.Address
	Limit  *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterAutoPayoutChanged is a free log retrieval operation binding the contract event 0x1ee471395ea2cfa5c7eec94aabde2b3c330825cb942ef8d0fd1eeb7a0d3d2752.
//
// Solidity: e AutoPayoutChanged(master indexed address, target indexed address, limit indexed uint256)
func (_AutoPayout *AutoPayoutFilterer) FilterAutoPayoutChanged(opts *bind.FilterOpts, master []common.Address, target []common.Address, limit []*big.Int) (*AutoPayoutAutoPayoutChangedIterator, error) {

	var masterRule []interface{}
	for _, masterItem := range master {
		masterRule = append(masterRule, masterItem)
	}
	var targetRule []interface{}
	for _, targetItem := range target {
		targetRule = append(targetRule, targetItem)
	}
	var limitRule []interface{}
	for _, limitItem := range limit {
		limitRule = append(limitRule, limitItem)
	}

	logs, sub, err := _AutoPayout.contract.FilterLogs(opts, "AutoPayoutChanged", masterRule, targetRule, limitRule)
	if err != nil {
		return nil, err
	}
	return &AutoPayoutAutoPayoutChangedIterator{contract: _AutoPayout.contract, event: "AutoPayoutChanged", logs: logs, sub: sub}, nil
}

// WatchAutoPayoutChanged is a free log subscription operation binding the contract event 0x1ee471395ea2cfa5c7eec94aabde2b3c330825cb942ef8d0fd1eeb7a0d3d2752.
//
// Solidity: e AutoPayoutChanged(master indexed address, target indexed address, limit indexed uint256)
func (_AutoPayout *AutoPayoutFilterer) WatchAutoPayoutChanged(opts *bind.WatchOpts, sink chan<- *AutoPayoutAutoPayoutChanged, master []common.Address, target []common.Address, limit []*big.Int) (event.Subscription, error) {

	var masterRule []interface{}
	for _, masterItem := range master {
		masterRule = append(masterRule, masterItem)
	}
	var targetRule []interface{}
	for _, targetItem := range target {
		targetRule = append(targetRule, targetItem)
	}
	var limitRule []interface{}
	for _, limitItem := range limit {
		limitRule = append(limitRule, limitItem)
	}

	logs, sub, err := _AutoPayout.contract.WatchLogs(opts, "AutoPayoutChanged", masterRule, targetRule, limitRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AutoPayoutAutoPayoutChanged)
				if err := _AutoPayout.contract.UnpackLog(event, "AutoPayoutChanged", log); err != nil {
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

// AutoPayoutOwnershipRenouncedIterator is returned from FilterOwnershipRenounced and is used to iterate over the raw logs and unpacked data for OwnershipRenounced events raised by the AutoPayout contract.
type AutoPayoutOwnershipRenouncedIterator struct {
	Event *AutoPayoutOwnershipRenounced // Event containing the contract specifics and raw log

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
func (it *AutoPayoutOwnershipRenouncedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AutoPayoutOwnershipRenounced)
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
		it.Event = new(AutoPayoutOwnershipRenounced)
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
func (it *AutoPayoutOwnershipRenouncedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AutoPayoutOwnershipRenouncedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AutoPayoutOwnershipRenounced represents a OwnershipRenounced event raised by the AutoPayout contract.
type AutoPayoutOwnershipRenounced struct {
	PreviousOwner common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipRenounced is a free log retrieval operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_AutoPayout *AutoPayoutFilterer) FilterOwnershipRenounced(opts *bind.FilterOpts, previousOwner []common.Address) (*AutoPayoutOwnershipRenouncedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _AutoPayout.contract.FilterLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return &AutoPayoutOwnershipRenouncedIterator{contract: _AutoPayout.contract, event: "OwnershipRenounced", logs: logs, sub: sub}, nil
}

// WatchOwnershipRenounced is a free log subscription operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_AutoPayout *AutoPayoutFilterer) WatchOwnershipRenounced(opts *bind.WatchOpts, sink chan<- *AutoPayoutOwnershipRenounced, previousOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _AutoPayout.contract.WatchLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AutoPayoutOwnershipRenounced)
				if err := _AutoPayout.contract.UnpackLog(event, "OwnershipRenounced", log); err != nil {
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

// AutoPayoutOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the AutoPayout contract.
type AutoPayoutOwnershipTransferredIterator struct {
	Event *AutoPayoutOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *AutoPayoutOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AutoPayoutOwnershipTransferred)
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
		it.Event = new(AutoPayoutOwnershipTransferred)
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
func (it *AutoPayoutOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AutoPayoutOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AutoPayoutOwnershipTransferred represents a OwnershipTransferred event raised by the AutoPayout contract.
type AutoPayoutOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_AutoPayout *AutoPayoutFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*AutoPayoutOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AutoPayout.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &AutoPayoutOwnershipTransferredIterator{contract: _AutoPayout.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_AutoPayout *AutoPayoutFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *AutoPayoutOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AutoPayout.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AutoPayoutOwnershipTransferred)
				if err := _AutoPayout.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// AutoPayoutSuicideIterator is returned from FilterSuicide and is used to iterate over the raw logs and unpacked data for Suicide events raised by the AutoPayout contract.
type AutoPayoutSuicideIterator struct {
	Event *AutoPayoutSuicide // Event containing the contract specifics and raw log

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
func (it *AutoPayoutSuicideIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AutoPayoutSuicide)
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
		it.Event = new(AutoPayoutSuicide)
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
func (it *AutoPayoutSuicideIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AutoPayoutSuicideIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AutoPayoutSuicide represents a Suicide event raised by the AutoPayout contract.
type AutoPayoutSuicide struct {
	Block *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterSuicide is a free log retrieval operation binding the contract event 0xa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde.
//
// Solidity: e Suicide(block uint256)
func (_AutoPayout *AutoPayoutFilterer) FilterSuicide(opts *bind.FilterOpts) (*AutoPayoutSuicideIterator, error) {

	logs, sub, err := _AutoPayout.contract.FilterLogs(opts, "Suicide")
	if err != nil {
		return nil, err
	}
	return &AutoPayoutSuicideIterator{contract: _AutoPayout.contract, event: "Suicide", logs: logs, sub: sub}, nil
}

// WatchSuicide is a free log subscription operation binding the contract event 0xa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde.
//
// Solidity: e Suicide(block uint256)
func (_AutoPayout *AutoPayoutFilterer) WatchSuicide(opts *bind.WatchOpts, sink chan<- *AutoPayoutSuicide) (event.Subscription, error) {

	logs, sub, err := _AutoPayout.contract.WatchLogs(opts, "Suicide")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AutoPayoutSuicide)
				if err := _AutoPayout.contract.UnpackLog(event, "Suicide", log); err != nil {
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
