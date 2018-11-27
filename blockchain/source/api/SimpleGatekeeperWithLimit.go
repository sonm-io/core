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

// SimpleGatekeeperWithLimitABI is the input ABI used to generate the binding from.
const SimpleGatekeeperWithLimitABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"keepers\",\"outputs\":[{\"name\":\"dayLimit\",\"type\":\"uint256\"},{\"name\":\"lastDay\",\"type\":\"uint256\"},{\"name\":\"spentToday\",\"type\":\"uint256\"},{\"name\":\"frozen\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"paid\",\"outputs\":[{\"name\":\"commitTS\",\"type\":\"uint256\"},{\"name\":\"paid\",\"type\":\"bool\"},{\"name\":\"keeper\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"transactionAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"commissionBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"commission\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"},{\"name\":\"_freezingTime\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"txNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"PayinTx\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"txNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"commitTimestamp\",\"type\":\"uint256\"}],\"name\":\"CommitTx\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"txNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"PayoutTx\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"block\",\"type\":\"uint256\"}],\"name\":\"Suicide\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"keeper\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"dayLimit\",\"type\":\"uint256\"}],\"name\":\"LimitChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"keeper\",\"type\":\"address\"}],\"name\":\"KeeperFreezed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"keeper\",\"type\":\"address\"}],\"name\":\"KeeperUnfreezed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"commission\",\"type\":\"uint256\"}],\"name\":\"CommissionChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"}],\"name\":\"OwnershipRenounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_keeper\",\"type\":\"address\"},{\"name\":\"_limit\",\"type\":\"uint256\"}],\"name\":\"ChangeKeeperLimit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_keeper\",\"type\":\"address\"}],\"name\":\"FreezeKeeper\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_keeper\",\"type\":\"address\"}],\"name\":\"UnfreezeKeeper\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Payin\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_value\",\"type\":\"uint256\"},{\"name\":\"_target\",\"type\":\"address\"}],\"name\":\"PayinTargeted\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"},{\"name\":\"_txNumber\",\"type\":\"uint256\"}],\"name\":\"Payout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_freezingTime\",\"type\":\"uint256\"}],\"name\":\"SetFreezingTime\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetFreezingTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_commission\",\"type\":\"uint256\"}],\"name\":\"SetCommission\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"GetCommission\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"TransferCommission\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"kill\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// SimpleGatekeeperWithLimitBin is the compiled bytecode used for deploying new contracts.
const SimpleGatekeeperWithLimitBin = `0x608060405260006003556000600455600060055534801561001f57600080fd5b50604051604080610eea8339810160405280516020909101516000805460018054600160a060020a03909516600160a060020a0319958616179055831633908117909316909217909155600755610e6f8061007b6000396000f3006080604052600436106101115763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306d8e8b1811461011657806328f727f01461012d57806336ab802e146101455780633bbd64bc1461016c57806341c0e1b5146101b557806358712633146101ca578063634235fc146101df5780636ea5803114610206578063715018a61461021e5780638da5cb5b14610233578063ad835c0b14610264578063add89bb214610288578063b38ad8e7146102c7578063cc38d7ca146102e8578063d942bffa14610300578063dcf1a9ef14610315578063e14891911461032a578063e3fcd18e1461033f578063e5837a7b14610363578063f2fde38b14610384575b600080fd5b34801561012257600080fd5b5061012b6103a5565b005b34801561013957600080fd5b5061012b600435610473565b34801561015157600080fd5b5061015a610480565b60408051918252519081900360200190f35b34801561017857600080fd5b5061018d600160a060020a0360043516610486565b6040805194855260208501939093528383019190915215156060830152519081900360800190f35b3480156101c157600080fd5b5061012b6104b2565b3480156101d657600080fd5b5061015a610647565b3480156101eb57600080fd5b5061012b600160a060020a036004351660243560443561064d565b34801561021257600080fd5b5061012b6004356108d5565b34801561022a57600080fd5b5061012b61091f565b34801561023f57600080fd5b5061024861098b565b60408051600160a060020a039092168252519081900360200190f35b34801561027057600080fd5b5061012b600160a060020a036004351660243561099a565b34801561029457600080fd5b506102a06004356109f8565b604080519384529115156020840152600160a060020a031682820152519081900360600190f35b3480156102d357600080fd5b5061012b600160a060020a0360043516610a24565b3480156102f457600080fd5b5061012b600435610ab0565b34801561030c57600080fd5b5061015a610acc565b34801561032157600080fd5b5061015a610ad2565b34801561033657600080fd5b5061015a610ad8565b34801561034b57600080fd5b5061012b600435600160a060020a0360243516610ade565b34801561036f57600080fd5b5061012b600160a060020a0360043516610c06565b34801561039057600080fd5b5061012b600160a060020a0360043516610c8c565b600054600160a060020a031633146103bc57600080fd5b60015460008054600554604080517fa9059cbb000000000000000000000000000000000000000000000000000000008152600160a060020a039384166004820152602481019290925251919093169263a9059cbb9260448083019360209390929083900390910190829087803b15801561043557600080fd5b505af1158015610449573d6000803e3d6000fd5b505050506040513d602081101561045f57600080fd5b5051151561046c57600080fd5b6000600555565b61047d8133610ade565b50565b60075490565b600260208190526000918252604090912080546001820154928201546003909201549092919060ff1684565b600054600160a060020a031633146104c957600080fd5b60015460008054604080517f70a082310000000000000000000000000000000000000000000000000000000081523060048201529051600160a060020a039485169463a9059cbb9493169285926370a082319260248083019360209383900390910190829087803b15801561053d57600080fd5b505af1158015610551573d6000803e3d6000fd5b505050506040513d602081101561056757600080fd5b5051604080517c010000000000000000000000000000000000000000000000000000000063ffffffff8616028152600160a060020a03909316600484015260248301919091525160448083019260209291908290030181600087803b1580156105cf57600080fd5b505af11580156105e3573d6000803e3d6000fd5b505050506040513d60208110156105f957600080fd5b5051151561060657600080fd5b6040805142815290517fa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde9181900360200190a1600054600160a060020a0316ff5b60045490565b3360009081526002602052604081206003015460ff161561066d57600080fd5b336000908152600260205260408120541161068757600080fd5b50604080516c01000000000000000000000000600160a060020a0386160281526014810183905260348101849052815190819003605401902060008181526006602052919091206001015460ff16156106df57600080fd5b600081815260066020526040902054151561078b576106fe3384610cac565b151561070957600080fd5b600081815260066020908152604091829020428082556001909101805474ffffffffffffffffffffffffffffffffffffffff00191633610100021790558251908152915185928592600160a060020a038916927f65546c3bc3a77ffc91667da85018004299542e28a511328cfb4b3f86974902ee9281900390910190a46108cf565b6000818152600660205260409020600101546101009004600160a060020a031633146107b657600080fd5b60075460008281526006602052604090205442910111156107d657600080fd5b600154604080517fa9059cbb000000000000000000000000000000000000000000000000000000008152600160a060020a038781166004830152602482018790529151919092169163a9059cbb9160448083019260209291908290030181600087803b15801561084557600080fd5b505af1158015610859573d6000803e3d6000fd5b505050506040513d602081101561086f57600080fd5b5051151561087c57600080fd5b6000818152600660205260408082206001908101805460ff191690911790555184918491600160a060020a038816917f731af16374848c2c73a6154fd410cb421138e7db45c5a904e5a475c756faa8d991a45b50505050565b600054600160a060020a031633146108ec57600080fd5b600481905560405181907f839e4456845dbc05c7d8638cf0b0976161331b5f9163980d71d9a6444a326c6190600090a250565b600054600160a060020a0316331461093657600080fd5b60008054604051600160a060020a03909116917ff8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c6482091a26000805473ffffffffffffffffffffffffffffffffffffffff19169055565b600054600160a060020a031681565b600054600160a060020a031633146109b157600080fd5b600160a060020a038216600081815260026020526040808220849055518392917fef9c668177207fb68ca5e3894a1efacebb659762b27a737fde58ceebc4f30ad391a35050565b6006602052600090815260409020805460019091015460ff8116906101009004600160a060020a031683565b3360009081526002602052604081205411610a3e57600080fd5b600160a060020a03811660009081526002602052604081205411610a6157600080fd5b600160a060020a038116600081815260026020526040808220600301805460ff19166001179055517fdf4868d2f39f6ab9f41b92c6917da5aec882c461ce7316bb62076865108502bd9190a250565b600054600160a060020a03163314610ac757600080fd5b600755565b60035481565b60055481565b60045481565b6004548211610aec57600080fd5b600154604080517f23b872dd000000000000000000000000000000000000000000000000000000008152336004820152306024820152604481018590529051600160a060020a03909216916323b872dd916064808201926020929091908290030181600087803b158015610b5f57600080fd5b505af1158015610b73573d6000803e3d6000fd5b505050506040513d6020811015610b8957600080fd5b50511515610b9657600080fd5b600380546001019055600454600554610bb49163ffffffff610d9e16565b600555600454610bcb90839063ffffffff610dab16565b600354604051600160a060020a038416907f14312725abbc46ad798bc078b2663e1fcbace97be0247cd177176f3b4df2538e90600090a45050565b600054600160a060020a03163314610c1d57600080fd5b600160a060020a03811660009081526002602052604081205411610c4057600080fd5b600160a060020a038116600081815260026020526040808220600301805460ff19169055517fbbe17a7427b5192903e1b3f0f2b6ef8b2a1af9b33e1079faf8f8383f2fb63b539190a250565b600054600160a060020a03163314610ca357600080fd5b61047d81610dbd565b600160a060020a038216600090815260026020526040812060010154610cd0610e3a565b1115610d1757600160a060020a038316600090815260026020819052604082200155610cfa610e3a565b600160a060020a0384166000908152600260205260409020600101555b600160a060020a0383166000908152600260208190526040909120015482810110801590610d655750600160a060020a03831660009081526002602081905260409091208054910154830111155b15610d945750600160a060020a0382166000908152600260208190526040909120018054820190556001610d98565b5060005b92915050565b81810182811015610d9857fe5b600082821115610db757fe5b50900390565b600160a060020a0381161515610dd257600080fd5b60008054604051600160a060020a03808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a36000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0392909216919091179055565b620151804204905600a165627a7a723058203cff4e898fe0b069055c08319c3c3d2db6e51a57270c6d550f468526830d13e20029`

// DeploySimpleGatekeeperWithLimit deploys a new Ethereum contract, binding an instance of SimpleGatekeeperWithLimit to it.
func DeploySimpleGatekeeperWithLimit(auth *bind.TransactOpts, backend bind.ContractBackend, _token common.Address, _freezingTime *big.Int) (common.Address, *types.Transaction, *SimpleGatekeeperWithLimit, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleGatekeeperWithLimitABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SimpleGatekeeperWithLimitBin), backend, _token, _freezingTime)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SimpleGatekeeperWithLimit{SimpleGatekeeperWithLimitCaller: SimpleGatekeeperWithLimitCaller{contract: contract}, SimpleGatekeeperWithLimitTransactor: SimpleGatekeeperWithLimitTransactor{contract: contract}, SimpleGatekeeperWithLimitFilterer: SimpleGatekeeperWithLimitFilterer{contract: contract}}, nil
}

