import increaseTime from './helpers/increaseTime';
import assertRevert from './helpers/assertRevert';

const SNMD = artifacts.require('./SNMD.sol');
const Market = artifacts.require('./Market.sol');
const OracleUSD = artifacts.require('./OracleUSD.sol');
const Blacklist = artifacts.require('./Blacklist.sol');
const ProfileRegistry = artifacts.require('./ProfileRegistry.sol');

const ORDER_TYPE = {
    UNKNOWN: 0,
    BID: 1,
    ASK: 2,
};

const IdentityLevel = {
    UNKNOWN: 0,
    ANONIMOUS: 1,
    PSEUDOANONIMOUS: 2,
    IDENTIFIED: 3,
};

contract('Market', async function (accounts) {
    let market;
    let token;
    let oracle;
    let blacklist;
    let profileRegistry;
    let supplier = accounts[1];
    let consumer = accounts[2];
    let master = accounts[3];
    let specialConsumer = accounts[4];
    let specialConsumer2 = accounts[5];
    let blacklistedSupplier = accounts[6];

    let benchmarks = [40, 21, 2, 256, 160, 1000, 1000, 6, 3, 1200, 1860000, 3000];
    let testPrice = 1e15;
    let testDuration = 90000;

    before(async function () {
        token = await SNMD.new();
        oracle = await OracleUSD.new();
        await oracle.setCurrentPrice(1e18);
        blacklist = await Blacklist.new();
        profileRegistry = await ProfileRegistry.new();
        market = await Market.new(token.address, blacklist.address, oracle.address, profileRegistry.address, 12);
        await token.AddMarket(market.address);
        await blacklist.SetMarketAddress(market.address);

        await token.transfer(consumer, 1000000 * 1e18, { from: accounts[0] });
        await token.transfer(supplier, 1000000 * 1e18, { from: accounts[0] });
        await token.transfer(specialConsumer, 3600, { from: accounts[0] });
        await token.transfer(specialConsumer2, 3605, { from: accounts[0] });
    });

    it('test balances', async function () {
        await token.balanceOf.call(supplier);
        await token.balanceOf.call(consumer);
    });

    it('test CreateOrder forward ask', async function () {
        let stateBefore = await market.GetOrdersAmount();
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            testDuration, // duration
            testPrice, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: supplier });
        let stateAfter = await market.GetOrdersAmount();
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));
        // TODO: get this order and match params
    });

    it('test CreateOrder forward bid', async function () {
        // TODO: test above normal deal
        let stateBefore = await market.GetOrdersAmount();
        let balanceBefore = await token.balanceOf(consumer);
        await market.PlaceOrder(
            ORDER_TYPE.BID, // type
            '0x0', // counter_party
            testDuration, // duration
            testPrice, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: consumer });

        let stateAfter = await market.GetOrdersAmount();
        let balanceAfter = await token.balanceOf(consumer);
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));
        assert.equal(balanceBefore.toNumber(10) - 86400 * testPrice, balanceAfter.toNumber(10));
        // TODO: get this order and match params
        // TODO: get balance and check block for 1 day

        // TODO: test above normal deal - VAR#2 - for deal-duration
        // TODO: get this order and match params
        // TODO: get balance and check block for deal duration

        // TODO: test above spot deal
        // TODO: get this order and match params
        // TODO: get balance and check block for 1 hour
    });

    it('test CreateOrder spot ask', async function () {
        let stateBefore = await market.GetOrdersAmount();
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            0, // duration
            testPrice, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: supplier });
        let stateAfter = await market.GetOrdersAmount();
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));
        // TODO: get this order and match params
    });

    it(' test makret balance', async function () {
        await token.balanceOf(market.address);
    });

    it('test CreateOrder spot bid', async function () {
        let stateBefore = await market.GetOrdersAmount();
        let balanceBefore = await token.balanceOf(consumer);
        // TODO: test above normal deal
        await market.PlaceOrder(
            ORDER_TYPE.BID, // type
            '0x0', // counter_party
            0, // duration
            testPrice, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: consumer });
        let stateAfter = await market.GetOrdersAmount();
        let balanceAfter = await token.balanceOf(consumer);
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));
        assert.equal(balanceBefore.toNumber(10) - 3600 * testPrice, balanceAfter.toNumber(10));
    });

    it('test CancelOrder: cancel ask order', async function () {
        // TODO: success case - cancel ask order
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            testDuration, // duration
            testPrice, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: supplier });
        await market.CancelOrder(5, { from: supplier });

        // TODO: verify order = inactive
        // TODO: verify token not(!) transfered
        // TODO: verify catch the event - not properly

        // TODO: verify order = inactive
        // TODO: verify token transfered
        // TODO: verify catch the event - not properly
    });

    it('test CancelOrder: cancel bid order', async function () {
        // TODO: success case - cancel bid order
        await market.PlaceOrder(
            ORDER_TYPE.BID, // type
            '0x0', // counter_party
            testDuration, // duration
            testPrice, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: consumer });
        await market.CancelOrder(6, { from: consumer });
    });

    it('test CancelOrder: error case 1', async function () {
        // error case - order not active
        await assertRevert(market.CancelOrder(6, { from: supplier }));
    });

    it('test CancelOrder: error case 2', async function () {
        // error case - foreign address of sender
        await market.PlaceOrder(
            ORDER_TYPE.BID, // type
            '0x0', // counter_party
            testDuration, // duration
            testPrice, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: supplier });
        await assertRevert(market.CancelOrder(7, { from: consumer }));
    });

    it('test CancelOrder: error case 3', async function () {
        // error case - order does not exists
        await assertRevert(market.CancelOrder(100500, { from: supplier }));
    });

    it('test OpenDeal forward', async function () {
        let stateBefore = await market.GetDealsAmount();
        await market.OpenDeal(1, 2, { from: consumer });
        let stateAfter = await market.GetDealsAmount();
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));
    });

    it('test OpenDeal: spot', async function () {
        let stateBefore = await market.GetDealsAmount();
        await market.OpenDeal(3, 4, { from: master });
        let stateAfter = await market.GetDealsAmount();
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));
    });

    it('test OpenDeal:closing after ending', async function () {
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            1, // duration
            1, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: supplier });

        await market.PlaceOrder(
            ORDER_TYPE.BID, // type
            '0x0', // counter_party
            1, // duration
            1, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: consumer });
        let ordersAmount = await market.GetOrdersAmount();
        ordersAmount = ordersAmount.toNumber(10);
        let stateBefore = await market.GetDealsAmount();
        await market.OpenDeal(ordersAmount - 1, ordersAmount, { from: consumer });
        let stateAfter = await market.GetDealsAmount();
        await increaseTime(2);
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));
    });

    it('test BlockedBalance: forward', async function () {
        let rate = (await oracle.getCurrentPrice()).toNumber(10);
        let shouldBlocked = testPrice * 86400 * rate / 1e18;
        let deal = await market.GetDealParams(1);
        let nowBlocked = deal[4].toNumber(10); // we do not cast getters as as struct
        assert.equal(shouldBlocked, nowBlocked);
    });

    it('test USD price changing ', async function () {
        await oracle.setCurrentPrice(2000000000000, { from: accounts[0] });
    });

    it('test RegisterWorker', async function () {
        await market.RegisterWorker(master, { from: supplier });
    });

    it('test ConfirmWorker', async function () {
        let stateBefore = await market.GetMaster(supplier);
        await market.ConfirmWorker(supplier, { from: master });
        let stateAfter = await market.GetMaster(supplier);
        assert.ok(stateBefore !== stateAfter && stateAfter === master);
    });

    it('test RemoveWorker', async function () {
        await market.RemoveWorker(supplier, master, { from: master });
        let check = await market.GetMaster(supplier);
        assert.equal(check, supplier);
    });

    it('test GetDealInfo', async function () {
        await market.GetDealInfo(2);
    });

    it('test GetDealParams', async function () {
        await market.GetDealParams(1);
    });

    it('test GetOrderInfo', async function () {
        await market.GetOrderInfo(2);
    });

    it('test GetOrderParams', async function () {
        await market.GetOrderParams(2);
    });

    it('test Bill: spot', async function () {
        let deal = await market.GetDealParams(2);
        let lastBillTS = deal[6];
        await increaseTime(2);
        await market.Bill(2, { from: supplier });
        let dealAfter = await market.GetDealParams(2);
        let lastBillAfter = dealAfter[6];
        assert.ok(lastBillTS.toNumber(10) < lastBillAfter.toNumber(10));
    });

    it('test Bill: forward 1', async function () {
        let deal = await market.GetDealParams(1);
        let lastBillTS = deal[6];
        await increaseTime(2);
        await market.Bill(1, { from: supplier });
        let dealAfter = await market.GetDealParams(1);
        let lastBillAfter = dealAfter[6];
        assert.ok(lastBillTS.toNumber(10) < lastBillAfter.toNumber(10));
        assert.equal(dealAfter[3].toNumber(10), 1);
    });

    it('test Bill: forward 2', async function () {
        let deal = await market.GetDealParams(1);
        let lastBillTS = deal[6];
        await increaseTime(2);
        await market.Bill(1, { from: supplier });
        let dealAfter = await market.GetDealParams(1);
        let lastBillAfter = dealAfter[6];
        assert.ok(lastBillTS.toNumber(10) < lastBillAfter.toNumber(10));
        assert.equal(dealAfter[3].toNumber(10), 1);
    });

    it('test CreateChangeRequest for cancel: ask', async function () {
        let stateBefore = await market.GetChangeRequestsAmount();
        await market.CreateChangeRequest(1, 3, 100000, { from: supplier });
        let stateAfter = await market.GetChangeRequestsAmount();
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));
    });

    it('test CancelChangeRequest after creation: ask', async function () {
        let lastCR = (await market.GetChangeRequestsAmount()).toNumber(10);
        let stateBefore = await market.GetChangeRequestInfo(lastCR);
        await market.CancelChangeRequest(lastCR, { from: supplier });
        let stateAfter = await market.GetChangeRequestInfo(lastCR);
        assert.ok(stateBefore[4] !== stateAfter[4]);
    });

    it('test CreateChangeRequest for cancel: bid', async function () {
        let stateBefore = await market.GetChangeRequestsAmount();
        await market.CreateChangeRequest(1, 3, 100000, { from: consumer });
        let stateAfter = await market.GetChangeRequestsAmount();
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));
    });

    it('test CancelChangeRequest after creation: bid', async function () {
        let lastCR = (await market.GetChangeRequestsAmount()).toNumber(10);
        let stateBefore = await market.GetChangeRequestInfo(lastCR);
        await market.CancelChangeRequest(lastCR, { from: consumer });
        let stateAfter = await market.GetChangeRequestInfo(lastCR);
        assert.ok(stateBefore[4] !== stateAfter[4]);
    });

    it('test CreateChangeRequest for rejecting: bid', async function () {
        let stateBefore = await market.GetChangeRequestsAmount();
        await market.CreateChangeRequest(3, 3, 100000, { from: consumer });
        let stateAfter = await market.GetChangeRequestsAmount();
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));
    });

    it('test CancelChangeRequest rejecting: ask', async function () {
        let lastCR = await market.GetChangeRequestsAmount();
        let stateBefore = await market.GetChangeRequestInfo(lastCR.toNumber(10));
        await market.CancelChangeRequest(lastCR.toNumber(10), { from: supplier });
        let stateAfter = await market.GetChangeRequestInfo(lastCR.toNumber(10));
        assert.ok(stateBefore[4] !== stateAfter[4]);
    });

    it('test CreateChangeRequest for rejecting: ask', async function () {
        let stateBefore = await market.GetChangeRequestsAmount();
        await market.CreateChangeRequest(3, 3, 100000, { from: supplier });
        let stateAfter = await market.GetChangeRequestsAmount();
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));
    });

    it('test CancelChangeRequest rejecting: bid', async function () {
        let lastCR = await market.GetChangeRequestsAmount();
        let stateBefore = await market.GetChangeRequestInfo(lastCR.toNumber(10));
        await market.CancelChangeRequest(lastCR.toNumber(10), { from: consumer });
        let stateAfter = await market.GetChangeRequestInfo(lastCR.toNumber(10));
        assert.ok(stateBefore[4] !== stateAfter[4]);
    });

    it('test CreateChangeRequest for spot deal: ask', async function () {
        await market.CreateChangeRequest(2, 2, 0, { from: supplier });
    });

    it('test CreateChangeRequest for spot deal: bid', async function () {
        await increaseTime(2);
        await market.CreateChangeRequest(2, 2, 0, { from: consumer });
        let newState = await market.GetDealParams(2);
        let newPrice = newState[1].toNumber(10);
        assert.equal(2, newPrice);
    });

    it('test CreateChangeRequest for forward deal: bid', async function () {
        await market.CreateChangeRequest(1, 2, 2999, { from: consumer });
    });

    it('test CreateChangeRequest for forward deal: ask', async function () {
        await increaseTime(2);
        await market.CreateChangeRequest(1, 2, 3000, { from: supplier });
        let newState = await market.GetDealParams(1);
        let newPrice = newState[1].toNumber(10);
        let newDuration = newState[0].toNumber(10);
        assert.equal(2, newPrice);
        assert.equal(2999, newDuration);
    });

    it('test CreateChangeRequest for forward deal: automatch bid', async function () {
        let stateBefore = await market.GetDealParams(1);
        let oldPrice = stateBefore[1].toNumber(10);
        let oldDuration = stateBefore[0].toNumber(10);
        await market.CreateChangeRequest(1, 3, oldDuration, { from: consumer });
        let newState = await market.GetDealParams(1);
        let newPrice = newState[1].toNumber(10);
        let newDuration = newState[0].toNumber(10);
        assert.ok(oldPrice < newPrice);
        assert.equal(oldDuration, newDuration);
    });

    it('test re-OpenDeal forward: close it with blacklist', async function () {
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            5, // duration
            1, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: supplier });

        await market.PlaceOrder(
            ORDER_TYPE.BID, // type
            '0x0', // counter_party
            4, // duration
            1, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: consumer });
        let ordersAmount = await market.GetOrdersAmount();
        ordersAmount = ordersAmount.toNumber(10);
        let stateBefore = await market.GetDealsAmount();
        await market.OpenDeal(ordersAmount - 1, ordersAmount, { from: consumer });
        let stateAfter = await market.GetDealsAmount();
        await increaseTime(2);
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));
    });

    it('test CloseDeal spot: close it when paid amount > blocked balance, not enough money', async function () {
        await oracle.setCurrentPrice(1e12);
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            0, // duration
            1e6, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: supplier });

        await market.PlaceOrder(
            ORDER_TYPE.BID, // type
            '0x0', // counter_party
            0, // duration
            1e6, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: specialConsumer });
        let ordersAmount = await market.GetOrdersAmount();
        ordersAmount = ordersAmount.toNumber(10);
        let stateBefore = await market.GetDealsAmount();
        await market.OpenDeal(ordersAmount - 1, ordersAmount, { from: specialConsumer });
        let dealId = (await market.GetDealsAmount()).toNumber(10);

        let stateAfterOpen = await market.GetDealParams(dealId);
        let stateAfter = await market.GetDealsAmount();
        await increaseTime(2);

        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));

        await oracle.setCurrentPrice(1e18);
        await market.CloseDeal(dealId, 0, { from: specialConsumer });
        let stateAfterClose = await market.GetDealParams(dealId);
        assert.equal(stateAfterClose[3].toNumber(10), 2);
        assert.equal(stateAfterClose[4].toNumber(10), 0); // balance
        assert.equal(stateAfterClose[5].toNumber(10), 3600); // payout
        assert.equal(stateAfterOpen[3].toNumber(10), 1);
        assert.equal(stateAfterOpen[4].toNumber(10), 3600); // balance
        assert.equal(stateAfterOpen[5].toNumber(10), 0); // payout
    });

    it('test CloseDeal spot: close it when paid amount > blocked balance', async function () {
        await oracle.setCurrentPrice(1e12);
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            0, // duration
            1e6, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: supplier });

        await market.PlaceOrder(
            ORDER_TYPE.BID, // type
            '0x0', // counter_party
            0, // duration
            1e6, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: consumer });
        let ordersAmount = await market.GetOrdersAmount();
        ordersAmount = ordersAmount.toNumber(10);
        let stateBefore = await market.GetDealsAmount();
        await market.OpenDeal(ordersAmount - 1, ordersAmount, { from: consumer });
        let dealId = (await market.GetDealsAmount()).toNumber(10);

        let stateAfter = await market.GetDealsAmount();
        let stateAfterOpen = await market.GetDealParams(dealId);
        await increaseTime(2);
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));

        await oracle.setCurrentPrice(1e18);

        await market.CloseDeal(dealId, 0, { from: consumer });
        let stateAfterClose = await market.GetDealParams(dealId);
        let infoAfterClose = await market.GetDealInfo(dealId);
        let dealTime = stateAfterClose[2].toNumber(10) - infoAfterClose[6].toNumber(10);
        assert.equal(stateAfterClose[3].toNumber(10), 2);
        assert.equal(stateAfterClose[4].toNumber(10), 3600000000); // balance
        assert.equal(stateAfterClose[5].toNumber(10), dealTime * 1e18 * 1e6 / 1e18); // payout
        assert.equal(stateAfterOpen[3].toNumber(10), 1);
        assert.equal(stateAfterOpen[4].toNumber(10), 3600); // balance
        assert.equal(stateAfterOpen[5].toNumber(10), 0); // payout
    });

    it('test create deal from spot BID and forward ASK and close with blacklist', async function () {
        await oracle.setCurrentPrice(1e12);
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            36000, // duration
            1e6, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: blacklistedSupplier });

        await market.PlaceOrder(
            ORDER_TYPE.BID, // type
            '0x0', // counter_party
            0, // duration
            1e6, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: consumer });
        let ordersAmount = await market.GetOrdersAmount();
        ordersAmount = ordersAmount.toNumber(10);
        let stateBefore = await market.GetDealsAmount();
        await market.OpenDeal(ordersAmount - 1, ordersAmount, { from: consumer });
        let dealId = (await market.GetDealsAmount()).toNumber(10);

        let stateAfter = await market.GetDealsAmount();
        await market.GetDealParams(dealId);
        await increaseTime(2);
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));

        await oracle.setCurrentPrice(1e18);

        await market.CloseDeal(dealId, 2, { from: consumer });
        let blacklisted = await blacklist.Check(consumer, blacklistedSupplier);
        assert.ok(blacklisted);
    });

    it('test create deal from spot BID and forward ASK with blacklisted supplier', async function () {
        await oracle.setCurrentPrice(1e12);
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            36000, // duration
            1e6, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: blacklistedSupplier });

        await market.PlaceOrder(
            ORDER_TYPE.BID, // type
            '0x0', // counter_party
            0, // duration
            1e6, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: consumer });
        let ordersAmount = await market.GetOrdersAmount();
        ordersAmount = ordersAmount.toNumber(10);
        await assertRevert(market.OpenDeal(ordersAmount - 1, ordersAmount, { from: consumer }));
    });

    it('test CreateChangeRequest for forward deal: fullcheck ask', async function () {
        let stateBefore = await market.GetDealParams(4);
        let oldPrice = stateBefore[1].toNumber(10);
        let oldDuration = stateBefore[0].toNumber(10);
        await market.CreateChangeRequest(4, 2, 100, { from: supplier });
        let newState = await market.GetDealParams(4);
        let newPrice = newState[1].toNumber(10);
        let newDuration = newState[0].toNumber(10);
        assert.ok(newPrice === oldPrice);
        assert.ok(newDuration === oldDuration);
        let a = (await market.GetChangeRequestsAmount()).toNumber(10);
        await market.GetChangeRequestInfo(a);
    });

    it('test CreateChangeRequest for forward deal: fullcheck bid', async function () {
        await increaseTime(1);
        let stateBefore = await market.GetDealParams(4);
        let oldPrice = stateBefore[1].toNumber(10);
        let oldDuration = stateBefore[0].toNumber(10);
        await market.CreateChangeRequest(4, 3, 99, { from: consumer });
        let newState = await market.GetDealParams(4);
        let newPrice = newState[1].toNumber(10);
        let newDuration = newState[0].toNumber(10);
        assert.ok(newPrice > oldPrice);
        assert.ok(newDuration > oldDuration);
    });

    it('test Bill delayed I: spot', async function () {
        await increaseTime(2);
        await market.Bill(2, { from: supplier });
    });

    it('test Bill delayed II: spot', async function () {
        await increaseTime(2);
        await market.Bill(2, { from: supplier });
    });

    it('test CloseDeal: spot w/o blacklist', async function () {
        await market.CloseDeal(2, 0, { from: supplier });
        let stateAfter = await market.GetDealParams(2);
        assert.equal(stateAfter[3].toNumber(10), 2);
    });

    it('test CloseDeal: closing after ending', async function () {
        await market.CloseDeal(3, 0, { from: supplier });
        let stateAfter = await market.GetDealParams(3);
        assert.equal(stateAfter[3].toNumber(10), 2);
    });

    it('test CloseDeal: forward w blacklist', async function () {
        await market.CloseDeal(4, 2, { from: consumer });
        let stateAfter = await market.GetDealParams(4);
        assert.equal(stateAfter[3].toNumber(10), 2);
    });

    it('test Bill spot: close it when next period sum > consumer balance', async function () {
        await oracle.setCurrentPrice(1e12);
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            0, // duration
            1e6, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: supplier });

        await market.PlaceOrder(
            ORDER_TYPE.BID, // type
            '0x0', // counter_party
            0, // duration
            1e6, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: specialConsumer2 });
        let ordersAmount = await market.GetOrdersAmount();
        ordersAmount = ordersAmount.toNumber(10);
        let stateBefore = await market.GetDealsAmount();
        await market.OpenDeal(ordersAmount - 1, ordersAmount, { from: specialConsumer2 });
        let dealId = (await market.GetDealsAmount()).toNumber(10);

        let stateAfter = await market.GetDealsAmount();
        await market.GetDealParams(dealId);
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));

        increaseTime(3600);
        await market.Bill(dealId, { from: specialConsumer2 });
        let stateAfterClose = await market.GetDealParams(dealId);
        await market.GetDealInfo(dealId);
        assert.equal(stateAfterClose[3].toNumber(10), 2);
        await oracle.setCurrentPrice(1e18);
    });

    it('test Set new blacklist', async function () {
        let newBl = await Blacklist.new();
        await market.SetBlacklistAddress(newBl.address);
        await newBl.SetMarketAddress(market.address);
        blacklist = newBl;
    });

    it('test Set new pr', async function () {
        let newPR = await ProfileRegistry.new();
        await market.SetProfileRegistryAddress(newPR.address);
    });

    it('test Set new oracle', async function () {
        let newOracle = await OracleUSD.new();
        await newOracle.setCurrentPrice(20000000000000);
        await market.SetOracleAddress(newOracle.address);
    });

    it('test QuickBuy', async function () {
        let stateBefore = await market.GetOrdersAmount();
        let dealsBefore = await market.GetDealsAmount();
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            testDuration, // duration
            testPrice, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: supplier });
        let stateAfter = await market.GetOrdersAmount();
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));

        await market.QuickBuy(stateAfter, 10, { from: consumer });
        let dealsAfter = await market.GetDealsAmount();
        assert.equal(dealsBefore.toNumber(10) + 1, dealsAfter.toNumber(10));
    });

    it('test QuickBuy w master', async function () {
        await market.RegisterWorker(master, { from: supplier });
        await market.ConfirmWorker(supplier, { from: master });
        let stateBefore = await market.GetOrdersAmount();
        let dealsBefore = await market.GetDealsAmount();
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            testDuration, // duration
            testPrice, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: supplier });
        let stateAfter = await market.GetOrdersAmount();
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));

        await market.QuickBuy(stateAfter, 10, { from: consumer });
        let dealsAfter = await market.GetDealsAmount();
        assert.equal(dealsBefore.toNumber(10) + 1, dealsAfter.toNumber(10));
    });

    it('test re-OpenDeal forward: close it with blacklist', async function () {
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            3600, // duration
            10, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: supplier });

        await market.PlaceOrder(
            ORDER_TYPE.BID, // type
            '0x0', // counter_party
            3600, // duration
            10, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: consumer });
        let ordersAmount = await market.GetOrdersAmount();
        ordersAmount = ordersAmount.toNumber(10);
        await market.GetDealsAmount();
        await market.OpenDeal(ordersAmount - 1, ordersAmount, { from: supplier });
        let stateAfter = await market.GetDealsAmount();
        await market.CloseDeal(stateAfter.toNumber(10), 1, { from: consumer });
        await blacklist.Remove(supplier, { from: consumer });
    });

    it('test CloseDeal w blacklist from supplier (assert revert)', async function () {
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            3600, // duration
            10, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: supplier });

        let ordersAmount = await market.GetOrdersAmount();
        ordersAmount = ordersAmount.toNumber(10);
        await market.GetDealsAmount();
        await market.QuickBuy(ordersAmount, 0, { from: consumer });
        let stateAfter = await market.GetDealsAmount();
        await assertRevert(market.CloseDeal(stateAfter.toNumber(10), 1, { from: supplier }));
    });

    it('test OpenDeal forward: close it after ending', async function () {
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            3600, // duration
            10, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: consumer });

        await market.PlaceOrder(
            ORDER_TYPE.BID, // type
            '0x0', // counter_party
            3600, // duration
            10, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: consumer });
        let ordersAmount = await market.GetOrdersAmount();
        ordersAmount = ordersAmount.toNumber(10);
        await market.GetDealsAmount();
        await market.OpenDeal(ordersAmount - 1, ordersAmount, { from: consumer });
        let stateAfter = await market.GetDealsAmount();
        await increaseTime(2);
        await market.CloseDeal(stateAfter.toNumber(10), 0, { from: consumer });
    });

    it('test CreateOrder forward bid duration 2 hours', async function () {
        await oracle.setCurrentPrice(1e18, { from: accounts[0] });
        let stateBefore = await market.GetOrdersAmount();
        let balanceBefore = await token.balanceOf(consumer);
        await market.PlaceOrder(
            ORDER_TYPE.BID, // type
            '0x0', // counter_party
            7200, // duration
            1e5, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            [88, 222], // benchmarks
            { from: consumer });

        let stateAfter = await market.GetOrdersAmount();
        let balanceAfter = await token.balanceOf(consumer);
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));
        assert.equal(balanceBefore.toNumber(10) - 7200 * 1e3, balanceAfter.toNumber(10));
    });

    it('test UpdateBenchmarks', async function () {
        await market.SetBenchmarksQuantity(20);
        assert.equal((await market.GetBenchmarksQuantity()).toNumber(10), 20);
    });

    it('test CreateOrder with num benchmarks < current benchmarks', async function () {
        let stateBefore = await market.GetOrdersAmount();
        let dealsBefore = await market.GetDealsAmount();
        await market.PlaceOrder(
            ORDER_TYPE.ASK, // type
            '0x0', // counter_party
            3600, // duration
            1, // price
            [0, 0, 0], // netflags
            IdentityLevel.ANONIMOUS, // identity level
            0x0, // blacklist
            '00000', // tag
            benchmarks, // benchmarks
            { from: supplier });

        let stateAfter = await market.GetOrdersAmount();
        assert.equal(stateBefore.toNumber(10) + 1, stateAfter.toNumber(10));
        await market.QuickBuy(stateAfter, 10, { from: consumer });
        let dealsAfter = await market.GetDealsAmount();
        assert.equal(dealsBefore.toNumber(10) + 1, dealsAfter.toNumber(10));
        let deal = await market.GetDealInfo(dealsAfter.toNumber(10));
        let dealBenchmarks = deal[0];
        assert.equal(dealBenchmarks.length, (await market.GetBenchmarksQuantity()).toNumber(10));
    });

    it('test SetProfileRegistryAddress: bug while we can cast any contract as valid (for example i cast token as a Profile Registry)', async function () { // eslint-disable-line max-len
        await market.SetProfileRegistryAddress(token.address);
    });
});
