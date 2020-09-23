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
import { useSelector } from 'react-redux'
import PaginatedTableView from './PaginatedTableView'

const useRowStyles = makeStyles({
  root: {
    '& > *': {
      borderBottom: 'unset',
    },
  },
})

export function EventTable({ searchReport, address }) {
  return (
    <PaginatedTableView
      title={searchReport.label}
      HeaderView={EventHeader}
      ItemView={EventRowItem}
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

export function EventHeader() {
  return (
    <TableHead>
      <TableRow>
        <TableCell width="5%" />
        <TableCell width="5%"><strong>Block</strong></TableCell>
        <TableCell width="20%"><strong>Event Topic</strong></TableCell>
        <TableCell width="20%"><strong>Transaction Hash</strong></TableCell>
      </TableRow>
    </TableHead>
  )
}

export function EventRowItem(event) {
  return (
    <ExpandableEventRow
      key={event.txHash + JSON.stringify(event.parsedEvent)}
      topic={event.topic}
      txHash={event.txHash}
      address={event.address}
      blockNumber={event.blockNumber}
      parsedEvent={event.parsedEvent}
    />
  )
}

export function ExpandableEventRow({
  blockNumber, parsedEvent, topic, txHash,
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
          {topic}
        </TableCell>
        <TableCell>
          <Link className={classes.link} to={`/transactions/${txHash}`}>{txHash}</Link>
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
              <Typography>Parsed Event</Typography>
              <Table size="small" aria-label="a dense table">
                <TableHead>
                  <TableRow>
                    <TableCell><strong>Event Signature</strong></TableCell>
                    <TableCell><strong>Parsed Data</strong></TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  <TableRow>
                    <TableCell>{parsedEvent.eventSig}</TableCell>
                    <TableCell>{JSON.stringify(parsedEvent.parsedData)}</TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </Box>
          </Collapse>
        </TableCell>
      </TableRow>
    </>
  )
}
