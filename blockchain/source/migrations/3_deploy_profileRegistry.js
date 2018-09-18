let ProfileRegistry = artifacts.require('./ProfileRegistry.sol');

module.exports = function (deployer, network) {
    if ((network === 'private') || (network === 'privateLive')) {
        //deployer.deploy(ProfileRegistry, { gasPrice: 0 });
    } else if (network === 'master') {
        // we do not deploy ProfileRegistry at master anytime
    } else if (network === 'rinkeby') {
        // we do not deploy ProfileRegistry at rinkeby anytime
    } else {
        deployer.deploy(ProfileRegistry);
    }
};
