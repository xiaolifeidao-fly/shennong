import { http } from './request'
import type {
  AppUserProfile,
  AuthState,
  ChangePasswordRequest,
  LoginRequest,
  LoginResponse,
  UpdateAppUserProfileRequest,
  WechatLoginRequest,
  WechatLoginResponse,
  WechatPhoneRequest,
  WechatPhoneResponse,
} from '@/types/api'

export function login(data: LoginRequest) {
  return http.post<LoginResponse, LoginRequest>('/login', data, { showLoading: true })
}

export function wechatLogin(data: WechatLoginRequest) {
  return http.post<WechatLoginResponse, WechatLoginRequest>('/wechat-login', data, { showLoading: true })
}

export function wechatPhoneLogin(data: WechatPhoneRequest) {
  return http.post<WechatLoginResponse, WechatPhoneRequest>('/wechat-phone-login', data, { showLoading: true })
}

export function logout() {
  return http.post<{ loggedOut: boolean }>('/logout')
}

export function getAuthState() {
  return http.get<AuthState>('/auth-state')
}

export function getCurrentUserProfile() {
  return http.get<AppUserProfile>('/app-user-profile')
}

export function updateCurrentUserProfile(data: UpdateAppUserProfileRequest) {
  return http.put<AppUserProfile, UpdateAppUserProfileRequest>('/app-user-profile', data, { showLoading: true })
}

export function changeCurrentUserPassword(data: ChangePasswordRequest) {
  return http.put<{ changed: boolean }, ChangePasswordRequest>('/app-user-profile/password', data, { showLoading: true })
}

export function updateWechatPhone(data: WechatPhoneRequest) {
  return http.post<WechatPhoneResponse, WechatPhoneRequest>('/app-user-profile/wechat-phone', data, { showLoading: true })
}
