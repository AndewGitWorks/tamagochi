export type LoginRequestDTO = {
  login: string;
  password: string;
};

export type TokenPairDTO = {
  access_token: string;
  access_expires_at: string;
  refresh_token: string;
  refresh_expires_at: string;
  token_type: 'Bearer';
};

export type LoginResponseDTO = TokenPairDTO;

export type RegistrationRequestDTO = {
  login: string;
  name: string;
  surname: string;
  email: string;
  phone_number: string;
  password: string;
};

export type RegistrationResponseDTO = TokenPairDTO;

export type RefreshRequestDTO = {
  refreshToken: string;
};

export type RefreshResponseDTO = TokenPairDTO;
