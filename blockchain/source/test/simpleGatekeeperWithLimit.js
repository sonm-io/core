import assertRevert from './helpers/assertRevert';
import increaseTime from './helpers/increaseTime';

const SimpleGatekeeperWithLimit = artifacts.require('./SimpleGatekeeperWithLimit.sol');
const SNM = artifacts.require('./SNM.sol');

contract('SimpleGatekeeperWithLimit', async function (accounts) {
    describe('PayIn', function () {
        let token;
        let gatekeeper;

        let owner = accounts[0];
        let user = accounts[1];

        let tx;

        let startGatekeeperBalance;
        let startUserBalance;

        let testValue = 100;

        before(async function () {
            token = await SNM.new({ from: owner });
            gatekeeper = await SimpleGatekeeperWithLimit.new(token.address, { from: owner });

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
        let token;
        let gatekeeper;

        let owner = accounts[0];
        let user = accounts[1];

        let tx;

        let startGatekeeperBalance;
        let startUserBalance;

        let testValue = 100;

        let lowKeeper = accounts[3];
        let highKeeper = accounts[4];

        let lowLimit = 10;
        let highLimit = 1000;

        before(async function () {
            token = await SNM.new({ from: owner });
            gatekeeper = await SimpleGatekeeperWithLimit.new(token.address, { from: owner });

            await token.transfer(gatekeeper.address, testValue, { from: owner });

            startGatekeeperBalance = (await token.balanceOf(gatekeeper.address)).toNumber();
            startUserBalance = (await token.balanceOf(user)).toNumber();

            await gatekeeper.ChangeKeeperLimit(lowKeeper, lowLimit, { from: owner });
            await gatekeeper.ChangeKeeperLimit(highKeeper, highLimit, { from: owner });
        });

        it('should exec', async function () {
            await gatekeeper.Payout(user, testValue, 1, { from: highKeeper });
            await increaseTime(86500);
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
    });

    describe('kill', async function () {
        let owner = accounts[0];
        let creeper = accounts[3];

        let token;
        let gatekeeper;

        let startOwnerBalance;
        let startGatekeeperBalance;
        let diff;

        let tx;

        before(async function () {
            token = await SNM.new({ from: owner });
            gatekeeper = await SimpleGatekeeperWithLimit.new(token.address, { from: owner });

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
                gatekeeper = await SimpleGatekeeperWithLimit.new(token.address);
            });

            it('should revert', async function () {
                await assertRevert(gatekeeper.kill({ from: creeper }));
            });
        });
    });
});
