import inMemoryJWT from './in_memory_jwt';

export const authProvider = {
    login: ({username, password}: any) => {
        const request = new Request('http://turing:8000/login', {
            method: 'POST',
            body: JSON.stringify({ 'username': username, 'password': password }),
            headers: new Headers({ 'Content-Type': 'application/json' }),
            credentials: 'include',
        });
        inMemoryJWT.setRefreshTokenEndpoint('http://turing:8000/refresh-token');
        return fetch(request)
            .then((response) => {
                if (response.status < 200 || response.status >= 300) {
                    throw new Error(response.statusText);
                }
                return response.json();
            })
            .then(({ token, tokenExpiry }) => {
                return inMemoryJWT.setToken(token, tokenExpiry);
            });
    },

    logout: () => {
        const request = new Request('http://turing:8000/logout', {
            method: 'POST',
            headers: new Headers({ 'Content-Type': 'application/json' }),
            credentials: 'include',
        });
        inMemoryJWT.eraseToken();

        return fetch(request).then(() => '/login');
    },

    checkAuth: () => {
        return inMemoryJWT.waitForTokenRefresh().then(() => {
            return inMemoryJWT.getToken() ? Promise.resolve() : Promise.reject();
        });
    },

    checkError: (error: any) => {
        const status = error.status;
        if (status === 401 || status === 403) {
            inMemoryJWT.eraseToken();
            return Promise.reject();
        }
        return Promise.resolve();
    },

    getPermissions: () => {
        return inMemoryJWT.waitForTokenRefresh().then(() => {
            return inMemoryJWT.getToken() ? Promise.resolve() : Promise.reject();
        });
    },
};

export default authProvider; 