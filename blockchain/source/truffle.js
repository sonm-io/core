/* eslint-disable camelcase */
const fs = require('fs');
const path = require('path');
const Web3 = require('web3');
require('babel-register');
require('babel-polyfill');
require('dotenv').config();
let PrivateKeyProvider = require('truffle-privatekey-provider');

let privateKey;
let msPrivateKey;
if (process.env.PRV_KEY !== undefined) {
    privateKey = process.env.PRV_KEY;
}
if (process.env.MS_PRV_KEY !== undefined) {
    msPrivateKey = process.env.MS_PRV_KEY;
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
    development: 'http://localhost:8535',
    dev_side: 'http://localhost:8525',
    dev_main: 'http://localhost:8545',
    coverage: 'http://localhost:8555',
    master: 'https://mainnet.infura.io/',
    rinkeby: 'https://rinkeby.infura.io/',
    privateLive: 'https://sidechain.livenet.sonm.com',
    private: 'https://sidechain-testnet.sonm.com',
};

let networks = {
    development: {
        network_id: '*',
    },
    dev_side: {
        network_id: '8525',
    },
    dev_main: {
        network_id: '8545',
    },
    coverage: {
        network_id: '*',
        gas: 0xfffffffffff,
        gasPrice: 0x01,
    },

    master: {
        network_id: '1',
        gasPrice: 5000000000,
        gas: 1500000,
    },
    rinkeby: {
        network_id: '4',
    },

    privateLive: {
        network_id: '444',
        gasPrice: 0x0,
    },
    private: {
        network_id: '4242',
        gasPrice: 0x0,
    },
};

for (let net in networks) {
    let provider;
    if (privateKey !== undefined) {
        provider = () => new PrivateKeyProvider(privateKey, urls[net]);
    } else {
        provider = () => new Web3.providers.HttpProvider(urls[net]);
    }
    networks[net].provider = provider;
}

let networkMapping = {
    dev_main: 'dev_side',
    dev_side: 'dev_main',
    rinkeby: 'private',
    private: 'rinkeby',
    master: 'privateLive',
    privateLive: 'master',
};

module.exports = {
    networks: networks,
    urls: urls,
    solc: {
        optimizer: {
            enabled: true,
            runs: 200,
        },
    },
    mocha: mochaConfig,
    contracts_directory: './contracts',
    build_directory: buildFolder,
    isSidechain: function (network) {
        return network === 'dev_side' || network === 'privateLive' || network === 'private';
    },
    isMainChain: function (network) {
        return network === 'dev_main' || network === 'master' || network === 'rinkeby';
    },
    oppositeNetName: function (network) {
        return networkMapping[network];
    },
    multisigPrivateKey: msPrivateKey,
};
