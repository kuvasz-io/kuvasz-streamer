import {
  Admin,
  Resource,
  useStore
} from "react-admin";
import { dataProvider } from "./dataProvider";
import { authProvider } from "./authProvider";
import { softLightTheme } from './softTheme';
import Layout  from "./layout";

import { DbList, DbEdit, DbShow, DbCreate } from "./db";
import { UrlList, UrlEdit, UrlShow, UrlCreate } from "./url";
import { TblList, TblEdit, TblShow, TblCreate } from "./tbl";

export const App = () => (
  <Admin 
    disableTelemetry
    dataProvider={dataProvider} 
    authProvider={authProvider}
    theme={softLightTheme}
    layout={Layout}
    >
    <Resource
      name="db"
      options={{ label: 'Database schemas' }}
      list={DbList}
      edit={DbEdit}
      show={DbShow}
      create={DbCreate}
    />
    <Resource
      name="url"
      options={{ label: 'Sources' }}
      list={UrlList}
      edit={UrlEdit}
      show={UrlShow}
      create={UrlCreate}
    />
    <Resource
      name="tbl"
      options={{ label: 'Tables' }}
      list={TblList}
      edit={TblEdit}
      show={TblShow}
      create={TblCreate}
    />
  </Admin>
);
