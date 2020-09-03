import React, { useEffect, useState } from 'react'
import Alert from '@material-ui/lab/Alert'
import { useDispatch, useSelector } from 'react-redux'
import { makeStyles } from '@material-ui/core/styles'
import ContractInfoContainer from './ContractInfoContainer'
import ReportContainer from './ReportContainer'
import shallowEqual from 'react-redux/lib/utils/shallowEqual'
import Paper from '@material-ui/core/Paper'
import Typography from '@material-ui/core/Typography'
import TextareaAutosize from '@material-ui/core/TextareaAutosize'

const useStyles = makeStyles((theme) => ({
  root: {
    marginTop: 10,
    marginBottom: 10,
    maxWidth: 1080,
  },
  alert: {
    marginTop: 5,
    width: 1000,
  },
  details: {
    display: 'flex',
    flexDirection: 'column',
    padding: 24,
  },
  label: {
    fontSize: 13,
    marginBottom: 2,
  },
  value: {
    marginBottom: 8,
  },
}))

export function ContractDetail ({ address }) {

  const classes = useStyles()
  const [contractDetail, setContractDetail] = useState()
  const [errorMessage, setErrorMessage] = useState()
  const { lastPersistedBlockNumber, rpcEndpoint, isConnected } = useSelector(state => state.system, shallowEqual)
  const { contracts = [] } = useSelector(state => state.user, shallowEqual)
  const dispatch = useDispatch()

  useEffect(() => {
    console.log('contracts', contracts)
    setContractDetail(contracts.find((contract) => contract.address === address))
  }, [address, contracts])

  return (
    <div className={classes.root}>
      {errorMessage &&
      <Alert severity="error" className={classes.alert}>{errorMessage}</Alert>
      }
      {contractDetail &&
      <Paper className={classes.details}>
        <Typography variant="caption" className={classes.label}>Type</Typography>
        <Typography variant={'h6'} className={classes.value}>{contractDetail.name}</Typography>
        <Typography variant="caption" className={classes.label}>Address</Typography>
        <Typography variant="h6" className={classes.value}>{contractDetail.address}</Typography>
        <Typography variant="caption" className={classes.label}>ABI</Typography>
        <TextareaAutosize
          readOnly
          rowsMax={4}
          style={{ fontSize: '16px', width: '1000px' }}
          defaultValue={contractDetail.abi}
          className={classes.value}/>
        <Typography variant="caption" className={classes.label}>Storage</Typography>
        <TextareaAutosize
          readOnly
          rowsMax={4}
          style={{ fontSize: '16px', width: '1000px' }}
          defaultValue={contractDetail.storageLayout}
          className={classes.value}/>
      </Paper>
      }
      {contractDetail &&
      <ContractInfoContainer address={address}/>
      }
      {contractDetail &&
      <ReportContainer address={address}/>
      }
    </div>
  )
}
