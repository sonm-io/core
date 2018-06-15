var SNMD = artifacts.require('./SNMD.sol');
var Market = artifacts.require('./Market.sol');

module.exports = function (deployer, network) {
    if (network === 'private') {
        SNMD.at('0x26524b1234e361eb4e3ddf7600d41271620fcb0a')
            .then(function (token) {
                token.AddMarket(Market.address, { gasPrice: 0 });
            })
            .catch(function (err) {
                console.log(err);
            });
    }
};
