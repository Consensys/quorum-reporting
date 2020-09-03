import {
    getLastPersistedBlockNumber, getAddresses, getABI, getStorageABI, getContractCreationTransaction,
    getAllTransactionsToAddress, getAllTransactionsInternalToAddress, getAllEventsFromAddress, getStorageHistory,
    addAddress, addABI, addStorageABI, deleteAddress, getBlock, getTransaction,
} from '../client/rpcClient';

export const getBlockNumber = (baseURL) => {
    return getLastPersistedBlockNumber(baseURL).then( (res) => {
        if (res.data.error) {
            throw res.data.error.message
        }
        return res.data.result
    })
};

export const addContract = (baseURL, newContract) => {
    return addAddress(baseURL, newContract.address).then( (res) => {
        if (res.data.error) {
            throw res.data.error.message
        }
        return addABI(baseURL, newContract.address, newContract.abi).then( (res) => {
            if (res.data.error) {
                throw res.data.error.message
            }
            return addStorageABI(baseURL, newContract.address, newContract.template)
        })
    })
};

export const deleteContract = (baseURL, address) => {
    return deleteAddress(baseURL, address).then( (res) => {
        if (res.data.error) {
            throw res.data.error.message
        }
        return res
    })
};

export const getContracts = (baseURL) => {
    return getAddresses(baseURL).then( (res) => {
        if (res.data.error) {
            throw res.data.error.message
        }
        return getContractsDetail(baseURL, res.data.result)
    })
};

export const getContractCreationTx = (baseURL, address) => {
    return getContractCreationTransaction(baseURL, address).then( (res) => {
        if (res.data.error) {
            throw res.data.error.message
        }
        return getTransactionDetail(baseURL, res.data.result)
    })
};

export const getToTxs = (baseURL, address) => {
    return getAllTransactionsToAddress(baseURL, address).then( (res) => {
        if (res.data.error) {
            throw res.data.error.message
        }
        return getTransactionsDetail(baseURL, res.data.result["transactions"]).then( (txs) => {
            return {
                data: txs,
                total: res.data.result["total"]
            }
        })
    })
};

export const getInternalToTxs = (baseURL, address) => {
    return getAllTransactionsInternalToAddress(baseURL, address).then( (res) => {
        if (res.data.error) {
            throw res.data.error.message
        }
        return getTransactionsDetail(baseURL, res.data.result["transactions"]).then( (txs) => {
            return {
                data: txs,
                total: res.data.result["total"]
            }
        })
    })
};

export const getEvents = (baseURL, address) => {
    return getAllEventsFromAddress(baseURL, address).then( (res) => {
        if (res.data.error) {
            throw res.data.error.message
        }
        return {
            data: res.data.result["events"].map( (event) => ({
                topic: event.rawEvent.topics[0],
                txHash: event.rawEvent.transactionHash,
                address: event.rawEvent.address,
                blockNumber: event.rawEvent.blockNumber,
                parsedEvent: {
                    eventSig: event.eventSig,
                    parsedData: event.parsedData,
                },
            })),
            total: res.data.result["total"]
        }
    })
};

export const getReportData = (baseURL, address, startBlockNumber, endBlockNumber, currentPage) => {
    return getStorageHistory(baseURL, address, startBlockNumber, endBlockNumber, currentPage).then( (res) => {
        if (res.data.error) {
            throw res.data.error.message
        }
        return {
            data: generateReportData(res.data.result.historicState),
            total: res.data.result["total"]
    }
    })
};

export const getSingleBlock = (baseURL, blockNumber) => {
    return getBlock(baseURL, blockNumber).then( (res) => {
        if (res.data.error) {
            throw res.data.error.message
        }
        return res.data.result
    })
};

export const getSingleTransaction = (baseURL, txHash) => {
    return getTransaction(baseURL, txHash).then( (res) => {
        if (res.data.error) {
            throw res.data.error.message
        }
        return {
            txSig: res.data.result.txSig,
            func4Bytes: res.data.result.func4Bytes,
            parsedData: res.data.result.parsedData,
            parsedEvents: res.data.result.parsedEvents,
            ...res.data.result.rawTransaction
        }
    })
};

const getContractsDetail = (baseURL, addresses) => {
    return new Promise( (resolve, reject) => {
        if (addresses.length === 0) {
            resolve([])
        }
        let contracts = [];
        let abiCounter = addresses.length;
        let templateCounter = addresses.length;
        for (let i = 0; i < addresses.length; i++) {
            contracts.push({address: addresses[i]});
            ( (x) => {
                getABI(baseURL, addresses[x]).then( (res) => {
                    contracts[x].abi = res.data.result;
                    abiCounter--;
                    if (abiCounter === 0 && templateCounter === 0) {
                        resolve(contracts)
                    }
                }).catch(reject)
            })(i);
            ( (x) => {
                getStorageABI(baseURL, addresses[x]).then( (res) => {
                    contracts[x].template = res.data.result;
                    templateCounter--;
                    if (abiCounter === 0 && templateCounter === 0) {
                        resolve(contracts)
                    }
                }).catch(reject)
            })(i)
        }
    })
};

const getTransactionsDetail = (baseURL, txs) => {
    return new Promise( (resolve, reject) => {
        if (txs.length === 0) {
            resolve([])
        }
        let txsWithDetails = [];
        let counter = txs.length;
        for (let i = 0; i < txs.length; i++) {
            txsWithDetails.push({hash: txs[i]});
            ( (x) => {
                getTransactionDetail(baseURL, txs[x]).then( (res) => {
                    txsWithDetails[x] = res;
                    counter--;
                    if (counter === 0) {
                        resolve(txsWithDetails)
                    }
                }).catch(reject)
            })(i)
        }
    })
};

const getTransactionDetail = (baseURL, txHash) => {
    return getTransaction(baseURL, txHash).then( (res) => {
        return {
            hash: res.data.result.rawTransaction.hash,
            from: res.data.result.rawTransaction.from,
            to: res.data.result.rawTransaction.to,
            blockNumber: res.data.result.rawTransaction.blockNumber,
            parsedTransaction: {
                txSig: res.data.result.txSig,
                func4Bytes: res.data.result.func4Bytes,
                parsedData: res.data.result.parsedData,
            },
            parsedEvents: res.data.result.parsedEvents,
            internalCalls: res.data.result.rawTransaction.internalCalls,
        }
    })
};

const generateReportData = (historicState) => {
    if (historicState.length === 0) {
        return []
    }
    let reportData = [historicState[0]];
    let currentState = historicState[0];
    for (let i = 1; i < historicState.length; i++) {
        let nextState = isStateEqual(currentState, historicState[i]);
        if (nextState) {
            reportData.unshift(nextState);
            currentState = nextState
        }
    }
    return reportData
};

const isStateEqual = (state1, state2) => {
    let markedState2 = state2;
    let noChange = true;
    for (let i = 0; i < state1.historicStorage.length; i++) {
        if (state1.historicStorage[i].value !== state2.historicStorage[i].value) {
            markedState2.historicStorage[i].changed = true;
            noChange = false
        }
    }
    if (noChange) {
        return false
    }
    return markedState2
};