// SimpleGatekeeperWithLimit is an auto generated Go binding around an Ethereum contract.
type SimpleGatekeeperWithLimit struct {
	SimpleGatekeeperWithLimitCaller     // Read-only binding to the contract
	SimpleGatekeeperWithLimitTransactor // Write-only binding to the contract
	SimpleGatekeeperWithLimitFilterer   // Log filterer for contract events
}

// SimpleGatekeeperWithLimitCaller is an auto generated read-only Go binding around an Ethereum contract.
type SimpleGatekeeperWithLimitCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleGatekeeperWithLimitTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SimpleGatekeeperWithLimitTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleGatekeeperWithLimitFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SimpleGatekeeperWithLimitFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleGatekeeperWithLimitSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SimpleGatekeeperWithLimitSession struct {
	Contract     *SimpleGatekeeperWithLimit // Generic contract binding to set the session for
	CallOpts     bind.CallOpts              // Call options to use throughout this session
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// SimpleGatekeeperWithLimitCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SimpleGatekeeperWithLimitCallerSession struct {
	Contract *SimpleGatekeeperWithLimitCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                    // Call options to use throughout this session
}

// SimpleGatekeeperWithLimitTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SimpleGatekeeperWithLimitTransactorSession struct {
	Contract     *SimpleGatekeeperWithLimitTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                    // Transaction auth options to use throughout this session
}

