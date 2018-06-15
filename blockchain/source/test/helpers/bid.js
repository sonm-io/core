import { defaultBenchmarks, IdentityLevel, OrderType, testDuration, testPrice } from './constants';

export async function Bid (
    {
        market, consumer,
        counterparty = '0x0',
        duration = testDuration,
        price = testPrice,
        netFlags = [0, 0, 0],
        identityLvl = IdentityLevel.ANONIMOUS,
        blacklist = 0x0,
        tag = '000000',
        benchmarks = defaultBenchmarks,
    } = {}) {
    let tx = await market.PlaceOrder(
        OrderType.BID,
        counterparty,
        duration,
        price,
        netFlags,
        identityLvl,
        blacklist,
        tag,
        benchmarks,
        { from: consumer });
    let orderId = tx.logs[0].args.orderID;
    assert.isNotNull(orderId);
    return orderId;
}
