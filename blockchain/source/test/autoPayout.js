import assertRevert from './helpers/assertRevert';

let AutoPayout = artifacts.require('./AutoPayout.sol');
let SimpleGatekeeperWithLimit = artifacts.require('./SimpleGatekeeperWithLimit.sol');
let SNM = artifacts.require('./SNM.sol');

contract('AutoPayout', (accounts) => {
    let token, gatekeeper, autopayout;

    let owner = accounts[0];
    let master = accounts[2];
    let target = accounts[3];
    let creeper = accounts[7];

    const oneMinute = 60;
    const testValue = 100;

    before(async () => {
        token = await SNM.new({ from: owner });

        gatekeeper = await SimpleGatekeeperWithLimit.new(token.address, oneMinute, { from: owner });

        autopayout = await AutoPayout.new(token.address, gatekeeper.address, { from: owner });
    });

    describe('SetAutoPayout', () => {
        let tx;

        it('should set payout limit', async () => {
            tx = await autopayout.SetAutoPayout(testValue, target, { from: master });
        });

        it('should spend `AutoPayoutChanged` event', async () => {
            assert.equal(tx.logs.length, 1);
            assert.equal(tx.logs[0].event, 'AutoPayoutChanged');
            assert.equal(tx.logs[0].args.master, master);
            assert.equal(tx.logs[0].args.target, target);
            assert.equal(tx.logs[0].args.limit, testValue);
        });
    });

    describe('DoAutoPayout', () => {
        let tx;

        before(async () => {
            await autopayout.SetAutoPayout(testValue, target, { from: master });

            // Send tokens for 100 test iterations
            await token.transfer(master, testValue * 2, { from: owner });
            await token.approve(autopayout.address, testValue * 1000000, { from: master });
        });

        it('should payout and transfer tokens to gate', async () => {
            let startGateBalance = (await token.balanceOf(gatekeeper.address)).toNumber();
            let startMasterBalance = (await token.balanceOf(master)).toNumber();

            tx = await autopayout.DoAutoPayout(master, { from: creeper });

            let endGateBalance = (await token.balanceOf(gatekeeper.address)).toNumber();
            let endMasterBalance = (await token.balanceOf(master)).toNumber();

            assert.equal(endMasterBalance, 0);
            assert.equal(startGateBalance + startMasterBalance, endGateBalance);
        });

        it('should spend `AutoPayout` event', async () => {
            assert.equal(tx.logs.length, 1);
            assert.equal(tx.logs[0].event, 'AutoPayout');
            assert.equal(tx.logs[0].args.master, master);
        });

        describe('when balance lower than limit', () => {
            it('should revert', async () => {
                // balance of master at this moment is equal zero
                await assertRevert(autopayout.DoAutoPayout(master, { from: creeper }));
            });
        });

        describe('when allowance lower than limit', () => {
            before(async () => {
                await token.transfer(master, testValue * 2, { from: owner });
                await token.approve(autopayout.address, 0, { from: master });
            });

            it('should revert', async () => {
                await assertRevert(autopayout.DoAutoPayout(creeper, { from: creeper }));
            });
        });
    });

    describe('kill', async () => {
        let startOwnerBalance;
        let startAutoPayoutBalance;
        let diff;

        let tx;

        before(async () => {
            autopayout = await AutoPayout.new(token.address, gatekeeper.address, { from: owner });

            // Accidentally transfer some tokens to AutoPayout address
            await token.transfer(autopayout.address, testValue, { from: owner });
            startOwnerBalance = await token.balanceOf(owner);
            startAutoPayoutBalance = await token.balanceOf(autopayout.address);
        });

        it('should killed', async () => {
            tx = await autopayout.kill({ from: owner });
        });

        it('should transfer contract balance to owner', async () => {
            let endOwnerBalance = await token.balanceOf(owner);
            let endAutoPayoutBalance = await token.balanceOf(autopayout.address);

            assert.equal(endAutoPayoutBalance.toNumber(), 0);
            diff = startOwnerBalance.toNumber() - startAutoPayoutBalance.toNumber();
            assert.equal(endOwnerBalance.toNumber() - endAutoPayoutBalance.toNumber(), diff);
        });

        it('should spend `Suicide` event', () => {
            assert.equal(tx.logs.length, 1);
            assert.equal(tx.logs[0].event, 'Suicide');
        });

        describe('when not owner want to kill contract', () => {
            before(async () => {
                token = await SNM.new();
                autopayout = await AutoPayout.new(token.address, gatekeeper.address);
            });

            it('should revert', async () => {
                await assertRevert(autopayout.kill({ from: creeper }));
            });
        });
    });
});
