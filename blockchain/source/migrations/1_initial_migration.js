var Migrations = artifacts.require('./Migrations.sol');

module.exports = function (deployer) {
    // Deploy the Migrations contract as our only task
    deployer.deploy(Migrations);
};
