import assertRevert from './helpers/assertRevert';

const ProfileRegistry = artifacts.require('./ProfileRegistry.sol');

contract('ProfileRegistry', async function (accounts) {
    let registry;

    const owner = accounts[0];
    const master = accounts[1];
    const creeper = accounts[2];
    const testValidator = accounts[3];

    const validatorLvl1 = accounts[5];
    const validatorLvl2 = accounts[6];
    const validatorLvl3 = accounts[7];
    const validatorLvl4 = accounts[8];

    const testLevel = 3;

    describe('Defaults', function () {
        before(async function () {
            registry = await ProfileRegistry.new();
        });

        it('owner has level `-1`', async function () {
            let validator = await registry.GetValidatorLevel(owner);
            assert.equal(validator.toNumber(), -1);
        });
    });

    describe('AddValidator', function () {
        before(async function () {
            registry = await ProfileRegistry.new();
        });

        it('add new validator and spend `ValidatorCreated` event', async function () {
            const { logs } = await registry.AddValidator(testValidator, testLevel, { from: owner });
            assert.equal(logs.length, 1);
            assert.equal(logs[0].event, 'ValidatorCreated');
            assert.equal(logs[0].args.validator, testValidator);
        });

        it('with preffred lvl', async function () {
            let validator = await registry.GetValidatorLevel(testValidator);
            assert.equal(validator.toNumber(), testLevel);
        });

        it('add new validator twice - should revert', async function () {
            await assertRevert(registry.AddValidator(testValidator, testLevel, { from: owner }));
        });

        describe('when not owner want to add validator', function () {
            it('should revert', async function () {
                await assertRevert(registry.AddValidator(master, testLevel, { from: creeper }));
            });
        });

        describe('when the set value less than zero', function () {
            it('should revert', async function () {
                await assertRevert(registry.AddValidator(master, 0), { from: owner });
            });
        });
    });

    describe('RemoveValidator', function () {
        before(async function () {
            registry = await ProfileRegistry.new();

            await registry.AddValidator(testValidator, testLevel, { from: owner });
        });

        it('remove validator and spend `ValidatorDeleted` event', async function () {
            await registry.RemoveValidator(testValidator, { from: owner });
        });

        it('set removed validator level to `0`', async function () {
            let validator = await registry.GetValidatorLevel(testValidator);
            assert.equal(validator.toNumber(), 0);
        });

        describe('when validator want to remove itself', function () {
            before(async function () {
                await registry.AddValidator(testValidator, testLevel, { from: owner });
            });

            it('should revert', async function () {
                await assertRevert(registry.RemoveValidator(testValidator, { from: creeper }));
            });

            after(async function () {
                await registry.RemoveValidator(testValidator, { from: owner });
            });
        });

        describe('when stranger want to remove validator', function () {
            before(async function () {
                await registry.AddValidator(testValidator, testLevel, { from: owner });
            });

            it('should revert', async function () {
                await assertRevert(registry.RemoveValidator(testValidator, { from: creeper }));
            });

            after(async function () {
                await registry.RemoveValidator(testValidator, { from: owner });
            });
        });

        describe('when validator is not setted', function () {
            it('should revert', async function () {
                await assertRevert(registry.RemoveValidator(master, { from: owner }));
            });
        });
    });

    describe('CreateCertificate', function () {
        before(async function () {
            registry = await ProfileRegistry.new();

            await registry.AddValidator(validatorLvl1, 1, { from: owner });
            await registry.AddValidator(validatorLvl2, 2, { from: owner });
            await registry.AddValidator(validatorLvl3, 3, { from: owner });
            await registry.AddValidator(validatorLvl4, 4, { from: owner });
        });

        describe('when set value is empty', function () {
            it('should revert', async function () {
                await assertRevert(registry.CreateCertificate(master, 1002, '', { from: master }));
            });
        });

        describe('when certificate has single attribute', function () {
            describe('user can set attribute with 0lvl itself', function () {
                it('should setting value', async function () {
                    await registry.CreateCertificate(master, 1002, 'value', { from: master });
                    const value = await registry.GetAttributeValue(master, 1002);
                    assert.equal(web3.toAscii(value).toString(), 'value');
                });
            });

            it('validator with lvl1 should create certificate 1lvl', async function () {
                await registry.CreateCertificate(master, 1102, 'value', { from: validatorLvl1 });
                const value = await registry.GetAttributeValue(master, 1102);
                assert.equal(web3.toAscii(value), 'value');
            });

            it('validator with lvl2 should create certificate 2lvl', async function () {
                await registry.CreateCertificate(master, 1201, 'value', { from: validatorLvl2 });
                const value = await registry.GetAttributeValue(master, 1201);
                assert.equal(web3.toAscii(value), 'value');
            });

            it('validator with lvl3 should create certificate 3lvl', async function () {
                await registry.CreateCertificate(master, 1301, 'value', { from: validatorLvl3 });
                const value = await registry.GetAttributeValue(master, 1301);
                assert.equal(web3.toAscii(value), 'value');
            });

            it('validator with lvl4 should create certificate 3lvl', async function () {
                await registry.CreateCertificate(master, 1301, 'value', { from: validatorLvl4 });
                const count = await registry.GetAttributeCount(master, 1301);
                assert.equal(count.toNumber(), 2);
            });

            it('validator with lvl4 should create certificate 4lvl', async function () {
                await registry.CreateCertificate(master, 1401, 'value', { from: validatorLvl4 });
                const value = await registry.GetAttributeValue(master, 1401);
                assert.equal(web3.toAscii(value), 'value');
            });

            describe('when validator create certificate with same value', function () {
                let firstCertificateId;
                let secondCertificateId;
                let startValue;
                let startCount;
                let testType = 1203;

                before(async function () {
                    const { logs } = await registry.CreateCertificate(master, testType, 'value',
                        { from: validatorLvl2 });
                    const value = await registry.GetAttributeValue(master, testType);
                    assert.equal(web3.toAscii(value), 'value');

                    firstCertificateId = logs[0].args.id;
                    startValue = await registry.GetAttributeValue(master, testType);
                    startCount = await registry.GetAttributeCount(master, testType);
                });

                it('should created', async function () {
                    const { logs } = await registry.CreateCertificate(master, testType, 'value',
                        { from: validatorLvl3 });
                    secondCertificateId = logs[0].args.id;
                });

                it('value cant changed', async function () {
                    const value = await registry.GetAttributeValue(master, testType);
                    assert.equal(web3.toAscii(value), 'value');
                    assert.equal(value, startValue);
                });

                it('attribute count should increase', async function () {
                    const count = await registry.GetAttributeCount(master, testType);
                    assert.equal(count.toNumber(), parseInt(startCount) + 1);
                });

                after(async function () {
                    await registry.RemoveCertificate(firstCertificateId, { from: master });
                    await registry.RemoveCertificate(secondCertificateId, { from: master });
                });
            });

            describe('when validator create certificate with other value', function () {
                let firstCertificateId;
                let testType = 1203;
                let testValue = 'value';
                before(async function () {
                    const { logs } = await registry.CreateCertificate(master, testType, testValue,
                        { from: validatorLvl2 });
                    const value = await registry.GetAttributeValue(master, testType);
                    assert.equal(web3.toAscii(value), testValue);

                    firstCertificateId = logs[0].args.id;
                });

                it('should revert', async function () {
                    await assertRevert(registry.CreateCertificate(master, testType, testValue + testValue,
                        { from: validatorLvl3 }));
                });

                after(async function () {
                    await registry.RemoveCertificate(firstCertificateId, { from: master });
                });
            });
        });

        describe('when certificate has multiple attribute', function () {
            it('should setting, but GetAttributeValue returns `` for multiple', async function () {
                await registry.CreateCertificate(master, 2202, 'value', { from: validatorLvl4 });
                const value = await registry.GetAttributeValue(master, 2202);
                assert.equal(web3.toAscii(value), '');
            });

            describe('when certificate with same type added twice', function () {
                it('attribute count should increase by 1', async function () {
                    await registry.CreateCertificate(master, 2202, 'value', { from: validatorLvl4 });
                    const count = await registry.GetAttributeCount(master, 2202);
                    assert.equal(count.toNumber(), 2);
                });
            });
        });

        describe('when stranger want to create 0lvl cert to other', function () {
            it('should revert', async function () {
                await assertRevert(registry.CreateCertificate(master, 1002, 'value', { from: creeper }));
            });
        });

        describe('when stranger want to create 1lvl cert', function () {
            it('should revert', async function () {
                await assertRevert(registry.CreateCertificate(master, 1102, 'value', { from: creeper }));
            });
        });

        describe('when validator with 2lvl want to create 4lvl cert', function () {
            it('should revert', async function () {
                await assertRevert(registry.CreateCertificate(creeper, 1401, 'value', { from: validatorLvl2 }));
            });
        });

        describe('when validator with 3lvl want to create 4lvl cert', function () {
            it('should revert', async function () {
                await assertRevert(registry.CreateCertificate(creeper, 1401, 'value', { from: validatorLvl3 }));
            });
        });
    });

    describe('GetCertificate', function () {
        let certificateId;
        let testType = 1203;
        let testValue = 'val';
        before(async function () {
            const { logs } = await registry.CreateCertificate(master, testType, testValue, { from: validatorLvl2 });
            certificateId = logs[0].args.id;
        });

        it('should return existing certificate with correct values', async function () {
            const result = await registry.GetCertificate(certificateId);
            assert.equal(result[0], validatorLvl2);
            assert.equal(result[1], master);
            assert.equal(result[2], testType);
            assert.equal(web3.toAscii(result[3]), testValue);
        });

        describe('when certificate does not exist', async function () {
            it('should return zero values', async function () {
                const result = await registry.GetCertificate(99999999999);
                assert.equal(result[0], '0x0000000000000000000000000000000000000000');
                assert.equal(result[1], '0x0000000000000000000000000000000000000000');
                assert.equal(result[2].toNumber(), 0);
                assert.equal(web3.toAscii(result[3]), '');
            });
        });
    });

    describe('RemoveCertificate', function () {
        before(async function () {
            registry = await ProfileRegistry.new();

            await registry.AddValidator(validatorLvl1, 1, { from: owner });
            await registry.AddValidator(validatorLvl2, 2, { from: owner });
            await registry.AddValidator(validatorLvl3, 3, { from: owner });
            await registry.AddValidator(validatorLvl4, 4, { from: owner });
        });

        describe('when other validator want remove certificate', function () {
            let certId;

            before(async function () {
                const { logs } = await registry.CreateCertificate(master, 1201, 'value', { from: validatorLvl2 });
                certId = logs[0].args.id;
            });

            it('should revert', async function () {
                await assertRevert(registry.RemoveCertificate(certId, { from: validatorLvl3 }));
            });

            after(async function () {
                await registry.RemoveCertificate(certId, { from: master });
            });
        });

        describe('when stranger want remove certificate', function () {
            let certId;

            before(async function () {
                const { logs } = await registry.CreateCertificate(master, 1201, 'value', { from: validatorLvl2 });
                certId = logs[0].args.id;
            });

            it('should revert', async function () {
                await assertRevert(registry.RemoveCertificate(certId, { from: creeper }));
            });

            after(async function () {
                await registry.RemoveCertificate(certId, { from: master });
            });
        });

        describe('when SONM want remove certificate', function () {
            let certId;

            before(async function () {
                const { logs } = await registry.CreateCertificate(master, 1201, 'value', { from: validatorLvl2 });
                certId = logs[0].args.id;
            });

            it('should removed', async function () {
                await registry.RemoveCertificate(certId, { from: owner });
            });
        });

        describe('when certificates has single attribute with one confirmation', function () {
            let certId;

            before(async function () {
                const { logs } = await registry.CreateCertificate(master, 1201, 'value', { from: validatorLvl2 });
                certId = logs[0].args.id;
            });

            it('should removed', async function () {
                await registry.RemoveCertificate(certId, { from: validatorLvl2 });
            });

            it('should decrease attribute count to 0', async function () {
                const count = await registry.GetAttributeCount(master, 1401);
                assert.equal(count.toNumber(), 0);
            });

            it('should set attribute value to ``', async function () {
                const value = await registry.GetAttributeValue(master, 1401);
                assert.equal(web3.toAscii(value), '');
            });
        });

        describe('when certificate had single attribute with two confirmations', function () {
            let firstCertId;
            let secondCertId;
            let startedCount;
            let startedValue;
            let testValue = 'val';
            let testType = 1213;

            before(async function () {
                let { logs } = await registry.CreateCertificate(master, testType, testValue, { from: validatorLvl2 });
                firstCertId = logs[0].args.id;

                let tx = await registry.CreateCertificate(master, testType, testValue, { from: validatorLvl4 });
                secondCertId = tx.logs[0].args.id;

                startedValue = await registry.GetAttributeValue(master, testType);
                assert.equal(web3.toAscii(startedValue), testValue);
                startedCount = await registry.GetAttributeCount(master, testType);
            });

            describe('when first certificate removed', function () {
                it('should remove', async function () {
                    await registry.RemoveCertificate(firstCertId, { from: master });
                });

                it('attribute value should not changed', async function () {
                    let value = await registry.GetAttributeValue(master, testType);
                    assert.equal(web3.toAscii(value), web3.toAscii(startedValue));
                });

                it('attribute count should decrease by 1', async function () {
                    let count = await registry.GetAttributeCount(master, testType);
                    assert.equal(count.toNumber(), startedCount.toNumber() - 1);
                });
            });

            describe('when second certificate removed', function () {
                it('should removed', async function () {
                    await registry.RemoveCertificate(secondCertId, { from: master });
                });

                it('attribute value should changed to ``', async function () {
                    let value = await registry.GetAttributeValue(master, testType);
                    assert.equal(web3.toAscii(value), '');
                });

                it('attribute count should decrease to 0', async function () {
                    let count = await registry.GetAttributeCount(master, testType);
                    assert.equal(count.toNumber(), 0);
                });
            });
        });

        describe('when certificate has multiple attribute', function () {
            let firstCertId;
            let secondCertId;
            let startedCount;
            let startedValue;
            let testValue = 'val';
            let testType = 2213;

            before(async function () {
                let { logs } = await registry.CreateCertificate(master, testType, testValue, { from: validatorLvl2 });
                firstCertId = logs[0].args.id;

                let tx = await registry.CreateCertificate(master, testType, testValue, { from: validatorLvl4 });
                secondCertId = tx.logs[0].args.id;

                startedValue = await registry.GetAttributeValue(master, testType);
                assert.equal(web3.toAscii(startedValue), '');
                startedCount = await registry.GetAttributeCount(master, testType);
            });

            describe('when first certificate removed', function () {
                it('should remove', async function () {
                    await registry.RemoveCertificate(firstCertId, { from: master });
                });

                it('attribute value should not changed', async function () {
                    let value = await registry.GetAttributeValue(master, testType);
                    assert.equal(web3.toAscii(value), web3.toAscii(startedValue));
                });

                it('attribute count should decrease by 1', async function () {
                    let count = await registry.GetAttributeCount(master, testType);
                    assert.equal(count.toNumber(), startedCount.toNumber() - 1);
                });
            });

            describe('when second certificate removed', function () {
                it('should removed', async function () {
                    await registry.RemoveCertificate(secondCertId, { from: master });
                });

                it('attribute value should changed to ``', async function () {
                    let value = await registry.GetAttributeValue(master, testType);
                    assert.equal(web3.toAscii(value), '');
                });

                it('attribute count should decrease to 0', async function () {
                    let count = await registry.GetAttributeCount(master, testType);
                    assert.equal(count.toNumber(), 0);
                });
            });
        });

        describe('when certificate already removed', function () {
            let certId;
            let testValue = 'val';
            let testType = 1213;

            before(async function () {
                let { logs } = await registry.CreateCertificate(master, testType, testValue, { from: validatorLvl2 });
                certId = logs[0].args.id;
                await registry.RemoveCertificate(certId, { from: master });
            });

            it('should revert', async function () {
                await assertRevert(registry.RemoveCertificate(certId, { from: master }));
            });
        });
    });

    describe('CheckProfileLevel', function () {
        before(async function () {
            registry = await ProfileRegistry.new();

            await registry.AddValidator(validatorLvl1, 1, { from: owner });
            await registry.AddValidator(validatorLvl2, 2, { from: owner });
            await registry.AddValidator(validatorLvl3, 3, { from: owner });
            await registry.AddValidator(validatorLvl4, 4, { from: owner });

            await registry.CreateCertificate(master, 1102, 'value', { from: validatorLvl1 });
            await registry.CreateCertificate(master, 1201, 'value', { from: validatorLvl2 });
            await registry.CreateCertificate(master, 1301, 'value', { from: validatorLvl4 });
            await registry.CreateCertificate(master, 1401, 'value', { from: validatorLvl4 });
        });

        describe('when trying to check >4 lvl ', function () {
            it('should return false', async function () {
                let result = await registry.CheckProfileLevel.call(master, 5);
                assert.equal(result, false);
            });
        });

        describe('when trying to check 4 lvl and certificate exist', function () {
            it('should return true', async function () {
                let result = await registry.CheckProfileLevel.call(master, 4);
                assert.equal(result, true);
            });
        });

        describe('when trying to check 4 lvl and certificate doesnt exist', function () {
            it('should return false', async function () {
                let result = await registry.CheckProfileLevel.call(creeper, 4);
                assert.equal(result, false);
            });
        });

        describe('when trying to check 3 lvl and certificate exist', function () {
            it('should returns true', async function () {
                let result = await registry.CheckProfileLevel.call(master, 3);
                assert.equal(result, true);
            });
        });

        describe('when trying to check 3 lvl and certificate doesnt exist', function () {
            it('should return false', async function () {
                let result = await registry.CheckProfileLevel.call(creeper, 3);
                assert.equal(result, false);
            });
        });

        describe('when trying to check 2 lvl and certificate exist', function () {
            it('should return true', async function () {
                let result = await registry.CheckProfileLevel.call(master, 2);
                assert.equal(result, true);
            });
        });

        describe('when trying to check 2 lvl and certificate doesnt exist', function () {
            it('should returns false', async function () {
                let result = await registry.CheckProfileLevel.call(creeper, 3);
                assert.equal(result, false);
            });
        });

        describe('when trying to check 1 lvl', function () {
            it('should return true', async function () {
                let result = await registry.CheckProfileLevel.call(master, 1);
                assert.equal(result, true);
            });
        });

        describe('when trying to check 0 lvl', function () {
            it('should return true', async function () {
                let result = await registry.CheckProfileLevel.call(master, 0);
                assert.equal(result, true);
            });
        });
    });
});
