import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import InputLabel from '@material-ui/core/InputLabel';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import MenuItem from '@material-ui/core/MenuItem';

const useStyles = makeStyles((theme) => ({
    formControl: {
        margin: theme.spacing(1),
    }
}));

function ContractSelector(props) {
    const classes = useStyles();
    return (
            <FormControl variant="filled" size="small" className={classes.formControl} style={{width: 400, maxWidth: '90%'}}>
                <InputLabel>Actions</InputLabel>
                <Select
                    value={props.selectedAction}
                    onChange={props.handleSelectedActionChange}
                >
                    {props.actions.map(({ label, value }) => (
                        <MenuItem key={value} value={value}>{label}</MenuItem>
                    ))}
                </Select>
            </FormControl>
    )
}

export default ContractSelector