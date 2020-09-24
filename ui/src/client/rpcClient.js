import axios from 'axios'

// lower level RPC services, should only be used by fetcher.js

export const DEFAULT_RPC_URL = 'http://localhost:4000'

let baseUrl = DEFAULT_RPC_URL
let requestCount = 0

export function setBaseUrl(url) {
  baseUrl = url
}

export function request(method, params) {
  return axios.post(
    baseUrl,
    {
      jsonrpc: '2.0',
      method,
      params,
      // eslint-disable-next-line no-plusplus
      id: requestCount++,
    },
  )
    .then((res) => res.data.result)
    .catch((e) => {
      throw new Error(e.response.data.error)
    })
}

export function getLastPersistedBlockNumber() {
  return request('reporting.GetLastPersistedBlockNumber', [])
}

export function getAddresses() {
  return request('reporting.GetAddresses', [])
}

export function addAddress(address) {
  return request('reporting.AddAddress', [{ address }])
}

export function deleteAddress(address) {
  return request('reporting.DeleteAddress', [address])
}

export function getTemplates() {
  return request('reporting.GetTemplates', [])
}

export function addTemplate(name, abi, storageLayout) {
  return request('reporting.AddTemplate', [{ name, abi, storageLayout }])
}

export function assignTemplate(address, templateName) {
  return request('reporting.AssignTemplate', [
    {
      address,
      data: templateName,
    },
  ])
}

export function getContractTemplate(address) {
  return request('reporting.GetContractTemplate', [address])
}

export function getABI(address) {
  return request('reporting.GetABI', [address])
}

export function getStorageABI(address) {
  return request('reporting.GetStorageABI', [address])
}

export function getContractCreationTransaction(address) {
  return request('reporting.GetContractCreationTransaction', [address])
}

export function getAllTransactionsToAddress(address, options) {
  return request('reporting.GetAllTransactionsToAddress', [
    {
      address,
      options,
    },
  ])
}

export function getAllTransactionsInternalToAddress(address, options) {
  return request('reporting.GetAllTransactionsInternalToAddress', [
    {
      address,
      options,
    },
  ])
}

export function getAllEventsFromAddress(address, options) {
  return request('reporting.GetAllEventsFromAddress', [
    {
      address,
      options,
    },
  ])
}

export function getBlock(blockNumber) {
  return request('reporting.GetBlock', [blockNumber])
}

export function getTransaction(txHash) {
  return request('reporting.GetTransaction', [txHash])
}

export function getStorageHistory(address, options) {
  return request('reporting.GetStorageHistory', [
    {
      address,
      options,
    },
  ])
}

export function getStorageHistoryCount(address, startBlockNumber, endBlockNumber) {
  return request('reporting.GetStorageHistoryCount', [
    {
      address,
      options: {
        beginBlockNumber: parseInt(startBlockNumber, 10),
        endBlockNumber: parseInt(endBlockNumber, 10),
      },
    },
  ])
}

export function getERC20TokenHolders(address, block, options) {
  return request('token.GetERC20TokenHoldersAtBlock', [
    {
      contract: address,
      block: parseInt(block, 10),
      options,
    },
  ])
}

export function getERC20TokenBalance(
  address, holder, startBlockNumber, endBlockNumber, options,
) {
  return request('token.GetERC20TokenBalance', [
    {
      contract: address,
      holder,
      options: {
        ...options,
        beginBlockNumber: parseInt(startBlockNumber, 10),
        endBlockNumber: parseInt(endBlockNumber, 10),
        after: undefined, // use page numbers, not after value
      },
    },
  ])
}

export function getERC721TokenHolders(address, block, options) {
  return request('token.AllERC721HoldersAtBlock', [
    {
      contract: address,
      block: parseInt(block, 10),
      options,
    },
  ])
}

export function getERC721TokensAtBlock(address, block, options) {
  return request('token.AllERC721TokensAtBlock', [
    {
      contract: address,
      block: parseInt(block, 10),
      options: {
        ...options,
        after: options.after ? options.after.tokenId : undefined,
      },
    },
  ])
}

export function getERC721TokensForAccountAtBlock(address, holder, block, options) {
  return request('token.ERC721TokensForAccountAtBlock', [
    {
      contract: address,
      holder,
      block: parseInt(block, 10),
      options: {
        ...options,
        after: options.after ? options.after.tokenId : undefined,
      },
    },
  ])
}

export function getHolderForERC721TokenAtBlock(address, tokenId, block) {
  return request('token.GetHolderForERC721TokenAtBlock', [
    {
      contract: address,
      tokenId: parseInt(tokenId, 10),
      block: parseInt(block, 10),
    },
  ])
}
