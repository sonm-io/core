let SimpleGatekeeper = artifacts.require('./SimpleGatekeeper.sol');
let SNM = artifacts.require('./SNM.sol');

module.exports = function (deployer, network) {
    if (network === 'private') {
        deployer.deploy(SimpleGatekeeper, SNM.address, { gasPrice: 0 });
    } else if (network === 'rinkeby') {
        deployer.deploy(SimpleGatekeeper, SNM.address);
    } else {
        deployer.deploy(SimpleGatekeeper, SNM.address);
    }
};
