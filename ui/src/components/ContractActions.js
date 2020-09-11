import React, { useState } from 'react'
import CardContent from '@material-ui/core/CardContent'
import Card from '@material-ui/core/Card'
import Button from '@material-ui/core/Button'
import SearchIcon from '@material-ui/icons/Search'
import ContractSelector from '../components/ContractSelector'
import {
  Actions,
  GenerateReport,
  ContractCreationTx,
  ERC20TokenBalance,
  ERC20TokenHolders,
  ERC721HolderForToken,
  ERC721Holders,
  ERC721Tokens,
  ERC721TokensForAccount,
  Events,
  InternalToTxs,
  ToTxs,
} from '../constants'
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

export default function ContractActions ({ onSearch, contractDetail }) {
  let actions = [ToTxs, ContractCreationTx, InternalToTxs, Events]
  switch (contractDetail.name) {
    case 'ERC20':
      actions = [ERC20TokenHolders, ERC20TokenBalance, ...actions]
      break
    case 'ERC721':
      actions = [ERC721Holders, ERC721Tokens, ERC721TokensForAccount, ERC721HolderForToken, ...actions]
      break
    default:
      actions = [...actions, GenerateReport]
      break
  }
  const classes = useStyles()
  const lastPersistedBlockNumber = useSelector(state => state.system.lastPersistedBlockNumber)
  const [error, setError] = useState('')
  const [account, setAccount] = useState('')
  const [atBlock, setAtBlock] = useState(lastPersistedBlockNumber)
  const [tokenId, setTokenId] = useState('')
  const [startNumber, setStartNumber] = useState("1")
  const [endNumber, setEndNumber] = useState(lastPersistedBlockNumber)
  const [selectedAction, setSelectedAction] = useState(actions[0])
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
            selectedAction={selectedAction.value}
            handleSelectedActionChange={(e) => {
              setSelectedAction(Actions[e.target.value])
            }}
          />
          {selectedAction.fields.account &&
          <FormControl className={classes.formControl}>
            <TextField
              label="For Account"
              value={account}
              onChange={(e) => setAccount(e.target.value)}
              variant="filled"
              size="small"
            />
          </FormControl>
          }
          {selectedAction.fields.tokenId &&
          <FormControl className={classes.formControl}>
            <TextField
              label="Token ID"
              value={tokenId}
              onChange={(e) => setTokenId(e.target.value)}
              variant="filled"
              size="small"
            />
          </FormControl>
          }
          {selectedAction.fields.block &&
          <FormControl className={classes.formControl}>
            <TextField
              label="At Block Number"
              value={atBlock}
              onChange={(e) => setAtBlock(e.target.value)}
              variant="filled"
              size="small"
            />
          </FormControl>
          }
          {selectedAction.fields.startBlock &&
          <FormControl className={classes.formControl}>
            <TextField
              label="Start Block Number"
              value={startNumber}
              onChange={(e) => setStartNumber(e.target.value)}
              variant="filled"
              size="small"
            />
          </FormControl>
          }
          {selectedAction.fields.endBlock &&
          <FormControl className={classes.formControl}>
            <TextField
              label="End Block Number"
              value={endNumber}
              onChange={(e) => setEndNumber(e.target.value)}
              variant="filled"
              size="small"
            />
          </FormControl>
          }
          <br/>
          <br/>
          <Button align="right" variant="contained" color="primary" onClick={() => {
            if (selectedAction.fields.startBlock === 'required' &&
              (startNumber === '' || isNaN(startNumber))) {
              setError('Invalid start block number')
              return
            }
            if (selectedAction.fields.startBlock === 'required' &&
              (endNumber === '' || isNaN(endNumber))) {
              setError('Invalid end block number')
              return
            }
            if (selectedAction.fields.block === 'required' &&
              (atBlock === '' || isNaN(atBlock))) {
              setError('Invalid block number')
              return
            }
            if (selectedAction.fields.account === 'required' &&
              (account === '')) {
              setError('Account cannot be empty')
              return
            }
            if (selectedAction.fields.tokenId === 'required' &&
              (tokenId === '')) {
              setError('Token ID cannot be empty')
              return
            }
            onSearch({
              ...selectedAction,
              params: {
                startNumber,
                endNumber,
                atBlock,
                account,
                tokenId,
              }
            })
            setStartNumber("1")
            setEndNumber(lastPersistedBlockNumber)
            setAccount("")
            setAtBlock(lastPersistedBlockNumber)
            setTokenId("")
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
