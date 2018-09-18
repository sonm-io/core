let Orders = artifacts.require('./Orders.sol');
let Deals = artifacts.require('./Deals.sol');
let ChangeRequests = artifacts.require('./ChangeRequests.sol');

module.exports = function (deployer, network) {
  deployer.then(async () => {
    if ((network === 'private') || (network === 'privateLive')) {
        //await deployer.deploy(Blacklist, { gasPrice: 0 });
    } else if (network === 'master') {
        // will filled later
    } else if (network === 'rinkeby') {
        // later
    } else {
        await deployer.deploy(Orders, {gasLimit: 20000000});
        await deployer.deploy(Deals, {gasLimit: 20000000});
        await deployer.deploy(ChangeRequests, {gasLimit: 20000000})
    }
  });
};
