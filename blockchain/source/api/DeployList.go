// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package api

import (
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// DeployListABI is the input ABI used to generate the binding from.
const DeployListABI = "[{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_deployers\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"deployer\",\"type\":\"address\"}],\"name\":\"DeployerAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"deployer\",\"type\":\"address\"}],\"name\":\"DeployerRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"}],\"name\":\"OwnershipRenounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_deployer\",\"type\":\"address\"}],\"name\":\"addDeployer\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_deployer\",\"type\":\"address\"}],\"name\":\"removeDeployer\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getDeployers\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// DeployListBin is the compiled bytecode used for deploying new contracts.
const DeployListBin = `0x608060405234801561001057600080fd5b5060405161063e38038061063e8339810160405280516000805433600160a060020a0319918216811790911617905501805161005390600190602084019061005a565b50506100e6565b8280548282559060005260206000209081019282156100af579160200282015b828111156100af5782518254600160a060020a031916600160a060020a0390911617825560209092019160019091019061007a565b506100bb9291506100bf565b5090565b6100e391905b808211156100bb578054600160a060020a03191681556001016100c5565b90565b610549806100f56000396000f3006080604052600436106100775763ffffffff7c0100000000000000000000000000000000000000000000000000000000600035041663607c12b5811461007c578063715018a6146100e1578063880f4039146100f85780638da5cb5b14610119578063f2fde38b1461014a578063f315df861461016b575b600080fd5b34801561008857600080fd5b5061009161018c565b60408051602080825283518183015283519192839290830191858101910280838360005b838110156100cd5781810151838201526020016100b5565b505050509050019250505060405180910390f35b3480156100ed57600080fd5b506100f66101ef565b005b34801561010457600080fd5b506100f6600160a060020a036004351661025b565b34801561012557600080fd5b5061012e610302565b60408051600160a060020a039092168252519081900360200190f35b34801561015657600080fd5b506100f6600160a060020a0360043516610311565b34801561017757600080fd5b506100f6600160a060020a0360043516610334565b606060018054806020026020016040519081016040528092919081815260200182805480156101e457602002820191906000526020600020905b8154600160a060020a031681526001909101906020018083116101c6575b505050505090505b90565b600054600160a060020a0316331461020657600080fd5b60008054604051600160a060020a03909116917ff8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c6482091a26000805473ffffffffffffffffffffffffffffffffffffffff19169055565b600054600160a060020a0316331461027257600080fd5b6001805480820182556000919091527fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6018054600160a060020a03831673ffffffffffffffffffffffffffffffffffffffff19909116811790915560408051918252517f861a21548a3ee34d896ccac3668a9d65030aaf2cb7367a2ed13608014016a0329181900360200190a150565b600054600160a060020a031681565b600054600160a060020a0316331461032857600080fd5b61033181610459565b50565b60008054600160a060020a0316331461034c57600080fd5b5060005b600154600019018110156104055781600160a060020a031660018281548110151561037757fe5b600091825260209091200154600160a060020a031614156103fd576001805460001981019081106103a457fe5b60009182526020909120015460018054600160a060020a0390921691839081106103ca57fe5b9060005260206000200160006101000a815481600160a060020a030219169083600160a060020a03160217905550610405565b600101610350565b60018054600019019061041890826104d6565b5060408051600160a060020a038416815290517ffdb22628e87f888d060acc53d048a6a8400a5024f81f9dcb0606e723f238864a9181900360200190a15050565b600160a060020a038116151561046e57600080fd5b60008054604051600160a060020a03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a36000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0392909216919091179055565b8154818355818111156104fa576000838152602090206104fa9181019083016104ff565b505050565b6101ec91905b808211156105195760008155600101610505565b50905600a165627a7a72305820e9fc8432afc30cdd5dc2cb14c650ee92c84cddf8014f5d38b3a7dc97efabcbc90029`

