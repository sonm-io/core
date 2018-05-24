var OracleUSD = artifacts.require('./OracleUSD.sol');

module.exports = function (deployer) {
    deployer.deploy(OracleUSD);
};
