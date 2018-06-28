let sidechainGateMS = artifacts.require('./MultiSigWallet.sol');
let Gatekeeper = artifacts.require('./SimpleGatekeeperWithLimit.sol');
let SNM = artifacts.require('./SNM.sol');

let MSOwners = [
    '0xdaec8F2cDf27aD3DF5438E5244aE206c5FcF7fCd',
    '0xd9a43e16e78c86cf7b525c305f8e72723e0fab5e',
    '0x72cb2a9AD34aa126fC02b7d32413725A1B478888',
    '0x1f50Be5cbFBFBF3aBD889e17cb77D31dA2Bd7227',
    '0xe062C67207F7E478a93EF9BEA39535d8EfFAE3cE',
    '0x5fa359a9137cc5ac2a85d701ce2991cab5dcd538',
    '0x7aa5237e0f999a9853a9cc8c56093220142ce48e',
    '0xd43f262536e916a4a807d27080092f190e25d774',
    '0xdd8422eed7fe5f85ea8058d273d3f5c17ef41d1c',
];

let MSRequired = 5;
let freezingTime = 60 * 15;

module.exports = function (deployer, network) {
    deployer.then(async () => { // eslint-disable-line promise/catch-or-return
        if (network === 'privateLive') {
            // 0) deploy `Gatekeeper` multisig
            await deployer.deploy(sidechainGateMS, MSOwners, MSRequired, { gasPrice: 0 });
            let multiSig = await sidechainGateMS.deployed();

            // 1) deploy SNM token
            await deployer.deploy(SNM, { gasPrice: 0 });
            let token = await SNM.deployed();

            // 2) deploy Gatekeper
            await deployer.deploy(Gatekeeper, token.address, freezingTime, { gasPrice: 0 });
            let gk = await Gatekeeper.deployed();

            // 3) transfer all tokens to Gatekeeper
            await token.transfer(multiSig.address, 444 * 1e6 * 1e18, { gasPrice: 0 });

            // 3.1): add keeper with 10k limit for testing
            await gk.ChangeKeeperLimit('0xAfA5a3b6675024af5C6D56959eF366d6b1FBa0d4', 10000 * 1e18);

            // 4) transfer Gatekeeper ownership to `Gatekeeper` multisig
            await gk.transferOwnership(multiSig.address, { gasPrice: 0 });
        }
    });
};
