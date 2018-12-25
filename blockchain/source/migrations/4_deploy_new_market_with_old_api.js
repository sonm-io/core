const AddressHashMap = artifacts.require('./AddressHashMap.sol');
const Multisig = artifacts.require('./MultiSigWallet.sol');
const SNM = artifacts.require('./SNM.sol');
const Blacklist = artifacts.require('./Blacklist.sol');
const ProfileRegistry = artifacts.require('./ProfileRegistry');
const Oracle = artifacts.require('./OracleUSD.sol');

let Orders = artifacts.require('./Orders.sol');
let Deals = artifacts.require('./Deals.sol');
let AdministratumCrud = artifacts.require('./AdministratumCrud.sol');
let Administratum = artifacts.require('Administratum.sol');
let ChangeRequests = artifacts.require('./ChangeRequests.sol');
let Market = artifacts.require('./Market.sol');

const TruffleConfig = require('../truffle');
const ContractRegistry = require('../migration_utils/address_hashmap');

const benchmarksNum = 13;
const netflagsNum = 3;

async function deploySidechain (deployer, network, accounts) {
    let registry = new ContractRegistry(AddressHashMap, network, Multisig);
    console.log('registry created');
    await registry.init();
    console.log('registry initialized');

    let pr = await registry.resolve(ProfileRegistry, 'profileRegistryAddress');
    console.log('profile registry initialized');

    let bl = await registry.resolve(Blacklist, 'blacklistAddress');
    console.log('blacklist initialized')

    let snm = await registry.resolve(SNM, 'sidechainSNMAddress');
    console.log('snm initialized');

    let oracle = await registry.resolve(Oracle, 'oracleAddress');
    console.log('oracle initialized');

    // deploy crud
    await deployer.deploy(Orders, { gasPrice:  0 });
    let orders = await Orders.deployed();
    console.log('orders crud deployed');

    await deployer.deploy(Deals, { gasPrice: 0 });
    console.log('deals crud deployed');
    let deals = await Deals.deployed();
    console.log('deployed object fetched');

    await deployer.deploy(AdministratumCrud, { gasPrice: 0 });
    console.log('administratum crud deployed');
    let ac = await AdministratumCrud.deployed();
    console.log('deployed object fetched');

    await deployer.deploy(Administratum, ac.address, { gasPrice: 0 });
    console.log('administratum deployed');
    let adm = await Administratum.deployed();
    console.log('deployed object fetched');

    await deployer.deploy(ChangeRequests, { gasPrice: 0 });
    console.log('change requests crud deployed');
    let cr = await ChangeRequests.deployed();
    console.log('deployed object fetched');

    // deploy market
    await deployer.deploy(Market, snm.address, bl.address, oracle.address, pr.address, adm.address, orders.address, deals.address, cr.address, benchmarksNum, netflagsNum )
    console.log('market deployed');
    let market = await Market.deployed();
    console.log('deployed object fetched');

    // link market to administratum
    await adm.SetMarketAddress(market.address, { gasPrice:0 });
    let ms = await registry.resolve(Multisig, 'migrationMultSigAddress');
    console.log('ms resolved');

    // link administratum crud to its contract
    await ac.tranferOwnership(adm.address, { gasPrice: 0 });

    //transfer crud ownerships
    await orders.tranferOwnership(market.address, { gasPrice: 0 });
    await deals.tranferOwnership(market.address, { gasPrice: 0 });
    await cr.tranferOwnership(market.address, { gasPrice: 0 });
    await orders.transferAdministratorship(market.address, { gasPrice: 0 });
    await deals.transferAdministratorship(market.address, { gasPrice: 0 });
    await cr.transferAdministratorship(market.address, { gasPrice: 0 });

    // transfer main contracts ownership
    await adm.transferOwnership(ms.address, { gasPrice: 0 });
    await market.transferOwnership(ms.address, { gasPrice: 0 });
    console.log('ownerships transferred');


    console.log('and submitted all theese addresses to registry');

    await registry.write('migrationMultiSigAddress', ms.address);
    await registry.write('administratumAddress', adm.address);
    await registry.write('ordersCrudAddress', orders.address);
    await registry.write('dealsCrudAddress', deals.address);
    await registry.write('administratumCrudAddress', ac.address);
    await registry.write('administratumAddress', adm.address);
    await registry.write('marketAddress', market.address);
}

module.exports = function (deployer, network, accounts) {
    deployer.then(async () => { // eslint-disable-line promise/catch-or-return
        if (TruffleConfig.isSidechain(network)) {
            await deploySidechain(deployer, network, accounts);
        }
    });
};
