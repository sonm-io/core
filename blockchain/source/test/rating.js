const FixedPoint = artifacts.require('./FixedPoint.sol');
const FixedPoint128 = artifacts.require('./FixedPoint128.sol');
const RatingData = artifacts.require('./RatingData.sol');
const Rating = artifacts.require('./Rating.sol');

contract('Rating', async function (accounts) {
    let rating;
    let ratingData;
    let owner = accounts[0];
    let somebody = accounts[9];
    let defaultRating;
    let defaultSum = 1000;
    let decayValue;
    let fp;
    before(async function () {
        fp = await FixedPoint.new(128);
        let decayDivider = await fp.FromNatural(8);
        decayValue = await fp.FromNatural(1);
        defaultRating = await fp.FromNatural(1);
        decayValue = await fp.Div(decayValue, decayDivider);
        ratingData = await RatingData.new(decayValue);
        rating = await Rating.new(ratingData.address, { from: owner })
        ratingData.transferOwnership(rating.address);
    });

    describe('Get decay param', function() {
        it('should return decay param passed in ctor', async function() {
            let curDecayValue = (await rating.DecayValue({from: somebody}));
            assert.equal(decayValue.toNumber(), curDecayValue.toNumber());
        });
    });

    describe('Set decay param', function(){
        it('should properly set new decay value', async function() {
            let tx = await rating.SetDecayValue(42, {from: owner});
            assert.equal(tx.logs.length, 1);
            assert.equal(tx.logs[0].event, 'DecayValueUpdated');
            assert.equal(tx.logs[0].args.decayValue.toNumber(), 42);
            let curDecayValue = (await rating.DecayValue({from: somebody}))
            assert.equal(curDecayValue.toNumber(), 42);

            // restore decay value
            tx = await rating.SetDecayValue(decayValue, {from: owner});
            assert.equal(tx.logs.length, 1);
            assert.equal(tx.logs[0].event, 'DecayValueUpdated');
            assert.equal(tx.logs[0].args.decayValue.toNumber(), decayValue.toNumber());
            curDecayValue = (await rating.DecayValue({from: somebody}))
            assert.equal(curDecayValue.toNumber(), decayValue.toNumber());
        });
    });

    describe('Initial rating', function () {
        it('should return 1.0 rating for empty profile', async function () {
            for(var role = 0; role < 3; role ++) {
                let curRating = (await rating.Unclaimed(role, accounts[3])).toNumber();
                assert.equal(curRating, defaultRating);
                curRating = (await rating.Unproblematic(role, accounts[3])).toNumber();
                assert.equal(curRating, defaultRating);
                curRating = (await rating.ClaimResolveRating(role, accounts[3])).toNumber();
                assert.equal(curRating, defaultRating);
            }
        });
    });

    describe('Calculate rating', function () {
        let tx;
        let positiveSumC = 0;
        let allSumC = 0;
        let positiveSumS = 0;
        let allSumS = 0;
        let blockNum = 0
        it('should return 1.0 rating for all positive outcomes', async function() {
            for (i = 0; i <5; i++) {
                tx = await rating.RegisterOutcome(0, accounts[1], accounts[2], defaultSum, {from: owner});
                decay = calculateDecay();
                positiveSumC = decayValue ** (positiveSumC * precision / decayValue)
                positiveSumC += defaultSum;
                assert.equal(tx.logs.length, 2);
                assert.equal(tx.logs[0].event, 'RatingUpdated');
                assert.equal(tx.logs[0].args.rating, defaultRating);
                assert.equal(tx.logs[0].args.who, accounts[1]);
                assert.equal(tx.logs[1].event, 'RatingUpdated');
                assert.equal(tx.logs[1].args.rating, defaultRating);
                assert.equal(tx.logs[1].args.who, accounts[2]);
                rate1 = await rating.Current(accounts[1])
                assert.equal(rate1.toNumber(), defaultRating)
                rate2 = await rating.Current(accounts[2])
                assert.equal(rate2.toNumber(), defaultRating)
            }
        });

        it('should decrease rating on negative outcome', async function(){
            rating =
            tx = await rating.RegisterOutcome(1, accounts[1], accounts[2], defaultSum, {from: owner});
            assert.equal(tx.logs.length, 2);
            assert.equal(tx.logs[0].event, 'RatingUpdated');
            assert.equal(tx.logs[0].args.rating, defaultRating);
            assert.equal(tx.logs[0].args.who, accounts[1]);
            assert.equal(tx.logs[1].event, 'RatingUpdated');
            assert.equal(tx.logs[1].args.rating, defaultRating);
            assert.equal(tx.logs[1].args.who, accounts[2]);
            rate1 = await rating.Current(accounts[1])
            assert.equal(rate1.toNumber(), defaultRating)
            rate2 = await rating.Current(accounts[2])
            assert.equal(rate2.toNumber(), defaultRating)
        })
    });

    describe('TransferData', function() {
        it('should properly transfer data to new account', async function(){
            let decayDivider = await fp.FromNatural(8);
            decayValue = await fp.FromNatural(1);
            defaultRating = await fp.FromNatural(1);
            decayValue = await fp.Div(decayValue, decayDivider);
            ratingData = await RatingData.new(decayValue);
            rating = await Rating.new(ratingData.address, { from: owner })
            ratingData.transferOwnership(rating.address);
        });
    });
});
