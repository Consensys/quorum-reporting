import React, { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import AppBar from '@material-ui/core/AppBar'
import Toolbar from '@material-ui/core/Toolbar'
import Typography from '@material-ui/core/Typography'
import IconButton from '@material-ui/core/IconButton'
import SettingsIcon from '@material-ui/icons/Settings'
import SettingForm from '../components/SettingForm'
import {
  connectAction,
  disconnectAction,
  updateBlockNumberAction,
  updateEndpointAction
} from '../redux/actions/systemActions'
import { getBlockNumber, getContracts } from '../client/fetcher'
import SearchField from '../components/SearchField'
import { makeStyles } from '@material-ui/styles'
import { shallowEqual, useDispatch, useSelector } from 'react-redux'
import { getContractsAction } from '../redux/actions/contractActions'
import { Lens } from '@material-ui/icons'

const useStyles = makeStyles((theme) => ({
  grow: {
    flexGrow: 1,
  },
  home: {
    display: 'flex',
    flexDirection: 'row',
    alignItems: 'center',
    textDecoration: 'none',
    color: 'inherit',
    marginRight: 16,
  },
  homeText: {
    paddingLeft: 12,
    paddingRight: 12,
    fontSize: 20,
  },
  link: {
    textDecoration: 'none',
    color: 'inherit',
    margin: 12,
  },
}))

export default function HeaderContainer () {

  const classes = useStyles()
  const [formIsOpen, setFormIsOpen] = useState(false)
  const [newRPCEndpoint, setNewRPCEndpoint] = useState('')
  const {
    rpcEndpoint,
    isConnected,
    lastPersistedBlockNumber,
  } = useSelector(state => state.system, shallowEqual)
  const dispatch = useDispatch()

  useEffect(() => {
    getContracts(rpcEndpoint).then( (contracts) => {
      dispatch(getContractsAction(contracts))
    })
    const timerID = setInterval(
      () => connectReporting(),
      1000
    )
    return () => {
      clearInterval(timerID)
    }
  }, [])

  const connectReporting = () => {
    getBlockNumber(rpcEndpoint).then((res) => {
      if (lastPersistedBlockNumber !== res) {
        if (!isConnected) {
          dispatch(connectAction())
        }
        dispatch(updateBlockNumberAction(res))
      }
    }).catch((e) => {
      if (isConnected) {
        dispatch(disconnectAction())
        dispatch(updateBlockNumberAction(''))
      }
    })
  }

  return (
    <AppBar color="transparent" position="static">
      <Toolbar>
        <Link to="/" className={classes.home}>
          <img src={require('../resources/quorum-logo.png')} width="40" height="20" alt=""/>
          <Typography className={classes.homeText}>
            Quorum Reporting
          </Typography>
        </Link>
        <span className={classes.grow}/>
        <SearchField/>
        <Lens style={{ fontSize: 16, color: isConnected ? 'green' : 'red', margin: 6 }}/>
        <Typography variant="h5" color="inherit">
          {isConnected ? ('#' + lastPersistedBlockNumber) : '#N/A'}
          &nbsp;
        </Typography>
        <IconButton variant="h4" onClick={() => setFormIsOpen(true)}>
          <SettingsIcon color="action"/>
        </IconButton>
        <SettingForm
          rpcEndpoint={rpcEndpoint}
          isOpen={formIsOpen}
          handleCloseSetting={() => setFormIsOpen(false)}
          handleRPCEndpointChange={(e) => setNewRPCEndpoint(e.target.value)}
          handleSetRPCEndpoint={() => {
            dispatch(updateEndpointAction(newRPCEndpoint))
            connectReporting()
            setFormIsOpen(false)
          }}
          newRPCEndpoint={newRPCEndpoint}
        />
      </Toolbar>
    </AppBar>
  )
}

