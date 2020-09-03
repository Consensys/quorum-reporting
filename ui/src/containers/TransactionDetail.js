import React, { useEffect, useState } from 'react'
import Alert from '@material-ui/lab/Alert'
import RecursiveInfoList from '../components/RecursiveInfoList'
import { getSingleBlock, getSingleTransaction } from '../client/fetcher'
import { useSelector } from 'react-redux'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme) => ({
  root: {
    marginTop: 10,
    marginBottom: 10,
  },
  alert: {
    marginTop: 5,
    width: 1000,
  }
}))

export function TransactionDetail ({ id }) {

  const classes = useStyles()
  const [displayData, setDisplayData] = useState()
  const [errorMessage, setErrorMessage] = useState()
  const { lastPersistedBlockNumber, rpcEndpoint, isConnected } = useSelector(state => state.system)

  useEffect(() => {
    setDisplayData(undefined)
    getSingleTransaction(rpcEndpoint, id).then((res) => {
      setDisplayData(res)
    }).catch((e) => {
      setErrorMessage(`Transaction not found (${e.toString()})`)
      setDisplayData(undefined)
    })
  }, [id])

  return (
    <div className={classes.root} align="center">
      {errorMessage &&
      <Alert severity="error" className={classes.alert}>{errorMessage}</Alert>
      }
      {displayData &&
      <RecursiveInfoList
        displayData={displayData}
        // handleReturn={handleReturn}
      />
      }
    </div>
  )
}
