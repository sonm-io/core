var ProfileRegistry = artifacts.require("./ProfileRegistry.sol");

module.exports = function (deployer, network) {
    if (network === "private") {
        ProfileRegistry.deployed()
            .then(function (pr) {
                pr.AddValidator('0x074dbe6017d6cb0bfb594046f694da8b7dc31266', 1);
                pr.AddValidator('0xf9c176c276dc8c04ee9f01166f70fd238e5a16cf', 2);
                pr.AddValidator('0xbeeeff0a0f4dd2dbacfbf4ff4d4838962f761cc4', 3);
                pr.AddValidator('0xac4b829daa17c686ac5264b70c9f4d9ce54a2ec9', 4);
            })
    }
};
