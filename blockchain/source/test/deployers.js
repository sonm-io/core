const DeployList = artifacts.require('./DeployList.sol');

contract('DeployList', async function (accounts) {
    let list;
    let owner = accounts[0];

    let deployer1 = accounts[2];
    let deployer2 = accounts[3];
    let deployer3 = accounts[4];

    let newDeployer = accounts[6];

    before(async function () {
        list = await DeployList.new([deployer1, deployer2, deployer3], { from: owner });
    });

    describe('Default', function () {
        it('should return constructor values', async function () {
            let addresses = await list.getDeployers();
            assert.equal(addresses[0], deployer1);
            assert.equal(addresses[1], deployer2);
            assert.equal(addresses[2], deployer3);
        });
    });

    describe('AddDeployer', function () {
        it('should add new element', async function () {
            await list.addDeployer(newDeployer);
            let addresses = await list.getDeployers();
            assert.notEqual(addresses.indexOf(newDeployer), -1);
        });
    });

    describe('RemoveDeployer', function () {
        it('should remove element', async function () {
            await list.removeDeployer(deployer1);
            let addresses = await list.getDeployers();
            assert.equal(addresses.indexOf(deployer1), -1);
        });
    });
});
