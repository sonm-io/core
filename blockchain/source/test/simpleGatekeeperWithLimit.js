import assertRevert from './helpers/assertRevert';
import increaseTime from './helpers/increaseTime';
import { eventInTransaction } from './helpers/expectEvent';

let SimpleGatekeeperWithLimit = artifacts.require('./SimpleGatekeeperWithLimit.sol');
const SNM = artifacts.require('./SNM.sol');
const MultiSig = artifacts.require('./MultiSigWallet.sol');

let simpleGateTests = function (accounts) {
    let token;
    let gatekeeper;

    let owner = accounts[0];
    let user = accounts[1];
    let creeper = accounts[3];

    let keeper = accounts[5];

    const oneMinute = 60;
    const fiveMinute = 300;

    describe('PayIn', function () {
        let tx;

        let startGatekeeperBalance;
        let startUserBalance;

        let testValue = 100;

        before(async function () {
            token = await SNM.new({ from: owner });
            gatekeeper = await SimpleGatekeeperWithLimit.new(token.address, fiveMinute, { from: owner });

            await token.transfer(user, testValue, { from: owner });
            await token.approve(gatekeeper.address, testValue, { from: user });

            startGatekeeperBalance = (await token.balanceOf(gatekeeper.address)).toNumber();
            startUserBalance = (await token.balanceOf(user)).toNumber();
        });

        it('should exec', async function () {
            tx = await gatekeeper.Payin(testValue, { from: user });
        });

        it('should transfer token from user to gatekeeper', async function () {
            let endGatekeeperBalance = (await token.balanceOf(gatekeeper.address)).toNumber();
            assert.equal(startGatekeeperBalance, endGatekeeperBalance - testValue);

            let endUserBalance = (await token.balanceOf(user)).toNumber();
            assert.equal(startUserBalance, endUserBalance + testValue);
        });

        it('should spend `PayOut` event', function () {
            assert.equal(tx.logs.length, 1);
            assert.equal(tx.logs[0].event, 'PayinTx');
            assert.equal(tx.logs[0].args.from, user);
            assert.equal(tx.logs[0].args.txNumber, 1);
            assert.equal(tx.logs[0].args.value, testValue);
        });

        describe('when user hasnt balance', function () {
            let userWithoutBalance = accounts[3];

            before(async function () {
                await token.approve(gatekeeper.address, testValue, { from: user });
            });

            it('should revert', async function () {
                await assertRevert(gatekeeper.Payin(testValue, { from: userWithoutBalance }));
            });
        });

        describe('when user hasnt allowance', function () {
            let userWithoutAllowance = accounts[4];

            before(async function () {
                await token.transfer(userWithoutAllowance, testValue, { from: owner });
            });

            it('should revert', async function () {
                await assertRevert(gatekeeper.Payin(testValue, { from: userWithoutAllowance }));
            });
        });
    });

    describe('Payout', function () {
        let tx;

        let startGatekeeperBalance;
        let startUserBalance;

        let testValue = 100;

        let lowKeeper = accounts[3];
        let highKeeper = accounts[4];

        let lowLimit = 100;
        let highLimit = 10000;

        before(async function () {
            token = await SNM.new({ from: owner });
            gatekeeper = await SimpleGatekeeperWithLimit.new(token.address, fiveMinute, { from: owner });

            await token.transfer(gatekeeper.address, testValue, { from: owner });

            startGatekeeperBalance = (await token.balanceOf(gatekeeper.address)).toNumber();
            startUserBalance = (await token.balanceOf(user)).toNumber();

            await gatekeeper.ChangeKeeperLimit(lowKeeper, lowLimit, { from: owner });
            await gatekeeper.ChangeKeeperLimit(highKeeper, highLimit, { from: owner });
        });

        it('should exec', async function () {
            await gatekeeper.Payout(user, testValue, 1, { from: highKeeper });
            await increaseTime(fiveMinute + 1);
            tx = await gatekeeper.Payout(user, testValue, 1, { from: highKeeper });
        });

        it('should transfer token from gatekeeper to user', async function () {
            let endGatekeeperBalance = (await token.balanceOf(gatekeeper.address)).toNumber();
            assert.equal(startGatekeeperBalance - testValue, endGatekeeperBalance);

            let endUserBalance = (await token.balanceOf(user)).toNumber();
            assert.equal(startUserBalance + testValue, endUserBalance);
        });

        it('should spend `PayOut` event', function () {
            assert.equal(tx.logs.length, 1);
            assert.equal(tx.logs[0].event, 'PayoutTx');
            assert.equal(tx.logs[0].args.from, user);
            assert.equal(tx.logs[0].args.txNumber, 1);
            assert.equal(tx.logs[0].args.value, testValue);
        });

        describe('when transaction already paid', function () {
            it('should revert', async function () {
                await assertRevert(gatekeeper.Payout(user, testValue, 1, { from: highKeeper }));
            });
        });

        describe('when keeper want release tx before freeze time', function () {
            it('should revert', async function () {
                await gatekeeper.Payout(user, testValue, 2, { from: highKeeper });
                await increaseTime(fiveMinute - 100);
                await assertRevert(gatekeeper.Payout(user, testValue, 1, { from: highKeeper }));
            });
        });

        describe('when other keeper want release tx', function () {
            it('should revert', async function () {
                await assertRevert(gatekeeper.Payout(user, testValue, 2, { from: lowKeeper }));
            });
        });

        describe('when keeper is not under daily limit', function () {
            it('should revert', async function () {
                await assertRevert(gatekeeper.Payout(user, highLimit, 10, { from: lowKeeper }));
            });
        });

        describe('when freezed keeper want to release tx', function () {
            before(async function () {
                await gatekeeper.Payout(user, testValue, 15, { from: highKeeper });
                await gatekeeper.FreezeKeeper(highKeeper, { from: lowKeeper });
            });

            it('should revert', async function () {
                await assertRevert(gatekeeper.Payout(user, testValue, 15, { from: highKeeper }));
            });
        });

        describe('when freezed keeper want to submit tx', function () {
            it('should revert', async function () {
                await assertRevert(gatekeeper.Payout(user, testValue, 20, { from: highKeeper }));
            });
        });

        describe('when keeper renew it limit', function () {
            before(async function () {
                await increaseTime(86500);
                await gatekeeper.Payout(user, lowLimit, 25, { from: lowKeeper });
                await increaseTime(86500);
            });

            it('should exec correct', async function () {
                await gatekeeper.Payout(user, lowLimit, 30, { from: lowKeeper });
            });
        });

        describe('when not keeper want to submit transaction', function () {
            it('should revert', async function () {
                await assertRevert(gatekeeper.Payout(user, testValue, 35, { from: creeper }));
            });
        });
    });

    describe('FreezeKeeper', function () {
        let tx;
        let freezedKeeper = accounts[6];
        before(async function () {
            token = await SNM.new({ from: owner });
            gatekeeper = await SimpleGatekeeperWithLimit.new(token.address, fiveMinute, { from: owner });
            await gatekeeper.ChangeKeeperLimit(keeper, 1, { from: owner });
            await gatekeeper.ChangeKeeperLimit(freezedKeeper, 1, { from: owner });
        });

        it('should exec', async function () {
            tx = await gatekeeper.FreezeKeeper(freezedKeeper, { from: keeper });
        });

        it('should spend `KeeperFreezed`', async function () {
            await eventInTransaction(tx, 'KeeperFreezed');
        });

        describe('when keeper want to freeze not keeper address', function () {
            it('should revert', async function () {
                await assertRevert(gatekeeper.FreezeKeeper(creeper, { from: keeper }));
            });
        });

        describe('when not keeper want to freeze', function () {
            it('should revert', async function () {
                await assertRevert(gatekeeper.FreezeKeeper(keeper, { from: creeper }));
            });
        });
    });

    describe('UnfreezeKeeper', function () {
        let tx;
        before(async function () {
            token = await SNM.new({ from: owner });
            gatekeeper = await SimpleGatekeeperWithLimit.new(token.address, fiveMinute, { from: owner });
            await gatekeeper.ChangeKeeperLimit(keeper, 1, { from: owner });
        });

        it('should exec', async function () {
            tx = await gatekeeper.UnfreezeKeeper(keeper, { from: owner });
        });

        it('should spend `KeeperUnfreezed` event', async function () {
            await eventInTransaction(tx, 'KeeperUnfreezed');
        });

        describe('when not owner want unfreeze', function () {
            it('should revert', async function () {
                await assertRevert(gatekeeper.UnfreezeKeeper(keeper, { from: creeper }));
            });
        });

        describe('when owner want unfreeze not keeper address', function () {
            it('should revert', async function () {
                await assertRevert(gatekeeper.UnfreezeKeeper(creeper, { from: owner }));
            });
        });
    });

    describe('SetFreezingTime', function () {
        let newFreezeTime = 1345;
        before(async function () {
            token = await SNM.new({ from: owner });
            gatekeeper = await SimpleGatekeeperWithLimit.new(token.address, fiveMinute, { from: owner });
        });

        it('should changed', async function () {
            let beforeChanging = (await gatekeeper.GetFreezingTime.call()).toNumber();
            await gatekeeper.SetFreezingTime(newFreezeTime, { from: owner });
            let afterChanging = (await gatekeeper.GetFreezingTime.call()).toNumber();
            assert.notEqual(beforeChanging, afterChanging);
            assert.equal(afterChanging, newFreezeTime);
        });

        describe('when not owner want to change freezing time', function () {
            it('should revert', async function () {
                await assertRevert(gatekeeper.SetFreezingTime(1000, { from: creeper }));
            });
        });
    });

    describe('ChangeKeeperLimit', function () {
        let tx;
        let newKeeperLimit = 65430;

        before(async function () {
            token = await SNM.new({ from: owner });
            gatekeeper = await SimpleGatekeeperWithLimit.new(token.address, fiveMinute, { from: owner });
        });

        it('should changed', async function () {
            tx = await gatekeeper.ChangeKeeperLimit(keeper, newKeeperLimit, { from: owner });
        });

        it('should spend `LimitChanged` event', async function () {
            await eventInTransaction(tx, 'LimitChanged');
        });

        describe('when not owner should change limit', function () {
            it('should revert', async function () {
                await assertRevert(gatekeeper.ChangeKeeperLimit(keeper, 1000, { from: creeper }));
            });
        });
    });

    describe('kill', async function () {
        let startOwnerBalance;
        let startGatekeeperBalance;
        let diff;

        let tx;

        before(async function () {
            token = await SNM.new({ from: owner });
            gatekeeper = await SimpleGatekeeperWithLimit.new(token.address, fiveMinute, { from: owner });

            let ownerBalance = (await token.balanceOf(owner)).toNumber();
            await token.transfer(gatekeeper.address, ownerBalance, { from: owner });

            startOwnerBalance = await token.balanceOf(owner);
            startGatekeeperBalance = await token.balanceOf(gatekeeper.address);
            diff = startGatekeeperBalance.toNumber() - startOwnerBalance.toNumber();
        });

        it('should killed', async function () {
            tx = await gatekeeper.kill({ from: owner });
        });

        it('should transfer gatekeeper balance to owner', async function () {
            let endOwnerBalance = await token.balanceOf(owner);
            let endGatekeeperBalance = await token.balanceOf(gatekeeper.address);

            assert.equal(endGatekeeperBalance.toNumber(), 0);
            assert.equal(endOwnerBalance.toNumber() - endGatekeeperBalance.toNumber(), diff);
        });

        it('should spend `Suicide` event', function () {
            assert.equal(tx.logs.length, 1);
            assert.equal(tx.logs[0].event, 'Suicide');
        });

        describe('when not owner want to kill contract', function () {
            before(async function () {
                token = await SNM.new();
                gatekeeper = await SimpleGatekeeperWithLimit.new(token.address, fiveMinute);
            });

            it('should revert', async function () {
                await assertRevert(gatekeeper.kill({ from: creeper }));
            });
        });
    });

    describe('multisig', function () {
        let token;
        let gatekeeper;
        let multisig;

        let keeper = accounts[4];
        let freezedKeeper = accounts[5];

        let owner1 = accounts[6];
        let owner2 = accounts[7];
        let owner3 = accounts[8];

        async function execMultiSig (data) {
            let tx = await multisig.submitTransaction(gatekeeper.address, 0, data, { from: owner3 });
            let txId = tx.logs[0].args.transactionId;
            return multisig.confirmTransaction(txId, { from: owner2 });
        }

        before(async function () {
            token = await SNM.new({ from: owner });
            gatekeeper = await SimpleGatekeeperWithLimit.new(token.address, oneMinute, { from: owner });
            multisig = await MultiSig.new([owner1, owner2, owner3], 2);
        });

        it('should change owner', async function () {
            assert.equal(await gatekeeper.owner.call(), owner);
            await gatekeeper.transferOwnership(multisig.address, { from: owner });
            let newOwner = await gatekeeper.owner.call();
            assert.equal(newOwner, multisig.address);
        });

        it('should change owner through MultiSig', async function () {
            assert.equal(await gatekeeper.owner.call(), multisig.address);
            let data = await gatekeeper.contract.transferOwnership.getData(owner);
            await execMultiSig(data);
            let newOwner = await gatekeeper.owner.call();
            assert.equal(newOwner, owner);

            // return past state
            await gatekeeper.transferOwnership(multisig.address, { from: owner });
        });

        it('should ChangeKeeperLimit through MultiSig', async function () {
            let data = await gatekeeper.contract.ChangeKeeperLimit.getData(keeper, 1000000);
            await execMultiSig(data);
        });

        it('should UnfreezeKeeper through MultiSig', async function () {
            let data = await gatekeeper.contract.ChangeKeeperLimit.getData(freezedKeeper, 1000000);
            await execMultiSig(data);

            await gatekeeper.FreezeKeeper(freezedKeeper, { from: keeper });

            data = gatekeeper.contract.UnfreezeKeeper.getData(freezedKeeper);
            await execMultiSig(data);
        });

        it('should Change freeze time through MultiSig', async function () {
            let freezingTime = (await gatekeeper.GetFreezingTime.call()).toNumber();

            await assertRevert(gatekeeper.SetFreezingTime(fiveMinute, { from: owner }));

            let data = await gatekeeper.contract.SetFreezingTime.getData(fiveMinute);
            let tx = await multisig.submitTransaction(gatekeeper.address, 0, data, { from: owner3 });
            let txId = tx.logs[0].args.transactionId;
            await multisig.confirmTransaction(txId, { from: owner2 });

            let newFreezingTime = (await gatekeeper.GetFreezingTime.call()).toNumber();
            assert.notEqual(freezingTime, newFreezingTime);
            assert.equal(fiveMinute, newFreezingTime);
        });

        it('should killed through MultiSig', async function () {
            let startOwnerBalance;
            let startGatekeeperBalance;
            let diff;

            startOwnerBalance = await token.balanceOf(multisig.address);
            startGatekeeperBalance = await token.balanceOf(gatekeeper.address);
            diff = startGatekeeperBalance.toNumber() - startOwnerBalance.toNumber();

            let data = await gatekeeper.contract.kill.getData();
            await execMultiSig(data);

            let endOwnerBalance = await token.balanceOf(multisig.address);
            let endGatekeeperBalance = await token.balanceOf(gatekeeper.address);

            assert.equal(endGatekeeperBalance.toNumber(), 0);
            assert.equal(endOwnerBalance.toNumber() - endGatekeeperBalance.toNumber(), diff);
        });

        it('should migrate to new gate', async function () {
            let startOwnerBalance;
            let startGatekeeperBalance;
            let diff;

            gatekeeper = await SimpleGatekeeperWithLimit.new(token.address, oneMinute, { from: owner });

            startOwnerBalance = await token.balanceOf(multisig.address);
            startGatekeeperBalance = await token.balanceOf(gatekeeper.address);
            diff = startGatekeeperBalance.toNumber() - startOwnerBalance.toNumber();

            let data = token.contract.transfer.getData(gatekeeper.address, startOwnerBalance);
            await execMultiSig(data);

            let endOwnerBalance = await token.balanceOf(multisig.address);
            let endGatekeeperBalance = await token.balanceOf(gatekeeper.address);

            assert.equal(endGatekeeperBalance.toNumber(), 0);
            assert.equal(endGatekeeperBalance.toNumber() - endOwnerBalance.toNumber(), diff);
        });
    });
};

contract('SimpleGatekeeperWithLimit', simpleGateTests);

SimpleGatekeeperWithLimit = artifacts.require('./SimpleGatekeeperWithLimitLive.sol');

contract('SimpleGatekeeperWithLimitLive', simpleGateTests);
