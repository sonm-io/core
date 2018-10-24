import BigNumber from "../node_modules/web3/bower/bignumber.js/bignumber";
import assertRevert from "./helpers/assertRevert";

const Rating = artifacts.require('./Rating.sol');
const FixedPoint = artifacts.require('./FixedPoint.sol');

contract('FixedPoint', async function (accounts) {
    let fp1;
    let fp10;
    let fp40;
    let fp128;
    let fp248;
    let fp255;
    let somebody = accounts[9];
    before(async function () {
        fp1 = await FixedPoint.new(1);
        fp10 = await FixedPoint.new(10);
        fp40 = await FixedPoint.new(10);
        fp128 = await FixedPoint.new(128);
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

        it('should properly construct fp numbers from natural numbers', async function(){
            let fp = fp248;
            let mul = (await fp.FromNatural(255));
            let round = (await fp.Round(mul)).toNumber();
            assert.equal(round, 255);
            assertRevert(fp.FromNatural(256));
        });

        it('should properly sum numbers and detect overflow', async function(){
            let fp = fp248;
            let val1 = (await fp.FromNatural(127));
            let val2 = (await fp.FromNatural(128));
            let sum = (await fp.Add(val1, val2));
            let round = (await fp.Round(sum)).toNumber();
            assert.equal(round, 255);
            assertRevert(fp.Add(val2, val2));
            let one = (await fp.FromNatural(1));
            assertRevert(fp.Add(sum, one));
        });

        it('should properly sub numbers and detect underflow', async function(){
            let fp = fp248;
            let val1 = (await fp.FromNatural(127));
            let val2 = (await fp.FromNatural(128));
            let diff = (await fp.Sub(val2, val1));
            let round = (await fp.Round(diff)).toNumber();
            assert.equal(round, 1);
            assertRevert(fp.Sub(val1, val2));
        });

        it('should properly multiply numbers', async function(){
            let fp = fp248;
            let precision = await fp.Precision();
            let val1 = new BigNumber(2).pow(precision).mul(15);
            let val2 = new BigNumber(2).pow(precision).mul(16);
            let mul = (await fp.Mul(val1, val2));
            let round = (await fp.Round(mul)).toNumber();
            assert.equal(round, 15*16);
            assertRevert(fp.Mul(val2, val2));
        });

        it('should properly divide numbers', async function(){
            let fp = fp10;
            let div = (await fp.Div(1024, 2048)).toNumber();
            assert.equal(div, 512);
            div = (await fp.Div(1024, 1024*1024)).toNumber();
            assert.equal(div, 1);
            div = (await fp.Div(1024, 1024*1024 + 1)).toNumber();
            assert.equal(div, 0);
            div = (await fp.Div(1024, 128)).toNumber();
            assert.equal(div, 1024*8);
        });

        it('should properly make integer power', async function(){
            let fp = fp10;
            let num1 = (await fp.FromNatural(3)).toNumber();
            let num2 = (await fp.FromNatural(2)).toNumber();
            let num3 = (await fp.Div(num1, num2)).toNumber();
            let num4 = (await fp.Ipow(num3, 13)).toNumber()
            let round = (await fp.Round(num4)).toNumber();
            // 1.5^13
            assert.equal(round, 195);

            fp = fp128;
            num1 = (await fp.FromNatural(101)).toNumber();
            num2 = (await fp.FromNatural(100)).toNumber();
            num3 = (await fp.Div(num1, num2)).toNumber();
            num4 = (await fp.Ipow(num3, 70)).toNumber()
            round = (await fp.Round(num4)).toNumber();
            // 1.01^70
            assert.equal(round, 2);

            fp = fp128;
            num1 = (await fp.FromNatural(101)).toNumber();
            num2 = (await fp.FromNatural(100)).toNumber();
            num3 = (await fp.Div(num1, num2)).toNumber();
            num4 = (await fp.Ipow(num3, 3000)).toNumber()
            round = (await fp.Round(num4)).toNumber();
            // 1.01^3000
            assert.equal(round, 9207067941190);
        });
    });
});