// SimpleGatekeeperWithLimitRaw is an auto generated low-level Go binding around an Ethereum contract.
type SimpleGatekeeperWithLimitRaw struct {
	Contract *SimpleGatekeeperWithLimit // Generic contract binding to access the raw methods on
}

// SimpleGatekeeperWithLimitCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SimpleGatekeeperWithLimitCallerRaw struct {
	Contract *SimpleGatekeeperWithLimitCaller // Generic read-only contract binding to access the raw methods on
}

// SimpleGatekeeperWithLimitTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SimpleGatekeeperWithLimitTransactorRaw struct {
	Contract *SimpleGatekeeperWithLimitTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSimpleGatekeeperWithLimit creates a new instance of SimpleGatekeeperWithLimit, bound to a specific deployed contract.
func NewSimpleGatekeeperWithLimit(address common.Address, backend bind.ContractBackend) (*SimpleGatekeeperWithLimit, error) {
	contract, err := bindSimpleGatekeeperWithLimit(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimit{SimpleGatekeeperWithLimitCaller: SimpleGatekeeperWithLimitCaller{contract: contract}, SimpleGatekeeperWithLimitTransactor: SimpleGatekeeperWithLimitTransactor{contract: contract}, SimpleGatekeeperWithLimitFilterer: SimpleGatekeeperWithLimitFilterer{contract: contract}}, nil
}

// NewSimpleGatekeeperWithLimitCaller creates a new read-only instance of SimpleGatekeeperWithLimit, bound to a specific deployed contract.
func NewSimpleGatekeeperWithLimitCaller(address common.Address, caller bind.ContractCaller) (*SimpleGatekeeperWithLimitCaller, error) {
	contract, err := bindSimpleGatekeeperWithLimit(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitCaller{contract: contract}, nil
}

// NewSimpleGatekeeperWithLimitTransactor creates a new write-only instance of SimpleGatekeeperWithLimit, bound to a specific deployed contract.
func NewSimpleGatekeeperWithLimitTransactor(address common.Address, transactor bind.ContractTransactor) (*SimpleGatekeeperWithLimitTransactor, error) {
	contract, err := bindSimpleGatekeeperWithLimit(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitTransactor{contract: contract}, nil
}

// NewSimpleGatekeeperWithLimitFilterer creates a new log filterer instance of SimpleGatekeeperWithLimit, bound to a specific deployed contract.
func NewSimpleGatekeeperWithLimitFilterer(address common.Address, filterer bind.ContractFilterer) (*SimpleGatekeeperWithLimitFilterer, error) {
	contract, err := bindSimpleGatekeeperWithLimit(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitFilterer{contract: contract}, nil
}

// bindSimpleGatekeeperWithLimit binds a generic wrapper to an already deployed contract.
func bindSimpleGatekeeperWithLimit(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleGatekeeperWithLimitABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleGatekeeperWithLimit.Contract.SimpleGatekeeperWithLimitCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.SimpleGatekeeperWithLimitTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.SimpleGatekeeperWithLimitTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleGatekeeperWithLimit.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.contract.Transact(opts, method, params...)
}

// GetCommission is a free data retrieval call binding the contract method 0x58712633.
//
// Solidity: function GetCommission() constant returns(uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCaller) GetCommission(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleGatekeeperWithLimit.contract.Call(opts, out, "GetCommission")
	return *ret0, err
}

// GetCommission is a free data retrieval call binding the contract method 0x58712633.
//
// Solidity: function GetCommission() constant returns(uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) GetCommission() (*big.Int, error) {
	return _SimpleGatekeeperWithLimit.Contract.GetCommission(&_SimpleGatekeeperWithLimit.CallOpts)
}

// GetCommission is a free data retrieval call binding the contract method 0x58712633.
//
// Solidity: function GetCommission() constant returns(uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCallerSession) GetCommission() (*big.Int, error) {
	return _SimpleGatekeeperWithLimit.Contract.GetCommission(&_SimpleGatekeeperWithLimit.CallOpts)
}

// GetFreezingTime is a free data retrieval call binding the contract method 0x36ab802e.
//
// Solidity: function GetFreezingTime() constant returns(uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCaller) GetFreezingTime(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleGatekeeperWithLimit.contract.Call(opts, out, "GetFreezingTime")
	return *ret0, err
}

// GetFreezingTime is a free data retrieval call binding the contract method 0x36ab802e.
//
// Solidity: function GetFreezingTime() constant returns(uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) GetFreezingTime() (*big.Int, error) {
	return _SimpleGatekeeperWithLimit.Contract.GetFreezingTime(&_SimpleGatekeeperWithLimit.CallOpts)
}

// GetFreezingTime is a free data retrieval call binding the contract method 0x36ab802e.
//
// Solidity: function GetFreezingTime() constant returns(uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCallerSession) GetFreezingTime() (*big.Int, error) {
	return _SimpleGatekeeperWithLimit.Contract.GetFreezingTime(&_SimpleGatekeeperWithLimit.CallOpts)
}

// Commission is a free data retrieval call binding the contract method 0xe1489191.
//
// Solidity: function commission() constant returns(uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCaller) Commission(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleGatekeeperWithLimit.contract.Call(opts, out, "commission")
	return *ret0, err
}

// Commission is a free data retrieval call binding the contract method 0xe1489191.
//
// Solidity: function commission() constant returns(uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) Commission() (*big.Int, error) {
	return _SimpleGatekeeperWithLimit.Contract.Commission(&_SimpleGatekeeperWithLimit.CallOpts)
}

// Commission is a free data retrieval call binding the contract method 0xe1489191.
//
// Solidity: function commission() constant returns(uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCallerSession) Commission() (*big.Int, error) {
	return _SimpleGatekeeperWithLimit.Contract.Commission(&_SimpleGatekeeperWithLimit.CallOpts)
}

// CommissionBalance is a free data retrieval call binding the contract method 0xdcf1a9ef.
//
// Solidity: function commissionBalance() constant returns(uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCaller) CommissionBalance(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleGatekeeperWithLimit.contract.Call(opts, out, "commissionBalance")
	return *ret0, err
}

