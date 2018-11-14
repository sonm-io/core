import assertRevert from './helpers/assertRevert';

const FixedPoint128 = artifacts.require('./FixedPoint128.sol');
const RatingData = artifacts.require('./RatingData.sol');
const Rating = artifacts.require('./Rating.sol');

let roleWorker = 0;
let roleMaster = 1;
let roleConsumer = 2;
let Roles = ['worker', 'master', 'consumer'];

let Unclaimed = 0;
let ClaimAccepted = 1;
let ClaimRejected = 2;

let UnclaimedRating = 0;
let UnproblematicRating = 1;
let ClaimResolveRating = 2;
let RatingFunctions = ['Unclaimed', 'Unproblematic', 'ClaimResolveRating'];

async function curRating (ratingContract, type, role, account) {
    return (await ratingContract[type](role, account)).toNumber() / 2 ** 128;
}

async function assertRating (ratingContract, type, role, account, value) {
    let r = await curRating(ratingContract, type, role, account);
    assert.equal(r, value, `Invalid rating for ${account}, role ${Roles[role]}, type ${type}`);
}

class FullRating {
    constructor (ratingContract, address) {
        this.contract = ratingContract;
        this.address = address;
        this.sums = [
            [0.0, 0.0, 0.0],
            [0.0, 0.0, 0.0],
            [0.0, 0.0, 0.0],
        ];
        this.ratings = [
            [0.0, 0.0, 0.0],
            [0.0, 0.0, 0.0],
            [0.0, 0.0, 0.0],
        ];
    }

    async load () {
        for (let role of [roleConsumer, roleMaster, roleWorker]) {
            for (let outcome of [Unclaimed, ClaimAccepted, ClaimRejected]) {
                this.sums[role][outcome] = (await this.contract.Sum(role, outcome, this.address)).toNumber() / 2 ** 128;
            }
            for (let func of [UnclaimedRating, UnproblematicRating, ClaimResolveRating]) {
                let r = await this.contract[RatingFunctions[func]](role, this.address);
                this.ratings[role][func] = r.toNumber() / 2 ** 128;
            }
        }
        return this;
    }

    assertEq (another) {
        for (let role of [roleConsumer, roleMaster, roleWorker]) {
            for (let func of [UnclaimedRating, UnproblematicRating, ClaimResolveRating]) {
                assert.equal(this.ratings[role][func], another.ratings[role][func]);
            }
        }
        return this;
    }
}

function assertEmitLogsUpdated (tx, account1, account2, account3) {
    assert.equal(tx.logs.length, 3);
    assert.equal(tx.logs[0].event, 'RatingUpdated');
    assert.equal(tx.logs[0].args.who, account1);
    assert.equal(tx.logs[1].event, 'RatingUpdated');
    assert.equal(tx.logs[1].args.who, account2);
    assert.equal(tx.logs[2].event, 'RatingUpdated');
    assert.equal(tx.logs[2].args.who, account3);
}

