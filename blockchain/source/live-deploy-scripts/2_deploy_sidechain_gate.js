var sidechainGateMS = artifacts.require('./MultiSigWallet.sol');
var GateKeeper = artifacts.require('./SimpleGatekeeperWithLimit.sol');
var SNM = artifacts.require('./SNM.sol');

var MSOwners = ['0x34', '0x35'];
var MSRequired = 1;
var freezingTime = 30;

module.exports = function (deployer, network) {
	deployer.then(async () => {
	    if (network ==='privateLive') {
	    	await deployer.deploy(SNM, { gasPrice: 90000000 });
	    	let token = await SNM.deployed();
	        await deployer.deploy(sidechainGateMS, MSOwners, MSRequired, { gasPrice: 90000000000 });
	        let multiSig = await sidechainGateMS.deployed();
	        await token.transfer(multiSig.address, 444 * 1e6 * 1e18, { gasPrice: 9000000000 });
	        await deployer.deploy(GateKeeper, SNM.address, freezingTime, { gasPrice : 90000000000 });
	        let Gate = await GateKeeper.deployed();
	        await Gate.transferOwnership(multiSig.address);
	    }
	})
};
