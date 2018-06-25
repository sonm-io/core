let SNM = artifacts.require('./SNM.sol');
let TestnetFaucet = artifacts.require('./TestnetFaucet.sol');

module.exports = function (deployer, network) {
    if ((network === 'private') || (network === 'privateLive')) {
        deployer.deploy(SNM, { gasPrice: 0 });
    } else if (network === 'rinkeby') {
        deployer.deploy(TestnetFaucet);
    } else if (network === 'master') {
        // token already deployed at master to address 0x983f6d60db79ea8ca4eb9968c6aff8cfa04b3c63
    } else {
        deployer.deploy(SNM);
    }
};
