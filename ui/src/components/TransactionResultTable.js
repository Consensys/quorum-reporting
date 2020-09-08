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
import ExpandableTxRow from './ExpandableTxRow';

function TransactionResultTable(props) {
    return (
        <div>
            { !props.isLoading &&
                <div>
                    <TableContainer component={Paper}>
                        <Table size="small" aria-label="collapsible table">
                            <TableHead>
                                <TableRow>
                                    <TableCell width="5%"/>
                                    <TableCell width="5%"><strong>Block</strong></TableCell>
                                    <TableCell width="45%"><strong>Transaction Hash</strong></TableCell>
                                    <TableCell width="45%"><strong>From</strong></TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {(props.displayData.slice(props.currentPage * props.pageSize, props.currentPage * props.pageSize + props.pageSize)).map((tx, i) => (
                                    <ExpandableTxRow
                                        key={i}
                                        txHash={tx.hash}
                                        from={tx.from}
                                        to={tx.to}
                                        blockNumber={tx.blockNumber}
                                        parsedTransaction={tx.parsedTransaction}
                                        parsedEvents={tx.parsedEvents}
                                        internalCalls={tx.internalCalls}
                                    />
                                ))}
                            </TableBody>
                        </Table>
                    </TableContainer>
                    <TablePagination
                        component="div"
                        rowsPerPageOptions={[]}
                        count={props.totalTxs}
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