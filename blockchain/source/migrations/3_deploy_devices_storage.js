const DevicesStorage = artifacts.require('./DevicesStorage.sol');
const AddressHashMap = artifacts.require('./AddressHashMap.sol');
const Multisig = artifacts.require('./MultiSigWallet.sol');
const TruffleConfig = require('../truffle');
const ContractRegistry = require('../migration_utils/address_hashmap');

async function deploySidechain(deployer, network, accounts) {
    let registry = new ContractRegistry(AddressHashMap, network, Multisig);
    await registry.init();
    await deployer.deploy(DevicesStorage);
    let ds = await DevicesStorage.deployed();
    registry.write("devicesStorageAddress", ds.address);
    let ms = await registry.resolve(Multisig, 'multiSigAddress');
    await ds.transferOwnership(ms.address);
}

module.exports = function (deployer, network, accounts) {
    deployer.then(async () => { // eslint-disable-line promise/catch-or-return
        if (TruffleConfig.isSidechain(network)) {
            await deploySidechain(deployer, network, accounts);
        }
    });
};
