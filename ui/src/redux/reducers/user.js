import * as types from '../actionTypes';
import { HomePageId } from "../../constants";

const initialState = {
    page: HomePageId,
    selectedContract: "",
    contracts: [],
};

const userReducer = (state=initialState, action) => {
    switch (action.type) {
        case types.CHANGE_PAGE:
            return {
                ...state,
                page: action.page,
            };
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