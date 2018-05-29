var Blacklist = artifacts.require('./Blacklist.sol');

module.exports = function (deployer, network) {
    if (network === 'private') {
        deployer.deploy(Blacklist, { gasPrice: 0 });
    } else {
        deployer.deploy(Blacklist);
    }
};
