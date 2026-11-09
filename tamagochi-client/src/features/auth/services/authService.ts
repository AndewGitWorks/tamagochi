import { axiosApi } from '@/shared/api';
import {
  type RegistrationResponseDTO,
  type LoginRequestDTO,
  type LoginResponseDTO,
  type RegistrationRequestDTO,
  type RefreshResponseDTO,
} from './auth.types';
import type { AxiosRequestConfig } from 'axios';

export const loginRequest = async (credentials: LoginRequestDTO) => {
  const res = await axiosApi.post<LoginResponseDTO>('/login', credentials);
  return res;
};

export const registrationRequest = async (credentials: RegistrationRequestDTO) => {
  const res = await axiosApi.post<RegistrationResponseDTO>('/register', credentials);
  return res;
};

export const refreshRequest = async (config: AxiosRequestConfig) => {
  const res = await axiosApi.post<RefreshResponseDTO>('/refresh', null, config);
  return res;
};
