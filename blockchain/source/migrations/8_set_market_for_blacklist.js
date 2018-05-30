var Market = artifacts.require("./Market.sol");
var Blacklist = artifacts.require("./Blacklist.sol");

module.exports = function (deployer, network) {
    Blacklist.deployed()
        .then(function (blacklist) {
            if (network === "private") {
                blacklist.SetMarketAddress(Market.address, {gasPrice: 0});
            }else{
                blacklist.SetMarketAddress(Market.address);
            }
    })

};
