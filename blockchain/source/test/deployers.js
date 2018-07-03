import assertRevert from './helpers/assertRevert';
const DeployList = artifacts.require('./DeployList.sol');

contract('DeployList', async (accounts) => {
    let list;
    let owner = accounts[0];

    let deployer1 = accounts[2];
    let deployer2 = accounts[3];
    let deployer3 = accounts[4];

    let newDeployer = accounts[6];
    let creeper = accounts[7];

    before(async () => {
        list = await DeployList.new([deployer1, deployer2, deployer3], { from: owner });
    });

    describe('Default', () => {
        it('should return constructor values', async () => {
            let addresses = await list.getDeployers();
            assert.deepEqual(addresses, [deployer1, deployer2, deployer3]);
        });
    });

    describe('Add/Remove deployers', () => {
        it('should add new deployer', async () => {
            await list.addDeployer(newDeployer);
            let addresses = await list.getDeployers();
            assert.equal(addresses.length, 4);
            assert.notEqual(addresses.indexOf(newDeployer), -1);
        });

        it('should remove deployer', async () => {
            await list.removeDeployer(deployer1);
            let addresses = await list.getDeployers();
            assert.equal(addresses.length, 3);
            assert.equal(addresses.indexOf(deployer1), -1);
        });

        it('creeper cannot add deployer', async () => {
            await assertRevert(list.addDeployer(newDeployer, { from: creeper }));
        });

        it('creeper cannot remove deployer', async () => {
            await assertRevert(list.addDeployer(newDeployer, { from: creeper }));
        });
    });
});
