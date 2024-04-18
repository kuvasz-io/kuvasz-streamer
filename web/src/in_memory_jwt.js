const inMemoryJWTManager = () => {
    let inMemoryJWT = null;
    let isRefreshing = null;
    let logoutEventName = 'ra-logout';
    let refreshEndpoint = '/refresh-token';
    let refreshTimeOutId;

    const setLogoutEventName = name => logoutEventName = name;
    const setRefreshTokenEndpoint = endpoint => refreshEndpoint = endpoint;

    // This countdown feature is used to renew the JWT before it's no longer valid
    // in a way that is transparent to the user.
    const refreshToken = (delay) => {
        refreshTimeOutId = window.setTimeout(
            getRefreshedToken,
            delay * 1000 - 5000
        ); // Validity period of the token in seconds, minus 5 seconds
    };

    const abortRefreshToken = () => {
        if (refreshTimeOutId) {
            window.clearTimeout(refreshTimeOutId);
        }
    };

    const waitForTokenRefresh = () => {
        if (!isRefreshing) {
            return Promise.resolve();
        }
        return isRefreshing.then(() => {
            isRefreshing = null;
            return true;
        });
    }

    // The method make a call to the refresh-token endpoint
    // If there is a valid cookie, the endpoint will set a fresh jwt in memory.
    const getRefreshedToken = () => {
        const request = new Request(refreshEndpoint, {
            method: 'GET',
            headers: new Headers({ 'Content-Type': 'application/json' }),
            credentials: 'include',
        });

        isRefreshing = fetch(request)
            .then((response) => {
                if (response.status !== 200) {
                    eraseToken();
                    console.log(
                        'Token renewal failure'
                    );
                    return { token: null };
                }
                return response.json();
            })
            .then(({ token, tokenExpiry }) => {
                if (token) {
                    setToken(token, tokenExpiry);
                    return true;
                }
                eraseToken();
                return false;
            });

        return isRefreshing;
    };


    const getToken = () => inMemoryJWT;

    const setToken = (token, delay) => {
        inMemoryJWT = token;
        refreshToken(delay);
        return true;
    };

    const eraseToken = () => {
        inMemoryJWT = null;
        abortRefreshToken();
        window.localStorage.setItem(logoutEventName, Date.now());
        return true;
    }

    // This listener will allow to disconnect a session of ra started in another tab
    window.addEventListener('storage', (event) => {
        if (event.key === logoutEventName) {
            inMemoryJWT = null;
        }
    });

    return {
        eraseToken,
        getRefreshedToken,
        getToken,
        setLogoutEventName,
        setRefreshTokenEndpoint,
        setToken,
        waitForTokenRefresh,
    }
};

export default inMemoryJWTManager();