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

// SimpleGatekeeperWithLimitLiveABI is the input ABI used to generate the binding from.
const SimpleGatekeeperWithLimitLiveABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"keepers\",\"outputs\":[{\"name\":\"dayLimit\",\"type\":\"uint256\"},{\"name\":\"lastDay\",\"type\":\"uint256\"},{\"name\":\"spentToday\",\"type\":\"uint256\"},{\"name\":\"frozen\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"paid\",\"outputs\":[{\"name\":\"commitTS\",\"type\":\"uint256\"},{\"name\":\"paid\",\"type\":\"bool\"},{\"name\":\"keeper\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"transactionAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"},{\"name\":\"_freezingTime\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"txNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"PayinTx\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"txNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"commitTimestamp\",\"type\":\"uint256\"}],\"name\":\"CommitTx\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"txNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"PayoutTx\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"block\",\"type\":\"uint256\"}],\"name\":\"Suicide\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"keeper\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"dayLimit\",\"type\":\"uint256\"}],\"name\":\"LimitChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"keeper\",\"type\":\"address\"}],\"name\":\"KeeperFreezed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"keeper\",\"type\":\"address\"}],\"name\":\"KeeperUnfreezed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"}],\"name\":\"OwnershipRenounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_keeper\",\"type\":\"address\"},{\"name\":\"_limit\",\"type\":\"uint256\"}],\"name\":\"ChangeKeeperLimit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_keeper\",\"type\":\"address\"}],\"name\":\"FreezeKeeper\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_keeper\",\"type\":\"address\"}],\"name\":\"UnfreezeKeeper\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Payin\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"},{\"name\":\"_txNumber\",\"type\":\"uint256\"}],\"name\":\"Payout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_freezingTime\",\"type\":\"uint256\"}],\"name\":\"SetFreezingTime\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetFreezingTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"kill\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// SimpleGatekeeperWithLimitLiveBin is the compiled bytecode used for deploying new contracts.
const SimpleGatekeeperWithLimitLiveBin = `0x6080604052600060035534801561001557600080fd5b50604051604080610c1c8339810160405280516020909101516000805460018054600160a060020a03909516600160a060020a0319958616179055831633908117909316909217909155600555610bab806100716000396000f3006080604052600436106100cf5763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166328f727f081146100d457806336ab802e146100ee5780633bbd64bc1461011557806341c0e1b51461015e578063634235fc14610173578063715018a61461019a5780638da5cb5b146101af578063ad835c0b146101e0578063add89bb214610204578063b38ad8e714610243578063cc38d7ca14610264578063d942bffa1461027c578063e5837a7b14610291578063f2fde38b146102b2575b600080fd5b3480156100e057600080fd5b506100ec6004356102d3565b005b3480156100fa57600080fd5b5061010361039a565b60408051918252519081900360200190f35b34801561012157600080fd5b50610136600160a060020a03600435166103a0565b6040805194855260208501939093528383019190915215156060830152519081900360800190f35b34801561016a57600080fd5b506100ec6103cc565b34801561017f57600080fd5b506100ec600160a060020a0360043516602435604435610542565b3480156101a657600080fd5b506100ec6107ab565b3480156101bb57600080fd5b506101c4610817565b60408051600160a060020a039092168252519081900360200190f35b3480156101ec57600080fd5b506100ec600160a060020a0360043516602435610826565b34801561021057600080fd5b5061021c600435610884565b604080519384529115156020840152600160a060020a031682820152519081900360600190f35b34801561024f57600080fd5b506100ec600160a060020a03600435166108b0565b34801561027057600080fd5b506100ec60043561093c565b34801561028857600080fd5b50610103610958565b34801561029d57600080fd5b506100ec600160a060020a036004351661095e565b3480156102be57600080fd5b506100ec600160a060020a03600435166109e4565b600154604080517f23b872dd000000000000000000000000000000000000000000000000000000008152336004820152306024820152604481018490529051600160a060020a03909216916323b872dd9160648082019260009290919082900301818387803b15801561034557600080fd5b505af1158015610359573d6000803e3d6000fd5b5050600380546001019081905560405184935090915033907f14312725abbc46ad798bc078b2663e1fcbace97be0247cd177176f3b4df2538e90600090a450565b60055490565b600260208190526000918252604090912080546001820154928201546003909201549092919060ff1684565b600054600160a060020a031633146103e357600080fd5b60015460008054604080517f70a082310000000000000000000000000000000000000000000000000000000081523060048201529051600160a060020a039485169463a9059cbb9493169285926370a082319260248083019360209383900390910190829087803b15801561045757600080fd5b505af115801561046b573d6000803e3d6000fd5b505050506040513d602081101561048157600080fd5b5051604080517c010000000000000000000000000000000000000000000000000000000063ffffffff8616028152600160a060020a039093166004840152602483019190915251604480830192600092919082900301818387803b1580156104e857600080fd5b505af11580156104fc573d6000803e3d6000fd5b50506040805142815290517fa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde9350908190036020019150a1600054600160a060020a0316ff5b3360009081526002602052604081206003015460ff161561056257600080fd5b336000908152600260205260408120541161057c57600080fd5b50604080516c01000000000000000000000000600160a060020a0386160281526014810183905260348101849052815190819003605401902060008181526004602052919091206001015460ff16156105d457600080fd5b6000818152600460205260409020541515610680576105f33384610a07565b15156105fe57600080fd5b600081815260046020908152604091829020428082556001909101805474ffffffffffffffffffffffffffffffffffffffff00191633610100021790558251908152915185928592600160a060020a038916927f65546c3bc3a77ffc91667da85018004299542e28a511328cfb4b3f86974902ee9281900390910190a46107a5565b6000818152600460205260409020600101546101009004600160a060020a031633146106ab57600080fd5b60055460008281526004602052604090205442910111156106cb57600080fd5b600154604080517fa9059cbb000000000000000000000000000000000000000000000000000000008152600160a060020a038781166004830152602482018790529151919092169163a9059cbb91604480830192600092919082900301818387803b15801561073957600080fd5b505af115801561074d573d6000803e3d6000fd5b5050506000828152600460205260408082206001908101805460ff19169091179055518592508491600160a060020a038816917f731af16374848c2c73a6154fd410cb421138e7db45c5a904e5a475c756faa8d99190a45b50505050565b600054600160a060020a031633146107c257600080fd5b60008054604051600160a060020a03909116917ff8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c6482091a26000805473ffffffffffffffffffffffffffffffffffffffff19169055565b600054600160a060020a031681565b600054600160a060020a0316331461083d57600080fd5b600160a060020a038216600081815260026020526040808220849055518392917fef9c668177207fb68ca5e3894a1efacebb659762b27a737fde58ceebc4f30ad391a35050565b6004602052600090815260409020805460019091015460ff8116906101009004600160a060020a031683565b33600090815260026020526040812054116108ca57600080fd5b600160a060020a038116600090815260026020526040812054116108ed57600080fd5b600160a060020a038116600081815260026020526040808220600301805460ff19166001179055517fdf4868d2f39f6ab9f41b92c6917da5aec882c461ce7316bb62076865108502bd9190a250565b600054600160a060020a0316331461095357600080fd5b600555565b60035481565b600054600160a060020a0316331461097557600080fd5b600160a060020a0381166000908152600260205260408120541161099857600080fd5b600160a060020a038116600081815260026020526040808220600301805460ff19169055517fbbe17a7427b5192903e1b3f0f2b6ef8b2a1af9b33e1079faf8f8383f2fb63b539190a250565b600054600160a060020a031633146109fb57600080fd5b610a0481610af9565b50565b600160a060020a038216600090815260026020526040812060010154610a2b610b76565b1115610a7257600160a060020a038316600090815260026020819052604082200155610a55610b76565b600160a060020a0384166000908152600260205260409020600101555b600160a060020a0383166000908152600260208190526040909120015482810110801590610ac05750600160a060020a03831660009081526002602081905260409091208054910154830111155b15610aef5750600160a060020a0382166000908152600260208190526040909120018054820190556001610af3565b5060005b92915050565b600160a060020a0381161515610b0e57600080fd5b60008054604051600160a060020a03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a36000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0392909216919091179055565b620151804204905600a165627a7a72305820e7c887d3422c4f9a9a340d62d676e86f8c6e9ca6019825f49e43b298b4996cb50029`

// DeploySimpleGatekeeperWithLimitLive deploys a new Ethereum contract, binding an instance of SimpleGatekeeperWithLimitLive to it.
func DeploySimpleGatekeeperWithLimitLive(auth *bind.TransactOpts, backend bind.ContractBackend, _token common.Address, _freezingTime *big.Int) (common.Address, *types.Transaction, *SimpleGatekeeperWithLimitLive, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleGatekeeperWithLimitLiveABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SimpleGatekeeperWithLimitLiveBin), backend, _token, _freezingTime)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SimpleGatekeeperWithLimitLive{SimpleGatekeeperWithLimitLiveCaller: SimpleGatekeeperWithLimitLiveCaller{contract: contract}, SimpleGatekeeperWithLimitLiveTransactor: SimpleGatekeeperWithLimitLiveTransactor{contract: contract}, SimpleGatekeeperWithLimitLiveFilterer: SimpleGatekeeperWithLimitLiveFilterer{contract: contract}}, nil
}

