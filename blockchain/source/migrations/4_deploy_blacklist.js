let Blacklist = artifacts.require('./Blacklist.sol');

module.exports = function (deployer, network) {
    if ((network === 'private') || (network === 'privateLive')) {
        // deployer.deploy(Blacklist, { gasPrice: 0 });
    } else if (network === 'master') {
        // blacklist haven't reason to deploy to mainnet
    } else if (network === 'rinkeby') {
        // blacklist haven't reason to deploy to rinkeby
    } else {
        deployer.deploy(Blacklist);
    }
};
