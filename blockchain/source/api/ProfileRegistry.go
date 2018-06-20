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

// ProfileRegistryABI is the input ABI used to generate the binding from.
const ProfileRegistryABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"certificates\",\"outputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"attributeType\",\"type\":\"uint256\"},{\"name\":\"value\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"validators\",\"outputs\":[{\"name\":\"\",\"type\":\"int8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"validator\",\"type\":\"address\"}],\"name\":\"ValidatorCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"validator\",\"type\":\"address\"}],\"name\":\"ValidatorDeleted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"CertificateCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"CertificateUpdated\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_validator\",\"type\":\"address\"},{\"name\":\"_level\",\"type\":\"int8\"}],\"name\":\"AddValidator\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_validator\",\"type\":\"address\"}],\"name\":\"RemoveValidator\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_validator\",\"type\":\"address\"}],\"name\":\"GetValidatorLevel\",\"outputs\":[{\"name\":\"\",\"type\":\"int8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_type\",\"type\":\"uint256\"},{\"name\":\"_value\",\"type\":\"bytes\"}],\"name\":\"CreateCertificate\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_id\",\"type\":\"uint256\"}],\"name\":\"RemoveCertificate\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_id\",\"type\":\"uint256\"}],\"name\":\"GetCertificate\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_type\",\"type\":\"uint256\"}],\"name\":\"GetAttributeValue\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_type\",\"type\":\"uint256\"}],\"name\":\"GetAttributeCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"GetProfileLevel\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// ProfileRegistryBin is the compiled bytecode used for deploying new contracts.
const ProfileRegistryBin = `0x60806040526000805534801561001457600080fd5b50336000908152600160205260409020805460ff191660ff179055610ec58061003e6000396000f3006080604052600436106100ad5763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166274fc3981146100b25780631af60f721461014b57806323eee3e6146101885780632eb4b2a5146101cd5780633e34e129146101e75780636209a633146102b0578063663b3e22146102e65780638997d27a146102fe57806393d7674214610338578063e7bcef44146103a1578063fa52c7d8146103c8575b600080fd5b3480156100be57600080fd5b506100d6600160a060020a03600435166024356103e9565b6040805160208082528351818301528351919283929083019185019080838360005b838110156101105781810151838201526020016100f8565b50505050905090810190601f16801561013d5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561015757600080fd5b5061016c600160a060020a036004351661049d565b60408051600160a060020a039092168252519081900360200190f35b34801561019457600080fd5b506101a9600160a060020a0360043516610520565b604051808260048111156101b957fe5b60ff16815260200191505060405180910390f35b3480156101d957600080fd5b506101e5600435610580565b005b3480156101f357600080fd5b506101ff600435610849565b6040518085600160a060020a0316600160a060020a0316815260200184600160a060020a0316600160a060020a0316815260200183815260200180602001828103825283818151815260200191508051906020019080838360005b8381101561027257818101518382015260200161025a565b50505050905090810190601f16801561029f5780820380516001836020036101000a031916815260200191505b509550505050505060405180910390f35b3480156102bc57600080fd5b506102d4600160a060020a036004351660243561091a565b60408051918252519081900360200190f35b3480156102f257600080fd5b506101ff600435610942565b34801561030a57600080fd5b5061031f600160a060020a0360043516610a01565b60408051600092830b90920b8252519081900360200190f35b34801561034457600080fd5b50604080516020600460443581810135601f81018490048402850184019095528484526101e5948235600160a060020a0316946024803595369594606494920191908190840183828082843750949750610a1e9650505050505050565b3480156103ad57600080fd5b5061016c600160a060020a036004351660243560000b610d2e565b3480156103d457600080fd5b5061031f600160a060020a0360043516610dc4565b600160a060020a038216600090815260036020908152604080832084845282529182902080548351601f60026000196101006001861615020190931692909204918201849004840281018401909452808452606093928301828280156104905780601f1061046557610100808354040283529160200191610490565b820191906000526020600020905b81548152906001019060200180831161047357829003601f168201915b5050505050905092915050565b60006104a833610a01565b60000b6000191415156104ba57600080fd5b60006104c583610a01565b60000b136104d257600080fd5b600160a060020a038216600081815260016020526040808220805460ff19169055517fa7a579573d398d7b67cd7450121bb250bbd060b29eabafdebc3ce0918658635c9190a250805b919050565b60008061052f836105796103e9565b51111561053e5750600461051b565b600061054c836105156103e9565b51111561055b5750600361051b565b6000610569836104b16103e9565b5111156105785750600261051b565b50600161051b565b610588610dd8565b60008281526002602081815260409283902083516080810185528154600160a060020a03908116825260018084015490911682850152828501548287015260038301805487516101009382161593909302600019011695909504601f8101859004850282018501909652858152909491936060860193919290918301828280156106535780601f1061062857610100808354040283529160200191610653565b820191906000526020600020905b81548152906001019060200180831161063657829003601f168201915b505050505081525050905033600160a060020a03168160200151600160a060020a0316148061068b57508051600160a060020a031633145b806106a2575061069a33610a01565b60000b600019145b15156106ad57600080fd5b604051606082015180517fc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a4709290819060208401908083835b602083106107045780518252601f1990920191602091820191016106e5565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040518091039020600019161415151561074257600080fd5b60208082018051600160a060020a03908116600090815260048085526040808320818801805185529087528184205486518616855283885282852082518652885282852060001990910190559451909316825284528181209251815291909252205415156107eb57604080516020818101808452600080845285830151600160a060020a031681526003835284812086860151825290925292902090516107e99290610dfe565b505b6040805160208181018084526000808452868152600290925292902090516108199260039092019190610dfe565b5060405182907f9a100d2018161ede6ca34c8007992b09bbffc636a636014a922e4c875041262890600090a25050565b60008181526002602081815260408084208054600180830154838701546003909401805486516101009482161594909402600019011697909704601f810187900487028301870190955284825287968796606096600160a060020a0395861696959093169493919283918301828280156109045780601f106108d957610100808354040283529160200191610904565b820191906000526020600020905b8154815290600101906020018083116108e757829003601f168201915b5050505050905093509350935093509193509193565b600160a060020a03919091166000908152600460209081526040808320938352929052205490565b60026020818152600092835260409283902080546001808301548386015460038501805489516101009582161595909502600019011697909704601f8101879004870284018701909852878352600160a060020a039384169793909116959094919290918301828280156109f75780601f106109cc576101008083540402835291602001916109f7565b820191906000526020600020905b8154815290600101906020018083116109da57829003601f168201915b5050505050905084565b600160a060020a0316600090815260016020526040812054900b90565b60008061044c8410610a5557600a60648504069150610a3c33610a01565b60000b8260000b13151515610a5057600080fd5b610a6a565b600160a060020a0385163314610a6a57600080fd5b60405183517fc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470918591819060208401908083835b60208310610abd5780518252601f199092019160209182019101610a9e565b6001836020036101000a03801982511681845116808217855250505050505090500191505060405180910390206000191614151515610afb57600080fd5b5060026103e884041480610c4157600160a060020a03851660009081526004602090815260408083208784529091529020541515610b6a57600160a060020a038516600090815260036020908152604080832087845282529091208451610b6492860190610dfe565b50610c41565b826040518082805190602001908083835b60208310610b9a5780518252601f199092019160209182019101610b7b565b6001836020036101000a038019825116818451168082178552505050505050905001915050604051809103902060001916610bd586866103e9565b6040518082805190602001908083835b60208310610c045780518252601f199092019160209182019101610be5565b6001836020036101000a038019825116818451168082178552505050505050905001915050604051809103902060001916141515610c4157600080fd5b600160a060020a0380861660008181526004602090815260408083208984528252808320805460019081019091558354810180855582516080810184523381528085019687528084018c8152606082018c815292875260028087529490962081518154908a1673ffffffffffffffffffffffffffffffffffffffff1991821617825597519381018054949099169390971692909217909655925190840155925180519193610cf792600385019290910190610dfe565b5050600080546040519092507fb9bb1df26fde5c1295a7ccd167330e5d6cb9df14fe4c3884669a64433cc9e7609190a25050505050565b6000610d3933610a01565b60000b600019141515610d4b57600080fd5b600082810b13610d5a57600080fd5b610d6383610a01565b60000b15610d7057600080fd5b600160a060020a038316600081815260016020526040808220805460ff191660ff87850b16179055517f02db26aafd16e8ecd93c4fa202917d50b1693f30b1594e57f7a432ede944eefc9190a25090919050565b600160205260009081526040812054900b81565b604080516080810182526000808252602082018190529181019190915260608082015290565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f10610e3f57805160ff1916838001178555610e6c565b82800160010185558215610e6c579182015b82811115610e6c578251825591602001919060010190610e51565b50610e78929150610e7c565b5090565b610e9691905b80821115610e785760008155600101610e82565b905600a165627a7a72305820ecf8be9eac07b9a3ff5215935ea42936a9a973181dfdbdd207c7319c201fe9200029`