// SimpleGatekeeperWithLimitLive is an auto generated Go binding around an Ethereum contract.
type SimpleGatekeeperWithLimitLive struct {
	SimpleGatekeeperWithLimitLiveCaller     // Read-only binding to the contract
	SimpleGatekeeperWithLimitLiveTransactor // Write-only binding to the contract
	SimpleGatekeeperWithLimitLiveFilterer   // Log filterer for contract events
}

// SimpleGatekeeperWithLimitLiveCaller is an auto generated read-only Go binding around an Ethereum contract.
type SimpleGatekeeperWithLimitLiveCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleGatekeeperWithLimitLiveTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SimpleGatekeeperWithLimitLiveTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleGatekeeperWithLimitLiveFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SimpleGatekeeperWithLimitLiveFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleGatekeeperWithLimitLiveSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SimpleGatekeeperWithLimitLiveSession struct {
	Contract     *SimpleGatekeeperWithLimitLive // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                  // Call options to use throughout this session
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// SimpleGatekeeperWithLimitLiveCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SimpleGatekeeperWithLimitLiveCallerSession struct {
	Contract *SimpleGatekeeperWithLimitLiveCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                        // Call options to use throughout this session
}

// SimpleGatekeeperWithLimitLiveTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SimpleGatekeeperWithLimitLiveTransactorSession struct {
	Contract     *SimpleGatekeeperWithLimitLiveTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                        // Transaction auth options to use throughout this session
}

