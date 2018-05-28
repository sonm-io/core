var ProfileRegistry = artifacts.require('./ProfileRegistry.sol');

module.exports = function (deployer, network) {
    if (network !== 'rinkeby') { // no reason to deploy to rinkeby
        deployer.deploy(ProfileRegistry);
    }
};
