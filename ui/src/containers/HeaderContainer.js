import React from 'react';
import { connect } from 'react-redux';
import { withStyles } from '@material-ui/core/styles';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import IconButton from '@material-ui/core/IconButton';
import SettingsIcon from '@material-ui/icons/Settings';
import SyncIcon from '@material-ui/icons/Sync';
import SyncDisabledIcon from '@material-ui/icons/SyncDisabled';
import Fade from '@material-ui/core/Fade';
import ButtonBase from '@material-ui/core/ButtonBase';
import SettingForm from '../components/SettingForm';
import MenuDropdown from '../components/MenuDropdown';
import { changePageAction } from '../redux/actions/pageActions';
import { connectAction, disconnectAction, updateEndpointAction, updateBlockNumberAction } from '../redux/actions/systemActions';
import { getBlockNumber } from '../client/fetcher';
import { HomePageId } from '../constants';

const styles = {
    grow: {
        flexGrow: 1,
    },
};

class HeaderContainer extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            formIsOpen: false,
            newRPCEndpoint: "",
            blockNumberAppear: true,
        }
    }

    componentDidMount() {
        this.timerID = setInterval(
            () => this.connectReporting(),
            1000
        );
    }

    componentWillUnmount() {
        clearInterval(this.timerID);
    }

    blockNumberBlinkEffect = () => {
        this.setState({blockNumberAppear: false});
        setTimeout(() => {
            this.setState({blockNumberAppear: true})
        }, 500)
    };

    connectReporting = () => {
        getBlockNumber(this.props.rpcEndpoint).then( (res) => {
            if (this.props.lastPersistedBlockNumber !== res) {
                this.blockNumberBlinkEffect();
                if (!this.props.isConnected) {
                    this.props.dispatch(connectAction())
                }
                this.props.dispatch(updateBlockNumberAction(res));
            }
        }).catch( (e) => {
            if (this.props.isConnected) {
                this.blockNumberBlinkEffect();
                this.props.dispatch(disconnectAction());
                this.props.dispatch(updateBlockNumberAction(""));
            }
        })
    };

    toHomePage = () => {
        this.handleMenuClick(HomePageId)
    };

    handleMenuClick = (page) => {
        this.props.dispatch(changePageAction(page))
    };

    handleOpenSetting = () => {
        this.setState({ formIsOpen: true })
    };

    handleCloseSetting = () => {
        this.setState({ formIsOpen: false })
    };

    handleRPCEndpointChange = (e) => {
        this.setState({ newRPCEndpoint: e.target.value })
    };

    handleSetRPCEndpoint = () => {
        this.props.dispatch(updateEndpointAction(this.state.newRPCEndpoint));
        this.connectReporting();
        this.setState({ formIsOpen: false })
    };

    render(){
        return (
            <AppBar position="static">
                <Toolbar>
                    <ButtonBase onClick={this.toHomePage}>
                        <Typography variant="h6" color="inherit">
                            <img src={require('../resources/quorum-logo.png')} width="40" height="20" alt="" />
                            &nbsp;
                            Quorum Reporting - Development
                            &nbsp;
                        </Typography>
                    </ButtonBase>
                    <span className={this.props.classes.grow}/>
                    <Typography variant="h4">
                        {
                            this.props.isConnected?<SyncIcon color="inherit" />:<SyncDisabledIcon color="error" />
                        }
                    </Typography>
                    <Fade in={this.state.blockNumberAppear} timeout={1000}>
                        <Typography variant="h5" color="inherit">
                            &nbsp;
                            {
                                this.props.isConnected?("# " + this.props.lastPersistedBlockNumber):"# N/A"
                            }
                            &nbsp;
                        </Typography>
                    </Fade>
                    <MenuDropdown
                        handleMenuClick={this.handleMenuClick}
                    />
                    <IconButton variant="h4" onClick={this.handleOpenSetting}>
                        <SettingsIcon color="action" />
                    </IconButton>
                    <SettingForm
                        rpcEndpoint={this.props.rpcEndpoint}
                        isOpen={this.state.formIsOpen}
                        handleCloseSetting={this.handleCloseSetting}
                        handleRPCEndpointChange={this.handleRPCEndpointChange}
                        handleSetRPCEndpoint={this.handleSetRPCEndpoint}
                        newRPCEndpoint={this.state.newRPCEndpoint}
                    />
                </Toolbar>
            </AppBar>
        )
    }
}

const mapStateToProps = state => {
    return {
        rpcEndpoint: state.system.rpcEndpoint,
        isConnected: state.system.isConnected,
        lastPersistedBlockNumber: state.system.lastPersistedBlockNumber,
    }
};

export default connect(mapStateToProps)(withStyles(styles)(HeaderContainer))
