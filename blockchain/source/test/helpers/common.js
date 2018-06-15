export async function checkBenchmarks (info, benchmarks) {
    let b = info.map((item) => parseInt(item, 10));
    assert.equal(JSON.stringify(benchmarks), JSON.stringify(b), 'Incorrect benchmarks');
}

export async function checkOrderStatus (market, key, orderId, status) {
    let res = await market.GetOrderParams(orderId, { from: key });
    assert.equal(status, res[0], 'Incorrect order status');
}

export async function getDealIdFromOrder (market, key, orderId) {
    let orderParams = await market.GetOrderParams(orderId, { from: key });
    return orderParams[1].toNumber(10);
}

export async function getDealInfoFromOrder (market, key, orderId) {
    let dealId = await getDealIdFromOrder(market, key, orderId);
    return market.GetDealInfo(dealId, { from: key });
}
