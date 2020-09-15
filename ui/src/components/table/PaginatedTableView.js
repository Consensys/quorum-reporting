import React, { useEffect, useState } from 'react'
import Paper from '@material-ui/core/Paper'
import Typography from '@material-ui/core/Typography'
import CircularProgress from '@material-ui/core/CircularProgress'
import InfoIcon from '@material-ui/icons/Info'
import TableContainer from '@material-ui/core/TableContainer'
import Table from '@material-ui/core/Table'
import TableRow from '@material-ui/core/TableRow'
import TableBody from '@material-ui/core/TableBody'
import TablePagination from '@material-ui/core/TablePagination'
import { makeStyles } from '@material-ui/core/styles'
import { useDispatch, useSelector } from 'react-redux'
import { updateRowsPerPageAction } from '../../redux/actions/systemActions'
import Tooltip from '@material-ui/core/Tooltip'

const useStyles = makeStyles((theme) => ({
  container: {
    marginTop: theme.spacing(0.5),
    marginBottom: theme.spacing(0.5),
    marginLeft: theme.spacing(1),
    marginRight: theme.spacing(1),
  },
  titleContainer: {
    padding: 12,
    display: 'flex',
    flexDirection: 'row',
    alignItems: 'center',
  },
  table: {
    minWidth: 650,
  },
  loading: {
    marginRight: 8,
  }
}))

export function PaginatedTableView ({ title, note, getItems, ItemView, HeaderView, startingRowsPerPage = 10 }) {
  const classes = useStyles()
  const [total, setTotal] = useState(0)
  const [list, setList] = useState([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(0)
  const [lastItemEachPage, setLastItemEachPage] = useState([])
  const dispatch = useDispatch()
  const rowsPerPage = useSelector(state => state.system.rowsPerPage)

  useEffect(() => {
    setLoading(true)
    const lastItem = page === 0 ? undefined : lastItemEachPage[page-1]
    getItems(page, rowsPerPage, lastItem)
      .then(({data, total}) => {
        setLastItemEachPage([...lastItemEachPage, data[data.length-1]])
        setTotal(total)
        setList(data)
        setLoading(false)
      })

  }, [page, rowsPerPage, getItems])
  const handleChangePage = (event, newPage) => {
    setPage(newPage)
  }

  const handleChangeRowsPerPage = (event) => {
    const newRowsPerPage = parseInt(event.target.value, 10)
    dispatch(updateRowsPerPageAction(newRowsPerPage))
    setPage(0)
    setLastItemEachPage([])
  }

  return <Paper className={classes.container}>
    <div className={classes.titleContainer}>
      <Typography variant="h6">{title}</Typography>
      <div style={{flex: 1}} />
      {loading && <CircularProgress size={18} className={classes.loading}/>}
      {note &&
        <Tooltip title={note}>
          <InfoIcon color="secondary" />
        </Tooltip>
      }
    </div>
    <TableContainer component={Paper}>
      <Table size="small" className={classes.table} aria-label="simple table">
        <HeaderView/>
        <TableBody>
          {list.map(ItemView)}
          <TableRow key="pagination">
            <TablePagination
              count={total}
              rowsPerPage={rowsPerPage}
              page={page}
              SelectProps={{
                inputProps: { 'aria-label': 'rows per page' },
                native: true,
              }}
              onChangePage={handleChangePage}
              onChangeRowsPerPage={handleChangeRowsPerPage}
            />
          </TableRow>
        </TableBody>
      </Table>
    </TableContainer>
  </Paper>

}
