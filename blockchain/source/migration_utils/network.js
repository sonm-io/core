let networkMapping = {
    dev_main: 'dev_side',
    dev_side: 'dev_main',
    rinkeby: 'private',
    private: 'rinkeby',
    master: 'privateLive',
    privateLive: 'master',
};

//TODO: move to truffle config
module.exports = {
    isSidechain: function (network) {
        return network === 'dev_side' || network === 'privateLive' || network === 'private';
    },
    isMainChain: function (network) {
        return network === 'dev_main' || network === 'master' || network === 'rinkeby';
    },
    oppositeNetName: function (network) {
        return networkMapping[network];
    },
};
