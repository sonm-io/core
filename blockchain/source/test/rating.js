import BigNumber from "../node_modules/web3/bower/bignumber.js/bignumber";

const Rating = artifacts.require('./Rating.sol');
const FixedPoint = artifacts.require('./FixedPoint.sol');

contract('FixedPoint', async function (accounts) {
    let fp1;
    let fp10;
    let fp248;
    let fp255;
    let somebody = accounts[9];
    before(async function () {
        fp1 = await FixedPoint.new(1);
        fp10 = await FixedPoint.new(10);
        fp248 = await FixedPoint.new(248);
        fp255 = await FixedPoint.new(255);
    });

    describe('Round', function(accounts) {
        it('should properly round numbers', async function(){
            let fp = fp10;
            let precision = await fp.Precision();
            let val = 1 << precision;
            let round = (await fp.Round(val));
            console.log(round);
            assert.equal(round.toNumber(), 1);
            val *= 42;
            round = (await fp.Round(val));
            assert.equal(round.toNumber(), 42);
            round = (await fp.Round(val + 2 ** (precision - 1)));
            assert.equal(round.toNumber(), 43);
            round = (await fp.Round(val + 2 ** (precision - 1) - 1));
            assert.equal(round, 42);
        });
        it('should properly multiply numbers', async function(){
            let fp = fp248;
            let precision = await fp.Precision();
            let val1 = new BigNumber(2).pow(precision).mul(15);
            let val2 = new BigNumber(2).pow(precision).mul(16);
            console.log(val1.toString(2));
            console.log(val2.toString(2));
            let mul = (await fp.Mul(val1, val2));
            console.log(mul.toString(2));
            let round = (await fp.Round(mul)).toNumber();
            console.log(round);
            assert.equal(round, 15*16);
        });
    });
});

// contract('Rating', async function (accounts) {
//     let rating;
//     let owner = accounts[0];
//     let somebody = accounts[9];
//     // let stranger = accounts[9];
//     let precision = 1024;
//     let defaultRating = 1;
//     let defaultSum = 1000;
//     let defaultDecayValue = 9e17;
//     let defaultDecayEpoch = 2;
//     before(async function () {
//         rating = await Rating.new(defaultDecayEpoch, defaultDecayValue, precision, { from: owner });
//     });
//
//     describe('Get decay param', function() {
//         it('should return decay epoch passed in ctor', async function() {
//             let decayEpoch = (await rating.DecayEpoch({from: somebody}))
//             assert.equal(decayEpoch, defaultDecayEpoch)
//         });
//         it('should return decay param passed in ctor', async function() {
//             let decayValue = (await rating.DecayValue({from: somebody}))
//             assert.equal(decayValue, defaultDecayValue)
//         });
//     });
//
//     describe('Set decay param', function(){
//         it('should set decay epoch', async function() {
//             tx = await rating.SetDecayEpoch(3, {from: owner});
//             assert.equal(tx.logs.length, 1);
//             assert.equal(tx.logs[0].event, 'DecayEpochUpdated');
//             assert.equal(tx.logs[0].args.decayEpoch, 'DecayEpochUpdated');
//             decayEpoch = (await rating.DecayEpoch({from: somebody}))
//             assert.equal(decayEpoch.toNumber(), defaultDecayEpoch);
//         });
//         it('should return decay param passed in ctor', async function() {
//             tx = await rating.SetDecayValue(3, {from: owner});
//             assert.equal(tx.logs.length, 1);
//             assert.equal(tx.logs[0].event, 'DecayEpochUpdated');
//             assert.equal(tx.logs[0].args.decayEpoch, 'DecayEpochUpdated');
//             decayEpoch = (await rating.DecayEpoch({from: somebody}))
//             assert.equal(decayEpoch.toNumber(), defaultDecayEpoch);
//
//         });
//     });
//
//     describe('Calculate decay', function () {
//         it('should correctly calculate decay', async function(){
//
//         });
//     });
//
//     describe('Initial rating', function () {
//         it('should return 1.0 rating for empty profile', async function () {
//             let curRating = (await rating.Current(address[1])).toNumber();
//             assert.equal(curRating, defaultRating);
//         });
//     });
//
//     describe('Calculate rating', function () {
//         let tx;
//         let positiveSumC = 0;
//         let allSumC = 0;
//         let positiveSumS = 0;
//         let allSumS = 0;
//         let blockNum = 0
//         it('should return 1.0 rating for all positive outcomes', async function() {
//             for (i = 0; i <5; i++) {
//                 tx = await rating.RegisterOutcome(0, accounts[1], accounts[2], defaultSum, {from: owner});
//                 if (blockNum == 0) {
//                     blockNum = tx.blockNumber
//                 }
//                 decay = calculateDecay()
//                 positiveSumC = decayValue ** (positiveSumC * precision / decayValue)
//                 positiveSumC += defaultSum
//                 positiveSumC += defaultSum
//                 assert.equal(tx.logs.length, 2);
//                 assert.equal(tx.logs[0].event, 'RatingUpdated');
//                 assert.equal(tx.logs[0].args.rating, defaultRating);
//                 assert.equal(tx.logs[0].args.who, accounts[1]);
//                 assert.equal(tx.logs[1].event, 'RatingUpdated');
//                 assert.equal(tx.logs[1].args.rating, defaultRating);
//                 assert.equal(tx.logs[1].args.who, accounts[2]);
//                 rate1 = await rating.Current(accounts[1])
//                 assert.equal(rate1.toNumber(), defaultRating)
//                 rate2 = await rating.Current(accounts[2])
//                 assert.equal(rate2.toNumber(), defaultRating)
//             }
//         });
//
//         it('should decrease rating on negative outcome', async function(){
//             rating =
//             tx = await rating.RegisterOutcome(1, accounts[1], accounts[2], defaultSum, {from: owner});
//             assert.equal(tx.logs.length, 2);
//             assert.equal(tx.logs[0].event, 'RatingUpdated');
//             assert.equal(tx.logs[0].args.rating, defaultRating);
//             assert.equal(tx.logs[0].args.who, accounts[1]);
//             assert.equal(tx.logs[1].event, 'RatingUpdated');
//             assert.equal(tx.logs[1].args.rating, defaultRating);
//             assert.equal(tx.logs[1].args.who, accounts[2]);
//             rate1 = await rating.Current(accounts[1])
//             assert.equal(rate1.toNumber(), defaultRating)
//             rate2 = await rating.Current(accounts[2])
//             assert.equal(rate2.toNumber(), defaultRating)
//         })
//     });
// });
