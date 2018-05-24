import assertRevert from './helpers/assertRevert';

const OracleUSD = artifacts.require('./OracleUSD.sol');

contract('OracleUSD', async function (accounts) {
    let oracle;
    let owner = accounts[0];
    let stranger = accounts[9];
    let defaultPrice = 1;

    before(async function () {
        oracle = await OracleUSD.new({ from: owner });
    });

    describe('Default', function () {
        it('current price should set to `1`', async function () {
            let price = (await oracle.getCurrentPrice.call()).toNumber();
            assert.equal(price, defaultPrice);
        });
    });

    describe('SetCurrentPrice', function () {
        let testPrice = 238985;
        let tx;

        it('should set new price', async function () {
            tx = await oracle.setCurrentPrice(testPrice, { from: owner });
        });

        it('should spend `PriceChanged` event', async function () {
            assert.equal(tx.logs.length, 1);
            assert.equal(tx.logs[0].event, 'PriceChanged');
            assert.equal(tx.logs[0].args.price, testPrice);
        });

        describe('when new price lower or equal 0', function () {
            it('should revert', async function () {
                await assertRevert(oracle.setCurrentPrice(0, { from: owner }));
            });
        });

        describe('when not owner want to set new price', function () {
            it('should revert', async function () {
                await assertRevert(oracle.setCurrentPrice(testPrice, { from: stranger }));
            });
        });
    });
});
