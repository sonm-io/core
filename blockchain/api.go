package blockchain

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/noxiouz/zapctx/ctxlog"
	marketAPI "github.com/sonm-io/core/blockchain/source/api"
	sonm "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

const (
	txRetryTimes     = 20
	txRetryDelayTime = 100 * time.Millisecond
	//TODO: make configurable
	blockGenPeriod = 4 * time.Second
)

type API interface {
	ProfileRegistry() ProfileRegistryAPI
	Events() EventsAPI
	Market() MarketAPI
	Blacklist() BlacklistAPI
	MasterchainToken() TokenAPI
	SidechainToken() TokenAPI
	TestToken() TestTokenAPI
	OracleUSD() OracleAPI
	MasterchainGate() SimpleGatekeeperAPI
	SidechainGate() SimpleGatekeeperAPI
	OracleMultiSig() MultiSigAPI
	ContractRegistry() ContractRegistry
}

type ContractRegistry interface {
	SidechainSNMAddress() common.Address
	MasterchainSNMAddress() common.Address
	BlacklistAddress() common.Address
	MarketAddress() common.Address
	ProfileRegistryAddress() common.Address
	OracleUsdAddress() common.Address
	GatekeeperMasterchainAddress() common.Address
	GatekeeperSidechainAddress() common.Address
	TestnetFaucetAddress() common.Address
	OracleMultiSig() common.Address
}

type ProfileRegistryAPI interface {
	AddValidator(ctx context.Context, key *ecdsa.PrivateKey, validator common.Address, level int8) (*types.Transaction, error)
	RemoveValidator(ctx context.Context, key *ecdsa.PrivateKey, validator common.Address) (*types.Transaction, error)
	GetValidator(ctx context.Context, validatorID common.Address) (*sonm.Validator, error)
	GetValidatorLevel(ctx context.Context, validatorID common.Address) (int8, error)
	CreateCertificate(ctx context.Context, key *ecdsa.PrivateKey, owner common.Address, attributeType *big.Int, value []byte) (*types.Transaction, error)
	RemoveCertificate(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) error
	GetCertificate(ctx context.Context, certificateID *big.Int) (*sonm.Certificate, error)
	GetAttributeCount(ctx context.Context, owner common.Address, attributeType *big.Int) (*big.Int, error)
	GetAttributeValue(ctx context.Context, owner common.Address, attributeType *big.Int) ([]byte, error)
	GetProfileLevel(ctx context.Context, owner common.Address) (sonm.IdentityLevel, error)
}

type EventsAPI interface {
	GetEvents(ctx context.Context, filter *EventFilter) (chan *Event, error)
	GetLastBlock(ctx context.Context) (uint64, error)
	GetMarketFilter(fromBlock *big.Int) *EventFilter
	GetMultiSigFilter(addresses []common.Address, fromBlock *big.Int) *EventFilter
}

type MarketAPI interface {
	QuickBuy(ctx context.Context, key *ecdsa.PrivateKey, askId *big.Int, duration uint64) (*sonm.Deal, error)
	OpenDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bigID *big.Int) (*sonm.Deal, error)
	CloseDeal(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int, blacklistType sonm.BlacklistType) error
	GetDealInfo(ctx context.Context, dealID *big.Int) (*sonm.Deal, error)
	GetDealsAmount(ctx context.Context) (*big.Int, error)
	PlaceOrder(ctx context.Context, key *ecdsa.PrivateKey, order *sonm.Order) (*sonm.Order, error)
	CancelOrder(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) error
	GetOrderInfo(ctx context.Context, orderID *big.Int) (*sonm.Order, error)
	GetOrdersAmount(ctx context.Context) (*big.Int, error)
	Bill(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int) error
	RegisterWorker(ctx context.Context, key *ecdsa.PrivateKey, master common.Address) error
	ConfirmWorker(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address) error
	RemoveWorker(ctx context.Context, key *ecdsa.PrivateKey, master, slave common.Address) error
	GetMaster(ctx context.Context, slave common.Address) (common.Address, error)
	GetDealChangeRequestInfo(ctx context.Context, id *big.Int) (*sonm.DealChangeRequest, error)
	CreateChangeRequest(ctx context.Context, key *ecdsa.PrivateKey, request *sonm.DealChangeRequest) (*big.Int, error)
	CancelChangeRequest(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) error
	GetNumBenchmarks(ctx context.Context) (uint64, error)
}

type BlacklistAPI interface {
	Check(ctx context.Context, who, whom common.Address) (bool, error)
	Add(ctx context.Context, key *ecdsa.PrivateKey, who, whom common.Address) (*types.Transaction, error)
	Remove(ctx context.Context, key *ecdsa.PrivateKey, whom common.Address) error
	AddMaster(ctx context.Context, key *ecdsa.PrivateKey, root common.Address) (*types.Transaction, error)
	RemoveMaster(ctx context.Context, key *ecdsa.PrivateKey, root common.Address) (*types.Transaction, error)
	SetMarketAddress(ctx context.Context, key *ecdsa.PrivateKey, market common.Address) (*types.Transaction, error)
}

// TokenAPI is a go implementation of ERC20-compatibility token with full functionality high-level interface
// standard description with placed: https://github.com/ethereum/EIPs/blob/master/EIPS/eip-20-token-standard.md
type TokenAPI interface {
	// Approve allows managing some amount of your tokens by given address.
	// Note that setting an allowance value is prohibited if current allowance value is non-zero.
	// This behavior will persist until migrate our token to the ERC827 standard.
	// You're forced to reset current allowance value for your address, then set a new value.
	// This method performs resetting an old allowance value and setting a new one.
	Approve(ctx context.Context, key *ecdsa.PrivateKey, to common.Address, amount *big.Int) error
	// ApproveAtLeast acts as Approve, but change allowance only if it less than required.
	ApproveAtLeast(ctx context.Context, key *ecdsa.PrivateKey, to common.Address, amount *big.Int) error
	// Transfer token from caller
	Transfer(ctx context.Context, key *ecdsa.PrivateKey, to common.Address, amount *big.Int) error
	// TransferFrom fallback function for contracts to transfer you allowance
	TransferFrom(ctx context.Context, key *ecdsa.PrivateKey, from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error)
	// BalanceOf returns balance of given address
	BalanceOf(ctx context.Context, address common.Address) (*Balance, error)
	// AllowanceOf returns allowance of given address to spender account
	AllowanceOf(ctx context.Context, from, to common.Address) (*big.Int, error)
	// TotalSupply - all amount of emitted token
	TotalSupply(ctx context.Context) (*big.Int, error)
	// IncreaseApproval increase the amount of tokens that an owner allowed to a spender
	// approve should be called when current allowance = 0. To increment
	// allowed value is better to use this function to avoid 2 calls (and wait until the first transaction is mined)
	IncreaseApproval(ctx context.Context, key *ecdsa.PrivateKey, spender common.Address, value *big.Int) error
	// DecreaseApproval - decrease the amount of tokens that an owner allowed to a spender
	// approve should be called when current allowance = 0. To decrement
	// allowed value is better to use this function to avoid 2 calls (and wait until the first transaction is mined)
	DecreaseApproval(ctx context.Context, key *ecdsa.PrivateKey, spender common.Address, value *big.Int) error
}

type TestTokenAPI interface {
	// GetTokens - send 100 SNMT token for message caller
	// this function added for MVP purposes and has been deleted later
	GetTokens(ctx context.Context, key *ecdsa.PrivateKey) (*types.Transaction, error)
}

// OracleAPI manage price relation between some currency and SNM token
type OracleAPI interface {
	// SetCurrentPrice sets current price relation between some currency and SONM token
	SetCurrentPrice(ctx context.Context, key *ecdsa.PrivateKey, price *big.Int) (*types.Transaction, error)
	// GetCurrentPrice returns current price relation between some currency and SONM token
	GetCurrentPrice(ctx context.Context) (*big.Int, error)
	// PackSetCurrentPriceTransactionData pack `SetCurrentPrice` method call
	PackSetCurrentPriceTransactionData(price *big.Int) ([]byte, error)
	// UnpackSetCurrentPriceTransactionData unpack `SetCurrentPrice` method call
	UnpackSetCurrentPriceTransactionData(data []byte) (*big.Int, error)
	// GetOwner get owner address
	GetOwner(ctx context.Context) (common.Address, error)
	// SetOwner set owner address, owner can change oracle price
	SetOwner(ctx context.Context, key *ecdsa.PrivateKey, owner common.Address) error
}

