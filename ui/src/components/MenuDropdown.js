import React from 'react';
import Menu from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';
import IconButton from '@material-ui/core/IconButton';
import MenuIcon from '@material-ui/icons/Menu';
import { ContractPageId, HomePageId, ReportPageId } from '../constants';

function MenuDropdown(props) {
    const [anchorEl, setAnchorEl] = React.useState(null);

    const handleClick = (e) => {
        setAnchorEl(e.currentTarget);
    };

    const handleClose = (e) => {
        setAnchorEl(null);
    };

    const handleMenuHome = (e) => {
        props.handleMenuClick("Home");
        handleClose(e);
    };

    const handleMenuContract = (e) => {
        props.handleMenuClick("Contract");
        handleClose(e);
    };

    const handleMenuReport = (e) => {
        props.handleMenuClick("Report");
        handleClose(e);
    };

    return (
        <div>
            <IconButton variant="h4" onClick={handleClick}>
                <MenuIcon color="action" />
            </IconButton>
            <Menu
                id="simple-menu"
                anchorEl={anchorEl}
                keepMounted
                open={Boolean(anchorEl)}
                onClose={handleClose}
            >
                <MenuItem value={HomePageId} onClick={handleMenuHome}>Home</MenuItem>
                <MenuItem value={ContractPageId} onClick={handleMenuContract}>Contract</MenuItem>
                <MenuItem value={ReportPageId} onClick={handleMenuReport}>Report</MenuItem>
            </Menu>
        </div>
    );
}

export default MenuDropdown