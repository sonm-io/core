import assertRevert from './helpers/assertRevert';

const MultiSigWallet = artifacts.require('./MultiSigWallet.sol');
const OracleUSD = artifacts.require('./OracleUSD.sol');

contract('MultiSigWallet', async function (accounts) {
    let multisig;

    const owner = accounts[0];
    const master1 = accounts[1];
    const master2 = accounts[2];
    const master3 = accounts[3];

    const creeper = accounts[9];

    describe('Oracle', function () {
        let oracle;

        let defaultPrice = 1;

        before(async function () {
            multisig = await MultiSigWallet.new([master1, master2, master3], 2);
            oracle = await OracleUSD.new();
            await oracle.transferOwnership(multisig.address, { from: owner });
        });

        describe('Default', function () {
            it('current price should set to `1`', async function () {
                let price = (await oracle.getCurrentPrice.call()).toNumber();
                assert.equal(price, defaultPrice);
            });
        });

        describe('SetCurrentPrice', function () {
            let testPrice = 238985;

            it('should set new price', async function () {
                let data = oracle.contract.setCurrentPrice.getData(testPrice);
                // tokenInstance.contract.transfer.getData(accounts[1], 1000000);
                let tx = await multisig.submitTransaction(oracle.address, 0, data, { from: master1 });
                let txId = tx.logs[0].args.transactionId;
                await multisig.confirmTransaction(txId, { from: master2 });
                let price = (await oracle.getCurrentPrice.call()).toNumber();
                assert.equal(price, testPrice);
            });

            describe('when new price lower or equal 0', function () {
                it('should revert', async function () {
                    await assertRevert(oracle.setCurrentPrice(0, { from: owner }));
                });
            });

            describe('when not owner want to set new price', function () {
                it('should revert', async function () {
                    await assertRevert(oracle.setCurrentPrice(testPrice, { from: creeper }));
                });
            });
        });
    });
});
