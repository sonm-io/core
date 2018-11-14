import assertRevert from './helpers/assertRevert';

const Dummy = artifacts.require('./test/Dummy.sol');

class Assertions {
    constructor (contract, somebody, owner, administrator) {
        this.contract = contract;
        this.somebody = somebody;
        this.owner = owner;
        this.administrator = administrator;
    }

    async assertTransferOwnershipViaAdmin (to) {
        await this.assertTransferOwnership(this.administrator, to);
    }

    async assertTransferOwnershipViaOwner (to) {
        await this.assertTransferOwnership(this.owner, to);
    }

    async assertTransferOwnership (via, to) {
        assert.notEqual(to, this.somebody);
        assert.notEqual(to, this.owner);
        await assertRevert(this.contract.testOnlyOwner({ from: to }));
        let tx = await this.contract.transferOwnership(to, { from: via });
        assert.equal(tx.logs.length, 1);
        assert.equal(tx.logs[0].event, 'OwnershipTransferred');
        assert.equal(tx.logs[0].args.newOwner, to);
        if (this.owner !== this.administrator) {
            await assertRevert(this.contract.testOnlyOwner({ from: this.owner }));
        } else {
            await assertRevert(this.contract.testOnlyOwner({ from: this.somebody }));
        }
        this.owner = to;
    }

    async assertTransferAdministratorship (to) {
        assert.notEqual(to, this.administrator);
        await assertRevert(this.contract.transferAdministratorship(to, { from: to }));
        let tx = await this.contract.transferAdministratorship(to, { from: this.administrator });
        assert.equal(tx.logs.length, 1);
        assert.equal(tx.logs[0].event, 'AdministratorshipTransferred');
        assert.equal(tx.logs[0].args.newAdministrator, to);
        await assertRevert(this.contract.transferAdministratorship(to, { from: this.administrator }));
        this.administrator = to;
    }

    async assertOnlyOwner () {
        await assertRevert(this.contract.testOnlyOwner({ from: this.somebody }));
        let tx = await this.contract.testOnlyOwner({ from: this.owner });
        assert.equal(tx.logs.length, 1);
        assert.equal(tx.logs[0].event, 'Call');
        assert.equal(tx.logs[0].args.method.toString(), 'onlyOwner');
    }

    async assertOwnerOrAdministratorViaOwner () {
        await this.assertOwnerOrAdministrator(this.owner);
    }

    async assertOwnerOrAdministratorViaAdministrator () {
        await this.assertOwnerOrAdministrator(this.administrator);
    }

    async assertOwnerOrAdministrator (ownerOrAdministrator) {
        await assertRevert(this.contract.testOwnerOrAdministrator({ from: this.somebody }));
        let tx = await this.contract.testOwnerOrAdministrator({ from: ownerOrAdministrator });
        assert.equal(tx.logs.length, 1);
        assert.equal(tx.logs[0].event, 'Call');
        assert.equal(tx.logs[0].args.method.toString(), 'ownerOrAdministrator');
    }
}

async function makeAssertions (accounts) {
    let dummy = await Dummy.new({ from: accounts[0] });
    await dummy.transferAdministratorship(accounts[2], { from: accounts[0] });
    return new Assertions(dummy, accounts[9], accounts[0], accounts[2]);
}

contract('Administratable', function (accounts) {
    let firstOwner = accounts[0];
    let secondOwner = accounts[1];
    let firstAdministrator = accounts[2];
    let secondAdministrator = accounts[3];
    let assertions;
    before(async () => {
        assertions = await makeAssertions(accounts);
    });

    describe('OwnerOrAdministratorModifier', () => {
        it('should not permit to execute ownerOrAdministrator functions to anybody except owner or administrator',
            async () => {
                await assertions.assertOwnerOrAdministratorViaAdministrator();
                await assertions.assertOwnerOrAdministratorViaOwner();
            }
        );
        it('should correctly transfer administratorship', async () => {
            // check case when admin and owner is one address
            await assertions.assertTransferAdministratorship(firstOwner);
            await assertions.assertOwnerOrAdministratorViaAdministrator();

            // transfer via administrator
            await assertions.assertTransferAdministratorship(secondAdministrator);
            await assertions.assertTransferAdministratorship(firstAdministrator);
            await assertions.assertOwnerOrAdministratorViaAdministrator();
        });
    });

    describe('OnlyOwner modifier', () => {
        it('should not permit to execute onlyOwner functions to anybody except owner', async () => {
            await assertions.assertOnlyOwner();
        });
        it('should correctly transfer ownership via owner', async () => {
            await assertions.assertTransferOwnershipViaOwner(secondOwner);
            await assertions.assertTransferOwnershipViaOwner(firstOwner);
        });
        it('should correctly transfer ownership via administrator', async () => {
            await assertions.assertTransferOwnershipViaAdmin(secondOwner);
            await assertions.assertTransferOwnershipViaAdmin(firstOwner);
        });
    });
});
