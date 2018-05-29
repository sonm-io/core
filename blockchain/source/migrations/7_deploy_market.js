var SNM = artifacts.require('./SNM.sol');
var Market = artifacts.require('./Market.sol');
var Blacklist = artifacts.require('./Blacklist.sol');
var OracleUSD = artifacts.require('./OracleUSD.sol');
var ProfileRegistry = artifacts.require('./ProfileRegistry.sol');

module.exports = function (deployer, network) {
    if (network === 'private') {
        deployer.deploy(Market, '0x5db024c6469634f4b307135cb76e8e6806f007b3', Blacklist.address, OracleUSD.address, '0xcc1cb65bdea124520dbdcc9e82b21b4352e45ad9', 12, { gasPrice: 0 });
    } else {
        deployer.deploy(Market, SNM.address, Blacklist.address, OracleUSD.address, ProfileRegistry.address, 12);
    }
};
