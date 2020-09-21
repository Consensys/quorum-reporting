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
        if(newContract.template === 'new') {
            return addTemplate(baseURL, newContract.newTemplate)
              .then((template) => assignTemplate(baseURL, newContract.address, newContract.newTemplate.name))
        }
        return assignTemplate(baseURL, newContract.address, newContract.template)
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
        if (res.data.error) {
            throw res.data.error.message
        }
        return res.data.result
    })
};

export const getToTxs = (baseURL, address, options) => {
    return getAllTransactionsToAddress(baseURL, address, options).then( (res) => {
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

export const getEvents = (baseURL, address, options) => {
    return getAllEventsFromAddress(baseURL, address, options).then( (res) => {
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

export const getReportData = (baseURL, address, startBlockNumber, endBlockNumber, options) => {
    return getStorageHistoryCount(baseURL, address, startBlockNumber, endBlockNumber)
      .then(res => {
          if (res.data.error) {
              throw res.data.error.message
          }
          const { ranges } = res.data.result
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
              if (res.data.error) {
                  throw res.data.error.message
              }
              return {
                  data: res.data.result.historicState,
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
        if (res.data.error) {
            throw res.data.error.message
        }
        let total = calculateTotal(res.data.result, options)
        return Promise.all(res.data.result.map((holder) => {
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
        if (res.data.error) {
            throw res.data.error.message
        }
        let data = Object.entries(res.data.result).map(([key, value]) => ({ block: key, balance: value })).sort((one, two) => two.block - one.block)
        let total = calculateTotal(res.data.result, options)
        return {
            data: data,
            total,
        }
    })
};

export const getERC721Holders = (baseURL, address, block, options) => {
    return getERC721TokenHolders(baseURL, address, block, options).then( (res) => {
        if (res.data.error) {
            throw res.data.error.message
        }
        let total = calculateTotal(res.data.result, options)
        return Promise.all(res.data.result.map((holder) => {
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
        if (res.data.error) {
            throw res.data.error.message
        }
        let total = calculateTotal(res.data.result, options)
        return {
            data: res.data.result,
            total
        }
    })
};

export const getERC721TokensForAccount = (baseURL, address, holder, block, options) => {
    return getERC721TokensForAccountAtBlock(baseURL, address, holder, block, options).then( (res) => {
        if (res.data.error) {
            throw res.data.error.message
        }
        let total = calculateTotal(res.data.result, options)
        return {
            data: res.data.result,
            total
        }
    })
};

export const getHolderForERC721Token = (baseURL, address, tokenId, block) => {
    return getHolderForERC721TokenAtBlock(baseURL, address, tokenId, block).then( (res) => {
        if (res.data.error) {
            throw res.data.error.message
        }
        return {
            data: [{ holder: res.data.result.replace('0x0x', '0x'), value: tokenId}], // TODO remove this when fixed
            total: 1,
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
  return Promise.all(
    addresses.map((address) => {
        return Promise.all([
          getABI(baseURL, address).then((res) => res.data.result),
          getStorageABI(baseURL, address).then((res) => res.data.result),
          getContractTemplate(baseURL, address).then((res) => res.data.result),
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
