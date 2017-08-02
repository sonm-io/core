var SonmDummyToken = artifacts.require("./SonmDummyToken.sol");
var Factory = artifacts.require("./Factory.sol");

module.exports = function(deployer) {

    deployer.deploy(SonmDummyToken).then(function () {
        return deployer.deploy(Factory, SonmDummyToken.address, web3.eth.accounts[0]);
    })

};
