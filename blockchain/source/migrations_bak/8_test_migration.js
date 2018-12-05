let SNM = artifacts.require('./SNM.sol');

module.exports = function (deployer, network) {
    deployer.deploy(SNM);
};
