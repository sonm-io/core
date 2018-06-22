let SimpleGatekeeper = artifacts.require('./SimpleGatekeeper.sol');
let SNM = artifacts.require('./SNM.sol');
var TestnetFaucet = artifacts.require('./TestnetFaucet.sol');

module.exports = function (deployer, network) {
    if (network === 'private') {
        deployer.deploy(SimpleGatekeeper, SNM.address, { gasPrice: 0 });
    } else if (network === 'rinkeby') {
        TestnetFaucet.deployed()
            .then(function (instance) {
                instance.getTokenAddress()
                    .then(function (result) {
                        deployer.deploy(SimpleGatekeeper, SNM.address, { gasPrice: 0 });
                    })
                    .catch(function (err) {
                        console.log(err);
                    });
            })
            .catch(function (err) {
                console.log(err);
            });
    } else {
        deployer.deploy(SimpleGatekeeper, SNM.address);
    }
};
