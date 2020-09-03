import React from 'react'
import { connect } from 'react-redux'
import HeaderContainer from '../containers/HeaderContainer'
import ContractInfoContainer from '../containers/ContractInfoContainer'
import ReportContainer from '../containers/ReportContainer'
import { BrowserRouter as Router, Route, Switch } from 'react-router-dom'
import { BlockDetail } from '../containers/BlockDetail'
import { TransactionDetail } from '../containers/TransactionDetail'
import { ContractDetail } from '../containers/ContractDetail'
import ContractListContainer from '../containers/ContractListContainer'
import makeStyles from '@material-ui/core/styles/makeStyles'

const useStyles = makeStyles((theme) => ({
    root: {
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    }

}))
function HomePage () {
    const classes = useStyles()
    return (
      <Router>
          <div className={classes.root}>
              <HeaderContainer/>
              {/* A <Switch> looks through its children <Route>s and
              renders the first one that matches the current URL. */}
              <Switch>
                  <Route path="/blocks/:id" render={({ match }) => <BlockDetail number={match.params.id}/>}/>
                  <Route path="/transactions/:id" render={({ match }) => <TransactionDetail id={match.params.id}/>}/>
                  <Route path="/contracts/:id" render={({ match }) => <ContractDetail address={match.params.id}/>}/>

                  <Route path="/">
                      <ContractListContainer/>
                  </Route>
              </Switch>
          </div>
      </Router>
    )
}

export default HomePage