import React, { useEffect, useState } from 'react'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import Typography from '@material-ui/core/Typography'
import IconButton from '@material-ui/core/IconButton'
import AddIcon from '@material-ui/icons/Add'
import RefreshIcon from '@material-ui/icons/Refresh'
import Button from '@material-ui/core/Button'
import { makeStyles } from '@material-ui/core/styles'
import { useDispatch, useSelector } from 'react-redux'
import Alert from '@material-ui/lab/Alert'
import ContractTable from '../components/ContractTable'
import ContractForm from '../components/ContractForm'
import { getContractsAction } from '../redux/actions/systemActions'
import { deleteContract, getContracts } from '../client/fetcher'

const useStyles = makeStyles(() => ({
  card: {
    minWidth: 275,
    marginTop: 5,
    marginBottom: 5,
    width: '95%',
    maxWidth: 1080,
  },
  cardContent: {
    display: 'flex',
    flexDirection: 'column',
  },
}))

export default function ContractListContainer() {
  const classes = useStyles()
  const dispatch = useDispatch()
  const [formIsOpen, setFormIsOpen] = useState(false)
  const [errorMessage, setErrorMessage] = useState()
  const contracts = useSelector((state) => state.system.contracts)

  useEffect(() => {
    getAllRegisteredContracts()
  }, [])

  const getAllRegisteredContracts = () => {
    getContracts()
      .then((res) => dispatch(getContractsAction(res)))
      .catch((e) => {
        console.error('Could not fetch contracts', e)
      })
  }

  const handleCloseSetting = () => {
    setFormIsOpen(false)
    // give a small timeout to avoid fetch too fast
    setTimeout(() => {
      getAllRegisteredContracts()
    }, 500)
  }

  const handleContractDelete = (address) => {
    deleteContract(address)
      .then(() => {
        // give a small timeout to avoid fetch too fast
        setTimeout(() => {
          getAllRegisteredContracts()
        }, 500)
      })
      .catch((e) => {
        setErrorMessage(e.message)
      })
  }

  return (
    <Card className={classes.card}>
      <CardContent className={classes.cardContent}>
        {errorMessage
        && (
          <Alert
            severity="error"
            className={classes.alert}
            onClose={() => setErrorMessage(undefined)}
          >
            {errorMessage}
          </Alert>
        )}
        <Typography variant="h6" align="left">
          Registered Contract List&nbsp;
          <IconButton onClick={getAllRegisteredContracts}>
            <RefreshIcon />
          </IconButton>
        </Typography>
        <br />
        {
          contracts.length === 0
          && <h1 align="center">No Contracts Registered</h1>
        }
        {
          contracts.length !== 0
          && (
            <ContractTable
              contracts={contracts}
              handleContractDelete={handleContractDelete}
            />
          )
        }
        <br />
        <Button
          color="primary"
          onClick={() => {
            setFormIsOpen(true)
          }}
        >
          <AddIcon />
          &nbsp;Add Contract
        </Button>
        <ContractForm
          isOpen={formIsOpen}
          handleCloseSetting={handleCloseSetting}
        />
      </CardContent>
    </Card>
  )
}
