package blockchain

import "github.com/ethereum/go-ethereum/common"

// Market topics
var (
	DealOpenedTopic               = common.HexToHash("0xb9ffc65567b7238dd641372277b8c93ed03df73945932dd84fd3cbb33f3eddbf")
	DealUpdatedTopic              = common.HexToHash("0x0b27183934cfdbeb1fbbe288c2e163ed7aa8f458a954054970f78446bccb36e0")
	OrderPlacedTopic              = common.HexToHash("0xffa896d8919f0556f53ace1395617969a3b53ab5271a085e28ac0c4a3724e63d")
	OrderUpdatedTopic             = common.HexToHash("0xb8b459bc0688c37baf5f735d17f1711684bc14ab7db116f88bc18bf409b9309a")
	DealChangeRequestSentTopic    = common.HexToHash("0x7ff56b2eb3ce318aad93d0ba39a3e4a406992a136f9554f17f6bcc43509275d1")
	DealChangeRequestUpdatedTopic = common.HexToHash("0x4b92d35447745e95b7344414a41ae94984787d0ebcd2c12021169197bb59af39")
	BilledTopic                   = common.HexToHash("0x51f87cd83a2ce6c4ff7957861f7aba400dc3857d2325e0c94cc69f468874515c")
	WorkerAnnouncedTopic          = common.HexToHash("0xe398d33bf7e881cdfc9f34c743822904d4e45a0be0db740dd88cb132e4ce2ed9")
	WorkerConfirmedTopic          = common.HexToHash("0x4940ef08d5aed63b7d3d3db293d69d6ed1d624995b90e9e944839c8ea0ae450d")
	WorkerRemovedTopic            = common.HexToHash("0x7822736ed69a5fe0ad6dc2c6669e8053495d711118e5435b047f9b83deda4c37")
	NumBenchmarksUpdatedTopic     = common.HexToHash("0x1acf16d0a0451282e1d2cac3f5473ca7c931bcda610ff6e061041af50e2abc13")
)

// Gatekeeper topics
var (
	PayinTopic   = common.HexToHash("0x14312725abbc46ad798bc078b2663e1fcbace97be0247cd177176f3b4df2538e")
	PayoutTopic  = common.HexToHash("0x731af16374848c2c73a6154fd410cb421138e7db45c5a904e5a475c756faa8d9")
	CommitTopic  = common.HexToHash("0x65546c3bc3a77ffc91667da85018004299542e28a511328cfb4b3f86974902ee")
	SuicideTopic = common.HexToHash("0xa1ea9b09ea114021983e9ecf71cf2ffddfd80f5cb4f925e5bf24f9bdb5e55fde")
)

// Blacklist topics
var (
	AddedToBlacklistTopic     = common.HexToHash("0x708802ac7da0a63d9f6b2df693b53345ad263e42d74c245110e1ec1e03a1567e")
	RemovedFromBlacklistTopic = common.HexToHash("0x576a9aef294e1b4baf3617fde4cbc80ba5344d5eb508222f29e558981704a457")
)

// ProfileRegistry topics
var (
	ValidatorCreatedTopic   = common.HexToHash("0x02db26aafd16e8ecd93c4fa202917d50b1693f30b1594e57f7a432ede944eefc")
	ValidatorDeletedTopic   = common.HexToHash("0xa7a579573d398d7b67cd7450121bb250bbd060b29eabafdebc3ce0918658635c")
	CertificateCreatedTopic = common.HexToHash("0xb9bb1df26fde5c1295a7ccd167330e5d6cb9df14fe4c3884669a64433cc9e760")
	CertificateUpdatedTopic = common.HexToHash("0x9a100d2018161ede6ca34c8007992b09bbffc636a636014a922e4c8750412628")
)

// StandardToken topics
var (
	TransferTopic = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	ApprovalTopic = common.HexToHash("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925")
)

// Ownable topics
var (
	OwnershipTransferredTopic = common.HexToHash("0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0")
)

// SNMT topics
var (
	GiveAwayTopic = common.HexToHash("0xe08e9d066634006283658128ec91f58d444719d7a07d49f72924da4352ff94ad")
)

// MultiSig topics
var (
	ConfirmationTopic      = common.HexToHash("0x4a504a94899432a9846e1aa406dceb1bcfd538bb839071d49d1e5e23f5be30ef")
	RevocationTopic        = common.HexToHash("0xf6a317157440607f36269043eb55f1287a5a19ba2216afeab88cd46cbcfb88e9")
	SubmissionTopic        = common.HexToHash("0xc0ba8fe4b176c1714197d43b9cc6bcf797a4a7461c5fe8d0ef6e184ae7601e51")
	ExecutionTopic         = common.HexToHash("0x33e13ecb54c3076d8e8bb8c2881800a4d972b792045ffae98fdf46df365fed75")
	ExecutionFailureTopic  = common.HexToHash("0x526441bb6c1aba3c9a4a6ca1d6545da9c2333c8c48343ef398eb858d72b79236")
	DepositTopic           = common.HexToHash("0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c")
	OwnerAdditionTopic     = common.HexToHash("0xf39e6e1eb0edcf53c221607b54b00cd28f3196fed0a24994dc308b8f611b682d")
	OwnerRemovalTopic      = common.HexToHash("0x8001553a916ef2f495d26a907cc54d96ed840d7bda71e73194bf5a9df7a76b90")
	RequirementChangeTopic = common.HexToHash("0xa3f1ee9126a074d9326c682f561767f710e927faa811f7a99829d49dc421797a")
)
