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

// DevicesStorageABI is the input ABI used to generate the binding from.
const DevicesStorageABI = "[{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"DevicesHasSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"DevicesUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"DevicesTimestampUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"block\",\"type\":\"uint256\"}],\"name\":\"Suicide\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"}],\"name\":\"OwnershipRenounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_deviceList\",\"type\":\"bytes\"}],\"name\":\"SetDevices\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_hash\",\"type\":\"bytes32\"}],\"name\":\"Touch\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"Hash\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"GetDevices\",\"outputs\":[{\"name\":\"devices\",\"type\":\"bytes\"},{\"name\":\"timestamp\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"KillDevicesStorage\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// DevicesStorageBin is the compiled bytecode used for deploying new contracts.
const DevicesStorageBin = `0x608060405234801561001057600080fd5b5060008054600160a060020a0319908116339081179091161790556109b38061003a6000396000f30060806040526004361061008d5763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416632d8dce7b81146100925780633e550cd4146100be57806359ef39501461015e578063715018a61461019157806375ff58db146101a85780638da5cb5b14610201578063cd33174a14610232578063f2fde38b14610247575b600080fd5b34801561009e57600080fd5b506100aa600435610268565b604080519115158252519081900360200190f35b3480156100ca57600080fd5b506100df600160a060020a03600435166103de565b6040518080602001838152602001828103825284818151815260200191508051906020019080838360005b8381101561012257818101518382015260200161010a565b50505050905090810190601f16801561014f5780820380516001836020036101000a031916815260200191505b50935050505060405180910390f35b34801561016a57600080fd5b5061017f600160a060020a03600435166104b7565b60408051918252519081900360200190f35b34801561019d57600080fd5b506101a66105b4565b005b3480156101b457600080fd5b506040805160206004803580820135601f81018490048402850184019095528484526101a69436949293602493928401919081908401838280828437509497506106209650505050505050565b34801561020d57600080fd5b506102166107cd565b60408051600160a060020a039092168252519081900360200190f35b34801561023e57600080fd5b506101a66107dc565b34801561025357600080fd5b506101a6600160a060020a0360043516610834565b6000806001600033600160a060020a0316600160a060020a0316815260200190815260200160002060000160405160200180828054600181600116156101000203166002900480156102f15780601f106102cf5761010080835404028352918201916102f1565b820191906000526020600020905b8154815290600101906020018083116102dd575b50509150506040516020818303038152906040526040518082805190602001908083835b602083106103345780518252601f199092019160209182019101610315565b5181516020939093036101000a60001901801990911692169190911790526040519201829003909120935050508382149050801561039257507fc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a4708114155b1561008d5733600081815260016020819052604080832042920191909155517f84769e7ee13c01432c7d8b384e4a9ba21ead8189731b1049244807be07df834a9190a250600192915050565b606060006103ea6108d4565b600160a060020a0384166000908152600160208181526040928390208351815460029481161561010002600019011693909304601f810183900490920283016060908101855293830182815292939092849290918491908401828280156104925780601f1061046757610100808354040283529160200191610492565b820191906000526020600020905b81548152906001019060200180831161047557829003601f168201915b5050509183525050600191909101546020918201528151910151909590945092505050565b60006001600083600160a060020a0316600160a060020a03168152602001908152602001600020600001604051602001808280546001816001161561010002031660029004801561053f5780601f1061051d57610100808354040283529182019161053f565b820191906000526020600020905b81548152906001019060200180831161052b575b50509150506040516020818303038152906040526040518082805190602001908083835b602083106105825780518252601f199092019160209182019101610563565b5181516020939093036101000a6000190180199091169216919091179052604051920182900390912095945050505050565b600054600160a060020a031633146105cb57600080fd5b60008054604051600160a060020a03909116917ff8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c6482091a26000805473ffffffffffffffffffffffffffffffffffffffff19169055565b60408051336000908152600160208181529390912080547fc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a47094919391909101918291849160026101009183161591909102600019019091160480156106bc5780601f1061069a5761010080835404028352918201916106bc565b820191906000526020600020905b8154815290600101906020018083116106a8575b50509150506040516020818303038152906040526040518082805190602001908083835b602083106106ff5780518252601f1990920191602091820191016106e0565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040518091039020600019161415156107675760405133907f1cd3cdf902f594faafe10e417201b346dcfaceed5eb9fc913379951f0d40068f90600090a2610793565b60405133907fc7582f50a7f5f4ab5898d20e888a592d9338cb6bdd957aa01f89fd3198d0fa9690600090a25b33600090815260016020908152604090912082516107b3928401906108ec565b505033600090815260016020819052604090912042910155565b600054600160a060020a031681565b600054600160a060020a031633146107f357600080fd5b6040805142815290517fa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde9181900360200190a1600054600160a060020a0316ff5b600054600160a060020a0316331461084b57600080fd5b61085481610857565b50565b600160a060020a038116151561086c57600080fd5b60008054604051600160a060020a03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a36000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0392909216919091179055565b60408051808201909152606081526000602082015290565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061092d57805160ff191683800117855561095a565b8280016001018555821561095a579182015b8281111561095a57825182559160200191906001019061093f565b5061096692915061096a565b5090565b61098491905b808211156109665760008155600101610970565b905600a165627a7a72305820df50965bc39729ab0b29d9c276304718203612d4916d40e28786def0e53dbca50029`

