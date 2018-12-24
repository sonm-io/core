const EthereumjsWallet = require('ethereumjs-wallet');
const PrivateKeyProvider = require('truffle-privatekey-provider');
const TruffleConfig = require('../truffle');

function sleep (millis) {
    return new Promise(resolve => setTimeout(resolve, millis));
}

class MSWrapper {
    static async new (msContract, resolver, network, wrappedContract, multisigKey) {
        let netID = TruffleConfig.networks[network].network_id;
        let msWrapper = new MSWrapper();
        msWrapper.wrappedContract = wrappedContract;
        let alt;
        if (msContract.network_id !== netID) {
            console.log('cloning multisig');
            alt = msContract.clone(netID);
            alt.class_defaults.gas = msContract.class_defaults.gas;
            alt.class_defaults.from = msContract.class_defaults.from;
            alt.setProvider(TruffleConfig.networks[network].provider());
        } else {
            console.log('using existing multisig');
            alt = msContract;
        }

        let pKey = TruffleConfig.multisigPrivateKey;
        if (pKey !== undefined) {
            let wallet = EthereumjsWallet.fromPrivateKey(Buffer.from(pKey, 'hex'));
            let from = '0x' + wallet.getAddress().toString('hex');
            if (from !== alt.class_defaults.from) {
                console.log('using separate pKey provider for multisig');
                alt.class_defaults.from = from;
                let provider = new PrivateKeyProvider(pKey, TruffleConfig.urls[network]);
                alt.setProvider(provider);
            }
        }

        if (multisigKey === undefined) {
            // intentional misspelling, as it was deployed in live under this name
            multisigKey = 'migrationMultSigAddress';
        }
        console.log('resolving ms');
        msWrapper.ms = await resolver.resolve(alt, multisigKey);
        console.log('ms resolved');
        return msWrapper;
    }

    async call (method, ...args) {
        return this.callImpl(false, method, ...args);
    }

    async callAndWait (method, ...args) {
        return this.callImpl(true, method, ...args);
    }

    async callImpl (shouldWait, method, ...args) {
        let expandedArgs = '(' + args.join(', ') + ')';
        console.log(`handling ${this.wrappedContract.constructor.contractName}.${method}${expandedArgs}` +
            ` at ${this.wrappedContract.address} via multisig at ${this.ms.address}`);
        let tx = this.wrappedContract.contract[method].getData(...args);
        console.log('TX: ', tx);
        let submitTx = await this.ms.submitTransaction(this.wrappedContract.address, 0, tx);
        let txID = submitTx.logs[0].args.transactionId.toNumber();
        console.log('TX id: ', txID);
        // eslint-disable-next-line no-unmodified-loop-condition
        while (shouldWait) {
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
