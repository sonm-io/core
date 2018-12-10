let Multisig = artifacts.require('./MultiSigWallet.sol');
let OracleMultisig = artifacts.require('./MultiSigWallet.sol');
let SNM = artifacts.require('./SNM.sol');

let Market = artifacts.require('./Market.sol');
let Blacklist = artifacts.require('./Blacklist.sol');
let ProfileRegistry = artifacts.require('./ProfileRegistry');
let Oracle = artifacts.require('./OracleUSD.sol');
let GateKeeper = artifacts.require('./SimpleGatekeeperWithLimit.sol');
let GateKeeperLive = artifacts.require('./SimpleGatekeeperWithLimitLive.sol');
let AddressHashMap = artifacts.require('./AddressHashMap.sol');
let TestnetFaucet = artifacts.require('./TestnetFaucet.sol');

let { isSidechain, isMainChain, oppositeNetName } = require('../migration_utils/network');

const TruffleConfig = require('../truffle');

const MSWrapper = require('../migration_utils/multisig');

let freezingTime = 60 * 15;
let actualGasPrice = 3000000000;
let benchmarksQuantity = 12;
let netflagsQuantity = 3;

function MSRequired (network) {
    if (network === 'rinkeby' || network === 'private' || network === 'dev_main' || network === 'dev_side') {
        return 1;
    } else {
        return 5;
    }
}