contract('Rating', async function (accounts) {
    let rating;
    let ratingData;
    let owner = accounts[0];
    let somebody = accounts[9];
    let defaultRating = 1.0;
    let defaultSum = 1000;
    let decayValue = 2 ** 127;
    before(async () => {
        let fpLib = await FixedPoint128.new();
        RatingData.link('FixedPoint128', fpLib.address);
        Rating.link('FixedPoint128', fpLib.address);
        ratingData = await RatingData.new(decayValue);
        rating = await Rating.new(ratingData.address, { from: owner });
        ratingData.transferOwnership(rating.address);
    });

    describe('Get decay param', () => {
        it('should return decay param passed in ctor', async () => {
            let curDecayValue = (await rating.DecayValue({ from: somebody }));
            assert.equal(decayValue, curDecayValue.toNumber());
        });
    });

    describe('Set decay param', () => {
        it('should properly set new decay value', async () => {
            let tx = await rating.SetDecayValue(42, { from: owner });
            assert.equal(tx.logs.length, 1);
            assert.equal(tx.logs[0].event, 'DecayValueUpdated');
            assert.equal(tx.logs[0].args.decayValue.toNumber(), 42);
            let curDecayValue = (await rating.DecayValue({ from: somebody }));
            assert.equal(curDecayValue.toNumber(), 42);

            // restore decay value
            tx = await rating.SetDecayValue(decayValue, { from: owner });
            assert.equal(tx.logs.length, 1);
            assert.equal(tx.logs[0].event, 'DecayValueUpdated');
            assert.equal(tx.logs[0].args.decayValue.toNumber(), decayValue);
            curDecayValue = (await rating.DecayValue({ from: somebody }));
            assert.equal(curDecayValue.toNumber(), decayValue);
        });
    });

    describe('Initial rating', () => {
        it('should return 1.0 rating for empty profile', async () => {
            for (let role = 0; role < 3; role++) {
                for (let type of ['Unclaimed', 'Unproblematic', 'ClaimResolveRating']) {
                    await assertRating(rating, type, role, accounts[3], defaultRating);
                }
            }
        });
    });

    describe('Calculate rating', () => {
        let tx;
        it('should return 1.0 rating for unclaimed outcomes', async () => {
            tx = await rating.RegisterOutcome(
                Unclaimed, accounts[1], accounts[2], accounts[3], defaultSum, 1, { from: owner });
            assertEmitLogsUpdated(tx, accounts[1], accounts[2], accounts[3]);
            for (let type of ['Unclaimed', 'Unproblematic', 'ClaimResolveRating']) {
                await assertRating(rating, type, roleConsumer, accounts[1], defaultRating);
                await assertRating(rating, type, roleWorker, accounts[2], defaultRating);
                await assertRating(rating, type, roleMaster, accounts[3], defaultRating);
            }
        });

        it('should decrease rating on negative outcome', async () => {
            tx = await rating.RegisterOutcome(
                ClaimAccepted, accounts[1], accounts[2], accounts[3], defaultSum, 2, { from: owner });
            assertEmitLogsUpdated(tx, accounts[1], accounts[2], accounts[3]);
            await assertRating(rating, 'Unproblematic', roleConsumer, accounts[1], defaultRating);
            await assertRating(rating, 'Unclaimed', roleWorker, accounts[2], 0.3333333333333333);
            await assertRating(rating, 'Unproblematic', roleMaster, accounts[3], 1.0);
        });
    });

    describe('TransferData', () => {
        it('should properly transfer data to new account', async () => {
            let r1 = await new FullRating(rating, accounts[1]).load();
            let r2 = await new FullRating(rating, accounts[2]).load();
            let r3 = await new FullRating(rating, accounts[3]).load();
            let newRating = await Rating.new(ratingData.address, { from: owner });
            r1.assertEq(await new FullRating(newRating, accounts[1]).load());
            r2.assertEq(await new FullRating(newRating, accounts[2]).load());
            r3.assertEq(await new FullRating(newRating, accounts[3]).load());
            await assertRevert(newRating.RegisterOutcome(
                ClaimAccepted, accounts[1], accounts[2], accounts[3], defaultSum, 2, { from: owner })
            );
            await rating.TransferData(newRating.address);
            await assertRevert(rating.RegisterOutcome(
                ClaimAccepted, accounts[1], accounts[2], accounts[3], defaultSum, 2, { from: owner })
            );
            await newRating.RegisterOutcome(
                ClaimAccepted, accounts[1], accounts[2], accounts[3], defaultSum, 2, { from: owner }
            );
            await assertRating(newRating, 'Unproblematic', roleConsumer, accounts[1], defaultRating);
            await assertRating(newRating, 'Unclaimed', roleWorker, accounts[2], 0.2);
            await assertRating(newRating, 'Unproblematic', roleMaster, accounts[3], 1.0);
        });
    });
});
