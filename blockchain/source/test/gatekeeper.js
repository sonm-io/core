import sha3 from 'solidity-sha3';

import assertRevert from './helpers/assertRevert';
import MerkleTree from 'merkle-tree-solidity';
import util from 'ethereumjs-util';

const Gatekeeper = artifacts.require('./Gatekeeper.sol');
const SNMTToken = artifacts.require('./SNM.sol');

contract('Gatekeeper', async function (accounts) {
    let gatekeeper;
    let token;

    let owner = accounts[0];
    let user = accounts[1];
    let userWithoutBalance = accounts[6];
    let userWithoutAllowance = accounts[5];
    let creeper = accounts[9];

    before(async function () {
        gatekeeper = await Gatekeeper.deployed();
        token = await SNMTToken.deployed();
        await token.transfer(gatekeeper.address, web3.toWei(100000000), { from: owner });

        await token.transfer(user, 100000000, { from: owner });
        await token.approve(gatekeeper.address, 50000000, { from: user });

        await token.transfer(userWithoutAllowance, web3.toWei(100000000), { from: owner });

        await token.approve(gatekeeper.address, 50000000, { from: userWithoutBalance });
    });

    describe('Defaults', function () {
        it('transaction amount default = 512', async function () {
            let transactionAmount = await gatekeeper.GetCurrentTransactionAmountForBlock.call();
            assert.equal(transactionAmount, 512);
        });
    });

    describe('SetTransactionAmount', function () {
        let tx;
        const testAmount = 4;

        it('should not revert', async function () {
            tx = await gatekeeper.SetTransactionAmountForBlock(testAmount, { from: owner });
        });

        it('CurrentTransactionAmountForBlock should set to new value', async function () {
            let transactionAmount = await gatekeeper.GetCurrentTransactionAmountForBlock.call();
            assert.equal(transactionAmount, testAmount);
        });

        it('should spend `TransactionAmountForBlockChanged` event', function () {
            assert.equal(tx.logs.length, 1);
            assert.equal(tx.logs[0].event, 'TransactionAmountForBlockChanged');
            assert.equal(tx.logs[0].args.amount.toNumber(), testAmount);
        });

        describe('when newTransactionAmount is greeter than processed transactions', function () {
            before(async function () {
                await gatekeeper.PayIn(100, { from: user });
                await gatekeeper.PayIn(100, { from: user });
                await gatekeeper.PayIn(100, { from: user });
            });

            it('should revert', async function () {
                await assertRevert(gatekeeper.SetTransactionAmountForBlock(2, { from: owner }));
            });

            it('not changed', async function () {
                let transactionAmount = await gatekeeper.GetCurrentTransactionAmountForBlock.call();
                assert.equal(transactionAmount.toNumber(), 4);
            });

            after(async function () {
                await gatekeeper.PayIn(100, { from: user });
            });
        });

        describe('when not owner want to change transaction amount', function () {
            it('reverts', async function () {
                await assertRevert(gatekeeper.SetTransactionAmountForBlock(16, { from: creeper }));
            });
        });
    });

    describe('PayIn', function () {
        const testValue = 100;
        let startedBalance;
        let startedTransactionCount;
        let tx;

        before(async function () {
            startedBalance = (await token.balanceOf.call(user)).toNumber();
            startedTransactionCount = (await gatekeeper.GetTransactionCount.call()).toNumber();
        });

        it('transfer token to gatekeeper contract', async function () {
            tx = await gatekeeper.PayIn(testValue, { from: user });
            let balance = (await token.balanceOf.call(user)).toNumber();
            assert.equal(startedBalance - testValue, balance);
        });

        it('should spend `PayInTx` event', function () {
            assert.equal(tx.logs.length, 1);
            assert.equal(tx.logs[0].event, 'PayInTx');
            assert.equal(tx.logs[0].args.from, user);
            assert.equal(tx.logs[0].args.txNumber, 1);
            assert(tx.logs[0].args.value.eq(testValue));
        });

        it('should increase transaction count', async function () {
            let transactionCount = (await gatekeeper.GetTransactionCount.call()).toNumber();
            assert.equal(startedTransactionCount + 1, transactionCount);
        });

        describe('when conditions to emit block reached', function () {
            let tx;

            it('emit block', async function () {
                await gatekeeper.PayIn(testValue, { from: user });
                await gatekeeper.PayIn(testValue, { from: user });
                tx = await gatekeeper.PayIn(testValue, { from: user });
            });

            it('should spend `BlockEmitted` events', function () {
                assert.equal(tx.logs.length, 2);
                assert.equal(tx.logs[1].event, 'BlockEmitted');
            });

            it('should set transaction count to 0', async function () {
                let txCount = await gatekeeper.GetTransactionCount.call();
                assert.equal(txCount.toNumber(), 0);
            });
        });

        describe('when sender doesnt have balance', function () {
            it('should revert', async function () {
                await assertRevert(gatekeeper.PayIn(testValue, { from: userWithoutBalance }));
            });
        });

        describe('when sender doesnt have allowance', function () {
            it('should revert', async function () {
                await assertRevert(gatekeeper.PayIn(testValue, { from: userWithoutAllowance }));
            });
        });
    });

    describe('VoteRoot', function () {
        let tx;
        let startedRootsCount;
        let testRoot = generateTestMerkle(getTx(owner)).root;

        before(async function () {
            startedRootsCount = (await gatekeeper.GetRootsCount.call()).toNumber();
        });

        it('should add new root', async function () {
            tx = await gatekeeper.VoteRoot(stringBufferToHex(testRoot), { from: owner });
        });

        it('should spend `RootAdded` event', async function () {
            assert.equal(tx.logs.length, 1);
            assert.equal(tx.logs[0].event, 'RootAdded');
            assert.equal(tx.logs[0].args.root, stringBufferToHex(testRoot));
        });

        it('should increase root count', async function () {
            let rootsCount = await gatekeeper.GetRootsCount.call();
            assert.equal(startedRootsCount + 1, rootsCount);
        });

        describe('when not owner want to VoteRoot', function () {
            it('should revert', async function () {
                await assertRevert(gatekeeper.VoteRoot(stringBufferToHex(testRoot), { from: creeper }));
            });
        });
    });

    describe('Payout', function () {
        let startedUserBalance;
        let startedGatekeeperBalance;
        let testMerkle = generateTestMerkle(getTx(user));
        let tx;
        let rootId;

        before(async function () {
            let txVoteRoot = await gatekeeper.VoteRoot(stringBufferToHex(testMerkle.root), { from: owner });
            rootId = txVoteRoot.logs[0].args.id;

            startedUserBalance = (await token.balanceOf.call(user)).toNumber();
            startedGatekeeperBalance = (await token.balanceOf.call(gatekeeper.address)).toNumber();
        });

        it('doesnt should revert', async function () {
            const proof = bytesBufferToHex(testMerkle.proof);
            tx = await gatekeeper.Payout(proof, rootId, user, 1, 200, { from: user });
        });

        it('should transfer balance from gatekeeper to user', async function () {
            let userBalance = (await token.balanceOf.call(user)).toNumber();
            assert.equal(startedUserBalance + 200, userBalance);
            let gatekeeperBalance = (await token.balanceOf.call(gatekeeper.address)).toNumber();
            assert.equal(startedGatekeeperBalance - 200, gatekeeperBalance);
        });

        it('should spend `PayoutTx` event', function () {
            assert.equal(tx.logs.length, 1);
            assert.equal(tx.logs[0].event, 'PayoutTx');
            assert.equal(tx.logs[0].args.from, user);
            assert.equal(tx.logs[0].args.txNumber, 1);
            assert.equal(tx.logs[0].args.value, 200);
        });

        it('test other proof', async function () {
            const proof2 = bytesBufferToHex(testMerkle.proof2);
            await gatekeeper.Payout(proof2, rootId, user, 512, 200, { from: user });
            let balance = (await token.balanceOf.call(user)).toNumber();
            assert.equal(balance, startedUserBalance + 400);
        });

        it('test other proof', async function () {
            const proof3 = bytesBufferToHex(testMerkle.proof3);
            await gatekeeper.Payout(proof3, rootId, user, 435, 200, { from: user });
            let balance = (await token.balanceOf.call(user)).toNumber();
            assert.equal(balance, startedUserBalance + 600);
        });
    });
});

function stringBufferToHex (element) {
    return Buffer.isBuffer(element) ? '0x' + element.toString('hex') : element;
}

function bytesBufferToHex (element) {
    return '0x' + element.map(e => e.toString('hex')).join('');
}

function getTx (_from) {
    let el = [];
    for (let i = 1; i <= 512; i++) {
        el.push(util.toBuffer(sha3(_from, i, (2 * 100))));
    }
    return el;
}

function generateTestMerkle (elements) {
    const merkleTree = new MerkleTree(elements);

    // get the merkle root
    // returns 32 byte buffer
    const root = merkleTree.getRoot();

    const proof = merkleTree.getProof(elements[0]);
    const proof2 = merkleTree.getProof(elements[511]);
    const proof3 = merkleTree.getProof(elements[434]);
    // const elem = elements[511];

    // merkleTree.checkProof(proof, root, hash)
    const elem = elements[511];

    return { root, proof, proof2, proof3, elem };
}
