import React, { useEffect, useState } from 'react'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import Typography from '@material-ui/core/Typography'
import IconButton from '@material-ui/core/IconButton'
import AddIcon from '@material-ui/icons/Add'
import RefreshIcon from '@material-ui/icons/Refresh'
import ContractTable from '../components/ContractTable'
import ContractForm from '../components/ContractForm'
import { getContractsAction } from '../redux/actions/contractActions'
import { deleteContract, getContracts } from '../client/fetcher'
import Button from '@material-ui/core/Button'
import { makeStyles } from '@material-ui/core/styles'
import { useDispatch, useSelector } from 'react-redux'

const useStyles = makeStyles((theme) => ({
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
  }
}))

export default function ContractListContainer() {
  const classes = useStyles()
  const dispatch = useDispatch()
  const [formIsOpen, setFormIsOpen] = useState(false)
  const rpcEndpoint = useSelector(state => state.system.rpcEndpoint)
  const contracts = useSelector(state => state.user.contracts)

  useEffect(() => {
    getAllRegisteredContract()
  }, [])

  const getAllRegisteredContract = () => {
    getContracts(rpcEndpoint).then((contracts) => {
      const sortedContracts = contracts.sort((a, b) => a.name.localeCompare(b.name))
      dispatch(getContractsAction(sortedContracts))
    })
  }

  const handleCloseSetting = () => {
    setFormIsOpen(false)
    // give a small timeout to avoid fetch too fast
    setTimeout(() => {
      getAllRegisteredContract()
    }, 500)
  }

  const handleContractDelete = (address) => {
    deleteContract(rpcEndpoint, address).then(() => {
      // TODO: handle error?
      // give a small timeout to avoid fetch too fast
      setTimeout(() => {
        getAllRegisteredContract()
      }, 500)
    })
  }

  return (
    <Card className={classes.card}>
      <CardContent className={classes.cardContent}>
        <Typography variant="h6" align="left">
          Registered Contract List&nbsp;
          <IconButton onClick={getAllRegisteredContract}>
            <RefreshIcon/>
          </IconButton>
        </Typography>
        <br/>
        {
          contracts.length === 0 &&
          <h1 align="center">No Contracts Registered</h1>
        }
        {
          contracts.length !== 0 &&
          <ContractTable
            contracts={contracts}
            handleContractDelete={handleContractDelete}
          />
        }
        <br/>
        <Button color="primary" onClick={() => {
          setFormIsOpen(true)
        }}>
          <AddIcon/>&nbsp;Add Contract
        </Button>
        <ContractForm
          isOpen={formIsOpen}
          handleCloseSetting={handleCloseSetting}
        />
      </CardContent>
    </Card>
  )
}
