import React, { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import AppBar from '@material-ui/core/AppBar'
import Toolbar from '@material-ui/core/Toolbar'
import Typography from '@material-ui/core/Typography'
import IconButton from '@material-ui/core/IconButton'
import SettingsIcon from '@material-ui/icons/Settings'
import SyncIcon from '@material-ui/icons/Sync'
import SyncDisabledIcon from '@material-ui/icons/SyncDisabled'
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

const useStyles = makeStyles((theme) => ({
  grow: {
    flexGrow: 1,
  },
  home: {
    textDecoration: 'none',
    color: 'inherit',
    marginRight: 16,
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
  const [blockNumberAppear, setBlockNumberAppear] = useState(false)
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

  const blockNumberBlinkEffect = () => {
    setBlockNumberAppear(false)
    setTimeout(() => {
      setBlockNumberAppear(true)
    }, 500)
  }

  const connectReporting = () => {
    getBlockNumber(rpcEndpoint).then((res) => {
      if (lastPersistedBlockNumber !== res) {
        blockNumberBlinkEffect()
        if (!isConnected) {
          dispatch(connectAction())
        }
        dispatch(updateBlockNumberAction(res))
      }
    }).catch((e) => {
      if (isConnected) {
        blockNumberBlinkEffect()
        dispatch(disconnectAction())
        dispatch(updateBlockNumberAction(''))
      }
    })
  }

  return (
    <AppBar position="static">
      <Toolbar>
        <Link to="/" className={classes.home}>
          <Typography variant="h6" color="inherit">
            <img src={require('../resources/quorum-logo.png')} width="40" height="20" alt=""/>
            &nbsp;
            Quorum Reporting
            &nbsp;
          </Typography>
        </Link>
        <span className={classes.grow}/>
        <SearchField/>
        <Typography variant="h4">
          {
            isConnected ? <SyncIcon color="inherit"/> : <SyncDisabledIcon color="error"/>
          }
        </Typography>
        {/*<Fade in={blockNumberAppear} timeout={1000}>*/}
        <Typography variant="h5" color="inherit">
          &nbsp;
          {isConnected ? ('# ' + lastPersistedBlockNumber) : '# N/A'}
          &nbsp;
        </Typography>
        {/*</Fade>*/}
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

