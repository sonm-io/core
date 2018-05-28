let SimpleGatekeeper = artifacts.require('./SimpleGatekeeper.sol');
let SNM = artifacts.require('./SNM.sol');
let SNMD = artifacts.require('./SNMD.sol');

module.exports = function (deployer, network) {
    if (network === 'private') {
        deployer.deploy(SimpleGatekeeper, SNMD.address, { gasPrice: 0 });
    } else if (network === 'rinkeby') {
        deployer.deploy(SimpleGatekeeper, SNMD.address);
    } else {
        deployer.deploy(SimpleGatekeeper, SNM.address);
    }
};
