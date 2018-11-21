import assertRevert from './helpers/assertRevert';
import increaseTime from './helpers/increaseTime';

let DevicesStorage = artifacts.require('./DevicesStorage.sol');

contract('DevicesStora1ge', (accounts) => {
    let devicesStorage;
    let worker = accounts[1];

    let devices = '0x1337';
    let hash = '0x2636a8beb2c41b8ccafa9a55a5a5e333892a83b491df3a67d2768946a9f9c6dc';

    before(async () => {
        devicesStorage = await DevicesStorage.new();
    });

    describe('SetDevices', () => {
        it('should set devices', async () => {
            await devicesStorage.SetDevices(devices, { from: worker });
        });

        it('should update devices by hash', async () => {
            await increaseTime(86400);
            await devicesStorage.UpdateDevicesByHash(hash, { from: worker });
            increaseTime(86000);
            let returnedDevices = await devicesStorage.GetDevices(worker);
            assert.equal(returnedDevices, devices);
        });

        it('should do not update devices by hash, revert', async () => {
            await increaseTime(1);
            await assertRevert(devicesStorage.UpdateDevicesByHash('0x1', { from: worker }));
        });

        it('should return Devices same as written', async () => {
            let returnedDevices = await devicesStorage.GetDevices(worker);
            assert.equal(returnedDevices, devices);
        });

        it('should return empty devices due ttl ending', async () => {
            await increaseTime(10000000);
            let returnedDevices = await devicesStorage.GetDevices(worker);
            assert.equal(returnedDevices, '0x');
        });
    });
});
