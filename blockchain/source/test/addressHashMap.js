const AddressHashMap = artifacts.require('./AddressHashMap.sol');

contract('AddressHashMap', async function (accounts) {
    let hm;
    let owner = accounts[0];

    before(async function () {
        hm = await AddressHashMap.new({ from: owner });
    });

    it('should write and read', async function () {
        await hm.write('market', accounts[2], { from: owner });
        await hm.write('market', accounts[3], { from: owner });
        await hm.write('profiles', accounts[4], { from: owner });

        let market = await hm.read('market');
        assert.equal(market, accounts[3]);

        let pr = await hm.read('profiles');
        assert.equal(pr, accounts[4]);
    });
});
