import React from 'react'
import { connect } from 'react-redux';
import SearchContainer from '../containers/SearchContainer';
import HeaderContainer from '../containers/HeaderContainer';
import ContractInfoContainer from '../containers/ContractInfoContainer';
import ReportContainer from '../containers/ReportContainer';
import { ContractPageId, HomePageId, ReportPageId } from '../constants';

class HomePage extends React.Component {

    renderPageContent = () => {
        switch (this.props.page) {
            case HomePageId:
                return <SearchContainer/>;
            case ContractPageId:
                return <ContractInfoContainer/>;
            case ReportPageId:
                return <ReportContainer/>;
            default:
                return <SearchContainer/>;
        }
    };

    render() {
        return (
            <div>
                <HeaderContainer/>
                { this.renderPageContent() }
            </div>
        )
    }
}

const mapStateToProps = state => {
    return {
        page: state.user.page,
    }
};

export default connect(mapStateToProps)(HomePage)