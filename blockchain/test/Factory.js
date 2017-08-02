'use strict';

const assertJump = require('./helpers/assertJump');
var SonmDummyToken = artifacts.require('../contracts/SonmDummyToken.sol');
var Factory = artifacts.require('../contracts/SonmDummyToken.sol');


contract('Factory', function (accounts) {
    let token;
    let factory;

    beforeEach(async function() {
        token = SonmDummyToken.new(accounts[0]);
    });

    it('should return be correctly deployed', async function () {
        try {
            factory = await Factory.new(accounts[0], token);
        } catch (error){
            assert.fail("")
        }
    });

});