import {
    addAddress,
    addTemplate,
    assignTemplate,
    deleteAddress,
    getABI,
    getAddresses,
    getAllEventsFromAddress,
    getAllTransactionsInternalToAddress,
    getAllTransactionsToAddress,
    getBlock,
    getContractCreationTransaction,
    getERC20TokenHolders,
    getERC721TokensAtBlock,
    getHolderForERC721TokenAtBlock,
    getLastPersistedBlockNumber,
    getStorageABI,
    getStorageHistory,
    getTransaction,
} from '../client/rpcClient'
import {
    getContractTemplate,
    getERC20TokenBalance,
    getERC721TokenHolders,
    getERC721TokensForAccountAtBlock,
    getStorageHistoryCount
} from './rpcClient'

export const getBlockNumber = (baseURL) => {
    return getLastPersistedBlockNumber(baseURL).then( (res) => {
        return res
    })
};

export const addContract = (baseURL, newContract) => {
    return addAddress(baseURL, newContract.address).then( (res) => {
        if(newContract.template === 'new') {
            return addTemplate(baseURL, newContract.newTemplate)
              .then((template) => assignTemplate(baseURL, newContract.address, newContract.newTemplate.name))
        }
        return assignTemplate(baseURL, newContract.address, newContract.template)
    })
};

export const deleteContract = (baseURL, address) => {
    return deleteAddress(baseURL, address)
};

export const getContracts = (baseURL) => {
    return getAddresses(baseURL).then( (res) => {
        return getContractsDetail(baseURL, res)
          .then((contracts) => {
                // sort by template name + address for consistent order
                return contracts.sort((a, b) =>
                  `${a.name}${a.address}`.localeCompare(`${b.name}${b.address}`))
            }
          )
    })
};

export const getContractCreationTx = (baseURL, address) => {
    return getContractCreationTransaction(baseURL, address).then( (res) => {
        return res
    })
};

export const getToTxs = (baseURL, address, options) => {
    return getAllTransactionsToAddress(baseURL, address, options).then( (res) => {
        return getTransactionsDetail(baseURL, res["transactions"]).then( (txs) => {
            return {
                data: txs,
                total: res["total"]
            }
        })
    })
};

export const getInternalToTxs = (baseURL, address) => {
    return getAllTransactionsInternalToAddress(baseURL, address).then( (res) => {
        return getTransactionsDetail(baseURL, res["transactions"]).then( (txs) => {
            return {
                data: txs,
                total: res["total"]
            }
        })
    })
};

export const getEvents = (baseURL, address, options) => {
    return getAllEventsFromAddress(baseURL, address, options).then( (res) => {
        return {
            data: res["events"].map( (event) => ({
                topic: event.rawEvent.topics[0],
                txHash: event.rawEvent.transactionHash,
                address: event.rawEvent.address,
                blockNumber: event.rawEvent.blockNumber,
                parsedEvent: {
                    eventSig: event.eventSig,
                    parsedData: event.parsedData,
                },
            })),
            total: res["total"]
        }
    })
};

export const getReportData = (baseURL, address, startBlockNumber, endBlockNumber, options) => {
    return getStorageHistoryCount(baseURL, address, startBlockNumber, endBlockNumber)
      .then(res => {
          const { ranges } = res
          const total = ranges.reduce((sum, range) => sum + range.resultCount, 0)
          const pagesPerRange = 1000 / options.pageSize
          const rangeIndex = Math.floor((options.pageNumber * options.pageSize) / 1000)
          const range = ranges[rangeIndex]
          const pageNumberWithinRange = options.pageNumber % pagesPerRange

          return getStorageHistory(baseURL, address, {
              pageSize: options.pageSize,
              pageNumber: pageNumberWithinRange,
              beginBlockNumber: range.start,
              endBlockNumber: range.end,
          }).then((res) => {
              return {
                  data: res.historicState,
                  total: total
              }
          })
      })
}

