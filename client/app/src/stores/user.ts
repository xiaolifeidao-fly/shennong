import { defineStore } from 'pinia'
import { getAuthState, getCurrentUserProfile, login, logout, updateCurrentUserProfile, updateWechatPhone } from '@/services/auth'
import { loginWithWechatCodeOnly, loginWithWechatProfile } from '@/services/wechat'
import { clearToken, setToken } from '@/utils/token'
import type { AppUserProfile, AuthState, LoginRequest, UpdateAppUserProfileRequest } from '@/types/api'

interface UserState {
  token: string
  authState: AuthState | null
  profile: AppUserProfile | null
}

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    token: '',
    authState: null,
    profile: null,
  }),
  getters: {
    isLoggedIn: (state) => Boolean(state.token || state.authState?.authenticated),
    displayName: (state) => state.authState?.displayName || state.profile?.name || state.profile?.wxNickname || '未登录',
  },
  actions: {
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
      const result = await loginWithWechatCodeOnly()
      this.token = result.token
      this.profile = result.user || null
      setToken(result.token)
      await this.bindWechatPhone(phoneCode)
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
    async bindWechatPhone(code: string) {
      await updateWechatPhone({ code })
      return this.refreshProfile()
    },
    async logout() {
      try {
        await logout()
      } finally {
        this.token = ''
        this.authState = null
        this.profile = null
        clearToken()
      }
    },
  },
})
