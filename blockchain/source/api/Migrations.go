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

// MigrationsABI is the input ABI used to generate the binding from.
const MigrationsABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"last_completed_migration\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"}],\"name\":\"OwnershipRenounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"completed\",\"type\":\"uint256\"}],\"name\":\"setCompleted\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"new_address\",\"type\":\"address\"}],\"name\":\"upgrade\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// MigrationsBin is the compiled bytecode used for deploying new contracts.
const MigrationsBin = `0x608060405234801561001057600080fd5b5060008054600160a060020a0319908116339081179091161790556103498061003a6000396000f3006080604052600436106100775763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416630900f010811461007c578063445df0ac1461009f578063715018a6146100c65780638da5cb5b146100db578063f2fde38b1461010c578063fdacd5761461012d575b600080fd5b34801561008857600080fd5b5061009d600160a060020a0360043516610145565b005b3480156100ab57600080fd5b506100b46101e0565b60408051918252519081900360200190f35b3480156100d257600080fd5b5061009d6101e6565b3480156100e757600080fd5b506100f0610252565b60408051600160a060020a039092168252519081900360200190f35b34801561011857600080fd5b5061009d600160a060020a0360043516610261565b34801561013957600080fd5b5061009d600435610284565b60008054600160a060020a0316331461015d57600080fd5b81905080600160a060020a031663fdacd5766001546040518263ffffffff167c010000000000000000000000000000000000000000000000000000000002815260040180828152602001915050600060405180830381600087803b1580156101c457600080fd5b505af11580156101d8573d6000803e3d6000fd5b505050505050565b60015481565b600054600160a060020a031633146101fd57600080fd5b60008054604051600160a060020a03909116917ff8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c6482091a26000805473ffffffffffffffffffffffffffffffffffffffff19169055565b600054600160a060020a031681565b600054600160a060020a0316331461027857600080fd5b610281816102a0565b50565b600054600160a060020a0316331461029b57600080fd5b600155565b600160a060020a03811615156102b557600080fd5b60008054604051600160a060020a03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a36000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a03929092169190911790555600a165627a7a72305820774bb1e7dc98687afa76cad6808ac981da0465576214be259cbeaba2f29c72060029`