// SimpleGatekeeperAPI facade to interact with deposit/withdraw functions through gates
type SimpleGatekeeperAPI interface {
	// PayIn grab sender tokens and signal gate to transfer it to mirrored chain.
	// On Masterchain ally as `Deposit`
	// On Sidecain ally as `Withdraw`
	PayIn(ctx context.Context, key *ecdsa.PrivateKey, value *big.Int) error
	// Payout release payout transaction from mirrored chain.
	// Accessible only by owner.
	Payout(ctx context.Context, key *ecdsa.PrivateKey, to common.Address, value *big.Int, txNumber *big.Int) (*types.Transaction, error)
	// Kill calls contract to suicide, all ether and tokens funds transfer to owner.
	// Accessible only by owner.
	Kill(ctx context.Context, key *ecdsa.PrivateKey) (*types.Transaction, error)
}

type MultiSigAPI interface {
	AddOwner(ctx context.Context, key *ecdsa.PrivateKey, owner common.Address) error
	RemoveOwner(ctx context.Context, key *ecdsa.PrivateKey, owner common.Address) error
	ReplaceOwner(ctx context.Context, key *ecdsa.PrivateKey, owner common.Address, newOwner common.Address) error
	ChangeRequirement(ctx context.Context, key *ecdsa.PrivateKey, required *big.Int) error
	SubmitTransaction(ctx context.Context, key *ecdsa.PrivateKey, destination common.Address, value *big.Int, data []byte) error
	ConfirmTransaction(ctx context.Context, key *ecdsa.PrivateKey, transactionID *big.Int) error
	RevokeConfirmation(ctx context.Context, key *ecdsa.PrivateKey, transactionID *big.Int) error
	ExecuteTransaction(ctx context.Context, key *ecdsa.PrivateKey, transactionID *big.Int) error
	IsConfirmed(ctx context.Context, transactionID *big.Int) (bool, error)
	GetConfirmationCount(ctx context.Context, transactionID *big.Int) (*big.Int, error)
	GetTransactionCount(ctx context.Context, penging bool, executed bool) (*big.Int, error)
	GetOwners(ctx context.Context) ([]common.Address, error)
	GetConfirmations(ctx context.Context, transactionID *big.Int) ([]common.Address, error)
	GetTransactionIDs(ctx context.Context, from *big.Int, to *big.Int, pending bool, executed bool) ([]*big.Int, error)
	GetTransaction(ctx context.Context, transactionID *big.Int) (*MultiSigTransactionData, error)
}

type BasicAPI struct {
	options          *options
	contractRegistry ContractRegistry
	masterchainToken TokenAPI
	sidechainToken   TokenAPI
	testToken        TestTokenAPI
	blacklist        BlacklistAPI
	market           MarketAPI
	profileRegistry  ProfileRegistryAPI
	events           EventsAPI
	oracle           OracleAPI
	masterchainGate  SimpleGatekeeperAPI
	sidechainGate    SimpleGatekeeperAPI
	oracleMultiSig   MultiSigAPI
}

func NewAPI(ctx context.Context, opts ...Option) (API, error) {
	defaults := defaultOptions()
	for _, o := range opts {
		o(defaults)
	}

	api := &BasicAPI{options: defaults}

	setup := []func(ctx context.Context) error{
		api.setupContractRegistry,
		api.setupMasterchainToken,
		api.setupSidechainToken,
		api.setupTestToken,
		api.setupBlacklist,
		api.setupMarket,
		api.setupProfileRegistry,
		api.setupEvents,
		api.setupOracle,
		api.setupMasterchainGate,
		api.setupSidechainGate,
		api.setupOracleMultiSig,
	}

	for _, setupFunc := range setup {
		if err := setupFunc(ctx); err != nil {
			return nil, err
		}
	}
	return api, nil
}

func (api *BasicAPI) setupContractRegistry(ctx context.Context) error {
	registry, err := NewRegistry(ctx, api.options.contractRegistry, api.options.sidechain)
	if err != nil {
		return fmt.Errorf("failed to setup contract registry: %s", err)
	}
	api.contractRegistry = registry
	return nil
}

func (api *BasicAPI) setupMasterchainToken(ctx context.Context) error {
	masterchainTokenAddr := api.contractRegistry.MasterchainSNMAddress()
	masterchainToken, err := NewStandardToken(masterchainTokenAddr, api.options.masterchain)
	if err != nil {
		return fmt.Errorf("failed to setup masterchain token: %s", err)
	}
	api.masterchainToken = masterchainToken

	return nil
}

func (api *BasicAPI) setupSidechainToken(ctx context.Context) error {
	sidechainTokenAddr := api.contractRegistry.SidechainSNMAddress()
	sidechainToken, err := NewStandardToken(sidechainTokenAddr, api.options.sidechain)
	if err != nil {
		return fmt.Errorf("failed to setup sidechain token: %s", err)
	}
	api.sidechainToken = sidechainToken

	return nil
}

func (api *BasicAPI) setupTestToken(ctx context.Context) error {
	masterchainTokenAddr := api.contractRegistry.MasterchainSNMAddress()
	testToken, err := NewTestToken(masterchainTokenAddr, api.options.masterchain)
	if err != nil {
		return fmt.Errorf("failed to setup test token: %s", err)
	}
	api.testToken = testToken
	return nil
}

func (api *BasicAPI) setupBlacklist(ctx context.Context) error {
	blacklistAddr := api.contractRegistry.BlacklistAddress()
	blacklist, err := NewBasicBlacklist(blacklistAddr, api.options.sidechain)
	if err != nil {
		return fmt.Errorf("failed to setup blacklist: %s", err)
	}
	api.blacklist = blacklist
	return nil
}

func (api *BasicAPI) setupMarket(ctx context.Context) error {
	marketAddr := api.contractRegistry.MarketAddress()
	market, err := NewBasicMarket(marketAddr, api.sidechainToken, api.options.sidechain)
	if err != nil {
		return fmt.Errorf("failed to setup market: %s", err)
	}
	api.market = market
	return nil
}

func (api *BasicAPI) setupProfileRegistry(ctx context.Context) error {
	profileRegistryAddr := api.contractRegistry.ProfileRegistryAddress()
	profileRegistry, err := NewProfileRegistry(profileRegistryAddr, api.options.sidechain)
	if err != nil {
		return fmt.Errorf("failed to setup profile registry: %s", err)
	}
	api.profileRegistry = profileRegistry
	return nil
}

func (api *BasicAPI) setupEvents(ctx context.Context) error {
	events, err := NewEventsAPI(api, api.options.sidechain, ctxlog.GetLogger(ctx), api.options.blocksBatchSize)
	if err != nil {
		return fmt.Errorf("failed to setup events: %s", err)
	}
	api.events = events
	return nil
}

func (api *BasicAPI) setupOracle(ctx context.Context) error {
	oracleAddr := api.contractRegistry.OracleUsdAddress()
	oracle, err := NewOracleUSDAPI(oracleAddr, api.options.sidechain)
	if err != nil {
		return fmt.Errorf("failed to setup oracle: %s", err)
	}
	api.oracle = oracle
	return nil
}

func (api *BasicAPI) setupOracleMultiSig(ctx context.Context) error {
	multiSigAddr := api.contractRegistry.OracleMultiSig()
	oracleMultiSig, err := NewMultiSigAPI(multiSigAddr, api.options.sidechain)
	if err != nil {
		return fmt.Errorf("failed to setup oracle: %s", err)
	}
	api.oracleMultiSig = oracleMultiSig
	return nil
}

func (api *BasicAPI) setupMasterchainGate(ctx context.Context) error {
	gatekeeperMasterchainAddr := api.contractRegistry.GatekeeperMasterchainAddress()
	masterchainGate, err := NewSimpleGatekeeper(gatekeeperMasterchainAddr, api.options.masterchain)
	if err != nil {
		return fmt.Errorf("failed to setup masterchain gatekeeper: %s", err)
	}
	api.masterchainGate = masterchainGate
	return nil
}

func (api *BasicAPI) setupSidechainGate(ctx context.Context) error {
	gatekeeperSidechainAddr := api.contractRegistry.GatekeeperSidechainAddress()
	sidechainGate, err := NewSimpleGatekeeper(gatekeeperSidechainAddr, api.options.sidechain)
	if err != nil {
		return fmt.Errorf("failed to setup sidechain gatekeeper: %s", err)
	}
	api.sidechainGate = sidechainGate
	return nil
}

func (api *BasicAPI) Market() MarketAPI {
	if api.options.niceMarket {
		return &niceMarketAPI{
			MarketAPI: api.market,
			profiles:  api.profileRegistry,
			blacklist: api.blacklist,
		}
	}
	return api.market
}

func (api *BasicAPI) MasterchainToken() TokenAPI {
	return api.masterchainToken
}

func (api *BasicAPI) SidechainToken() TokenAPI {
	return api.sidechainToken
}

func (api *BasicAPI) TestToken() TestTokenAPI {
	return api.testToken
}

