function sleep(millis) {
    return new Promise(resolve => setTimeout(resolve, millis));
}

class MSWrapper {
    constructor(ms, wrappedContract) {
        this.ms = ms;
        this.wrappedContract = wrappedContract;
    }
    async call(method, ...args) {
        console.log('handling ', this.wrappedContract.contractName, '.', method, '(', ...args, ') via multisig at ', this.ms.address);
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
                console.log(transaction);
                return;
            }
            console.log('transaction has not been confirmed yet, waiting...');
            await sleep(1000);
        }

    }
}
module.exports = MSWrapper;
