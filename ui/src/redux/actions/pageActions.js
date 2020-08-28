import * as types from '../actionTypes';

export const changePageAction = (page) => ({
    type: types.CHANGE_PAGE,
    page,
});