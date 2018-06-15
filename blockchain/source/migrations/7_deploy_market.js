var SNM = artifacts.require('./SNM.sol');
var Market = artifacts.require('./Market.sol');
var Blacklist = artifacts.require('./Blacklist.sol');
var OracleUSD = artifacts.require('./OracleUSD.sol');
var ProfileRegistry = artifacts.require('./ProfileRegistry.sol');

module.exports = function (deployer, network) {
    if (network === 'private') {
        deployer.deploy(Market,
            '0x5db024c6469634f4b307135cb76e8e6806f007b3', // token address
            '0x9ad1e969ec5842ee5d67414536813e224ceb56b1', // Blacklist address
            '0x1f995e52dcbec7c0d00d45b8b1bf43b29dd12b5b', // Oracle address
            '0xacfe1a688649fe9798b44a76b906fdca6e584a8d', // ProfileRegistry address
            12, // benchmarks quantity
            3,
            { gasPrice: 0 });
    } else {
        deployer.deploy(Market, SNM.address, Blacklist.address, OracleUSD.address, ProfileRegistry.address, 12, 3);
    }
};