function calculateTotal (result, options) {
    let lastPage = result.length < options.pageSize
    let currentTotal = options.pageSize * options.pageNumber + result.length
    // -1 means total unknown, set current total to disable next on last page
    let total = lastPage ? currentTotal : -1
    return total
}

export const getERC20Holders = (baseURL, address, block, options) => {
    return getERC20TokenHolders(baseURL, address, block, options).then( (res) => {
        let total = calculateTotal(res, options)
        return Promise.all(res.map((holder) => {
            return getERC20Balance(baseURL, address, holder, block, block, {})
              .then((balanceRes) => {
                  const value = balanceRes.data && balanceRes.data[0].balance
                  return {
                      holder,
                      value,
                  }
              })
        }))
          .then((balances) => {
              return {
                  data: balances,
                  total
              }
          })
    })
};

export const getERC20Balance = (baseURL, address, holder, startBlockNumber, endBlockNumber, options) => {
    return getERC20TokenBalance(baseURL, address, holder, startBlockNumber, endBlockNumber, options).then( (res) => {
        let data = Object.entries(res).map(([key, value]) => ({ block: key, balance: value })).sort((one, two) => two.block - one.block)
        let total = calculateTotal(res, options)
        return {
            data: data,
            total,
        }
    })
};

export const getERC721Holders = (baseURL, address, block, options) => {
    return getERC721TokenHolders(baseURL, address, block, options).then( (res) => {
        let total = calculateTotal(res, options)
        return Promise.all(res.map((holder) => {
            return getERC721TokensForAccount(baseURL, address, holder, block, block, {})
              .then((tokensRes) => {
                  const value = tokensRes.data && tokensRes.data.length
                  return {
                      holder,
                      value,
                  }
              })
        }))
          .then((balances) => {
              return {
                  data: balances,
                  total
              }
          })
    })
};

export const getERC721Tokens = (baseURL, address, block, options) => {
    return getERC721TokensAtBlock(baseURL, address, block, options).then( (res) => {
        let total = calculateTotal(res, options)
        return {
            data: res,
            total
        }
    })
};

export const getERC721TokensForAccount = (baseURL, address, holder, block, options) => {
    return getERC721TokensForAccountAtBlock(baseURL, address, holder, block, options).then( (res) => {
        let total = calculateTotal(res, options)
        return {
            data: res,
            total
        }
    })
};

export const getHolderForERC721Token = (baseURL, address, tokenId, block) => {
    return getHolderForERC721TokenAtBlock(baseURL, address, tokenId, block).then( (res) => {
        return {
            data: [{ holder: res.replace('0x0x', '0x'), value: tokenId}], // TODO remove this when fixed
            total: 1,
        }
    })
};


export const getSingleBlock = (baseURL, blockNumber) => {
    return getBlock(baseURL, blockNumber).then( (res) => {
        return res
    })
};

export const getSingleTransaction = (baseURL, txHash) => {
    return getTransaction(baseURL, txHash).then( (res) => {
        return {
            txSig: res.txSig,
            func4Bytes: res.func4Bytes,
            parsedData: res.parsedData,
            parsedEvents: res.parsedEvents,
            ...res.rawTransaction
        }
    })
};

const getContractsDetail = (baseURL, addresses) => {
  return Promise.all(
    addresses.map((address) => {
        return Promise.all([
          getABI(baseURL, address).then((res) => res),
          getStorageABI(baseURL, address).then((res) => res),
          getContractTemplate(baseURL, address).then((res) => res),
          ]
        ).then(([abi, storageLayout, name]) => {
            return { address, abi, storageLayout, name }
        })
    })
  )
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
            hash: res.rawTransaction.hash,
            from: res.rawTransaction.from,
            to: res.rawTransaction.to,
            blockNumber: res.rawTransaction.blockNumber,
            parsedTransaction: {
                txSig: res.txSig,
                func4Bytes: res.func4Bytes,
                parsedData: res.parsedData,
            },
            parsedEvents: res.parsedEvents,
            internalCalls: res.rawTransaction.internalCalls,
        }
    })
};