// DeployProfileRegistry deploys a new Ethereum contract, binding an instance of ProfileRegistry to it.
func DeployProfileRegistry(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ProfileRegistry, error) {
	parsed, err := abi.JSON(strings.NewReader(ProfileRegistryABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ProfileRegistryBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ProfileRegistry{ProfileRegistryCaller: ProfileRegistryCaller{contract: contract}, ProfileRegistryTransactor: ProfileRegistryTransactor{contract: contract}, ProfileRegistryFilterer: ProfileRegistryFilterer{contract: contract}}, nil
}

// ProfileRegistry is an auto generated Go binding around an Ethereum contract.
type ProfileRegistry struct {
	ProfileRegistryCaller     // Read-only binding to the contract
	ProfileRegistryTransactor // Write-only binding to the contract
	ProfileRegistryFilterer   // Log filterer for contract events
}

// ProfileRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type ProfileRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProfileRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ProfileRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProfileRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ProfileRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProfileRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ProfileRegistrySession struct {
	Contract     *ProfileRegistry  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ProfileRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ProfileRegistryCallerSession struct {
	Contract *ProfileRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// ProfileRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ProfileRegistryTransactorSession struct {
	Contract     *ProfileRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// ProfileRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type ProfileRegistryRaw struct {
	Contract *ProfileRegistry // Generic contract binding to access the raw methods on
}

// ProfileRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ProfileRegistryCallerRaw struct {
	Contract *ProfileRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// ProfileRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ProfileRegistryTransactorRaw struct {
	Contract *ProfileRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewProfileRegistry creates a new instance of ProfileRegistry, bound to a specific deployed contract.
func NewProfileRegistry(address common.Address, backend bind.ContractBackend) (*ProfileRegistry, error) {
	contract, err := bindProfileRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ProfileRegistry{ProfileRegistryCaller: ProfileRegistryCaller{contract: contract}, ProfileRegistryTransactor: ProfileRegistryTransactor{contract: contract}, ProfileRegistryFilterer: ProfileRegistryFilterer{contract: contract}}, nil
}

// NewProfileRegistryCaller creates a new read-only instance of ProfileRegistry, bound to a specific deployed contract.
func NewProfileRegistryCaller(address common.Address, caller bind.ContractCaller) (*ProfileRegistryCaller, error) {
	contract, err := bindProfileRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ProfileRegistryCaller{contract: contract}, nil
}

// NewProfileRegistryTransactor creates a new write-only instance of ProfileRegistry, bound to a specific deployed contract.
func NewProfileRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*ProfileRegistryTransactor, error) {
	contract, err := bindProfileRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ProfileRegistryTransactor{contract: contract}, nil
}

// NewProfileRegistryFilterer creates a new log filterer instance of ProfileRegistry, bound to a specific deployed contract.
func NewProfileRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*ProfileRegistryFilterer, error) {
	contract, err := bindProfileRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ProfileRegistryFilterer{contract: contract}, nil
}

