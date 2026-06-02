<template>
  <view class="page change-password-page">
    <view class="header">
      <text class="title">修改密码</text>
      <text class="subtitle">设置新的登录密码后，请妥善保管。</text>
    </view>

    <view class="card password-card">
      <view v-if="hasOriginPassword" class="field">
        <text class="label">原密码</text>
        <input v-model="form.oldPassword" class="input" password placeholder="请输入原密码" />
      </view>
      <view class="field">
        <text class="label">新密码</text>
        <input v-model="form.newPassword" class="input" password placeholder="至少 6 位" />
      </view>
      <view class="field">
        <text class="label">确认密码</text>
        <input v-model="form.confirmPassword" class="input" password placeholder="再次输入新密码" />
      </view>

      <button class="submit icon-text-btn" :loading="submitting" @click="savePassword">
        <text class="check-icon"></text>
        <text>保存密码</text>
      </button>
      <button class="skip icon-text-btn" @click="goMine">
        <text class="back-icon"></text>
        <text>返回我的</text>
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/user'
import type { ChangePasswordRequest } from '@/types/api'

const userStore = useUserStore()
const submitting = ref(false)
const hasOriginPassword = computed(() => Boolean(userStore.profile?.hasOriginPassword))
const form = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

onShow(() => {
  void loadProfile()
})

async function loadProfile() {
  try {
    await userStore.refreshProfile()
  } catch (error) {
    uni.showToast({
      title: error instanceof Error ? error.message : '获取资料失败',
      icon: 'none',
    })
  }
}

async function savePassword() {
  const oldPassword = form.oldPassword.trim()
  const newPassword = form.newPassword.trim()
  const confirmPassword = form.confirmPassword.trim()

  if (hasOriginPassword.value && !oldPassword) {
    uni.showToast({ title: '请输入原密码', icon: 'none' })
    return
  }
  if (!newPassword) {
    uni.showToast({ title: '请输入新密码', icon: 'none' })
    return
  }
  if (newPassword.length < 6) {
    uni.showToast({ title: '密码至少 6 位', icon: 'none' })
    return
  }
  if (newPassword !== confirmPassword) {
    uni.showToast({ title: '两次输入的密码不一致', icon: 'none' })
    return
  }

  submitting.value = true
  try {
    const payload: ChangePasswordRequest = {
      newPassword,
    }
    if (hasOriginPassword.value) {
      payload.oldPassword = oldPassword
    }

    await userStore.changePassword(payload)
    await userStore.refreshProfile()
    form.oldPassword = ''
    form.newPassword = ''
    form.confirmPassword = ''
    uni.showToast({ title: '密码已保存', icon: 'success' })
    goMine()
  } catch (error) {
    uni.showToast({
      title: error instanceof Error ? error.message : '密码保存失败',
      icon: 'none',
    })
  } finally {
    submitting.value = false
  }
}

function goMine() {
  uni.switchTab({ url: '/pages/mine/index' })
}
</script>

<style lang="scss" scoped>
.change-password-page {
  display: flex;
  flex-direction: column;
  gap: 32rpx;
}

.header {
  padding-top: 48rpx;
}

.title,
.subtitle {
  display: block;
}

.title {
  color: #172018;
  font-size: 44rpx;
  font-weight: 800;
}

.subtitle {
  margin-top: 12rpx;
  color: #6d776c;
  font-size: 28rpx;
  line-height: 1.5;
}

.password-card {
  padding: 32rpx;
}

.field {
  margin-top: 28rpx;
}

.field:first-child {
  margin-top: 0;
}

.label {
  display: block;
  margin-bottom: 14rpx;
  color: #445044;
  font-size: 26rpx;
  font-weight: 700;
}

.input {
  width: 100%;
  height: 88rpx;
  padding: 0 24rpx;
  border: 1rpx solid #e2e8dd;
  border-radius: 18rpx;
  background: #fbfcfa;
  color: #172018;
  font-size: 28rpx;
}

.submit,
.skip {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  width: 100%;
  min-height: 88rpx;
  border-radius: 18rpx;
  font-size: 28rpx;
  font-weight: 800;
  line-height: 88rpx;
}

.submit::after,
.skip::after {
  border: 0;
}

.icon-text-btn {
  padding: 0 24rpx;
}

.submit {
  margin-top: 36rpx;
  border: 1rpx solid #237a4b;
  background: linear-gradient(135deg, #237a4b, #145535);
  color: #ffffff;
  box-shadow: 0 16rpx 30rpx rgba(35, 122, 75, 0.18);
}

.skip {
  margin-top: 16rpx;
  border: 1rpx solid #e2e8dd;
  background: #ffffff;
  color: #445044;
}

.check-icon,
.back-icon {
  position: relative;
  display: inline-block;
  flex: 0 0 auto;
}

.check-icon {
  width: 32rpx;
  height: 20rpx;
  border-left: 5rpx solid #ffffff;
  border-bottom: 5rpx solid #ffffff;
  transform: rotate(-45deg) translateY(-4rpx);
}

.back-icon {
  width: 28rpx;
  height: 4rpx;
  border-radius: 99rpx;
  background: #445044;
}

.back-icon::before,
.back-icon::after {
  position: absolute;
  left: 0;
  width: 16rpx;
  height: 4rpx;
  border-radius: 99rpx;
  background: #445044;
  content: '';
}

.back-icon::before {
  top: -5rpx;
  transform: rotate(-42deg);
}

.back-icon::after {
  top: 5rpx;
  transform: rotate(42deg);
}
</style>
