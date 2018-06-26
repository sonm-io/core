var GateMultisig = artifacts.require('./MultiSigWallet.sol');
var GateKeeperLive = artifacts.require('./SimpleGatekeeperWithLimitLive.sol');

var MSOwners = ['0x0', '0x0'];
var MSRequired = 1;
var SNMAddress = '0x983f6d60db79ea8ca4eb9968c6aff8cfa04b3c63';
var freezingTime = 30;
var actualGasPrice = 400;

module.exports = function (deployer, network) {
	deployer.then(async () => {
	    if (network === 'master') {
	        await deployer.deploy(GateMultisig, MSOwners, MSRequired, { gasPrice: actualGasPrice});
	        await deployer.deploy(GateKeeperLive, SNMAddress, freezingTime, { gasPrice : actualGasPrice});
	        let multisig = await GateMultisig.deployed();
	        await GateKeeperLive.deployed().tranferOwnership(multisig.address);
	    }
	})
};
