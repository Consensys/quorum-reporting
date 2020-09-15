import React from 'react';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import TableCell from '@material-ui/core/TableCell';
import { Link } from 'react-router-dom'
import { useSelector } from 'react-redux'
import { PaginatedTableView } from './PaginatedTableView'

export function TokenHolderTable ({ searchReport, address }) {
    const rpcEndpoint = useSelector(state => state.system.rpcEndpoint)
    return <PaginatedTableView
      title={searchReport.label}
      HeaderView={TokenHolderHeader}
      ItemView={TokenHolderRowItem}
      getItems={(page, rowsPerPage, lastItem) => {
          return searchReport.getItems(rpcEndpoint, { address, ...searchReport.params }, {
              pageNumber: page,
              pageSize: rowsPerPage,
              after: lastItem
          })
      }}
    />
}

export function TokenHolderHeader () {
    return <TableHead>
        <TableRow>
            <TableCell><strong>Account</strong></TableCell>
        </TableRow>
    </TableHead>
}

export function TokenHolderRowItem (item) {
    return <TableRow key={item}>
        <TableCell component="th" scope="row">
            {item}
        </TableCell>
    </TableRow>
}

export function TokenTable ({ searchReport, address }) {
    const rpcEndpoint = useSelector(state => state.system.rpcEndpoint)
    return <PaginatedTableView
      title={searchReport.label}
      HeaderView={TokenHeader}
      ItemView={TokenRowItem}
      getItems={(page, rowsPerPage, lastItem) => {
          return searchReport.getItems(rpcEndpoint, { address, ...searchReport.params }, {
              pageNumber: page,
              pageSize: rowsPerPage,
              after: lastItem
          })
      }}
    />
}


export function TokenHeader () {
    return <TableHead>
        <TableRow>
            <TableCell><strong>Token</strong></TableCell>
            <TableCell><strong>Holder</strong></TableCell>
            <TableCell><strong>From</strong></TableCell>
            <TableCell><strong>Until</strong></TableCell>
        </TableRow>
    </TableHead>
}

export function TokenRowItem (item) {
    return <TableRow key={item}>
        <TableCell component="th" scope="row">
            {item.token}
        </TableCell>
        <TableCell component="th" scope="row">
            {item.holder}
        </TableCell>
        <TableCell component="th" scope="row">
            <Link to={`/blocks/${item.heldFrom}`}>{item.heldFrom}</Link>
        </TableCell>
        <TableCell component="th" scope="row">
            {item.heldUntil ? <Link to={`/blocks/${item.heldFrom}`}>{item.heldFrom}</Link> : ""}
        </TableCell>
    </TableRow>
}

export function TokenBalanceTable ({ searchReport, address }) {
    const rpcEndpoint = useSelector(state => state.system.rpcEndpoint)
    return <PaginatedTableView
      title={searchReport.label}
      HeaderView={BalanceHeader}
      ItemView={BalanceRowItem}
      getItems={(page, rowsPerPage, lastItem) => {
          return searchReport.getItems(rpcEndpoint, { address, ...searchReport.params }, {
              pageNumber: page,
              pageSize: rowsPerPage,
              after: lastItem
          })
      }}
    />
}

export function BalanceHeader () {
    return <TableHead>
        <TableRow>
            <TableCell><strong>Block</strong></TableCell>
            <TableCell><strong>Balance</strong></TableCell>
        </TableRow>
    </TableHead>
}
export function BalanceRowItem (item) {
    return <TableRow key={item}>
        <TableCell component="th" scope="row">
            <Link to={`/blocks/${item.block}`}>{item.block}</Link>
        </TableCell>
        <TableCell component="th" scope="row">
            {item.balance}
        </TableCell>
    </TableRow>
}

