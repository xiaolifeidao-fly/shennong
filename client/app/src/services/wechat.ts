import { wechatLogin } from '@/services/auth'
import type { WechatLoginRequest } from '@/types/api'

interface UniLoginResult {
  code?: string
}

interface UniUserProfileResult {
  rawData?: string
  signature?: string
  userInfo?: WechatLoginRequest['userInfo']
}

function uniLogin() {
  return new Promise<UniLoginResult>((resolve, reject) => {
    uni.login({
      provider: 'weixin',
      success: resolve,
      fail: reject,
    })
  })
}

function getUserProfile() {
  return new Promise<UniUserProfileResult>((resolve, reject) => {
    uni.getUserProfile({
      desc: '用于完善业务员身份资料',
      success: resolve,
      fail: reject,
    })
  })
}

export async function loginWithWechatProfile() {
  const [loginResult, profileResult] = await Promise.all([uniLogin(), getUserProfile()])
  if (!loginResult.code) {
    throw new Error('未获取到微信登录 code')
  }

  return wechatLogin({
    code: loginResult.code,
    rawData: profileResult.rawData || '',
    signature: profileResult.signature || '',
    userInfo: profileResult.userInfo || {},
  })
}

export async function loginWithWechatCodeOnly() {
  const loginResult = await uniLogin()
  if (!loginResult.code) {
    throw new Error('未获取到微信登录 code')
  }

  return wechatLogin({
    code: loginResult.code,
    rawData: '',
    signature: '',
    userInfo: {},
  })
}
