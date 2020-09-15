import React, { useEffect, useState } from 'react'
import Alert from '@material-ui/lab/Alert'
import { useDispatch, useSelector } from 'react-redux'
import { makeStyles } from '@material-ui/core/styles'
import shallowEqual from 'react-redux/lib/utils/shallowEqual'
import Typography from '@material-ui/core/Typography'
import TextareaAutosize from '@material-ui/core/TextareaAutosize'
import ContractActions from '../components/ContractActions'
import Grid from '@material-ui/core/Grid'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import { PaginatedTableView } from '../components/table/PaginatedTableView'

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
}))

export function ContractDetail ({ address }) {

  const classes = useStyles()
  const [contractDetail, setContractDetail] = useState()
  const [errorMessage, setErrorMessage] = useState()
  const [searchReport, setSearchReport] = useState()
  const { contracts = [] } = useSelector(state => state.user, shallowEqual)
  const { rpcEndpoint, lastPersistedBlockNumber } = useSelector(state => state.system, shallowEqual)
  const dispatch = useDispatch()

  useEffect(() => {
    setContractDetail(contracts.find((contract) => contract.address === address))
  }, [address, contracts])

  return (
    <div className={classes.root}>
      <Grid container
            direction="row"
            justify="center"
            className={classes.grid} alignItems={'stretch'}>
        {errorMessage &&
        <Grid item xs={12}>
          <Alert severity="error" className={classes.alert}>{errorMessage}</Alert>
        </Grid>
        }
        {contractDetail &&
        <Grid item xs={12} md={8}>
          <Card className={classes.details}>
            <CardContent>
              <Typography variant={'h5'}>Contract Details</Typography>
              <br/>
              <Typography variant="caption" className={classes.label}>Type</Typography>
              <Typography variant={'h6'} className={classes.value}>{contractDetail.name}</Typography>
              <Typography variant="caption" className={classes.label}>Address</Typography>
              <Typography variant="h6" className={classes.value}>{contractDetail.address}</Typography>
              <Typography variant="caption" className={classes.label}>ABI</Typography>
              <TextareaAutosize
                readOnly
                rowsMax={4}
                style={{ fontSize: '14px', width: '720px', maxWidth: '100%' }}
                defaultValue={contractDetail.abi}
                className={classes.value}/>
              {contractDetail.storageLayout &&
              <div>
                <Typography variant="caption" className={classes.label}>Storage</Typography>,
                <TextareaAutosize
                  readOnly
                  rowsMax={4}
                  style={{ fontSize: '14px', width: '720px', maxWidth: '100%' }}
                  defaultValue={contractDetail.storageLayout}
                  className={classes.value}/>
              </div>
              }
            </CardContent>
          </Card>
        </Grid>
        }
        {contractDetail &&
        <Grid item xs={12} md={4}>
          <ContractActions onSearch={setSearchReport} contractDetail={contractDetail}/>
        </Grid>
        }
        {searchReport &&
        <Grid item xs={12}>
          <searchReport.View searchReport={searchReport} address={address} />
        </Grid>
        }
      </Grid>
    </div>
  )
}

