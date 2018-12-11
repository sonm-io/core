const fs = require('fs');
const path = require('path');
const Web3 = require('web3');
require('babel-register');
require('babel-polyfill');
require('dotenv').config();
let PrivateKeyProvider = require('truffle-privatekey-provider');

let privateKey;
if (process.env.PRV_KEY !== undefined) {
    privateKey = process.env.PRV_KEY;
}

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

let urls = {
    development: 'http://localhost:8525',
    dev_side: 'http://localhost:8535',
    dev_main: 'http://localhost:8545',
    coverage: 'http://localhost:8555',
    master: 'https://mainnet.infura.io/',
    rinkeby: 'https://rinkeby.infura.io/',
    privateLive: 'https://sidechain.livenet.sonm.com',
    private: 'https://sidechain-dev.sonm.com',
};

let networks = {
    development: {
        network_id: '*', // eslint-disable-line camelcase
    },
    // eslint-disable-next-line camelcase
    dev_side: {
        network_id: '8535', // eslint-disable-line camelcase
        main_network_id: '8545', // eslint-disable-line camelcase
    },
    // eslint-disable-next-line camelcase
    dev_main: {
        network_id: '8545', // eslint-disable-line camelcase
        side_network_id: '8535', // eslint-disable-line camelcase
    },
    coverage: {
        network_id: '*', // eslint-disable-line camelcase
        gas: 0xfffffffffff,
        gasPrice: 0x01,
    },

    master: {
        network_id: '1', // eslint-disable-line camelcase
        side_network_id: '444', // eslint-disable-line camelcase
    },
    rinkeby: {
        network_id: '4', // eslint-disable-line camelcase
        side_network_id: '4242', // eslint-disable-line camelcase
    },

    privateLive: {
        network_id: '444', // eslint-disable-line camelcase
        main_network_id: '1', // eslint-disable-line camelcase
        gasPrice: 0x0,
    },
    private: {
        network_id: '4242', // eslint-disable-line camelcase
        main_network_id: '4', // eslint-disable-line camelcase
        gasPrice: 0x0,
    },
};

for (let net in networks) {
    let provider;
    if(privateKey !== undefined) {
        provider = () => new PrivateKeyProvider(privateKey, urls[net])
    } else {
        provider = () => new Web3.providers.HttpProvider(urls[net]);
    }
    networks[net].provider = provider;
}

module.exports = {
    networks: networks,
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
