import React from 'react';
import { connect } from 'react-redux';
import { withStyles } from '@material-ui/core/styles';
import Alert from '@material-ui/lab/Alert';
import SearchField from '../components/SearchField';
import ContractListContainer from './ContractListContainer';
import RecursiveInfoList from '../components/RecursiveInfoList';
import {getSingleBlock, getSingleTransaction} from '../client/fetcher';

const styles = {
    root: {
        marginTop: 10,
        marginBottom: 10,
    },
    alert: {
        marginTop: 5,
        width: 1000,
    }
};

class SearchContainer extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            searchText: "",
            displayData: "",
            errorMessage: "",
        }
    }

    handleSearchTextChange = (e) => {
        this.setState({
            searchText: e.target.value,
            errorMessage: "",
        })
    };

    handleSearch = () => {
        if ((/^[1-9][0-9]*$/g).test(this.state.searchText)) {
            // search block
            let blockNumber = parseInt(this.state.searchText);
            if (this.props.lastPersistedBlockNumber < blockNumber) {
                this.setState({errorMessage: "block number exceed the last persisted"});
                return
            }
            getSingleBlock(this.props.rpcEndpoint, blockNumber).then( (res) => {
                this.setState({
                    displayData: res,
                    errorMessage: "",
                })
            }).catch( (e) => {
                this.setState({errorMessage: e.toString()})
            });
            return
        }
        if ((/^0x[0-9a-fA-F]{64}$/g).test(this.state.searchText)) {
            // search tx
            getSingleTransaction(this.props.rpcEndpoint, this.state.searchText).then( (res) => {
                this.setState({
                    displayData: res,
                    errorMessage: "",
                })
            }).catch( (e) => {
                this.setState({errorMessage: e.toString()})
            });
            return
        }
        this.setState({errorMessage: "invalid search text"});
    };

    handleReturn = () => {
        this.setState({
            displayData: "",
        })
    };

    render() {
        return (
            <div className={this.props.classes.root} align="center">
                <br/>
                <SearchField
                    searchText={this.state.searchText}
                    handleSearchTextChange={this.handleSearchTextChange}
                    handleSearch={this.handleSearch}
                />
                {
                    this.state.errorMessage &&
                    <Alert severity="error" className={this.props.classes.alert}>{this.state.errorMessage}</Alert>
                }
                {
                    this.state.displayData === "" &&
                    <ContractListContainer />
                }
                {
                    this.state.displayData !== "" &&
                    <RecursiveInfoList
                        displayData={this.state.displayData}
                        handleReturn={this.handleReturn}
                    />
                }
            </div>
        )
    }
}

const mapStateToProps = state => {
    return {
        rpcEndpoint: state.system.rpcEndpoint,
        lastPersistedBlockNumber: state.system.lastPersistedBlockNumber,
    }
};

export default connect(mapStateToProps)(withStyles(styles)(SearchContainer))