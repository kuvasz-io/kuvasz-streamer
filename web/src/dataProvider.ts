import simpleRestProvider from "ra-data-simple-rest";
import { fetchUtils } from 'ra-core';

export const dataProvider = simpleRestProvider(
  'http://turing:8000/api', 
  fetchUtils.fetchJson, 
  'X-Total-Count'
);
