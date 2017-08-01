var SonmDummyToken = artifacts.require("./SonmDummyToken.sol");

module.exports = function(deployer) {

    deployer.deploy(SonmDummyToken).then(function () {
    })

};
