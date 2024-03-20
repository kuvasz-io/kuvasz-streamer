import simpleRestProvider from "ra-data-simple-rest";
import { fetchUtils } from 'ra-core';

export const baseDataProvider = simpleRestProvider(
  '/api', 
  fetchUtils.fetchJson, 
  'X-Total-Count'
);


export const dataProvider = {
  ...baseDataProvider,
  urlTables: (urlId) => {
      return fetch(`/api/url/${urlId}/tables`, { method: 'GET' })
          .then(response => response.json());
  },
}
