type TokenPair = {
  accessToken: string;
  refreshToken: string;
};

export type TokenStoragePolicy = 'session' | 'local';

const ACCESS_TOKEN_KEY = 'accessToken';
const REFRESH_TOKEN_KEY = 'refreshToken';

const getStorage = (policy: TokenStoragePolicy) => {
  return policy === 'session' ? sessionStorage : localStorage;
};

const getActiveStorage = () => {
  return sessionStorage.getItem(REFRESH_TOKEN_KEY) ? sessionStorage : localStorage;
};

const clearTokens = () => {
  [sessionStorage, localStorage].forEach((storage) => {
    storage.removeItem(ACCESS_TOKEN_KEY);
    storage.removeItem(REFRESH_TOKEN_KEY);
  });
};

const setTokens = (
  { accessToken, refreshToken }: TokenPair,
  policy?: TokenStoragePolicy,
) => {
  const storage = policy ? getStorage(policy) : getActiveStorage();

  clearTokens();
  storage.setItem(ACCESS_TOKEN_KEY, accessToken);
  storage.setItem(REFRESH_TOKEN_KEY, refreshToken);
};

const getAccessToken = () => {
  return getActiveStorage().getItem(ACCESS_TOKEN_KEY);
};

const getRefreshToken = () => {
  return getActiveStorage().getItem(REFRESH_TOKEN_KEY);
};

export const tokenStorage = {
  setTokens,
  getAccessToken,
  getRefreshToken,
  clearTokens,
};
