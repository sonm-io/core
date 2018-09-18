let Market = artifacts.require('./Market.sol');
let Blacklist = artifacts.require('./Blacklist.sol');

module.exports = function (deployer, network) {
    if (network === 'private') {
        // Blacklist.deployed()
        //     .then(function (blacklist) {
        //         blacklist.SetMarketAddress(Market.address, { gasPrice: 0 });
        //     })
        //     .catch(function (err) {
        //         console.log(err);
        //     });
    } else if (network === 'master') {
        // blacklist haven't reason to deploy to mainnet
    } else if (network === 'rinkeby') {
        // blacklist haven't reason to deploy to rinkeby
    } else {
        Blacklist.deployed()
            .then(function (blacklist) {
                blacklist.SetMarketAddress(Market.address);
            })
            .catch(function (err) {
                console.log(err);
            });
    }
};
