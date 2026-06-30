export interface ApiResponse<T> {
  success: boolean
  code: number
  data: T
  message: string
  error: string | null
}

export interface PageResponse<T> {
  total: number
  data: T[]
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
}

export interface WechatUserInfo {
  nickName?: string
  avatarUrl?: string
  gender?: number
  country?: string
  province?: string
  city?: string
  language?: string
}

export interface WechatLoginRequest {
  code: string
  rawData?: string
  signature?: string
  userInfo?: WechatUserInfo
}

export interface WechatLoginResponse {
  token: string
  user?: AppUserProfile
}

export interface UpdateAppUserProfileRequest {
  name?: string
  idNumber?: string
  email?: string
  phone?: string
  department?: string
  remark?: string
  wxNickname?: string
  wxAvatar?: string
}

export interface ChangePasswordRequest {
  oldPassword?: string
  newPassword: string
}

export interface WechatPhoneRequest {
  code: string
}

export interface WechatPhoneResponse {
  phone: string
}

export interface AuthState {
  authenticated: boolean
  username: string
  displayName: string
}

export interface AgreementDocument {
  key: string
  title: string
  content: string
}

export interface AgreementResponse {
  version: string
  documents: AgreementDocument[]
}

export interface AgreementStatusResponse {
  agreed: boolean
  version: string
}

export interface AppUserProfile {
  id?: number
  name?: string
  idNumber?: string
  username?: string
  avatar?: string
  phone?: string
  openUid?: string
  unionId?: string
  wxNickname?: string
  wxAvatar?: string
  wxLastLoginTime?: string
  hasOriginPassword?: boolean
  stationId?: number
  stationName?: string
  [key: string]: unknown
}
