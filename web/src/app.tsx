import {
  Admin,
  Resource,
} from "react-admin";
import { Route } from 'react-router-dom';

import { dataProvider } from "./data-provider";
import { authProvider } from "./auth-provider";
import { softLightTheme } from './soft-theme';
import Layout  from "./layout";

import { DbList, DbView, DbEdit, DbShow, DbCreate } from "./db";
import { UrlList, UrlEdit, UrlShow, UrlCreate } from "./url";
import { TblList, TblEdit, TblShow, TblCreate } from "./tbl";
import { MapList, MapEdit } from "./map";
// import { MapList, MapEdit, MapShow } from "./map";

export const App = () => (
  <Admin 
    disableTelemetry
    dataProvider={dataProvider} 
    authProvider={authProvider}
    theme={softLightTheme}
    layout={Layout}
    >
      {permissions => (
      <>
        <Resource
          name="db"
          options={{ label: 'Databases' }}
          list={permissions === 'admin' ? DbList : DbView}
          edit={permissions === 'admin' ? DbEdit : null}
          show={DbShow}
          create={permissions === 'admin' ? DbCreate: null}
        />
        <Resource
          name="url"
          options={{ label: 'Sources' }}
          list={UrlList}
          edit={permissions === 'admin' ? UrlEdit : null}
          show={UrlShow}
          create={permissions === 'admin' ? UrlCreate : null}
        />
        <Resource
          name="tbl"
          options={{ label: 'Tables' }}
          list={TblList}
          edit={permissions === 'admin' ? TblEdit : null}
          show={TblShow}
          create={permissions === 'admin' ? TblCreate : null}
        />
        <Resource 
          name="map"
          options={{ label: 'Map'}}
          list={MapList}
          edit={permissions === 'admin' ? MapEdit : null}
          // show={MapShow}
        />
    </>
      )}
  </Admin>
);