func (api *BasicAPI) Blacklist() BlacklistAPI {
	return api.blacklist
}

func (api *BasicAPI) ProfileRegistry() ProfileRegistryAPI {
	return api.profileRegistry
}

func (api *BasicAPI) Events() EventsAPI {
	return api.events
}

func (api *BasicAPI) OracleUSD() OracleAPI {
	return api.oracle
}

func (api *BasicAPI) OracleMultiSig() MultiSigAPI {
	return api.oracleMultiSig
}

func (api *BasicAPI) MasterchainGate() SimpleGatekeeperAPI {
	return api.masterchainGate
}

func (api *BasicAPI) SidechainGate() SimpleGatekeeperAPI {
	return api.sidechainGate
}

func (api *BasicAPI) ContractRegistry() ContractRegistry {
	return api.contractRegistry
}

func NewRegistry(ctx context.Context, address common.Address, opts *chainOpts) (*BasicContractRegistry, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	contract, err := marketAPI.NewAddressHashMap(address, client)
	if err != nil {
		return nil, err
	}

	registry := &BasicContractRegistry{
		registryContract: contract,
	}
	if err := registry.setup(ctx); err != nil {
		return nil, err
	}
	return registry, nil
}

type BasicContractRegistry struct {
	sidechainSNMAddress          common.Address
	masterchainSNMAddress        common.Address
	blacklistAddress             common.Address
	marketAddress                common.Address
	profileRegistryAddress       common.Address
	oracleUsdAddress             common.Address
	gatekeeperMasterchainAddress common.Address
	gatekeeperSidechainAddress   common.Address
	testnetFaucetAddress         common.Address
	oracleMultiSigAddress        common.Address

	registryContract *marketAPI.AddressHashMap
}

func registryKey(key string) [32]byte {
	if len(key) > 32 {
		panic("registry key exceeds 32 byte limit")
	}
	result := [32]byte{}
	copy(result[:], key)
	return result
}

func (m *BasicContractRegistry) readContract(ctx context.Context, key string, target *common.Address) error {
	data, err := m.registryContract.Read(getCallOptions(ctx), registryKey(key))
	if err != nil {
		return err
	}
	*target = data
	return nil
}

func (m *BasicContractRegistry) setup(ctx context.Context) error {
	addresses := []struct {
		key    string
		target *common.Address
	}{
		{sidechainSNMAddressKey, &m.sidechainSNMAddress},
		{masterchainSNMAddressKey, &m.masterchainSNMAddress},
		{blacklistAddressKey, &m.blacklistAddress},
		{marketAddressKey, &m.marketAddress},
		{profileRegistryAddressKey, &m.profileRegistryAddress},
		{oracleUsdAddressKey, &m.oracleUsdAddress},
		{gatekeeperMasterchainAddressKey, &m.gatekeeperMasterchainAddress},
		{gatekeeperSidechainAddressKey, &m.gatekeeperSidechainAddress},
		{testnetFaucetAddressKey, &m.testnetFaucetAddress},
		{oracleMultiSigAddressKey, &m.oracleMultiSigAddress},
	}

	for _, param := range addresses {
		if err := m.readContract(ctx, param.key, param.target); err != nil {
			return err
		}
		ctxlog.S(ctx).Infof("fetched `%s` contract address: %s", param.key, param.target.String())
	}
	return nil
}

func (m *BasicContractRegistry) SidechainSNMAddress() common.Address {
	return m.sidechainSNMAddress
}

func (m *BasicContractRegistry) MasterchainSNMAddress() common.Address {
	return m.masterchainSNMAddress
}

func (m *BasicContractRegistry) BlacklistAddress() common.Address {
	return m.blacklistAddress
}

func (m *BasicContractRegistry) MarketAddress() common.Address {
	return m.marketAddress
}

func (m *BasicContractRegistry) ProfileRegistryAddress() common.Address {
	return m.profileRegistryAddress
}

func (m *BasicContractRegistry) OracleUsdAddress() common.Address {
	return m.oracleUsdAddress
}

func (m *BasicContractRegistry) GatekeeperMasterchainAddress() common.Address {
	return m.gatekeeperMasterchainAddress
}

func (m *BasicContractRegistry) GatekeeperSidechainAddress() common.Address {
	return m.gatekeeperSidechainAddress
}

func (m *BasicContractRegistry) TestnetFaucetAddress() common.Address {
	return m.testnetFaucetAddress
}

func (m *BasicContractRegistry) OracleMultiSig() common.Address {
	return m.oracleMultiSigAddress
}

type BasicMarketAPI struct {
	client             CustomEthereumClient
	token              TokenAPI
	marketContractAddr common.Address
	marketContract     *marketAPI.Market
	opts               *chainOpts
}

func NewBasicMarket(address common.Address, token TokenAPI, opts *chainOpts) (MarketAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	marketContract, err := marketAPI.NewMarket(address, client)
	if err != nil {
		return nil, err
	}

	return &BasicMarketAPI{
		client:             client,
		token:              token,
		marketContractAddr: address,
		marketContract:     marketContract,
		opts:               opts,
	}, nil
}

func txRetryWrapper(fn func() (*types.Transaction, error)) (*types.Transaction, error) {
	var err error
	var tx *types.Transaction

	for i := 0; i < txRetryTimes; i++ {
		tx, err = fn()

		if err == nil {
			break
		} else {
			if err.Error() != "replacement transaction underpriced" &&
				err.Error() != "nonce too low" &&
				!strings.Contains(err.Error(), "known transaction:") {
				break
			} else {
				time.Sleep(txRetryDelayTime)
			}
		}
	}

	return tx, err
}

func (api *BasicMarketAPI) checkAllowance(ctx context.Context, key *ecdsa.PrivateKey) error {
	maxAllowance := big.NewInt(1)
	maxAllowance = maxAllowance.Lsh(maxAllowance, 256)
	maxAllowance = maxAllowance.Sub(maxAllowance, big.NewInt(1))
	minAllowance := big.NewInt(0).Div(maxAllowance, big.NewInt(2))
	curAllowance, err := api.token.AllowanceOf(ctx, crypto.PubkeyToAddress(key.PublicKey), api.marketContractAddr)
	if err != nil {
		return fmt.Errorf("failed to get allowance: %s", err)
	}
	if curAllowance.Cmp(minAllowance) < 0 {
		if err := api.token.Approve(ctx, key, api.marketContractAddr, maxAllowance); err != nil {
			return fmt.Errorf("failed to set allowance: %s", err)
		}
	}
	return nil
}

func (api *BasicMarketAPI) QuickBuy(ctx context.Context, key *ecdsa.PrivateKey, askId *big.Int, duration uint64) (*sonm.Deal, error) {
	if err := api.checkAllowance(ctx, key); err != nil {
		return nil, err
	}
	opts := api.opts.getTxOpts(ctx, key, quickBuyGasLimit)
	tx, err := api.marketContract.QuickBuy(opts, askId, big.NewInt(0).SetUint64(duration))
	if err != nil {
		return nil, err
	}

	return api.extractOpenDealData(ctx, tx)
}

func (api *BasicMarketAPI) OpenDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bidID *big.Int) (*sonm.Deal, error) {
	opts := api.opts.getTxOpts(ctx, key, openDealGasLimit)
	tx, err := txRetryWrapper(func() (*types.Transaction, error) {
		return api.marketContract.OpenDeal(opts, askID, bidID)
	})
	if err != nil {
		return nil, err
	}

	return api.extractOpenDealData(ctx, tx)
}

func (api *BasicMarketAPI) extractOpenDealData(ctx context.Context, tx *types.Transaction) (*sonm.Deal, error) {
	logs, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, DealOpenedTopic)
	if err != nil {
		return nil, err
	}

	id, err := extractBig(logs.Topics, 1)
	if err != nil {
		return nil, err
	}

	return api.GetDealInfo(ctx, id)
}

func (api *BasicMarketAPI) CloseDeal(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int, blacklistType sonm.BlacklistType) error {
	if err := api.checkAllowance(ctx, key); err != nil {
		return err
	}
	opts := api.opts.getTxOpts(ctx, key, closeDealGasLimit)
	tx, err := txRetryWrapper(func() (*types.Transaction, error) {
		return api.marketContract.CloseDeal(opts, dealID, uint8(blacklistType))
	})
	if err != nil {
		return err
	}

	if _, err = WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, DealUpdatedTopic); err != nil {
		return err
	}

	return nil
}

