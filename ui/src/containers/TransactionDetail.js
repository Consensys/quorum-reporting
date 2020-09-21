import React, { useEffect, useState } from 'react'
import Alert from '@material-ui/lab/Alert'
import { getSingleTransaction } from '../client/fetcher'
import { useSelector } from 'react-redux'
import { makeStyles } from '@material-ui/core/styles'
import TableRow from '@material-ui/core/TableRow'
import TableCell from '@material-ui/core/TableCell'
import TableBody from '@material-ui/core/TableBody'
import Table from '@material-ui/core/Table'
import TableContainer from '@material-ui/core/TableContainer'
import Typography from '@material-ui/core/Typography'
import { Link } from 'react-router-dom'
import Paper from '@material-ui/core/Paper'
import Grid from '@material-ui/core/Grid'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'

const useStyles = makeStyles((theme) => ({
  root: {
    width: '100%',
  },
  grid: {
    maxWidth: 1080,
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
  title: {
    padding: 12,
    textOverflow: 'ellipsis',
    overflow: 'hidden',
    whiteSpace: 'nowrap',
  },
  table: {
  },
}))

export function TransactionDetail ({ id }) {

  const classes = useStyles()
  const [transaction, setDisplayData] = useState()
  const [errorMessage, setErrorMessage] = useState()
  const { rpcEndpoint, } = useSelector(state => state.system)

  useEffect(() => {
    setDisplayData(undefined)
    getSingleTransaction(rpcEndpoint, id).then((res) => {
      setDisplayData(res)
    }).catch((e) => {
      setErrorMessage(`Transaction not found (${e.message})`)
      setDisplayData(undefined)
    })
  }, [id, rpcEndpoint])

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
        {transaction &&
        <Grid item xs={12}>
          <Card className={classes.details}>
            <CardContent>
              <Typography variant="h6" className={classes.title}>Transaction {id}</Typography>
              <TableContainer>
                <Table className={classes.table} aria-label="simple table">
                  <TableBody>
                    <TableRow key={'from'}>
                      <TableCell width="25%" size="small">from</TableCell>
                      <TableCell align="left" padding="default" data-value={transaction.from}>
                        <Link to={`/contracts/${transaction.from}`}>{transaction.from}</Link>
                      </TableCell>
                    </TableRow>
                    {transaction.to &&
                    transaction.to !== '0x0000000000000000000000000000000000000000' &&
                    <TableRow key={'to'}>
                      <TableCell width="25%" size="small">to</TableCell>
                      <TableCell align="left" padding="default" data-value={transaction.to}>
                        <Link to={`/contracts/${transaction.to}`}>{transaction.to}</Link>
                      </TableCell>
                    </TableRow>
                    }
                    {transaction.createdContract &&
                    transaction.createdContract !== '0x0000000000000000000000000000000000000000' &&
                    <TableRow key={'createdContract'}>
                      <TableCell width="25%" size="small">createdContract</TableCell>
                      <TableCell align="left" padding="default" data-value={transaction.createdContract}>
                        <Link to={`/contracts/${transaction.createdContract}`}>{transaction.createdContract}</Link>
                      </TableCell>
                    </TableRow>
                    }
                    <TableRow key={'value'}>
                      <TableCell width="25%" size="small">value</TableCell>
                      <TableCell align="left" padding="default"
                                 data-value={transaction.value}>{transaction.value}</TableCell>
                    </TableRow>
                    <TableRow key={'gas'}>
                      <TableCell width="25%" size="small">gas</TableCell>
                      <TableCell align="left" padding="default"
                                 data-value={transaction.gas}>{transaction.gas}</TableCell>
                    </TableRow>
                    <TableRow key={'gasPrice'}>
                      <TableCell width="25%" size="small">gasPrice</TableCell>
                      <TableCell align="left" padding="default"
                                 data-value={transaction.gasPrice}>{transaction.gasPrice}</TableCell>
                    </TableRow>
                    <TableRow key={'data'}>
                      <TableCell width="25%" size="small">data</TableCell>
                      <TableCell align="left" padding="default"
                                 data-value={transaction.data}>{transaction.data}</TableCell>
                    </TableRow>
                    <TableRow key={'blockNumber'}>
                      <TableCell width="25%" size="small">blockNumber</TableCell>
                      <TableCell align="left" padding="default" data-value={transaction.blockNumber}>
                        <Link to={`/blocks/${transaction.blockNumber}`}>{transaction.blockNumber}</Link>
                      </TableCell>
                    </TableRow>
                    <TableRow key={'blockHash'}>
                      <TableCell width="25%" size="small">blockHash</TableCell>
                      <TableCell align="left" padding="default" data-value={transaction.blockHash}>
                        {transaction.blockHash}
                      </TableCell>
                    </TableRow>
                    <TableRow key={'status'}>
                      <TableCell width="25%" size="small">status</TableCell>
                      <TableCell align="left" padding="default"
                                 data-value={transaction.status}>{transaction.status ? 1 : 0}</TableCell>
                    </TableRow>
                    <TableRow key={'nonce'}>
                      <TableCell width="25%" size="small">nonce</TableCell>
                      <TableCell align="left" padding="default"
                                 data-value={transaction.nonce}>{transaction.nonce}</TableCell>
                    </TableRow>
                    <TableRow key={'index'}>
                      <TableCell width="25%" size="small">index</TableCell>
                      <TableCell align="left" padding="default" data-value={transaction.index}>
                        {transaction.index}
                      </TableCell>
                    </TableRow>
                    <TableRow key={'parsedData'}>
                      <TableCell width="25%" size="small">parsedData</TableCell>
                      <TableCell align="left" padding="default" data-value={transaction.parsedData}>
                        {transaction.parsedData ? JSON.stringify(transaction.parsedData) : ''}
                      </TableCell>
                    </TableRow>
                    <TableRow key={'parsedEvents'}>
                      <TableCell width="25%" size="small">Events</TableCell>
                      <TableCell align="left" padding="default" data-value={transaction.parsedEvents}>
                        {transaction.parsedEvents ? transaction.parsedEvents.length : ''}
                      </TableCell>
                    </TableRow>
                    <TableRow key={'cumulativeGasUsed'}>
                      <TableCell width="25%" size="small">cumulativeGasUsed</TableCell>
                      <TableCell align="left" padding="default"
                                 data-value={transaction.cumulativeGasUsed}>{transaction.cumulativeGasUsed}</TableCell>
                    </TableRow>
                    <TableRow key={'gasUsed'}>
                      <TableCell width="25%" size="small">gasUsed</TableCell>
                      <TableCell align="left" padding="default" data-value={transaction.gasUsed}>
                        {transaction.gasUsed}
                      </TableCell>
                    </TableRow>
                  </TableBody>
                </Table>
              </TableContainer>
            </CardContent>
          </Card>
        </Grid>
        }
      </Grid>
    </div>
  )
}
