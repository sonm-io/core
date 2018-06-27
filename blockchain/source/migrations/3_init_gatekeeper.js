let SimpleGatekeeper = artifacts.require('./SimpleGatekeeper.sol');
let SNM = artifacts.require('./SNM.sol');

module.exports = function (deployer, network) {
    if ((network === 'private') || (network === 'privateLive')) {
        SNM.deployed()
            .then(function (res) {
                res.transfer(SimpleGatekeeper.address, 444 * 1e6 * 1e18, { gasPrice: 0 });
            })
            .catch(function (err) {
                console.log(err);
            });
    } else if (network === 'master') {
        // we do not transfer token to gatekeeper at mainnet anytime
    } else if (network === 'rinkeby') {
        // we do not transfer token to gatekeeper at rinkeby anytime
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
