var SimpleGatekeeper = artifacts.require('./SimpleGatekeeper.sol');
var SNM = artifacts.require('./SNM.sol');
var SNMD = artifacts.require('./SNMD.sol');

module.exports = function (deployer, network) {
    if (network === 'private') {
        SNMD.deployed()
            .then(function (res) {
                res.transfer(SimpleGatekeeper.address, 444 * 1e6 * 1e18, { gasPrice: 0 });
            })
            .catch(function (err) {
                console.log(err);
            });
    } else if (network === 'rinkeby') {
        // we do not transfer token to simple gatekeeper at rinkeby anytime
    } else {
        SNM.deployed()
            .then(function (res) {
                res.transfer(SimpleGatekeeper.address, 444 * 1e6 * 1e18);
            })
            .catch(function (err) {
                console.log(err);
            });
    }
};
