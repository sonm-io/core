var SNM = artifacts.require('./SNM.sol');
var Gatekeeper = artifacts.require('./Gatekeeper.sol');

module.exports = function (deployer) {
    deployer.deploy(Gatekeeper, SNM.address);
};
