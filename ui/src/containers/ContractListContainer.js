import React from 'react';
import { connect } from 'react-redux';
import { withStyles } from '@material-ui/core/styles';
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import Typography from '@material-ui/core/Typography';
import IconButton from '@material-ui/core/IconButton';
import AddIcon from '@material-ui/icons/Add';
import RefreshIcon from '@material-ui/icons/Refresh';
import ContractTable from '../components/ContractTable';
import ContractForm from '../components/ContractForm';
import { getContractsAction, selectContractAction } from '../redux/actions/contractActions';
import { addContract, deleteContract, getContracts } from '../client/fetcher';
import Button from '@material-ui/core/Button'

const styles = {
    card: {
        minWidth: 275,
        marginTop: 5,
        marginBottom: 5,
        width: '95%',
        maxWidth: 1080,
    },
    cardContent: {
        display: 'flex',
        flexDirection: 'column',
    }
};

class ContractListContainer extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            formIsOpen: false,
            errorMessage: "",
        }
    }

    componentDidMount() {
        this.getAllRegisteredContract()
    }

    getAllRegisteredContract = () => {
        getContracts(this.props.rpcEndpoint).then( (contracts) => {
            const sortedContracts = contracts.sort((a, b) => a.name.localeCompare(b.name))
            this.props.dispatch(getContractsAction(sortedContracts))
        })
    };

    handleOpenSetting = () => {
        this.setState({ formIsOpen: true })
    };

    handleCloseSetting = () => {
        this.setState({ formIsOpen: false })
        // give a small timeout to avoid fetch too fast
        setTimeout( () => {
            this.getAllRegisteredContract()
        }, 500)
    };

    handleContractDelete = (address) => {
        deleteContract(this.props.rpcEndpoint, address).then( () => {
            // TODO: handle error?
            // give a small timeout to avoid fetch too fast
            setTimeout( () => {
                this.getAllRegisteredContract()
            }, 500)
        })
    };

    render(){
        return (
            <Card className={this.props.classes.card}>
                <CardContent className={this.props.classes.cardContent}>
                    <Typography variant="h6" align="left">
                        Registered Contract List&nbsp;
                        <IconButton onClick={this.getAllRegisteredContract} >
                            <RefreshIcon/>
                        </IconButton>
                    </Typography>
                    <br/>
                    {
                        this.props.contracts.length === 0 &&
                        <h1 align="center">No Contracts Registered</h1>
                    }
                    {
                        this.props.contracts.length !== 0 &&
                        <ContractTable
                            contracts={this.props.contracts}
                            handleContractDelete={this.handleContractDelete}
                        />
                    }
                    <br/>
                    <Button color="primary" onClick={this.handleOpenSetting}>
                        <AddIcon />&nbsp;Add Contract
                    </Button>
                    <ContractForm
                        isOpen={this.state.formIsOpen}
                        handleCloseSetting={this.handleCloseSetting}
                    />
                </CardContent>
            </Card>
        )
    }
}

const mapStateToProps = state => {
    return {
        rpcEndpoint: state.system.rpcEndpoint,
        contracts: state.user.contracts,
    }
};

export default connect(mapStateToProps)(withStyles(styles)(ContractListContainer))