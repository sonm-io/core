require('babel-register');
require('babel-polyfill');
require('dotenv').config();
let PrivateKeyProvider = require('truffle-privatekey-provider');

let privateKey = '0000000000000000000000000000000000000000000000000000000000000000';

if (process.env.PRV_KEY !== undefined) {
    privateKey = process.env.PRV_KEY;
}
let masterchainEndpoint = 'https://mainnet.infura.io/';
let rinkebyEndpoint = 'https://rinkeby.infura.io/';
let sidechainEndpoint = 'https://sidechain.livenet.sonm.com';
let sidechainDevEndpoint = 'https://sidechain-dev.sonm.com';

let mochaConfig = {};
if (process.env.BUILD_TYPE === 'CI') {
    mochaConfig = {
        reporter: 'mocha-junit-reporter',
        reporterOptions: {
            mochaFile: 'result.xml',
        },
    };
}

module.exports = {
    networks: {
        development: {
            host: 'localhost',
            port: 8535,
            network_id: '*', // eslint-disable-line camelcase
            gas: 100000000,
        },
        coverage: {
            host: 'localhost',
            network_id: '*', // eslint-disable-line camelcase
            port: 8555,
            gas: 0xfffffffffff,
            gasPrice: 0x01,
        },

        master: {
            provider: () => new PrivateKeyProvider(privateKey, masterchainEndpoint),
            network_id: '1', // eslint-disable-line camelcase
        },
        rinkeby: {
            provider: () => new PrivateKeyProvider(privateKey, rinkebyEndpoint),
            network_id: '4', // eslint-disable-line camelcase
        },

        privateLive: {
            provider: () => new PrivateKeyProvider(privateKey, sidechainEndpoint),
            network_id: '444', // eslint-disable-line camelcase
        },
        private: {
            provider: () => new PrivateKeyProvider(privateKey, sidechainDevEndpoint),
            network_id: '444', // eslint-disable-line camelcase
        },
    },
    solc: {
        optimizer: {
            enabled: false,
            runs: 200,
        },
    },
    mocha: mochaConfig,
    // eslint-disable-next-line camelcase
    contracts_directory: './contracts',
};
