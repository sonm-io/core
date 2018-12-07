module.exports = {
    isSidechain: function (network) {
        return network === 'dev_side' || network === 'privateLive' || network === 'private';
    },
    isMainChain: function (network) {
        return network === 'dev_main' || network === 'master' || network === 'rinkeby';
    },
};
