import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import InputLabel from '@material-ui/core/InputLabel';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import MenuItem from '@material-ui/core/MenuItem';
import TextField from '@material-ui/core/TextField';
import Button from '@material-ui/core/Button';
import GavelIcon from '@material-ui/icons/Gavel';

const useStyles = makeStyles((theme) => ({
    formControl: {
        margin: theme.spacing(1),
    }
}));

function ReportForm(props) {
    const classes = useStyles();
    return (
        <div align="center">
            <FormControl variant="filled" size="small" className={classes.formControl} style={{minWidth:400}}>
                <InputLabel>Contract</InputLabel>
                <Select
                    value={props.selectedContract}
                    onChange={props.handleSelectedContractChange}
                >
                    {props.contracts.map( c => (
                        <MenuItem key={c.address} value={c.address}>{c.address}</MenuItem>
                    ))}
                </Select>
            </FormControl>
            <FormControl className={classes.formControl}>
                <TextField
                    label="Start Block Number"
                    value={props.startBlockNumber}
                    onChange={props.handleStartBlockChange}
                    variant="filled"
                    size="small"
                />
            </FormControl>
            <FormControl className={classes.formControl}>
                <TextField
                    label="End Block Number"
                    value={props.endBlockNumber}
                    onChange={props.handleEndBlockChange}
                    variant="filled"
                    size="small"
                />
            </FormControl>
            <br/>
            <br/>
            <Button variant="contained" color="primary" onClick={props.handleReport}>
                <GavelIcon />
                &nbsp;
                Generate Report
            </Button>
        </div>
    )
}

export default ReportForm