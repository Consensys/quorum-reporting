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
  getContractTemplate,
  getERC20TokenBalance,
  getERC20TokenHolders,
  getERC721TokenHolders,
  getERC721TokensAtBlock,
  getERC721TokensForAccountAtBlock,
  getHolderForERC721TokenAtBlock,
  getLastPersistedBlockNumber,
  getStorageABI,
  getStorageHistory,
  getStorageHistoryCount,
  getTransaction,
} from './rpcClient'

export function getBlockNumber() {
  return getLastPersistedBlockNumber()
}

export function addContract(newContract) {
  return addAddress(newContract.address)
    .then(() => {
      if (newContract.template === 'new') {
        return addTemplate(newContract.newTemplate)
          .then(() => assignTemplate(newContract.address, newContract.newTemplate.name))
      }
      return assignTemplate(newContract.address, newContract.template)
    })
}

export function deleteContract(address) {
  return deleteAddress(address)
}

export function getContracts() {
  return getAddresses()
    .then((res) => {
      return getContractsDetail(res)
        .then((contracts) => {
          // sort by template name + address for consistent order
          return contracts.sort((a, b) => `${a.name}${a.address}`.localeCompare(`${b.name}${b.address}`))
        })
    })
}

export function getContractCreationTx(address) {
  return getContractCreationTransaction(address)
}

export function getToTxs(address, options) {
  return getAllTransactionsToAddress(address, options)
    .then((res) => {
      return getTransactionsDetail(res.transactions)
        .then((txs) => {
          return {
            data: txs,
            total: res.total,
          }
        })
    })
}

export function getInternalToTxs(address, options) {
  return getAllTransactionsInternalToAddress(address, options)
    .then((res) => {
      return getTransactionsDetail(res.transactions)
        .then((txs) => {
          return {
            data: txs,
            total: res.total,
          }
        })
    })
}

export function getEvents(address, options) {
  return getAllEventsFromAddress(address, options)
    .then((res) => {
      return {
        data: res.events.map((event) => ({
          topic: event.rawEvent.topics[0],
          txHash: event.rawEvent.transactionHash,
          address: event.rawEvent.address,
          blockNumber: event.rawEvent.blockNumber,
          parsedEvent: {
            eventSig: event.eventSig,
            parsedData: event.parsedData,
          },
        })),
        total: res.total,
      }
    })
}

export function getReportData(address, startBlockNumber, endBlockNumber, options) {
  return getStorageHistoryCount(address, startBlockNumber, endBlockNumber)
    .then(({ ranges }) => {
      const total = ranges.reduce((sum, range) => sum + range.resultCount, 0)
      const pagesPerRange = 1000 / options.pageSize
      const rangeIndex = Math.floor((options.pageNumber * options.pageSize) / 1000)
      const range = ranges[rangeIndex]
      const pageNumberWithinRange = options.pageNumber % pagesPerRange

      return getStorageHistory(address, {
        pageSize: options.pageSize,
        pageNumber: pageNumberWithinRange,
        beginBlockNumber: range.start,
        endBlockNumber: range.end,
      })
        .then((res) => {
          return {
            data: res.historicState,
            total,
          }
        })
    })
}

function calculateTotal(result, options) {
  const lastPage = result.length < options.pageSize
  const currentTotal = options.pageSize * options.pageNumber + result.length
  // -1 means total unknown, set current total to disable next on last page
  const total = lastPage ? currentTotal : -1
  return total
}

export function getERC20Holders(address, block, options) {
  return getERC20TokenHolders(address, block, options)
    .then((res) => {
      const total = calculateTotal(res, options)
      return Promise.all(res.map((holder) => {
        return getERC20Balance(address, holder, block, block, {})
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
            total,
          }
        })
    })
}

export function getERC20Balance(address, holder, startBlockNumber, endBlockNumber, options) {
  return getERC20TokenBalance(address, holder, startBlockNumber, endBlockNumber, options)
    .then((res) => {
      const data = Object.entries(res)
        .map(([key, value]) => ({
          block: key,
          balance: value,
        }))
        .sort((one, two) => two.block - one.block)
      const total = calculateTotal(res, options)
      return {
        data,
        total,
      }
    })
}

export function getERC721Holders(address, block, options) {
  return getERC721TokenHolders(address, block, options)
    .then((res) => {
      const total = calculateTotal(res, options)
      return Promise.all(res.map((holder) => {
        return getERC721TokensForAccount(address, holder, block, block)
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
            total,
          }
        })
    })
}

export function getERC721Tokens(address, block, options) {
  return getERC721TokensAtBlock(address, block, options)
    .then((res) => {
      const total = calculateTotal(res, options)
      return {
        data: res,
        total,
      }
    })
}

export function getERC721TokensForAccount(address, holder, block, options) {
  return getERC721TokensForAccountAtBlock(address, holder, block, options)
    .then((res) => {
      const total = calculateTotal(res, options)
      return {
        data: res,
        total,
      }
    })
}

export function getHolderForERC721Token(address, tokenId, block) {
  return getHolderForERC721TokenAtBlock(address, tokenId, block)
    .then((res) => {
      return {
        data: [
          {
            holder: res,
            value: tokenId,
          },
        ],
        total: 1,
      }
    })
}

export function getSingleBlock(blockNumber) {
  return getBlock(blockNumber)
}

export function getSingleTransaction(txHash) {
  return getTransaction(txHash)
    .then((res) => {
      return {
        txSig: res.txSig,
        func4Bytes: res.func4Bytes,
        parsedData: res.parsedData,
        parsedEvents: res.parsedEvents,
        ...res.rawTransaction,
      }
    })
}

function getContractsDetail(addresses) {
  return Promise.all(
    addresses.map((address) => {
      return Promise.all([
        getABI(address)
          .then((res) => res),
        getStorageABI(address)
          .then((res) => res),
        getContractTemplate(address)
          .then((res) => res),
      ])
        .then(([abi, storageLayout, name]) => {
          return {
            address,
            abi,
            storageLayout,
            name,
          }
        })
    }),
  )
}

function getTransactionsDetail(txs) {
  return Promise.all(
    txs.map((hash) => getTransactionDetail(hash)),
  )
}

function getTransactionDetail(txHash) {
  return getTransaction(txHash)
    .then((res) => {
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
}
