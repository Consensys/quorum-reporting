import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import Paper from '@material-ui/core/Paper';
import InputBase from '@material-ui/core/InputBase';
import IconButton from '@material-ui/core/IconButton';
import SearchIcon from '@material-ui/icons/Search';

const useStyles = makeStyles((theme) => ({
    root: {
        padding: '2px 4px',
        display: 'flex',
        alignItems: 'center',
        width: 1000,
    },
    input: {
        marginLeft: theme.spacing(1),
        flex: 1,
    },
    iconButton: {
        padding: 10,
    },
}));

export default function SearchField(props) {
    const classes = useStyles();

    const handleKeyPress = (e) => {
        if (e.key === 'Enter') {
            props.handleSearch()
        }
    };

    return (
        <Paper className={classes.root}>
            <InputBase
                className={classes.input}
                placeholder="Search by Tx Hash or Block Number"
                onChange={props.handleSearchTextChange}
                onKeyPress={handleKeyPress}
                value={props.searchText}
            />
            <IconButton type="submit" className={classes.iconButton} aria-label="search" onClick={props.handleSearch}>
                <SearchIcon />
            </IconButton>
        </Paper>
    );
}