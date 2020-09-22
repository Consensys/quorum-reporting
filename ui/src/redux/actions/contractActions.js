import * as types from '../actionTypes'

export const selectContractAction = (selectedContract) => ({
  type: types.SELECT_CONTRACT,
  selectedContract,
})

export const getContractsAction = (contracts) => ({
  type: types.GET_CONTRACTS,
  contracts,
})
