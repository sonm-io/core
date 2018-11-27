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

// OracleUSDABI is the input ABI used to generate the binding from.
const OracleUSDABI = "[{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"price\",\"type\":\"uint256\"}],\"name\":\"PriceChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"}],\"name\":\"OwnershipRenounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_price\",\"type\":\"uint256\"}],\"name\":\"setCurrentPrice\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getCurrentPrice\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// OracleUSDBin is the compiled bytecode used for deploying new contracts.
const OracleUSDBin = `0x60806040526001805534801561001457600080fd5b5060008054600160a060020a0319908116339081179091161790556102c58061003e6000396000f30060806040526004361061006c5763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166318b200718114610071578063715018a61461008b5780638da5cb5b146100a0578063eb91d37e146100d1578063f2fde38b146100f8575b600080fd5b34801561007d57600080fd5b50610089600435610119565b005b34801561009757600080fd5b50610089610178565b3480156100ac57600080fd5b506100b56101e4565b60408051600160a060020a039092168252519081900360200190f35b3480156100dd57600080fd5b506100e66101f3565b60408051918252519081900360200190f35b34801561010457600080fd5b50610089600160a060020a03600435166101f9565b600054600160a060020a0316331461013057600080fd5b6000811161013d57600080fd5b60018190556040805182815290517fa6dc15bdb68da224c66db4b3838d9a2b205138e8cff6774e57d0af91e196d6229181900360200190a150565b600054600160a060020a0316331461018f57600080fd5b60008054604051600160a060020a03909116917ff8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c6482091a26000805473ffffffffffffffffffffffffffffffffffffffff19169055565b600054600160a060020a031681565b60015490565b600054600160a060020a0316331461021057600080fd5b6102198161021c565b50565b600160a060020a038116151561023157600080fd5b60008054604051600160a060020a03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a36000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a03929092169190911790555600a165627a7a72305820be6f69c6a3f6722d576a1b501aff2cf9b9de7b8c862ddb317bd0d7808c5a41ad0029`

