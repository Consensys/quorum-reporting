import React, { useState } from 'react'
import CardContent from '@material-ui/core/CardContent'
import Card from '@material-ui/core/Card'
import Button from '@material-ui/core/Button'
import SearchIcon from '@material-ui/icons/Search'
import ContractSelector from '../components/ContractSelector'
import { GenerateReport, GetContractCreationTx, GetEvents, GetInternalToTxs, GetToTxs } from '../constants'
import { makeStyles } from '@material-ui/core/styles'
import FormControl from '@material-ui/core/FormControl'
import TextField from '@material-ui/core/TextField'
import Alert from '@material-ui/lab/Alert'
import { useSelector } from 'react-redux'
import Typography from '@material-ui/core/Typography'

const useStyles = makeStyles((theme) => ({
  card: {
    marginTop: theme.spacing(0.5),
    marginBottom: theme.spacing(0.5),
    marginLeft: theme.spacing(1),
    marginRight: theme.spacing(1),
  },
  formControl: {
    margin: theme.spacing(1),
  }
}))

const actions = [GetContractCreationTx, GetToTxs, GetInternalToTxs, GetEvents, GenerateReport]

export default function ContractActions ({ onSearch }) {
  const classes = useStyles()
  const lastPersistedBlockNumber = useSelector(state => state.system.lastPersistedBlockNumber)
  const [selectedAction, setSelectedAction] = useState(GetToTxs)
  const [error, setError] = useState('')
  const [startNumber, setStartNumber] = useState(1)
  const [endNumber, setEndNumber] = useState(lastPersistedBlockNumber)
  return (
    <Card className={classes.card}>
      <CardContent>
        <Typography variant="h5">Reports</Typography>
        <div align="center">
          {
            error &&
            <div>
              <br/>
              <Alert severity="error">{error}</Alert>
            </div>
          }
          <br/>
          <ContractSelector
            actions={actions}
            selectedAction={selectedAction}
            handleSelectedActionChange={(e) => {
              setSelectedAction(e.target.value)
            }}
          />
          {selectedAction === GenerateReport &&
          <div>
            <FormControl className={classes.formControl}>
              <TextField
                label="Start Block Number"
                value={startNumber}
                onChange={(e) => setStartNumber(e.target.value)}
                variant="filled"
                size="small"
              />
            </FormControl>
            <FormControl className={classes.formControl}>
              <TextField
                label="End Block Number"
                value={endNumber}
                onChange={(e) => setEndNumber(e.target.value)}
                variant="filled"
                size="small"
              />
            </FormControl>
          </div>
          }
          <br/>
          <br/>
          <Button align="right" variant="contained" color="primary" onClick={() => {
            if(selectedAction === GenerateReport) {
              if (startNumber === "" || isNaN(startNumber)) {
                setError('Invalid start block number')
                return
              }
              if (endNumber === "" || isNaN(endNumber)) {
                setError('Invalid end block number')
                return
              }
            }
            onSearch(selectedAction, startNumber, endNumber)
          }}>
            <SearchIcon/>
            &nbsp;
            Search
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
