import React from 'react'
import ReactDOM from 'react-dom'
import { Provider } from 'react-redux'
import store from './redux/store'
import App from './App'
import CssBaseline from '@material-ui/core/CssBaseline'
import { createMuiTheme, MuiThemeProvider } from '@material-ui/core/styles'
import blue from '@material-ui/core/colors/blue'
import amber from '@material-ui/core/colors/amber'

const theme = createMuiTheme({
  palette: {
    primary: blue,
    secondary: amber,
  },
  overrides: {
    MuiCssBaseline: {
      '@global': {
        a: {
          textDecoration: 'none',
          color: blue[700],
        },
        '.MuiTooltip-tooltip': {
          fontSize: 14,
        }
      },
    },
  },
})

function render (TheApp) {
  ReactDOM.render((
    <Provider store={store}>
      <MuiThemeProvider theme={theme}>
        <CssBaseline/>
        <App/>
      </MuiThemeProvider>
    </Provider>
  ), document.getElementById('root'))
}

if (process.env.NODE_ENV === 'development' && module.hot) {
  module.hot.accept('./App', () => {
    const NextApp = require('./App').default
    render(NextApp)
  })
}

render(App)
