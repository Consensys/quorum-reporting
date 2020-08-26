import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import Paper from '@material-ui/core/Paper';
import Typography from '@material-ui/core/Typography';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import TextareaAutosize from "@material-ui/core/TextareaAutosize";
import ArrowBackIcon from '@material-ui/icons/ArrowBack';
import Button from '@material-ui/core/Button';

const useStyles = makeStyles((theme) => ({
    root: {
        padding: '2px 4px',
        marginTop: 10,
        marginBottom: 10,
        alignItems: 'center',
        width: 1200,
    },
}));

export default function RecursiveInfoList(props) {
    const classes = useStyles();
    return (
        <Paper className={classes.root} align="center">
            <br/>
            <div>
                <Button variant="contained" color="primary" onClick={props.handleReturn}>
                    <ArrowBackIcon />
                    &nbsp;
                    Return
                </Button>
            </div>
            <br/>
            <List>
                {
                    Object.keys(props.displayData).map( (k, i) => (
                        <ListItem key={i}>
                            <Typography variant="caption">{k + ":"}&nbsp;</Typography>
                            {
                                JSON.stringify(props.displayData[k]).length > 100 ?
                                    <TextareaAutosize
                                        readOnly
                                        rowsMax={4}
                                        aria-label="maximum height"
                                        style={{fontSize: "16px", width: "1000px"}}
                                        defaultValue={JSON.stringify(props.displayData[k])}
                                    /> : <Typography>{JSON.stringify(props.displayData[k])}</Typography>
                            }
                        </ListItem>
                    ))
                }
            </List>
        </Paper>
    );
}