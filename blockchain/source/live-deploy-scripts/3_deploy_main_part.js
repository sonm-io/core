var multiSigMigrations = artifacts.require('./MultiSigWallet.sol');
var multiSigOracle = artifacts.require('./MultiSigWallet.sol');
var Market = artifacts.require('./Market.sol');
var SNM = artifacts.require('./SNM.sol');
var Blacklist = artifacts.require('./Blacklist.sol');
var ProfileRegistry = artifacts.require('./ProfileRegistry');
var Oracle = artifacts.require('./OracleUSD.sol');
var DeployList = artifacts.require('./DeployList.sol');
var AddressHashMap = artifacts.require('./AddressHashMap.sol');

// filled before deploy 

var MSOwners = ['0x34', '0x35'];
var MSRequired = 1;

var benchmarksQuantity = 13;
var netflagsQuantity = 3;

var Deployers = ['0x488']; 

// main part

module.exports = function (deployer, network) {
	deployer.then(async () => {
	    if (network === 'privateLive') {
	        await deployer.deploy(multiSigMigrations, MSOwners, MSRequired, { gasPrice: 0 });
	        let multiSigMig = await multiSigMigrations.deployed();

	        await deployer.deploy(ProfileRegistry, { gasPrice: 0 });
	        let pr = await ProfileRegistry.deployed();
	        await pr.transferOwnership(multiSigMig.address, { gasPrice: 0});
	    	await deployer.deploy(Blacklist, { gasPrice: 0});
	        let bl = await Blacklist.deployed();

	        await deployer.deploy(multiSigOracle, MSOwners, MSRequired, { gasPrice: 0 });
	    	let multiSigOrac = await multiSigOracle.deployed();
	        await deployer.deploy(Oracle, { gasPrice: 0 });
	        let oracle = await Oracle.deployed()
	       	oracle.transferOwnership(multiSigOrac.address);
	        
	        await deployer.deploy(Market, SNM.address, Blacklist.address, Oracle.address, ProfileRegistry.address, benchmarksQuantity, netflagsQuantity, { gasPrice: 3000000000 });
	        let market = await Market.deployed();
	        await market.transferOwnership(multiSigMigrations.address);
	        await bl.SetMarketAddress(market.address);
	  		await bl.transferOwnership(multiSigMig.address);

	  		await deployer.deploy(DeployList, Deployers, { gasPrice: 0 });
	  		await deployer.deploy(AddressHashMap, { gasPrice: 0 });
	  		let dl = await DeployList.deployed();
	  		let ahm = await AddressHashMap.deployed();
	  		await dl.transferOwnership(multiSigMig.address);
	  		await ahm.transferOwnership(multiSigMig.address);
	    }
	})
};