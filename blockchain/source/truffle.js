require('babel-register');
require('babel-polyfill');
require('dotenv').config();
let PrivateKeyProvider = require('truffle-privatekey-provider');

let privateKey = 'a000000000000000000000000000000000000000000000000000000000000000'; // for test purposes
if (process.env.PRV_KEY !== undefined) {
    privateKey = process.env.PRV_KEY;
}
let privateEndpoint = 'https://sidechain-dev.sonm.com';

module.exports = {
    networks: {
        development: {
            host: 'localhost',
            port: 8535,
            network_id: '*', // eslint-disable-line camelcase
        },
        rinkeby: {
            host: 'localhost',
            port: 8545,
            network_id: '4', // eslint-disable-line camelcase
        },
        coverage: {
            host: 'localhost',
            network_id: '*', // eslint-disable-line camelcase
            port: 8555,
            gas: 0xfffffffffff,
            gasPrice: 0x01,
        },
        private: {
            provider: new PrivateKeyProvider(privateKey, privateEndpoint),
            network_id: '444', // eslint-disable-line camelcase
        },
    },
    solc: {
        optimizer: {
            enabled: true,
            runs: 200,
        },
    },
};
