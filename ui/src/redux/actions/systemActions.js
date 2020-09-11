import * as types from '../actionTypes';

export const connectAction = () => ({
    type: types.CONNECT,
});

export const disconnectAction = () => ({
    type: types.DISCONNECT,
});

export const updateEndpointAction = (rpcEndpoint) => ({
    type: types.UPDATE_ENDPOINT,
    rpcEndpoint,
});

export const updateBlockNumberAction = (lastPersistedBlockNumber) => ({
    type: types.UPDATE_BLOCK_NUMBER,
    lastPersistedBlockNumber,
});

export const updateRowsPerPageAction = (rowsPerPage) => ({
    type: types.UPDATE_ROWS_PER_PAGE,
    rowsPerPage,
});