// bindProfileRegistry binds a generic wrapper to an already deployed contract.
func bindProfileRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ProfileRegistryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ProfileRegistry *ProfileRegistryRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ProfileRegistry.Contract.ProfileRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ProfileRegistry *ProfileRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.ProfileRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ProfileRegistry *ProfileRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.ProfileRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ProfileRegistry *ProfileRegistryCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ProfileRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ProfileRegistry *ProfileRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ProfileRegistry *ProfileRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.contract.Transact(opts, method, params...)
}

// GetAttributeCount is a free data retrieval call binding the contract method 0x6209a633.
//
// Solidity: function GetAttributeCount(_owner address, _type uint256) constant returns(uint256)
func (_ProfileRegistry *ProfileRegistryCaller) GetAttributeCount(opts *bind.CallOpts, _owner common.Address, _type *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ProfileRegistry.contract.Call(opts, out, "GetAttributeCount", _owner, _type)
	return *ret0, err
}

// GetAttributeCount is a free data retrieval call binding the contract method 0x6209a633.
//
// Solidity: function GetAttributeCount(_owner address, _type uint256) constant returns(uint256)
func (_ProfileRegistry *ProfileRegistrySession) GetAttributeCount(_owner common.Address, _type *big.Int) (*big.Int, error) {
	return _ProfileRegistry.Contract.GetAttributeCount(&_ProfileRegistry.CallOpts, _owner, _type)
}

