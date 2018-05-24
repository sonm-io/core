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

// GatekeeperABI is the input ABI used to generate the binding from.
const GatekeeperABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"txNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"PayInTx\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"txNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"PayoutTx\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"BlockEmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"TransactionAmountForBlockChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"root\",\"type\":\"bytes32\"}],\"name\":\"RootAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"SetTransactionAmountForBlock\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"PayIn\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetCurrentTransactionAmountForBlock\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetTransactionCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_root\",\"type\":\"bytes32\"}],\"name\":\"VoteRoot\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetRootsCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_proof\",\"type\":\"bytes\"},{\"name\":\"_root\",\"type\":\"uint256\"},{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_txNumber\",\"type\":\"uint256\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Payout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// GatekeeperBin is the compiled bytecode used for deploying new contracts.
const GatekeeperBin = `0x60806040526000600255610200600355600060055534801561002057600080fd5b506040516020806107e383398101604052516000805460018054600160a060020a03909416600160a060020a03199485161790558216339081179092169091179055610772806100716000396000f3006080604052600436106100985763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631ade66cb811461009d57806352b75ca1146100c45780638da5cb5b146100d95780639fab56ac1461010a578063b8a7b4ae14610124578063d457547b1461013c578063e5a745b114610151578063eaf42d2714610169578063f2fde38b146101dc575b600080fd5b3480156100a957600080fd5b506100b26101fd565b60408051918252519081900360200190f35b3480156100d057600080fd5b506100b2610203565b3480156100e557600080fd5b506100ee610209565b60408051600160a060020a039092168252519081900360200190f35b34801561011657600080fd5b50610122600435610218565b005b34801561013057600080fd5b5061012260043561033a565b34801561014857600080fd5b506100b261039a565b34801561015d57600080fd5b506101226004356103a0565b34801561017557600080fd5b506040805160206004803580820135601f810184900484028501840190955284845261012294369492936024939284019190819084018382808284375094975050843595505050506020820135600160a060020a0316916040810135915060600135610480565b3480156101e857600080fd5b50610122600160a060020a036004351661061e565b60025490565b60035490565b600054600160a060020a031681565b600154604080517f23b872dd000000000000000000000000000000000000000000000000000000008152336004820152306024820152604481018490529051600160a060020a03909216916323b872dd916064808201926020929091908290030181600087803b15801561028b57600080fd5b505af115801561029f573d6000803e3d6000fd5b505050506040513d60208110156102b557600080fd5b505115156102c257600080fd5b600280546001019081905560405182919033907f63768eabd21c026cb17439a3c6556436c1b0219c2046875297ad3f4b14e6700f90600090a46003546002541415610337576000600281905560405143917fefc6c60ac095a7bb2aa44f6dba3421076ce7baccf540ff61f4d6150d1f8440d791a25b50565b600054600160a060020a0316331461035157600080fd5b60058054600101808255600090815260046020526040808220849055915491518392917f47e13ec4cc37e31e3a4f25115640068ffbe4bee53b32f0953fa593388e69fc0f91a350565b60055490565b600054600160a060020a031633146103b757600080fd5b600254811161044d57604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602560248201527f6e657720616d6f756e74206c6f776572207468616e2063757272656e7420636f60448201527f756e746572000000000000000000000000000000000000000000000000000000606482015290519081900360840190fd5b600381905560405181907f2c718a4f8bf79a664538b8187c8ffe6090a0bbb24820a968005806edf02d98a790600090a250565b604080516c01000000000000000000000000600160a060020a03861602815260148101849052603481018390528151908190036054019020600086815260046020908152838220548252600681528382208383529052919091205460ff16156104e857600080fd5b600085815260046020526040902054610503908790836106b2565b151561050e57600080fd5b600154604080517fa9059cbb000000000000000000000000000000000000000000000000000000008152600160a060020a038781166004830152602482018690529151919092169163a9059cbb9160448083019260209291908290030181600087803b15801561057d57600080fd5b505af1158015610591573d6000803e3d6000fd5b505050506040513d60208110156105a757600080fd5b505115156105b457600080fd5b600085815260046020908152604080832054835260068252808320848452909152808220805460ff191660011790555183918591600160a060020a038816917f731af16374848c2c73a6154fd410cb421138e7db45c5a904e5a475c756faa8d991a4505050505050565b600054600160a060020a0316331461063557600080fd5b600160a060020a038116151561064a57600080fd5b60008054604051600160a060020a03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a36000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0392909216919091179055565b600080600080602087518115156106c557fe5b06156106d4576000935061073c565b5083905060205b865161ffff8216116107365786810151925082821015610713576040805192835260208301849052805192839003019091209061072e565b60408051848152602081019390935280519283900301909120905b6020016106db565b81861493505b50505093925050505600a165627a7a7230582003614b159741b294bd5053e1f3f6890f6f35ef9425ef2d6560dfcbb1dfc54dc40029`

