import * as types from '../actionTypes';

const initialState = {
    selectedContract: "",
    contracts: [],
};

const userReducer = (state=initialState, action) => {
    switch (action.type) {
        case types.SELECT_CONTRACT:
            return {
                ...state,
                selectedContract: action.selectedContract,
            };
        case types.GET_CONTRACTS:
            return {
                ...state,
                contracts: action.contracts,
            };
        default:
            return state
    }
};

export default userReducer