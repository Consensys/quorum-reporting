import React from 'react'
import makeStyles from '@material-ui/core/styles/makeStyles'
import { BrowserRouter as Router, Route, Switch } from 'react-router-dom'
import { useSelector } from 'react-redux'
import CircularProgress from '@material-ui/core/CircularProgress'
import HeaderContainer from './containers/HeaderContainer'
import BlockDetail from './containers/BlockDetail'
import TransactionDetail from './containers/TransactionDetail'
import ContractDetail from './containers/ContractDetail'
import ContractListContainer from './containers/ContractListContainer'

const useStyles = makeStyles(() => ({
  root: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
  },
  loading: {
    margin: 24,
  },
}))

function App() {
  const classes = useStyles()
  const ready = useSelector(
    (state) => state.system.isConnected && state.system.lastPersistedBlockNumber !== undefined,
  )
  return (
    <Router>
      <div className={classes.root}>
        <HeaderContainer />
        {/* A <Switch> looks through its children <Route>s and
              renders the first one that matches the current URL. */}
        {!ready && <CircularProgress className={classes.loading} />}
        {ready
        && (
          <Switch>
            <Route
              path="/blocks/:id"
              render={({ match }) => <BlockDetail number={match.params.id} />}
            />
            <Route
              path="/transactions/:id"
              render={({ match }) => <TransactionDetail id={match.params.id} />}
            />
            <Route
              path="/contracts/:id"
              render={({ match }) => <ContractDetail address={match.params.id} />}
            />

            <Route path="/">
              <ContractListContainer />
            </Route>
          </Switch>
        )}
      </div>
    </Router>
  )
}

export default App