// CommissionBalance is a free data retrieval call binding the contract method 0xdcf1a9ef.
//
// Solidity: function commissionBalance() constant returns(uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) CommissionBalance() (*big.Int, error) {
	return _SimpleGatekeeperWithLimit.Contract.CommissionBalance(&_SimpleGatekeeperWithLimit.CallOpts)
}

// CommissionBalance is a free data retrieval call binding the contract method 0xdcf1a9ef.
//
// Solidity: function commissionBalance() constant returns(uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCallerSession) CommissionBalance() (*big.Int, error) {
	return _SimpleGatekeeperWithLimit.Contract.CommissionBalance(&_SimpleGatekeeperWithLimit.CallOpts)
}

// Keepers is a free data retrieval call binding the contract method 0x3bbd64bc.
//
// Solidity: function keepers( address) constant returns(dayLimit uint256, lastDay uint256, spentToday uint256, frozen bool)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCaller) Keepers(opts *bind.CallOpts, arg0 common.Address) (struct {
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
	err := _SimpleGatekeeperWithLimit.contract.Call(opts, out, "keepers", arg0)
	return *ret, err
}

// Keepers is a free data retrieval call binding the contract method 0x3bbd64bc.
//
// Solidity: function keepers( address) constant returns(dayLimit uint256, lastDay uint256, spentToday uint256, frozen bool)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) Keepers(arg0 common.Address) (struct {
	DayLimit   *big.Int
	LastDay    *big.Int
	SpentToday *big.Int
	Frozen     bool
}, error) {
	return _SimpleGatekeeperWithLimit.Contract.Keepers(&_SimpleGatekeeperWithLimit.CallOpts, arg0)
}

// Keepers is a free data retrieval call binding the contract method 0x3bbd64bc.
//
// Solidity: function keepers( address) constant returns(dayLimit uint256, lastDay uint256, spentToday uint256, frozen bool)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCallerSession) Keepers(arg0 common.Address) (struct {
	DayLimit   *big.Int
	LastDay    *big.Int
	SpentToday *big.Int
	Frozen     bool
}, error) {
	return _SimpleGatekeeperWithLimit.Contract.Keepers(&_SimpleGatekeeperWithLimit.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _SimpleGatekeeperWithLimit.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) Owner() (common.Address, error) {
	return _SimpleGatekeeperWithLimit.Contract.Owner(&_SimpleGatekeeperWithLimit.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCallerSession) Owner() (common.Address, error) {
	return _SimpleGatekeeperWithLimit.Contract.Owner(&_SimpleGatekeeperWithLimit.CallOpts)
}

// Paid is a free data retrieval call binding the contract method 0xadd89bb2.
//
// Solidity: function paid( bytes32) constant returns(commitTS uint256, paid bool, keeper address)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCaller) Paid(opts *bind.CallOpts, arg0 [32]byte) (struct {
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
	err := _SimpleGatekeeperWithLimit.contract.Call(opts, out, "paid", arg0)
	return *ret, err
}

// Paid is a free data retrieval call binding the contract method 0xadd89bb2.
//
// Solidity: function paid( bytes32) constant returns(commitTS uint256, paid bool, keeper address)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) Paid(arg0 [32]byte) (struct {
	CommitTS *big.Int
	Paid     bool
	Keeper   common.Address
}, error) {
	return _SimpleGatekeeperWithLimit.Contract.Paid(&_SimpleGatekeeperWithLimit.CallOpts, arg0)
}

// Paid is a free data retrieval call binding the contract method 0xadd89bb2.
//
// Solidity: function paid( bytes32) constant returns(commitTS uint256, paid bool, keeper address)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCallerSession) Paid(arg0 [32]byte) (struct {
	CommitTS *big.Int
	Paid     bool
	Keeper   common.Address
}, error) {
	return _SimpleGatekeeperWithLimit.Contract.Paid(&_SimpleGatekeeperWithLimit.CallOpts, arg0)
}

// TransactionAmount is a free data retrieval call binding the contract method 0xd942bffa.
//
// Solidity: function transactionAmount() constant returns(uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCaller) TransactionAmount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleGatekeeperWithLimit.contract.Call(opts, out, "transactionAmount")
	return *ret0, err
}

// TransactionAmount is a free data retrieval call binding the contract method 0xd942bffa.
//
// Solidity: function transactionAmount() constant returns(uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) TransactionAmount() (*big.Int, error) {
	return _SimpleGatekeeperWithLimit.Contract.TransactionAmount(&_SimpleGatekeeperWithLimit.CallOpts)
}

// TransactionAmount is a free data retrieval call binding the contract method 0xd942bffa.
//
// Solidity: function transactionAmount() constant returns(uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitCallerSession) TransactionAmount() (*big.Int, error) {
	return _SimpleGatekeeperWithLimit.Contract.TransactionAmount(&_SimpleGatekeeperWithLimit.CallOpts)
}

// ChangeKeeperLimit is a paid mutator transaction binding the contract method 0xad835c0b.
//
// Solidity: function ChangeKeeperLimit(_keeper address, _limit uint256) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactor) ChangeKeeperLimit(opts *bind.TransactOpts, _keeper common.Address, _limit *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.contract.Transact(opts, "ChangeKeeperLimit", _keeper, _limit)
}

// ChangeKeeperLimit is a paid mutator transaction binding the contract method 0xad835c0b.
//
// Solidity: function ChangeKeeperLimit(_keeper address, _limit uint256) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) ChangeKeeperLimit(_keeper common.Address, _limit *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.ChangeKeeperLimit(&_SimpleGatekeeperWithLimit.TransactOpts, _keeper, _limit)
}

// ChangeKeeperLimit is a paid mutator transaction binding the contract method 0xad835c0b.
//
// Solidity: function ChangeKeeperLimit(_keeper address, _limit uint256) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactorSession) ChangeKeeperLimit(_keeper common.Address, _limit *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.ChangeKeeperLimit(&_SimpleGatekeeperWithLimit.TransactOpts, _keeper, _limit)
}

// FreezeKeeper is a paid mutator transaction binding the contract method 0xb38ad8e7.
//
// Solidity: function FreezeKeeper(_keeper address) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactor) FreezeKeeper(opts *bind.TransactOpts, _keeper common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.contract.Transact(opts, "FreezeKeeper", _keeper)
}