func (api *BasicMarketAPI) GetDealInfo(ctx context.Context, dealID *big.Int) (*sonm.Deal, error) {
	deal1, err := api.marketContract.GetDealInfo(getCallOptions(ctx), dealID)
	if err != nil {
		return nil, err
	}

	noAsk := deal1.AskID.Cmp(big.NewInt(0)) == 0
	noBid := deal1.BidID.Cmp(big.NewInt(0)) == 0
	if noAsk && noBid {
		return nil, fmt.Errorf("no deal with id = %s", dealID.String())
	}

	deal2, err := api.marketContract.GetDealParams(getCallOptions(ctx), dealID)
	if err != nil {
		return nil, err
	}
	if deal2.Status == 0 {
		return nil, fmt.Errorf("deal fetching inconsistency for deal %s", dealID.String())
	}

	benchmarks, err := sonm.NewBenchmarks(deal1.Benchmarks)
	if err != nil {
		return nil, err
	}

	return &sonm.Deal{
		Id:             sonm.NewBigInt(dealID),
		Benchmarks:     benchmarks,
		SupplierID:     sonm.NewEthAddress(deal1.SupplierID),
		ConsumerID:     sonm.NewEthAddress(deal1.ConsumerID),
		MasterID:       sonm.NewEthAddress(deal1.MasterID),
		AskID:          sonm.NewBigInt(deal1.AskID),
		BidID:          sonm.NewBigInt(deal1.BidID),
		Duration:       deal2.Duration.Uint64(),
		Price:          sonm.NewBigInt(deal2.Price),
		StartTime:      &sonm.Timestamp{Seconds: deal1.StartTime.Int64()},
		EndTime:        &sonm.Timestamp{Seconds: deal2.EndTime.Int64()},
		Status:         sonm.DealStatus(deal2.Status),
		BlockedBalance: sonm.NewBigInt(deal2.BlockedBalance),
		TotalPayout:    sonm.NewBigInt(deal2.TotalPayout),
		LastBillTS:     &sonm.Timestamp{Seconds: deal2.LastBillTS.Int64()},
	}, nil
}

func (api *BasicMarketAPI) GetDealsAmount(ctx context.Context) (*big.Int, error) {
	return api.marketContract.GetDealsAmount(getCallOptions(ctx))
}

func (api *BasicMarketAPI) PlaceOrder(ctx context.Context, key *ecdsa.PrivateKey, order *sonm.Order) (*sonm.Order, error) {
	if err := api.checkAllowance(ctx, key); err != nil {
		return nil, err
	}
	opts := api.opts.getTxOpts(ctx, key, placeOrderGasLimit)

	var fixedTag [32]byte
	copy(fixedTag[:], order.Tag[:])

	tx, err := txRetryWrapper(func() (*types.Transaction, error) {
		return api.marketContract.PlaceOrder(opts,
			uint8(order.OrderType),
			order.CounterpartyID.Unwrap(),
			big.NewInt(int64(order.Duration)),
			order.Price.Unwrap(),
			order.Netflags.ToBoolSlice(),
			uint8(order.IdentityLevel),
			common.HexToAddress(order.Blacklist),
			fixedTag,
			order.GetBenchmarks().ToArray())
	})
	if err != nil {
		return nil, err
	}

	logs, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, OrderPlacedTopic)
	if err != nil {
		return nil, err
	}

	id, err := extractBig(logs.Topics, 1)
	if err != nil {
		return nil, err
	}

	return api.GetOrderInfo(ctx, id)
}

func (api *BasicMarketAPI) CancelOrder(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) error {
	opts := api.opts.getTxOpts(ctx, key, cancelOrderGasLimit)
	tx, err := txRetryWrapper(func() (*types.Transaction, error) {
		return api.marketContract.CancelOrder(opts, id)
	})
	if err != nil {
		return err
	}

	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, OrderUpdatedTopic); err != nil {
		return err
	}

	return nil
}

func (api *BasicMarketAPI) GetOrderInfo(ctx context.Context, orderID *big.Int) (*sonm.Order, error) {
	order1, err := api.marketContract.GetOrderInfo(getCallOptions(ctx), orderID)
	if err != nil {
		return nil, err
	}

	noAuthor := order1.Author.Big().Cmp(big.NewInt(0)) == 0
	noType := sonm.OrderType(order1.OrderType) == sonm.OrderType_ANY

	if noAuthor && noType {
		return nil, fmt.Errorf("no order with id = %s", orderID.String())
	}

	order2, err := api.marketContract.GetOrderParams(getCallOptions(ctx), orderID)
	if err != nil {
		return nil, err
	}
	if order2.OrderStatus == 0 {
		return nil, fmt.Errorf("order fetching inconsistency for order %s", orderID.String())
	}

	netflags := sonm.NetFlagsFromBoolSlice(order1.Netflags[:])

	benchmarks, err := sonm.NewBenchmarks(order1.Benchmarks)
	if err != nil {
		return nil, err
	}

	return &sonm.Order{
		Id:             sonm.NewBigInt(orderID),
		DealID:         sonm.NewBigInt(order2.DealID),
		OrderType:      sonm.OrderType(order1.OrderType),
		OrderStatus:    sonm.OrderStatus(order2.OrderStatus),
		AuthorID:       sonm.NewEthAddress(order1.Author),
		CounterpartyID: sonm.NewEthAddress(order1.Counterparty),
		Duration:       order1.Duration.Uint64(),
		Price:          sonm.NewBigInt(order1.Price),
		Netflags:       netflags,
		IdentityLevel:  sonm.IdentityLevel(order1.IdentityLevel),
		Blacklist:      order1.Blacklist.String(),
		Tag:            order1.Tag[:],
		Benchmarks:     benchmarks,
		FrozenSum:      sonm.NewBigInt(order1.FrozenSum),
	}, nil
}

func (api *BasicMarketAPI) GetOrdersAmount(ctx context.Context) (*big.Int, error) {
	return api.marketContract.GetOrdersAmount(getCallOptions(ctx))
}

func (api *BasicMarketAPI) Bill(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int) error {
	if err := api.checkAllowance(ctx, key); err != nil {
		return err
	}
	opts := api.opts.getTxOpts(ctx, key, billGasLimit)
	tx, err := txRetryWrapper(func() (*types.Transaction, error) {
		return api.marketContract.Bill(opts, dealID)
	})
	if err != nil {
		return err
	}

	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, BilledTopic); err != nil {
		return err
	}

	return nil
}

func (api *BasicMarketAPI) RegisterWorker(ctx context.Context, key *ecdsa.PrivateKey, master common.Address) error {
	opts := api.opts.getTxOpts(ctx, key, registerWorkerGasLimit)
	tx, err := txRetryWrapper(func() (*types.Transaction, error) {
		return api.marketContract.RegisterWorker(opts, master)
	})
	if err != nil {
		return err
	}

	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, WorkerAnnouncedTopic); err != nil {
		return err
	}

	return nil
}

func (api *BasicMarketAPI) ConfirmWorker(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address) error {
	opts := api.opts.getTxOpts(ctx, key, confirmWorkerGasLimit)
	tx, err := txRetryWrapper(func() (*types.Transaction, error) {
		return api.marketContract.ConfirmWorker(opts, slave)
	})
	if err != nil {
		return err
	}

	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, WorkerConfirmedTopic); err != nil {
		return err
	}

	return nil
}

func (api *BasicMarketAPI) RemoveWorker(ctx context.Context, key *ecdsa.PrivateKey, master, slave common.Address) error {
	opts := api.opts.getTxOpts(ctx, key, removeWorkerGasLimit)
	tx, err := txRetryWrapper(func() (*types.Transaction, error) {
		return api.marketContract.RemoveWorker(opts, master, slave)
	})
	if err != nil {
		return err
	}

	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, WorkerRemovedTopic); err != nil {
		return err
	}

	return nil
}

func (api *BasicMarketAPI) GetMaster(ctx context.Context, slave common.Address) (common.Address, error) {
	return api.marketContract.GetMaster(getCallOptions(ctx), slave)
}

func (api *BasicMarketAPI) GetDealChangeRequestInfo(ctx context.Context, changeRequestID *big.Int) (*sonm.DealChangeRequest, error) {
	changeRequest, err := api.marketContract.GetChangeRequestInfo(getCallOptions(ctx), changeRequestID)
	if err != nil {
		return nil, err
	}

	return &sonm.DealChangeRequest{
		Id:          sonm.NewBigInt(changeRequestID),
		DealID:      sonm.NewBigInt(changeRequest.DealID),
		RequestType: sonm.OrderType(changeRequest.RequestType),
		Duration:    changeRequest.Duration.Uint64(),
		Price:       sonm.NewBigInt(changeRequest.Price),
		Status:      sonm.ChangeRequestStatus(changeRequest.Status),
	}, nil
}

