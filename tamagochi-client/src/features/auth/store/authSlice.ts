import { createSlice } from '@reduxjs/toolkit';
import { restoreSession } from './authThunk';

type AuthStatus = 'initializing' | 'authenticated' | 'unauthenticated';

type AuthState = {
  authStatus: AuthStatus;
};

const initialState: AuthState = {
  authStatus: 'initializing',
};

export const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setAuthenticated: (state) => {
      state.authStatus = 'authenticated';
    },
    setUnauthenticated: (state) => {
      state.authStatus = 'unauthenticated';
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(restoreSession.pending, (state) => {
        state.authStatus = 'initializing';
      })
      .addCase(restoreSession.fulfilled, (state) => {
        state.authStatus = 'authenticated';
      })
      .addCase(restoreSession.rejected, (state) => {
        state.authStatus = 'unauthenticated';
      });
  },
});

export const { setAuthenticated, setUnauthenticated } = authSlice.actions;
export default authSlice.reducer;
