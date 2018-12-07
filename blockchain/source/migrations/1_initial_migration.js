var Migrations = artifacts.require('./Migrations.sol');
var util = require('../migration_utils/network');

module.exports = function (deployer, network) {
    // Deploy the Migrations contract as our only task
    if (util.isSidechain (network)) {
        deployer.deploy(Migrations, { gasPrice: 0 });
    } else {
        deployer.deploy(Migrations);
    }

};