// DeployMigrations deploys a new Ethereum contract, binding an instance of Migrations to it.
func DeployMigrations(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Migrations, error) {
	parsed, err := abi.JSON(strings.NewReader(MigrationsABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(MigrationsBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Migrations{MigrationsCaller: MigrationsCaller{contract: contract}, MigrationsTransactor: MigrationsTransactor{contract: contract}, MigrationsFilterer: MigrationsFilterer{contract: contract}}, nil
}

// Migrations is an auto generated Go binding around an Ethereum contract.
type Migrations struct {
	MigrationsCaller     // Read-only binding to the contract
	MigrationsTransactor // Write-only binding to the contract
	MigrationsFilterer   // Log filterer for contract events
}

// MigrationsCaller is an auto generated read-only Go binding around an Ethereum contract.
type MigrationsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MigrationsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MigrationsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MigrationsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MigrationsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MigrationsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MigrationsSession struct {
	Contract     *Migrations       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MigrationsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MigrationsCallerSession struct {
	Contract *MigrationsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// MigrationsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MigrationsTransactorSession struct {
	Contract     *MigrationsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// MigrationsRaw is an auto generated low-level Go binding around an Ethereum contract.
type MigrationsRaw struct {
	Contract *Migrations // Generic contract binding to access the raw methods on
}

// MigrationsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MigrationsCallerRaw struct {
	Contract *MigrationsCaller // Generic read-only contract binding to access the raw methods on
}

// MigrationsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MigrationsTransactorRaw struct {
	Contract *MigrationsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMigrations creates a new instance of Migrations, bound to a specific deployed contract.
func NewMigrations(address common.Address, backend bind.ContractBackend) (*Migrations, error) {
	contract, err := bindMigrations(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Migrations{MigrationsCaller: MigrationsCaller{contract: contract}, MigrationsTransactor: MigrationsTransactor{contract: contract}, MigrationsFilterer: MigrationsFilterer{contract: contract}}, nil
}

// NewMigrationsCaller creates a new read-only instance of Migrations, bound to a specific deployed contract.
func NewMigrationsCaller(address common.Address, caller bind.ContractCaller) (*MigrationsCaller, error) {
	contract, err := bindMigrations(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MigrationsCaller{contract: contract}, nil
}

// NewMigrationsTransactor creates a new write-only instance of Migrations, bound to a specific deployed contract.
func NewMigrationsTransactor(address common.Address, transactor bind.ContractTransactor) (*MigrationsTransactor, error) {
	contract, err := bindMigrations(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MigrationsTransactor{contract: contract}, nil
}

// NewMigrationsFilterer creates a new log filterer instance of Migrations, bound to a specific deployed contract.
func NewMigrationsFilterer(address common.Address, filterer bind.ContractFilterer) (*MigrationsFilterer, error) {
	contract, err := bindMigrations(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MigrationsFilterer{contract: contract}, nil
}

// bindMigrations binds a generic wrapper to an already deployed contract.
func bindMigrations(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MigrationsABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Migrations *MigrationsRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Migrations.Contract.MigrationsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Migrations *MigrationsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Migrations.Contract.MigrationsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Migrations *MigrationsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Migrations.Contract.MigrationsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Migrations *MigrationsCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Migrations.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Migrations *MigrationsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Migrations.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Migrations *MigrationsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Migrations.Contract.contract.Transact(opts, method, params...)
}

// LastCompletedMigration is a free data retrieval call binding the contract method 0x445df0ac.
//
// Solidity: function last_completed_migration() constant returns(uint256)
func (_Migrations *MigrationsCaller) LastCompletedMigration(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Migrations.contract.Call(opts, out, "last_completed_migration")
	return *ret0, err
}

// LastCompletedMigration is a free data retrieval call binding the contract method 0x445df0ac.
//
// Solidity: function last_completed_migration() constant returns(uint256)
func (_Migrations *MigrationsSession) LastCompletedMigration() (*big.Int, error) {
	return _Migrations.Contract.LastCompletedMigration(&_Migrations.CallOpts)
}

// LastCompletedMigration is a free data retrieval call binding the contract method 0x445df0ac.
//
// Solidity: function last_completed_migration() constant returns(uint256)
func (_Migrations *MigrationsCallerSession) LastCompletedMigration() (*big.Int, error) {
	return _Migrations.Contract.LastCompletedMigration(&_Migrations.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Migrations *MigrationsCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Migrations.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Migrations *MigrationsSession) Owner() (common.Address, error) {
	return _Migrations.Contract.Owner(&_Migrations.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Migrations *MigrationsCallerSession) Owner() (common.Address, error) {
	return _Migrations.Contract.Owner(&_Migrations.CallOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Migrations *MigrationsTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Migrations.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Migrations *MigrationsSession) RenounceOwnership() (*types.Transaction, error) {
	return _Migrations.Contract.RenounceOwnership(&_Migrations.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Migrations *MigrationsTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Migrations.Contract.RenounceOwnership(&_Migrations.TransactOpts)
}

// SetCompleted is a paid mutator transaction binding the contract method 0xfdacd576.
//
// Solidity: function setCompleted(completed uint256) returns()
func (_Migrations *MigrationsTransactor) SetCompleted(opts *bind.TransactOpts, completed *big.Int) (*types.Transaction, error) {
	return _Migrations.contract.Transact(opts, "setCompleted", completed)
}

// SetCompleted is a paid mutator transaction binding the contract method 0xfdacd576.
//
// Solidity: function setCompleted(completed uint256) returns()
func (_Migrations *MigrationsSession) SetCompleted(completed *big.Int) (*types.Transaction, error) {
	return _Migrations.Contract.SetCompleted(&_Migrations.TransactOpts, completed)
}

// SetCompleted is a paid mutator transaction binding the contract method 0xfdacd576.
//
// Solidity: function setCompleted(completed uint256) returns()
func (_Migrations *MigrationsTransactorSession) SetCompleted(completed *big.Int) (*types.Transaction, error) {
	return _Migrations.Contract.SetCompleted(&_Migrations.TransactOpts, completed)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_Migrations *MigrationsTransactor) TransferOwnership(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _Migrations.contract.Transact(opts, "transferOwnership", _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_Migrations *MigrationsSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _Migrations.Contract.TransferOwnership(&_Migrations.TransactOpts, _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_Migrations *MigrationsTransactorSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _Migrations.Contract.TransferOwnership(&_Migrations.TransactOpts, _newOwner)
}

// Upgrade is a paid mutator transaction binding the contract method 0x0900f010.
//
// Solidity: function upgrade(new_address address) returns()
func (_Migrations *MigrationsTransactor) Upgrade(opts *bind.TransactOpts, new_address common.Address) (*types.Transaction, error) {
	return _Migrations.contract.Transact(opts, "upgrade", new_address)
}

// Upgrade is a paid mutator transaction binding the contract method 0x0900f010.
//
// Solidity: function upgrade(new_address address) returns()
func (_Migrations *MigrationsSession) Upgrade(new_address common.Address) (*types.Transaction, error) {
	return _Migrations.Contract.Upgrade(&_Migrations.TransactOpts, new_address)
}

// Upgrade is a paid mutator transaction binding the contract method 0x0900f010.
//
// Solidity: function upgrade(new_address address) returns()
func (_Migrations *MigrationsTransactorSession) Upgrade(new_address common.Address) (*types.Transaction, error) {
	return _Migrations.Contract.Upgrade(&_Migrations.TransactOpts, new_address)
}

// MigrationsOwnershipRenouncedIterator is returned from FilterOwnershipRenounced and is used to iterate over the raw logs and unpacked data for OwnershipRenounced events raised by the Migrations contract.
type MigrationsOwnershipRenouncedIterator struct {
	Event *MigrationsOwnershipRenounced // Event containing the contract specifics and raw log

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
func (it *MigrationsOwnershipRenouncedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MigrationsOwnershipRenounced)
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
		it.Event = new(MigrationsOwnershipRenounced)
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
func (it *MigrationsOwnershipRenouncedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MigrationsOwnershipRenouncedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MigrationsOwnershipRenounced represents a OwnershipRenounced event raised by the Migrations contract.
type MigrationsOwnershipRenounced struct {
	PreviousOwner common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipRenounced is a free log retrieval operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_Migrations *MigrationsFilterer) FilterOwnershipRenounced(opts *bind.FilterOpts, previousOwner []common.Address) (*MigrationsOwnershipRenouncedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _Migrations.contract.FilterLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return &MigrationsOwnershipRenouncedIterator{contract: _Migrations.contract, event: "OwnershipRenounced", logs: logs, sub: sub}, nil
}

// WatchOwnershipRenounced is a free log subscription operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_Migrations *MigrationsFilterer) WatchOwnershipRenounced(opts *bind.WatchOpts, sink chan<- *MigrationsOwnershipRenounced, previousOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _Migrations.contract.WatchLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MigrationsOwnershipRenounced)
				if err := _Migrations.contract.UnpackLog(event, "OwnershipRenounced", log); err != nil {
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

// MigrationsOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Migrations contract.
type MigrationsOwnershipTransferredIterator struct {
	Event *MigrationsOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *MigrationsOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MigrationsOwnershipTransferred)
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
		it.Event = new(MigrationsOwnershipTransferred)
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
func (it *MigrationsOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MigrationsOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MigrationsOwnershipTransferred represents a OwnershipTransferred event raised by the Migrations contract.
type MigrationsOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_Migrations *MigrationsFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*MigrationsOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Migrations.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &MigrationsOwnershipTransferredIterator{contract: _Migrations.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_Migrations *MigrationsFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *MigrationsOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Migrations.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MigrationsOwnershipTransferred)
				if err := _Migrations.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
