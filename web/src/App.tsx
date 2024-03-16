import {
  Admin,
  Resource,
} from "react-admin";
import { dataProvider } from "./dataProvider";
import { authProvider } from "./authProvider";
import { DbList, DbEdit, DbShow, DbCreate } from "./db";
import { UrlList, UrlEdit, UrlShow, UrlCreate } from "./url";
import { TblList, TblEdit, TblShow, TblCreate } from "./tbl";

export const App = () => (
  <Admin dataProvider={dataProvider} authProvider={authProvider}>
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
