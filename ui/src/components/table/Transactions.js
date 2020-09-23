import React from 'react'
import { makeStyles } from '@material-ui/core/styles'
import Table from '@material-ui/core/Table'
import TableHead from '@material-ui/core/TableHead'
import TableBody from '@material-ui/core/TableBody'
import TableRow from '@material-ui/core/TableRow'
import TableCell from '@material-ui/core/TableCell'
import IconButton from '@material-ui/core/IconButton'
import KeyboardArrowUpIcon from '@material-ui/icons/KeyboardArrowUp'
import KeyboardArrowDownIcon from '@material-ui/icons/KeyboardArrowDown'
import Collapse from '@material-ui/core/Collapse'
import Box from '@material-ui/core/Box'
import Typography from '@material-ui/core/Typography'
import { Link } from 'react-router-dom'
import PaginatedTableView from './PaginatedTableView'

const useRowStyles = makeStyles({
  root: {
    '& > *': {
      borderBottom: 'unset',
    },
  },
  link: {
    cursor: 'pointer',
    textDecoration: 'none',
    color: '#1976D2',
  },
})

export function TransactionTable({ searchReport, address }) {
  return (
    <PaginatedTableView
      title={searchReport.label}
      HeaderView={TransactionHeader}
      ItemView={TransactionRowItem}
      getItems={(page, rowsPerPage, lastItem) => {
        return searchReport.getItems({ address, ...searchReport.params }, {
          pageNumber: page,
          pageSize: rowsPerPage,
          after: lastItem,
        })
      }}
    />
  )
}

export function TransactionHeader() {
  return (
    <TableHead>
      <TableRow>
        <TableCell width="5%" />
        <TableCell width="5%"><strong>Block</strong></TableCell>
        <TableCell width="45%"><strong>Transaction Hash</strong></TableCell>
        <TableCell width="45%"><strong>From</strong></TableCell>
      </TableRow>
    </TableHead>
  )
}

export function TransactionRowItem(tx) {
  return (
    <ExpandableTxRow
      key={tx.hash}
      txHash={tx.hash}
      from={tx.from}
      to={tx.to}
      blockNumber={tx.blockNumber}
      parsedTransaction={tx.parsedTransaction}
      parsedEvents={tx.parsedEvents}
      internalCalls={tx.internalCalls}
    />
  )
}

export function ExpandableTxRow({
  blockNumber, from, internalCalls, parsedEvents, parsedTransaction, txHash,
}) {
  const [open, setOpen] = React.useState(false)
  const classes = useRowStyles()

  return (
    <>
      <TableRow className={classes.root}>
        <TableCell component="th">
          <IconButton aria-label="expand row" size="small" onClick={() => setOpen(!open)}>
            {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
          </IconButton>
        </TableCell>
        <TableCell align="center">
          <Link
            className={classes.link}
            to={`/blocks/${blockNumber}`}
          >
            {blockNumber}
          </Link>
        </TableCell>
        <TableCell>
          <Link className={classes.link} to={`/transactions/${txHash}`}>{txHash}</Link>
        </TableCell>
        <TableCell>
          {from}
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell
          style={{
            paddingBottom: 0,
            paddingTop: 0,
          }}
          colSpan={6}
        >
          <Collapse in={open} timeout="auto" unmountOnExit>
            <Box margin={1} maxWidth="800px">
              <Typography>Parsed Transaction</Typography>
              <Table size="small" aria-label="a dense table">
                <TableHead>
                  <TableRow>
                    <TableCell><strong>Transaction Signature</strong></TableCell>
                    <TableCell><strong>Function 4 Bytes</strong></TableCell>
                    <TableCell><strong>Parsed Data</strong></TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  <TableRow>
                    <TableCell>{parsedTransaction.txSig}</TableCell>
                    <TableCell>{parsedTransaction.func4Bytes}</TableCell>
                    <TableCell>{JSON.stringify(parsedTransaction.parsedData)}</TableCell>
                  </TableRow>
                </TableBody>
              </Table>
              <br />
              {
                (parsedEvents && parsedEvents.length > 0)
                && (
                  <div>
                    <Typography>Parsed Events</Typography>
                    <Table size="small" aria-label="a dense table">
                      <TableHead>
                        <TableRow>
                          <TableCell><strong>Event Signature</strong></TableCell>
                          <TableCell><strong>Parsed Data</strong></TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {
                          parsedEvents.map((event) => (
                            <TableRow key={event.eventSig + JSON.stringify(event.parsedData)}>
                              <TableCell>{event.eventSig}</TableCell>
                              <TableCell>{JSON.stringify(event.parsedData)}</TableCell>
                            </TableRow>
                          ))
                        }
                      </TableBody>
                    </Table>
                  </div>
                )
              }
              <br />
              {
                (internalCalls && internalCalls.length > 0)
                && (
                  <div>
                    <Typography>Internal Calls</Typography>
                    <Table size="small" aria-label="a dense table">
                      <TableHead>
                        <TableRow>
                          <TableCell><strong>From</strong></TableCell>
                          <TableCell><strong>To</strong></TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {
                          internalCalls.map((c) => (
                            <TableRow key={c.from + c.to}>
                              <TableCell>{c.from}</TableCell>
                              <TableCell>{c.to}</TableCell>
                            </TableRow>
                          ))
                        }
                      </TableBody>
                    </Table>
                  </div>
                )
              }
            </Box>
          </Collapse>
        </TableCell>
      </TableRow>
    </>
  )
}
