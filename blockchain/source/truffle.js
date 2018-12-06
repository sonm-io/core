const fs = require('fs');
const path = require('path');
const Web3 = require('web3');
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

let buildFolder = path.join(process.cwd(), 'build');
if (process.env.MIGRATION === 'true') {
    buildFolder = path.join(process.cwd(), 'migration_artifacts');
    fs.mkdir(buildFolder, { recursive: true }, (err) => {
        if (err) throw err;
    });
}

module.exports = {
    networks: {
        development: {
            provider: () => new Web3.providers.HttpProvider("http://localhost:8525"),
            network_id: '*', // eslint-disable-line camelcase
        },
        // eslint-disable-next-line camelcase
        dev_side: {
            provider: () => new Web3.providers.HttpProvider("http://localhost:8535"),
            network_id: '8535', // eslint-disable-line camelcase
            main_network_id: '8545', // eslint-disable-line camelcase
        },
        // eslint-disable-next-line camelcase
        dev_main: {
            provider: () => new Web3.providers.HttpProvider("http://localhost:8545"),
            network_id: '*', // eslint-disable-line camelcase
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
            main_network_id: '1', // eslint-disable-line camelcase
        },
        private: {
            provider: () => new PrivateKeyProvider(privateKey, sidechainDevEndpoint),
            network_id: '4444', // eslint-disable-line camelcase
            main_network_id: '4', // eslint-disable-line camelcase
        },
    },
    solc: {
        optimizer: {
            enabled: true,
            runs: 200,
        },
    },
    mocha: mochaConfig,
    // eslint-disable-next-line camelcase
    contracts_directory: './contracts',
    // eslint-disable-next-line camelcase
    build_directory: buildFolder,
};
