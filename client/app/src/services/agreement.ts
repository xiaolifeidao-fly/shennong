import { http } from './request'
import type { AgreementResponse, AgreementStatusResponse } from '@/types/api'

function getWechatLoginCode() {
  return new Promise<string>((resolve, reject) => {
    uni.login({
      provider: 'weixin',
      success: (res) => {
        if (res.code) {
          resolve(res.code)
        } else {
          reject(new Error('未获取到微信登录 code'))
        }
      },
      fail: () => reject(new Error('微信登录失败')),
    })
  })
}

// 获取协议正文与当前版本。
export function getAgreement() {
  return http.get<AgreementResponse>('/agreement')
}

// 静默判断当前微信用户（openid）是否已同意当前版本协议。
export async function checkAgreementStatus() {
  const code = await getWechatLoginCode()
  return http.post<AgreementStatusResponse, { code: string }>('/agreement/status', { code })
}

// 记录当前微信用户同意协议。
export async function agreeAgreement() {
  const code = await getWechatLoginCode()
  return http.post<AgreementStatusResponse, { code: string }>('/agreement/agree', { code })
}
