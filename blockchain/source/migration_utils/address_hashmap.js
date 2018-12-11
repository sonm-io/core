let { isSidechain, oppositeNetName } = require('../migration_utils/network');
const TruffleConfig = require('../truffle');
const MSWrapper = require('../migration_utils/multisig');

class ContractRegistry {
    constructor(ahm, network, multisigContract) {
        this.ahmContract = ahm;
        this.network = network;
        this.multisigContract = multisigContract;
    }
    async init() {
        if (isSidechain(this.network)) {
            this.hm = await this.ahmContract.deployed();
            this.multisig = await this.resolve(this.multisigContract, 'multiSigAddress');
        } else {
            let alt = this.ahmContract.clone();
            let sideNet = TruffleConfig.networks[this.network].side_network_id;
            alt.setNetwork(sideNet);
            alt.setProvider(TruffleConfig.networks[oppositeNetName(this.network)].provider());
            this.hm = await alt.deployed();

            alt = this.multisigContract.clone();
            alt.class_defaults.from = this.multisigContract.class_defaults.from;
            alt.class_defaults.gas = this.multisigContract.class_defaults.gas;
            alt.setNetwork(sideNet);
            alt.setProvider(TruffleConfig.networks[oppositeNetName(this.network)].provider());
            this.multisig = await this.resolve(alt, 'multiSigAddress');
        }
    }

    async resolve(contract, name) {
        let address = await this.hm.read(name);
        return contract.at(address)
    }

    async write(name, value) {
        if(this.multisig !== undefined) {
            let ms = new MSWrapper(this.multisig, this.hm);
            return ms.call('write', name, value);
        }
        return this.hm.write(name, value);
    }

}

module.exports = ContractRegistry;
