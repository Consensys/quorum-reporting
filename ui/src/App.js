import React from 'react';
import { MuiThemeProvider, createMuiTheme } from '@material-ui/core/styles';
import HomePage from './pages/HomePage';

const theme = createMuiTheme({});

class App extends React.Component {
    render() {
        return (
            <MuiThemeProvider theme={theme}>
                <HomePage />
            </MuiThemeProvider>
        )
    }
}

export default App