// GetAttributeCount is a free data retrieval call binding the contract method 0x6209a633.
//
// Solidity: function GetAttributeCount(_owner address, _type uint256) constant returns(uint256)
func (_ProfileRegistry *ProfileRegistryCallerSession) GetAttributeCount(_owner common.Address, _type *big.Int) (*big.Int, error) {
	return _ProfileRegistry.Contract.GetAttributeCount(&_ProfileRegistry.CallOpts, _owner, _type)
}

// GetAttributeValue is a free data retrieval call binding the contract method 0x0074fc39.
//
// Solidity: function GetAttributeValue(_owner address, _type uint256) constant returns(bytes)
func (_ProfileRegistry *ProfileRegistryCaller) GetAttributeValue(opts *bind.CallOpts, _owner common.Address, _type *big.Int) ([]byte, error) {
	var (
		ret0 = new([]byte)
	)
	out := ret0
	err := _ProfileRegistry.contract.Call(opts, out, "GetAttributeValue", _owner, _type)
	return *ret0, err
}

// GetAttributeValue is a free data retrieval call binding the contract method 0x0074fc39.
//
// Solidity: function GetAttributeValue(_owner address, _type uint256) constant returns(bytes)
func (_ProfileRegistry *ProfileRegistrySession) GetAttributeValue(_owner common.Address, _type *big.Int) ([]byte, error) {
	return _ProfileRegistry.Contract.GetAttributeValue(&_ProfileRegistry.CallOpts, _owner, _type)
}

// GetAttributeValue is a free data retrieval call binding the contract method 0x0074fc39.
//
// Solidity: function GetAttributeValue(_owner address, _type uint256) constant returns(bytes)
func (_ProfileRegistry *ProfileRegistryCallerSession) GetAttributeValue(_owner common.Address, _type *big.Int) ([]byte, error) {
	return _ProfileRegistry.Contract.GetAttributeValue(&_ProfileRegistry.CallOpts, _owner, _type)
}

// GetCertificate is a free data retrieval call binding the contract method 0x3e34e129.
//
// Solidity: function GetCertificate(_id uint256) constant returns(address, address, uint256, bytes)
func (_ProfileRegistry *ProfileRegistryCaller) GetCertificate(opts *bind.CallOpts, _id *big.Int) (common.Address, common.Address, *big.Int, []byte, error) {
	var (
		ret0 = new(common.Address)
		ret1 = new(common.Address)
		ret2 = new(*big.Int)
		ret3 = new([]byte)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
	}
	err := _ProfileRegistry.contract.Call(opts, out, "GetCertificate", _id)
	return *ret0, *ret1, *ret2, *ret3, err
}

// GetCertificate is a free data retrieval call binding the contract method 0x3e34e129.
//
// Solidity: function GetCertificate(_id uint256) constant returns(address, address, uint256, bytes)
func (_ProfileRegistry *ProfileRegistrySession) GetCertificate(_id *big.Int) (common.Address, common.Address, *big.Int, []byte, error) {
	return _ProfileRegistry.Contract.GetCertificate(&_ProfileRegistry.CallOpts, _id)
}

// GetCertificate is a free data retrieval call binding the contract method 0x3e34e129.
//
// Solidity: function GetCertificate(_id uint256) constant returns(address, address, uint256, bytes)
func (_ProfileRegistry *ProfileRegistryCallerSession) GetCertificate(_id *big.Int) (common.Address, common.Address, *big.Int, []byte, error) {
	return _ProfileRegistry.Contract.GetCertificate(&_ProfileRegistry.CallOpts, _id)
}

// GetProfileLevel is a free data retrieval call binding the contract method 0x23eee3e6.
//
// Solidity: function GetProfileLevel(_owner address) constant returns(uint8)
func (_ProfileRegistry *ProfileRegistryCaller) GetProfileLevel(opts *bind.CallOpts, _owner common.Address) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _ProfileRegistry.contract.Call(opts, out, "GetProfileLevel", _owner)
	return *ret0, err
}

// GetProfileLevel is a free data retrieval call binding the contract method 0x23eee3e6.
//
// Solidity: function GetProfileLevel(_owner address) constant returns(uint8)
func (_ProfileRegistry *ProfileRegistrySession) GetProfileLevel(_owner common.Address) (uint8, error) {
	return _ProfileRegistry.Contract.GetProfileLevel(&_ProfileRegistry.CallOpts, _owner)
}

// GetProfileLevel is a free data retrieval call binding the contract method 0x23eee3e6.
//
// Solidity: function GetProfileLevel(_owner address) constant returns(uint8)
func (_ProfileRegistry *ProfileRegistryCallerSession) GetProfileLevel(_owner common.Address) (uint8, error) {
	return _ProfileRegistry.Contract.GetProfileLevel(&_ProfileRegistry.CallOpts, _owner)
}

