var SNM = artifacts.require('./SNM.sol');
var Market = artifacts.require('./Market.sol');
var Blacklist = artifacts.require('./Blacklist.sol');
var OracleUSD = artifacts.require('./OracleUSD.sol');
var ProfileRegistry = artifacts.require('./ProfileRegistry.sol');

module.exports = function (deployer, network) {
    if (network === 'private') {
        deployer.deploy(Market,
            SNM.address, // token address
            Blacklist.address, // Blacklist address
            '0x1f995e52dcbec7c0d00d45b8b1bf43b29dd12b5b', // Oracle address
            ProfileRegistry.address, // ProfileRegistry address
            12, // benchmarks quantity
            3, // netflags quantity
            { gasPrice: 0 });
    } else {
        deployer.deploy(Market, SNM.address, Blacklist.address, OracleUSD.address, ProfileRegistry.address, 12, 3);
    }
};
