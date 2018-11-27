export async function checkBenchmarks (info, benchmarks) {
    let b = info.map((item) => parseInt(item, 10));
    assert.equal(JSON.stringify(benchmarks), JSON.stringify(b), 'Incorrect benchmarks');
}

export async function checkOrderStatus (orders, key, orderId, status) {
    let res = await orders.GetOrderParams(orderId, { from: key });
    assert.equal(status, res[0], 'Incorrect order status');
}

export async function getDealIdFromOrder (orders, key, orderId) {
    let orderParams = await orders.GetOrderParams(orderId, { from: key });
    return orderParams[1].toNumber(10);
}

export async function getDealInfoFromOrder (deals, orders, key, orderId) {
    let dealId = await getDealIdFromOrder(orders, key, orderId);
    return deals.GetDealInfo(dealId, { from: key });
}