// DeployDevicesStorage deploys a new Ethereum contract, binding an instance of DevicesStorage to it.
func DeployDevicesStorage(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *DevicesStorage, error) {
	parsed, err := abi.JSON(strings.NewReader(DevicesStorageABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(DevicesStorageBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &DevicesStorage{DevicesStorageCaller: DevicesStorageCaller{contract: contract}, DevicesStorageTransactor: DevicesStorageTransactor{contract: contract}, DevicesStorageFilterer: DevicesStorageFilterer{contract: contract}}, nil
}

// DevicesStorage is an auto generated Go binding around an Ethereum contract.
type DevicesStorage struct {
	DevicesStorageCaller     // Read-only binding to the contract
	DevicesStorageTransactor // Write-only binding to the contract
	DevicesStorageFilterer   // Log filterer for contract events
}

// DevicesStorageCaller is an auto generated read-only Go binding around an Ethereum contract.
type DevicesStorageCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DevicesStorageTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DevicesStorageTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DevicesStorageFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DevicesStorageFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DevicesStorageSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DevicesStorageSession struct {
	Contract     *DevicesStorage   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DevicesStorageCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DevicesStorageCallerSession struct {
	Contract *DevicesStorageCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// DevicesStorageTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DevicesStorageTransactorSession struct {
	Contract     *DevicesStorageTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// DevicesStorageRaw is an auto generated low-level Go binding around an Ethereum contract.
type DevicesStorageRaw struct {
	Contract *DevicesStorage // Generic contract binding to access the raw methods on
}

// DevicesStorageCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DevicesStorageCallerRaw struct {
	Contract *DevicesStorageCaller // Generic read-only contract binding to access the raw methods on
}

// DevicesStorageTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DevicesStorageTransactorRaw struct {
	Contract *DevicesStorageTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDevicesStorage creates a new instance of DevicesStorage, bound to a specific deployed contract.
func NewDevicesStorage(address common.Address, backend bind.ContractBackend) (*DevicesStorage, error) {
	contract, err := bindDevicesStorage(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &DevicesStorage{DevicesStorageCaller: DevicesStorageCaller{contract: contract}, DevicesStorageTransactor: DevicesStorageTransactor{contract: contract}, DevicesStorageFilterer: DevicesStorageFilterer{contract: contract}}, nil
}

// NewDevicesStorageCaller creates a new read-only instance of DevicesStorage, bound to a specific deployed contract.
func NewDevicesStorageCaller(address common.Address, caller bind.ContractCaller) (*DevicesStorageCaller, error) {
	contract, err := bindDevicesStorage(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DevicesStorageCaller{contract: contract}, nil
}

// NewDevicesStorageTransactor creates a new write-only instance of DevicesStorage, bound to a specific deployed contract.
func NewDevicesStorageTransactor(address common.Address, transactor bind.ContractTransactor) (*DevicesStorageTransactor, error) {
	contract, err := bindDevicesStorage(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DevicesStorageTransactor{contract: contract}, nil
}

// NewDevicesStorageFilterer creates a new log filterer instance of DevicesStorage, bound to a specific deployed contract.
func NewDevicesStorageFilterer(address common.Address, filterer bind.ContractFilterer) (*DevicesStorageFilterer, error) {
	contract, err := bindDevicesStorage(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DevicesStorageFilterer{contract: contract}, nil
}

// bindDevicesStorage binds a generic wrapper to an already deployed contract.
func bindDevicesStorage(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(DevicesStorageABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DevicesStorage *DevicesStorageRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _DevicesStorage.Contract.DevicesStorageCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DevicesStorage *DevicesStorageRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DevicesStorage.Contract.DevicesStorageTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DevicesStorage *DevicesStorageRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DevicesStorage.Contract.DevicesStorageTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DevicesStorage *DevicesStorageCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _DevicesStorage.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DevicesStorage *DevicesStorageTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DevicesStorage.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DevicesStorage *DevicesStorageTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DevicesStorage.Contract.contract.Transact(opts, method, params...)
}

// GetDevices is a free data retrieval call binding the contract method 0x3e550cd4.
//
// Solidity: function GetDevices(_owner address) constant returns(devices bytes, timestamp uint256)
func (_DevicesStorage *DevicesStorageCaller) GetDevices(opts *bind.CallOpts, _owner common.Address) (struct {
	Devices   []byte
	Timestamp *big.Int
}, error) {
	ret := new(struct {
		Devices   []byte
		Timestamp *big.Int
	})
	out := ret
	err := _DevicesStorage.contract.Call(opts, out, "GetDevices", _owner)
	return *ret, err
}

// GetDevices is a free data retrieval call binding the contract method 0x3e550cd4.
//
// Solidity: function GetDevices(_owner address) constant returns(devices bytes, timestamp uint256)
func (_DevicesStorage *DevicesStorageSession) GetDevices(_owner common.Address) (struct {
	Devices   []byte
	Timestamp *big.Int
}, error) {
	return _DevicesStorage.Contract.GetDevices(&_DevicesStorage.CallOpts, _owner)
}

// GetDevices is a free data retrieval call binding the contract method 0x3e550cd4.
//
// Solidity: function GetDevices(_owner address) constant returns(devices bytes, timestamp uint256)
func (_DevicesStorage *DevicesStorageCallerSession) GetDevices(_owner common.Address) (struct {
	Devices   []byte
	Timestamp *big.Int
}, error) {
	return _DevicesStorage.Contract.GetDevices(&_DevicesStorage.CallOpts, _owner)
}

// Hash is a free data retrieval call binding the contract method 0x59ef3950.
//
// Solidity: function Hash(_owner address) constant returns(bytes32)
func (_DevicesStorage *DevicesStorageCaller) Hash(opts *bind.CallOpts, _owner common.Address) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _DevicesStorage.contract.Call(opts, out, "Hash", _owner)
	return *ret0, err
}

// Hash is a free data retrieval call binding the contract method 0x59ef3950.
//
// Solidity: function Hash(_owner address) constant returns(bytes32)
func (_DevicesStorage *DevicesStorageSession) Hash(_owner common.Address) ([32]byte, error) {
	return _DevicesStorage.Contract.Hash(&_DevicesStorage.CallOpts, _owner)
}

// Hash is a free data retrieval call binding the contract method 0x59ef3950.
//
// Solidity: function Hash(_owner address) constant returns(bytes32)
func (_DevicesStorage *DevicesStorageCallerSession) Hash(_owner common.Address) ([32]byte, error) {
	return _DevicesStorage.Contract.Hash(&_DevicesStorage.CallOpts, _owner)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_DevicesStorage *DevicesStorageCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _DevicesStorage.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_DevicesStorage *DevicesStorageSession) Owner() (common.Address, error) {
	return _DevicesStorage.Contract.Owner(&_DevicesStorage.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_DevicesStorage *DevicesStorageCallerSession) Owner() (common.Address, error) {
	return _DevicesStorage.Contract.Owner(&_DevicesStorage.CallOpts)
}

// KillDevicesStorage is a paid mutator transaction binding the contract method 0xcd33174a.
//
// Solidity: function KillDevicesStorage() returns()
func (_DevicesStorage *DevicesStorageTransactor) KillDevicesStorage(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DevicesStorage.contract.Transact(opts, "KillDevicesStorage")
}

// KillDevicesStorage is a paid mutator transaction binding the contract method 0xcd33174a.
//
// Solidity: function KillDevicesStorage() returns()
func (_DevicesStorage *DevicesStorageSession) KillDevicesStorage() (*types.Transaction, error) {
	return _DevicesStorage.Contract.KillDevicesStorage(&_DevicesStorage.TransactOpts)
}

// KillDevicesStorage is a paid mutator transaction binding the contract method 0xcd33174a.
//
// Solidity: function KillDevicesStorage() returns()
func (_DevicesStorage *DevicesStorageTransactorSession) KillDevicesStorage() (*types.Transaction, error) {
	return _DevicesStorage.Contract.KillDevicesStorage(&_DevicesStorage.TransactOpts)
}

// SetDevices is a paid mutator transaction binding the contract method 0x75ff58db.
//
// Solidity: function SetDevices(_deviceList bytes) returns()
func (_DevicesStorage *DevicesStorageTransactor) SetDevices(opts *bind.TransactOpts, _deviceList []byte) (*types.Transaction, error) {
	return _DevicesStorage.contract.Transact(opts, "SetDevices", _deviceList)
}

// SetDevices is a paid mutator transaction binding the contract method 0x75ff58db.
//
// Solidity: function SetDevices(_deviceList bytes) returns()
func (_DevicesStorage *DevicesStorageSession) SetDevices(_deviceList []byte) (*types.Transaction, error) {
	return _DevicesStorage.Contract.SetDevices(&_DevicesStorage.TransactOpts, _deviceList)
}

// SetDevices is a paid mutator transaction binding the contract method 0x75ff58db.
//
// Solidity: function SetDevices(_deviceList bytes) returns()
func (_DevicesStorage *DevicesStorageTransactorSession) SetDevices(_deviceList []byte) (*types.Transaction, error) {
	return _DevicesStorage.Contract.SetDevices(&_DevicesStorage.TransactOpts, _deviceList)
}

// Touch is a paid mutator transaction binding the contract method 0x2d8dce7b.
//
// Solidity: function Touch(_hash bytes32) returns(bool)
func (_DevicesStorage *DevicesStorageTransactor) Touch(opts *bind.TransactOpts, _hash [32]byte) (*types.Transaction, error) {
	return _DevicesStorage.contract.Transact(opts, "Touch", _hash)
}

// Touch is a paid mutator transaction binding the contract method 0x2d8dce7b.
//
// Solidity: function Touch(_hash bytes32) returns(bool)
func (_DevicesStorage *DevicesStorageSession) Touch(_hash [32]byte) (*types.Transaction, error) {
	return _DevicesStorage.Contract.Touch(&_DevicesStorage.TransactOpts, _hash)
}

// Touch is a paid mutator transaction binding the contract method 0x2d8dce7b.
//
// Solidity: function Touch(_hash bytes32) returns(bool)
func (_DevicesStorage *DevicesStorageTransactorSession) Touch(_hash [32]byte) (*types.Transaction, error) {
	return _DevicesStorage.Contract.Touch(&_DevicesStorage.TransactOpts, _hash)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_DevicesStorage *DevicesStorageTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DevicesStorage.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_DevicesStorage *DevicesStorageSession) RenounceOwnership() (*types.Transaction, error) {
	return _DevicesStorage.Contract.RenounceOwnership(&_DevicesStorage.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_DevicesStorage *DevicesStorageTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _DevicesStorage.Contract.RenounceOwnership(&_DevicesStorage.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_DevicesStorage *DevicesStorageTransactor) TransferOwnership(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _DevicesStorage.contract.Transact(opts, "transferOwnership", _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_DevicesStorage *DevicesStorageSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _DevicesStorage.Contract.TransferOwnership(&_DevicesStorage.TransactOpts, _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_DevicesStorage *DevicesStorageTransactorSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _DevicesStorage.Contract.TransferOwnership(&_DevicesStorage.TransactOpts, _newOwner)
}

// DevicesStorageDevicesHasSetIterator is returned from FilterDevicesHasSet and is used to iterate over the raw logs and unpacked data for DevicesHasSet events raised by the DevicesStorage contract.
type DevicesStorageDevicesHasSetIterator struct {
	Event *DevicesStorageDevicesHasSet // Event containing the contract specifics and raw log

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
func (it *DevicesStorageDevicesHasSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DevicesStorageDevicesHasSet)
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
		it.Event = new(DevicesStorageDevicesHasSet)
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
func (it *DevicesStorageDevicesHasSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DevicesStorageDevicesHasSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DevicesStorageDevicesHasSet represents a DevicesHasSet event raised by the DevicesStorage contract.
type DevicesStorageDevicesHasSet struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDevicesHasSet is a free log retrieval operation binding the contract event 0xc7582f50a7f5f4ab5898d20e888a592d9338cb6bdd957aa01f89fd3198d0fa96.
//
// Solidity: e DevicesHasSet(owner indexed address)
func (_DevicesStorage *DevicesStorageFilterer) FilterDevicesHasSet(opts *bind.FilterOpts, owner []common.Address) (*DevicesStorageDevicesHasSetIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _DevicesStorage.contract.FilterLogs(opts, "DevicesHasSet", ownerRule)
	if err != nil {
		return nil, err
	}
	return &DevicesStorageDevicesHasSetIterator{contract: _DevicesStorage.contract, event: "DevicesHasSet", logs: logs, sub: sub}, nil
}

// WatchDevicesHasSet is a free log subscription operation binding the contract event 0xc7582f50a7f5f4ab5898d20e888a592d9338cb6bdd957aa01f89fd3198d0fa96.
//
// Solidity: e DevicesHasSet(owner indexed address)
func (_DevicesStorage *DevicesStorageFilterer) WatchDevicesHasSet(opts *bind.WatchOpts, sink chan<- *DevicesStorageDevicesHasSet, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _DevicesStorage.contract.WatchLogs(opts, "DevicesHasSet", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DevicesStorageDevicesHasSet)
				if err := _DevicesStorage.contract.UnpackLog(event, "DevicesHasSet", log); err != nil {
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

// DevicesStorageDevicesTimestampUpdatedIterator is returned from FilterDevicesTimestampUpdated and is used to iterate over the raw logs and unpacked data for DevicesTimestampUpdated events raised by the DevicesStorage contract.
type DevicesStorageDevicesTimestampUpdatedIterator struct {
	Event *DevicesStorageDevicesTimestampUpdated // Event containing the contract specifics and raw log

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
func (it *DevicesStorageDevicesTimestampUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DevicesStorageDevicesTimestampUpdated)
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
		it.Event = new(DevicesStorageDevicesTimestampUpdated)
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
func (it *DevicesStorageDevicesTimestampUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DevicesStorageDevicesTimestampUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DevicesStorageDevicesTimestampUpdated represents a DevicesTimestampUpdated event raised by the DevicesStorage contract.
type DevicesStorageDevicesTimestampUpdated struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDevicesTimestampUpdated is a free log retrieval operation binding the contract event 0x84769e7ee13c01432c7d8b384e4a9ba21ead8189731b1049244807be07df834a.
//
// Solidity: e DevicesTimestampUpdated(owner indexed address)
func (_DevicesStorage *DevicesStorageFilterer) FilterDevicesTimestampUpdated(opts *bind.FilterOpts, owner []common.Address) (*DevicesStorageDevicesTimestampUpdatedIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _DevicesStorage.contract.FilterLogs(opts, "DevicesTimestampUpdated", ownerRule)
	if err != nil {
		return nil, err
	}
	return &DevicesStorageDevicesTimestampUpdatedIterator{contract: _DevicesStorage.contract, event: "DevicesTimestampUpdated", logs: logs, sub: sub}, nil
}

// WatchDevicesTimestampUpdated is a free log subscription operation binding the contract event 0x84769e7ee13c01432c7d8b384e4a9ba21ead8189731b1049244807be07df834a.
//
// Solidity: e DevicesTimestampUpdated(owner indexed address)
func (_DevicesStorage *DevicesStorageFilterer) WatchDevicesTimestampUpdated(opts *bind.WatchOpts, sink chan<- *DevicesStorageDevicesTimestampUpdated, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _DevicesStorage.contract.WatchLogs(opts, "DevicesTimestampUpdated", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DevicesStorageDevicesTimestampUpdated)
				if err := _DevicesStorage.contract.UnpackLog(event, "DevicesTimestampUpdated", log); err != nil {
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

// DevicesStorageDevicesUpdatedIterator is returned from FilterDevicesUpdated and is used to iterate over the raw logs and unpacked data for DevicesUpdated events raised by the DevicesStorage contract.
type DevicesStorageDevicesUpdatedIterator struct {
	Event *DevicesStorageDevicesUpdated // Event containing the contract specifics and raw log

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
func (it *DevicesStorageDevicesUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DevicesStorageDevicesUpdated)
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
		it.Event = new(DevicesStorageDevicesUpdated)
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
func (it *DevicesStorageDevicesUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DevicesStorageDevicesUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DevicesStorageDevicesUpdated represents a DevicesUpdated event raised by the DevicesStorage contract.
type DevicesStorageDevicesUpdated struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDevicesUpdated is a free log retrieval operation binding the contract event 0x1cd3cdf902f594faafe10e417201b346dcfaceed5eb9fc913379951f0d40068f.
//
// Solidity: e DevicesUpdated(owner indexed address)
func (_DevicesStorage *DevicesStorageFilterer) FilterDevicesUpdated(opts *bind.FilterOpts, owner []common.Address) (*DevicesStorageDevicesUpdatedIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _DevicesStorage.contract.FilterLogs(opts, "DevicesUpdated", ownerRule)
	if err != nil {
		return nil, err
	}
	return &DevicesStorageDevicesUpdatedIterator{contract: _DevicesStorage.contract, event: "DevicesUpdated", logs: logs, sub: sub}, nil
}

// WatchDevicesUpdated is a free log subscription operation binding the contract event 0x1cd3cdf902f594faafe10e417201b346dcfaceed5eb9fc913379951f0d40068f.
//
// Solidity: e DevicesUpdated(owner indexed address)
func (_DevicesStorage *DevicesStorageFilterer) WatchDevicesUpdated(opts *bind.WatchOpts, sink chan<- *DevicesStorageDevicesUpdated, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _DevicesStorage.contract.WatchLogs(opts, "DevicesUpdated", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DevicesStorageDevicesUpdated)
				if err := _DevicesStorage.contract.UnpackLog(event, "DevicesUpdated", log); err != nil {
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

// DevicesStorageOwnershipRenouncedIterator is returned from FilterOwnershipRenounced and is used to iterate over the raw logs and unpacked data for OwnershipRenounced events raised by the DevicesStorage contract.
type DevicesStorageOwnershipRenouncedIterator struct {
	Event *DevicesStorageOwnershipRenounced // Event containing the contract specifics and raw log

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
func (it *DevicesStorageOwnershipRenouncedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DevicesStorageOwnershipRenounced)
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
		it.Event = new(DevicesStorageOwnershipRenounced)
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
func (it *DevicesStorageOwnershipRenouncedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DevicesStorageOwnershipRenouncedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DevicesStorageOwnershipRenounced represents a OwnershipRenounced event raised by the DevicesStorage contract.
type DevicesStorageOwnershipRenounced struct {
	PreviousOwner common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipRenounced is a free log retrieval operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_DevicesStorage *DevicesStorageFilterer) FilterOwnershipRenounced(opts *bind.FilterOpts, previousOwner []common.Address) (*DevicesStorageOwnershipRenouncedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _DevicesStorage.contract.FilterLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return &DevicesStorageOwnershipRenouncedIterator{contract: _DevicesStorage.contract, event: "OwnershipRenounced", logs: logs, sub: sub}, nil
}

// WatchOwnershipRenounced is a free log subscription operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_DevicesStorage *DevicesStorageFilterer) WatchOwnershipRenounced(opts *bind.WatchOpts, sink chan<- *DevicesStorageOwnershipRenounced, previousOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _DevicesStorage.contract.WatchLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DevicesStorageOwnershipRenounced)
				if err := _DevicesStorage.contract.UnpackLog(event, "OwnershipRenounced", log); err != nil {
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

// DevicesStorageOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the DevicesStorage contract.
type DevicesStorageOwnershipTransferredIterator struct {
	Event *DevicesStorageOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *DevicesStorageOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DevicesStorageOwnershipTransferred)
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
		it.Event = new(DevicesStorageOwnershipTransferred)
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
func (it *DevicesStorageOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DevicesStorageOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DevicesStorageOwnershipTransferred represents a OwnershipTransferred event raised by the DevicesStorage contract.
type DevicesStorageOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_DevicesStorage *DevicesStorageFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*DevicesStorageOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _DevicesStorage.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &DevicesStorageOwnershipTransferredIterator{contract: _DevicesStorage.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_DevicesStorage *DevicesStorageFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *DevicesStorageOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _DevicesStorage.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DevicesStorageOwnershipTransferred)
				if err := _DevicesStorage.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// DevicesStorageSuicideIterator is returned from FilterSuicide and is used to iterate over the raw logs and unpacked data for Suicide events raised by the DevicesStorage contract.
type DevicesStorageSuicideIterator struct {
	Event *DevicesStorageSuicide // Event containing the contract specifics and raw log

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
func (it *DevicesStorageSuicideIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DevicesStorageSuicide)
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
		it.Event = new(DevicesStorageSuicide)
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
func (it *DevicesStorageSuicideIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DevicesStorageSuicideIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DevicesStorageSuicide represents a Suicide event raised by the DevicesStorage contract.
type DevicesStorageSuicide struct {
	Block *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterSuicide is a free log retrieval operation binding the contract event 0xa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde.
//
// Solidity: e Suicide(block uint256)
func (_DevicesStorage *DevicesStorageFilterer) FilterSuicide(opts *bind.FilterOpts) (*DevicesStorageSuicideIterator, error) {

	logs, sub, err := _DevicesStorage.contract.FilterLogs(opts, "Suicide")
	if err != nil {
		return nil, err
	}
	return &DevicesStorageSuicideIterator{contract: _DevicesStorage.contract, event: "Suicide", logs: logs, sub: sub}, nil
}

// WatchSuicide is a free log subscription operation binding the contract event 0xa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde.
//
// Solidity: e Suicide(block uint256)
func (_DevicesStorage *DevicesStorageFilterer) WatchSuicide(opts *bind.WatchOpts, sink chan<- *DevicesStorageSuicide) (event.Subscription, error) {

	logs, sub, err := _DevicesStorage.contract.WatchLogs(opts, "Suicide")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DevicesStorageSuicide)
				if err := _DevicesStorage.contract.UnpackLog(event, "Suicide", log); err != nil {
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
