const Migrations = artifacts.require('./Migrations.sol');
const TruffleConfig = require('../truffle');

module.exports = function (deployer, network) {
    // Deploy the Migrations contract as our only task
    if (TruffleConfig.isSidechain(network)) {
        deployer.deploy(Migrations, { gasPrice: 0 });
    } else {
        deployer.deploy(Migrations, { gas: 500000 });
    }
};
