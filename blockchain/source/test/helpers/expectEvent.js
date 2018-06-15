const assert = require('chai').assert;

const eventInLogs = (logs, eventName) => {
    const event = logs.find(e => e.event === eventName);
    assert.exists(event, `event ${eventName} wasn't emitted`);
    return event.args;
};

const eventInTransaction = async (tx, eventName) => {
    const { logs } = await tx;
    return eventInLogs(logs, eventName);
};

const allEventsInLogs = (logs, eventName) => {
    let partLogs;
    if (eventName === undefined) {
        partLogs = logs;
    } else {
        partLogs = logs.filter(logs => logs.event === eventName);
    }
    const events = [];
    for (let l of partLogs) {
        events.push({ event: l.event, args: l.args });
    }
    assert.exists(events, `events ${eventName} wasn't emitted`);
    return events;
};

const allEventsInTransaction = async (tx, eventName) => {
    const { logs } = await tx;
    return allEventsInLogs(logs, eventName);
};

module.exports = {
    eventInLogs,
    allEventsInLogs,
    eventInTransaction,
    allEventsInTransaction,
};
