import axios from 'axios';

// lower level RPC services, should only be used by fetcher.js

let requestCount = 0;

export const getLastPersistedBlockNumber = (baseURL) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: "reporting.GetLastPersistedBlockNumber",
        params: [],
        id: requestCount++,
    })
};

export const getAddresses = (baseURL) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: "reporting.GetAddresses",
        params: [],
        id: requestCount++,
    })
};

export const addAddress = (baseURL, address) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: "reporting.AddAddress",
        params: [{address}],
        id: requestCount++,
    })
};

export const deleteAddress = (baseURL, address) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: "reporting.DeleteAddress",
        params: [address],
        id: requestCount++,
    })
};

export const getABI = (baseURL, address) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: "reporting.GetABI",
        params: [address],
        id: requestCount++,
    })
};

export const getStorageABI = (baseURL, address) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: "reporting.GetStorageABI",
        params: [address],
        id: requestCount++,
    })
};

export const addABI = (baseURL, address, abi) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: "reporting.AddABI",
        params: [{address, data: abi}],
        id: requestCount++,
    })
};

export const addStorageABI = (baseURL, address, template) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: "reporting.AddStorageABI",
        params: [{address, data: template}],
        id: requestCount++,
    })
};

export const getContractCreationTransaction = (baseURL, address) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: "reporting.GetContractCreationTransaction",
        params: [address],
        id: requestCount++,
    })
};

export const getAllTransactionsToAddress = (baseURL, address) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: "reporting.GetAllTransactionsToAddress",
        params: [{address, options: {pageSize: 1000}}], //TODO: use very large page size temporarily
        id: requestCount++,
    })
};

export const getAllTransactionsInternalToAddress = (baseURL, address) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: "reporting.GetAllTransactionsInternalToAddress",
        params: [{address, options: {pageSize: 1000}}], //TODO: use very large page size temporarily
        id: requestCount++,
    })
};

export const getAllEventsFromAddress = (baseURL, address) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: "reporting.GetAllEventsFromAddress",
        params: [{address, options: {pageSize: 1000}}], //TODO: use very large page size temporarily
        id: requestCount++,
    })
};

export const getBlock = (baseURL, blockNumber) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: "reporting.GetBlock",
        params: [blockNumber],
        id: requestCount++,
    })
};

export const getTransaction = (baseURL, txHash) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: "reporting.GetTransaction",
        params: [txHash],
        id: requestCount++,
    })
};

export const getStorageHistory = (baseURL, address, startBlockNumber, endBlockNumber, currentPage) => {
    return axios.post(baseURL, {
        jsonrpc: '2.0',
        method: "reporting.GetStorageHistory",
        params: [{address, options:{beginBlockNumber: parseInt(startBlockNumber), endBlockNumber: parseInt(endBlockNumber),pageNumber:currentPage, pageSize:10}}],
        id: requestCount++,
    })
};