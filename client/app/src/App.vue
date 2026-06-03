<script setup lang="ts">
import { onLaunch, onShow } from '@dcloudio/uni-app'
import { installAuthGuard, redirectLoggedInUser, ensureAuthenticated } from '@/utils/authGuard'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const apiTransport = import.meta.env.VITE_APP_API_TRANSPORT || 'https'
const cloudEnv = import.meta.env.VITE_APP_CLOUD_ENV

declare const wx: {
  cloud?: {
    init(options: {
      env: string
      traceUser?: boolean
    }): void
  }
} | undefined

installAuthGuard()

function initWechatCloud() {
  if (apiTransport !== 'cloud') {
    return
  }
  if (!cloudEnv) {
    console.warn('VITE_APP_CLOUD_ENV is empty')
    return
  }
  if (typeof wx === 'undefined' || !wx.cloud?.init) {
    console.warn('wx.cloud.init is unavailable')
    return
  }

  wx.cloud.init({
    env: cloudEnv,
    traceUser: true,
  })
}

onLaunch(() => {
  initWechatCloud()
  userStore.restoreLoginState()
  ensureAuthenticated()
  redirectLoggedInUser()
})

onShow(() => {
  userStore.restoreLoginState()
  ensureAuthenticated()
  redirectLoggedInUser()
})

uni.$on('auth:expired', () => {
  userStore.clearLocalLoginState()
})
</script>

<style lang="scss">
page {
  min-height: 100%;
  background: #eef4ed;
  color: #172018;
  font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', Helvetica, Arial, sans-serif;
}

view,
text,
button,
input,
textarea {
  box-sizing: border-box;
}

button {
  margin: 0;
  transition: transform 0.18s ease, box-shadow 0.18s ease, opacity 0.18s ease;
}

button::after {
  border: 0;
}

button:active {
  transform: scale(0.97);
  opacity: 0.9;
}

.page {
  min-height: 100vh;
  padding: 32rpx 28rpx calc(120rpx + env(safe-area-inset-bottom));
  background:
    radial-gradient(circle at 15% 0%, rgba(255, 186, 73, 0.2), transparent 260rpx),
    radial-gradient(circle at 88% 8%, rgba(35, 122, 75, 0.18), transparent 300rpx),
    linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(238, 244, 237, 0) 380rpx),
    #eef4ed;
  animation: page-rise 0.38s ease both;
}

.section {
  margin-top: 24rpx;
}

.card {
  border: 1rpx solid rgba(255, 255, 255, 0.72);
  border-radius: 24rpx;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 18rpx 44rpx rgba(31, 47, 31, 0.08);
}

@keyframes page-rise {
  from {
    opacity: 0;
    transform: translateY(18rpx);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