// GetValidatorLevel is a free data retrieval call binding the contract method 0x8997d27a.
//
// Solidity: function GetValidatorLevel(_validator address) constant returns(int8)
func (_ProfileRegistry *ProfileRegistryCaller) GetValidatorLevel(opts *bind.CallOpts, _validator common.Address) (int8, error) {
	var (
		ret0 = new(int8)
	)
	out := ret0
	err := _ProfileRegistry.contract.Call(opts, out, "GetValidatorLevel", _validator)
	return *ret0, err
}

// GetValidatorLevel is a free data retrieval call binding the contract method 0x8997d27a.
//
// Solidity: function GetValidatorLevel(_validator address) constant returns(int8)
func (_ProfileRegistry *ProfileRegistrySession) GetValidatorLevel(_validator common.Address) (int8, error) {
	return _ProfileRegistry.Contract.GetValidatorLevel(&_ProfileRegistry.CallOpts, _validator)
}

// GetValidatorLevel is a free data retrieval call binding the contract method 0x8997d27a.
//
// Solidity: function GetValidatorLevel(_validator address) constant returns(int8)
func (_ProfileRegistry *ProfileRegistryCallerSession) GetValidatorLevel(_validator common.Address) (int8, error) {
	return _ProfileRegistry.Contract.GetValidatorLevel(&_ProfileRegistry.CallOpts, _validator)
}

// Certificates is a free data retrieval call binding the contract method 0x663b3e22.
//
// Solidity: function certificates( uint256) constant returns(from address, to address, attributeType uint256, value bytes)
func (_ProfileRegistry *ProfileRegistryCaller) Certificates(opts *bind.CallOpts, arg0 *big.Int) (struct {
	From          common.Address
	To            common.Address
	AttributeType *big.Int
	Value         []byte
}, error) {
	ret := new(struct {
		From          common.Address
		To            common.Address
		AttributeType *big.Int
		Value         []byte
	})
	out := ret
	err := _ProfileRegistry.contract.Call(opts, out, "certificates", arg0)
	return *ret, err
}

// Certificates is a free data retrieval call binding the contract method 0x663b3e22.
//
// Solidity: function certificates( uint256) constant returns(from address, to address, attributeType uint256, value bytes)
func (_ProfileRegistry *ProfileRegistrySession) Certificates(arg0 *big.Int) (struct {
	From          common.Address
	To            common.Address
	AttributeType *big.Int
	Value         []byte
}, error) {
	return _ProfileRegistry.Contract.Certificates(&_ProfileRegistry.CallOpts, arg0)
}

// Certificates is a free data retrieval call binding the contract method 0x663b3e22.
//
// Solidity: function certificates( uint256) constant returns(from address, to address, attributeType uint256, value bytes)
func (_ProfileRegistry *ProfileRegistryCallerSession) Certificates(arg0 *big.Int) (struct {
	From          common.Address
	To            common.Address
	AttributeType *big.Int
	Value         []byte
}, error) {
	return _ProfileRegistry.Contract.Certificates(&_ProfileRegistry.CallOpts, arg0)
}

// Validators is a free data retrieval call binding the contract method 0xfa52c7d8.
//
// Solidity: function validators( address) constant returns(int8)
func (_ProfileRegistry *ProfileRegistryCaller) Validators(opts *bind.CallOpts, arg0 common.Address) (int8, error) {
	var (
		ret0 = new(int8)
	)
	out := ret0
	err := _ProfileRegistry.contract.Call(opts, out, "validators", arg0)
	return *ret0, err
}

// Validators is a free data retrieval call binding the contract method 0xfa52c7d8.
//
// Solidity: function validators( address) constant returns(int8)
func (_ProfileRegistry *ProfileRegistrySession) Validators(arg0 common.Address) (int8, error) {
	return _ProfileRegistry.Contract.Validators(&_ProfileRegistry.CallOpts, arg0)
}

// Validators is a free data retrieval call binding the contract method 0xfa52c7d8.
//
// Solidity: function validators( address) constant returns(int8)
func (_ProfileRegistry *ProfileRegistryCallerSession) Validators(arg0 common.Address) (int8, error) {
	return _ProfileRegistry.Contract.Validators(&_ProfileRegistry.CallOpts, arg0)
}

