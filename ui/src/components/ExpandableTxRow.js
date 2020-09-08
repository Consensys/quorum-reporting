import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableHead from '@material-ui/core/TableHead';
import TableBody from '@material-ui/core/TableBody';
import TableRow from '@material-ui/core/TableRow';
import TableCell from '@material-ui/core/TableCell';
import IconButton from '@material-ui/core/IconButton';
import KeyboardArrowUpIcon from '@material-ui/icons/KeyboardArrowUp';
import KeyboardArrowDownIcon from '@material-ui/icons/KeyboardArrowDown';
import Collapse from '@material-ui/core/Collapse';
import Box from '@material-ui/core/Box';
import Typography from '@material-ui/core/Typography';
import { Link } from 'react-router-dom'

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
    }
});

function ExpandableTxRow(props) {
    const [open, setOpen] = React.useState(false);
    const classes = useRowStyles();

    return (
        <React.Fragment>
            <TableRow className={classes.root}>
                <TableCell>
                    <IconButton aria-label="expand row" size="small" onClick={() => setOpen(!open) }>
                        {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
                    </IconButton>
                </TableCell>
                <TableCell component="th" scope="row" align="center">
                    <Link className={classes.link} to={`/blocks/${props.blockNumber}`}>{props.blockNumber}</Link>
                </TableCell>
                <TableCell component="th" scope="row">
                    <Link className={classes.link} to={`/transactions/${props.txHash}`}>{props.txHash}</Link>
                </TableCell>
                <TableCell component="th" scope="row">
                    {props.from}
                </TableCell>
            </TableRow>
            <TableRow>
                <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={6} >
                    <Collapse in={open} timeout="auto" unmountOnExit>
                        <Box margin={1} maxWidth={"800px"}>
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
                                        <TableCell component="th" scope="row">{props.parsedTransaction.txSig}</TableCell>
                                        <TableCell>{props.parsedTransaction.func4Bytes}</TableCell>
                                        <TableCell>{JSON.stringify(props.parsedTransaction.parsedData)}</TableCell>
                                    </TableRow>
                                </TableBody>
                            </Table>
                            <br/>
                            {
                                (props.parsedEvents && props.parsedEvents.length > 0) &&
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
                                                props.parsedEvents.map((event, i) => (
                                                    <TableRow key={i}>
                                                        <TableCell>{event.eventSig}</TableCell>
                                                        <TableCell>{JSON.stringify(event.parsedData)}</TableCell>
                                                    </TableRow>
                                                ))
                                            }
                                        </TableBody>
                                    </Table>
                                </div>
                            }
                            <br/>
                            {
                                (props.internalCalls && props.internalCalls.length > 0) &&
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
                                                props.internalCalls.map((c, i) => (
                                                    <TableRow key={i}>
                                                        <TableCell>{c.from}</TableCell>
                                                        <TableCell>{c.to}</TableCell>
                                                    </TableRow>
                                                ))
                                            }
                                        </TableBody>
                                    </Table>
                                </div>
                            }
                        </Box>
                    </Collapse>
                </TableCell>
            </TableRow>
        </React.Fragment>
    );
}

export default ExpandableTxRow