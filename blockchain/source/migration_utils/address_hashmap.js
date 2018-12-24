const TruffleConfig = require('../truffle');
const Multisig = require('./multisig');

class ContractRegistry {
    constructor (ahm, network, multisigContract) {
        this.ahmContract = ahm;
        this.network = network;
        this.multisigContract = multisigContract;
    }
    async init () {
        if (TruffleConfig.isSidechain(this.network)) {
            this.hm = await this.ahmContract.deployed();
            this.multisig = await Multisig.new(this.multisigContract, this, this.network, this.hm);
        } else {
            let alt = this.ahmContract.clone();
            let sideNet = TruffleConfig.networks[this.network].side_network_id;
            alt.setNetwork(sideNet);
            alt.setProvider(TruffleConfig.networks[TruffleConfig.oppositeNetName(this.network)].provider());
            this.hm = await alt.deployed();
            let sideNetName = TruffleConfig.oppositeNetName(this.network);
            this.multisig = await Multisig.new(this.multisigContract, this, sideNetName, this.hm);
        }
    }

    async resolve (contract, name) {
        let address = await this.hm.read(name);
        console.log('resolved', name, 'to', address);
        return contract.at(address);
    }

    async write (name, value) {
        if (this.multisig !== undefined) {
            return this.multisig.call('write', name, value);
        }
        return this.hm.write(name, value);
    }
}

module.exports = ContractRegistry;
