import React from 'react'
import { makeStyles } from '@material-ui/core/styles'
import Table from '@material-ui/core/Table'
import TableBody from '@material-ui/core/TableBody'
import TableCell from '@material-ui/core/TableCell'
import TableRow from '@material-ui/core/TableRow'
import IconButton from '@material-ui/core/IconButton'
import DeleteIcon from '@material-ui/icons/Delete'
import { useHistory } from 'react-router-dom'
import { Edit } from '@material-ui/icons'

const useStyles = makeStyles(() => ({
  row: {
    textDecoration: 'none',
    cursor: 'pointer',
  },
}))

function ContractTable({ contracts, handleContractDelete, handleContractEdit }) {
  const history = useHistory()
  const classes = useStyles()
  return (
    <Table stickyHeader>
      <TableBody>
        {contracts.map((c, i) => {
          return (
            <TableRow
              hover
              className={classes.row}
              onClick={() => history.push(`/contracts/${c.address}`)}
              key={c.address}
            >
              <TableCell width="10%">
                {i + 1}
              </TableCell>
              <TableCell width="30%">
                {c.name}
              </TableCell>
              <TableCell>
                {c.address}
              </TableCell>
              <TableCell component="th" width="15%" align="right">
                <IconButton
                  onClick={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                    handleContractEdit(c)
                  }}
                >
                  <Edit color="action" />
                </IconButton>
                <IconButton
                  onClick={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                    handleContractDelete(c.address)
                  }}
                >
                  <DeleteIcon color="action" />
                </IconButton>
              </TableCell>
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}

export default ContractTable
