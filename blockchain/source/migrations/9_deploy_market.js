let SNM = artifacts.require('./SNM.sol');
let Market = artifacts.require('./Market.sol');
let Blacklist = artifacts.require('./Blacklist.sol');
let OracleUSD = artifacts.require('./OracleUSD.sol');
let ProfileRegistry = artifacts.require('./ProfileRegistry.sol');
let Administratum = artifacts.require('./Administratum.sol');
let Orders = artifacts.require('./Orders.sol');
let Deals = artifacts.require('./Deals.sol');
let ChangeRequests = artifacts.require('./ChangeRequests.sol');

module.exports = function (deployer, network) {
    deployer.then(async () => {
      if ((network === 'private') || (network === 'privateLive')) {
          // await deployer.deploy(Market,
          //       SNM.address, // token address
          //       Blacklist.address, // Blacklist address
          //       OracleUSD.address, // Oracle address
          //       ProfileRegistry.address, // ProfileRegistry address
          //       Administratum.address,
          //       Orders.address,
          //       Deals.address,
          //       ChangeRequests.address,
          //       12, // benchmarks quantity
          //       3, // netflags quantity
          //       { gasPrice: 0 });
          await deployer.deploy(Market, '0x0', '0x0', '0x0', '0x0', '0x0', '0x0', '0x0', '0x0', 12, 3, {gasLimit: 99000000, gasPrice: 0});
      } else if (network === 'master') {
          // market haven't reason to deploy to mainnet
      } else if (network === 'rinkeby') {
          // market haven't reason to deploy to rinkeby
          //await deployer.deploy(Market, '0x0', '0x0', '0x0', '0x0', '0x0', '0x0', '0x0', '0x0', 12, 3, {gasLimit: 100000000, gasPrice: 1000000};
      } else {
        await deployer.deploy(Market, SNM.address, Blacklist.address, OracleUSD.address, ProfileRegistry.address, Administratum.address, Orders.address, Deals.address, ChangeRequests.address, 12, 3, {gasLimit: 1000000000});
        //await deployer.deploy(Market, '0x0', '0x0', '0x0', '0x0', '0x0', '0x0', '0x0', '0x0', 12, 3, {gasLimit: 15000000, gasPrice: 1})
      }
   });
};