// FreezeKeeper is a paid mutator transaction binding the contract method 0xb38ad8e7.
//
// Solidity: function FreezeKeeper(_keeper address) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) FreezeKeeper(_keeper common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.FreezeKeeper(&_SimpleGatekeeperWithLimit.TransactOpts, _keeper)
}

// FreezeKeeper is a paid mutator transaction binding the contract method 0xb38ad8e7.
//
// Solidity: function FreezeKeeper(_keeper address) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactorSession) FreezeKeeper(_keeper common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.FreezeKeeper(&_SimpleGatekeeperWithLimit.TransactOpts, _keeper)
}

// Payin is a paid mutator transaction binding the contract method 0x28f727f0.
//
// Solidity: function Payin(_value uint256) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactor) Payin(opts *bind.TransactOpts, _value *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.contract.Transact(opts, "Payin", _value)
}

// Payin is a paid mutator transaction binding the contract method 0x28f727f0.
//
// Solidity: function Payin(_value uint256) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) Payin(_value *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.Payin(&_SimpleGatekeeperWithLimit.TransactOpts, _value)
}

// Payin is a paid mutator transaction binding the contract method 0x28f727f0.
//
// Solidity: function Payin(_value uint256) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactorSession) Payin(_value *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.Payin(&_SimpleGatekeeperWithLimit.TransactOpts, _value)
}

// PayinTargeted is a paid mutator transaction binding the contract method 0xe3fcd18e.
//
// Solidity: function PayinTargeted(_value uint256, _target address) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactor) PayinTargeted(opts *bind.TransactOpts, _value *big.Int, _target common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.contract.Transact(opts, "PayinTargeted", _value, _target)
}

// PayinTargeted is a paid mutator transaction binding the contract method 0xe3fcd18e.
//
// Solidity: function PayinTargeted(_value uint256, _target address) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) PayinTargeted(_value *big.Int, _target common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.PayinTargeted(&_SimpleGatekeeperWithLimit.TransactOpts, _value, _target)
}

// PayinTargeted is a paid mutator transaction binding the contract method 0xe3fcd18e.
//
// Solidity: function PayinTargeted(_value uint256, _target address) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactorSession) PayinTargeted(_value *big.Int, _target common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.PayinTargeted(&_SimpleGatekeeperWithLimit.TransactOpts, _value, _target)
}

// Payout is a paid mutator transaction binding the contract method 0x634235fc.
//
// Solidity: function Payout(_to address, _value uint256, _txNumber uint256) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactor) Payout(opts *bind.TransactOpts, _to common.Address, _value *big.Int, _txNumber *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.contract.Transact(opts, "Payout", _to, _value, _txNumber)
}

// Payout is a paid mutator transaction binding the contract method 0x634235fc.
//
// Solidity: function Payout(_to address, _value uint256, _txNumber uint256) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) Payout(_to common.Address, _value *big.Int, _txNumber *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.Payout(&_SimpleGatekeeperWithLimit.TransactOpts, _to, _value, _txNumber)
}

// Payout is a paid mutator transaction binding the contract method 0x634235fc.
//
// Solidity: function Payout(_to address, _value uint256, _txNumber uint256) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactorSession) Payout(_to common.Address, _value *big.Int, _txNumber *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.Payout(&_SimpleGatekeeperWithLimit.TransactOpts, _to, _value, _txNumber)
}

// SetCommission is a paid mutator transaction binding the contract method 0x6ea58031.
//
// Solidity: function SetCommission(_commission uint256) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactor) SetCommission(opts *bind.TransactOpts, _commission *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.contract.Transact(opts, "SetCommission", _commission)
}

// SetCommission is a paid mutator transaction binding the contract method 0x6ea58031.
//
// Solidity: function SetCommission(_commission uint256) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) SetCommission(_commission *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.SetCommission(&_SimpleGatekeeperWithLimit.TransactOpts, _commission)
}

// SetCommission is a paid mutator transaction binding the contract method 0x6ea58031.
//
// Solidity: function SetCommission(_commission uint256) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactorSession) SetCommission(_commission *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.SetCommission(&_SimpleGatekeeperWithLimit.TransactOpts, _commission)
}

// SetFreezingTime is a paid mutator transaction binding the contract method 0xcc38d7ca.
//
// Solidity: function SetFreezingTime(_freezingTime uint256) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactor) SetFreezingTime(opts *bind.TransactOpts, _freezingTime *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.contract.Transact(opts, "SetFreezingTime", _freezingTime)
}

// SetFreezingTime is a paid mutator transaction binding the contract method 0xcc38d7ca.
//
// Solidity: function SetFreezingTime(_freezingTime uint256) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) SetFreezingTime(_freezingTime *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.SetFreezingTime(&_SimpleGatekeeperWithLimit.TransactOpts, _freezingTime)
}

// SetFreezingTime is a paid mutator transaction binding the contract method 0xcc38d7ca.
//
// Solidity: function SetFreezingTime(_freezingTime uint256) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactorSession) SetFreezingTime(_freezingTime *big.Int) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.SetFreezingTime(&_SimpleGatekeeperWithLimit.TransactOpts, _freezingTime)
}

// TransferCommission is a paid mutator transaction binding the contract method 0x06d8e8b1.
//
// Solidity: function TransferCommission() returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactor) TransferCommission(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.contract.Transact(opts, "TransferCommission")
}

// TransferCommission is a paid mutator transaction binding the contract method 0x06d8e8b1.
//
// Solidity: function TransferCommission() returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) TransferCommission() (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.TransferCommission(&_SimpleGatekeeperWithLimit.TransactOpts)
}

// TransferCommission is a paid mutator transaction binding the contract method 0x06d8e8b1.
//
// Solidity: function TransferCommission() returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactorSession) TransferCommission() (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.TransferCommission(&_SimpleGatekeeperWithLimit.TransactOpts)
}

// UnfreezeKeeper is a paid mutator transaction binding the contract method 0xe5837a7b.
//
// Solidity: function UnfreezeKeeper(_keeper address) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactor) UnfreezeKeeper(opts *bind.TransactOpts, _keeper common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.contract.Transact(opts, "UnfreezeKeeper", _keeper)
}

