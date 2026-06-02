import { defineStore } from 'pinia'
import {
  changeCurrentUserPassword,
  getAuthState,
  getCurrentUserProfile,
  login,
  logout,
  updateCurrentUserProfile,
  updateWechatPhone,
  wechatPhoneLogin,
} from '@/services/auth'
import { loginWithWechatProfile } from '@/services/wechat'
import { clearToken, getToken, setToken } from '@/utils/token'
import type {
  AppUserProfile,
  AuthState,
  ChangePasswordRequest,
  LoginRequest,
  UpdateAppUserProfileRequest,
} from '@/types/api'

interface UserState {
  token: string
  authState: AuthState | null
  profile: AppUserProfile | null
}

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    token: getToken() || '',
    authState: null,
    profile: null,
  }),
  getters: {
    isLoggedIn: (state) => Boolean(state.token || state.authState?.authenticated),
    displayName: (state) => state.authState?.displayName || state.profile?.name || state.profile?.wxNickname || '未登录',
  },
  actions: {
    restoreLoginState() {
      this.token = getToken() || ''
    },
    clearLocalLoginState() {
      this.token = ''
      this.authState = null
      this.profile = null
    },
    async loginWithPassword(form: LoginRequest) {
      const result = await login(form)
      this.token = result.token
      setToken(result.token)
      await this.refreshAuthState()
    },
    async loginWithWechat() {
      const result = await loginWithWechatProfile()
      this.token = result.token
      this.profile = result.user || null
      setToken(result.token)
      await this.refreshAuthState()
    },
    async loginWithWechatPhone(phoneCode: string) {
      const result = await wechatPhoneLogin({ code: phoneCode })
      this.token = result.token
      this.profile = result.user || null
      setToken(result.token)
      await this.refreshAuthState()
    },
    async refreshAuthState() {
      this.authState = await getAuthState()
      return this.authState
    },
    async refreshProfile() {
      this.profile = await getCurrentUserProfile()
      return this.profile
    },
    async updateProfile(data: UpdateAppUserProfileRequest) {
      this.profile = await updateCurrentUserProfile(data)
      await this.refreshAuthState()
      return this.profile
    },
    async changePassword(data: ChangePasswordRequest) {
      return changeCurrentUserPassword(data)
    },
    async bindWechatPhone(code: string) {
      await updateWechatPhone({ code })
      return this.refreshProfile()
    },
    async logout() {
      try {
        await logout()
      } finally {
        this.clearLocalLoginState()
        clearToken()
      }
    },
  },
})
