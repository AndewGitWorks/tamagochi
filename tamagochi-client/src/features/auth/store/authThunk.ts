import { createAsyncThunk } from '@reduxjs/toolkit';
import { tokenStorage } from '@/shared/utils/tokenStorage';
import { refreshRequest } from '../services/authService';

export const restoreSession = createAsyncThunk(
  'auth/restoreSession',
  async (_, { rejectWithValue }) => {
    const refreshToken = tokenStorage.getRefreshToken();

    if (!refreshToken) {
      return rejectWithValue('No refresh token found');
    }

    try {
      const tokens = await refreshRequest({
        headers: {
          Authorization: `Bearer ${refreshToken}`,
        },
      });

      tokenStorage.setTokens({
        accessToken: tokens.access_token,
        refreshToken: tokens.refresh_token,
      });

      return true;
    } catch {
      tokenStorage.clearTokens();
      return rejectWithValue('Session expired or invalid');
    }
  },
);
