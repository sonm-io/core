var ProfileRegistry = artifacts.require("./ProfileRegistry.sol");

module.exports = function (deployer, network) {
    if (network !== "private") {
        deployer.deploy(ProfileRegistry);
    }
};
