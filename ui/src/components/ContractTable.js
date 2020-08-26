import React from 'react';
import { withStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import IconButton from '@material-ui/core/IconButton';
import Box from '@material-ui/core/Box';
import Link from '@material-ui/core/Link';
import EditIcon from "@material-ui/icons/Edit";
import DeleteIcon from '@material-ui/icons/Delete';
import ReceiptIcon from '@material-ui/icons/Receipt';
import Popup from './Popup';

const StyledTableHeader = withStyles((theme) => ({
    head: {
        backgroundColor: theme.palette.primary.light,
        color: theme.palette.common.white,
    },
    body: {
        fontSize: 14,
    },
}))(TableCell);

function ContractTable(props) {
    return (
        <Table stickyHeader>
            <TableHead>
                <TableRow>
                    <StyledTableHeader width="5%">#</StyledTableHeader>
                    <StyledTableHeader width="60%">Contract Address</StyledTableHeader>
                    <StyledTableHeader width="20%">Contract Details</StyledTableHeader>
                    <StyledTableHeader width="15%">Contract Operations</StyledTableHeader>
                </TableRow>
            </TableHead>
            <TableBody>
                {props.contracts.map( (c, i) => (
                    <TableRow key={c.address}>
                        <TableCell>
                            {i+1}
                        </TableCell>
                        <TableCell>
                            <Link
                                component="button"
                                variant="body2"
                                onClick={props.handleNavigateContract}
                                value={c.address}
                            >
                                {c.address}
                            </Link>
                        </TableCell>
                        <TableCell>
                            <Box component="span" m={1}>
                                <Popup
                                    name="Show ABI"
                                    content={c.abi}
                                />
                            </Box>
                            <Box component="span" m={1}>
                                <Popup
                                    name="Show Template"
                                    content={c.template}
                                />
                            </Box>
                        </TableCell>
                        <TableCell>
                            <IconButton onClick={()=>{props.handleContractUpdate(c.address)}} >
                                <EditIcon color="primary" />
                            </IconButton>
                            <IconButton onClick={()=>{props.handleContractDelete(c.address)}} >
                                <DeleteIcon color="primary" />
                            </IconButton>
                            <IconButton onClick={()=>{props.handleNavigateReport(c.address)}} >
                                <ReceiptIcon color="primary" />
                            </IconButton>
                        </TableCell>
                    </TableRow>
                ))}
            </TableBody>
        </Table>
    )
}

export default ContractTable