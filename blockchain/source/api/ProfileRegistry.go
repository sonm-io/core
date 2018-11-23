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
const ProfileRegistryABI = "[{\"constant\":false,\"inputs\":[],\"name\":\"unpause\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"paused\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"certificates\",\"outputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"attributeType\",\"type\":\"uint256\"},{\"name\":\"value\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"pause\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"validators\",\"outputs\":[{\"name\":\"\",\"type\":\"int8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"validator\",\"type\":\"address\"}],\"name\":\"ValidatorCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"validator\",\"type\":\"address\"}],\"name\":\"ValidatorDeleted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"CertificateCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"CertificateUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"Pause\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"Unpause\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"}],\"name\":\"OwnershipRenounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_validator\",\"type\":\"address\"},{\"name\":\"_level\",\"type\":\"int8\"}],\"name\":\"AddValidator\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_validator\",\"type\":\"address\"}],\"name\":\"RemoveValidator\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_validator\",\"type\":\"address\"}],\"name\":\"GetValidatorLevel\",\"outputs\":[{\"name\":\"\",\"type\":\"int8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_type\",\"type\":\"uint256\"},{\"name\":\"_value\",\"type\":\"bytes\"}],\"name\":\"CreateCertificate\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_id\",\"type\":\"uint256\"}],\"name\":\"RemoveCertificate\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_id\",\"type\":\"uint256\"}],\"name\":\"GetCertificate\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_type\",\"type\":\"uint256\"}],\"name\":\"GetAttributeValue\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_type\",\"type\":\"uint256\"}],\"name\":\"GetAttributeCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"GetProfileLevel\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_validator\",\"type\":\"address\"}],\"name\":\"AddSonmValidator\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_validator\",\"type\":\"address\"}],\"name\":\"RemoveSonmValidator\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// ProfileRegistryBin is the compiled bytecode used for deploying new contracts.
const ProfileRegistryBin = `0x60806040526000805460a060020a60ff021916815560015534801561002357600080fd5b5060008054600160a060020a031990811633908117909116811782558152600260205260409020805460ff191660ff179055611311806100646000396000f3006080604052600436106101045763ffffffff7c0100000000000000000000000000000000000000000000000000000000600035041662707c75811461010957806274fc391461013e5780630553701a146101d75780631af60f72146101f857806323eee3e6146102355780632eb4b2a51461027a5780633e34e129146102945780633f4ba83a1461035d5780635c975abb146103725780636209a63314610387578063663b3e22146103bd578063715018a6146103d55780638456cb59146103ea5780638997d27a146103ff5780638da5cb5b1461043957806393d767421461044e578063e7bcef44146104b7578063f2fde38b146104de578063fa52c7d8146104ff575b600080fd5b34801561011557600080fd5b5061012a600160a060020a0360043516610520565b604080519115158252519081900360200190f35b34801561014a57600080fd5b50610162600160a060020a0360043516602435610564565b6040805160208082528351818301528351919283929083019185019080838360005b8381101561019c578181015183820152602001610184565b50505050905090810190601f1680156101c95780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b3480156101e357600080fd5b5061012a600160a060020a0360043516610618565b34801561020457600080fd5b50610219600160a060020a0360043516610670565b60408051600160a060020a039092168252519081900360200190f35b34801561024157600080fd5b50610256600160a060020a0360043516610706565b6040518082600481111561026657fe5b60ff16815260200191505060405180910390f35b34801561028657600080fd5b50610292600435610766565b005b3480156102a057600080fd5b506102ac600435610a48565b6040518085600160a060020a0316600160a060020a0316815260200184600160a060020a0316600160a060020a0316815260200183815260200180602001828103825283818151815260200191508051906020019080838360005b8381101561031f578181015183820152602001610307565b50505050905090810190601f16801561034c5780820380516001836020036101000a031916815260200191505b509550505050505060405180910390f35b34801561036957600080fd5b50610292610b1a565b34801561037e57600080fd5b5061012a610b90565b34801561039357600080fd5b506103ab600160a060020a0360043516602435610ba0565b60408051918252519081900360200190f35b3480156103c957600080fd5b506102ac600435610bc8565b3480156103e157600080fd5b50610292610c87565b3480156103f657600080fd5b50610292610cf3565b34801561040b57600080fd5b50610420600160a060020a0360043516610d6e565b60408051600092830b90920b8252519081900360200190f35b34801561044557600080fd5b50610219610d8b565b34801561045a57600080fd5b50604080516020600460443581810135601f8101849004840285018401909552848452610292948235600160a060020a0316946024803595369594606494920191908190840183828082843750949750610d9a9650505050505050565b3480156104c357600080fd5b50610219600160a060020a036004351660243560000b6110c3565b3480156104ea57600080fd5b50610292600160a060020a0360043516611170565b34801561050b57600080fd5b50610420600160a060020a0360043516611193565b60008054600160a060020a0316331461053857600080fd5b50600160a060020a0381166000908152600260205260409020805460ff191660ff17905560015b919050565b600160a060020a038216600090815260046020908152604080832084845282529182902080548351601f600260001961010060018616150201909316929092049182018490048402810184019094528084526060939283018282801561060b5780601f106105e05761010080835404028352916020019161060b565b820191906000526020600020905b8154815290600101906020018083116105ee57829003601f168201915b5050505050905092915050565b60008054600160a060020a0316331461063057600080fd5b61063982610d6e565b60000b60001914151561064b57600080fd5b50600160a060020a03166000908152600260205260409020805460ff19169055600190565b600061067b33610d6e565b60000b60001914151561068d57600080fd5b60005460a060020a900460ff16156106a457600080fd5b60006106af83610d6e565b60000b136106bc57600080fd5b600160a060020a038216600081815260026020526040808220805460ff19169055517fa7a579573d398d7b67cd7450121bb250bbd060b29eabafdebc3ce0918658635c9190a25090565b60008061071583610579610564565b5111156107245750600461055f565b600061073283610515610564565b5111156107415750600361055f565b600061074f836104b1610564565b51111561075e5750600261055f565b50600161055f565b61076e611224565b60005460a060020a900460ff161561078557600080fd5b60008281526003602081815260409283902083516080810185528154600160a060020a0390811682526001808401549091168285015260028084015483880152948301805487516101009382161593909302600019011695909504601f8101859004850282018501909652858152909491936060860193919290918301828280156108515780601f1061082657610100808354040283529160200191610851565b820191906000526020600020905b81548152906001019060200180831161083457829003601f168201915b505050505081525050905033600160a060020a03168160200151600160a060020a0316148061088957508051600160a060020a031633145b806108a0575061089833610d6e565b60000b600019145b15156108ab57600080fd5b604051606082015180517fc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a4709290819060208401908083835b602083106109025780518252601f1990920191602091820191016108e3565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040518091039020600019161415151561094057600080fd5b60208082018051600160a060020a03908116600090815260058085526040808320818801805185529087528184205486518616855283885282852082518652885282852060001990910190559451909316825284528181209251815291909252205415156109e957604080516020818101808452600080845285830151600160a060020a031681526004835284812086860151825290925292902090516109e7929061124a565b505b6040805160208181018084526000808452868152600392839052939093209151610a189392909101919061124a565b5060405182907f9a100d2018161ede6ca34c8007992b09bbffc636a636014a922e4c875041262890600090a25050565b6000818152600360208181526040808420805460018083015460028085015494909701805486516101009482161594909402600019011697909704601f810187900487028301870190955284825287968796606096600160a060020a039586169695909316949391928391830182828015610b045780601f10610ad957610100808354040283529160200191610b04565b820191906000526020600020905b815481529060010190602001808311610ae757829003601f168201915b5050505050905093509350935093509193509193565b600054600160a060020a03163314610b3157600080fd5b60005460a060020a900460ff161515610b4957600080fd5b6000805474ff0000000000000000000000000000000000000000191681556040517f7805862f689e2f13df9f062ff482ad3ad112aca9e0847911ed832e158c525b339190a1565b60005460a060020a900460ff1681565b600160a060020a03919091166000908152600560209081526040808320938352929052205490565b6003602081815260009283526040928390208054600180830154600280850154968501805489516101009582161595909502600019011691909104601f8101879004870284018701909852878352600160a060020a039384169793909116959491929091830182828015610c7d5780601f10610c5257610100808354040283529160200191610c7d565b820191906000526020600020905b815481529060010190602001808311610c6057829003601f168201915b5050505050905084565b600054600160a060020a03163314610c9e57600080fd5b60008054604051600160a060020a03909116917ff8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c6482091a26000805473ffffffffffffffffffffffffffffffffffffffff19169055565b600054600160a060020a03163314610d0a57600080fd5b60005460a060020a900460ff1615610d2157600080fd5b6000805474ff0000000000000000000000000000000000000000191660a060020a1781556040517f6985a02210a168e66602d3235cb6db0e70f92b3ba4d376a33c0f3d9434bff6259190a1565b600160a060020a0316600090815260026020526040812054900b90565b600054600160a060020a031681565b60008054819060a060020a900460ff1615610db457600080fd5b61044c8410610de857600a60648504069150610dcf33610d6e565b60000b8260000b13151515610de357600080fd5b610dfd565b600160a060020a0385163314610dfd57600080fd5b60405183517fc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470918591819060208401908083835b60208310610e505780518252601f199092019160209182019101610e31565b6001836020036101000a03801982511681845116808217855250505050505090500191505060405180910390206000191614151515610e8e57600080fd5b5060026103e884041480610fd457600160a060020a03851660009081526005602090815260408083208784529091529020541515610efd57600160a060020a038516600090815260046020908152604080832087845282529091208451610ef79286019061124a565b50610fd4565b826040518082805190602001908083835b60208310610f2d5780518252601f199092019160209182019101610f0e565b6001836020036101000a038019825116818451168082178552505050505050905001915050604051809103902060001916610f688686610564565b6040518082805190602001908083835b60208310610f975780518252601f199092019160209182019101610f78565b6001836020036101000a038019825116818451168082178552505050505050905001915050604051809103902060001916141515610fd457600080fd5b600160a060020a0380861660008181526005602090815260408083208984528252808320805460019081019091558054810180825582516080810184523381528085019687528084018c8152606082018c815292875260038087529490962081518154908a1673ffffffffffffffffffffffffffffffffffffffff19918216178255975193810180549490991693909716929092179096559251600285015593518051929461108b9390850192919091019061124a565b50506001546040519091507fb9bb1df26fde5c1295a7ccd167330e5d6cb9df14fe4c3884669a64433cc9e76090600090a25050505050565b60006110ce33610d6e565b60000b6000191415156110e057600080fd5b60005460a060020a900460ff16156110f757600080fd5b600082810b1361110657600080fd5b61110f83610d6e565b60000b1561111c57600080fd5b600160a060020a038316600081815260026020526040808220805460ff191660ff87850b16179055517f02db26aafd16e8ecd93c4fa202917d50b1693f30b1594e57f7a432ede944eefc9190a25090919050565b600054600160a060020a0316331461118757600080fd5b611190816111a7565b50565b600260205260009081526040812054900b81565b600160a060020a03811615156111bc57600080fd5b60008054604051600160a060020a03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a36000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0392909216919091179055565b604080516080810182526000808252602082018190529181019190915260608082015290565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061128b57805160ff19168380011785556112b8565b828001600101855582156112b8579182015b828111156112b857825182559160200191906001019061129d565b506112c49291506112c8565b5090565b6112e291905b808211156112c457600081556001016112ce565b905600a165627a7a7230582062063412dc64c0ee657ce05f8e88f1f13a7c8937a61cd4f42db5d324754e10750029`

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

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_ProfileRegistry *ProfileRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _ProfileRegistry.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_ProfileRegistry *ProfileRegistrySession) Owner() (common.Address, error) {
	return _ProfileRegistry.Contract.Owner(&_ProfileRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_ProfileRegistry *ProfileRegistryCallerSession) Owner() (common.Address, error) {
	return _ProfileRegistry.Contract.Owner(&_ProfileRegistry.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() constant returns(bool)
func (_ProfileRegistry *ProfileRegistryCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _ProfileRegistry.contract.Call(opts, out, "paused")
	return *ret0, err
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() constant returns(bool)
func (_ProfileRegistry *ProfileRegistrySession) Paused() (bool, error) {
	return _ProfileRegistry.Contract.Paused(&_ProfileRegistry.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() constant returns(bool)
func (_ProfileRegistry *ProfileRegistryCallerSession) Paused() (bool, error) {
	return _ProfileRegistry.Contract.Paused(&_ProfileRegistry.CallOpts)
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

// AddSonmValidator is a paid mutator transaction binding the contract method 0x00707c75.
//
// Solidity: function AddSonmValidator(_validator address) returns(bool)
func (_ProfileRegistry *ProfileRegistryTransactor) AddSonmValidator(opts *bind.TransactOpts, _validator common.Address) (*types.Transaction, error) {
	return _ProfileRegistry.contract.Transact(opts, "AddSonmValidator", _validator)
}

// AddSonmValidator is a paid mutator transaction binding the contract method 0x00707c75.
//
// Solidity: function AddSonmValidator(_validator address) returns(bool)
func (_ProfileRegistry *ProfileRegistrySession) AddSonmValidator(_validator common.Address) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.AddSonmValidator(&_ProfileRegistry.TransactOpts, _validator)
}

// AddSonmValidator is a paid mutator transaction binding the contract method 0x00707c75.
//
// Solidity: function AddSonmValidator(_validator address) returns(bool)
func (_ProfileRegistry *ProfileRegistryTransactorSession) AddSonmValidator(_validator common.Address) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.AddSonmValidator(&_ProfileRegistry.TransactOpts, _validator)
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

// RemoveSonmValidator is a paid mutator transaction binding the contract method 0x0553701a.
//
// Solidity: function RemoveSonmValidator(_validator address) returns(bool)
func (_ProfileRegistry *ProfileRegistryTransactor) RemoveSonmValidator(opts *bind.TransactOpts, _validator common.Address) (*types.Transaction, error) {
	return _ProfileRegistry.contract.Transact(opts, "RemoveSonmValidator", _validator)
}

// RemoveSonmValidator is a paid mutator transaction binding the contract method 0x0553701a.
//
// Solidity: function RemoveSonmValidator(_validator address) returns(bool)
func (_ProfileRegistry *ProfileRegistrySession) RemoveSonmValidator(_validator common.Address) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.RemoveSonmValidator(&_ProfileRegistry.TransactOpts, _validator)
}

// RemoveSonmValidator is a paid mutator transaction binding the contract method 0x0553701a.
//
// Solidity: function RemoveSonmValidator(_validator address) returns(bool)
func (_ProfileRegistry *ProfileRegistryTransactorSession) RemoveSonmValidator(_validator common.Address) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.RemoveSonmValidator(&_ProfileRegistry.TransactOpts, _validator)
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

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_ProfileRegistry *ProfileRegistryTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProfileRegistry.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_ProfileRegistry *ProfileRegistrySession) Pause() (*types.Transaction, error) {
	return _ProfileRegistry.Contract.Pause(&_ProfileRegistry.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_ProfileRegistry *ProfileRegistryTransactorSession) Pause() (*types.Transaction, error) {
	return _ProfileRegistry.Contract.Pause(&_ProfileRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ProfileRegistry *ProfileRegistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProfileRegistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ProfileRegistry *ProfileRegistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _ProfileRegistry.Contract.RenounceOwnership(&_ProfileRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ProfileRegistry *ProfileRegistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ProfileRegistry.Contract.RenounceOwnership(&_ProfileRegistry.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_ProfileRegistry *ProfileRegistryTransactor) TransferOwnership(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _ProfileRegistry.contract.Transact(opts, "transferOwnership", _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_ProfileRegistry *ProfileRegistrySession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.TransferOwnership(&_ProfileRegistry.TransactOpts, _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_ProfileRegistry *ProfileRegistryTransactorSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _ProfileRegistry.Contract.TransferOwnership(&_ProfileRegistry.TransactOpts, _newOwner)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_ProfileRegistry *ProfileRegistryTransactor) Unpause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProfileRegistry.contract.Transact(opts, "unpause")
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_ProfileRegistry *ProfileRegistrySession) Unpause() (*types.Transaction, error) {
	return _ProfileRegistry.Contract.Unpause(&_ProfileRegistry.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_ProfileRegistry *ProfileRegistryTransactorSession) Unpause() (*types.Transaction, error) {
	return _ProfileRegistry.Contract.Unpause(&_ProfileRegistry.TransactOpts)
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
// Solidity: e CertificateCreated(id indexed uint256)
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
// Solidity: e CertificateCreated(id indexed uint256)
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
// Solidity: e CertificateUpdated(id indexed uint256)
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
// Solidity: e CertificateUpdated(id indexed uint256)
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

// ProfileRegistryOwnershipRenouncedIterator is returned from FilterOwnershipRenounced and is used to iterate over the raw logs and unpacked data for OwnershipRenounced events raised by the ProfileRegistry contract.
type ProfileRegistryOwnershipRenouncedIterator struct {
	Event *ProfileRegistryOwnershipRenounced // Event containing the contract specifics and raw log

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
func (it *ProfileRegistryOwnershipRenouncedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProfileRegistryOwnershipRenounced)
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
		it.Event = new(ProfileRegistryOwnershipRenounced)
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
func (it *ProfileRegistryOwnershipRenouncedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProfileRegistryOwnershipRenouncedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProfileRegistryOwnershipRenounced represents a OwnershipRenounced event raised by the ProfileRegistry contract.
type ProfileRegistryOwnershipRenounced struct {
	PreviousOwner common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipRenounced is a free log retrieval operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_ProfileRegistry *ProfileRegistryFilterer) FilterOwnershipRenounced(opts *bind.FilterOpts, previousOwner []common.Address) (*ProfileRegistryOwnershipRenouncedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _ProfileRegistry.contract.FilterLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ProfileRegistryOwnershipRenouncedIterator{contract: _ProfileRegistry.contract, event: "OwnershipRenounced", logs: logs, sub: sub}, nil
}

// WatchOwnershipRenounced is a free log subscription operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_ProfileRegistry *ProfileRegistryFilterer) WatchOwnershipRenounced(opts *bind.WatchOpts, sink chan<- *ProfileRegistryOwnershipRenounced, previousOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _ProfileRegistry.contract.WatchLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProfileRegistryOwnershipRenounced)
				if err := _ProfileRegistry.contract.UnpackLog(event, "OwnershipRenounced", log); err != nil {
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

// ProfileRegistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ProfileRegistry contract.
type ProfileRegistryOwnershipTransferredIterator struct {
	Event *ProfileRegistryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ProfileRegistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProfileRegistryOwnershipTransferred)
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
		it.Event = new(ProfileRegistryOwnershipTransferred)
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
func (it *ProfileRegistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProfileRegistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProfileRegistryOwnershipTransferred represents a OwnershipTransferred event raised by the ProfileRegistry contract.
type ProfileRegistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_ProfileRegistry *ProfileRegistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ProfileRegistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ProfileRegistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ProfileRegistryOwnershipTransferredIterator{contract: _ProfileRegistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_ProfileRegistry *ProfileRegistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ProfileRegistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ProfileRegistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProfileRegistryOwnershipTransferred)
				if err := _ProfileRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ProfileRegistryPauseIterator is returned from FilterPause and is used to iterate over the raw logs and unpacked data for Pause events raised by the ProfileRegistry contract.
type ProfileRegistryPauseIterator struct {
	Event *ProfileRegistryPause // Event containing the contract specifics and raw log

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
func (it *ProfileRegistryPauseIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProfileRegistryPause)
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
		it.Event = new(ProfileRegistryPause)
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
func (it *ProfileRegistryPauseIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProfileRegistryPauseIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProfileRegistryPause represents a Pause event raised by the ProfileRegistry contract.
type ProfileRegistryPause struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterPause is a free log retrieval operation binding the contract event 0x6985a02210a168e66602d3235cb6db0e70f92b3ba4d376a33c0f3d9434bff625.
//
// Solidity: e Pause()
func (_ProfileRegistry *ProfileRegistryFilterer) FilterPause(opts *bind.FilterOpts) (*ProfileRegistryPauseIterator, error) {

	logs, sub, err := _ProfileRegistry.contract.FilterLogs(opts, "Pause")
	if err != nil {
		return nil, err
	}
	return &ProfileRegistryPauseIterator{contract: _ProfileRegistry.contract, event: "Pause", logs: logs, sub: sub}, nil
}

// WatchPause is a free log subscription operation binding the contract event 0x6985a02210a168e66602d3235cb6db0e70f92b3ba4d376a33c0f3d9434bff625.
//
// Solidity: e Pause()
func (_ProfileRegistry *ProfileRegistryFilterer) WatchPause(opts *bind.WatchOpts, sink chan<- *ProfileRegistryPause) (event.Subscription, error) {

	logs, sub, err := _ProfileRegistry.contract.WatchLogs(opts, "Pause")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProfileRegistryPause)
				if err := _ProfileRegistry.contract.UnpackLog(event, "Pause", log); err != nil {
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

// ProfileRegistryUnpauseIterator is returned from FilterUnpause and is used to iterate over the raw logs and unpacked data for Unpause events raised by the ProfileRegistry contract.
type ProfileRegistryUnpauseIterator struct {
	Event *ProfileRegistryUnpause // Event containing the contract specifics and raw log

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
func (it *ProfileRegistryUnpauseIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProfileRegistryUnpause)
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
		it.Event = new(ProfileRegistryUnpause)
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
func (it *ProfileRegistryUnpauseIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProfileRegistryUnpauseIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProfileRegistryUnpause represents a Unpause event raised by the ProfileRegistry contract.
type ProfileRegistryUnpause struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterUnpause is a free log retrieval operation binding the contract event 0x7805862f689e2f13df9f062ff482ad3ad112aca9e0847911ed832e158c525b33.
//
// Solidity: e Unpause()
func (_ProfileRegistry *ProfileRegistryFilterer) FilterUnpause(opts *bind.FilterOpts) (*ProfileRegistryUnpauseIterator, error) {

	logs, sub, err := _ProfileRegistry.contract.FilterLogs(opts, "Unpause")
	if err != nil {
		return nil, err
	}
	return &ProfileRegistryUnpauseIterator{contract: _ProfileRegistry.contract, event: "Unpause", logs: logs, sub: sub}, nil
}

// WatchUnpause is a free log subscription operation binding the contract event 0x7805862f689e2f13df9f062ff482ad3ad112aca9e0847911ed832e158c525b33.
//
// Solidity: e Unpause()
func (_ProfileRegistry *ProfileRegistryFilterer) WatchUnpause(opts *bind.WatchOpts, sink chan<- *ProfileRegistryUnpause) (event.Subscription, error) {

	logs, sub, err := _ProfileRegistry.contract.WatchLogs(opts, "Unpause")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProfileRegistryUnpause)
				if err := _ProfileRegistry.contract.UnpackLog(event, "Unpause", log); err != nil {
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
// Solidity: e ValidatorCreated(validator indexed address)
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
// Solidity: e ValidatorCreated(validator indexed address)
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
// Solidity: e ValidatorDeleted(validator indexed address)
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
// Solidity: e ValidatorDeleted(validator indexed address)
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
