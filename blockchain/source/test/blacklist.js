import assertRevert from './helpers/assertRevert';

const Market = artifacts.require('./Market.sol');
const Blacklist = artifacts.require('./Blacklist.sol');
const SNM = artifacts.require('./SNM.sol');
const OracleUSD = artifacts.require('./OracleUSD.sol');
const ProfileRegistry = artifacts.require('./ProfileRegistry.sol');

contract('Blacklist', async function (accounts) {
    let market;
    let blacklist;
    let token;
    let oracle;
    let pr;

    const owner = accounts[0];
    const creeper = accounts[1];
    const testMaster = accounts[2];
    const testBlacklister = accounts[3];
    const testBlacklisted = accounts[3];
    const master = accounts[5];

    before(async function () {
        token = await SNM.new();
        blacklist = await Blacklist.new();
        oracle = await OracleUSD.new();
        pr = await ProfileRegistry.new();
        await blacklist.AddMaster(master, { from: owner });
        market = await Market.new(token.address, blacklist.address, oracle.address, pr.address, 12, 3);
    });

    it('test ACL', async function () {
        await assertRevert(blacklist.AddMaster('0x0', { from: creeper }));

        await blacklist.AddMaster('0x0', { from: owner });

        await assertRevert(blacklist.RemoveMaster('0x0', { from: creeper }));

        await blacklist.RemoveMaster('0x0', { from: owner });

        await assertRevert(blacklist.SetMarketAddress('0x0', { from: creeper }));

        await blacklist.SetMarketAddress('0x0', { from: owner });
    });

    it('test ACL - OnlyMarket', async function () {
        // TODO: implement this after market is done!
        console.warn('TODO: implement this after market is done!');

        await assertRevert(blacklist.Add(testBlacklister, testBlacklisted, { from: creeper }));

        await blacklist.SetMarketAddress(master, { from: owner });

        await blacklist.Add(testBlacklister, testBlacklisted, { from: master });
    });

    it('test SetMarketAddress', async function () {
        await blacklist.SetMarketAddress(market.address, { from: owner });

        let marketAddressInBlacklist = await blacklist.market.call();
        assert.equal(marketAddressInBlacklist, market.address);
    });

    it('test Add', async function () {
        let marketAddressInBlacklist = await blacklist.market.call();

        assert.notEqual(marketAddressInBlacklist, '0x0');
        await blacklist.Add(testBlacklister, testBlacklisted, { from: master });

        let check = await blacklist.Check(testBlacklister, testBlacklisted);
        assert(check);
    });

    it('test Remove', async function () {
        let marketAddressInBlacklist = await blacklist.market.call();
        assert.notEqual(marketAddressInBlacklist, '0x0');

        await blacklist.Remove(testBlacklisted, { from: testBlacklister });

        let check = await blacklist.Check(testBlacklister, testBlacklisted);
        assert(!check);

        await assertRevert(blacklist.Remove(testBlacklisted, { from: testBlacklister }));
    });

    it('test AddMaster', async function () {
        await blacklist.AddMaster(testMaster, { from: owner });

        await assertRevert(blacklist.AddMaster(testMaster, { from: owner }));
    });

    it('test RemoveMaster', async function () {
        await blacklist.RemoveMaster(testMaster, { from: owner });
        await assertRevert(blacklist.RemoveMaster(testMaster, { from: owner }));
    });
});
