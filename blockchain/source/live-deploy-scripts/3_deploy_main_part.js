let MultiSigWallet = artifacts.require('./MultiSigWallet.sol');
let multiSigOracle = artifacts.require('./MultiSigWallet.sol');
let Market = artifacts.require('./Market.sol');
let SNM = artifacts.require('./SNM.sol');
let Blacklist = artifacts.require('./Blacklist.sol');
let ProfileRegistry = artifacts.require('./ProfileRegistry');
let Oracle = artifacts.require('./OracleUSD.sol');
let GateKeeper = artifacts.require('./SimpleGatekeeperWithLimit.sol');
let DeployList = artifacts.require('./DeployList.sol');
let AddressHashMap = artifacts.require('./AddressHashMap.sol');

// filled before deploy
let SNMMasterchainAddress = '0x983f6d60db79ea8ca4eb9968c6aff8cfa04b3c63';
let GatekeeperMasterchainAddress = '';

let MSOwners = [
    '0xdaec8F2cDf27aD3DF5438E5244aE206c5FcF7fCd',
    '0xd9a43e16e78c86cf7b525c305f8e72723e0fab5e',
    '0x72cb2a9AD34aa126fC02b7d32413725A1B478888',
    '0x1f50Be5cbFBFBF3aBD889e17cb77D31dA2Bd7227',
    '0xe062C67207F7E478a93EF9BEA39535d8EfFAE3cE',
    '0x5fa359a9137cc5ac2a85d701ce2991cab5dcd538',
    '0x7aa5237e0f999a9853a9cc8c56093220142ce48e',
    '0xd43f262536e916a4a807d27080092f190e25d774',
    '0xdd8422eed7fe5f85ea8058d273d3f5c17ef41d1c',
];

let MSRequired = 5;
let benchmarksQuantity = 12;
let netflagsQuantity = 3;

let Deployers = ['0x7aa5237e0f999a9853a9cc8c56093220142ce48e', '0xd9a43e16e78c86cf7b525c305f8e72723e0fab5e'];

// main part

module.exports = function (deployer, network) {
    deployer.then(async () => { // eslint-disable-line promise/catch-or-return
        if (network === 'privateLive') {
            if (GatekeeperMasterchainAddress === '') {
                console.log('GatekeeperMasterchainAddress is not set!!!');
                return;
            }

            // deploy `Migration` MultiSig
            await deployer.deploy(MultiSigWallet, MSOwners, MSRequired, { gasPrice: 0 });
            let multiSigMig = await MultiSigWallet.deployed();

            // deploy `Oracle` MultiSig
            await deployer.deploy(multiSigOracle, MSOwners, MSRequired, { gasPrice: 0 });
            let msOracle = await multiSigOracle.deployed();

            // deploy ProfileRegistry
            await deployer.deploy(ProfileRegistry, { gasPrice: 0 });
            let pr = await ProfileRegistry.deployed();

            // transfer ProfileRegistry ownership to `Migration` multisig
            await pr.transferOwnership(multiSigMig.address, { gasPrice: 0 });

            // deploy Blacklist
            await deployer.deploy(Blacklist, { gasPrice: 0 });
            let bl = await Blacklist.deployed();

            // deploy Oracle
            await deployer.deploy(Oracle, { gasPrice: 0 });
            let oracle = await Oracle.deployed();

            // Transfer Oracle ownership to `Oracle` multisig
            oracle.transferOwnership(msOracle.address, { gasPrice: 0 });

            // deploy Market
            await deployer.deploy(Market, SNM.address, Blacklist.address, Oracle.address, ProfileRegistry.address, benchmarksQuantity, netflagsQuantity, { gasPrice: 0 }); // eslint-disable-line max-len
            let market = await Market.deployed();

            // transfer Market ownership to `Migration`
            await market.transferOwnership(multiSigMig.address, { gasPrice: 0 });

            // set Market address to Blacklist
            await bl.SetMarketAddress(market.address, { gasPrice: 0 });

            // transfer Blacklist ownership to Migration multisig
            await bl.transferOwnership(multiSigMig.address, { gasPrice: 0 });

            // deploy Deploylist
            await deployer.deploy(DeployList, Deployers, { gasPrice: 0 });
            let deployList = await DeployList.deployed();

            // transfer DeployList ownership
            await deployList.transferOwnership(multiSigMig.address, { gasPrice: 0 });

            // deploy AddressHashMap
            await deployer.deploy(AddressHashMap, { gasPrice: 0 });
            let ahm = await AddressHashMap.deployed();

            let gk = await GateKeeper.deployed();

            // write
            await ahm.write('sidechainSNMAddress', SNM.address, { gasPrice: 0 });
            await ahm.write('masterchainSNMAddress', SNMMasterchainAddress, { gasPrice: 0 });
            await ahm.write('blacklistAddress', bl.address, { gasPrice: 0 });
            await ahm.write('marketAddress', market.address, { gasPrice: 0 });
            await ahm.write('profileRegistryAddress', pr.address, { gasPrice: 0 });
            await ahm.write('oracleUsdAddress', oracle.address, { gasPrice: 0 });
            await ahm.write('gatekeeperSidechainAddress', gk.address, { gasPrice: 0 });
            await ahm.write('gatekeeperMasterchainAddress', GatekeeperMasterchainAddress, { gasPrice: 0 });

            // transfer AddressHashMap ownership to `Migration` multisig
            await ahm.transferOwnership(multiSigMig.address, { gasPrice: 0 });
        }
    });
};
