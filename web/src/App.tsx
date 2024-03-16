import {
  Admin,
  Resource,
  ListGuesser,
  EditGuesser,
  ShowGuesser,
} from "react-admin";
import { dataProvider } from "./dataProvider";
import { authProvider } from "./authProvider";
import { TblList, TblEdit, TblShow } from "./tbl";
import { DbList, DbEdit, DbShow } from "./db";
import { UrlList, UrlEdit, UrlShow } from "./url";

export const App = () => (
  <Admin dataProvider={dataProvider} authProvider={authProvider}>
    <Resource
      name="db"
      options={{ label: 'Database schemas' }}
      list={DbList}
      edit={DbEdit}
      show={DbShow}
    />
    <Resource
      name="url"
      options={{ label: 'Sources' }}
      list={UrlList}
      edit={UrlEdit}
      show={UrlShow}
    />
    <Resource
      name="tbl"
      options={{ label: 'Tables' }}
      list={TblList}
      edit={TblEdit}
      show={TblShow}
    />
  </Admin>
);
