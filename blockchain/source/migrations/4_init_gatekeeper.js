var SimpleGatekeeper = artifacts.require('./SimpleGatekeeper.sol');
var SNM = artifacts.require('./SNM.sol');

module.exports = function (deployer, network) {
    if (network === 'private') {
        SNM.deployed()
            .then(function (res) {
                res.transfer(SimpleGatekeeper.address, 444 * 1e6 * 1e18, { gasPrice: 0 });
            })
            .catch(function (err) {
                console.log(err);
            });
    } else if (network === 'rinkeby') {
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
