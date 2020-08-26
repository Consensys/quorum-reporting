import React from 'react';
import { connect } from 'react-redux';
import { withStyles } from '@material-ui/core/styles';
import CardContent from '@material-ui/core/CardContent';
import Card from '@material-ui/core/Card';
import Typography from '@material-ui/core/Typography';
import Alert from '@material-ui/lab/Alert';
import ReportForm from '../components/ReportForm';
import Report from '../components/Report';
import { getReportData } from '../client/fetcher';
import { selectContractAction } from '../redux/actions/contractActions';

const styles = {
    card: {
        minWidth: 275,
        marginTop: 5,
        marginBottom: 5,
    },
};

class ReportContainer extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            startBlockNumber: "",
            endBlockNumber: "",
            reportData: [],
            isLoading: false,
            displayReport: false,
            errorMessage: "",
        }
    }

    componentDidMount() {
        this.setState({
            startBlockNumber: 1,
            endBlockNumber: this.props.lastPersistedBlockNumber,
        })
    }

    componentWillUnmount() {
        this.props.dispatch(selectContractAction(""))
    }

    handleSelectedContractChange = (e) => {
        this.setState({ errorMessage: "" });
        this.props.dispatch(selectContractAction(e.target.value))
    };

    handleStartBlockChange = (e) => {
        this.setState({
            startBlockNumber: e.target.value,
            errorMessage: "",
        })
    };

    handleEndBlockChange = (e) => {
        this.setState({
            endBlockNumber: e.target.value,
            errorMessage: "",
        })
    };

    handleReport = () => {
        if (this.props.selectedContract === "") {
            this.setState({ errorMessage: "no contract selected" });
            return
        }
        if (this.state.startBlockNumber === "" || isNaN(this.state.startBlockNumber)) {
            this.setState({
                startBlockNumber: "",
                errorMessage: "invalid start block number",
            });
            return
        }
        if (this.state.endBlockNumber === "" || isNaN(this.state.endBlockNumber)) {
            this.setState({
                endBlockNumber: "",
                errorMessage: "invalid end block number",
            });
            return
        }
        this.setState({
            reportData: [],
            displayReport: true,
            isLoading: true,
            errorMessage: "",
        });
        getReportData(this.props.rpcEndpoint, this.props.selectedContract, this.state.startBlockNumber, this.state.endBlockNumber).then( (res) => {
            this.setState({
                reportData: res,
                isLoading: false,
            })
        }).catch( (e) => {
            this.setState({
                isLoading: false,
                errorMessage: e.toString(),
            })
        })
    };

    render(){
        return (
            <Card className={this.props.classes.card}>
                <CardContent>
                    <Typography variant="h6" align="center">
                        Report
                    </Typography>
                    <br/>
                    {
                        this.props.contracts.length === 0 &&
                        <h1 align="center">&lt; No Records Found &gt;</h1>
                    }
                    {
                        this.props.contracts.length !== 0 &&
                        <ReportForm
                            selectedContract={this.props.selectedContract}
                            startBlockNumber={this.state.startBlockNumber}
                            endBlockNumber={this.state.endBlockNumber}
                            contracts={this.props.contracts}
                            handleSelectedContractChange={this.handleSelectedContractChange}
                            handleStartBlockChange={this.handleStartBlockChange}
                            handleEndBlockChange={this.handleEndBlockChange}
                            handleReport={this.handleReport}
                        />
                    }
                    <br/>
                    {
                        this.state.errorMessage &&
                        <div>
                            <br/>
                            <Alert severity="error">{this.state.errorMessage}</Alert>
                        </div>
                    }
                    {
                        this.state.displayReport &&
                        <Report
                            parsedStorage={this.state.reportData}
                            isLoading={this.state.isLoading}
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
        lastPersistedBlockNumber: state.system.lastPersistedBlockNumber,
    }
};

export default connect(mapStateToProps)(withStyles(styles)(ReportContainer))