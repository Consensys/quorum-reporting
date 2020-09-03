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
        <div>
            <FormControl variant="filled" size="small" className={classes.formControl} style={{minWidth: 400}}>
                <InputLabel>Actions</InputLabel>
                <Select
                    value={props.selectedAction}
                    onChange={props.handleSelectedActionChange}
                >
                    {props.actions.map(a => (
                        <MenuItem key={a} value={a}>{a}</MenuItem>
                    ))}
                </Select>
            </FormControl>
        </div>
    )
}

export default ContractSelector