// UnfreezeKeeper is a paid mutator transaction binding the contract method 0xe5837a7b.
//
// Solidity: function UnfreezeKeeper(_keeper address) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) UnfreezeKeeper(_keeper common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.UnfreezeKeeper(&_SimpleGatekeeperWithLimit.TransactOpts, _keeper)
}

// UnfreezeKeeper is a paid mutator transaction binding the contract method 0xe5837a7b.
//
// Solidity: function UnfreezeKeeper(_keeper address) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactorSession) UnfreezeKeeper(_keeper common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.UnfreezeKeeper(&_SimpleGatekeeperWithLimit.TransactOpts, _keeper)
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactor) Kill(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.contract.Transact(opts, "kill")
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) Kill() (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.Kill(&_SimpleGatekeeperWithLimit.TransactOpts)
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactorSession) Kill() (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.Kill(&_SimpleGatekeeperWithLimit.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) RenounceOwnership() (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.RenounceOwnership(&_SimpleGatekeeperWithLimit.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.RenounceOwnership(&_SimpleGatekeeperWithLimit.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactor) TransferOwnership(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.contract.Transact(opts, "transferOwnership", _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.TransferOwnership(&_SimpleGatekeeperWithLimit.TransactOpts, _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitTransactorSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _SimpleGatekeeperWithLimit.Contract.TransferOwnership(&_SimpleGatekeeperWithLimit.TransactOpts, _newOwner)
}

// SimpleGatekeeperWithLimitCommissionChangedIterator is returned from FilterCommissionChanged and is used to iterate over the raw logs and unpacked data for CommissionChanged events raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitCommissionChangedIterator struct {
	Event *SimpleGatekeeperWithLimitCommissionChanged // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitCommissionChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitCommissionChanged)
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
		it.Event = new(SimpleGatekeeperWithLimitCommissionChanged)
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
func (it *SimpleGatekeeperWithLimitCommissionChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitCommissionChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitCommissionChanged represents a CommissionChanged event raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitCommissionChanged struct {
	Commission *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterCommissionChanged is a free log retrieval operation binding the contract event 0x839e4456845dbc05c7d8638cf0b0976161331b5f9163980d71d9a6444a326c61.
//
// Solidity: e CommissionChanged(commission indexed uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) FilterCommissionChanged(opts *bind.FilterOpts, commission []*big.Int) (*SimpleGatekeeperWithLimitCommissionChangedIterator, error) {

	var commissionRule []interface{}
	for _, commissionItem := range commission {
		commissionRule = append(commissionRule, commissionItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.FilterLogs(opts, "CommissionChanged", commissionRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitCommissionChangedIterator{contract: _SimpleGatekeeperWithLimit.contract, event: "CommissionChanged", logs: logs, sub: sub}, nil
}

// WatchCommissionChanged is a free log subscription operation binding the contract event 0x839e4456845dbc05c7d8638cf0b0976161331b5f9163980d71d9a6444a326c61.
//
// Solidity: e CommissionChanged(commission indexed uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) WatchCommissionChanged(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitCommissionChanged, commission []*big.Int) (event.Subscription, error) {

	var commissionRule []interface{}
	for _, commissionItem := range commission {
		commissionRule = append(commissionRule, commissionItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.WatchLogs(opts, "CommissionChanged", commissionRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitCommissionChanged)
				if err := _SimpleGatekeeperWithLimit.contract.UnpackLog(event, "CommissionChanged", log); err != nil {
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

// SimpleGatekeeperWithLimitCommitTxIterator is returned from FilterCommitTx and is used to iterate over the raw logs and unpacked data for CommitTx events raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitCommitTxIterator struct {
	Event *SimpleGatekeeperWithLimitCommitTx // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitCommitTxIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitCommitTx)
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
		it.Event = new(SimpleGatekeeperWithLimitCommitTx)
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
func (it *SimpleGatekeeperWithLimitCommitTxIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitCommitTxIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitCommitTx represents a CommitTx event raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitCommitTx struct {
	From            common.Address
	TxNumber        *big.Int
	Value           *big.Int
	CommitTimestamp *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterCommitTx is a free log retrieval operation binding the contract event 0x65546c3bc3a77ffc91667da85018004299542e28a511328cfb4b3f86974902ee.
//
// Solidity: e CommitTx(from indexed address, txNumber indexed uint256, value indexed uint256, commitTimestamp uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) FilterCommitTx(opts *bind.FilterOpts, from []common.Address, txNumber []*big.Int, value []*big.Int) (*SimpleGatekeeperWithLimitCommitTxIterator, error) {

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

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.FilterLogs(opts, "CommitTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitCommitTxIterator{contract: _SimpleGatekeeperWithLimit.contract, event: "CommitTx", logs: logs, sub: sub}, nil
}

// WatchCommitTx is a free log subscription operation binding the contract event 0x65546c3bc3a77ffc91667da85018004299542e28a511328cfb4b3f86974902ee.
//
// Solidity: e CommitTx(from indexed address, txNumber indexed uint256, value indexed uint256, commitTimestamp uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) WatchCommitTx(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitCommitTx, from []common.Address, txNumber []*big.Int, value []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.WatchLogs(opts, "CommitTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitCommitTx)
				if err := _SimpleGatekeeperWithLimit.contract.UnpackLog(event, "CommitTx", log); err != nil {
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

// SimpleGatekeeperWithLimitKeeperFreezedIterator is returned from FilterKeeperFreezed and is used to iterate over the raw logs and unpacked data for KeeperFreezed events raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitKeeperFreezedIterator struct {
	Event *SimpleGatekeeperWithLimitKeeperFreezed // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitKeeperFreezedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitKeeperFreezed)
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
		it.Event = new(SimpleGatekeeperWithLimitKeeperFreezed)
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
func (it *SimpleGatekeeperWithLimitKeeperFreezedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitKeeperFreezedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitKeeperFreezed represents a KeeperFreezed event raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitKeeperFreezed struct {
	Keeper common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterKeeperFreezed is a free log retrieval operation binding the contract event 0xdf4868d2f39f6ab9f41b92c6917da5aec882c461ce7316bb62076865108502bd.
//
// Solidity: e KeeperFreezed(keeper indexed address)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) FilterKeeperFreezed(opts *bind.FilterOpts, keeper []common.Address) (*SimpleGatekeeperWithLimitKeeperFreezedIterator, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.FilterLogs(opts, "KeeperFreezed", keeperRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitKeeperFreezedIterator{contract: _SimpleGatekeeperWithLimit.contract, event: "KeeperFreezed", logs: logs, sub: sub}, nil
}

// WatchKeeperFreezed is a free log subscription operation binding the contract event 0xdf4868d2f39f6ab9f41b92c6917da5aec882c461ce7316bb62076865108502bd.
//
// Solidity: e KeeperFreezed(keeper indexed address)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) WatchKeeperFreezed(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitKeeperFreezed, keeper []common.Address) (event.Subscription, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.WatchLogs(opts, "KeeperFreezed", keeperRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitKeeperFreezed)
				if err := _SimpleGatekeeperWithLimit.contract.UnpackLog(event, "KeeperFreezed", log); err != nil {
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

// SimpleGatekeeperWithLimitKeeperUnfreezedIterator is returned from FilterKeeperUnfreezed and is used to iterate over the raw logs and unpacked data for KeeperUnfreezed events raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitKeeperUnfreezedIterator struct {
	Event *SimpleGatekeeperWithLimitKeeperUnfreezed // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitKeeperUnfreezedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitKeeperUnfreezed)
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
		it.Event = new(SimpleGatekeeperWithLimitKeeperUnfreezed)
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
func (it *SimpleGatekeeperWithLimitKeeperUnfreezedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitKeeperUnfreezedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitKeeperUnfreezed represents a KeeperUnfreezed event raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitKeeperUnfreezed struct {
	Keeper common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterKeeperUnfreezed is a free log retrieval operation binding the contract event 0xbbe17a7427b5192903e1b3f0f2b6ef8b2a1af9b33e1079faf8f8383f2fb63b53.
//
// Solidity: e KeeperUnfreezed(keeper indexed address)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) FilterKeeperUnfreezed(opts *bind.FilterOpts, keeper []common.Address) (*SimpleGatekeeperWithLimitKeeperUnfreezedIterator, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.FilterLogs(opts, "KeeperUnfreezed", keeperRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitKeeperUnfreezedIterator{contract: _SimpleGatekeeperWithLimit.contract, event: "KeeperUnfreezed", logs: logs, sub: sub}, nil
}

// WatchKeeperUnfreezed is a free log subscription operation binding the contract event 0xbbe17a7427b5192903e1b3f0f2b6ef8b2a1af9b33e1079faf8f8383f2fb63b53.
//
// Solidity: e KeeperUnfreezed(keeper indexed address)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) WatchKeeperUnfreezed(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitKeeperUnfreezed, keeper []common.Address) (event.Subscription, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.WatchLogs(opts, "KeeperUnfreezed", keeperRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitKeeperUnfreezed)
				if err := _SimpleGatekeeperWithLimit.contract.UnpackLog(event, "KeeperUnfreezed", log); err != nil {
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

// SimpleGatekeeperWithLimitLimitChangedIterator is returned from FilterLimitChanged and is used to iterate over the raw logs and unpacked data for LimitChanged events raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitLimitChangedIterator struct {
	Event *SimpleGatekeeperWithLimitLimitChanged // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitLimitChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitLimitChanged)
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
		it.Event = new(SimpleGatekeeperWithLimitLimitChanged)
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
func (it *SimpleGatekeeperWithLimitLimitChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitLimitChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitLimitChanged represents a LimitChanged event raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitLimitChanged struct {
	Keeper   common.Address
	DayLimit *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterLimitChanged is a free log retrieval operation binding the contract event 0xef9c668177207fb68ca5e3894a1efacebb659762b27a737fde58ceebc4f30ad3.
//
// Solidity: e LimitChanged(keeper indexed address, dayLimit indexed uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) FilterLimitChanged(opts *bind.FilterOpts, keeper []common.Address, dayLimit []*big.Int) (*SimpleGatekeeperWithLimitLimitChangedIterator, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}
	var dayLimitRule []interface{}
	for _, dayLimitItem := range dayLimit {
		dayLimitRule = append(dayLimitRule, dayLimitItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.FilterLogs(opts, "LimitChanged", keeperRule, dayLimitRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitLimitChangedIterator{contract: _SimpleGatekeeperWithLimit.contract, event: "LimitChanged", logs: logs, sub: sub}, nil
}

// WatchLimitChanged is a free log subscription operation binding the contract event 0xef9c668177207fb68ca5e3894a1efacebb659762b27a737fde58ceebc4f30ad3.
//
// Solidity: e LimitChanged(keeper indexed address, dayLimit indexed uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) WatchLimitChanged(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitLimitChanged, keeper []common.Address, dayLimit []*big.Int) (event.Subscription, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}
	var dayLimitRule []interface{}
	for _, dayLimitItem := range dayLimit {
		dayLimitRule = append(dayLimitRule, dayLimitItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.WatchLogs(opts, "LimitChanged", keeperRule, dayLimitRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitLimitChanged)
				if err := _SimpleGatekeeperWithLimit.contract.UnpackLog(event, "LimitChanged", log); err != nil {
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

// SimpleGatekeeperWithLimitOwnershipRenouncedIterator is returned from FilterOwnershipRenounced and is used to iterate over the raw logs and unpacked data for OwnershipRenounced events raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitOwnershipRenouncedIterator struct {
	Event *SimpleGatekeeperWithLimitOwnershipRenounced // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitOwnershipRenouncedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitOwnershipRenounced)
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
		it.Event = new(SimpleGatekeeperWithLimitOwnershipRenounced)
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
func (it *SimpleGatekeeperWithLimitOwnershipRenouncedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitOwnershipRenouncedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitOwnershipRenounced represents a OwnershipRenounced event raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitOwnershipRenounced struct {
	PreviousOwner common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipRenounced is a free log retrieval operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) FilterOwnershipRenounced(opts *bind.FilterOpts, previousOwner []common.Address) (*SimpleGatekeeperWithLimitOwnershipRenouncedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.FilterLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitOwnershipRenouncedIterator{contract: _SimpleGatekeeperWithLimit.contract, event: "OwnershipRenounced", logs: logs, sub: sub}, nil
}

// WatchOwnershipRenounced is a free log subscription operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) WatchOwnershipRenounced(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitOwnershipRenounced, previousOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.WatchLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitOwnershipRenounced)
				if err := _SimpleGatekeeperWithLimit.contract.UnpackLog(event, "OwnershipRenounced", log); err != nil {
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

// SimpleGatekeeperWithLimitOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitOwnershipTransferredIterator struct {
	Event *SimpleGatekeeperWithLimitOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitOwnershipTransferred)
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
		it.Event = new(SimpleGatekeeperWithLimitOwnershipTransferred)
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
func (it *SimpleGatekeeperWithLimitOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitOwnershipTransferred represents a OwnershipTransferred event raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SimpleGatekeeperWithLimitOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitOwnershipTransferredIterator{contract: _SimpleGatekeeperWithLimit.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitOwnershipTransferred)
				if err := _SimpleGatekeeperWithLimit.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// SimpleGatekeeperWithLimitPayinTxIterator is returned from FilterPayinTx and is used to iterate over the raw logs and unpacked data for PayinTx events raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitPayinTxIterator struct {
	Event *SimpleGatekeeperWithLimitPayinTx // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitPayinTxIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitPayinTx)
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
		it.Event = new(SimpleGatekeeperWithLimitPayinTx)
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
func (it *SimpleGatekeeperWithLimitPayinTxIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitPayinTxIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitPayinTx represents a PayinTx event raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitPayinTx struct {
	From     common.Address
	TxNumber *big.Int
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPayinTx is a free log retrieval operation binding the contract event 0x14312725abbc46ad798bc078b2663e1fcbace97be0247cd177176f3b4df2538e.
//
// Solidity: e PayinTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) FilterPayinTx(opts *bind.FilterOpts, from []common.Address, txNumber []*big.Int, value []*big.Int) (*SimpleGatekeeperWithLimitPayinTxIterator, error) {

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

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.FilterLogs(opts, "PayinTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitPayinTxIterator{contract: _SimpleGatekeeperWithLimit.contract, event: "PayinTx", logs: logs, sub: sub}, nil
}

// WatchPayinTx is a free log subscription operation binding the contract event 0x14312725abbc46ad798bc078b2663e1fcbace97be0247cd177176f3b4df2538e.
//
// Solidity: e PayinTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) WatchPayinTx(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitPayinTx, from []common.Address, txNumber []*big.Int, value []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.WatchLogs(opts, "PayinTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitPayinTx)
				if err := _SimpleGatekeeperWithLimit.contract.UnpackLog(event, "PayinTx", log); err != nil {
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

// SimpleGatekeeperWithLimitPayoutTxIterator is returned from FilterPayoutTx and is used to iterate over the raw logs and unpacked data for PayoutTx events raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitPayoutTxIterator struct {
	Event *SimpleGatekeeperWithLimitPayoutTx // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitPayoutTxIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitPayoutTx)
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
		it.Event = new(SimpleGatekeeperWithLimitPayoutTx)
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
func (it *SimpleGatekeeperWithLimitPayoutTxIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitPayoutTxIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitPayoutTx represents a PayoutTx event raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitPayoutTx struct {
	From     common.Address
	TxNumber *big.Int
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPayoutTx is a free log retrieval operation binding the contract event 0x731af16374848c2c73a6154fd410cb421138e7db45c5a904e5a475c756faa8d9.
//
// Solidity: e PayoutTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) FilterPayoutTx(opts *bind.FilterOpts, from []common.Address, txNumber []*big.Int, value []*big.Int) (*SimpleGatekeeperWithLimitPayoutTxIterator, error) {

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

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.FilterLogs(opts, "PayoutTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitPayoutTxIterator{contract: _SimpleGatekeeperWithLimit.contract, event: "PayoutTx", logs: logs, sub: sub}, nil
}

// WatchPayoutTx is a free log subscription operation binding the contract event 0x731af16374848c2c73a6154fd410cb421138e7db45c5a904e5a475c756faa8d9.
//
// Solidity: e PayoutTx(from indexed address, txNumber indexed uint256, value indexed uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) WatchPayoutTx(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitPayoutTx, from []common.Address, txNumber []*big.Int, value []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.WatchLogs(opts, "PayoutTx", fromRule, txNumberRule, valueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitPayoutTx)
				if err := _SimpleGatekeeperWithLimit.contract.UnpackLog(event, "PayoutTx", log); err != nil {
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

// SimpleGatekeeperWithLimitSuicideIterator is returned from FilterSuicide and is used to iterate over the raw logs and unpacked data for Suicide events raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitSuicideIterator struct {
	Event *SimpleGatekeeperWithLimitSuicide // Event containing the contract specifics and raw log

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
func (it *SimpleGatekeeperWithLimitSuicideIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleGatekeeperWithLimitSuicide)
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
		it.Event = new(SimpleGatekeeperWithLimitSuicide)
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
func (it *SimpleGatekeeperWithLimitSuicideIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleGatekeeperWithLimitSuicideIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleGatekeeperWithLimitSuicide represents a Suicide event raised by the SimpleGatekeeperWithLimit contract.
type SimpleGatekeeperWithLimitSuicide struct {
	Block *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterSuicide is a free log retrieval operation binding the contract event 0xa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde.
//
// Solidity: e Suicide(block uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) FilterSuicide(opts *bind.FilterOpts) (*SimpleGatekeeperWithLimitSuicideIterator, error) {

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.FilterLogs(opts, "Suicide")
	if err != nil {
		return nil, err
	}
	return &SimpleGatekeeperWithLimitSuicideIterator{contract: _SimpleGatekeeperWithLimit.contract, event: "Suicide", logs: logs, sub: sub}, nil
}

// WatchSuicide is a free log subscription operation binding the contract event 0xa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde.
//
// Solidity: e Suicide(block uint256)
func (_SimpleGatekeeperWithLimit *SimpleGatekeeperWithLimitFilterer) WatchSuicide(opts *bind.WatchOpts, sink chan<- *SimpleGatekeeperWithLimitSuicide) (event.Subscription, error) {

	logs, sub, err := _SimpleGatekeeperWithLimit.contract.WatchLogs(opts, "Suicide")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleGatekeeperWithLimitSuicide)
				if err := _SimpleGatekeeperWithLimit.contract.UnpackLog(event, "Suicide", log); err != nil {
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