// SimpleGatekeeperWithLimitLiveRaw is an auto generated low-level Go binding around an Ethereum contract.
type SimpleGatekeeperWithLimitLiveRaw struct {
	Contract *SimpleGatekeeperWithLimitLive // Generic contract binding to access the raw methods on
}

// SimpleGatekeeperWithLimitLiveCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SimpleGatekeeperWithLimitLiveCallerRaw struct {
	Contract *SimpleGatekeeperWithLimitLiveCaller // Generic read-only contract binding to access the raw methods on
}

// SimpleGatekeeperWithLimitLiveTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SimpleGatekeeperWithLimitLiveTransactorRaw struct {
	Contract *SimpleGatekeeperWithLimitLiveTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSimpleGatekeeperWithLimitLive creates a new instance of SimpleGatekeeperWithLimitLive, bound to a specific deployed contract.
func NewSimpleGatekeeperWithLimitLive(address common.Address, backend bind.ContractBackend) (*SimpleGatekeeperWithLimitLive, error) {
	contract, err := bindSimpleGatekeeperWithLimitLive(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitLive{SimpleGatekeeperWithLimitLiveCaller: SimpleGatekeeperWithLimitLiveCaller{contract: contract}, SimpleGatekeeperWithLimitLiveTransactor: SimpleGatekeeperWithLimitLiveTransactor{contract: contract}, SimpleGatekeeperWithLimitLiveFilterer: SimpleGatekeeperWithLimitLiveFilterer{contract: contract}}, nil
}

// NewSimpleGatekeeperWithLimitLiveCaller creates a new read-only instance of SimpleGatekeeperWithLimitLive, bound to a specific deployed contract.
func NewSimpleGatekeeperWithLimitLiveCaller(address common.Address, caller bind.ContractCaller) (*SimpleGatekeeperWithLimitLiveCaller, error) {
	contract, err := bindSimpleGatekeeperWithLimitLive(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitLiveCaller{contract: contract}, nil
}

// NewSimpleGatekeeperWithLimitLiveTransactor creates a new write-only instance of SimpleGatekeeperWithLimitLive, bound to a specific deployed contract.
func NewSimpleGatekeeperWithLimitLiveTransactor(address common.Address, transactor bind.ContractTransactor) (*SimpleGatekeeperWithLimitLiveTransactor, error) {
	contract, err := bindSimpleGatekeeperWithLimitLive(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitLiveTransactor{contract: contract}, nil
}

// NewSimpleGatekeeperWithLimitLiveFilterer creates a new log filterer instance of SimpleGatekeeperWithLimitLive, bound to a specific deployed contract.
func NewSimpleGatekeeperWithLimitLiveFilterer(address common.Address, filterer bind.ContractFilterer) (*SimpleGatekeeperWithLimitLiveFilterer, error) {
	contract, err := bindSimpleGatekeeperWithLimitLive(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitLiveFilterer{contract: contract}, nil
}

// bindSimpleGatekeeperWithLimitLive binds a generic wrapper to an already deployed contract.
func bindSimpleGatekeeperWithLimitLive(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleGatekeeperWithLimitLiveABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleGatekeeperWithLimitLive.Contract.SimpleGatekeeperWithLimitLiveCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.SimpleGatekeeperWithLimitLiveTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.SimpleGatekeeperWithLimitLiveTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleGatekeeperWithLimitLive.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.contract.Transact(opts, method, params...)
}

// GetFreezingTime is a free data retrieval call binding the contract method 0x36ab802e.
//
// Solidity: function GetFreezingTime() constant returns(uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveCaller) GetFreezingTime(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleGatekeeperWithLimitLive.contract.Call(opts, out, "GetFreezingTime")
	return *ret0, err
}

// GetFreezingTime is a free data retrieval call binding the contract method 0x36ab802e.
//
// Solidity: function GetFreezingTime() constant returns(uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveSession) GetFreezingTime() (*big.Int, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.GetFreezingTime(&_SimpleGatekeeperWithLimitLive.CallOpts)
}

// GetFreezingTime is a free data retrieval call binding the contract method 0x36ab802e.
//
// Solidity: function GetFreezingTime() constant returns(uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveCallerSession) GetFreezingTime() (*big.Int, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.GetFreezingTime(&_SimpleGatekeeperWithLimitLive.CallOpts)
}

// Keepers is a free data retrieval call binding the contract method 0x3bbd64bc.
//
// Solidity: function keepers( address) constant returns(dayLimit uint256, lastDay uint256, spentToday uint256, frozen bool)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveCaller) Keepers(opts *bind.CallOpts, arg0 common.Address) (struct {
	DayLimit   *big.Int
	LastDay    *big.Int
	SpentToday *big.Int
	Frozen     bool
}, error) {
	ret := new(struct {
		DayLimit   *big.Int
		LastDay    *big.Int
		SpentToday *big.Int
		Frozen     bool
	})
	out := ret
	err := _SimpleGatekeeperWithLimitLive.contract.Call(opts, out, "keepers", arg0)
	return *ret, err
}

