let Orders = artifacts.require('./Orders.sol');
let Deals = artifacts.require('./Deals.sol');
let ChangeRequests = artifacts.require('./ChangeRequests.sol');

module.exports = function (deployer, network) {
    if ((network === 'private') || (network === 'privateLive')) {
        // await deployer.deploy(Blacklist, { gasPrice: 0 });
    } else if (network === 'master') {
        // will filled later
    } else if (network === 'rinkeby') {
        // later
    } else {
        deployer.deploy(Orders, { gasLimit: 20000000 });
        deployer.deploy(Deals, { gasLimit: 20000000 });
        deployer.deploy(ChangeRequests, { gasLimit: 20000000 });
    }
};
