const Rating = artifacts.require('./Rating.sol');

contract('Rating', async function (accounts) {
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