// Keepers is a free data retrieval call binding the contract method 0x3bbd64bc.
//
// Solidity: function keepers( address) constant returns(dayLimit uint256, lastDay uint256, spentToday uint256, frozen bool)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveSession) Keepers(arg0 common.Address) (struct {
	DayLimit   *big.Int
	LastDay    *big.Int
	SpentToday *big.Int
	Frozen     bool
}, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.Keepers(&_SimpleGatekeeperWithLimitLive.CallOpts, arg0)
}

// Keepers is a free data retrieval call binding the contract method 0x3bbd64bc.
//
// Solidity: function keepers( address) constant returns(dayLimit uint256, lastDay uint256, spentToday uint256, frozen bool)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveCallerSession) Keepers(arg0 common.Address) (struct {
	DayLimit   *big.Int
	LastDay    *big.Int
	SpentToday *big.Int
	Frozen     bool
}, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.Keepers(&_SimpleGatekeeperWithLimitLive.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _SimpleGatekeeperWithLimitLive.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveSession) Owner() (common.Address, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.Owner(&_SimpleGatekeeperWithLimitLive.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveCallerSession) Owner() (common.Address, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.Owner(&_SimpleGatekeeperWithLimitLive.CallOpts)
}

// Paid is a free data retrieval call binding the contract method 0xadd89bb2.
//
// Solidity: function paid( bytes32) constant returns(commitTS uint256, paid bool, keeper address)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveCaller) Paid(opts *bind.CallOpts, arg0 [32]byte) (struct {
	CommitTS *big.Int
	Paid     bool
	Keeper   common.Address
}, error) {
	ret := new(struct {
		CommitTS *big.Int
		Paid     bool
		Keeper   common.Address
	})
	out := ret
	err := _SimpleGatekeeperWithLimitLive.contract.Call(opts, out, "paid", arg0)
	return *ret, err
}

// Paid is a free data retrieval call binding the contract method 0xadd89bb2.
//
// Solidity: function paid( bytes32) constant returns(commitTS uint256, paid bool, keeper address)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveSession) Paid(arg0 [32]byte) (struct {
	CommitTS *big.Int
	Paid     bool
	Keeper   common.Address
}, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.Paid(&_SimpleGatekeeperWithLimitLive.CallOpts, arg0)
}

// Paid is a free data retrieval call binding the contract method 0xadd89bb2.
//
// Solidity: function paid( bytes32) constant returns(commitTS uint256, paid bool, keeper address)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveCallerSession) Paid(arg0 [32]byte) (struct {
	CommitTS *big.Int
	Paid     bool
	Keeper   common.Address
}, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.Paid(&_SimpleGatekeeperWithLimitLive.CallOpts, arg0)
}

// TransactionAmount is a free data retrieval call binding the contract method 0xd942bffa.
//
// Solidity: function transactionAmount() constant returns(uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveCaller) TransactionAmount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleGatekeeperWithLimitLive.contract.Call(opts, out, "transactionAmount")
	return *ret0, err
}

// TransactionAmount is a free data retrieval call binding the contract method 0xd942bffa.
//
// Solidity: function transactionAmount() constant returns(uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveSession) TransactionAmount() (*big.Int, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.TransactionAmount(&_SimpleGatekeeperWithLimitLive.CallOpts)
}

// TransactionAmount is a free data retrieval call binding the contract method 0xd942bffa.
//
// Solidity: function transactionAmount() constant returns(uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveCallerSession) TransactionAmount() (*big.Int, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.TransactionAmount(&_SimpleGatekeeperWithLimitLive.CallOpts)
}

// ChangeKeeperLimit is a paid mutator transaction binding the contract method 0xad835c0b.
//
// Solidity: function ChangeKeeperLimit(_keeper address, _limit uint256) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactor) ChangeKeeperLimit(opts *bind.TransactOpts, _keeper common.Address, _limit *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.contract.Transact(opts, "ChangeKeeperLimit", _keeper, _limit)
}

// ChangeKeeperLimit is a paid mutator transaction binding the contract method 0xad835c0b.
//
// Solidity: function ChangeKeeperLimit(_keeper address, _limit uint256) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveSession) ChangeKeeperLimit(_keeper common.Address, _limit *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.ChangeKeeperLimit(&_SimpleGatekeeperWithLimitLive.TransactOpts, _keeper, _limit)
}

// ChangeKeeperLimit is a paid mutator transaction binding the contract method 0xad835c0b.
//
// Solidity: function ChangeKeeperLimit(_keeper address, _limit uint256) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactorSession) ChangeKeeperLimit(_keeper common.Address, _limit *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.ChangeKeeperLimit(&_SimpleGatekeeperWithLimitLive.TransactOpts, _keeper, _limit)
}

// FreezeKeeper is a paid mutator transaction binding the contract method 0xb38ad8e7.
//
// Solidity: function FreezeKeeper(_keeper address) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactor) FreezeKeeper(opts *bind.TransactOpts, _keeper common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.contract.Transact(opts, "FreezeKeeper", _keeper)
}

// FreezeKeeper is a paid mutator transaction binding the contract method 0xb38ad8e7.
//
// Solidity: function FreezeKeeper(_keeper address) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveSession) FreezeKeeper(_keeper common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.FreezeKeeper(&_SimpleGatekeeperWithLimitLive.TransactOpts, _keeper)
}

