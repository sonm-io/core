const EthereumjsWallet = require('ethereumjs-wallet');
const PrivateKeyProvider = require('truffle-privatekey-provider');
const TruffleConfig = require('../truffle');

function sleep (millis) {
    return new Promise(resolve => setTimeout(resolve, millis));
}

class MSWrapper {
    static async new (msContract, resolver, network, wrappedContract) {
        let netID = TruffleConfig.networks[network].network_id;
        let msWrapper = new MSWrapper();
        msWrapper.wrappedContract = wrappedContract;
        let alt = msContract.clone(netID);
        alt.class_defaults.gas = msContract.class_defaults.gas;

        let pKey = TruffleConfig.multisigPrivateKey;
        if (pKey !== undefined) {
            let wallet = EthereumjsWallet.fromPrivateKey(Buffer.from(pKey, 'hex'));
            alt.class_defaults.from = wallet.getAddress().toString('hex');
            let provider = new PrivateKeyProvider(pKey, TruffleConfig.urls[network]);
            alt.setProvider(provider);
        } else {
            alt.class_defaults.from = msContract.class_defaults.from;
            alt.setProvider(TruffleConfig.networks[network].provider());
        }

        msWrapper.ms = await resolver.resolve(alt, 'multiSigAddress');
        return msWrapper;
    }

    async call (method, ...args) {
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
            if (confirmed) {
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
