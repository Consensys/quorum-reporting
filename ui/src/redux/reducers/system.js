import * as types from '../actionTypes'

const initialState = {
  rpcEndpoint: 'http://localhost:4000',
  isConnected: false,
  lastPersistedBlockNumber: '',
  rowsPerPage: 25,
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
    default:
      return state
  }
}

export default systemReducer
