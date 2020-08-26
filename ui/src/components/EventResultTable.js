import React from 'react';
import Paper from '@material-ui/core/Paper';
import TableContainer from '@material-ui/core/TableContainer';
import Table from '@material-ui/core/Table';
import TableHead from '@material-ui/core/TableHead';
import TableBody from '@material-ui/core/TableBody';
import TableRow from '@material-ui/core/TableRow';
import TableCell from '@material-ui/core/TableCell';
import TablePagination from '@material-ui/core/TablePagination';
import CircularProgress from '@material-ui/core/CircularProgress';
import ExpandableEventRow from './ExpandableEventRow';

function TransactionResultTable(props) {
    return (
        <div>
            { !props.isLoading &&
                <div>
                    <TableContainer component={Paper}>
                        <Table size="small" aria-label="collapsible table">
                            <TableHead>
                                <TableRow>
                                    <TableCell width="20%"/>
                                    <TableCell width="20%"><strong>Event Topic</strong></TableCell>
                                    <TableCell width="20%"><strong>Transaction Hash</strong></TableCell>
                                    <TableCell width="20%"><strong>Address</strong></TableCell>
                                    <TableCell width="20%"><strong>Block Number</strong></TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {(props.displayData.slice(props.currentPage * props.pageSize, props.currentPage * props.pageSize + props.pageSize)).map((event, i) => (
                                    <ExpandableEventRow
                                        key={i}
                                        topic={event.topic}
                                        txHash={event.txHash}
                                        address={event.address}
                                        blockNumber={event.blockNumber}
                                        parsedEvent={event.parsedEvent}
                                    />
                                ))}
                            </TableBody>
                        </Table>
                    </TableContainer>
                    <TablePagination
                        component="div"
                        rowsPerPageOptions={[]}
                        count={props.totalEvents}
                        rowsPerPage={props.pageSize}
                        page={props.currentPage}
                        onChangePage={props.handleChangePage}
                    />
                </div>
            }
            { props.isLoading &&
                <div align="center">
                    <br/>
                    <CircularProgress/>
                </div>
            }
        </div>
    )
}

export default TransactionResultTable