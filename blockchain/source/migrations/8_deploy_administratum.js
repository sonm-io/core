let AdministratumCrud = artifacts.require('./AdministratumCrud.sol');
let Administratum = artifacts.require('./Administratum.sol');

module.exports = function (deployer, network) {
    if ((network === 'private') || (network === 'privateLive')) {
        deployer.deploy(Administratum, { gasPrice: 0 });
    } else if (network === 'master') {
        // will filled later
    } else if (network === 'rinkeby') {
        // later
    } else {
        deployer.deploy(Administratum, AdministratumCrud.address);
    }
};