// FreezeKeeper is a paid mutator transaction binding the contract method 0xb38ad8e7.
//
// Solidity: function FreezeKeeper(_keeper address) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactorSession) FreezeKeeper(_keeper common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.FreezeKeeper(&_SimpleGatekeeperWithLimitLive.TransactOpts, _keeper)
}

// Payin is a paid mutator transaction binding the contract method 0x28f727f0.
//
// Solidity: function Payin(_value uint256) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactor) Payin(opts *bind.TransactOpts, _value *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.contract.Transact(opts, "Payin", _value)
}

// Payin is a paid mutator transaction binding the contract method 0x28f727f0.
//
// Solidity: function Payin(_value uint256) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveSession) Payin(_value *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.Payin(&_SimpleGatekeeperWithLimitLive.TransactOpts, _value)
}

// Payin is a paid mutator transaction binding the contract method 0x28f727f0.
//
// Solidity: function Payin(_value uint256) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactorSession) Payin(_value *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.Payin(&_SimpleGatekeeperWithLimitLive.TransactOpts, _value)
}

// Payout is a paid mutator transaction binding the contract method 0x634235fc.
//
// Solidity: function Payout(_to address, _value uint256, _txNumber uint256) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactor) Payout(opts *bind.TransactOpts, _to common.Address, _value *big.Int, _txNumber *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.contract.Transact(opts, "Payout", _to, _value, _txNumber)
}

// Payout is a paid mutator transaction binding the contract method 0x634235fc.
//
// Solidity: function Payout(_to address, _value uint256, _txNumber uint256) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveSession) Payout(_to common.Address, _value *big.Int, _txNumber *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.Payout(&_SimpleGatekeeperWithLimitLive.TransactOpts, _to, _value, _txNumber)
}

// Payout is a paid mutator transaction binding the contract method 0x634235fc.
//
// Solidity: function Payout(_to address, _value uint256, _txNumber uint256) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactorSession) Payout(_to common.Address, _value *big.Int, _txNumber *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.Payout(&_SimpleGatekeeperWithLimitLive.TransactOpts, _to, _value, _txNumber)
}

// SetFreezingTime is a paid mutator transaction binding the contract method 0xcc38d7ca.
//
// Solidity: function SetFreezingTime(_freezingTime uint256) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactor) SetFreezingTime(opts *bind.TransactOpts, _freezingTime *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.contract.Transact(opts, "SetFreezingTime", _freezingTime)
}

// SetFreezingTime is a paid mutator transaction binding the contract method 0xcc38d7ca.
//
// Solidity: function SetFreezingTime(_freezingTime uint256) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveSession) SetFreezingTime(_freezingTime *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.SetFreezingTime(&_SimpleGatekeeperWithLimitLive.TransactOpts, _freezingTime)
}

// SetFreezingTime is a paid mutator transaction binding the contract method 0xcc38d7ca.
//
// Solidity: function SetFreezingTime(_freezingTime uint256) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactorSession) SetFreezingTime(_freezingTime *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.SetFreezingTime(&_SimpleGatekeeperWithLimitLive.TransactOpts, _freezingTime)
}

// UnfreezeKeeper is a paid mutator transaction binding the contract method 0xe5837a7b.
//
// Solidity: function UnfreezeKeeper(_keeper address) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactor) UnfreezeKeeper(opts *bind.TransactOpts, _keeper common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.contract.Transact(opts, "UnfreezeKeeper", _keeper)
}

// UnfreezeKeeper is a paid mutator transaction binding the contract method 0xe5837a7b.
//
// Solidity: function UnfreezeKeeper(_keeper address) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveSession) UnfreezeKeeper(_keeper common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.UnfreezeKeeper(&_SimpleGatekeeperWithLimitLive.TransactOpts, _keeper)
}

