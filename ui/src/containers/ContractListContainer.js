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
import { changePageAction } from '../redux/actions/pageActions';
import { ContractPageId, ReportPageId } from '../constants';
import { addContract, deleteContract, getContracts } from '../client/fetcher';

const styles = {
    card: {
        minWidth: 275,
        marginTop: 5,
        marginBottom: 5,
    },
};

class ContractListContainer extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            formIsOpen: false,
            newContract: {
                address: "",
                abi: "",
                template: "",
            },
            errorMessage: "",
        }
    }

    componentDidMount() {
        this.getAllRegisteredContract()
    }

    getAllRegisteredContract = () => {
        getContracts(this.props.rpcEndpoint).then( (contracts) => {
            this.props.dispatch(getContractsAction(contracts))
        })
    };

    handleOpenSetting = () => {
        this.setState({ formIsOpen: true })
    };

    handleCloseSetting = () => {
        this.setState({ formIsOpen: false })
    };

    handleNewContractAddressChange = (e) => {
        this.setState({
            newContract: {
                ...this.state.newContract,
                address: e.target.value,
            },
            errorMessage: "",
        })
    };

    handleNewContractABIChange = (e) => {
        this.setState({
            newContract: {
                ...this.state.newContract,
                abi: e.target.value,
            },
            errorMessage: "",
        })
    };

    handleNewContractTemplateChange = (e) => {
        this.setState({
            newContract: {
                ...this.state.newContract,
                template: e.target.value,
            },
            errorMessage: "",
        })
    };

    handleNavigateContract = (e) => {
        this.props.dispatch(selectContractAction(e.target.value));
        this.props.dispatch(changePageAction(ContractPageId))
    };

    handleNavigateReport = (address) => {
        this.props.dispatch(selectContractAction(address));
        this.props.dispatch(changePageAction(ReportPageId))
    };

    handleRegisterNewContract = () => {
        if (this.state.newContract.address === ""){
            this.setState({
                errorMessage: "address must not be empty",
            });
            return
        }
        if (this.state.newContract.abi === ""){
            this.setState({
                errorMessage: "abi must not be empty",
            });
            return
        }
        if (this.state.newContract.template === "") {
            this.setState({
                errorMessage: "template must not be empty",
            });
            return
        }
        addContract(this.props.rpcEndpoint, this.state.newContract).then( (res) => {
            if (res.data.error) {
                throw res.data.error.message
            }
            this.setState({ formIsOpen: false });
            // give a small timeout to avoid fetch too fast
            setTimeout( () => {
                this.getAllRegisteredContract()
            }, 500)
        }).catch( (e) => {
            this.setState({
                errorMessage: e.toString()
            });
        });
    };

    handleContractUpdate = () => {
        // TODO: update contract
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
                <CardContent>
                    <br/>
                    <Typography variant="h6" align="center">
                        Registered Contract List&nbsp;
                        <IconButton onClick={this.getAllRegisteredContract} >
                            <RefreshIcon/>
                        </IconButton>
                    </Typography>
                    <br/>
                    {
                        this.props.contracts.length === 0 &&
                        <h1 align="center">&lt; No Records Found &gt;</h1>
                    }
                    {
                        this.props.contracts.length !== 0 &&
                        <ContractTable
                            contracts={this.props.contracts}
                            handleContractUpdate={this.handleContractUpdate}
                            handleContractDelete={this.handleContractDelete}
                            handleNavigateContract={this.handleNavigateContract}
                            handleNavigateReport={this.handleNavigateReport}
                        />
                    }
                    <br/>
                    <IconButton color="primary" variant="h4" onClick={this.handleOpenSetting}>
                        <AddIcon />
                    </IconButton>
                    <ContractForm
                        isOpen={this.state.formIsOpen}
                        handleCloseSetting={this.handleCloseSetting}
                        handleNewContractAddressChange={this.handleNewContractAddressChange}
                        handleNewContractABIChange={this.handleNewContractABIChange}
                        handleNewContractTemplateChange={this.handleNewContractTemplateChange}
                        handleRegisterNewContract={this.handleRegisterNewContract}
                        newContract={this.state.newContract}
                        errorMessage={this.state.errorMessage}
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