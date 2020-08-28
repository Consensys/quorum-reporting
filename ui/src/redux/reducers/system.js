import * as types from '../actionTypes';
import { DefaultRPCEndpoint } from '../../config';

const initialState = {
    rpcEndpoint: DefaultRPCEndpoint,
    isConnected: false,
    lastPersistedBlockNumber: "",
};

const systemReducer = (state=initialState, action) => {
    switch (action.type) {
        case types.CONNECT:
            return {
                ...state,
                isConnected: true,
            };
        case types.DISCONNECT:
            return {
                ...state,
                isConnected: false,
            };
        case types.UPDATE_ENDPOINT:
            return {
                ...state,
                rpcEndpoint: action.rpcEndpoint,
            };
        case types.UPDATE_BLOCK_NUMBER:
            return {
                ...state,
                lastPersistedBlockNumber: action.lastPersistedBlockNumber,
            };
        default:
            return state
    }
};

export default systemReducer