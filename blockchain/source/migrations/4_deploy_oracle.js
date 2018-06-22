var OracleUSD = artifacts.require('./OracleUSD.sol');

module.exports = function (deployer, network) {
    if (network === 'private') {
        deployer.deploy(OracleUSD, { gasPrice: 0 });
    } else if (network === 'rinkeby') {
        // oracle haven't reason to deploy to rinkeby
    } else {
        deployer.deploy(OracleUSD);
    }
};