func (api *BasicMarketAPI) CreateChangeRequest(ctx context.Context, key *ecdsa.PrivateKey, req *sonm.DealChangeRequest) (*big.Int, error) {
	if err := api.checkAllowance(ctx, key); err != nil {
		return nil, err
	}
	duration := big.NewInt(int64(req.GetDuration()))
	opts := api.opts.getTxOpts(ctx, key, createChangeRequestGasLimit)
	tx, err := txRetryWrapper(func() (*types.Transaction, error) {
		return api.marketContract.CreateChangeRequest(opts, req.GetDealID().Unwrap(), req.GetPrice().Unwrap(), duration)
	})
	if err != nil {
		return nil, err
	}

	logs, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, DealChangeRequestSentTopic)
	if err != nil {
		return nil, err
	}

	id, err := extractBig(logs.Topics, 1)
	if err != nil {
		return nil, fmt.Errorf("cannot extract change request id from transaction logs: %v", err)
	}

	return id, nil
}

func (api *BasicMarketAPI) CancelChangeRequest(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) error {
	opts := api.opts.getTxOpts(ctx, key, cancelChangeRequestGasLimit)
	tx, err := txRetryWrapper(func() (*types.Transaction, error) {
		return api.marketContract.CancelChangeRequest(opts, id)
	})
	if err != nil {
		return err
	}

	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, DealChangeRequestUpdatedTopic); err != nil {
		return err
	}

	return nil
}

func (api *BasicMarketAPI) GetNumBenchmarks(ctx context.Context) (uint64, error) {
	num, err := api.marketContract.GetBenchmarksQuantity(getCallOptions(ctx))
	if err != nil {
		return 0, err
	}
	if !num.IsUint64() {
		return 0, errors.New("benchmarks quantity overflows int64")
	}
	return num.Uint64(), nil
}

type ProfileRegistry struct {
	client                  CustomEthereumClient
	profileRegistryContract *marketAPI.ProfileRegistry
	opts                    *chainOpts
}

func NewProfileRegistry(address common.Address, opts *chainOpts) (ProfileRegistryAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	profileRegistryContract, err := marketAPI.NewProfileRegistry(address, client)
	if err != nil {
		return nil, err
	}

	return &ProfileRegistry{
		client:                  client,
		profileRegistryContract: profileRegistryContract,
		opts:                    opts,
	}, nil
}

func (api *ProfileRegistry) CreateCertificate(ctx context.Context, key *ecdsa.PrivateKey, owner common.Address, attributeType *big.Int, value []byte) (*types.Transaction, error) {
	opts := api.opts.getTxOpts(ctx, key, api.opts.gasLimit)
	return api.profileRegistryContract.CreateCertificate(opts, owner, attributeType, value)
}

func (api *ProfileRegistry) RemoveCertificate(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) error {
	opts := api.opts.getTxOpts(ctx, key, api.opts.gasLimit)
	tx, err := api.profileRegistryContract.RemoveCertificate(opts, id)
	if err != nil {
		return err
	}
	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, CertificateUpdatedTopic); err != nil {
		return err
	}

	return nil
}

func (api *ProfileRegistry) AddValidator(ctx context.Context, key *ecdsa.PrivateKey, validator common.Address, level int8) (*types.Transaction, error) {
	opts := api.opts.getTxOpts(ctx, key, api.opts.gasLimit)
	return api.profileRegistryContract.AddValidator(opts, validator, level)
}

func (api *ProfileRegistry) GetValidatorLevel(ctx context.Context, validatorID common.Address) (int8, error) {
	return api.profileRegistryContract.GetValidatorLevel(getCallOptions(ctx), validatorID)
}

func (api *ProfileRegistry) RemoveValidator(ctx context.Context, key *ecdsa.PrivateKey, validator common.Address) (*types.Transaction, error) {
	opts := api.opts.getTxOpts(ctx, key, api.opts.gasLimit)
	return api.profileRegistryContract.RemoveValidator(opts, validator)
}

func (api *ProfileRegistry) GetAttributeCount(ctx context.Context, owner common.Address, attributeType *big.Int) (*big.Int, error) {
	return api.profileRegistryContract.GetAttributeCount(getCallOptions(ctx), owner, attributeType)
}

func (api *ProfileRegistry) GetAttributeValue(ctx context.Context, owner common.Address, attributeType *big.Int) ([]byte, error) {
	return api.profileRegistryContract.GetAttributeValue(getCallOptions(ctx), owner, attributeType)
}

func (api *ProfileRegistry) GetProfileLevel(ctx context.Context, owner common.Address) (sonm.IdentityLevel, error) {
	lev, err := api.profileRegistryContract.GetProfileLevel(getCallOptions(ctx), owner)
	if err != nil {
		return sonm.IdentityLevel_UNKNOWN, fmt.Errorf("failed to get profile level: %s", err)
	}
	return sonm.IdentityLevel(lev), err
}

func (api *ProfileRegistry) GetValidator(ctx context.Context, validatorID common.Address) (*sonm.Validator, error) {
	level, err := api.profileRegistryContract.GetValidatorLevel(getCallOptions(ctx), validatorID)
	if err != nil {
		return nil, err
	}

	return &sonm.Validator{
		Id:    sonm.NewEthAddress(validatorID),
		Level: uint64(level),
	}, nil
}

func (api *ProfileRegistry) GetCertificate(ctx context.Context, certificateID *big.Int) (*sonm.Certificate, error) {
	validatorID, ownerID, attribute, value, err := api.profileRegistryContract.GetCertificate(getCallOptions(ctx), certificateID)
	if err != nil {
		return nil, err
	}

	return &sonm.Certificate{
		Id:          sonm.NewBigInt(certificateID),
		ValidatorID: sonm.NewEthAddress(validatorID),
		OwnerID:     sonm.NewEthAddress(ownerID),
		Attribute:   attribute.Uint64(),
		Value:       value,
	}, nil
}

type BasicBlacklistAPI struct {
	client            CustomEthereumClient
	blacklistContract *marketAPI.Blacklist
	opts              *chainOpts
}

func NewBasicBlacklist(address common.Address, opts *chainOpts) (BlacklistAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	blacklistContract, err := marketAPI.NewBlacklist(address, client)
	if err != nil {
		return nil, err
	}

	return &BasicBlacklistAPI{
		client:            client,
		blacklistContract: blacklistContract,
		opts:              opts,
	}, nil
}

func (api *BasicBlacklistAPI) Check(ctx context.Context, who, whom common.Address) (bool, error) {
	return api.blacklistContract.Check(getCallOptions(ctx), who, whom)
}

func (api *BasicBlacklistAPI) Add(ctx context.Context, key *ecdsa.PrivateKey, who, whom common.Address) (*types.Transaction, error) {
	opts := api.opts.getTxOpts(ctx, key, api.opts.gasLimit)
	return api.blacklistContract.Add(opts, who, whom)
}

func (api *BasicBlacklistAPI) Remove(ctx context.Context, key *ecdsa.PrivateKey, whom common.Address) error {
	opts := api.opts.getTxOpts(ctx, key, api.opts.gasLimit)
	tx, err := txRetryWrapper(func() (*types.Transaction, error) {
		return api.blacklistContract.Remove(opts, whom)
	})
	if err != nil {
		return err
	}

	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, RemovedFromBlacklistTopic); err != nil {
		return err
	}

	return nil
}

func (api *BasicBlacklistAPI) AddMaster(ctx context.Context, key *ecdsa.PrivateKey, root common.Address) (*types.Transaction, error) {
	opts := api.opts.getTxOpts(ctx, key, addMasterGasLimit)
	return api.blacklistContract.AddMaster(opts, root)
}

func (api *BasicBlacklistAPI) RemoveMaster(ctx context.Context, key *ecdsa.PrivateKey, root common.Address) (*types.Transaction, error) {
	opts := api.opts.getTxOpts(ctx, key, removeMasterGasLimit)
	return api.blacklistContract.RemoveMaster(opts, root)
}

func (api *BasicBlacklistAPI) SetMarketAddress(ctx context.Context, key *ecdsa.PrivateKey, market common.Address) (*types.Transaction, error) {
	opts := api.opts.getTxOpts(ctx, key, api.opts.gasLimit)
	return api.blacklistContract.SetMarketAddress(opts, market)
}

type StandardTokenApi struct {
	client        CustomEthereumClient
	tokenContract *marketAPI.StandardToken
	opts          *chainOpts
}

func NewStandardToken(address common.Address, opts *chainOpts) (TokenAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	tokenContract, err := marketAPI.NewStandardToken(address, client)
	if err != nil {
		return nil, err
	}

	return &StandardTokenApi{
		client:        client,
		tokenContract: tokenContract,
		opts:          opts,
	}, nil
}