function MSOwners (network, accounts) {
    if (network === 'dev_main' || network === 'dev_side') {
        return [
            accounts[0]
        ]
    } else {
        return [
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
    }
}

function determineGatekeeperMasterchainAddress (network) {
    let targetNet = TruffleConfig.networks[network].main_network_id;
    if (!GateKeeperLive.hasNetwork(targetNet)) {
        throw new Error('GateKeeper was not deployed to ' + targetNet);
    }
    return GateKeeperLive.networks[targetNet].address;
}

function needFaucet (network) {
    return network === 'dev_main' || network === 'rinkeby';
}

function hasFaucetInMain (network) {
    return network === 'dev_side' || network === 'private';
}

async function determineSNMMasterchainAddress (network) {
    if (network === 'privateLive') {
        // In main net it is already deployed
        return '0x983f6d60db79ea8ca4eb9968c6aff8cfa04b3c63';
    }
    try {
        let faucet = await TestnetFaucet.deployed();
        return faucet.getTokenAddress();
    } catch {
        //pass
    }
    let faucetAddress = determineFaucetAddress(network);
    let alt = TestnetFaucet.clone();
    let mainNet = TruffleConfig.networks[network].main_network_id;
    alt.setNetwork(mainNet);
    alt.setProvider(TruffleConfig.networks[oppositeNetName(network)].provider());
    let faucet = alt.at(faucetAddress);

    return faucet.getTokenAddress();
}

function determineFaucetAddress (network) {
    let targetNet = TruffleConfig.networks[network].main_network_id;
    if (!TestnetFaucet.hasNetwork(targetNet)) {
        throw new Error('TestnetFaucet was not deployed to ' + targetNet);
    }
    return TestnetFaucet.networks[targetNet].address;
}

async function deployMainchain (deployer, network, accounts) {
    if (needFaucet(network)) {
        await deployer.deploy(TestnetFaucet);
    }
    // deploy Live Gatekeeper
    let snmAddr = await determineSNMMasterchainAddress(network);
    await deployer.deploy(GateKeeperLive, snmAddr, freezingTime, { gasPrice: actualGasPrice });
    let gk = await GateKeeperLive.deployed();

    // add keeper with 100k limit for testing
    await gk.ChangeKeeperLimit('0xAfA5a3b6675024af5C6D56959eF366d6b1FBa0d4', 100000 * 1e18, { gasPrice: actualGasPrice }); // eslint-disable-line max-len

    await deployer.deploy(Multisig, MSOwners(network, accounts), MSRequired(network), { gasPrice: actualGasPrice });
    let multisig = await Multisig.deployed();

    // transfer Live Gatekeeper ownership to `Gatekeeper` multisig
    await gk.transferOwnership(multisig.address, { gasPrice: actualGasPrice });
}

async function deploySidechain (deployer, network, accounts) {
    let GatekeeperMasterchainAddress = determineGatekeeperMasterchainAddress(network);
    if (GatekeeperMasterchainAddress === '') {
        console.log('GatekeeperMasterchainAddress is not set!!!');
        throw new Error('GatekeeperMasterchainAddress is not set!!!');
    }
    // 1) deploy SNM token
    await deployer.deploy(SNM, { gasPrice: 0 });
    let token = await SNM.deployed();

    // 2) deploy Gatekeper
    await deployer.deploy(GateKeeper, token.address, freezingTime, { gasPrice: 0 });
    let gk = await GateKeeper.deployed();

    // 3) transfer all tokens to Gatekeeper
    await token.transfer(gk.address, 444 * 1e6 * 1e18, { gasPrice: 0 });

    // 3.1): add keeper with 100k limit for testing
    await gk.ChangeKeeperLimit('0x1f0dc2f125a2df9e37f32242cc3e34328f096b3c', 100000 * 1e18, { gasPrice: 0 });

    // deploy ProfileRegistry
    await deployer.deploy(ProfileRegistry, { gasPrice: 0 });
    let pr = await ProfileRegistry.deployed();

    // deploy Blacklist
    await deployer.deploy(Blacklist, { gasPrice: 0 });
    let bl = await Blacklist.deployed();

    // deploy Oracle
    await deployer.deploy(Oracle, { gasPrice: 0 });
    let oracle = await Oracle.deployed();

    // set price in Oracle
    await oracle.setCurrentPrice('6244497036986155008', { gasPrice: 0 });

    // deploy Market
    await deployer.deploy(Market, SNM.address, Blacklist.address, Oracle.address, ProfileRegistry.address, benchmarksQuantity, netflagsQuantity, { gasPrice: 0 }); // eslint-disable-line max-len
    let market = await Market.deployed();

    // set Market address to Blacklist
    await bl.SetMarketAddress(market.address, { gasPrice: 0 });

    // deploy AddressHashMap
    await deployer.deploy(AddressHashMap, { gasPrice: 0 });
    let ahm = await AddressHashMap.deployed();

    // write
    await ahm.write('sidechainSNMAddress', SNM.address, { gasPrice: 0 });
    let snmAddr = await determineSNMMasterchainAddress(network);
    if (hasFaucetInMain(network)) {
        await ahm.write('testnetFaucetAddress', determineFaucetAddress(network), { gasPrice: 0 });
    }
    await ahm.write('masterchainSNMAddress', snmAddr, { gasPrice: 0 });
    await ahm.write('blacklistAddress', bl.address, { gasPrice: 0 });
    await ahm.write('marketAddress', market.address, { gasPrice: 0 });
    await ahm.write('profileRegistryAddress', pr.address, { gasPrice: 0 });
    await ahm.write('oracleUsdAddress', oracle.address, { gasPrice: 0 });
    await ahm.write('gatekeeperSidechainAddress', gk.address, { gasPrice: 0 });
    await ahm.write('gatekeeperMasterchainAddress', GatekeeperMasterchainAddress, { gasPrice: 0 });


     // 0) deploy `Gatekeeper` multisig
     await deployer.deploy(Multisig, MSOwners(network, accounts), MSRequired(network), { gasPrice: 0 });
     let multiSig = await Multisig.deployed();
     await ahm.write('multiSigAddress', multiSig.address, { gasPrice: 0 });
     // compatibility
     await ahm.write('oracleMultiSigAddress', multiSig.address, { gasPrice: 0 });

     // transfer AddressHashMap ownership to `Migration` multisig
     await ahm.transferOwnership(multiSig.address, { gasPrice: 0 });

     // 4) transfer Gatekeeper ownership to `Gatekeeper` multisig
     await gk.transferOwnership(multiSig.address, { gasPrice: 0 });

     // transfer ProfileRegistry ownership to `Migration` multisig
     await pr.transferOwnership(multiSig.address, { gasPrice: 0 });

     await deployer.deploy(OracleMultisig, MSOwners(network, accounts), MSRequired(network), { gasPrice: 0 });
     let oracleMS = await OracleMultisig.deployed();
     // Transfer Oracle ownership to `Oracle` multisig
     oracle.transferOwnership(oracleMS.address, { gasPrice: 0 });

     // transfer Market ownership to `Migration`
     await market.transferOwnership(multiSig.address, { gasPrice: 0 });

     // transfer Blacklist ownership to Migration multisig
     await bl.transferOwnership(multiSig.address, { gasPrice: 0 });

}

module.exports = function (deployer, network, accounts) {
    deployer.then(async () => { // eslint-disable-line promise/catch-or-return
        if (isSidechain(network)) {
            await deploySidechain(deployer, network, accounts);
        }
        if (isMainChain(network)) {
            await deployMainchain(deployer, network, accounts);
        }
    });
};