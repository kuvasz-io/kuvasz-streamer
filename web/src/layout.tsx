import * as React from 'react';
import { Layout, LayoutProps } from 'react-admin';
import AppBar from './app-bar';

export default (props: LayoutProps) => (
    <Layout {...props} appBar={AppBar} />
);