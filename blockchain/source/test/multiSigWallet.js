import assertRevert from './helpers/assertRevert'
import {eventInTransaction} from './helpers/expectEvent'
const MultiSigWallet = artifacts.require('./MultiSigWallet.sol');
const OracleUSD = artifacts.require('./OracleUSD.sol');

contract('MultiSigWallet', async (accounts) => {
    let multisig;

    const owner = accounts[0];
    const master1 = accounts[1];
    const master2 = accounts[2];
    const masterReplacee = accounts[3];
    const masterReplacer = accounts[4];

    const creeper = accounts[9];

    describe('Oracle', () => {
        let oracle;

        let defaultPrice = 1;

        before(async () => {
            multisig = await MultiSigWallet.new([master1, master2], 2);
            oracle = await OracleUSD.new();
            await oracle.transferOwnership(multisig.address, { from: owner });
        });

        describe('Default', () => {
            it('current price should set to `1`', async () => {
                let price = (await oracle.getCurrentPrice.call()).toNumber();
                assert.equal(price, defaultPrice);
            });
        });

        describe('Add/Remove owner', () => {
            it('add owner', async () => {
                let data = multisig.contract.addOwner.getData(masterReplacee);
                let tx = await multisig.submitTransaction(multisig.address, 0, data,
                    { from: master1 });
                let txId = tx.logs[0].args.transactionId;
                await multisig.confirmTransaction(txId, { from: master2 });
                let oAfter = await multisig.getOwners();
                assert.deepEqual(oAfter,
                    ['0x6704fbfcd5ef766b287262fa2281c105d57246a6',
                        '0x9e1ef1ec212f5dffb41d35d9e5c14054f26c6560',
                        '0xce42bdb34189a93c55de250e011c68faee374dd3']);
            });

            it('remove owner', async () => {
                let data = multisig.contract.removeOwner.getData(masterReplacee);
                let tx = await multisig.submitTransaction(multisig.address, 0, data,
                    { from: master1 });
                let txId = tx.logs[0].args.transactionId;
                await multisig.confirmTransaction(txId, { from: master2 });

                let oAfter = await multisig.getOwners();
                assert.deepEqual(oAfter,
                    ['0x6704fbfcd5ef766b287262fa2281c105d57246a6',
                        '0x9e1ef1ec212f5dffb41d35d9e5c14054f26c6560']);
            });

            it('replace owner', async () => {
                let data = multisig.contract.addOwner.getData(masterReplacee);
                let tx = await multisig.submitTransaction(multisig.address, 0, data,
                    { from: master1 });
                let txId = tx.logs[0].args.transactionId;
                await multisig.confirmTransaction(txId, { from: master2 });

                let replaceData = multisig.contract.replaceOwner.getData(masterReplacee, masterReplacer);
                let txReplace = await multisig.submitTransaction(multisig.address, 0, replaceData,
                    { from: master1 });
                let txReplaceId = txReplace.logs[0].args.transactionId;
                await multisig.confirmTransaction(txReplaceId, { from: master2 });

                let oAfter = await multisig.getOwners();
                assert.deepEqual(oAfter,
                    ['0x6704fbfcd5ef766b287262fa2281c105d57246a6',
                        '0x9e1ef1ec212f5dffb41d35d9e5c14054f26c6560',
                        '0x97a3fc5ee46852c1cf92a97b7bad42f2622267cc']);
            });

            it('remove owner in the middle of array', async () => {
                let data = multisig.contract.removeOwner.getData(master2);
                let tx = await multisig.submitTransaction(multisig.address, 0, data,
                    { from: master1 });
                let txId = tx.logs[0].args.transactionId;
                await multisig.confirmTransaction(txId, { from: masterReplacer });

                let oAfter = await multisig.getOwners();
                assert.deepEqual(oAfter,
                    ['0x6704fbfcd5ef766b287262fa2281c105d57246a6',
                        '0x97a3fc5ee46852c1cf92a97b7bad42f2622267cc']);

                // add master2 back
                let addData = multisig.contract.addOwner.getData(master2);
                let addTx = await multisig.submitTransaction(multisig.address, 0, addData,
                    { from: master1 });
                let addTxId = addTx.logs[0].args.transactionId;
                await multisig.confirmTransaction(addTxId, { from: masterReplacer });
            });

            it('remove 2 masters, confirmation requirement set to 1 owner', async () => {

                let data = multisig.contract.removeOwner.getData(masterReplacer);
                let tx = await multisig.submitTransaction(multisig.address, 0, data,
                    { from: master1 });
                let txId = tx.logs[0].args.transactionId;
                await multisig.confirmTransaction(txId, { from: master2 });

                let data2 = multisig.contract.removeOwner.getData(master2);
                let tx2 = await multisig.submitTransaction(multisig.address, 0, data2,
                    { from: master1 });
                let txId2 = tx2.logs[0].args.transactionId;
                let tx3 = await multisig.confirmTransaction(txId2, { from: master2 });

                await eventInTransaction(tx3, 'RequirementChange');

                //add master2 back
                let addData = multisig.contract.addOwner.getData(master2);
                let addTx = await multisig.submitTransaction(multisig.address, 0, addData,
                    { from: master1 });
            });
        });

        describe('SetCurrentPrice: check submit & confirm', () => {
            let testPrice = 238985;

            it('should set new price', async () => {
                let data = oracle.contract.setCurrentPrice.getData(testPrice);
                let tx = await multisig.submitTransaction(oracle.address, 0, data, { from: master1 });
                let txId = tx.logs[0].args.transactionId;
                await multisig.confirmTransaction(txId, { from: master2 });
                let price = (await oracle.getCurrentPrice.call()).toNumber();
                assert.equal(price, testPrice);
            });

            it('should revert confirm from creeper', async () => {
                let data = oracle.contract.setCurrentPrice.getData(testPrice + 1);
                let tx = await multisig.submitTransaction(oracle.address, 0, data, { from: master1 });
                let txId = tx.logs[0].args.transactionId;
                await assertRevert(multisig.confirmTransaction(txId, { from: creeper }));
                let price = (await oracle.getCurrentPrice.call()).toNumber();
                assert.equal(price, testPrice);
            });

            it('should revert submit from creeper', async () => {
                let data = oracle.contract.setCurrentPrice.getData(testPrice + 1);
                await assertRevert(multisig.submitTransaction(oracle.address, 0, data, { from: creeper }));
                let price = (await oracle.getCurrentPrice.call()).toNumber();
                assert.equal(price, testPrice);
            });

            it('should revert if tx not exists', async () => {
                await assertRevert(multisig.confirmTransaction(127836182736, { from: master2 }));
            });

            describe('when new price lower or equal 0', () => {
                it('should revert', async () => {
                    await assertRevert(oracle.setCurrentPrice(0, { from: owner }));
                });
            });

            describe('when not owner want to set new price', () => {
                it('should revert', async () => {
                    await assertRevert(oracle.setCurrentPrice(testPrice, { from: creeper }));
                });
            });
        });
    });
});
