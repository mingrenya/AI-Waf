import { get, post } from './index';
import { 
  UserLoginRequest, 
  LoginResponseData,
  UserPasswordResetRequest,
  GetUserInfoResponseData
} from '@/types/auth';

// Auth API endpoints
export const authApi = {
  // Login user
  login: (credentials: UserLoginRequest): Promise<LoginResponseData> => {
    return post<LoginResponseData>('/auth/login', credentials);
  },
  
  // Get current user info
  getCurrentUser: (): Promise<GetUserInfoResponseData> => {
    return get<GetUserInfoResponseData>('/auth/me');
  },
  
  // Reset password
  resetPassword: (data: UserPasswordResetRequest): Promise<void> => {
    return post<void>('/auth/reset-password', data);
  }
}; 