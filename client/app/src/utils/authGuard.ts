import { clearToken, getToken } from './token'

const LOGIN_URL = '/pages/login/index'
const HOME_URL = '/pages/index/index'
const PUBLIC_ROUTES = new Set(['pages/login/index'])

let installed = false

function normalizeRoute(url = '') {
  return url.replace(/^\//, '').split('?')[0]
}

function getCurrentRoute() {
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1]
  return normalizeRoute(currentPage?.route || '')
}

function isPublicRoute(url = '') {
  return PUBLIC_ROUTES.has(normalizeRoute(url))
}

export function hasLoginToken() {
  return Boolean(getToken())
}

export function redirectToLogin() {
  if (getCurrentRoute() === normalizeRoute(LOGIN_URL)) {
    return
  }

  uni.reLaunch({ url: LOGIN_URL })
}

export function redirectLoggedInUser() {
  if (hasLoginToken() && getCurrentRoute() === normalizeRoute(LOGIN_URL)) {
    uni.switchTab({ url: HOME_URL })
  }
}

export function ensureAuthenticated(url?: string) {
  if (hasLoginToken() || isPublicRoute(url || getCurrentRoute())) {
    return true
  }

  redirectToLogin()
  return false
}

export function expireLogin() {
  clearToken()
  uni.$emit('auth:expired')
  redirectToLogin()
}

export function installAuthGuard() {
  if (installed) {
    return
  }

  installed = true

  const guard = {
    invoke(args: { url?: string }) {
      return ensureAuthenticated(args.url)
    },
  }

  uni.addInterceptor('navigateTo', guard)
  uni.addInterceptor('redirectTo', guard)
  uni.addInterceptor('reLaunch', guard)
  uni.addInterceptor('switchTab', guard)
}
