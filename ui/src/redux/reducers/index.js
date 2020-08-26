import { combineReducers } from 'redux';
import user from './user';
import system from './system';

const rootReducer = combineReducers({
	user, system
});

export default rootReducer