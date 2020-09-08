import React, { useEffect, useState } from 'react'
import Alert from '@material-ui/lab/Alert'
import TransactionResultTable from '../components/TransactionResultTable'
import EventResultTable from '../components/EventResultTable'
import { GenerateReport, GetContractCreationTx, GetEvents, GetInternalToTxs, GetToTxs } from '../constants'
import { getContractCreationTx, getEvents, getInternalToTxs, getReportData, getToTxs } from '../client/fetcher'
import { makeStyles } from '@material-ui/core/styles'
import { useSelector } from 'react-redux'
import Report from '../components/Report'
import Paper from '@material-ui/core/Paper'

const useStyles = makeStyles((theme) => ({
  card: {
    marginTop: theme.spacing(0.5),
    marginBottom: theme.spacing(0.5),
    marginLeft: theme.spacing(1),
    marginRight: theme.spacing(1),
  },
}))

const pageSize = 10

export default function ContractInfoContainer (props) {

  const classes = useStyles()
  const rpcEndpoint = useSelector(state => state.system.rpcEndpoint)
  const [displayTxResult, setDisplayTxResult] = useState(false)
  const [displayEventResult, setDisplayEventResult] = useState(false)
  const [displayReportResult, setDisplayReportResult] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [displayData, setDisplayData] = useState([])
  const [displayDataLength, setDisplayDataLength] = useState(0)
  const [currentPage, setCurrentPage] = useState(0)
  const [error, setError] = useState('')

  const handleChangePage = (event, newPage) => {
    setCurrentPage(newPage)
  }

  useEffect(() => {
    // clear display
    setDisplayTxResult(false)
    setDisplayEventResult(false)
    setDisplayReportResult(false)
    // check new search condition
    setIsLoading(true)
    searchByPage(props.address, props.action, currentPage)
  }, [props.address, props.action, currentPage])

  function handleError (e) {
    setDisplayData([])
    setDisplayDataLength(0)
    setIsLoading(false)
    setError(e.toString)
  }

  const searchByPage = (contract, action, pageNumber) => {
    if (action.action === GetEvents) {
      getEvents(rpcEndpoint, contract, { pageSize, pageNumber }).then((res) => {
        setDisplayData(res.data)
        setDisplayDataLength(res.total)
        setIsLoading(false)
      }).catch(handleError)

      setDisplayEventResult(true)
    } else if (action.action === GenerateReport) {
      getReportData(rpcEndpoint, props.address, action.startBlock, action.endBlock, currentPage).then( (res) => {
        setDisplayData(res.data)
        setDisplayDataLength(res.total)
        setIsLoading(false)
      })
      setDisplayReportResult(true)
    } else {
      switch (action.action) {
        case GetContractCreationTx:
          getContractCreationTx(rpcEndpoint, contract).then((res) => {
            setDisplayData([res])
            setDisplayDataLength(1)
            setIsLoading(false)
          }).catch(handleError)
          break
        case GetToTxs:
          getToTxs(rpcEndpoint, contract, { pageSize, pageNumber }).then((res) => {
            setDisplayData(res.data)
            setDisplayDataLength(res.total)
            setIsLoading(false)
          }).catch(handleError)
          break
        case GetInternalToTxs:
          getInternalToTxs(rpcEndpoint, contract, { pageSize, pageNumber }).then((res) => {
            setDisplayData(res.data)
            setDisplayDataLength(res.total)
            setIsLoading(false)
          }).catch(handleError)
          break
        default:
          setDisplayData([])
          setDisplayDataLength(0)
          setIsLoading(false)
      }
      setDisplayTxResult(true)
    }
  }

  return (
    <Paper className={classes.card}>
      {
        error &&
        <div>
          <br/>
          <Alert severity="error">{error}</Alert>
        </div>
      }
      {
        displayTxResult &&
        <TransactionResultTable
          displayData={displayData}
          isLoading={isLoading}
          currentPage={currentPage}
          pageSize={pageSize}
          totalTxs={displayDataLength}
          handleChangePage={handleChangePage}
        />
      }
      {
        displayEventResult &&
        <EventResultTable
          displayData={displayData}
          isLoading={isLoading}
          currentPage={currentPage}
          pageSize={pageSize}
          totalEvents={displayDataLength}
          handleChangePage={handleChangePage}
        />
      }
      {
        displayReportResult &&
        <Report
          parsedStorage={displayData}
          isLoading={isLoading}
          currentPage={currentPage}
          pageSize={pageSize}
          totalEvents={displayDataLength}
          handleChangePage={handleChangePage}
        />
      }
    </Paper>
  )
}
