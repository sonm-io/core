let SimpleGatekeeper = artifacts.require('./SimpleGatekeeper.sol');
let SNM = artifacts.require('./SNM.sol');
let TestnetFaucet = artifacts.require('./TestnetFaucet.sol');

module.exports = function (deployer, network) {
    if ((network === 'private') || (network === 'privateLive')) {
        deployer.deploy(SimpleGatekeeper, SNM.address, { gasPrice: 0 });
    } else if (network === 'master') {
        deployer.deploy(SimpleGatekeeper, '0x983f6d60db79ea8ca4eb9968c6aff8cfa04b3c63');
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
