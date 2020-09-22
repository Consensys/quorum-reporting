import React, { useEffect, useState } from 'react'
import Alert from '@material-ui/lab/Alert'
import { useSelector } from 'react-redux'
import { makeStyles } from '@material-ui/core/styles'
import shallowEqual from 'react-redux/lib/utils/shallowEqual'
import Typography from '@material-ui/core/Typography'
import TextareaAutosize from '@material-ui/core/TextareaAutosize'
import Grid from '@material-ui/core/Grid'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import { Link } from 'react-router-dom'
import { getContractCreationTx } from '../client/fetcher'
import ContractActions from '../components/ContractActions'
import { getDefaultReportForTemplate } from '../reports'

const useStyles = makeStyles((theme) => ({
  root: {
    width: '100%',
  },
  grid: {
    maxWidth: 1280,
    margin: '0 auto',
  },
  alert: {
    marginTop: theme.spacing(0.5),
    marginBottom: theme.spacing(0.5),
    marginLeft: theme.spacing(1),
    marginRight: theme.spacing(1),
  },
  details: {
    display: 'flex',
    flexDirection: 'column',
    marginTop: theme.spacing(0.5),
    marginBottom: theme.spacing(0.5),
    marginLeft: theme.spacing(1),
    marginRight: theme.spacing(1),
  },
  label: {
    fontSize: 13,
    display: 'block',
  },
  value: {
    marginBottom: 8,
  },
  linkValue: {
    fontSize: 18,
    marginBottom: 8,
  },
}))

export default function ContractDetail({ address }) {
  const classes = useStyles()
  const [contractDetail, setContractDetail] = useState()
  const [errorMessage, setErrorMessage] = useState()
  const [searchReport, setSearchReport] = useState()
  const [creationTx, setCreationTx] = useState()
  const { contracts = [] } = useSelector((state) => state.user, shallowEqual)
  const { rpcEndpoint, lastPersistedBlockNumber } = useSelector(
    (state) => state.system,
    shallowEqual,
  )

  useEffect(() => {
    const detail = contracts.find((contract) => contract.address === address)
    if (!detail) {
      setErrorMessage(`No contract registered at ${address}`)
      return
    }
    setErrorMessage(undefined)
    setContractDetail(detail)
    const report = getDefaultReportForTemplate(detail.name)
    setSearchReport({
      ...report,
      params: {
        startNumber: 1,
        endNumber: lastPersistedBlockNumber,
        atBlock: lastPersistedBlockNumber,
      },
    })
    getContractCreationTx(rpcEndpoint, address)
      .then((transaction) => {
        setCreationTx(transaction)
      })
      .catch((e) => {
        console.log('Contract Creation Tx not found', e)
        setCreationTx(undefined)
      })
  }, [address, contracts])

  return (
    <div className={classes.root}>
      <Grid
        container
        direction="row"
        justify="center"
        className={classes.grid}
        alignItems="stretch"
      >
        {errorMessage
        && (
          <Grid item xs={12}>
            <Alert severity="error" className={classes.alert}>{errorMessage}</Alert>
          </Grid>
        )}
        {contractDetail
        && (
          <Grid item xs={12} md={8}>
            <Card className={classes.details}>
              <CardContent>
                <Typography variant="h5">Contract Details</Typography>
                <br />
                <Typography variant="caption" className={classes.label}>Type</Typography>
                <Typography
                  variant="h6"
                  className={classes.value}
                >
                  {contractDetail.name}
                </Typography>
                <Typography variant="caption" className={classes.label}>Address</Typography>
                <Typography
                  variant="h6"
                  className={classes.value}
                >
                  {contractDetail.address}
                </Typography>
                {creationTx
                && (
                  <div>
                    <Typography variant="caption" className={classes.label}>
                      Creation
                      Transaction
                    </Typography>
                    <Link to={`/transactions/${creationTx}`}>
                      <Typography variant="h6" className={classes.linkValue}>
                        {creationTx}
                      </Typography>
                    </Link>
                  </div>
                )}
                <Typography variant="caption" className={classes.label}>ABI</Typography>
                <TextareaAutosize
                  readOnly
                  rowsMax={4}
                  style={{
                    fontSize: '14px',
                    width: '720px',
                    maxWidth: '100%',
                  }}
                  defaultValue={contractDetail.abi}
                  className={classes.value}
                />
                {contractDetail.storageLayout
                && (
                  <div>
                    <Typography variant="caption" className={classes.label}>Storage</Typography>
                    <TextareaAutosize
                      readOnly
                      rowsMax={4}
                      style={{
                        fontSize: '14px',
                        width: '720px',
                        maxWidth: '100%',
                      }}
                      defaultValue={contractDetail.storageLayout}
                      className={classes.value}
                    />
                  </div>
                )}
              </CardContent>
            </Card>
          </Grid>
        )}
        {contractDetail
        && (
          <Grid item xs={12} md={4}>
            <ContractActions onSearch={setSearchReport} contractDetail={contractDetail} />
          </Grid>
        )}
        {searchReport
        && (
          <Grid item xs={12}>
            <searchReport.View searchReport={searchReport} address={address} />
          </Grid>
        )}
      </Grid>
    </div>
  )
}