// AddValidator is a paid mutator transaction binding the contract method 0xe7bcef44.
//
// Solidity: function AddValidator(_validator address, _level int8) returns(address)
func (_ProfileRegistry *ProfileRegistryTransactor) AddValidator(opts *bind.TransactOpts, _validator common.Address, _level int8) (*types.Transaction, error) {
	return _ProfileRegistry.contract.Transact(opts, "AddValidator", _validator, _level)
}

// AddValidator is a paid mutator transaction binding the contract method 0xe7bcef44.
//
// Solidity: function AddValidator(_validator address, _level int8) returns(address)
func (_ProfileRegistry *ProfileRegistrySession) AddValidator(_validator common.Address, _level int8) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.AddValidator(&_ProfileRegistry.TransactOpts, _validator, _level)
}

// AddValidator is a paid mutator transaction binding the contract method 0xe7bcef44.
//
// Solidity: function AddValidator(_validator address, _level int8) returns(address)
func (_ProfileRegistry *ProfileRegistryTransactorSession) AddValidator(_validator common.Address, _level int8) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.AddValidator(&_ProfileRegistry.TransactOpts, _validator, _level)
}

// CreateCertificate is a paid mutator transaction binding the contract method 0x93d76742.
//
// Solidity: function CreateCertificate(_owner address, _type uint256, _value bytes) returns()
func (_ProfileRegistry *ProfileRegistryTransactor) CreateCertificate(opts *bind.TransactOpts, _owner common.Address, _type *big.Int, _value []byte) (*types.Transaction, error) {
	return _ProfileRegistry.contract.Transact(opts, "CreateCertificate", _owner, _type, _value)
}

// CreateCertificate is a paid mutator transaction binding the contract method 0x93d76742.
//
// Solidity: function CreateCertificate(_owner address, _type uint256, _value bytes) returns()
func (_ProfileRegistry *ProfileRegistrySession) CreateCertificate(_owner common.Address, _type *big.Int, _value []byte) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.CreateCertificate(&_ProfileRegistry.TransactOpts, _owner, _type, _value)
}

// CreateCertificate is a paid mutator transaction binding the contract method 0x93d76742.
//
// Solidity: function CreateCertificate(_owner address, _type uint256, _value bytes) returns()
func (_ProfileRegistry *ProfileRegistryTransactorSession) CreateCertificate(_owner common.Address, _type *big.Int, _value []byte) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.CreateCertificate(&_ProfileRegistry.TransactOpts, _owner, _type, _value)
}

// RemoveCertificate is a paid mutator transaction binding the contract method 0x2eb4b2a5.
//
// Solidity: function RemoveCertificate(_id uint256) returns()
func (_ProfileRegistry *ProfileRegistryTransactor) RemoveCertificate(opts *bind.TransactOpts, _id *big.Int) (*types.Transaction, error) {
	return _ProfileRegistry.contract.Transact(opts, "RemoveCertificate", _id)
}

// RemoveCertificate is a paid mutator transaction binding the contract method 0x2eb4b2a5.
//
// Solidity: function RemoveCertificate(_id uint256) returns()
func (_ProfileRegistry *ProfileRegistrySession) RemoveCertificate(_id *big.Int) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.RemoveCertificate(&_ProfileRegistry.TransactOpts, _id)
}

// RemoveCertificate is a paid mutator transaction binding the contract method 0x2eb4b2a5.
//
// Solidity: function RemoveCertificate(_id uint256) returns()
func (_ProfileRegistry *ProfileRegistryTransactorSession) RemoveCertificate(_id *big.Int) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.RemoveCertificate(&_ProfileRegistry.TransactOpts, _id)
}

// RemoveValidator is a paid mutator transaction binding the contract method 0x1af60f72.
//
// Solidity: function RemoveValidator(_validator address) returns(address)
func (_ProfileRegistry *ProfileRegistryTransactor) RemoveValidator(opts *bind.TransactOpts, _validator common.Address) (*types.Transaction, error) {
	return _ProfileRegistry.contract.Transact(opts, "RemoveValidator", _validator)
}

// RemoveValidator is a paid mutator transaction binding the contract method 0x1af60f72.
//
// Solidity: function RemoveValidator(_validator address) returns(address)
func (_ProfileRegistry *ProfileRegistrySession) RemoveValidator(_validator common.Address) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.RemoveValidator(&_ProfileRegistry.TransactOpts, _validator)
}

// RemoveValidator is a paid mutator transaction binding the contract method 0x1af60f72.
//
// Solidity: function RemoveValidator(_validator address) returns(address)
func (_ProfileRegistry *ProfileRegistryTransactorSession) RemoveValidator(_validator common.Address) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.RemoveValidator(&_ProfileRegistry.TransactOpts, _validator)
}

