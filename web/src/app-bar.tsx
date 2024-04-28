import * as React from 'react';
import { AppBar, Button, TitlePortal, useDataProvider, useRefresh } from 'react-admin';
import { Box, useMediaQuery, Theme } from '@mui/material';

import { useMutation } from 'react-query';

import Logo from "./logo";

const RestartAllButton = () => {
    const dataProvider = useDataProvider();
    const refresh = useRefresh();

    const { mutate, isLoading } = useMutation(
        () => dataProvider.restartAll().then(() => refresh()));
        return <Button 
                label="Restart" 
                onClick={() => mutate()}
                disabled={isLoading} />;
    return null;
};


const CustomAppBar = () => {
    const isLargeEnough = useMediaQuery<Theme>(theme =>
        theme.breakpoints.up('sm')
    );
    return (
        <AppBar color="secondary">
            <TitlePortal />
            {isLargeEnough && <Logo />}
            {isLargeEnough && <Box component="span" sx={{ flex: 1 }} />}
            <RestartAllButton/>
        </AppBar>
    );
};

export default CustomAppBar;