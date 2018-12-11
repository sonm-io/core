const EthereumjsWallet = require('ethereumjs-wallet');
const PrivateKeyProvider = require('truffle-privatekey-provider');

function sleep(millis) {
    return new Promise(resolve => setTimeout(resolve, millis));
}

class MSWrapper {
    constructor(ms, wrappedContract) {
        this.ms = ms;
        this.wrappedContract = wrappedContract;
    }

    static async new(msContract, network, privateKey, providerUrl) {
        let alt = msContract.clone();
        let wallet = EthereumjsWallet.fromPrivateKey(new Buffer(privateKey, "hex"));
        let address = "0x" + this.wallet.getAddress().toString("hex");

        alt.class_defaults.from = address;
        alt.class_defaults.gas = msContract.class_defaults.gas;
        alt.setNetwork(network);
        let provider = new PrivateKeyProvider(privateKey, )
        alt.setProvider(TruffleConfig.networks[oppositeNetName(this.network)].provider());
        this.multisig = await this.resolve(alt, 'multiSigAddress');
    }

    async call(method, ...args) {
        let expandedArgs = '(' + args.join(', ') + ')';
        console.log(`handling ${this.wrappedContract.constructor.contractName}.${method}${expandedArgs}` +
            ` at ${this.wrappedContract.address} via multisig at ${this.ms.address}`);
        let tx = this.wrappedContract.contract[method].getData(...args);
        console.log('TX: ', tx);
        let submitTx = await this.ms.submitTransaction(this.wrappedContract.address, 0, tx);
        let txID = submitTx.logs[0].args.transactionId.toNumber();
        console.log('TX id: ', txID);
        while (true) {
            let confirmed = await this.ms.isConfirmed(txID);
            if(confirmed) {
                console.log('transaction has been confirmed');
                let transaction = await this.ms.transactions.call(txID);
                if (!transaction[3]) {
                    throw new Error('transaction was confirmed, but failed to execute');
                }
                return;
            }
            console.log('transaction has not been confirmed yet, waiting...');
            await sleep(1000);
        }

    }
}
module.exports = MSWrapper;
