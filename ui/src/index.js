import React from 'react'
import ReactDOM from 'react-dom'
import { Provider } from 'react-redux'
import CssBaseline from '@material-ui/core/CssBaseline'
import { createMuiTheme, MuiThemeProvider } from '@material-ui/core/styles'
import blue from '@material-ui/core/colors/blue'
import amber from '@material-ui/core/colors/amber'
import App from './App'
import store from './redux/store'

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
        },
        td: {
          maxWidth: 0,
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
          overflow: 'hidden',
        },
      },
    },
  },
})

function render(TheApp) {
  ReactDOM.render((
    <Provider store={store}>
      <MuiThemeProvider theme={theme}>
        <CssBaseline />
        <TheApp />
      </MuiThemeProvider>
    </Provider>
  ), document.getElementById('root'))
}

if (process.env.NODE_ENV === 'development' && module.hot) {
  module.hot.accept('./App', () => {
    // eslint-disable-next-line global-require
    const NextApp = require('./App').default
    render(NextApp)
  })
}

render(App)
