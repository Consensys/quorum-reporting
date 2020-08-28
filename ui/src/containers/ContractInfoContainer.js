import React from 'react';
import { connect } from 'react-redux';
import { withStyles } from '@material-ui/core/styles';
import CardContent from '@material-ui/core/CardContent';
import Card from '@material-ui/core/Card';
import Typography from '@material-ui/core/Typography';
import Button from '@material-ui/core/Button';
import SearchIcon from '@material-ui/icons/Search';
import Alert from '@material-ui/lab/Alert';
import ContractSelector from '../components/ContractSelector';
import TransactionResultTable from '../components/TransactionResultTable';
import EventResultTable from '../components/EventResultTable';
import { selectContractAction } from '../redux/actions/contractActions';
import { GetContractCreationTx, GetEvents, GetInternalToTxs, GetToTxs } from '../constants';
import { getContractCreationTx, getInternalToTxs, getToTxs, getEvents } from '../client/fetcher';

const styles = {
    card: {
        minWidth: 275,
        marginTop: 5,
        marginBottom: 5,
    },
};

const pageSize = 10;

const actions = [GetContractCreationTx, GetToTxs, GetInternalToTxs, GetEvents];

class ContractInfoContainer extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            selectedAction: "",
            displayTxResult: false,
            displayEventResult: false,
            isLoading: true,
            displayData: [],
            displayDataLength: 0,
            currentPage: 0,
            errorMessage: "",
            // fix current contract and action for paging
            currentContract: "",
            currentSelectedAction: "",
        }
    }

    componentWillUnmount() {
        this.props.dispatch(selectContractAction(""))
    }

    handleSelectedContractChange = (e) => {
        this.setState({ errorMessage: ""});
        this.props.dispatch(selectContractAction(e.target.value))
    };

    handleSelectedActionChange = (e) => {
        this.setState({
            errorMessage: "",
            selectedAction: e.target.value,
        })
    };

    handleChangePage = (event, newPage) => {
        this.setState({currentPage: newPage});
        this.handleSearch(false)
    };

    handleSearch = (newSearch) => {
        // clear display
        this.setState({
            displayTxResult: false,
            displayEventResult: false,
        });
        // check new search condition
        let currentContract = this.state.currentContract;
        let currentSelectedAction = this.state.currentSelectedAction;
        let currentPage = this.state.currentPage;
        if (newSearch) {
            if (this.props.selectedContract === "") {
                this.setState({ errorMessage: "no contract selected"});
                return
            }
            if (this.state.selectedAction === ""){
                this.setState({ errorMessage: "no action selected"});
                return
            }
            currentContract = this.props.selectedContract;
            currentSelectedAction = this.state.selectedAction;
            currentPage = 0;
            this.setState({ currentContract, currentSelectedAction, currentPage: 0 })
        }
        // start loading
        this.setState({
            isLoading: true,
            errorMessage: "",
        });
        this.searchByPage(currentContract, currentSelectedAction, currentPage)
    };

    searchByPage = (contract, action, pageNumber) => {
        if (action === GetEvents) {
            getEvents(this.props.rpcEndpoint, contract, {pageSize, pageNumber}).then( (res) => {
                this.setState({
                    displayData: res.data,
                    displayDataLength: res.total,
                    isLoading: false,
                })
            }).catch( (e) => {
                this.setState({
                    displayData: [],
                    displayDataLength: 0,
                    isLoading: false,
                    errorMessage: e.toString(),
                })
            });
            this.setState({displayEventResult: true});
        } else {
            switch (action) {
                case GetContractCreationTx:
                    getContractCreationTx(this.props.rpcEndpoint, contract).then( (res) => {
                        this.setState({
                            displayData: [res],
                            displayDataLength: 1,
                            isLoading: false,
                        })
                    }).catch( (e) => {
                        console.log(e);
                        this.setState({
                            displayData: [],
                            displayDataLength: 0,
                            isLoading: false,
                            errorMessage: e.toString(),
                        })
                    });
                    break;
                case GetToTxs:
                    getToTxs(this.props.rpcEndpoint, contract, {pageSize, pageNumber}).then( (res) => {
                        this.setState({
                            displayData: res.data,
                            displayDataLength: res.total,
                            isLoading: false,
                        })
                    }).catch( (e) => {
                        this.setState({
                            displayData: [],
                            displayDataLength: 0,
                            isLoading: false,
                            errorMessage: e.toString(),
                        })
                    });
                    break;
                case GetInternalToTxs:
                    getInternalToTxs(this.props.rpcEndpoint, contract, {pageSize, pageNumber}).then( (res) => {
                        this.setState({
                            displayData: res.data,
                            displayDataLength: res.total,
                            isLoading: false,
                        })
                    }).catch( (e) => {
                        this.setState({
                            displayData: [],
                            displayDataLength: 0,
                            isLoading: false,
                            errorMessage: e.toString(),
                        })
                    });
                    break;
                default:
                    this.setState({
                        displayData: [],
                        displayDataLength: 0,
                        isLoading: false,
                        errorMessage: "unknown action: " + action.toString(),
                    })
            }
            this.setState({displayTxResult: true});
        }
    };

    render(){
        return (
            <Card className={this.props.classes.card}>
                <CardContent>
                    <div align="center">
                        <Typography variant="h6">
                            Select Contract
                        </Typography>
                        <br/>
                        <ContractSelector
                            selectedContract={this.props.selectedContract}
                            contracts={this.props.contracts}
                            handleSelectedContractChange={this.handleSelectedContractChange}
                            actions={actions}
                            selectedAction={this.state.selectedAction}
                            handleSelectedActionChange={this.handleSelectedActionChange}
                        />
                        <br/>
                        <Button variant="contained" color="primary" onClick={this.handleSearch.bind(null, true)}>
                            <SearchIcon />
                            &nbsp;
                            Search
                        </Button>
                    </div>
                    <br/>
                    {
                        this.state.errorMessage &&
                        <div>
                            <br/>
                            <Alert severity="error">{this.state.errorMessage}</Alert>
                        </div>
                    }
                    {
                        this.state.displayTxResult &&
                        <TransactionResultTable
                            displayData={this.state.displayData}
                            isLoading={this.state.isLoading}
                            currentPage={this.state.currentPage}
                            pageSize={pageSize}
                            totalTxs={this.state.displayDataLength}
                            handleChangePage={this.handleChangePage}
                        />
                    }
                    {
                        this.state.displayEventResult &&
                        <EventResultTable
                            displayData={this.state.displayData}
                            isLoading={this.state.isLoading}
                            currentPage={this.state.currentPage}
                            pageSize={pageSize}
                            totalEvents={this.state.displayDataLength}
                            handleChangePage={this.handleChangePage}
                        />
                    }
                </CardContent>
            </Card>
        )
    }
}

const mapStateToProps = state => {
    return {
        rpcEndpoint: state.system.rpcEndpoint,
        contracts: state.user.contracts,
        selectedContract: state.user.selectedContract,
    }
};

export default connect(mapStateToProps)(withStyles(styles)(ContractInfoContainer))