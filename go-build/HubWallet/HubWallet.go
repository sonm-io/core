// This file is an automatically generated Go binding. Do not modify as any
// change will likely be lost upon the next re-generation!

package Hubwallet

import (
	"math/big"
	"strings"

	"github.com/sonm-io/go-ethereum/accounts/abi"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"github.com/sonm-io/go-ethereum/common"
	"github.com/sonm-io/go-ethereum/core/types"
)

// DeclarationABI is the input ABI used to generate the binding from.
const DeclarationABI = "[]"

// DeclarationBin is the compiled bytecode used for deploying new contracts.
const DeclarationBin = `0x60606040523415600b57fe5b5b60338060196000396000f30060606040525bfe00a165627a7a723058209be5a39b6105c214f918941e45e28f20b92862cd0e7cc1e7916a268ec3027f9a0029`

// DeployDeclaration deploys a new Ethereum contract, binding an instance of Declaration to it.
func DeployDeclaration(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Declaration, error) {
	parsed, err := abi.JSON(strings.NewReader(DeclarationABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(DeclarationBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Declaration{DeclarationCaller: DeclarationCaller{contract: contract}, DeclarationTransactor: DeclarationTransactor{contract: contract}}, nil
}

// Declaration is an auto generated Go binding around an Ethereum contract.
type Declaration struct {
	DeclarationCaller     // Read-only binding to the contract
	DeclarationTransactor // Write-only binding to the contract
}

// DeclarationCaller is an auto generated read-only Go binding around an Ethereum contract.
type DeclarationCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DeclarationTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DeclarationTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DeclarationSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DeclarationSession struct {
	Contract     *Declaration      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DeclarationCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DeclarationCallerSession struct {
	Contract *DeclarationCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// DeclarationTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DeclarationTransactorSession struct {
	Contract     *DeclarationTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// DeclarationRaw is an auto generated low-level Go binding around an Ethereum contract.
type DeclarationRaw struct {
	Contract *Declaration // Generic contract binding to access the raw methods on
}

// DeclarationCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DeclarationCallerRaw struct {
	Contract *DeclarationCaller // Generic read-only contract binding to access the raw methods on
}

// DeclarationTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DeclarationTransactorRaw struct {
	Contract *DeclarationTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDeclaration creates a new instance of Declaration, bound to a specific deployed contract.
func NewDeclaration(address common.Address, backend bind.ContractBackend) (*Declaration, error) {
	contract, err := bindDeclaration(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Declaration{DeclarationCaller: DeclarationCaller{contract: contract}, DeclarationTransactor: DeclarationTransactor{contract: contract}}, nil
}

// NewDeclarationCaller creates a new read-only instance of Declaration, bound to a specific deployed contract.
func NewDeclarationCaller(address common.Address, caller bind.ContractCaller) (*DeclarationCaller, error) {
	contract, err := bindDeclaration(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &DeclarationCaller{contract: contract}, nil
}

// NewDeclarationTransactor creates a new write-only instance of Declaration, bound to a specific deployed contract.
func NewDeclarationTransactor(address common.Address, transactor bind.ContractTransactor) (*DeclarationTransactor, error) {
	contract, err := bindDeclaration(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &DeclarationTransactor{contract: contract}, nil
}

// bindDeclaration binds a generic wrapper to an already deployed contract.
func bindDeclaration(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(DeclarationABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Declaration *DeclarationRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Declaration.Contract.DeclarationCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Declaration *DeclarationRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Declaration.Contract.DeclarationTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Declaration *DeclarationRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Declaration.Contract.DeclarationTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Declaration *DeclarationCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Declaration.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Declaration *DeclarationTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Declaration.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Declaration *DeclarationTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Declaration.Contract.contract.Transact(opts, method, params...)
}

// HubWalletABI is the input ABI used to generate the binding from.
const HubWalletABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"currentPhase\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"freezePeriod\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"gulag\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"suspect\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"sharesTokenAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"genesisTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint64\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"PayDay\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"freezeQuote\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"frozenTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint64\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"frozenFunds\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lockPercent\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"DAO\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lockedFunds\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"rehub\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"Factory\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"Registration\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"inputs\":[{\"name\":\"_hubowner\",\"type\":\"address\"},{\"name\":\"_dao\",\"type\":\"address\"},{\"name\":\"_whitelist\",\"type\":\"address\"},{\"name\":\"sharesAddress\",\"type\":\"address\"}],\"payable\":false,\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"newPhase\",\"type\":\"uint8\"}],\"name\":\"LogPhaseSwitch\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"pass\",\"type\":\"string\"}],\"name\":\"LogPass\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"val\",\"type\":\"uint256\"}],\"name\":\"ToVal\",\"type\":\"event\"}]"

// HubWalletBin is the compiled bytecode used for deploying new contracts.
const HubWalletBin = `0x60606040526000600855341561001157fe5b60405160808061102b83398101604090815281516020830151918301516060909301519092905b5b60008054600160a060020a03191633600160a060020a03161790555b60008054600160a060020a0319908116600160a060020a0387811691909117909255600180548216868416178155600380548316868516179055600280548316338516178155600b805467ffffffffffffffff19164267ffffffffffffffff161790556004805490931693851693909317909155670de0b6b3a76400006005908155601e600755600c5562278d00600a55600e8054909160ff1990911690835b02179055505b505050505b610f1c8061010f6000396000f3006060604052361561010f5763ffffffff7c0100000000000000000000000000000000000000000000000000000000600035041663055ad42e81146101115780630a3cb663146101455780630b3eeac8146101675780631e1683af1461017957806327ebcf0e1461018b5780633ccfd60b146101b757806342c6498a146101c95780634d78511c146101f6578063565f6c49146102085780638da5cb5b1461022a578063906db9ff1461025657806390a74e2c1461028357806391030cb6146102a557806398fabd3a146102c7578063a9059cbb146102f3578063b8afaa4814610314578063bd73820d14610336578063c83dd23114610348578063e8a3791914610374578063f2fde38b14610398575bfe5b341561011957fe5b6101216103b6565b6040518082600481111561013157fe5b60ff16815260200191505060405180910390f35b341561014d57fe5b6101556103bf565b60408051918252519081900360200190f35b341561016f57fe5b6101776103c5565b005b341561018157fe5b610177610505565b005b341561019357fe5b61019b6106c5565b60408051600160a060020a039092168252519081900360200190f35b34156101bf57fe5b6101776106d4565b005b34156101d157fe5b6101d9610821565b6040805167ffffffffffffffff9092168252519081900360200190f35b34156101fe57fe5b610177610831565b005b341561021057fe5b610155610a17565b60408051918252519081900360200190f35b341561023257fe5b61019b610a1d565b60408051600160a060020a039092168252519081900360200190f35b341561025e57fe5b6101d9610a2c565b6040805167ffffffffffffffff9092168252519081900360200190f35b341561028b57fe5b610155610a3c565b60408051918252519081900360200190f35b34156102ad57fe5b610155610a42565b60408051918252519081900360200190f35b34156102cf57fe5b61019b610a48565b60408051600160a060020a039092168252519081900360200190f35b34156102fb57fe5b610177600160a060020a0360043516602435610a57565b005b341561031c57fe5b610155610c02565b60408051918252519081900360200190f35b341561033e57fe5b610177610c08565b005b341561035057fe5b61019b610ca4565b60408051600160a060020a039092168252519081900360200190f35b341561037c57fe5b610384610cb3565b604080519115158252519081900360200190f35b34156103a057fe5b610177600160a060020a0360043516610e77565b005b600e5460ff1681565b600a5481565b60015460009033600160a060020a039081169116146103e45760006000fd5b60035b600e5460ff1660048111156103f857fe5b146104035760006000fd5b600a5460095467ffffffffffffffff16014210156104215760006000fd5b5060085460048054600154604080516000602091820181905282517fa9059cbb000000000000000000000000000000000000000000000000000000008152600160a060020a0394851696810196909652602486018790529151929093169363a9059cbb93604480830194919391928390030190829087803b15156104a157fe5b6102c65a03f115156104af57fe5b5050600e80546004925060ff19166001835b0217905550600e54604051600080516020610ed18339815191529160ff1690808260048111156104ed57fe5b60ff16815260200191505060405180910390a15b5b50565b60015433600160a060020a039081169116146105215760006000fd5b60015b600e5460ff16600481111561053557fe5b146105405760006000fd5b60048054604080516000602091820181905282517f70a08231000000000000000000000000000000000000000000000000000000008152600160a060020a0330811696820196909652925194909316936370a08231936024808501948390030190829087803b15156105ae57fe5b6102c65a03f115156105bc57fe5b505060405151600855506009805467ffffffffffffffff19164267ffffffffffffffff16179055600e80546003919060ff19166001835b0217905550600e54604051600080516020610ed18339815191529160ff16908082600481111561061f57fe5b60ff16815260200191505060405180910390a1629e3400600a556003546000805460408051602090810184905281517f0bdf0962000000000000000000000000000000000000000000000000000000008152600160a060020a039384166004820152308416602482015291519290941693630bdf09629360448084019492938390030190829087803b15156106b057fe5b6102c65a03f115156106be57fe5b5050505b5b565b600454600160a060020a031681565b6000805433600160a060020a039081169116146106f15760006000fd5b60025b600e5460ff16600481111561070557fe5b146107105760006000fd5b60048054604080516000602091820181905282517f70a08231000000000000000000000000000000000000000000000000000000008152600160a060020a0330811696820196909652925194909316936370a08231936024808501948390030190829087803b151561077e57fe5b6102c65a03f1151561078c57fe5b50506040805180516004805460008054602095860182905286517fa9059cbb000000000000000000000000000000000000000000000000000000008152600160a060020a03918216948101949094526024840185905295519397509416945063a9059cbb9360448083019493928390030190829087803b151561080b57fe5b6102c65a03f1151561081957fe5b5050505b5b50565b600b5467ffffffffffffffff1681565b60005433600160a060020a0390811691161461084d5760006000fd5b60015b600e5460ff16600481111561086157fe5b1461086c5760006000fd5b600c546008546103e891025b6006805492909104909101600d55600090819055600855600a5460095467ffffffffffffffff16014210156108ad5760006000fd5b60048054600154600d54604080516000602091820181905282517fa9059cbb000000000000000000000000000000000000000000000000000000008152600160a060020a039586169781019790975260248701939093529051929093169363a9059cbb93604480830194919391928390030190829087803b151561092d57fe5b6102c65a03f1151561093b57fe5b50506040805160035460008054602093840182905284517f0bdf0962000000000000000000000000000000000000000000000000000000008152600160a060020a0391821660048201523082166024820152945192169450630bdf0962936044808201949392918390030190829087803b15156109b457fe5b6102c65a03f115156109c257fe5b5050600e80546002925060ff19166001835b0217905550600e54604051600080516020610ed18339815191529160ff169080826004811115610a0057fe5b60ff16815260200191505060405180910390a15b5b565b60055481565b600054600160a060020a031681565b60095467ffffffffffffffff1681565b60065481565b60075481565b600154600160a060020a031681565b6000805481908190819033600160a060020a03908116911614610a7a5760006000fd5b60015b600e5460ff166004811115610a8e57fe5b14610a995760006000fd5b60075460649086025b04935083600854019250600654830191508385039050808201600460009054906101000a9004600160a060020a0316600160a060020a03166370a08231336000604051602001526040518263ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018082600160a060020a0316600160a060020a03168152602001915050602060405180830381600087803b1515610b4957fe5b6102c65a03f11515610b5757fe5b505050604051805190501015610b6d5760006000fd5b600883905560048054604080516000602091820181905282517f095ea7b3000000000000000000000000000000000000000000000000000000008152600160a060020a038c811696820196909652602481018790529251949093169363095ea7b3936044808501948390030190829087803b1515610be757fe5b6102c65a03f11515610bf557fe5b5050505b5b505050505050565b60085481565b60015433600160a060020a03908116911614610c245760006000fd5b60035b600e5460ff166004811115610c3857fe5b14610c435760006000fd5b60006008819055600655600e80546002919060ff19166001836109d4565b0217905550600e54604051600080516020610ed18339815191529160ff169080826004811115610a0057fe5b60ff16815260200191505060405180910390a15b5b565b600254600160a060020a031681565b600060025b600e5460ff166004811115610cc957fe5b14610cd45760006000fd5b60055460048054604080516000602091820181905282517f70a08231000000000000000000000000000000000000000000000000000000008152600160a060020a0333811696820196909652925194909316936370a08231936024808501948390030190829087803b1515610d4557fe5b6102c65a03f11515610d5357fe5b505060405151919091119050610d695760006000fd5b6005546006556009805467ffffffffffffffff19164267ffffffffffffffff90811691909117918290556003546000805460408051602090810184905281517f94baf301000000000000000000000000000000000000000000000000000000008152600160a060020a0393841660048201523084166024820152969095166044870152519216936394baf30193606480830194919391928390030190829087803b1515610e1257fe5b6102c65a03f11515610e2057fe5b5050600e80546001925060ff191682805b0217905550600e54604051600080516020610ed18339815191529160ff169080826004811115610e5d57fe5b60ff16815260200191505060405180910390a15060015b90565b60005433600160a060020a03908116911614610e935760006000fd5b600160a060020a03811615610501576000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0383161790555b5b5b5056008d9efa3fab1bd6476defa44f520afbf9337886a4947021fd7f2775e0efaf4571a165627a7a723058202f4f00a497bcd1334e900a5347e224558dd7ea81d8b3703b645b4f7d20a556b50029`

// DeployHubWallet deploys a new Ethereum contract, binding an instance of HubWallet to it.
func DeployHubWallet(auth *bind.TransactOpts, backend bind.ContractBackend, _hubowner common.Address, _dao common.Address, _whitelist common.Address, sharesAddress common.Address) (common.Address, *types.Transaction, *HubWallet, error) {
	parsed, err := abi.JSON(strings.NewReader(HubWalletABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(HubWalletBin), backend, _hubowner, _dao, _whitelist, sharesAddress)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &HubWallet{HubWalletCaller: HubWalletCaller{contract: contract}, HubWalletTransactor: HubWalletTransactor{contract: contract}}, nil
}

// HubWallet is an auto generated Go binding around an Ethereum contract.
type HubWallet struct {
	HubWalletCaller     // Read-only binding to the contract
	HubWalletTransactor // Write-only binding to the contract
}

// HubWalletCaller is an auto generated read-only Go binding around an Ethereum contract.
type HubWalletCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HubWalletTransactor is an auto generated write-only Go binding around an Ethereum contract.
type HubWalletTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HubWalletSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type HubWalletSession struct {
	Contract     *HubWallet        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HubWalletCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type HubWalletCallerSession struct {
	Contract *HubWalletCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// HubWalletTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type HubWalletTransactorSession struct {
	Contract     *HubWalletTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// HubWalletRaw is an auto generated low-level Go binding around an Ethereum contract.
type HubWalletRaw struct {
	Contract *HubWallet // Generic contract binding to access the raw methods on
}

// HubWalletCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type HubWalletCallerRaw struct {
	Contract *HubWalletCaller // Generic read-only contract binding to access the raw methods on
}

// HubWalletTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type HubWalletTransactorRaw struct {
	Contract *HubWalletTransactor // Generic write-only contract binding to access the raw methods on
}

// NewHubWallet creates a new instance of HubWallet, bound to a specific deployed contract.
func NewHubWallet(address common.Address, backend bind.ContractBackend) (*HubWallet, error) {
	contract, err := bindHubWallet(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &HubWallet{HubWalletCaller: HubWalletCaller{contract: contract}, HubWalletTransactor: HubWalletTransactor{contract: contract}}, nil
}

// NewHubWalletCaller creates a new read-only instance of HubWallet, bound to a specific deployed contract.
func NewHubWalletCaller(address common.Address, caller bind.ContractCaller) (*HubWalletCaller, error) {
	contract, err := bindHubWallet(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &HubWalletCaller{contract: contract}, nil
}

// NewHubWalletTransactor creates a new write-only instance of HubWallet, bound to a specific deployed contract.
func NewHubWalletTransactor(address common.Address, transactor bind.ContractTransactor) (*HubWalletTransactor, error) {
	contract, err := bindHubWallet(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &HubWalletTransactor{contract: contract}, nil
}

// bindHubWallet binds a generic wrapper to an already deployed contract.
func bindHubWallet(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(HubWalletABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HubWallet *HubWalletRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _HubWallet.Contract.HubWalletCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HubWallet *HubWalletRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.Contract.HubWalletTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HubWallet *HubWalletRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HubWallet.Contract.HubWalletTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HubWallet *HubWalletCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _HubWallet.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HubWallet *HubWalletTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HubWallet *HubWalletTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HubWallet.Contract.contract.Transact(opts, method, params...)
}

// DAO is a free data retrieval call binding the contract method 0x98fabd3a.
//
// Solidity: function DAO() constant returns(address)
func (_HubWallet *HubWalletCaller) DAO(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "DAO")
	return *ret0, err
}

// DAO is a free data retrieval call binding the contract method 0x98fabd3a.
//
// Solidity: function DAO() constant returns(address)
func (_HubWallet *HubWalletSession) DAO() (common.Address, error) {
	return _HubWallet.Contract.DAO(&_HubWallet.CallOpts)
}

// DAO is a free data retrieval call binding the contract method 0x98fabd3a.
//
// Solidity: function DAO() constant returns(address)
func (_HubWallet *HubWalletCallerSession) DAO() (common.Address, error) {
	return _HubWallet.Contract.DAO(&_HubWallet.CallOpts)
}

// Factory is a free data retrieval call binding the contract method 0xc83dd231.
//
// Solidity: function Factory() constant returns(address)
func (_HubWallet *HubWalletCaller) Factory(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "Factory")
	return *ret0, err
}

// Factory is a free data retrieval call binding the contract method 0xc83dd231.
//
// Solidity: function Factory() constant returns(address)
func (_HubWallet *HubWalletSession) Factory() (common.Address, error) {
	return _HubWallet.Contract.Factory(&_HubWallet.CallOpts)
}

// Factory is a free data retrieval call binding the contract method 0xc83dd231.
//
// Solidity: function Factory() constant returns(address)
func (_HubWallet *HubWalletCallerSession) Factory() (common.Address, error) {
	return _HubWallet.Contract.Factory(&_HubWallet.CallOpts)
}

// CurrentPhase is a free data retrieval call binding the contract method 0x055ad42e.
//
// Solidity: function currentPhase() constant returns(uint8)
func (_HubWallet *HubWalletCaller) CurrentPhase(opts *bind.CallOpts) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "currentPhase")
	return *ret0, err
}

// CurrentPhase is a free data retrieval call binding the contract method 0x055ad42e.
//
// Solidity: function currentPhase() constant returns(uint8)
func (_HubWallet *HubWalletSession) CurrentPhase() (uint8, error) {
	return _HubWallet.Contract.CurrentPhase(&_HubWallet.CallOpts)
}

// CurrentPhase is a free data retrieval call binding the contract method 0x055ad42e.
//
// Solidity: function currentPhase() constant returns(uint8)
func (_HubWallet *HubWalletCallerSession) CurrentPhase() (uint8, error) {
	return _HubWallet.Contract.CurrentPhase(&_HubWallet.CallOpts)
}

// FreezePeriod is a free data retrieval call binding the contract method 0x0a3cb663.
//
// Solidity: function freezePeriod() constant returns(uint256)
func (_HubWallet *HubWalletCaller) FreezePeriod(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "freezePeriod")
	return *ret0, err
}

// FreezePeriod is a free data retrieval call binding the contract method 0x0a3cb663.
//
// Solidity: function freezePeriod() constant returns(uint256)
func (_HubWallet *HubWalletSession) FreezePeriod() (*big.Int, error) {
	return _HubWallet.Contract.FreezePeriod(&_HubWallet.CallOpts)
}

// FreezePeriod is a free data retrieval call binding the contract method 0x0a3cb663.
//
// Solidity: function freezePeriod() constant returns(uint256)
func (_HubWallet *HubWalletCallerSession) FreezePeriod() (*big.Int, error) {
	return _HubWallet.Contract.FreezePeriod(&_HubWallet.CallOpts)
}

// FreezeQuote is a free data retrieval call binding the contract method 0x565f6c49.
//
// Solidity: function freezeQuote() constant returns(uint256)
func (_HubWallet *HubWalletCaller) FreezeQuote(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "freezeQuote")
	return *ret0, err
}

// FreezeQuote is a free data retrieval call binding the contract method 0x565f6c49.
//
// Solidity: function freezeQuote() constant returns(uint256)
func (_HubWallet *HubWalletSession) FreezeQuote() (*big.Int, error) {
	return _HubWallet.Contract.FreezeQuote(&_HubWallet.CallOpts)
}

// FreezeQuote is a free data retrieval call binding the contract method 0x565f6c49.
//
// Solidity: function freezeQuote() constant returns(uint256)
func (_HubWallet *HubWalletCallerSession) FreezeQuote() (*big.Int, error) {
	return _HubWallet.Contract.FreezeQuote(&_HubWallet.CallOpts)
}

// FrozenFunds is a free data retrieval call binding the contract method 0x90a74e2c.
//
// Solidity: function frozenFunds() constant returns(uint256)
func (_HubWallet *HubWalletCaller) FrozenFunds(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "frozenFunds")
	return *ret0, err
}

// FrozenFunds is a free data retrieval call binding the contract method 0x90a74e2c.
//
// Solidity: function frozenFunds() constant returns(uint256)
func (_HubWallet *HubWalletSession) FrozenFunds() (*big.Int, error) {
	return _HubWallet.Contract.FrozenFunds(&_HubWallet.CallOpts)
}

// FrozenFunds is a free data retrieval call binding the contract method 0x90a74e2c.
//
// Solidity: function frozenFunds() constant returns(uint256)
func (_HubWallet *HubWalletCallerSession) FrozenFunds() (*big.Int, error) {
	return _HubWallet.Contract.FrozenFunds(&_HubWallet.CallOpts)
}

// FrozenTime is a free data retrieval call binding the contract method 0x906db9ff.
//
// Solidity: function frozenTime() constant returns(uint64)
func (_HubWallet *HubWalletCaller) FrozenTime(opts *bind.CallOpts) (uint64, error) {
	var (
		ret0 = new(uint64)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "frozenTime")
	return *ret0, err
}

// FrozenTime is a free data retrieval call binding the contract method 0x906db9ff.
//
// Solidity: function frozenTime() constant returns(uint64)
func (_HubWallet *HubWalletSession) FrozenTime() (uint64, error) {
	return _HubWallet.Contract.FrozenTime(&_HubWallet.CallOpts)
}

// FrozenTime is a free data retrieval call binding the contract method 0x906db9ff.
//
// Solidity: function frozenTime() constant returns(uint64)
func (_HubWallet *HubWalletCallerSession) FrozenTime() (uint64, error) {
	return _HubWallet.Contract.FrozenTime(&_HubWallet.CallOpts)
}

// GenesisTime is a free data retrieval call binding the contract method 0x42c6498a.
//
// Solidity: function genesisTime() constant returns(uint64)
func (_HubWallet *HubWalletCaller) GenesisTime(opts *bind.CallOpts) (uint64, error) {
	var (
		ret0 = new(uint64)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "genesisTime")
	return *ret0, err
}

// GenesisTime is a free data retrieval call binding the contract method 0x42c6498a.
//
// Solidity: function genesisTime() constant returns(uint64)
func (_HubWallet *HubWalletSession) GenesisTime() (uint64, error) {
	return _HubWallet.Contract.GenesisTime(&_HubWallet.CallOpts)
}

// GenesisTime is a free data retrieval call binding the contract method 0x42c6498a.
//
// Solidity: function genesisTime() constant returns(uint64)
func (_HubWallet *HubWalletCallerSession) GenesisTime() (uint64, error) {
	return _HubWallet.Contract.GenesisTime(&_HubWallet.CallOpts)
}

// LockPercent is a free data retrieval call binding the contract method 0x91030cb6.
//
// Solidity: function lockPercent() constant returns(uint256)
func (_HubWallet *HubWalletCaller) LockPercent(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "lockPercent")
	return *ret0, err
}

// LockPercent is a free data retrieval call binding the contract method 0x91030cb6.
//
// Solidity: function lockPercent() constant returns(uint256)
func (_HubWallet *HubWalletSession) LockPercent() (*big.Int, error) {
	return _HubWallet.Contract.LockPercent(&_HubWallet.CallOpts)
}

// LockPercent is a free data retrieval call binding the contract method 0x91030cb6.
//
// Solidity: function lockPercent() constant returns(uint256)
func (_HubWallet *HubWalletCallerSession) LockPercent() (*big.Int, error) {
	return _HubWallet.Contract.LockPercent(&_HubWallet.CallOpts)
}

// LockedFunds is a free data retrieval call binding the contract method 0xb8afaa48.
//
// Solidity: function lockedFunds() constant returns(uint256)
func (_HubWallet *HubWalletCaller) LockedFunds(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "lockedFunds")
	return *ret0, err
}

// LockedFunds is a free data retrieval call binding the contract method 0xb8afaa48.
//
// Solidity: function lockedFunds() constant returns(uint256)
func (_HubWallet *HubWalletSession) LockedFunds() (*big.Int, error) {
	return _HubWallet.Contract.LockedFunds(&_HubWallet.CallOpts)
}

// LockedFunds is a free data retrieval call binding the contract method 0xb8afaa48.
//
// Solidity: function lockedFunds() constant returns(uint256)
func (_HubWallet *HubWalletCallerSession) LockedFunds() (*big.Int, error) {
	return _HubWallet.Contract.LockedFunds(&_HubWallet.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_HubWallet *HubWalletCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_HubWallet *HubWalletSession) Owner() (common.Address, error) {
	return _HubWallet.Contract.Owner(&_HubWallet.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_HubWallet *HubWalletCallerSession) Owner() (common.Address, error) {
	return _HubWallet.Contract.Owner(&_HubWallet.CallOpts)
}

// SharesTokenAddress is a free data retrieval call binding the contract method 0x27ebcf0e.
//
// Solidity: function sharesTokenAddress() constant returns(address)
func (_HubWallet *HubWalletCaller) SharesTokenAddress(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _HubWallet.contract.Call(opts, out, "sharesTokenAddress")
	return *ret0, err
}

// SharesTokenAddress is a free data retrieval call binding the contract method 0x27ebcf0e.
//
// Solidity: function sharesTokenAddress() constant returns(address)
func (_HubWallet *HubWalletSession) SharesTokenAddress() (common.Address, error) {
	return _HubWallet.Contract.SharesTokenAddress(&_HubWallet.CallOpts)
}

// SharesTokenAddress is a free data retrieval call binding the contract method 0x27ebcf0e.
//
// Solidity: function sharesTokenAddress() constant returns(address)
func (_HubWallet *HubWalletCallerSession) SharesTokenAddress() (common.Address, error) {
	return _HubWallet.Contract.SharesTokenAddress(&_HubWallet.CallOpts)
}

// PayDay is a paid mutator transaction binding the contract method 0x4d78511c.
//
// Solidity: function PayDay() returns()
func (_HubWallet *HubWalletTransactor) PayDay(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "PayDay")
}

// PayDay is a paid mutator transaction binding the contract method 0x4d78511c.
//
// Solidity: function PayDay() returns()
func (_HubWallet *HubWalletSession) PayDay() (*types.Transaction, error) {
	return _HubWallet.Contract.PayDay(&_HubWallet.TransactOpts)
}

// PayDay is a paid mutator transaction binding the contract method 0x4d78511c.
//
// Solidity: function PayDay() returns()
func (_HubWallet *HubWalletTransactorSession) PayDay() (*types.Transaction, error) {
	return _HubWallet.Contract.PayDay(&_HubWallet.TransactOpts)
}

// Registration is a paid mutator transaction binding the contract method 0xe8a37919.
//
// Solidity: function Registration() returns(success bool)
func (_HubWallet *HubWalletTransactor) Registration(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "Registration")
}

// Registration is a paid mutator transaction binding the contract method 0xe8a37919.
//
// Solidity: function Registration() returns(success bool)
func (_HubWallet *HubWalletSession) Registration() (*types.Transaction, error) {
	return _HubWallet.Contract.Registration(&_HubWallet.TransactOpts)
}

// Registration is a paid mutator transaction binding the contract method 0xe8a37919.
//
// Solidity: function Registration() returns(success bool)
func (_HubWallet *HubWalletTransactorSession) Registration() (*types.Transaction, error) {
	return _HubWallet.Contract.Registration(&_HubWallet.TransactOpts)
}

// Gulag is a paid mutator transaction binding the contract method 0x0b3eeac8.
//
// Solidity: function gulag() returns()
func (_HubWallet *HubWalletTransactor) Gulag(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "gulag")
}

// Gulag is a paid mutator transaction binding the contract method 0x0b3eeac8.
//
// Solidity: function gulag() returns()
func (_HubWallet *HubWalletSession) Gulag() (*types.Transaction, error) {
	return _HubWallet.Contract.Gulag(&_HubWallet.TransactOpts)
}

// Gulag is a paid mutator transaction binding the contract method 0x0b3eeac8.
//
// Solidity: function gulag() returns()
func (_HubWallet *HubWalletTransactorSession) Gulag() (*types.Transaction, error) {
	return _HubWallet.Contract.Gulag(&_HubWallet.TransactOpts)
}

// Rehub is a paid mutator transaction binding the contract method 0xbd73820d.
//
// Solidity: function rehub() returns()
func (_HubWallet *HubWalletTransactor) Rehub(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "rehub")
}

// Rehub is a paid mutator transaction binding the contract method 0xbd73820d.
//
// Solidity: function rehub() returns()
func (_HubWallet *HubWalletSession) Rehub() (*types.Transaction, error) {
	return _HubWallet.Contract.Rehub(&_HubWallet.TransactOpts)
}

// Rehub is a paid mutator transaction binding the contract method 0xbd73820d.
//
// Solidity: function rehub() returns()
func (_HubWallet *HubWalletTransactorSession) Rehub() (*types.Transaction, error) {
	return _HubWallet.Contract.Rehub(&_HubWallet.TransactOpts)
}

// Suspect is a paid mutator transaction binding the contract method 0x1e1683af.
//
// Solidity: function suspect() returns()
func (_HubWallet *HubWalletTransactor) Suspect(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "suspect")
}

// Suspect is a paid mutator transaction binding the contract method 0x1e1683af.
//
// Solidity: function suspect() returns()
func (_HubWallet *HubWalletSession) Suspect() (*types.Transaction, error) {
	return _HubWallet.Contract.Suspect(&_HubWallet.TransactOpts)
}

// Suspect is a paid mutator transaction binding the contract method 0x1e1683af.
//
// Solidity: function suspect() returns()
func (_HubWallet *HubWalletTransactorSession) Suspect() (*types.Transaction, error) {
	return _HubWallet.Contract.Suspect(&_HubWallet.TransactOpts)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns()
func (_HubWallet *HubWalletTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns()
func (_HubWallet *HubWalletSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _HubWallet.Contract.Transfer(&_HubWallet.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns()
func (_HubWallet *HubWalletTransactorSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _HubWallet.Contract.Transfer(&_HubWallet.TransactOpts, _to, _value)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_HubWallet *HubWalletTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_HubWallet *HubWalletSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _HubWallet.Contract.TransferOwnership(&_HubWallet.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_HubWallet *HubWalletTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _HubWallet.Contract.TransferOwnership(&_HubWallet.TransactOpts, newOwner)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_HubWallet *HubWalletTransactor) Withdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HubWallet.contract.Transact(opts, "withdraw")
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_HubWallet *HubWalletSession) Withdraw() (*types.Transaction, error) {
	return _HubWallet.Contract.Withdraw(&_HubWallet.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_HubWallet *HubWalletTransactorSession) Withdraw() (*types.Transaction, error) {
	return _HubWallet.Contract.Withdraw(&_HubWallet.TransactOpts)
}

// OwnableABI is the input ABI used to generate the binding from.
const OwnableABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"type\":\"constructor\"}]"

// OwnableBin is the compiled bytecode used for deploying new contracts.
const OwnableBin = `0x6060604052341561000c57fe5b5b60008054600160a060020a03191633600160a060020a03161790555b5b610119806100396000396000f300606060405263ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416638da5cb5b81146043578063f2fde38b14606c575bfe5b3415604a57fe5b60506087565b60408051600160a060020a039092168252519081900360200190f35b3415607357fe5b6085600160a060020a03600435166096565b005b600054600160a060020a031681565b60005433600160a060020a0390811691161460b15760006000fd5b600160a060020a0381161560e8576000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0383161790555b5b5b505600a165627a7a72305820b2910f42d40842f7d00082855d97f8bd466e1bd46cc1bad4f208faab8d314bb20029`

// DeployOwnable deploys a new Ethereum contract, binding an instance of Ownable to it.
func DeployOwnable(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Ownable, error) {
	parsed, err := abi.JSON(strings.NewReader(OwnableABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(OwnableBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Ownable{OwnableCaller: OwnableCaller{contract: contract}, OwnableTransactor: OwnableTransactor{contract: contract}}, nil
}

// Ownable is an auto generated Go binding around an Ethereum contract.
type Ownable struct {
	OwnableCaller     // Read-only binding to the contract
	OwnableTransactor // Write-only binding to the contract
}

// OwnableCaller is an auto generated read-only Go binding around an Ethereum contract.
type OwnableCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnableTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OwnableTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnableSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OwnableSession struct {
	Contract     *Ownable          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OwnableCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OwnableCallerSession struct {
	Contract *OwnableCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// OwnableTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OwnableTransactorSession struct {
	Contract     *OwnableTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// OwnableRaw is an auto generated low-level Go binding around an Ethereum contract.
type OwnableRaw struct {
	Contract *Ownable // Generic contract binding to access the raw methods on
}

// OwnableCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OwnableCallerRaw struct {
	Contract *OwnableCaller // Generic read-only contract binding to access the raw methods on
}

// OwnableTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OwnableTransactorRaw struct {
	Contract *OwnableTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOwnable creates a new instance of Ownable, bound to a specific deployed contract.
func NewOwnable(address common.Address, backend bind.ContractBackend) (*Ownable, error) {
	contract, err := bindOwnable(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Ownable{OwnableCaller: OwnableCaller{contract: contract}, OwnableTransactor: OwnableTransactor{contract: contract}}, nil
}

// NewOwnableCaller creates a new read-only instance of Ownable, bound to a specific deployed contract.
func NewOwnableCaller(address common.Address, caller bind.ContractCaller) (*OwnableCaller, error) {
	contract, err := bindOwnable(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &OwnableCaller{contract: contract}, nil
}

// NewOwnableTransactor creates a new write-only instance of Ownable, bound to a specific deployed contract.
func NewOwnableTransactor(address common.Address, transactor bind.ContractTransactor) (*OwnableTransactor, error) {
	contract, err := bindOwnable(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &OwnableTransactor{contract: contract}, nil
}

// bindOwnable binds a generic wrapper to an already deployed contract.
func bindOwnable(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(OwnableABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ownable *OwnableRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Ownable.Contract.OwnableCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ownable *OwnableRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ownable.Contract.OwnableTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ownable *OwnableRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ownable.Contract.OwnableTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ownable *OwnableCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Ownable.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ownable *OwnableTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ownable.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ownable *OwnableTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ownable.Contract.contract.Transact(opts, method, params...)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Ownable *OwnableCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Ownable.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Ownable *OwnableSession) Owner() (common.Address, error) {
	return _Ownable.Contract.Owner(&_Ownable.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Ownable *OwnableCallerSession) Owner() (common.Address, error) {
	return _Ownable.Contract.Owner(&_Ownable.CallOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_Ownable *OwnableTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Ownable.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_Ownable *OwnableSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Ownable.Contract.TransferOwnership(&_Ownable.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_Ownable *OwnableTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Ownable.Contract.TransferOwnership(&_Ownable.TransactOpts, newOwner)
}

// TokenABI is the input ABI used to generate the binding from.
const TokenABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"balances\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"remaining\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"}]"

// TokenBin is the compiled bytecode used for deploying new contracts.
const TokenBin = `0x`

// DeployToken deploys a new Ethereum contract, binding an instance of Token to it.
func DeployToken(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Token, error) {
	parsed, err := abi.JSON(strings.NewReader(TokenABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(TokenBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Token{TokenCaller: TokenCaller{contract: contract}, TokenTransactor: TokenTransactor{contract: contract}}, nil
}

// Token is an auto generated Go binding around an Ethereum contract.
type Token struct {
	TokenCaller     // Read-only binding to the contract
	TokenTransactor // Write-only binding to the contract
}

// TokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type TokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TokenSession struct {
	Contract     *Token            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TokenCallerSession struct {
	Contract *TokenCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// TokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TokenTransactorSession struct {
	Contract     *TokenTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type TokenRaw struct {
	Contract *Token // Generic contract binding to access the raw methods on
}

// TokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TokenCallerRaw struct {
	Contract *TokenCaller // Generic read-only contract binding to access the raw methods on
}

// TokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TokenTransactorRaw struct {
	Contract *TokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewToken creates a new instance of Token, bound to a specific deployed contract.
func NewToken(address common.Address, backend bind.ContractBackend) (*Token, error) {
	contract, err := bindToken(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Token{TokenCaller: TokenCaller{contract: contract}, TokenTransactor: TokenTransactor{contract: contract}}, nil
}

// NewTokenCaller creates a new read-only instance of Token, bound to a specific deployed contract.
func NewTokenCaller(address common.Address, caller bind.ContractCaller) (*TokenCaller, error) {
	contract, err := bindToken(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &TokenCaller{contract: contract}, nil
}

// NewTokenTransactor creates a new write-only instance of Token, bound to a specific deployed contract.
func NewTokenTransactor(address common.Address, transactor bind.ContractTransactor) (*TokenTransactor, error) {
	contract, err := bindToken(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &TokenTransactor{contract: contract}, nil
}

// bindToken binds a generic wrapper to an already deployed contract.
func bindToken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(TokenABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Token *TokenRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Token.Contract.TokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Token *TokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Token.Contract.TokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Token *TokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Token.Contract.TokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Token *TokenCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Token.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Token *TokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Token.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Token *TokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Token.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_Token *TokenCaller) Allowance(opts *bind.CallOpts, _owner common.Address, _spender common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Token.contract.Call(opts, out, "allowance", _owner, _spender)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_Token *TokenSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _Token.Contract.Allowance(&_Token.CallOpts, _owner, _spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_Token *TokenCallerSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _Token.Contract.Allowance(&_Token.CallOpts, _owner, _spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_Token *TokenCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Token.contract.Call(opts, out, "balanceOf", _owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_Token *TokenSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _Token.Contract.BalanceOf(&_Token.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_Token *TokenCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _Token.Contract.BalanceOf(&_Token.CallOpts, _owner)
}

// Balances is a free data retrieval call binding the contract method 0x27e235e3.
//
// Solidity: function balances( address) constant returns(uint256)
func (_Token *TokenCaller) Balances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Token.contract.Call(opts, out, "balances", arg0)
	return *ret0, err
}

// Balances is a free data retrieval call binding the contract method 0x27e235e3.
//
// Solidity: function balances( address) constant returns(uint256)
func (_Token *TokenSession) Balances(arg0 common.Address) (*big.Int, error) {
	return _Token.Contract.Balances(&_Token.CallOpts, arg0)
}

// Balances is a free data retrieval call binding the contract method 0x27e235e3.
//
// Solidity: function balances( address) constant returns(uint256)
func (_Token *TokenCallerSession) Balances(arg0 common.Address) (*big.Int, error) {
	return _Token.Contract.Balances(&_Token.CallOpts, arg0)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(success bool)
func (_Token *TokenTransactor) Approve(opts *bind.TransactOpts, _spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Token.contract.Transact(opts, "approve", _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(success bool)
func (_Token *TokenSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Token.Contract.Approve(&_Token.TransactOpts, _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(success bool)
func (_Token *TokenTransactorSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Token.Contract.Approve(&_Token.TransactOpts, _spender, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(success bool)
func (_Token *TokenTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Token.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(success bool)
func (_Token *TokenSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Token.Contract.Transfer(&_Token.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(success bool)
func (_Token *TokenTransactorSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Token.Contract.Transfer(&_Token.TransactOpts, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(success bool)
func (_Token *TokenTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Token.contract.Transact(opts, "transferFrom", _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(success bool)
func (_Token *TokenSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Token.Contract.TransferFrom(&_Token.TransactOpts, _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(success bool)
func (_Token *TokenTransactorSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _Token.Contract.TransferFrom(&_Token.TransactOpts, _from, _to, _value)
}

// WhitelistABI is the input ABI used to generate the binding from.
const WhitelistABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"wallet\",\"type\":\"address\"}],\"name\":\"UnRegisterHub\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"wallet\",\"type\":\"address\"},{\"name\":\"time\",\"type\":\"uint64\"},{\"name\":\"stakeShare\",\"type\":\"uint256\"}],\"name\":\"RegisterMin\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"wallet\",\"type\":\"address\"}],\"name\":\"UnRegisterMiner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"wallet\",\"type\":\"address\"},{\"name\":\"time\",\"type\":\"uint64\"}],\"name\":\"RegisterHub\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"}]"

// WhitelistBin is the compiled bytecode used for deploying new contracts.
const WhitelistBin = `0x`

// DeployWhitelist deploys a new Ethereum contract, binding an instance of Whitelist to it.
func DeployWhitelist(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Whitelist, error) {
	parsed, err := abi.JSON(strings.NewReader(WhitelistABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(WhitelistBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Whitelist{WhitelistCaller: WhitelistCaller{contract: contract}, WhitelistTransactor: WhitelistTransactor{contract: contract}}, nil
}

// Whitelist is an auto generated Go binding around an Ethereum contract.
type Whitelist struct {
	WhitelistCaller     // Read-only binding to the contract
	WhitelistTransactor // Write-only binding to the contract
}

// WhitelistCaller is an auto generated read-only Go binding around an Ethereum contract.
type WhitelistCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WhitelistTransactor is an auto generated write-only Go binding around an Ethereum contract.
type WhitelistTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WhitelistSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type WhitelistSession struct {
	Contract     *Whitelist        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// WhitelistCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type WhitelistCallerSession struct {
	Contract *WhitelistCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// WhitelistTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type WhitelistTransactorSession struct {
	Contract     *WhitelistTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// WhitelistRaw is an auto generated low-level Go binding around an Ethereum contract.
type WhitelistRaw struct {
	Contract *Whitelist // Generic contract binding to access the raw methods on
}

// WhitelistCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type WhitelistCallerRaw struct {
	Contract *WhitelistCaller // Generic read-only contract binding to access the raw methods on
}

// WhitelistTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type WhitelistTransactorRaw struct {
	Contract *WhitelistTransactor // Generic write-only contract binding to access the raw methods on
}

// NewWhitelist creates a new instance of Whitelist, bound to a specific deployed contract.
func NewWhitelist(address common.Address, backend bind.ContractBackend) (*Whitelist, error) {
	contract, err := bindWhitelist(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Whitelist{WhitelistCaller: WhitelistCaller{contract: contract}, WhitelistTransactor: WhitelistTransactor{contract: contract}}, nil
}

// NewWhitelistCaller creates a new read-only instance of Whitelist, bound to a specific deployed contract.
func NewWhitelistCaller(address common.Address, caller bind.ContractCaller) (*WhitelistCaller, error) {
	contract, err := bindWhitelist(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &WhitelistCaller{contract: contract}, nil
}

// NewWhitelistTransactor creates a new write-only instance of Whitelist, bound to a specific deployed contract.
func NewWhitelistTransactor(address common.Address, transactor bind.ContractTransactor) (*WhitelistTransactor, error) {
	contract, err := bindWhitelist(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &WhitelistTransactor{contract: contract}, nil
}

// bindWhitelist binds a generic wrapper to an already deployed contract.
func bindWhitelist(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(WhitelistABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Whitelist *WhitelistRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Whitelist.Contract.WhitelistCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Whitelist *WhitelistRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Whitelist.Contract.WhitelistTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Whitelist *WhitelistRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Whitelist.Contract.WhitelistTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Whitelist *WhitelistCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Whitelist.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Whitelist *WhitelistTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Whitelist.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Whitelist *WhitelistTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Whitelist.Contract.contract.Transact(opts, method, params...)
}

// RegisterHub is a paid mutator transaction binding the contract method 0x94baf301.
//
// Solidity: function RegisterHub(_owner address, wallet address, time uint64) returns(bool)
func (_Whitelist *WhitelistTransactor) RegisterHub(opts *bind.TransactOpts, _owner common.Address, wallet common.Address, time uint64) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "RegisterHub", _owner, wallet, time)
}

// RegisterHub is a paid mutator transaction binding the contract method 0x94baf301.
//
// Solidity: function RegisterHub(_owner address, wallet address, time uint64) returns(bool)
func (_Whitelist *WhitelistSession) RegisterHub(_owner common.Address, wallet common.Address, time uint64) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterHub(&_Whitelist.TransactOpts, _owner, wallet, time)
}

// RegisterHub is a paid mutator transaction binding the contract method 0x94baf301.
//
// Solidity: function RegisterHub(_owner address, wallet address, time uint64) returns(bool)
func (_Whitelist *WhitelistTransactorSession) RegisterHub(_owner common.Address, wallet common.Address, time uint64) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterHub(&_Whitelist.TransactOpts, _owner, wallet, time)
}

// RegisterMin is a paid mutator transaction binding the contract method 0x2edbb4fa.
//
// Solidity: function RegisterMin(_owner address, wallet address, time uint64, stakeShare uint256) returns(bool)
func (_Whitelist *WhitelistTransactor) RegisterMin(opts *bind.TransactOpts, _owner common.Address, wallet common.Address, time uint64, stakeShare *big.Int) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "RegisterMin", _owner, wallet, time, stakeShare)
}

// RegisterMin is a paid mutator transaction binding the contract method 0x2edbb4fa.
//
// Solidity: function RegisterMin(_owner address, wallet address, time uint64, stakeShare uint256) returns(bool)
func (_Whitelist *WhitelistSession) RegisterMin(_owner common.Address, wallet common.Address, time uint64, stakeShare *big.Int) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterMin(&_Whitelist.TransactOpts, _owner, wallet, time, stakeShare)
}

// RegisterMin is a paid mutator transaction binding the contract method 0x2edbb4fa.
//
// Solidity: function RegisterMin(_owner address, wallet address, time uint64, stakeShare uint256) returns(bool)
func (_Whitelist *WhitelistTransactorSession) RegisterMin(_owner common.Address, wallet common.Address, time uint64, stakeShare *big.Int) (*types.Transaction, error) {
	return _Whitelist.Contract.RegisterMin(&_Whitelist.TransactOpts, _owner, wallet, time, stakeShare)
}

// UnRegisterHub is a paid mutator transaction binding the contract method 0x0bdf0962.
//
// Solidity: function UnRegisterHub(_owner address, wallet address) returns(bool)
func (_Whitelist *WhitelistTransactor) UnRegisterHub(opts *bind.TransactOpts, _owner common.Address, wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "UnRegisterHub", _owner, wallet)
}

// UnRegisterHub is a paid mutator transaction binding the contract method 0x0bdf0962.
//
// Solidity: function UnRegisterHub(_owner address, wallet address) returns(bool)
func (_Whitelist *WhitelistSession) UnRegisterHub(_owner common.Address, wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterHub(&_Whitelist.TransactOpts, _owner, wallet)
}

// UnRegisterHub is a paid mutator transaction binding the contract method 0x0bdf0962.
//
// Solidity: function UnRegisterHub(_owner address, wallet address) returns(bool)
func (_Whitelist *WhitelistTransactorSession) UnRegisterHub(_owner common.Address, wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterHub(&_Whitelist.TransactOpts, _owner, wallet)
}

// UnRegisterMiner is a paid mutator transaction binding the contract method 0x7efad052.
//
// Solidity: function UnRegisterMiner(_owner address, wallet address) returns(bool)
func (_Whitelist *WhitelistTransactor) UnRegisterMiner(opts *bind.TransactOpts, _owner common.Address, wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.contract.Transact(opts, "UnRegisterMiner", _owner, wallet)
}

// UnRegisterMiner is a paid mutator transaction binding the contract method 0x7efad052.
//
// Solidity: function UnRegisterMiner(_owner address, wallet address) returns(bool)
func (_Whitelist *WhitelistSession) UnRegisterMiner(_owner common.Address, wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterMiner(&_Whitelist.TransactOpts, _owner, wallet)
}

// UnRegisterMiner is a paid mutator transaction binding the contract method 0x7efad052.
//
// Solidity: function UnRegisterMiner(_owner address, wallet address) returns(bool)
func (_Whitelist *WhitelistTransactorSession) UnRegisterMiner(_owner common.Address, wallet common.Address) (*types.Transaction, error) {
	return _Whitelist.Contract.UnRegisterMiner(&_Whitelist.TransactOpts, _owner, wallet)
}
