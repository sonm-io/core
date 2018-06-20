var ProfileRegistry = artifacts.require('./ProfileRegistry.sol');

module.exports = function (deployer, network) {
    if (network === 'private') { // no reason to deploy to rinkeby
        deployer.deploy(ProfileRegistry, { gasPrice: 0 });
    } else if (network === 'rinkeby') {
        // we do not deploy ProfileRegistry at rinkeby anytime
    } else {
        deployer.deploy(ProfileRegistry);
    }
};