func (api *StandardTokenApi) IncreaseApproval(ctx context.Context, key *ecdsa.PrivateKey, spender common.Address, value *big.Int) error {
	opts := api.opts.getTxOpts(ctx, key, increaseApprovalGasLimit)
	tx, err := api.tokenContract.IncreaseApproval(opts, spender, value)
	if err != nil {
		return err
	}

	receipt, err := WaitTransactionReceipt(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx)
	if err != nil {
		return err
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return fmt.Errorf("transaction failed: %s", err)
	}
	return nil
}

func (api *StandardTokenApi) DecreaseApproval(ctx context.Context, key *ecdsa.PrivateKey, spender common.Address, value *big.Int) error {
	opts := api.opts.getTxOpts(ctx, key, decreaseApprovalGasLimit)
	tx, err := api.tokenContract.DecreaseApproval(opts, spender, value)
	if err != nil {
		return err
	}

	receipt, err := WaitTransactionReceipt(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx)
	if err != nil {
		return err
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return fmt.Errorf("transaction failed: %s", err)
	}
	return nil
}

func (api *StandardTokenApi) BalanceOf(ctx context.Context, address common.Address) (*Balance, error) {
	block, err := api.client.GetLastBlock(ctx)
	if err != nil {
		return nil, err
	}

	ethBalance, err := api.client.GetEthereumBalanceAt(ctx, address, block)
	if err != nil {
		return nil, fmt.Errorf("failed to get Ethereum balance: %v", err)
	}

	snmBalance, err := api.tokenContract.BalanceOf(getCallOptions(ctx), address)
	if err != nil {
		return nil, fmt.Errorf("failed to get SNM balance: %v", err)
	}

	return &Balance{
		Eth: ethBalance,
		SNM: snmBalance,
	}, nil
}

func (api *StandardTokenApi) AllowanceOf(ctx context.Context, from, to common.Address) (*big.Int, error) {
	return api.tokenContract.Allowance(getCallOptions(ctx), from, to)
}

func (api *StandardTokenApi) ApproveAtLeast(ctx context.Context, key *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {
	curAmount, err := api.AllowanceOf(ctx, crypto.PubkeyToAddress(key.PublicKey), to)
	if err != nil {
		return err
	}
	if curAmount.Cmp(amount) >= 0 {
		return nil
	}
	return api.approve(ctx, key, to, amount, curAmount)
}

func (api *StandardTokenApi) Approve(ctx context.Context, key *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {
	curAmount, err := api.AllowanceOf(ctx, crypto.PubkeyToAddress(key.PublicKey), to)
	if err != nil {
		return err
	}
	return api.approve(ctx, key, to, amount, curAmount)
}

func (api *StandardTokenApi) approve(ctx context.Context, key *ecdsa.PrivateKey, to common.Address, amount *big.Int, curAmount *big.Int) error {
	opts := api.opts.getTxOpts(ctx, key, approveGasLimit)
	if curAmount.Cmp(big.NewInt(0)) != 0 {
		tx, err := api.tokenContract.Approve(opts, to, big.NewInt(0))
		if err != nil {
			return err
		}
		if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, ApprovalTopic); err != nil {
			return err
		}
	}

	// want to set a new zero value, but we just reset
	// allowance value to zero. Nothing to do there.
	if amount.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	// set new allowance value
	tx, err := api.tokenContract.Approve(opts, to, amount)
	if err != nil {
		return err
	}

	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, ApprovalTopic); err != nil {
		return err
	}

	return nil
}

func (api *StandardTokenApi) Transfer(ctx context.Context, key *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {
	opts := api.opts.getTxOpts(ctx, key, transferGasLimit)
	tx, err := txRetryWrapper(func() (*types.Transaction, error) {
		return api.tokenContract.Transfer(opts, to, amount)
	})
	if err != nil {
		return err
	}

	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, TransferTopic); err != nil {
		return err
	}

	return nil
}

func (api *StandardTokenApi) TransferFrom(ctx context.Context, key *ecdsa.PrivateKey, from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	opts := api.opts.getTxOpts(ctx, key, transferFromGasLimit)
	return api.tokenContract.TransferFrom(opts, from, to, amount)
}

func (api *StandardTokenApi) TotalSupply(ctx context.Context) (*big.Int, error) {
	return api.tokenContract.TotalSupply(getCallOptions(ctx))
}

type TestTokenApi struct {
	client        CustomEthereumClient
	tokenContract *marketAPI.SNMTToken
	opts          *chainOpts
}

func NewTestToken(address common.Address, opts *chainOpts) (TestTokenAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	tokenContract, err := marketAPI.NewSNMTToken(address, client)
	if err != nil {
		return nil, err
	}

	return &TestTokenApi{
		client:        client,
		tokenContract: tokenContract,
		opts:          opts,
	}, nil
}

func (api *TestTokenApi) GetTokens(ctx context.Context, key *ecdsa.PrivateKey) (*types.Transaction, error) {
	opts := api.opts.getTxOpts(ctx, key, getTokensGasLimit)
	return api.tokenContract.GetTokens(opts)
}

type BasicEventsAPI struct {
	parent  API
	options *chainOpts
	client  CustomEthereumClient
	logger  *zap.Logger
	// Number of blocks that will be read by FilterLogs() in one batch. This is required
	// so that we don't run out of memory trying to load all of the logs at once.
	blocksBatchSize uint64
}

type EventFilter struct {
	topics    [][]common.Hash
	addresses []common.Address
	fromBlock *big.Int
}

func NewEventsAPI(parent API, opts *chainOpts, logger *zap.Logger, blocksBatchSize uint64) (EventsAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	return &BasicEventsAPI{
		parent:          parent,
		options:         opts,
		client:          client,
		logger:          logger,
		blocksBatchSize: blocksBatchSize,
	}, nil
}

func (api *BasicEventsAPI) GetMarketFilter(fromBlock *big.Int) *EventFilter {
	var (
		topics     [][]common.Hash
		eventTopic = []common.Hash{
			DealOpenedTopic,
			DealUpdatedTopic,
			OrderPlacedTopic,
			OrderUpdatedTopic,
			DealChangeRequestSentTopic,
			DealChangeRequestUpdatedTopic,
			BilledTopic,
			WorkerAnnouncedTopic,
			WorkerConfirmedTopic,
			WorkerConfirmedTopic,
			WorkerRemovedTopic,
			AddedToBlacklistTopic,
			RemovedFromBlacklistTopic,
			ValidatorCreatedTopic,
			ValidatorDeletedTopic,
			CertificateCreatedTopic,
			NumBenchmarksUpdatedTopic,
		}
	)
	topics = append(topics, eventTopic)

	addresses := []common.Address{
		api.parent.ContractRegistry().MarketAddress(),
		api.parent.ContractRegistry().BlacklistAddress(),
		api.parent.ContractRegistry().ProfileRegistryAddress(),
	}

	return &EventFilter{
		fromBlock: fromBlock,
		topics:    topics,
		addresses: addresses,
	}
}

func (api *BasicEventsAPI) GetMultiSigFilter(addresses []common.Address, fromBlock *big.Int) *EventFilter {
	var (
		topics     [][]common.Hash
		eventTopic = []common.Hash{
			ConfirmationTopic,
			RevocationTopic,
			SubmissionTopic,
			ExecutionTopic,
			ExecutionFailureTopic,
			DepositTopic,
			OwnerAdditionTopic,
			OwnerRemovalTopic,
			RequirementChangeTopic,
		}
	)
	topics = append(topics, eventTopic)

	return &EventFilter{
		fromBlock: fromBlock,
		topics:    topics,
		addresses: addresses,
	}
}

func (api *BasicEventsAPI) GetLastBlock(ctx context.Context) (uint64, error) {
	block, err := api.client.GetLastBlock(ctx)
	if err != nil {
		return 0, err
	}
	if block.IsUint64() {
		return block.Uint64(), nil
	} else {
		return 0, errors.New("block number overflows uint64")
	}
}

func (api *BasicEventsAPI) GetEvents(ctx context.Context, filter *EventFilter) (chan *Event, error) {
	out := make(chan *Event, 1024)

	if filter.fromBlock != nil && !filter.fromBlock.IsUint64() {
		return nil, errors.New("filter fromBlock field is out of uint64 range")
	}
	// we are mutating filter, so prevent some strange behaviour on client side
	sFilter := simpleFilter{
		Addresses: filter.addresses,
		Topics:    filter.topics,
	}
	if filter.fromBlock != nil {
		sFilter.FromBlock = filter.fromBlock.Uint64()
	}

	go func() {
		if err := api.getEventsRoutine(ctx, sFilter, out); err != nil {
			out <- &Event{
				Data: &ErrorData{Err: err, Topic: "unknown"},
			}
		}
		close(out)
	}()
	return out, nil
}

