import simpleRestProvider from "ra-data-simple-rest";
import { fetchUtils } from 'ra-core';

export const baseDataProvider = simpleRestProvider(
  '/api', 
  fetchUtils.fetchJson, 
  'X-Total-Count'
);


export const dataProvider = {
  ...baseDataProvider,
  createTable: (mapId) => {
      return fetch(`/api/map/${mapId}/create`, { method: 'POST' })
          .then(response => response.json());
  },
  cloneTable: (mapId) => {
    return fetch(`/api/map/${mapId}/clone`, { method: 'POST' })
        .then(response => response.json());
  },
  restartAll: () => {
    return fetch(`/api/url/restart`, { method: 'POST' })
        .then(response => response.json());
  }
}
