import React from 'react';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';

function SettingForm(props) {

    const handleKeyPress = (e) => {
        if (e.key === 'Enter') {
            props.handleSetRPCEndpoint()
        }
    };

    return (
        <Dialog open={props.isOpen} onClose={props.handleCloseSetting} aria-labelledby="form-dialog-title" maximumwidth="400" fullWidth>
            <DialogTitle id="form-dialog-title">Connection ({props.rpcEndpoint})</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    Set the RPCEndpoint for Reporting Engine.
                </DialogContentText>
                <br/>
                <TextField
                    label="RPC Endpoint"
                    value={props.newRPCEndpoint}
                    onChange={props.handleRPCEndpointChange}
                    onKeyPress={handleKeyPress}
                    margin="dense"
                    fullWidth
                    autoFocus
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={props.handleCloseSetting} color="primary">
                    Cancel
                </Button>
                <Button onClick={props.handleSetRPCEndpoint} color="primary">
                    Update
                </Button>
            </DialogActions>
        </Dialog>
    )
}

export default SettingForm