type simpleFilter struct {
	FromBlock uint64
	ToBlock   uint64
	Addresses []common.Address
	Topics    [][]common.Hash
}

func (m *simpleFilter) EthFilter() ethereum.FilterQuery {
	return ethereum.FilterQuery{
		FromBlock: big.NewInt(0).SetUint64(m.FromBlock),
		ToBlock:   big.NewInt(0).SetUint64(m.ToBlock),
		Addresses: m.Addresses,
		Topics:    m.Topics,
	}
}

// Return block number which is beyond blockConfirmations count from the head
func (api *BasicEventsAPI) getLastConfirmedBlock(ctx context.Context) (uint64, error) {
	lastBlock, err := api.GetLastBlock(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block number: %v", err)
	}

	return lastBlock - uint64(api.options.blockConfirmations), nil

}

func (api *BasicEventsAPI) getEventsRoutine(ctx context.Context, filter simpleFilter, receiver chan<- *Event) error {
	tk := util.NewImmediateTicker(blockGenPeriod)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tk.C:
			// We do not want to read logs from block beyond blockConfirmations count from the head
			tillBlock, err := api.getLastConfirmedBlock(ctx)
			if err != nil {
				return fmt.Errorf("failed to get latest block number: %v", err)
			}
			if filter.FromBlock > tillBlock {
				continue
			}

			err = api.getEventsTill(ctx, filter, tillBlock, receiver)
			if err != nil {
				return err
			}
			filter.FromBlock = tillBlock + 1
		}
	}
}

func (api *BasicEventsAPI) getEventsTill(ctx context.Context, filter simpleFilter, tillBlock uint64, receiver chan<- *Event) error {
	ctxlog.S(ctx).Debugf("fetching events from %d till %d", filter.FromBlock, tillBlock)
	for {
		// we substract one, because of range inclusivity in FilterLogs call
		filter.ToBlock = filter.FromBlock + api.blocksBatchSize - 1
		if filter.ToBlock > tillBlock {
			filter.ToBlock = tillBlock
		}
		if err := api.fetchAndProcessLogs(ctx, filter, receiver); err != nil {
			return err
		}
		if filter.ToBlock == tillBlock {
			// we fetched all the logs
			return nil
		} else {
			// Start right after the last fetched block
			filter.FromBlock = filter.ToBlock + 1
		}
	}
}

func (api *BasicEventsAPI) fetchAndProcessLogs(ctx context.Context, filter simpleFilter, receiver chan<- *Event) error {
	logs, err := api.client.FilterLogs(ctx, filter.EthFilter())
	ctxlog.S(ctx).Debugf("filtering logs from %d to %d", filter.FromBlock, filter.ToBlock)
	if err != nil {
		return err
	}
	var curBlock uint64
	var curEventTS uint64
	for _, log := range logs {
		if log.BlockNumber != curBlock {
			curBlock = log.BlockNumber
			block, err := api.client.BlockByNumber(ctx, big.NewInt(0).SetUint64(curBlock))
			if err != nil {
				// TODO @aplodismerti: This place was changed, previously only log was written and old ts was used, is it right?
				return fmt.Errorf("failed to get event timestamp for block %d: %s", curBlock, err)
			}
			curEventTS = block.Time().Uint64()
			ctxlog.S(ctx).Debugf("switching to block %d", log.BlockNumber)
		}
		api.processLog(log, curEventTS, receiver)
	}
	ctxlog.S(ctx).Debugf("processed %d logs in blocks from %d to %d", len(logs), filter.FromBlock, filter.ToBlock)
	return nil
}

func (api *BasicEventsAPI) processLog(log types.Log, eventTS uint64, out chan<- *Event) {
	// This should never happen, but it's ethereum, and things might happen.
	if len(log.Topics) < 1 {
		out <- &Event{
			Data:        &ErrorData{Err: errors.New("malformed log entry"), Topic: "unknown"},
			BlockNumber: log.BlockNumber,
		}
		return
	}

	sendErr := func(out chan<- *Event, err error, topic common.Hash) {
		out <- &Event{Data: &ErrorData{Err: err, Topic: topic.String()}, BlockNumber: log.BlockNumber, TS: eventTS}
	}

	sendData := func(data interface{}) {
		out <- &Event{Data: data,
			BlockNumber:  log.BlockNumber,
			TxIndex:      uint64(log.TxIndex),
			ReceiptIndex: uint64(log.Index),
			TS:           eventTS,
		}
	}

	var topic = log.Topics[0]
	switch topic {
	case DealOpenedTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealOpenedData{ID: id})
	case DealUpdatedTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealUpdatedData{ID: id})
	case DealChangeRequestSentTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealChangeRequestSentData{ID: id})
	case DealChangeRequestUpdatedTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealChangeRequestUpdatedData{ID: id})
	case BilledTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		paidAmount, err := extractBig(log.Topics, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&BilledData{DealID: id, PaidAmount: paidAmount})
	case OrderPlacedTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&OrderPlacedData{ID: id})
	case OrderUpdatedTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&OrderUpdatedData{ID: id})
	case WorkerAnnouncedTopic:
		slaveID, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		masterID, err := extractAddress(log.Topics, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&WorkerAnnouncedData{WorkerID: slaveID, MasterID: masterID})
	case WorkerConfirmedTopic:
		slaveID, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		masterID, err := extractAddress(log.Topics, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&WorkerConfirmedData{WorkerID: slaveID, MasterID: masterID})
	case WorkerRemovedTopic:
		slaveID, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		masterID, err := extractAddress(log.Topics, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&WorkerRemovedData{WorkerID: slaveID, MasterID: masterID})
	case AddedToBlacklistTopic:
		adderID, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		addeeID, err := extractAddress(log.Topics, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&AddedToBlacklistData{AdderID: adderID, AddeeID: addeeID})
	case RemovedFromBlacklistTopic:
		removerID, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		removeeID, err := extractAddress(log.Topics, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&RemovedFromBlacklistData{RemoverID: removerID, RemoveeID: removeeID})
	case ValidatorCreatedTopic:
		id, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&ValidatorCreatedData{ID: id})
	case ValidatorDeletedTopic:
		id, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&ValidatorDeletedData{ID: id})
	case CertificateCreatedTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&CertificateCreatedData{ID: id})
	case CertificateUpdatedTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&CertificateUpdatedData{ID: id})
	case NumBenchmarksUpdatedTopic:
		numBenchmarksBig, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		if !numBenchmarksBig.IsUint64() {
			sendErr(out, errors.New("num benchmarks can not be used as uint64"), topic)
			return
		}
		sendData(&NumBenchmarksUpdatedData{NumBenchmarks: numBenchmarksBig.Uint64()})
	case ConfirmationTopic:
		sender, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		transactionID, err := extractBig(log.Topics, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&ConfirmationData{Sender: sender, TransactionId: transactionID})
	case RevocationTopic:
		sender, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		transactionID, err := extractBig(log.Topics, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&RevocationData{Sender: sender, TransactionId: transactionID})
	case SubmissionTopic:
		transactionID, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&SubmissionData{TransactionId: transactionID})
	case ExecutionTopic:
		transactionID, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&ExecutionData{TransactionId: transactionID})
	case ExecutionFailureTopic:
		transactionID, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&ExecutionFailureData{TransactionId: transactionID})
	case OwnerAdditionTopic:
		sender, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&OwnerAdditionData{Owner: sender})
	case OwnerRemovalTopic:
		sender, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&OwnerAdditionData{Owner: sender})
	default:
		out <- &Event{
			Data:        &ErrorData{Err: errors.New("unknown topic"), Topic: topic.String()},
			BlockNumber: log.BlockNumber,
		}
	}
}

type OracleUSDAPI struct {
	client         CustomEthereumClient
	oracleContract *marketAPI.OracleUSD
	opts           *chainOpts
}

func NewOracleUSDAPI(address common.Address, opts *chainOpts) (OracleAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	oracleContract, err := marketAPI.NewOracleUSD(address, client)
	if err != nil {
		return nil, err
	}

	return &OracleUSDAPI{
		client:         client,
		oracleContract: oracleContract,
		opts:           opts,
	}, nil

}

func (api *OracleUSDAPI) GetOwner(ctx context.Context) (common.Address, error) {
	return api.oracleContract.Owner(getCallOptions(ctx))
}

func (api *OracleUSDAPI) SetOwner(ctx context.Context, key *ecdsa.PrivateKey, owner common.Address) error {
	opts := api.opts.getTxOpts(ctx, key, api.opts.gasLimit)
	tx, err := api.oracleContract.TransferOwnership(opts, owner)
	if err != nil {
		return err
	}

	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, OwnershipTransferredTopic); err != nil {
		return err
	}

	return nil
}

