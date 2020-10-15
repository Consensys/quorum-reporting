import * as types from '../actionTypes'
import { DEFAULT_RPC_URL } from '../../client/rpcClient'

const initialState = {
  rpcEndpoint: DEFAULT_RPC_URL,
  isConnected: false,
  lastPersistedBlockNumber: undefined,
  rowsPerPage: 25,
  selectedContract: '',
  contracts: [],
}

const systemReducer = (state = initialState, action) => {
  switch (action.type) {
    case types.CONNECT:
      return {
        ...state,
        isConnected: true,
      }
    case types.DISCONNECT:
      return {
        ...state,
        isConnected: false,
      }
    case types.UPDATE_ENDPOINT:
      return {
        ...state,
        rpcEndpoint: action.rpcEndpoint,
      }
    case types.UPDATE_BLOCK_NUMBER:
      return {
        ...state,
        lastPersistedBlockNumber: action.lastPersistedBlockNumber,
      }
    case types.UPDATE_ROWS_PER_PAGE:
      return {
        ...state,
        rowsPerPage: action.rowsPerPage,
      }
    case types.GET_CONTRACTS:
      return {
        ...state,
        contracts: action.contracts,
      }
    default:
      return state
  }
}

export default systemReducer
