const DevicesStorage = artifacts.require('./DevicesStorage.sol');
const AddressHashMap = artifacts.require('./AddressHashMap.sol');
const Multisig = artifacts.require('./MultiSigWallet.sol');
const TruffleConfig = require('../truffle');
const ContractRegistry = require('../migration_utils/address_hashmap');

async function deploySidechain (deployer, network, accounts) {
    let registry = new ContractRegistry(AddressHashMap, network, Multisig);
    console.log('registry created');
    await registry.init();
    console.log('registry initialized');
    await deployer.deploy(DevicesStorage);
    console.log('devices storage deployed');
    let ds = await DevicesStorage.deployed();
    console.log('deployed object fetched');
    await registry.write('devicesStorageAddress', ds.address);
    console.log('written to registry');
    let ms = await registry.resolve(Multisig, 'migrationMultSigAddress');
    console.log('ms resolved');
    await ds.transferOwnership(ms.address);
    console.log('ownership transferred');

    await registry.write('migrationMultiSigAddress', ms.address);
}

module.exports = function (deployer, network, accounts) {
    deployer.then(async () => { // eslint-disable-line promise/catch-or-return
        if (TruffleConfig.isSidechain(network)) {
            await deploySidechain(deployer, network, accounts);
        }
    });
};