func (api *OracleUSDAPI) SetCurrentPrice(ctx context.Context, key *ecdsa.PrivateKey, price *big.Int) (*types.Transaction, error) {
	opts := api.opts.getTxOpts(ctx, key, api.opts.gasLimit)
	return api.oracleContract.SetCurrentPrice(opts, price)
}

func (api *OracleUSDAPI) PackSetCurrentPriceTransactionData(price *big.Int) ([]byte, error) {
	oracleABI, err := abi.JSON(bytes.NewBufferString(marketAPI.OracleUSDABI))
	if err != nil {
		return nil, err
	}
	return oracleABI.Pack("setCurrentPrice", price)
}

func (api *OracleUSDAPI) UnpackSetCurrentPriceTransactionData(data []byte) (*big.Int, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}

	if len(data) < 36 {
		return nil, fmt.Errorf("data is malformed")
	}
	// Cut method specify byte, keep only parameters data
	// Method ids are saves at the first 4 bytes of the hash of the
	// methods string signature. (signature = baz(uint32,string32))
	params := data[4:]

	price := common.HexToAddress(common.Bytes2Hex(params[:32])).Big()
	return price, nil
}

func (api *OracleUSDAPI) GetCurrentPrice(ctx context.Context) (*big.Int, error) {
	return api.oracleContract.GetCurrentPrice(getCallOptions(ctx))
}

type BasicSimpleGatekeeper struct {
	client   CustomEthereumClient
	contract *marketAPI.SimpleGatekeeperWithLimit
	opts     *chainOpts
}

func NewSimpleGatekeeper(address common.Address, opts *chainOpts) (SimpleGatekeeperAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	contract, err := marketAPI.NewSimpleGatekeeperWithLimit(address, client)
	if err != nil {
		return nil, err
	}

	return &BasicSimpleGatekeeper{
		client:   client,
		contract: contract,
		opts:     opts,
	}, nil
}

func (api *BasicSimpleGatekeeper) PayIn(ctx context.Context, key *ecdsa.PrivateKey, value *big.Int) error {
	opts := api.opts.getTxOpts(ctx, key, payinGasLimit)
	tx, err := api.contract.Payin(opts, value)
	if err != nil {
		return err
	}

	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, PayinTopic); err != nil {
		return err
	}

	return nil
}

func (api *BasicSimpleGatekeeper) Payout(ctx context.Context, key *ecdsa.PrivateKey, to common.Address, value *big.Int, txNumber *big.Int) (*types.Transaction, error) {
	opts := api.opts.getTxOpts(ctx, key, payoutGasLimit)
	return api.contract.Payout(opts, to, value, txNumber)
}

func (api *BasicSimpleGatekeeper) Kill(ctx context.Context, key *ecdsa.PrivateKey) (*types.Transaction, error) {
	opts := api.opts.getTxOpts(ctx, key, api.opts.gasLimit)
	return api.contract.Kill(opts)
}

type BasicMultiSigAPI struct {
	client   CustomEthereumClient
	contract *marketAPI.MultiSigWallet
	opts     *chainOpts
	address  common.Address
}

func NewMultiSigAPI(address common.Address, opts *chainOpts) (MultiSigAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	contract, err := marketAPI.NewMultiSigWallet(address, client)
	if err != nil {
		return nil, err
	}

	return &BasicMultiSigAPI{
		address:  address,
		client:   client,
		contract: contract,
		opts:     opts,
	}, nil
}

func (api *BasicMultiSigAPI) AddOwner(ctx context.Context, key *ecdsa.PrivateKey, owner common.Address) error {
	oracleABI, err := abi.JSON(bytes.NewBufferString(marketAPI.MultiSigWalletABI))
	if err != nil {
		return err
	}
	data, err := oracleABI.Pack("AddOwner", owner)

	return api.SubmitTransaction(ctx, key, api.address, big.NewInt(0), data)
}

func (api *BasicMultiSigAPI) RemoveOwner(ctx context.Context, key *ecdsa.PrivateKey, owner common.Address) error {
	oracleABI, err := abi.JSON(bytes.NewBufferString(marketAPI.MultiSigWalletABI))
	if err != nil {
		return err
	}
	data, err := oracleABI.Pack("RemoveOwner", owner)

	return api.SubmitTransaction(ctx, key, api.address, big.NewInt(0), data)
}

func (api *BasicMultiSigAPI) ReplaceOwner(ctx context.Context, key *ecdsa.PrivateKey, owner common.Address, newOwner common.Address) error {
	oracleABI, err := abi.JSON(bytes.NewBufferString(marketAPI.MultiSigWalletABI))
	if err != nil {
		return err
	}
	data, err := oracleABI.Pack("ReplaceOwner", owner, newOwner)

	return api.SubmitTransaction(ctx, key, api.address, big.NewInt(0), data)
}

func (api *BasicMultiSigAPI) ChangeRequirement(ctx context.Context, key *ecdsa.PrivateKey, required *big.Int) error {
	oracleABI, err := abi.JSON(bytes.NewBufferString(marketAPI.MultiSigWalletABI))
	if err != nil {
		return err
	}
	data, err := oracleABI.Pack("ChangeRequirement", required)

	return api.SubmitTransaction(ctx, key, api.address, big.NewInt(0), data)
}

func (api *BasicMultiSigAPI) SubmitTransaction(ctx context.Context, key *ecdsa.PrivateKey, destination common.Address, value *big.Int, data []byte) error {
	opts := api.opts.getTxOpts(ctx, key, api.opts.gasLimit)
	tx, err := api.contract.SubmitTransaction(opts, destination, value, data)
	if err != nil {
		return err
	}

	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, ConfirmationTopic); err != nil {
		return err
	}

	return nil
}

func (api *BasicMultiSigAPI) ConfirmTransaction(ctx context.Context, key *ecdsa.PrivateKey, transactionID *big.Int) error {
	opts := api.opts.getTxOpts(ctx, key, api.opts.gasLimit)
	tx, err := api.contract.ConfirmTransaction(opts, transactionID)
	if err != nil {
		return err
	}

	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, ConfirmationTopic); err != nil {
		return err
	}

	return nil
}

func (api *BasicMultiSigAPI) RevokeConfirmation(ctx context.Context, key *ecdsa.PrivateKey, transactionID *big.Int) error {
	opts := api.opts.getTxOpts(ctx, key, api.opts.gasLimit)
	tx, err := api.contract.RevokeConfirmation(opts, transactionID)
	if err != nil {
		return err
	}

	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, RevocationTopic); err != nil {
		return err
	}

	return nil
}

func (api *BasicMultiSigAPI) ExecuteTransaction(ctx context.Context, key *ecdsa.PrivateKey, transactionID *big.Int) error {
	opts := api.opts.getTxOpts(ctx, key, api.opts.gasLimit)
	tx, err := api.contract.ExecuteTransaction(opts, transactionID)
	if err != nil {
		return err
	}

	if _, err := WaitTxAndExtractLog(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx, PayinTopic); err != nil {
		return err
	}

	return nil
}

func (api *BasicMultiSigAPI) IsConfirmed(ctx context.Context, transactionID *big.Int) (bool, error) {
	return api.contract.IsConfirmed(getCallOptions(ctx), transactionID)
}

func (api *BasicMultiSigAPI) GetConfirmationCount(ctx context.Context, transactionID *big.Int) (*big.Int, error) {
	return api.contract.GetConfirmationCount(getCallOptions(ctx), transactionID)
}

func (api *BasicMultiSigAPI) GetTransactionCount(ctx context.Context, penging bool, executed bool) (*big.Int, error) {
	return api.contract.GetTransactionCount(getCallOptions(ctx), penging, executed)
}

func (api *BasicMultiSigAPI) GetOwners(ctx context.Context) ([]common.Address, error) {
	return api.contract.GetOwners(getCallOptions(ctx))
}

func (api *BasicMultiSigAPI) GetConfirmations(ctx context.Context, transactionID *big.Int) ([]common.Address, error) {
	return api.contract.GetConfirmations(getCallOptions(ctx), transactionID)
}

func (api *BasicMultiSigAPI) GetTransactionIDs(ctx context.Context, from *big.Int, to *big.Int, pending bool, executed bool) ([]*big.Int, error) {
	return api.contract.GetTransactionIds(getCallOptions(ctx), from, to, pending, executed)
}

func (api *BasicMultiSigAPI) GetTransaction(ctx context.Context, transactionID *big.Int) (*MultiSigTransactionData, error) {
	tx, err := api.contract.Transactions(getCallOptions(ctx), transactionID)
	if err != nil {
		return nil, err
	}
	return &MultiSigTransactionData{
		To:       tx.Destination,
		Value:    tx.Value,
		Data:     tx.Data,
		Executed: tx.Executed,
	}, nil
}
