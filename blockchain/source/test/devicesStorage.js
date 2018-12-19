import assertRevert from './helpers/assertRevert';
import increaseTime from './helpers/increaseTime'; // eslint-disable-next-line no-unused-vars
import { eventInTransaction, allEventsInTransaction } from './helpers/expectEvent';

let DevicesStorage = artifacts.require('./DevicesStorage.sol');

contract('DevicesStorage', (accounts) => {
    let devicesStorage;
    let worker = accounts[1];

    let devices = '0x1337';
    let hash = '0x2636a8beb2c41b8ccafa9a55a5a5e333892a83b491df3a67d2768946a9f9c6dc';

    before(async () => {
        devicesStorage = await DevicesStorage.new();
    });

    describe('SetDevices', () => {
        it('should set devices', async () => {
            let tx = await devicesStorage.SetDevices(devices, { from: worker });
            await eventInTransaction(tx, 'DevicesHasSet');
        });

        it('should return proper hash for devices', async () => {
            let curHash = await devicesStorage.Hash(worker);
            assert.equal(curHash, hash);
        });

        it('should update devices timestamp by hash', async () => {
            await increaseTime(86400);
            let tx = await devicesStorage.Touch(hash, { from: worker });
            await eventInTransaction(tx, 'DevicesTimestampUpdated');
            increaseTime(86000);
            let returnedDevices = await devicesStorage.GetDevices(worker);
            assert.equal(returnedDevices[0], devices);
        });

        it('should do not update devices by hash, revert', async () => {
            await increaseTime(1);
            await assertRevert(devicesStorage.Touch('0x1', { from: worker }));
        });

        it('should return Devices same as written', async () => {
            let returnedDevices = await devicesStorage.GetDevices(worker);
            assert.equal(returnedDevices[0], devices);
        });

        it('should emit DevicesUpdated event', async () => {
            let tx = await devicesStorage.SetDevices(devices, { from: worker });
            await eventInTransaction(tx, 'DevicesUpdated');
        });

        it('should be able to be killed', async () => {
            let tx = await devicesStorage.KillDevicesStorage();
            await eventInTransaction(tx, 'Suicide');
        });
    });
});
