let { isSidechain, oppositeNetName } = require('../migration_utils/network');
const TruffleConfig = require('./truffle');

class Resolver {

    constructor(ahm, network) {
        if (isSidechain(network)) {
            this.hm = ahm.deployed();
        } else {
            let alt = ahm.clone();
            let sideNet = TruffleConfig.networks[network].side_network_id;
            alt.setNetwork(sideNet);
            alt.setProvider(TruffleConfig.networks[oppositeNetName(network)].provider());
            this.hm = alt.deployed();
        }
    }

    async resolve(contract, name) {
        if (name === undefined) {
            name = contract.contractName;
        }
        address = await this.hm.read(name);
        return contract.at(address)
    }

}

module.exports = Resolver;