// ProfileRegistryCertificateCreatedIterator is returned from FilterCertificateCreated and is used to iterate over the raw logs and unpacked data for CertificateCreated events raised by the ProfileRegistry contract.
type ProfileRegistryCertificateCreatedIterator struct {
	Event *ProfileRegistryCertificateCreated // Event containing the contract specifics and raw log

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
func (it *ProfileRegistryCertificateCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProfileRegistryCertificateCreated)
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
		it.Event = new(ProfileRegistryCertificateCreated)
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
func (it *ProfileRegistryCertificateCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProfileRegistryCertificateCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProfileRegistryCertificateCreated represents a CertificateCreated event raised by the ProfileRegistry contract.
type ProfileRegistryCertificateCreated struct {
	Id  *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterCertificateCreated is a free log retrieval operation binding the contract event 0xb9bb1df26fde5c1295a7ccd167330e5d6cb9df14fe4c3884669a64433cc9e760.
//
// Solidity: event CertificateCreated(id indexed uint256)
func (_ProfileRegistry *ProfileRegistryFilterer) FilterCertificateCreated(opts *bind.FilterOpts, id []*big.Int) (*ProfileRegistryCertificateCreatedIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _ProfileRegistry.contract.FilterLogs(opts, "CertificateCreated", idRule)
	if err != nil {
		return nil, err
	}
	return &ProfileRegistryCertificateCreatedIterator{contract: _ProfileRegistry.contract, event: "CertificateCreated", logs: logs, sub: sub}, nil
}

// WatchCertificateCreated is a free log subscription operation binding the contract event 0xb9bb1df26fde5c1295a7ccd167330e5d6cb9df14fe4c3884669a64433cc9e760.
//
// Solidity: event CertificateCreated(id indexed uint256)
func (_ProfileRegistry *ProfileRegistryFilterer) WatchCertificateCreated(opts *bind.WatchOpts, sink chan<- *ProfileRegistryCertificateCreated, id []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _ProfileRegistry.contract.WatchLogs(opts, "CertificateCreated", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProfileRegistryCertificateCreated)
				if err := _ProfileRegistry.contract.UnpackLog(event, "CertificateCreated", log); err != nil {
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

// ProfileRegistryCertificateUpdatedIterator is returned from FilterCertificateUpdated and is used to iterate over the raw logs and unpacked data for CertificateUpdated events raised by the ProfileRegistry contract.
type ProfileRegistryCertificateUpdatedIterator struct {
	Event *ProfileRegistryCertificateUpdated // Event containing the contract specifics and raw log

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
func (it *ProfileRegistryCertificateUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProfileRegistryCertificateUpdated)
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
		it.Event = new(ProfileRegistryCertificateUpdated)
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
func (it *ProfileRegistryCertificateUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProfileRegistryCertificateUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProfileRegistryCertificateUpdated represents a CertificateUpdated event raised by the ProfileRegistry contract.
type ProfileRegistryCertificateUpdated struct {
	Id  *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterCertificateUpdated is a free log retrieval operation binding the contract event 0x9a100d2018161ede6ca34c8007992b09bbffc636a636014a922e4c8750412628.
//
// Solidity: event CertificateUpdated(id indexed uint256)
func (_ProfileRegistry *ProfileRegistryFilterer) FilterCertificateUpdated(opts *bind.FilterOpts, id []*big.Int) (*ProfileRegistryCertificateUpdatedIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _ProfileRegistry.contract.FilterLogs(opts, "CertificateUpdated", idRule)
	if err != nil {
		return nil, err
	}
	return &ProfileRegistryCertificateUpdatedIterator{contract: _ProfileRegistry.contract, event: "CertificateUpdated", logs: logs, sub: sub}, nil
}

// WatchCertificateUpdated is a free log subscription operation binding the contract event 0x9a100d2018161ede6ca34c8007992b09bbffc636a636014a922e4c8750412628.
//
// Solidity: event CertificateUpdated(id indexed uint256)
func (_ProfileRegistry *ProfileRegistryFilterer) WatchCertificateUpdated(opts *bind.WatchOpts, sink chan<- *ProfileRegistryCertificateUpdated, id []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _ProfileRegistry.contract.WatchLogs(opts, "CertificateUpdated", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProfileRegistryCertificateUpdated)
				if err := _ProfileRegistry.contract.UnpackLog(event, "CertificateUpdated", log); err != nil {
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

// ProfileRegistryValidatorCreatedIterator is returned from FilterValidatorCreated and is used to iterate over the raw logs and unpacked data for ValidatorCreated events raised by the ProfileRegistry contract.
type ProfileRegistryValidatorCreatedIterator struct {
	Event *ProfileRegistryValidatorCreated // Event containing the contract specifics and raw log

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
func (it *ProfileRegistryValidatorCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProfileRegistryValidatorCreated)
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
		it.Event = new(ProfileRegistryValidatorCreated)
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
func (it *ProfileRegistryValidatorCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProfileRegistryValidatorCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProfileRegistryValidatorCreated represents a ValidatorCreated event raised by the ProfileRegistry contract.
type ProfileRegistryValidatorCreated struct {
	Validator common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterValidatorCreated is a free log retrieval operation binding the contract event 0x02db26aafd16e8ecd93c4fa202917d50b1693f30b1594e57f7a432ede944eefc.
//
// Solidity: event ValidatorCreated(validator indexed address)
func (_ProfileRegistry *ProfileRegistryFilterer) FilterValidatorCreated(opts *bind.FilterOpts, validator []common.Address) (*ProfileRegistryValidatorCreatedIterator, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _ProfileRegistry.contract.FilterLogs(opts, "ValidatorCreated", validatorRule)
	if err != nil {
		return nil, err
	}
	return &ProfileRegistryValidatorCreatedIterator{contract: _ProfileRegistry.contract, event: "ValidatorCreated", logs: logs, sub: sub}, nil
}

// WatchValidatorCreated is a free log subscription operation binding the contract event 0x02db26aafd16e8ecd93c4fa202917d50b1693f30b1594e57f7a432ede944eefc.
//
// Solidity: event ValidatorCreated(validator indexed address)
func (_ProfileRegistry *ProfileRegistryFilterer) WatchValidatorCreated(opts *bind.WatchOpts, sink chan<- *ProfileRegistryValidatorCreated, validator []common.Address) (event.Subscription, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _ProfileRegistry.contract.WatchLogs(opts, "ValidatorCreated", validatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProfileRegistryValidatorCreated)
				if err := _ProfileRegistry.contract.UnpackLog(event, "ValidatorCreated", log); err != nil {
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

// ProfileRegistryValidatorDeletedIterator is returned from FilterValidatorDeleted and is used to iterate over the raw logs and unpacked data for ValidatorDeleted events raised by the ProfileRegistry contract.
type ProfileRegistryValidatorDeletedIterator struct {
	Event *ProfileRegistryValidatorDeleted // Event containing the contract specifics and raw log

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
func (it *ProfileRegistryValidatorDeletedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProfileRegistryValidatorDeleted)
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
		it.Event = new(ProfileRegistryValidatorDeleted)
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
func (it *ProfileRegistryValidatorDeletedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProfileRegistryValidatorDeletedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProfileRegistryValidatorDeleted represents a ValidatorDeleted event raised by the ProfileRegistry contract.
type ProfileRegistryValidatorDeleted struct {
	Validator common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterValidatorDeleted is a free log retrieval operation binding the contract event 0xa7a579573d398d7b67cd7450121bb250bbd060b29eabafdebc3ce0918658635c.
//
// Solidity: event ValidatorDeleted(validator indexed address)
func (_ProfileRegistry *ProfileRegistryFilterer) FilterValidatorDeleted(opts *bind.FilterOpts, validator []common.Address) (*ProfileRegistryValidatorDeletedIterator, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _ProfileRegistry.contract.FilterLogs(opts, "ValidatorDeleted", validatorRule)
	if err != nil {
		return nil, err
	}
	return &ProfileRegistryValidatorDeletedIterator{contract: _ProfileRegistry.contract, event: "ValidatorDeleted", logs: logs, sub: sub}, nil
}

// WatchValidatorDeleted is a free log subscription operation binding the contract event 0xa7a579573d398d7b67cd7450121bb250bbd060b29eabafdebc3ce0918658635c.
//
// Solidity: event ValidatorDeleted(validator indexed address)
func (_ProfileRegistry *ProfileRegistryFilterer) WatchValidatorDeleted(opts *bind.WatchOpts, sink chan<- *ProfileRegistryValidatorDeleted, validator []common.Address) (event.Subscription, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _ProfileRegistry.contract.WatchLogs(opts, "ValidatorDeleted", validatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProfileRegistryValidatorDeleted)
				if err := _ProfileRegistry.contract.UnpackLog(event, "ValidatorDeleted", log); err != nil {
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