// UnfreezeKeeper is a paid mutator transaction binding the contract method 0xe5837a7b.
//
// Solidity: function UnfreezeKeeper(_keeper address) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactorSession) UnfreezeKeeper(_keeper common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.UnfreezeKeeper(&_SimpleGatekeeperWithLimitLive.TransactOpts, _keeper)
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactor) Kill(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.contract.Transact(opts, "kill")
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveSession) Kill() (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.Kill(&_SimpleGatekeeperWithLimitLive.TransactOpts)
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactorSession) Kill() (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.Kill(&_SimpleGatekeeperWithLimitLive.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveSession) RenounceOwnership() (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.RenounceOwnership(&_SimpleGatekeeperWithLimitLive.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.RenounceOwnership(&_SimpleGatekeeperWithLimitLive.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactor) TransferOwnership(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.contract.Transact(opts, "transferOwnership", _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.TransferOwnership(&_SimpleGatekeeperWithLimitLive.TransactOpts, _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveTransactorSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimitLive.Contract.TransferOwnership(&_SimpleGatekeeperWithLimitLive.TransactOpts, _newOwner)
}

// SimpleGatekeeperWithLimitLiveCommitTxIterator is returned from FilterCommitTx and is used to iterate over the raw logs and unpacked data for CommitTx events raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLiveCommitTxIterator struct {
	Event *SimpleGatekeeperWithLimitLiveCommitTx // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitLiveCommitTxIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitLiveCommitTx)
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
		it.Event = new(SimpleGatekeeperWithLimitLiveCommitTx)
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
func (it *SimpleGatekeeperWithLimitLiveCommitTxIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitLiveCommitTxIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitLiveCommitTx represents a CommitTx event raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLiveCommitTx struct {
	From            common.Address
	TxNumber        *big.Int
	Value           *big.Int
	CommitTimestamp *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterCommitTx is a free log retrieval operation binding the contract event 0x65546c3bc3a77ffc91667da85018004299542e28a511328cfb4b3f86974902ee.
//
// Solidity: e CommitTx(from indexed address, txNumber indexed uint256, value indexed uint256, commitTimestamp uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) FilterCommitTx(opts *bind.FilterOpts, from []common.Address, txNumber []*big.Int, value []*big.Int) (*SimpleGatekeeperWithLimitLiveCommitTxIterator, error) {

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

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.FilterLogs(opts, "CommitTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitLiveCommitTxIterator{contract: _SimpleGatekeeperWithLimitLive.contract, event: "CommitTx", logs: logs, sub: sub}, nil
}

// WatchCommitTx is a free log subscription operation binding the contract event 0x65546c3bc3a77ffc91667da85018004299542e28a511328cfb4b3f86974902ee.
//
// Solidity: e CommitTx(from indexed address, txNumber indexed uint256, value indexed uint256, commitTimestamp uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) WatchCommitTx(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitLiveCommitTx, from []common.Address, txNumber []*big.Int, value []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.WatchLogs(opts, "CommitTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitLiveCommitTx)
				if err := _SimpleGatekeeperWithLimitLive.contract.UnpackLog(event, "CommitTx", log); err != nil {
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

// SimpleGatekeeperWithLimitLiveKeeperFreezedIterator is returned from FilterKeeperFreezed and is used to iterate over the raw logs and unpacked data for KeeperFreezed events raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLiveKeeperFreezedIterator struct {
	Event *SimpleGatekeeperWithLimitLiveKeeperFreezed // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitLiveKeeperFreezedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitLiveKeeperFreezed)
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
		it.Event = new(SimpleGatekeeperWithLimitLiveKeeperFreezed)
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
func (it *SimpleGatekeeperWithLimitLiveKeeperFreezedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitLiveKeeperFreezedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitLiveKeeperFreezed represents a KeeperFreezed event raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLiveKeeperFreezed struct {
	Keeper common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterKeeperFreezed is a free log retrieval operation binding the contract event 0xdf4868d2f39f6ab9f41b92c6917da5aec882c461ce7316bb62076865108502bd.
//
// Solidity: e KeeperFreezed(keeper indexed address)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) FilterKeeperFreezed(opts *bind.FilterOpts, keeper []common.Address) (*SimpleGatekeeperWithLimitLiveKeeperFreezedIterator, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.FilterLogs(opts, "KeeperFreezed", keeperRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitLiveKeeperFreezedIterator{contract: _SimpleGatekeeperWithLimitLive.contract, event: "KeeperFreezed", logs: logs, sub: sub}, nil
}

// WatchKeeperFreezed is a free log subscription operation binding the contract event 0xdf4868d2f39f6ab9f41b92c6917da5aec882c461ce7316bb62076865108502bd.
//
// Solidity: e KeeperFreezed(keeper indexed address)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) WatchKeeperFreezed(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitLiveKeeperFreezed, keeper []common.Address) (event.Subscription, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.WatchLogs(opts, "KeeperFreezed", keeperRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitLiveKeeperFreezed)
				if err := _SimpleGatekeeperWithLimitLive.contract.UnpackLog(event, "KeeperFreezed", log); err != nil {
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

// SimpleGatekeeperWithLimitLiveKeeperUnfreezedIterator is returned from FilterKeeperUnfreezed and is used to iterate over the raw logs and unpacked data for KeeperUnfreezed events raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLiveKeeperUnfreezedIterator struct {
	Event *SimpleGatekeeperWithLimitLiveKeeperUnfreezed // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitLiveKeeperUnfreezedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitLiveKeeperUnfreezed)
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
		it.Event = new(SimpleGatekeeperWithLimitLiveKeeperUnfreezed)
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
func (it *SimpleGatekeeperWithLimitLiveKeeperUnfreezedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitLiveKeeperUnfreezedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitLiveKeeperUnfreezed represents a KeeperUnfreezed event raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLiveKeeperUnfreezed struct {
	Keeper common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterKeeperUnfreezed is a free log retrieval operation binding the contract event 0xbbe17a7427b5192903e1b3f0f2b6ef8b2a1af9b33e1079faf8f8383f2fb63b53.
//
// Solidity: e KeeperUnfreezed(keeper indexed address)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) FilterKeeperUnfreezed(opts *bind.FilterOpts, keeper []common.Address) (*SimpleGatekeeperWithLimitLiveKeeperUnfreezedIterator, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.FilterLogs(opts, "KeeperUnfreezed", keeperRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitLiveKeeperUnfreezedIterator{contract: _SimpleGatekeeperWithLimitLive.contract, event: "KeeperUnfreezed", logs: logs, sub: sub}, nil
}

// WatchKeeperUnfreezed is a free log subscription operation binding the contract event 0xbbe17a7427b5192903e1b3f0f2b6ef8b2a1af9b33e1079faf8f8383f2fb63b53.
//
// Solidity: e KeeperUnfreezed(keeper indexed address)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) WatchKeeperUnfreezed(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitLiveKeeperUnfreezed, keeper []common.Address) (event.Subscription, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.WatchLogs(opts, "KeeperUnfreezed", keeperRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitLiveKeeperUnfreezed)
				if err := _SimpleGatekeeperWithLimitLive.contract.UnpackLog(event, "KeeperUnfreezed", log); err != nil {
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

// SimpleGatekeeperWithLimitLiveLimitChangedIterator is returned from FilterLimitChanged and is used to iterate over the raw logs and unpacked data for LimitChanged events raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLiveLimitChangedIterator struct {
	Event *SimpleGatekeeperWithLimitLiveLimitChanged // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitLiveLimitChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitLiveLimitChanged)
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
		it.Event = new(SimpleGatekeeperWithLimitLiveLimitChanged)
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
func (it *SimpleGatekeeperWithLimitLiveLimitChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitLiveLimitChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitLiveLimitChanged represents a LimitChanged event raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLiveLimitChanged struct {
	Keeper   common.Address
	DayLimit *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterLimitChanged is a free log retrieval operation binding the contract event 0xef9c668177207fb68ca5e3894a1efacebb659762b27a737fde58ceebc4f30ad3.
//
// Solidity: e LimitChanged(keeper indexed address, dayLimit indexed uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) FilterLimitChanged(opts *bind.FilterOpts, keeper []common.Address, dayLimit []*big.Int) (*SimpleGatekeeperWithLimitLiveLimitChangedIterator, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}
	var dayLimitRule []interface{}
	for _, dayLimitItem := range dayLimit {
		dayLimitRule = append(dayLimitRule, dayLimitItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.FilterLogs(opts, "LimitChanged", keeperRule, dayLimitRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitLiveLimitChangedIterator{contract: _SimpleGatekeeperWithLimitLive.contract, event: "LimitChanged", logs: logs, sub: sub}, nil
}

// WatchLimitChanged is a free log subscription operation binding the contract event 0xef9c668177207fb68ca5e3894a1efacebb659762b27a737fde58ceebc4f30ad3.
//
// Solidity: e LimitChanged(keeper indexed address, dayLimit indexed uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) WatchLimitChanged(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitLiveLimitChanged, keeper []common.Address, dayLimit []*big.Int) (event.Subscription, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}
	var dayLimitRule []interface{}
	for _, dayLimitItem := range dayLimit {
		dayLimitRule = append(dayLimitRule, dayLimitItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.WatchLogs(opts, "LimitChanged", keeperRule, dayLimitRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitLiveLimitChanged)
				if err := _SimpleGatekeeperWithLimitLive.contract.UnpackLog(event, "LimitChanged", log); err != nil {
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

// SimpleGatekeeperWithLimitLiveOwnershipRenouncedIterator is returned from FilterOwnershipRenounced and is used to iterate over the raw logs and unpacked data for OwnershipRenounced events raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLiveOwnershipRenouncedIterator struct {
	Event *SimpleGatekeeperWithLimitLiveOwnershipRenounced // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitLiveOwnershipRenouncedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitLiveOwnershipRenounced)
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
		it.Event = new(SimpleGatekeeperWithLimitLiveOwnershipRenounced)
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
func (it *SimpleGatekeeperWithLimitLiveOwnershipRenouncedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitLiveOwnershipRenouncedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitLiveOwnershipRenounced represents a OwnershipRenounced event raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLiveOwnershipRenounced struct {
	PreviousOwner common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipRenounced is a free log retrieval operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) FilterOwnershipRenounced(opts *bind.FilterOpts, previousOwner []common.Address) (*SimpleGatekeeperWithLimitLiveOwnershipRenouncedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.FilterLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitLiveOwnershipRenouncedIterator{contract: _SimpleGatekeeperWithLimitLive.contract, event: "OwnershipRenounced", logs: logs, sub: sub}, nil
}

// WatchOwnershipRenounced is a free log subscription operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) WatchOwnershipRenounced(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitLiveOwnershipRenounced, previousOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.WatchLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitLiveOwnershipRenounced)
				if err := _SimpleGatekeeperWithLimitLive.contract.UnpackLog(event, "OwnershipRenounced", log); err != nil {
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

// SimpleGatekeeperWithLimitLiveOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLiveOwnershipTransferredIterator struct {
	Event *SimpleGatekeeperWithLimitLiveOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitLiveOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitLiveOwnershipTransferred)
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
		it.Event = new(SimpleGatekeeperWithLimitLiveOwnershipTransferred)
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
func (it *SimpleGatekeeperWithLimitLiveOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitLiveOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitLiveOwnershipTransferred represents a OwnershipTransferred event raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLiveOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SimpleGatekeeperWithLimitLiveOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitLiveOwnershipTransferredIterator{contract: _SimpleGatekeeperWithLimitLive.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitLiveOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitLiveOwnershipTransferred)
				if err := _SimpleGatekeeperWithLimitLive.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// SimpleGatekeeperWithLimitLivePayinTxIterator is returned from FilterPayinTx and is used to iterate over the raw logs and unpacked data for PayinTx events raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLivePayinTxIterator struct {
	Event *SimpleGatekeeperWithLimitLivePayinTx // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitLivePayinTxIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitLivePayinTx)
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
		it.Event = new(SimpleGatekeeperWithLimitLivePayinTx)
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
func (it *SimpleGatekeeperWithLimitLivePayinTxIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitLivePayinTxIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitLivePayinTx represents a PayinTx event raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLivePayinTx struct {
	From     common.Address
	TxNumber *big.Int
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPayinTx is a free log retrieval operation binding the contract event 0x14312725abbc46ad798bc078b2663e1fcbace97be0247cd177176f3b4df2538e.
//
// Solidity: e PayinTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) FilterPayinTx(opts *bind.FilterOpts, from []common.Address, txNumber []*big.Int, value []*big.Int) (*SimpleGatekeeperWithLimitLivePayinTxIterator, error) {

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

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.FilterLogs(opts, "PayinTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitLivePayinTxIterator{contract: _SimpleGatekeeperWithLimitLive.contract, event: "PayinTx", logs: logs, sub: sub}, nil
}

// WatchPayinTx is a free log subscription operation binding the contract event 0x14312725abbc46ad798bc078b2663e1fcbace97be0247cd177176f3b4df2538e.
//
// Solidity: e PayinTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) WatchPayinTx(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitLivePayinTx, from []common.Address, txNumber []*big.Int, value []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.WatchLogs(opts, "PayinTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitLivePayinTx)
				if err := _SimpleGatekeeperWithLimitLive.contract.UnpackLog(event, "PayinTx", log); err != nil {
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

// SimpleGatekeeperWithLimitLivePayoutTxIterator is returned from FilterPayoutTx and is used to iterate over the raw logs and unpacked data for PayoutTx events raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLivePayoutTxIterator struct {
	Event *SimpleGatekeeperWithLimitLivePayoutTx // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitLivePayoutTxIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitLivePayoutTx)
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
		it.Event = new(SimpleGatekeeperWithLimitLivePayoutTx)
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
func (it *SimpleGatekeeperWithLimitLivePayoutTxIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitLivePayoutTxIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitLivePayoutTx represents a PayoutTx event raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLivePayoutTx struct {
	From     common.Address
	TxNumber *big.Int
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPayoutTx is a free log retrieval operation binding the contract event 0x731af16374848c2c73a6154fd410cb421138e7db45c5a904e5a475c756faa8d9.
//
// Solidity: e PayoutTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) FilterPayoutTx(opts *bind.FilterOpts, from []common.Address, txNumber []*big.Int, value []*big.Int) (*SimpleGatekeeperWithLimitLivePayoutTxIterator, error) {

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

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.FilterLogs(opts, "PayoutTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitLivePayoutTxIterator{contract: _SimpleGatekeeperWithLimitLive.contract, event: "PayoutTx", logs: logs, sub: sub}, nil
}

// WatchPayoutTx is a free log subscription operation binding the contract event 0x731af16374848c2c73a6154fd410cb421138e7db45c5a904e5a475c756faa8d9.
//
// Solidity: e PayoutTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) WatchPayoutTx(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitLivePayoutTx, from []common.Address, txNumber []*big.Int, value []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.WatchLogs(opts, "PayoutTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitLivePayoutTx)
				if err := _SimpleGatekeeperWithLimitLive.contract.UnpackLog(event, "PayoutTx", log); err != nil {
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

// SimpleGatekeeperWithLimitLiveSuicideIterator is returned from FilterSuicide and is used to iterate over the raw logs and unpacked data for Suicide events raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLiveSuicideIterator struct {
	Event *SimpleGatekeeperWithLimitLiveSuicide // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitLiveSuicideIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitLiveSuicide)
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
		it.Event = new(SimpleGatekeeperWithLimitLiveSuicide)
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
func (it *SimpleGatekeeperWithLimitLiveSuicideIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitLiveSuicideIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitLiveSuicide represents a Suicide event raised by the SimpleGatekeeperWithLimitLive contract.
type SimpleGatekeeperWithLimitLiveSuicide struct {
	Block *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterSuicide is a free log retrieval operation binding the contract event 0xa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde.
//
// Solidity: e Suicide(block uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) FilterSuicide(opts *bind.FilterOpts) (*SimpleGatekeeperWithLimitLiveSuicideIterator, error) {

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.FilterLogs(opts, "Suicide")
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitLiveSuicideIterator{contract: _SimpleGatekeeperWithLimitLive.contract, event: "Suicide", logs: logs, sub: sub}, nil
}

// WatchSuicide is a free log subscription operation binding the contract event 0xa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde.
//
// Solidity: e Suicide(block uint256)
func (_SimpleGatekeeperWithLimitLive *SimpleGatekeeperWithLimitLiveFilterer) WatchSuicide(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitLiveSuicide) (event.Subscription, error) {

	logs, sub, err := _SimpleGatekeeperWithLimitLive.contract.WatchLogs(opts, "Suicide")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitLiveSuicide)
				if err := _SimpleGatekeeperWithLimitLive.contract.UnpackLog(event, "Suicide", log); err != nil {
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
