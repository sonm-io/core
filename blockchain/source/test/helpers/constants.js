export const OrderType = {
    UNKNOWN: 0,
    BID: 1,
    ASK: 2,
};

export const IdentityLevel = {
    UNKNOWN: 0,
    ANONIMOUS: 1,
    PSEUDOANONIMOUS: 2,
    IDENTIFIED: 3,
};

export const OrderStatus = {
    UNKNOWN: 0,
    INACTIVE: 1,
    ACTIVE: 2,
};

export const DealStatus = {
    UNKNOWN: 0,
    ACCEPTED: 1,
    CLOSED: 2,
};

export const RequestStatus = {
    REQUEST_UNKNOWN: 0,
    REQUEST_CREATED: 1,
    REQUEST_CANCELED: 2,
    REQUEST_REJECTED: 3,
    REQUEST_ACCEPTED: 4,
};

export const BlackListPerson = {
    NOBODY: 0,
    WORKER: 1,
    MASTER: 2,
};

export const oraclePrice = 1e15;

export const testDuration = 90000;
export const testPrice = 1e3;
export const secInDay = 86400;
export const secInHour = 3600;

export const defaultBenchmarks = [40, 21, 2, 256, 160, 1000, 1000, 6, 3, 1200, 1860000, 3000];
export const benchmarkQuantity = 12;
export const netflagsQuantity = 3;

// response tuples mappings
export const orderInfo = {
    type: 0,
    author: 1,
    counterparty: 2,
    duration: 3,
    price: 4,
    netflags: 5,
    identityLvl: 6,
    blacklist: 7,
    tag: 8,
    benchmarks: 9,
    frozenSum: 10,
};

export const OrderParams = {
    status: 0,
    dealId: 1,
};

export const DealInfo = {
    benchmarks: 0,
    supplier: 1,
    consumer: 2,
    master: 3,
    ask: 4,
    bid: 5,
    startTime: 6,
};

export const DealParams = {
    duration: 0,
    price: 1,
    endTime: 2,
    status: 3,
    blockedBalance: 4,
    totalPayout: 5,
    lastBillTs: 6,
};

export const ChangeRequestInfo = {
    dealID: 0,
    requestType: 1,
    price: 2,
    duration: 3,
    status: 4,
};
