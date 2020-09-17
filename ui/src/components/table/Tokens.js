import React from 'react'
import TableHead from '@material-ui/core/TableHead'
import TableRow from '@material-ui/core/TableRow'
import TableCell from '@material-ui/core/TableCell'
import { Link } from 'react-router-dom'
import { useSelector } from 'react-redux'
import { PaginatedTableView } from './PaginatedTableView'
import Reports from '../../reports'

export function TokenHolderTable ({ searchReport, address }) {
    const rpcEndpoint = useSelector(state => state.system.rpcEndpoint)
    return <PaginatedTableView
      title={searchReport.label}
      HeaderView={getHeaderView(searchReport)}
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

function getHeaderView(report) {
    switch (report.value) {
        case Reports.ERC721Holders.value:
            return ERC721HolderCountHeader
        case Reports.ERC721HolderForToken.value:
            return ERC721HolderHeader
        case Reports.ERC20TokenHolders.value:
        default:
            return ERC20HolderHeader
    }
}

function ERC20HolderHeader () {
    return <TokenHolderHeader secondColumnName={'Balance'} />
}

function ERC721HolderCountHeader () {
    return <TokenHolderHeader secondColumnName={'Token Count'} />
}

function ERC721HolderHeader () {
    return <TokenHolderHeader secondColumnName={'Token ID'} />
}

function TokenHolderHeader ({ secondColumnName }) {
    return <TableHead>
        <TableRow>
            <TableCell><strong>Account</strong></TableCell>
            <TableCell width="10%"><strong>{secondColumnName}</strong></TableCell>
        </TableRow>
    </TableHead>
}

export function TokenHolderRowItem (item) {
    return <TableRow key={item.holder + item.value}>
        <TableCell>
            {item.holder}
        </TableCell>
        <TableCell>
            {item.value}
        </TableCell>
    </TableRow>
}

export function TokenTable ({ searchReport, address }) {
    const rpcEndpoint = useSelector(state => state.system.rpcEndpoint)
    return <PaginatedTableView
      title={searchReport.label}
      subtitle={searchReport.subtitle && searchReport.subtitle(searchReport.params)}
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
            <TableCell width="10%"><strong>Token</strong></TableCell>
            <TableCell><strong>Holder</strong></TableCell>
            <TableCell width="10%"><strong>From</strong></TableCell>
            <TableCell width="10%"><strong>Until</strong></TableCell>
        </TableRow>
    </TableHead>
}

export function TokenRowItem (item) {
    return <TableRow key={item.token}>
        <TableCell>
            {item.token}
        </TableCell>
        <TableCell>
            {item.holder}
        </TableCell>
        <TableCell>
            <Link to={`/blocks/${item.heldFrom}`}>{item.heldFrom}</Link>
        </TableCell>
        <TableCell>
            {item.heldUntil ? <Link to={`/blocks/${item.heldFrom}`}>{item.heldFrom}</Link> : ''}
        </TableCell>
    </TableRow>
}

export function TokenBalanceTable ({ searchReport, address }) {
    const rpcEndpoint = useSelector(state => state.system.rpcEndpoint)
    return <PaginatedTableView
      title={searchReport.label}
      subtitle={searchReport.subtitle && searchReport.subtitle(searchReport.params)}
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
    return <TableRow key={item.block}>
        <TableCell>
            <Link to={`/blocks/${item.block}`}>{item.block}</Link>
        </TableCell>
        <TableCell>
            {item.balance}
        </TableCell>
    </TableRow>
}