// DeployGatekeeper deploys a new Ethereum contract, binding an instance of Gatekeeper to it.
func DeployGatekeeper(auth *bind.TransactOpts, backend bind.ContractBackend, _token common.Address) (common.Address, *types.Transaction, *Gatekeeper, error) {
	parsed, err := abi.JSON(strings.NewReader(GatekeeperABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(GatekeeperBin), backend, _token)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Gatekeeper{GatekeeperCaller: GatekeeperCaller{contract: contract}, GatekeeperTransactor: GatekeeperTransactor{contract: contract}, GatekeeperFilterer: GatekeeperFilterer{contract: contract}}, nil
}

// Gatekeeper is an auto generated Go binding around an Ethereum contract.
type Gatekeeper struct {
	GatekeeperCaller     // Read-only binding to the contract
	GatekeeperTransactor // Write-only binding to the contract
	GatekeeperFilterer   // Log filterer for contract events
}

// GatekeeperCaller is an auto generated read-only Go binding around an Ethereum contract.
type GatekeeperCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GatekeeperTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GatekeeperTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GatekeeperFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GatekeeperFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GatekeeperSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GatekeeperSession struct {
	Contract     *Gatekeeper       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GatekeeperCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GatekeeperCallerSession struct {
	Contract *GatekeeperCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// GatekeeperTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GatekeeperTransactorSession struct {
	Contract     *GatekeeperTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// GatekeeperRaw is an auto generated low-level Go binding around an Ethereum contract.
type GatekeeperRaw struct {
	Contract *Gatekeeper // Generic contract binding to access the raw methods on
}

// GatekeeperCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GatekeeperCallerRaw struct {
	Contract *GatekeeperCaller // Generic read-only contract binding to access the raw methods on
}

// GatekeeperTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GatekeeperTransactorRaw struct {
	Contract *GatekeeperTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGatekeeper creates a new instance of Gatekeeper, bound to a specific deployed contract.
func NewGatekeeper(address common.Address, backend bind.ContractBackend) (*Gatekeeper, error) {
	contract, err := bindGatekeeper(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Gatekeeper{GatekeeperCaller: GatekeeperCaller{contract: contract}, GatekeeperTransactor: GatekeeperTransactor{contract: contract}, GatekeeperFilterer: GatekeeperFilterer{contract: contract}}, nil
}

// NewGatekeeperCaller creates a new read-only instance of Gatekeeper, bound to a specific deployed contract.
func NewGatekeeperCaller(address common.Address, caller bind.ContractCaller) (*GatekeeperCaller, error) {
	contract, err := bindGatekeeper(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GatekeeperCaller{contract: contract}, nil
}

// NewGatekeeperTransactor creates a new write-only instance of Gatekeeper, bound to a specific deployed contract.
func NewGatekeeperTransactor(address common.Address, transactor bind.ContractTransactor) (*GatekeeperTransactor, error) {
	contract, err := bindGatekeeper(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GatekeeperTransactor{contract: contract}, nil
}

// NewGatekeeperFilterer creates a new log filterer instance of Gatekeeper, bound to a specific deployed contract.
func NewGatekeeperFilterer(address common.Address, filterer bind.ContractFilterer) (*GatekeeperFilterer, error) {
	contract, err := bindGatekeeper(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GatekeeperFilterer{contract: contract}, nil
}

// bindGatekeeper binds a generic wrapper to an already deployed contract.
func bindGatekeeper(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(GatekeeperABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Gatekeeper *GatekeeperRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Gatekeeper.Contract.GatekeeperCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Gatekeeper *GatekeeperRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Gatekeeper.Contract.GatekeeperTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Gatekeeper *GatekeeperRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Gatekeeper.Contract.GatekeeperTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Gatekeeper *GatekeeperCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Gatekeeper.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Gatekeeper *GatekeeperTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Gatekeeper.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Gatekeeper *GatekeeperTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Gatekeeper.Contract.contract.Transact(opts, method, params...)
}

// GetCurrentTransactionAmountForBlock is a free data retrieval call binding the contract method 0x52b75ca1.
//
// Solidity: function GetCurrentTransactionAmountForBlock() constant returns(uint256)
func (_Gatekeeper *GatekeeperCaller) GetCurrentTransactionAmountForBlock(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Gatekeeper.contract.Call(opts, out, "GetCurrentTransactionAmountForBlock")
	return *ret0, err
}

// GetCurrentTransactionAmountForBlock is a free data retrieval call binding the contract method 0x52b75ca1.
//
// Solidity: function GetCurrentTransactionAmountForBlock() constant returns(uint256)
func (_Gatekeeper *GatekeeperSession) GetCurrentTransactionAmountForBlock() (*big.Int, error) {
	return _Gatekeeper.Contract.GetCurrentTransactionAmountForBlock(&_Gatekeeper.CallOpts)
}

// GetCurrentTransactionAmountForBlock is a free data retrieval call binding the contract method 0x52b75ca1.
//
// Solidity: function GetCurrentTransactionAmountForBlock() constant returns(uint256)
func (_Gatekeeper *GatekeeperCallerSession) GetCurrentTransactionAmountForBlock() (*big.Int, error) {
	return _Gatekeeper.Contract.GetCurrentTransactionAmountForBlock(&_Gatekeeper.CallOpts)
}

// GetRootsCount is a free data retrieval call binding the contract method 0xd457547b.
//
// Solidity: function GetRootsCount() constant returns(uint256)
func (_Gatekeeper *GatekeeperCaller) GetRootsCount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Gatekeeper.contract.Call(opts, out, "GetRootsCount")
	return *ret0, err
}

// GetRootsCount is a free data retrieval call binding the contract method 0xd457547b.
//
// Solidity: function GetRootsCount() constant returns(uint256)
func (_Gatekeeper *GatekeeperSession) GetRootsCount() (*big.Int, error) {
	return _Gatekeeper.Contract.GetRootsCount(&_Gatekeeper.CallOpts)
}

// GetRootsCount is a free data retrieval call binding the contract method 0xd457547b.
//
// Solidity: function GetRootsCount() constant returns(uint256)
func (_Gatekeeper *GatekeeperCallerSession) GetRootsCount() (*big.Int, error) {
	return _Gatekeeper.Contract.GetRootsCount(&_Gatekeeper.CallOpts)
}

// GetTransactionCount is a free data retrieval call binding the contract method 0x1ade66cb.
//
// Solidity: function GetTransactionCount() constant returns(uint256)
func (_Gatekeeper *GatekeeperCaller) GetTransactionCount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Gatekeeper.contract.Call(opts, out, "GetTransactionCount")
	return *ret0, err
}

// GetTransactionCount is a free data retrieval call binding the contract method 0x1ade66cb.
//
// Solidity: function GetTransactionCount() constant returns(uint256)
func (_Gatekeeper *GatekeeperSession) GetTransactionCount() (*big.Int, error) {
	return _Gatekeeper.Contract.GetTransactionCount(&_Gatekeeper.CallOpts)
}

// GetTransactionCount is a free data retrieval call binding the contract method 0x1ade66cb.
//
// Solidity: function GetTransactionCount() constant returns(uint256)
func (_Gatekeeper *GatekeeperCallerSession) GetTransactionCount() (*big.Int, error) {
	return _Gatekeeper.Contract.GetTransactionCount(&_Gatekeeper.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Gatekeeper *GatekeeperCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Gatekeeper.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Gatekeeper *GatekeeperSession) Owner() (common.Address, error) {
	return _Gatekeeper.Contract.Owner(&_Gatekeeper.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Gatekeeper *GatekeeperCallerSession) Owner() (common.Address, error) {
	return _Gatekeeper.Contract.Owner(&_Gatekeeper.CallOpts)
}

// PayIn is a paid mutator transaction binding the contract method 0x9fab56ac.
//
// Solidity: function PayIn(_value uint256) returns()
func (_Gatekeeper *GatekeeperTransactor) PayIn(opts *bind.TransactOpts, _value *big.Int) (*types.Transaction, error) {
	return _Gatekeeper.contract.Transact(opts, "PayIn", _value)
}

// PayIn is a paid mutator transaction binding the contract method 0x9fab56ac.
//
// Solidity: function PayIn(_value uint256) returns()
func (_Gatekeeper *GatekeeperSession) PayIn(_value *big.Int) (*types.Transaction, error) {
	return _Gatekeeper.Contract.PayIn(&_Gatekeeper.TransactOpts, _value)
}

// PayIn is a paid mutator transaction binding the contract method 0x9fab56ac.
//
// Solidity: function PayIn(_value uint256) returns()
func (_Gatekeeper *GatekeeperTransactorSession) PayIn(_value *big.Int) (*types.Transaction, error) {
	return _Gatekeeper.Contract.PayIn(&_Gatekeeper.TransactOpts, _value)
}

// Payout is a paid mutator transaction binding the contract method 0xeaf42d27.
//
// Solidity: function Payout(_proof bytes, _root uint256, _from address, _txNumber uint256, _value uint256) returns()
func (_Gatekeeper *GatekeeperTransactor) Payout(opts *bind.TransactOpts, _proof []byte, _root *big.Int, _from common.Address, _txNumber *big.Int, _value *big.Int) (*types.Transaction, error) {
	return _Gatekeeper.contract.Transact(opts, "Payout", _proof, _root, _from, _txNumber, _value)
}

// Payout is a paid mutator transaction binding the contract method 0xeaf42d27.
//
// Solidity: function Payout(_proof bytes, _root uint256, _from address, _txNumber uint256, _value uint256) returns()
func (_Gatekeeper *GatekeeperSession) Payout(_proof []byte, _root *big.Int, _from common.Address, _txNumber *big.Int, _value *big.Int) (*types.Transaction, error) {
	return _Gatekeeper.Contract.Payout(&_Gatekeeper.TransactOpts, _proof, _root, _from, _txNumber, _value)
}

// Payout is a paid mutator transaction binding the contract method 0xeaf42d27.
//
// Solidity: function Payout(_proof bytes, _root uint256, _from address, _txNumber uint256, _value uint256) returns()
func (_Gatekeeper *GatekeeperTransactorSession) Payout(_proof []byte, _root *big.Int, _from common.Address, _txNumber *big.Int, _value *big.Int) (*types.Transaction, error) {
	return _Gatekeeper.Contract.Payout(&_Gatekeeper.TransactOpts, _proof, _root, _from, _txNumber, _value)
}

// SetTransactionAmountForBlock is a paid mutator transaction binding the contract method 0xe5a745b1.
//
// Solidity: function SetTransactionAmountForBlock(_amount uint256) returns()
func (_Gatekeeper *GatekeeperTransactor) SetTransactionAmountForBlock(opts *bind.TransactOpts, _amount *big.Int) (*types.Transaction, error) {
	return _Gatekeeper.contract.Transact(opts, "SetTransactionAmountForBlock", _amount)
}

// SetTransactionAmountForBlock is a paid mutator transaction binding the contract method 0xe5a745b1.
//
// Solidity: function SetTransactionAmountForBlock(_amount uint256) returns()
func (_Gatekeeper *GatekeeperSession) SetTransactionAmountForBlock(_amount *big.Int) (*types.Transaction, error) {
	return _Gatekeeper.Contract.SetTransactionAmountForBlock(&_Gatekeeper.TransactOpts, _amount)
}

// SetTransactionAmountForBlock is a paid mutator transaction binding the contract method 0xe5a745b1.
//
// Solidity: function SetTransactionAmountForBlock(_amount uint256) returns()
func (_Gatekeeper *GatekeeperTransactorSession) SetTransactionAmountForBlock(_amount *big.Int) (*types.Transaction, error) {
	return _Gatekeeper.Contract.SetTransactionAmountForBlock(&_Gatekeeper.TransactOpts, _amount)
}

// VoteRoot is a paid mutator transaction binding the contract method 0xb8a7b4ae.
//
// Solidity: function VoteRoot(_root bytes32) returns()
func (_Gatekeeper *GatekeeperTransactor) VoteRoot(opts *bind.TransactOpts, _root [32]byte) (*types.Transaction, error) {
	return _Gatekeeper.contract.Transact(opts, "VoteRoot", _root)
}

// VoteRoot is a paid mutator transaction binding the contract method 0xb8a7b4ae.
//
// Solidity: function VoteRoot(_root bytes32) returns()
func (_Gatekeeper *GatekeeperSession) VoteRoot(_root [32]byte) (*types.Transaction, error) {
	return _Gatekeeper.Contract.VoteRoot(&_Gatekeeper.TransactOpts, _root)
}

// VoteRoot is a paid mutator transaction binding the contract method 0xb8a7b4ae.
//
// Solidity: function VoteRoot(_root bytes32) returns()
func (_Gatekeeper *GatekeeperTransactorSession) VoteRoot(_root [32]byte) (*types.Transaction, error) {
	return _Gatekeeper.Contract.VoteRoot(&_Gatekeeper.TransactOpts, _root)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_Gatekeeper *GatekeeperTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Gatekeeper.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_Gatekeeper *GatekeeperSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Gatekeeper.Contract.TransferOwnership(&_Gatekeeper.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_Gatekeeper *GatekeeperTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Gatekeeper.Contract.TransferOwnership(&_Gatekeeper.TransactOpts, newOwner)
}

// GatekeeperBlockEmittedIterator is returned from FilterBlockEmitted and is used to iterate over the raw logs and unpacked data for BlockEmitted events raised by the Gatekeeper contract.
type GatekeeperBlockEmittedIterator struct {
	Event *GatekeeperBlockEmitted // Event containing the contract specifics and raw log

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
func (it *GatekeeperBlockEmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatekeeperBlockEmitted)
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
		it.Event = new(GatekeeperBlockEmitted)
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
func (it *GatekeeperBlockEmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatekeeperBlockEmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatekeeperBlockEmitted represents a BlockEmitted event raised by the Gatekeeper contract.
type GatekeeperBlockEmitted struct {
	BlockNumber *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterBlockEmitted is a free log retrieval operation binding the contract event 0xefc6c60ac095a7bb2aa44f6dba3421076ce7baccf540ff61f4d6150d1f8440d7.
//
// Solidity: event BlockEmitted(blockNumber indexed uint256)
func (_Gatekeeper *GatekeeperFilterer) FilterBlockEmitted(opts *bind.FilterOpts, blockNumber []*big.Int) (*GatekeeperBlockEmittedIterator, error) {

	var blockNumberRule []interface{}
	for _, blockNumberItem := range blockNumber {
		blockNumberRule = append(blockNumberRule, blockNumberItem)
	}

	logs, sub, err := _Gatekeeper.contract.FilterLogs(opts, "BlockEmitted", blockNumberRule)
	if err != nil {
		return nil, err
	}
	return &GatekeeperBlockEmittedIterator{contract: _Gatekeeper.contract, event: "BlockEmitted", logs: logs, sub: sub}, nil
}

// WatchBlockEmitted is a free log subscription operation binding the contract event 0xefc6c60ac095a7bb2aa44f6dba3421076ce7baccf540ff61f4d6150d1f8440d7.
//
// Solidity: event BlockEmitted(blockNumber indexed uint256)
func (_Gatekeeper *GatekeeperFilterer) WatchBlockEmitted(opts *bind.WatchOpts, sink chan<- *GatekeeperBlockEmitted, blockNumber []*big.Int) (event.Subscription, error) {

	var blockNumberRule []interface{}
	for _, blockNumberItem := range blockNumber {
		blockNumberRule = append(blockNumberRule, blockNumberItem)
	}

	logs, sub, err := _Gatekeeper.contract.WatchLogs(opts, "BlockEmitted", blockNumberRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatekeeperBlockEmitted)
				if err := _Gatekeeper.contract.UnpackLog(event, "BlockEmitted", log); err != nil {
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

// GatekeeperOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Gatekeeper contract.
type GatekeeperOwnershipTransferredIterator struct {
	Event *GatekeeperOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *GatekeeperOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatekeeperOwnershipTransferred)
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
		it.Event = new(GatekeeperOwnershipTransferred)
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
func (it *GatekeeperOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatekeeperOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatekeeperOwnershipTransferred represents a OwnershipTransferred event raised by the Gatekeeper contract.
type GatekeeperOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_Gatekeeper *GatekeeperFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*GatekeeperOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Gatekeeper.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &GatekeeperOwnershipTransferredIterator{contract: _Gatekeeper.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_Gatekeeper *GatekeeperFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *GatekeeperOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Gatekeeper.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatekeeperOwnershipTransferred)
				if err := _Gatekeeper.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// GatekeeperPayInTxIterator is returned from FilterPayInTx and is used to iterate over the raw logs and unpacked data for PayInTx events raised by the Gatekeeper contract.
type GatekeeperPayInTxIterator struct {
	Event *GatekeeperPayInTx // Event containing the contract specifics and raw log

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
func (it *GatekeeperPayInTxIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatekeeperPayInTx)
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
		it.Event = new(GatekeeperPayInTx)
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
func (it *GatekeeperPayInTxIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatekeeperPayInTxIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatekeeperPayInTx represents a PayInTx event raised by the Gatekeeper contract.
type GatekeeperPayInTx struct {
	From     common.Address
	TxNumber *big.Int
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPayInTx is a free log retrieval operation binding the contract event 0x63768eabd21c026cb17439a3c6556436c1b0219c2046875297ad3f4b14e6700f.
//
// Solidity: event PayInTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_Gatekeeper *GatekeeperFilterer) FilterPayInTx(opts *bind.FilterOpts, from []common.Address, txNumber []*big.Int, value []*big.Int) (*GatekeeperPayInTxIterator, error) {

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

	logs, sub, err := _Gatekeeper.contract.FilterLogs(opts, "PayInTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return &GatekeeperPayInTxIterator{contract: _Gatekeeper.contract, event: "PayInTx", logs: logs, sub: sub}, nil
}

// WatchPayInTx is a free log subscription operation binding the contract event 0x63768eabd21c026cb17439a3c6556436c1b0219c2046875297ad3f4b14e6700f.
//
// Solidity: event PayInTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_Gatekeeper *GatekeeperFilterer) WatchPayInTx(opts *bind.WatchOpts, sink chan<- *GatekeeperPayInTx, from []common.Address, txNumber []*big.Int, value []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Gatekeeper.contract.WatchLogs(opts, "PayInTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatekeeperPayInTx)
				if err := _Gatekeeper.contract.UnpackLog(event, "PayInTx", log); err != nil {
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

// GatekeeperPayoutTxIterator is returned from FilterPayoutTx and is used to iterate over the raw logs and unpacked data for PayoutTx events raised by the Gatekeeper contract.
type GatekeeperPayoutTxIterator struct {
	Event *GatekeeperPayoutTx // Event containing the contract specifics and raw log

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
func (it *GatekeeperPayoutTxIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatekeeperPayoutTx)
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
		it.Event = new(GatekeeperPayoutTx)
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
func (it *GatekeeperPayoutTxIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatekeeperPayoutTxIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatekeeperPayoutTx represents a PayoutTx event raised by the Gatekeeper contract.
type GatekeeperPayoutTx struct {
	From     common.Address
	TxNumber *big.Int
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPayoutTx is a free log retrieval operation binding the contract event 0x731af16374848c2c73a6154fd410cb421138e7db45c5a904e5a475c756faa8d9.
//
// Solidity: event PayoutTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_Gatekeeper *GatekeeperFilterer) FilterPayoutTx(opts *bind.FilterOpts, from []common.Address, txNumber []*big.Int, value []*big.Int) (*GatekeeperPayoutTxIterator, error) {

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

	logs, sub, err := _Gatekeeper.contract.FilterLogs(opts, "PayoutTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return &GatekeeperPayoutTxIterator{contract: _Gatekeeper.contract, event: "PayoutTx", logs: logs, sub: sub}, nil
}

// WatchPayoutTx is a free log subscription operation binding the contract event 0x731af16374848c2c73a6154fd410cb421138e7db45c5a904e5a475c756faa8d9.
//
// Solidity: event PayoutTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_Gatekeeper *GatekeeperFilterer) WatchPayoutTx(opts *bind.WatchOpts, sink chan<- *GatekeeperPayoutTx, from []common.Address, txNumber []*big.Int, value []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Gatekeeper.contract.WatchLogs(opts, "PayoutTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatekeeperPayoutTx)
				if err := _Gatekeeper.contract.UnpackLog(event, "PayoutTx", log); err != nil {
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

// GatekeeperRootAddedIterator is returned from FilterRootAdded and is used to iterate over the raw logs and unpacked data for RootAdded events raised by the Gatekeeper contract.
type GatekeeperRootAddedIterator struct {
	Event *GatekeeperRootAdded // Event containing the contract specifics and raw log

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
func (it *GatekeeperRootAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatekeeperRootAdded)
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
		it.Event = new(GatekeeperRootAdded)
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
func (it *GatekeeperRootAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatekeeperRootAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatekeeperRootAdded represents a RootAdded event raised by the Gatekeeper contract.
type GatekeeperRootAdded struct {
	Id   *big.Int
	Root [32]byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterRootAdded is a free log retrieval operation binding the contract event 0x47e13ec4cc37e31e3a4f25115640068ffbe4bee53b32f0953fa593388e69fc0f.
//
// Solidity: event RootAdded(id indexed uint256, root indexed bytes32)
func (_Gatekeeper *GatekeeperFilterer) FilterRootAdded(opts *bind.FilterOpts, id []*big.Int, root [][32]byte) (*GatekeeperRootAddedIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var rootRule []interface{}
	for _, rootItem := range root {
		rootRule = append(rootRule, rootItem)
	}

	logs, sub, err := _Gatekeeper.contract.FilterLogs(opts, "RootAdded", idRule, rootRule)
	if err != nil {
		return nil, err
	}
	return &GatekeeperRootAddedIterator{contract: _Gatekeeper.contract, event: "RootAdded", logs: logs, sub: sub}, nil
}

// WatchRootAdded is a free log subscription operation binding the contract event 0x47e13ec4cc37e31e3a4f25115640068ffbe4bee53b32f0953fa593388e69fc0f.
//
// Solidity: event RootAdded(id indexed uint256, root indexed bytes32)
func (_Gatekeeper *GatekeeperFilterer) WatchRootAdded(opts *bind.WatchOpts, sink chan<- *GatekeeperRootAdded, id []*big.Int, root [][32]byte) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var rootRule []interface{}
	for _, rootItem := range root {
		rootRule = append(rootRule, rootItem)
	}

	logs, sub, err := _Gatekeeper.contract.WatchLogs(opts, "RootAdded", idRule, rootRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatekeeperRootAdded)
				if err := _Gatekeeper.contract.UnpackLog(event, "RootAdded", log); err != nil {
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

// GatekeeperTransactionAmountForBlockChangedIterator is returned from FilterTransactionAmountForBlockChanged and is used to iterate over the raw logs and unpacked data for TransactionAmountForBlockChanged events raised by the Gatekeeper contract.
type GatekeeperTransactionAmountForBlockChangedIterator struct {
	Event *GatekeeperTransactionAmountForBlockChanged // Event containing the contract specifics and raw log

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
func (it *GatekeeperTransactionAmountForBlockChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatekeeperTransactionAmountForBlockChanged)
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
		it.Event = new(GatekeeperTransactionAmountForBlockChanged)
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
func (it *GatekeeperTransactionAmountForBlockChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatekeeperTransactionAmountForBlockChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatekeeperTransactionAmountForBlockChanged represents a TransactionAmountForBlockChanged event raised by the Gatekeeper contract.
type GatekeeperTransactionAmountForBlockChanged struct {
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterTransactionAmountForBlockChanged is a free log retrieval operation binding the contract event 0x2c718a4f8bf79a664538b8187c8ffe6090a0bbb24820a968005806edf02d98a7.
//
// Solidity: event TransactionAmountForBlockChanged(amount indexed uint256)
func (_Gatekeeper *GatekeeperFilterer) FilterTransactionAmountForBlockChanged(opts *bind.FilterOpts, amount []*big.Int) (*GatekeeperTransactionAmountForBlockChangedIterator, error) {

	var amountRule []interface{}
	for _, amountItem := range amount {
		amountRule = append(amountRule, amountItem)
	}

	logs, sub, err := _Gatekeeper.contract.FilterLogs(opts, "TransactionAmountForBlockChanged", amountRule)
	if err != nil {
		return nil, err
	}
	return &GatekeeperTransactionAmountForBlockChangedIterator{contract: _Gatekeeper.contract, event: "TransactionAmountForBlockChanged", logs: logs, sub: sub}, nil
}

// WatchTransactionAmountForBlockChanged is a free log subscription operation binding the contract event 0x2c718a4f8bf79a664538b8187c8ffe6090a0bbb24820a968005806edf02d98a7.
//
// Solidity: event TransactionAmountForBlockChanged(amount indexed uint256)
func (_Gatekeeper *GatekeeperFilterer) WatchTransactionAmountForBlockChanged(opts *bind.WatchOpts, sink chan<- *GatekeeperTransactionAmountForBlockChanged, amount []*big.Int) (event.Subscription, error) {

	var amountRule []interface{}
	for _, amountItem := range amount {
		amountRule = append(amountRule, amountItem)
	}

	logs, sub, err := _Gatekeeper.contract.WatchLogs(opts, "TransactionAmountForBlockChanged", amountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatekeeperTransactionAmountForBlockChanged)
				if err := _Gatekeeper.contract.UnpackLog(event, "TransactionAmountForBlockChanged", log); err != nil {
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
