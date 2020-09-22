import React from 'react'
import Button from '@material-ui/core/Button'
import TextField from '@material-ui/core/TextField'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import DialogContentText from '@material-ui/core/DialogContentText'
import DialogTitle from '@material-ui/core/DialogTitle'

function SettingForm({
  handleCloseSetting,
  handleRPCEndpointChange,
  handleSetRPCEndpoint,
  isOpen,
  newRPCEndpoint,
  rpcEndpoint,
}) {
  const handleKeyPress = (e) => {
    if (e.key === 'Enter') {
      handleSetRPCEndpoint()
    }
  }

  return (
    <Dialog
      open={isOpen}
      onClose={handleCloseSetting}
      aria-labelledby="form-dialog-title"
      maximumwidth="400"
      fullWidth
    >
      <DialogTitle id="form-dialog-title">
        Connection (
        {rpcEndpoint}
        )
      </DialogTitle>
      <DialogContent>
        <DialogContentText>
          Set the RPCEndpoint for Reporting Engine.
        </DialogContentText>
        <br />
        <TextField
          label="RPC Endpoint"
          value={newRPCEndpoint}
          onChange={handleRPCEndpointChange}
          onKeyPress={handleKeyPress}
          margin="dense"
          fullWidth
          autoFocus
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={handleCloseSetting} color="primary">
          Cancel
        </Button>
        <Button onClick={handleSetRPCEndpoint} color="primary">
          Update
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default SettingForm
