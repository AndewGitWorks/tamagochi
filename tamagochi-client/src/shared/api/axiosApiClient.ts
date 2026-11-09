import axios, { type AxiosRequestConfig } from 'axios';
import { tokenStorage } from '../utils/tokenStorage';
import { store } from '@/store/store';
import { setUnauthenticated } from '@/features/auth/store';

const AUTH_ENDPOINTS = ['/login', '/register', '/refresh'];

export const createAxiosApiClient = (baseURL: string) => {
  const client = axios.create({ baseURL });

  let isRefreshing = false;
  let failedQueue: Array<{ resolve: (token: string) => void; reject: (error: unknown) => void }> =
    [];

  const processQueue = (error: unknown, token: string | null = null) => {
    failedQueue.forEach((promise) => {
      if (error) {
        promise.reject(error);
      } else if (token) {
        promise.resolve(token);
      }
    });

    failedQueue = [];
  };

  client.interceptors.request.use(
    (config) => {
      const isAuthRoute = AUTH_ENDPOINTS.some((route) => config.url?.includes(route));

      if (isAuthRoute) {
        return config;
      }

      const token = tokenStorage.getAccessToken();

      if (token) {
        config.headers.set('Authorization', `Bearer ${token}`);
      }
      return config;
    },
    (error) => {
      return Promise.reject(error);
    },
  );

  client.interceptors.response.use(
    (response) => response,
    async (error) => {
      const originalRequest = error.config;

      const isAuthRoute = AUTH_ENDPOINTS.some((route) => originalRequest.url?.includes(route));

      if (error.response?.status === 401 && !originalRequest._retry && !isAuthRoute) {
        originalRequest._retry = true;
        if (isRefreshing) {
          return new Promise((resolve, reject) => {
            failedQueue.push({ resolve, reject });
          })
            .then((token) => {
              originalRequest.headers.Authorization = `Bearer ${token}`;
              return client(originalRequest);
            })
            .catch((err) => Promise.reject(err));
        }

        isRefreshing = true;

        try {
          const refreshToken = tokenStorage.getRefreshToken();
          if (!refreshToken) throw new Error('No refresh token');

          const response = await axios.post(`${baseURL}/refresh`, null, {
            headers: { Authorization: `Bearer ${refreshToken}` },
          });

          const tokens = response.data;

          tokenStorage.setTokens({
            accessToken: tokens.access_token,
            refreshToken: tokens.refresh_token,
          });

          originalRequest.headers.set('Authorization', `Bearer ${tokens.access_token}`);
          processQueue(null, tokens.access_token);
          return client(originalRequest);
        } catch (refreshError) {
          processQueue(refreshError, null);

          tokenStorage.clearTokens();
          store.dispatch(setUnauthenticated());
          return Promise.reject(refreshError);
        } finally {
          isRefreshing = false;
        }
      }

      return Promise.reject(error);
    },
  );

  return {
    get: async <T>(url: string, config?: AxiosRequestConfig) => {
      const response = await client.get<T>(url, config);

      return response.data;
    },
    post: async <T>(url: string, data?: unknown, config?: AxiosRequestConfig) => {
      const response = await client.post<T>(url, data, config);
      return response.data;
    },
    put: async <T>(url: string, data?: unknown, config?: AxiosRequestConfig) => {
      const response = await client.put<T>(url, data, config);
      return response.data;
    },
  };
};
