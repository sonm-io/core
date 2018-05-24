let SimpleGatekeeper = artifacts.require('./SimpleGatekeeper.sol');
let SNM = artifacts.require('./SNM.sol');

module.exports = function (deployer) {
    deployer.deploy(SimpleGatekeeper, SNM.address);
};
