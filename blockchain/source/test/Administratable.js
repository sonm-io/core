import assertRevert from "./helpers/assertRevert";

const Dummy = artifacts.require('./Dummy.sol');
const Administratable = artifacts.require('./Administratable.sol');

class Assertions {

    constructor(contract, somebody, owner, administrator) {
        this.contract = contract;
        this.somebody = somebody;
        this.owner = owner;
        this.administrator = administrator;
    }

    async assertTransferOwnershipFromAdmin(to) {
        await ass
    }

    async assertTransferOwnershipFromOwner(to) {

    }

    async assertTransferOwnership(from, to) {
        let tx = await
        this.contract.transferOwnership(to, {from: from});
        assert.equal(tx.logs.length, 1);
        assert.equal(tx.logs[0].event, 'ownershipTransferred');
        assert.equal(tx.logs[0].args.newOwner, to);
    }

    async assertTransferAdministratorship(via, to) {
        let tx = await
        this.contract.transferAdministratorship(to, {from: via});
        assert.equal(tx.logs.length, 1);
        assert.equal(tx.logs[0].event, 'administratorshipTransferred');
        assert.equal(tx.logs[0].args.newAdministrator, to);
    }

    async assertOnlyOwner(owner) {
        assertRevert(contract.testOnlyOwner({from: this.somebody}));
        let tx = await
        this.contract.testOnlyOwner({from: owner});
        assert.equal(tx.logs.length, 1);
        assert.equal(tx.logs[0].event, 'Call');
        assert.equal(tx.logs[0].args.method.toString(), "onlyOwner");
    }

    async assertOwnerOrAdministrator(ownerOrAdministrator) {
        assertRevert(contract.testOwnerOrAdministartor({from: this.somebody}));
        let tx = await
        this.contract.testOwnerOrAdministrator({from: ownerOrAdministrator});
        assert.equal(tx.logs.length, 1);
        assert.equal(tx.logs[0].event, 'Call');
        assert.equal(tx.logs[0].args.method.toString(), "ownerOrAdministrator");
    }
}

contract('Administratable', function (accounts) {
    let dummy;
    let assertions;
    let firstOwner = accounts[0];
    let secondOwner = accounts[1];
    let firstAdministrator = accounts[2];
    let secondAdministrator = accounts[3];

    before(async function () {
        dummy = await Dummy.new({ from: firstOwner });
        await dummy.transferAdministratorship(firstAdministrator);
        assertions = Assertions.new(dummy, accounts[9]);
    });

    describe('OwnerOrAdministratorModifier', function() {
        it('should not permit to execute ownerOrAdministrator functions to anybody except owner or administrator', async function() {
            await assertOwnerOrAdministrator(dummy, firstOwner, somebody);
            await assertOwnerOrAdministrator(dummy, firstAdministrator, somebody);
        });
        it('should correctly transfer administratorship', async function (){
            // check case when admin and owner is one address
            await assertTransferAdministratorship(dummy, firstAdministrator, firstOwner);
            await assertOwnerOrAdministrator(dummy, firstOwner, firstAdministrator);

            // transfer via owner
            await assertTransferAdministratorship(dummy, firstOwner, secondAdministrator);
            await assertTransferAdministratorship(dummy, firstOwner, firstAdministrator);

            // transfer via
            await assertTransferAdministratorship(dummy, secondAdministrator, firstAdministrator);
        });
    });

    describe('OnlyOwner modifier', function() {
        it('should not permit to execute onlyOwner functions to anybody except owner', async function() {
            assertOnlyOwner(dummy, firstOwner, somebody);
        });
        it('should correctly transfer ownership from owner', async function(){
            await assertTransferOwnership(dummy, firstOwner, secondOwner);
            assertRevert(dummy.transferOwnership(secondOwner, {from: firstOwner}));
            await assertOnlyOwner(dummy, secondOwner, firstOwner);
            await assertTransferOwnership(dummy, secondOwner, firstOwner);
            await assertOnlyOwner(dummy, firstOwner, secondOwner);
        });
        it('should correctly transfer ownership via administrator', async function(){
    });

});
