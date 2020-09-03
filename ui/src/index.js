import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';
import store from './redux/store';
import App from './App';
import CssBaseline from '@material-ui/core/CssBaseline'

ReactDOM.render((
    <Provider store={store}>
        <CssBaseline />
        <App />
    </Provider>
), document.getElementById('root'));