import React, { useEffect, useState } from 'react'
import Button from '@material-ui/core/Button'
import TextField from '@material-ui/core/TextField'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import DialogContentText from '@material-ui/core/DialogContentText'
import DialogTitle from '@material-ui/core/DialogTitle'
import Alert from '@material-ui/lab/Alert'
import FormControl from '@material-ui/core/FormControl'
import InputLabel from '@material-ui/core/InputLabel'
import Select from '@material-ui/core/Select'
import MenuItem from '@material-ui/core/MenuItem'
import { getTemplates } from '../client/rpcClient'
import { useSelector } from 'react-redux'
import { addContract } from '../client/fetcher'

function ContractForm(props) {
     const [templates, setTemplates] = useState([])
    const [selectedTemplate, setSelectedTemplate] = useState('')
    const [address, setAddress] = useState('')
    const [abi, setAbi] = useState('')
    const [name, setName] = useState('')
    const [storageLayout, setStorageLayout] = useState('')
    const [errorMessage, setErrorMessage] = useState('')
    const rpcEndpoint = useSelector(state => state.system.rpcEndpoint)

  useEffect(() => {
    if(!rpcEndpoint) {
      console.log('returning', rpcEndpoint)
      return
    }
      getTemplates(rpcEndpoint)
        .then((res) => {
            setTemplates(res.data.result)
        })
    }, [rpcEndpoint])

    const handleKeyPress = (e) => {
        if (e.key === 'Enter') {
            props.handleRegisterNewContract()
        }
    };

    return (
        <Dialog open={props.isOpen} onClose={props.handleCloseSetting} aria-labelledby="form-dialog-title" maximumwidth="400" fullWidth>
            <DialogTitle id="form-dialog-title">
              Register a new contract for reporting.
            </DialogTitle>
          <DialogContent>
                <DialogContentText>
                </DialogContentText>
                <br/>
              <TextField
                label="Contract Address"
                value={address}
                onChange={(e) => setAddress(e.target.value)}
                onKeyPress={handleKeyPress}
                margin="dense"
                fullWidth
                autoFocus
              />
              <FormControl margin="dense" fullWidth>
                    <InputLabel>Contract Template</InputLabel>
                    <Select
                      value={selectedTemplate}
                      onChange={(e) => setSelectedTemplate(e.target.value)}
                    >
                        {templates.map( c => (
                          <MenuItem key={c} value={c}>{c}</MenuItem>
                        ))}
                      <MenuItem key={'new'} value={'new'}><strong>New Template</strong></MenuItem>
                    </Select>
                </FormControl>
                { selectedTemplate === 'new' && [
                    <TextField
                      label="Contract Template Name"
                      key="name"
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                      onKeyPress={handleKeyPress}
                      margin="dense"
                      fullWidth
                    />,
                    <TextField
                      label="Contract Template ABI"
                      key="abi"
                      value={abi}
                      onChange={(e) => setAbi(e.target.value)}
                      onKeyPress={handleKeyPress}
                      margin="dense"
                      fullWidth
                      multiline
                    />,
                    <TextField
                      label="Contract Template Storage Template"
                      key="storageLayout"
                      value={storageLayout}
                      onChange={(e) => setStorageLayout(e.target.value)}
                      onKeyPress={handleKeyPress}
                      margin="dense"
                      fullWidth
                      multiline
                    />
                ]}
            </DialogContent>
            {
                errorMessage &&
                <div>
                    <br/>
                    <Alert severity="error">{errorMessage}</Alert>
                </div>
            }
            <DialogActions>
                <Button onClick={props.handleCloseSetting} color="primary">
                    Cancel
                </Button>
                <Button onClick={() => {
                  if (address === ""){
                    setErrorMessage("Address must not be empty")
                    return
                  }
                  if (selectedTemplate === "new") {
                    if(name === '') {
                      setErrorMessage("Template name must not be empty")
                      return
                    }
                    if(abi === '') {
                      setErrorMessage("Template abi must not be empty")
                      return
                    }
                    if(storageLayout === '') {
                      setErrorMessage("Storage Template must not be empty")
                      return
                    }
                  } else if (selectedTemplate === '') {
                    setErrorMessage("Please select a template")
                    return
                  }
                  const newContract = {
                    address,
                    template: selectedTemplate,
                    newTemplate: {
                      name,
                      abi,
                      storageLayout,
                    }
                  }
                  addContract(rpcEndpoint, newContract).then( (res) => {
                    if (res.data.error) {
                      setErrorMessage(res.data.error)
                    } else {
                      props.handleCloseSetting()
                    }
                  }).catch( (e) => {
                    setErrorMessage(e.toString)
                  });
                }} color="primary">
                    Register
                </Button>
            </DialogActions>
        </Dialog>
    )
}

export default ContractForm