// DeployDeployList deploys a new Ethereum contract, binding an instance of DeployList to it.
func DeployDeployList(auth *bind.TransactOpts, backend bind.ContractBackend, _deployers []common.Address) (common.Address, *types.Transaction, *DeployList, error) {
	parsed, err := abi.JSON(strings.NewReader(DeployListABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(DeployListBin), backend, _deployers)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &DeployList{DeployListCaller: DeployListCaller{contract: contract}, DeployListTransactor: DeployListTransactor{contract: contract}, DeployListFilterer: DeployListFilterer{contract: contract}}, nil
}

// DeployList is an auto generated Go binding around an Ethereum contract.
type DeployList struct {
	DeployListCaller     // Read-only binding to the contract
	DeployListTransactor // Write-only binding to the contract
	DeployListFilterer   // Log filterer for contract events
}

// DeployListCaller is an auto generated read-only Go binding around an Ethereum contract.
type DeployListCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DeployListTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DeployListTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DeployListFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DeployListFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DeployListSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DeployListSession struct {
	Contract     *DeployList       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DeployListCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DeployListCallerSession struct {
	Contract *DeployListCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// DeployListTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DeployListTransactorSession struct {
	Contract     *DeployListTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// DeployListRaw is an auto generated low-level Go binding around an Ethereum contract.
type DeployListRaw struct {
	Contract *DeployList // Generic contract binding to access the raw methods on
}

// DeployListCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DeployListCallerRaw struct {
	Contract *DeployListCaller // Generic read-only contract binding to access the raw methods on
}

// DeployListTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DeployListTransactorRaw struct {
	Contract *DeployListTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDeployList creates a new instance of DeployList, bound to a specific deployed contract.
func NewDeployList(address common.Address, backend bind.ContractBackend) (*DeployList, error) {
	contract, err := bindDeployList(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &DeployList{DeployListCaller: DeployListCaller{contract: contract}, DeployListTransactor: DeployListTransactor{contract: contract}, DeployListFilterer: DeployListFilterer{contract: contract}}, nil
}

// NewDeployListCaller creates a new read-only instance of DeployList, bound to a specific deployed contract.
func NewDeployListCaller(address common.Address, caller bind.ContractCaller) (*DeployListCaller, error) {
	contract, err := bindDeployList(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DeployListCaller{contract: contract}, nil
}

// NewDeployListTransactor creates a new write-only instance of DeployList, bound to a specific deployed contract.
func NewDeployListTransactor(address common.Address, transactor bind.ContractTransactor) (*DeployListTransactor, error) {
	contract, err := bindDeployList(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DeployListTransactor{contract: contract}, nil
}

// NewDeployListFilterer creates a new log filterer instance of DeployList, bound to a specific deployed contract.
func NewDeployListFilterer(address common.Address, filterer bind.ContractFilterer) (*DeployListFilterer, error) {
	contract, err := bindDeployList(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DeployListFilterer{contract: contract}, nil
}

// bindDeployList binds a generic wrapper to an already deployed contract.
func bindDeployList(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(DeployListABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DeployList *DeployListRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _DeployList.Contract.DeployListCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DeployList *DeployListRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DeployList.Contract.DeployListTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DeployList *DeployListRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DeployList.Contract.DeployListTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DeployList *DeployListCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _DeployList.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DeployList *DeployListTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DeployList.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DeployList *DeployListTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DeployList.Contract.contract.Transact(opts, method, params...)
}

// GetDeployers is a free data retrieval call binding the contract method 0x607c12b5.
//
// Solidity: function getDeployers() constant returns(address[])
func (_DeployList *DeployListCaller) GetDeployers(opts *bind.CallOpts) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _DeployList.contract.Call(opts, out, "getDeployers")
	return *ret0, err
}

// GetDeployers is a free data retrieval call binding the contract method 0x607c12b5.
//
// Solidity: function getDeployers() constant returns(address[])
func (_DeployList *DeployListSession) GetDeployers() ([]common.Address, error) {
	return _DeployList.Contract.GetDeployers(&_DeployList.CallOpts)
}

// GetDeployers is a free data retrieval call binding the contract method 0x607c12b5.
//
// Solidity: function getDeployers() constant returns(address[])
func (_DeployList *DeployListCallerSession) GetDeployers() ([]common.Address, error) {
	return _DeployList.Contract.GetDeployers(&_DeployList.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_DeployList *DeployListCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _DeployList.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_DeployList *DeployListSession) Owner() (common.Address, error) {
	return _DeployList.Contract.Owner(&_DeployList.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_DeployList *DeployListCallerSession) Owner() (common.Address, error) {
	return _DeployList.Contract.Owner(&_DeployList.CallOpts)
}

// AddDeployer is a paid mutator transaction binding the contract method 0x880f4039.
//
// Solidity: function addDeployer(_deployer address) returns()
func (_DeployList *DeployListTransactor) AddDeployer(opts *bind.TransactOpts, _deployer common.Address) (*types.Transaction, error) {
	return _DeployList.contract.Transact(opts, "addDeployer", _deployer)
}

// AddDeployer is a paid mutator transaction binding the contract method 0x880f4039.
//
// Solidity: function addDeployer(_deployer address) returns()
func (_DeployList *DeployListSession) AddDeployer(_deployer common.Address) (*types.Transaction, error) {
	return _DeployList.Contract.AddDeployer(&_DeployList.TransactOpts, _deployer)
}

// AddDeployer is a paid mutator transaction binding the contract method 0x880f4039.
//
// Solidity: function addDeployer(_deployer address) returns()
func (_DeployList *DeployListTransactorSession) AddDeployer(_deployer common.Address) (*types.Transaction, error) {
	return _DeployList.Contract.AddDeployer(&_DeployList.TransactOpts, _deployer)
}

// RemoveDeployer is a paid mutator transaction binding the contract method 0xf315df86.
//
// Solidity: function removeDeployer(_deployer address) returns()
func (_DeployList *DeployListTransactor) RemoveDeployer(opts *bind.TransactOpts, _deployer common.Address) (*types.Transaction, error) {
	return _DeployList.contract.Transact(opts, "removeDeployer", _deployer)
}

// RemoveDeployer is a paid mutator transaction binding the contract method 0xf315df86.
//
// Solidity: function removeDeployer(_deployer address) returns()
func (_DeployList *DeployListSession) RemoveDeployer(_deployer common.Address) (*types.Transaction, error) {
	return _DeployList.Contract.RemoveDeployer(&_DeployList.TransactOpts, _deployer)
}

// RemoveDeployer is a paid mutator transaction binding the contract method 0xf315df86.
//
// Solidity: function removeDeployer(_deployer address) returns()
func (_DeployList *DeployListTransactorSession) RemoveDeployer(_deployer common.Address) (*types.Transaction, error) {
	return _DeployList.Contract.RemoveDeployer(&_DeployList.TransactOpts, _deployer)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_DeployList *DeployListTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DeployList.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_DeployList *DeployListSession) RenounceOwnership() (*types.Transaction, error) {
	return _DeployList.Contract.RenounceOwnership(&_DeployList.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_DeployList *DeployListTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _DeployList.Contract.RenounceOwnership(&_DeployList.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_DeployList *DeployListTransactor) TransferOwnership(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _DeployList.contract.Transact(opts, "transferOwnership", _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_DeployList *DeployListSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _DeployList.Contract.TransferOwnership(&_DeployList.TransactOpts, _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_DeployList *DeployListTransactorSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _DeployList.Contract.TransferOwnership(&_DeployList.TransactOpts, _newOwner)
}

// DeployListDeployerAddedIterator is returned from FilterDeployerAdded and is used to iterate over the raw logs and unpacked data for DeployerAdded events raised by the DeployList contract.
type DeployListDeployerAddedIterator struct {
	Event *DeployListDeployerAdded // Event containing the contract specifics and raw log

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
func (it *DeployListDeployerAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DeployListDeployerAdded)
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
		it.Event = new(DeployListDeployerAdded)
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
func (it *DeployListDeployerAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DeployListDeployerAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DeployListDeployerAdded represents a DeployerAdded event raised by the DeployList contract.
type DeployListDeployerAdded struct {
	Deployer common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterDeployerAdded is a free log retrieval operation binding the contract event 0x861a21548a3ee34d896ccac3668a9d65030aaf2cb7367a2ed13608014016a032.
//
// Solidity: e DeployerAdded(deployer address)
func (_DeployList *DeployListFilterer) FilterDeployerAdded(opts *bind.FilterOpts) (*DeployListDeployerAddedIterator, error) {

	logs, sub, err := _DeployList.contract.FilterLogs(opts, "DeployerAdded")
	if err != nil {
		return nil, err
	}
	return &DeployListDeployerAddedIterator{contract: _DeployList.contract, event: "DeployerAdded", logs: logs, sub: sub}, nil
}

// WatchDeployerAdded is a free log subscription operation binding the contract event 0x861a21548a3ee34d896ccac3668a9d65030aaf2cb7367a2ed13608014016a032.
//
// Solidity: e DeployerAdded(deployer address)
func (_DeployList *DeployListFilterer) WatchDeployerAdded(opts *bind.WatchOpts, sink chan<- *DeployListDeployerAdded) (event.Subscription, error) {

	logs, sub, err := _DeployList.contract.WatchLogs(opts, "DeployerAdded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DeployListDeployerAdded)
				if err := _DeployList.contract.UnpackLog(event, "DeployerAdded", log); err != nil {
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

// DeployListDeployerRemovedIterator is returned from FilterDeployerRemoved and is used to iterate over the raw logs and unpacked data for DeployerRemoved events raised by the DeployList contract.
type DeployListDeployerRemovedIterator struct {
	Event *DeployListDeployerRemoved // Event containing the contract specifics and raw log

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
func (it *DeployListDeployerRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DeployListDeployerRemoved)
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
		it.Event = new(DeployListDeployerRemoved)
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
func (it *DeployListDeployerRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DeployListDeployerRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DeployListDeployerRemoved represents a DeployerRemoved event raised by the DeployList contract.
type DeployListDeployerRemoved struct {
	Deployer common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterDeployerRemoved is a free log retrieval operation binding the contract event 0xfdb22628e87f888d060acc53d048a6a8400a5024f81f9dcb0606e723f238864a.
//
// Solidity: e DeployerRemoved(deployer address)
func (_DeployList *DeployListFilterer) FilterDeployerRemoved(opts *bind.FilterOpts) (*DeployListDeployerRemovedIterator, error) {

	logs, sub, err := _DeployList.contract.FilterLogs(opts, "DeployerRemoved")
	if err != nil {
		return nil, err
	}
	return &DeployListDeployerRemovedIterator{contract: _DeployList.contract, event: "DeployerRemoved", logs: logs, sub: sub}, nil
}

// WatchDeployerRemoved is a free log subscription operation binding the contract event 0xfdb22628e87f888d060acc53d048a6a8400a5024f81f9dcb0606e723f238864a.
//
// Solidity: e DeployerRemoved(deployer address)
func (_DeployList *DeployListFilterer) WatchDeployerRemoved(opts *bind.WatchOpts, sink chan<- *DeployListDeployerRemoved) (event.Subscription, error) {

	logs, sub, err := _DeployList.contract.WatchLogs(opts, "DeployerRemoved")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DeployListDeployerRemoved)
				if err := _DeployList.contract.UnpackLog(event, "DeployerRemoved", log); err != nil {
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

// DeployListOwnershipRenouncedIterator is returned from FilterOwnershipRenounced and is used to iterate over the raw logs and unpacked data for OwnershipRenounced events raised by the DeployList contract.
type DeployListOwnershipRenouncedIterator struct {
	Event *DeployListOwnershipRenounced // Event containing the contract specifics and raw log

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
func (it *DeployListOwnershipRenouncedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DeployListOwnershipRenounced)
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
		it.Event = new(DeployListOwnershipRenounced)
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
func (it *DeployListOwnershipRenouncedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DeployListOwnershipRenouncedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DeployListOwnershipRenounced represents a OwnershipRenounced event raised by the DeployList contract.
type DeployListOwnershipRenounced struct {
	PreviousOwner common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipRenounced is a free log retrieval operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_DeployList *DeployListFilterer) FilterOwnershipRenounced(opts *bind.FilterOpts, previousOwner []common.Address) (*DeployListOwnershipRenouncedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _DeployList.contract.FilterLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return &DeployListOwnershipRenouncedIterator{contract: _DeployList.contract, event: "OwnershipRenounced", logs: logs, sub: sub}, nil
}

// WatchOwnershipRenounced is a free log subscription operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_DeployList *DeployListFilterer) WatchOwnershipRenounced(opts *bind.WatchOpts, sink chan<- *DeployListOwnershipRenounced, previousOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _DeployList.contract.WatchLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DeployListOwnershipRenounced)
				if err := _DeployList.contract.UnpackLog(event, "OwnershipRenounced", log); err != nil {
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

// DeployListOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the DeployList contract.
type DeployListOwnershipTransferredIterator struct {
	Event *DeployListOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *DeployListOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DeployListOwnershipTransferred)
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
		it.Event = new(DeployListOwnershipTransferred)
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
func (it *DeployListOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DeployListOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DeployListOwnershipTransferred represents a OwnershipTransferred event raised by the DeployList contract.
type DeployListOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_DeployList *DeployListFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*DeployListOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _DeployList.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &DeployListOwnershipTransferredIterator{contract: _DeployList.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_DeployList *DeployListFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *DeployListOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _DeployList.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DeployListOwnershipTransferred)
				if err := _DeployList.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