// DeployOracleUSD deploys a new Ethereum contract, binding an instance of OracleUSD to it.
func DeployOracleUSD(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *OracleUSD, error) {
	parsed, err := abi.JSON(strings.NewReader(OracleUSDABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(OracleUSDBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &OracleUSD{OracleUSDCaller: OracleUSDCaller{contract: contract}, OracleUSDTransactor: OracleUSDTransactor{contract: contract}, OracleUSDFilterer: OracleUSDFilterer{contract: contract}}, nil
}

// OracleUSD is an auto generated Go binding around an Ethereum contract.
type OracleUSD struct {
	OracleUSDCaller     // Read-only binding to the contract
	OracleUSDTransactor // Write-only binding to the contract
	OracleUSDFilterer   // Log filterer for contract events
}

// OracleUSDCaller is an auto generated read-only Go binding around an Ethereum contract.
type OracleUSDCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OracleUSDTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OracleUSDTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OracleUSDFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OracleUSDFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OracleUSDSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OracleUSDSession struct {
	Contract     *OracleUSD        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OracleUSDCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OracleUSDCallerSession struct {
	Contract *OracleUSDCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// OracleUSDTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OracleUSDTransactorSession struct {
	Contract     *OracleUSDTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// OracleUSDRaw is an auto generated low-level Go binding around an Ethereum contract.
type OracleUSDRaw struct {
	Contract *OracleUSD // Generic contract binding to access the raw methods on
}

// OracleUSDCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OracleUSDCallerRaw struct {
	Contract *OracleUSDCaller // Generic read-only contract binding to access the raw methods on
}

// OracleUSDTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OracleUSDTransactorRaw struct {
	Contract *OracleUSDTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOracleUSD creates a new instance of OracleUSD, bound to a specific deployed contract.
func NewOracleUSD(address common.Address, backend bind.ContractBackend) (*OracleUSD, error) {
	contract, err := bindOracleUSD(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OracleUSD{OracleUSDCaller: OracleUSDCaller{contract: contract}, OracleUSDTransactor: OracleUSDTransactor{contract: contract}, OracleUSDFilterer: OracleUSDFilterer{contract: contract}}, nil
}

// NewOracleUSDCaller creates a new read-only instance of OracleUSD, bound to a specific deployed contract.
func NewOracleUSDCaller(address common.Address, caller bind.ContractCaller) (*OracleUSDCaller, error) {
	contract, err := bindOracleUSD(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OracleUSDCaller{contract: contract}, nil
}

// NewOracleUSDTransactor creates a new write-only instance of OracleUSD, bound to a specific deployed contract.
func NewOracleUSDTransactor(address common.Address, transactor bind.ContractTransactor) (*OracleUSDTransactor, error) {
	contract, err := bindOracleUSD(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OracleUSDTransactor{contract: contract}, nil
}

// NewOracleUSDFilterer creates a new log filterer instance of OracleUSD, bound to a specific deployed contract.
func NewOracleUSDFilterer(address common.Address, filterer bind.ContractFilterer) (*OracleUSDFilterer, error) {
	contract, err := bindOracleUSD(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OracleUSDFilterer{contract: contract}, nil
}

// bindOracleUSD binds a generic wrapper to an already deployed contract.
func bindOracleUSD(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(OracleUSDABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OracleUSD *OracleUSDRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _OracleUSD.Contract.OracleUSDCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OracleUSD *OracleUSDRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OracleUSD.Contract.OracleUSDTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OracleUSD *OracleUSDRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OracleUSD.Contract.OracleUSDTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OracleUSD *OracleUSDCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _OracleUSD.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OracleUSD *OracleUSDTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OracleUSD.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OracleUSD *OracleUSDTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OracleUSD.Contract.contract.Transact(opts, method, params...)
}

// GetCurrentPrice is a free data retrieval call binding the contract method 0xeb91d37e.
//
// Solidity: function getCurrentPrice() constant returns(uint256)
func (_OracleUSD *OracleUSDCaller) GetCurrentPrice(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _OracleUSD.contract.Call(opts, out, "getCurrentPrice")
	return *ret0, err
}

// GetCurrentPrice is a free data retrieval call binding the contract method 0xeb91d37e.
//
// Solidity: function getCurrentPrice() constant returns(uint256)
func (_OracleUSD *OracleUSDSession) GetCurrentPrice() (*big.Int, error) {
	return _OracleUSD.Contract.GetCurrentPrice(&_OracleUSD.CallOpts)
}

// GetCurrentPrice is a free data retrieval call binding the contract method 0xeb91d37e.
//
// Solidity: function getCurrentPrice() constant returns(uint256)
func (_OracleUSD *OracleUSDCallerSession) GetCurrentPrice() (*big.Int, error) {
	return _OracleUSD.Contract.GetCurrentPrice(&_OracleUSD.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_OracleUSD *OracleUSDCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _OracleUSD.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_OracleUSD *OracleUSDSession) Owner() (common.Address, error) {
	return _OracleUSD.Contract.Owner(&_OracleUSD.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_OracleUSD *OracleUSDCallerSession) Owner() (common.Address, error) {
	return _OracleUSD.Contract.Owner(&_OracleUSD.CallOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_OracleUSD *OracleUSDTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OracleUSD.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_OracleUSD *OracleUSDSession) RenounceOwnership() (*types.Transaction, error) {
	return _OracleUSD.Contract.RenounceOwnership(&_OracleUSD.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_OracleUSD *OracleUSDTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _OracleUSD.Contract.RenounceOwnership(&_OracleUSD.TransactOpts)
}

// SetCurrentPrice is a paid mutator transaction binding the contract method 0x18b20071.
//
// Solidity: function setCurrentPrice(_price uint256) returns()
func (_OracleUSD *OracleUSDTransactor) SetCurrentPrice(opts *bind.TransactOpts, _price *big.Int) (*types.Transaction, error) {
	return _OracleUSD.contract.Transact(opts, "setCurrentPrice", _price)
}

// SetCurrentPrice is a paid mutator transaction binding the contract method 0x18b20071.
//
// Solidity: function setCurrentPrice(_price uint256) returns()
func (_OracleUSD *OracleUSDSession) SetCurrentPrice(_price *big.Int) (*types.Transaction, error) {
	return _OracleUSD.Contract.SetCurrentPrice(&_OracleUSD.TransactOpts, _price)
}

// SetCurrentPrice is a paid mutator transaction binding the contract method 0x18b20071.
//
// Solidity: function setCurrentPrice(_price uint256) returns()
func (_OracleUSD *OracleUSDTransactorSession) SetCurrentPrice(_price *big.Int) (*types.Transaction, error) {
	return _OracleUSD.Contract.SetCurrentPrice(&_OracleUSD.TransactOpts, _price)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_OracleUSD *OracleUSDTransactor) TransferOwnership(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _OracleUSD.contract.Transact(opts, "transferOwnership", _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_OracleUSD *OracleUSDSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _OracleUSD.Contract.TransferOwnership(&_OracleUSD.TransactOpts, _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_OracleUSD *OracleUSDTransactorSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _OracleUSD.Contract.TransferOwnership(&_OracleUSD.TransactOpts, _newOwner)
}

// OracleUSDOwnershipRenouncedIterator is returned from FilterOwnershipRenounced and is used to iterate over the raw logs and unpacked data for OwnershipRenounced events raised by the OracleUSD contract.
type OracleUSDOwnershipRenouncedIterator struct {
	Event *OracleUSDOwnershipRenounced // Event containing the contract specifics and raw log

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
func (it *OracleUSDOwnershipRenouncedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleUSDOwnershipRenounced)
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
		it.Event = new(OracleUSDOwnershipRenounced)
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
func (it *OracleUSDOwnershipRenouncedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleUSDOwnershipRenouncedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleUSDOwnershipRenounced represents a OwnershipRenounced event raised by the OracleUSD contract.
type OracleUSDOwnershipRenounced struct {
	PreviousOwner common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipRenounced is a free log retrieval operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_OracleUSD *OracleUSDFilterer) FilterOwnershipRenounced(opts *bind.FilterOpts, previousOwner []common.Address) (*OracleUSDOwnershipRenouncedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _OracleUSD.contract.FilterLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return &OracleUSDOwnershipRenouncedIterator{contract: _OracleUSD.contract, event: "OwnershipRenounced", logs: logs, sub: sub}, nil
}

// WatchOwnershipRenounced is a free log subscription operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_OracleUSD *OracleUSDFilterer) WatchOwnershipRenounced(opts *bind.WatchOpts, sink chan<- *OracleUSDOwnershipRenounced, previousOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _OracleUSD.contract.WatchLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleUSDOwnershipRenounced)
				if err := _OracleUSD.contract.UnpackLog(event, "OwnershipRenounced", log); err != nil {
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

// OracleUSDOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the OracleUSD contract.
type OracleUSDOwnershipTransferredIterator struct {
	Event *OracleUSDOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *OracleUSDOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleUSDOwnershipTransferred)
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
		it.Event = new(OracleUSDOwnershipTransferred)
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
func (it *OracleUSDOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleUSDOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleUSDOwnershipTransferred represents a OwnershipTransferred event raised by the OracleUSD contract.
type OracleUSDOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_OracleUSD *OracleUSDFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*OracleUSDOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _OracleUSD.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &OracleUSDOwnershipTransferredIterator{contract: _OracleUSD.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_OracleUSD *OracleUSDFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *OracleUSDOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _OracleUSD.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleUSDOwnershipTransferred)
				if err := _OracleUSD.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// OracleUSDPriceChangedIterator is returned from FilterPriceChanged and is used to iterate over the raw logs and unpacked data for PriceChanged events raised by the OracleUSD contract.
type OracleUSDPriceChangedIterator struct {
	Event *OracleUSDPriceChanged // Event containing the contract specifics and raw log

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
func (it *OracleUSDPriceChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleUSDPriceChanged)
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
		it.Event = new(OracleUSDPriceChanged)
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
func (it *OracleUSDPriceChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleUSDPriceChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleUSDPriceChanged represents a PriceChanged event raised by the OracleUSD contract.
type OracleUSDPriceChanged struct {
	Price *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterPriceChanged is a free log retrieval operation binding the contract event 0xa6dc15bdb68da224c66db4b3838d9a2b205138e8cff6774e57d0af91e196d622.
//
// Solidity: e PriceChanged(price uint256)
func (_OracleUSD *OracleUSDFilterer) FilterPriceChanged(opts *bind.FilterOpts) (*OracleUSDPriceChangedIterator, error) {

	logs, sub, err := _OracleUSD.contract.FilterLogs(opts, "PriceChanged")
	if err != nil {
		return nil, err
	}
	return &OracleUSDPriceChangedIterator{contract: _OracleUSD.contract, event: "PriceChanged", logs: logs, sub: sub}, nil
}

// WatchPriceChanged is a free log subscription operation binding the contract event 0xa6dc15bdb68da224c66db4b3838d9a2b205138e8cff6774e57d0af91e196d622.
//
// Solidity: e PriceChanged(price uint256)
func (_OracleUSD *OracleUSDFilterer) WatchPriceChanged(opts *bind.WatchOpts, sink chan<- *OracleUSDPriceChanged) (event.Subscription, error) {

	logs, sub, err := _OracleUSD.contract.WatchLogs(opts, "PriceChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleUSDPriceChanged)
				if err := _OracleUSD.contract.UnpackLog(event, "PriceChanged", log); err != nil {
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
