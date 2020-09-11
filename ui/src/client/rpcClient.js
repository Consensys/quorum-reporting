import axios from 'axios'

// lower level RPC services, should only be used by fetcher.js

let requestCount = 0;

export const request = (baseURL, method, params) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: method,
        params: params,
        id: requestCount++,
    })
};

export const getLastPersistedBlockNumber = (baseURL) => {
    return request(baseURL,
        "reporting.GetLastPersistedBlockNumber",
        [],
    )
};

export const getAddresses = (baseURL) => {
    return request(baseURL,
        "reporting.GetAddresses",
        [],
    )
};

export const addAddress = (baseURL, address) => {
    return request(baseURL,
        "reporting.AddAddress",
        [{address}],
    )
};

export const deleteAddress = (baseURL, address) => {
    return request(baseURL,
        "reporting.DeleteAddress",
        [address],
    )
};

export const getTemplates = (baseURL) => {
    return request(baseURL,
        "reporting.GetTemplates",
        [],
    )
};

export const addTemplate = (baseURL, newTemplate) => {
    return request(baseURL,
        "reporting.AddTemplate",
        [newTemplate],
    )
};

export const assignTemplate = (baseURL, address, templateName) => {
    return request(baseURL,
        "reporting.AssignTemplate",
        [{address: address, data: templateName}],
    )
};

export const getContractTemplate = (baseURL, address) => {
    return request(baseURL,
        "reporting.GetContractTemplate",
        [address],
    )
};

export const getABI = (baseURL, address) => {
    return request(baseURL,
        "reporting.GetABI",
        [address],
    )
};

export const getStorageABI = (baseURL, address) => {
    return request(baseURL,
        "reporting.GetStorageABI",
        [address],
    )
};

export const getContractCreationTransaction = (baseURL, address) => {
    return request(baseURL,
        "reporting.GetContractCreationTransaction",
        [address],
    )
};

export const getAllTransactionsToAddress = (baseURL, address, options) => {
    return request(baseURL,
        "reporting.GetAllTransactionsToAddress",
        [{address, options}],
    )
};

export const getAllTransactionsInternalToAddress = (baseURL, address, options) => {
    return request(baseURL,
        "reporting.GetAllTransactionsInternalToAddress",
        [{address, options}],
    )
};

export const getAllEventsFromAddress = (baseURL, address, options) => {
    return request(baseURL,
        "reporting.GetAllEventsFromAddress",
        [{address, options}],
    )
};

export const getBlock = (baseURL, blockNumber) => {
    return request(baseURL,
        "reporting.GetBlock",
        [blockNumber],
    )
};

export const getTransaction = (baseURL, txHash) => {
    return request(baseURL,
        "reporting.GetTransaction",
        [txHash],
    )
};

export const getStorageHistory = (baseURL, address, startBlockNumber, endBlockNumber, options) => {
    return request(baseURL,
        "reporting.GetStorageHistory",
        [{address, options:{...options, beginBlockNumber: parseInt(startBlockNumber), endBlockNumber: parseInt(endBlockNumber)}}],
    )
};

export const getERC20TokenHolders = (baseURL, address, block, options) => {
    return request(baseURL,
      "token.GetERC20TokenHoldersAtBlock",
      [{contract: address, block: parseInt(block), options}],
    )
};

export const getERC20TokenBalance = (baseURL, address, holder, startBlockNumber, endBlockNumber, options) => {
    return request(baseURL,
        "token.GetERC20TokenBalance",
      [{
          contract: address,
          holder,
          options: {
              ...options,
              beginBlockNumber: parseInt(startBlockNumber),
              endBlockNumber: parseInt(endBlockNumber),
              after: undefined, // use page numbers, not after value
          },
      }],
    )
};

export const getERC721TokenHolders = (baseURL, address, block, options) => {
    return request(baseURL,
      "token.AllERC721HoldersAtBlock",
      [{contract: address, block: parseInt(block), options}],
    )
};

export const getERC721TokensAtBlock = (baseURL, address, block, options) => {
    return request(baseURL,
      "token.AllERC721TokensAtBlock",
      [{contract: address, block: parseInt(block),
          options: {
              ...options,
              after: options.after ? options.after.tokenId : undefined,
          },
      }],
    )
};

export const getERC721TokensForAccountAtBlock = (baseURL, address, holder, block, options) => {
    return request(baseURL,
      "token.ERC721TokensForAccountAtBlock",
      [{contract: address, holder: holder, block: parseInt(block),
          options: {
              ...options,
              after: options.after ? options.after.tokenId : undefined,
          },
      }],
    )
};
export const getHolderForERC721TokenAtBlock = (baseURL, address, tokenId, block) => {
    return request(baseURL,
      "token.GetHolderForERC721TokenAtBlock",
      [{contract: address, tokenId: parseInt(tokenId), block: parseInt(block)}],
    )
};
