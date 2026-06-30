<template>
  <view class="agreement-page">
    <view v-if="documents.length" class="tabs">
      <view
        v-for="doc in documents"
        :key="doc.key"
        class="tab"
        :class="{ active: doc.key === activeKey }"
        @click="activeKey = doc.key"
      >
        {{ doc.title }}
      </view>
    </view>

    <scroll-view scroll-y class="content-scroll">
      <text class="content-text">{{ activeContent }}</text>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getAgreement } from '@/services/agreement'
import type { AgreementDocument } from '@/types/api'

const documents = ref<AgreementDocument[]>([])
const activeKey = ref('')

const activeContent = computed(
  () => documents.value.find((doc) => doc.key === activeKey.value)?.content || '',
)

onLoad(async (query) => {
  try {
    const result = await getAgreement()
    documents.value = result.documents || []
    const requestedKey = (query?.key as string) || ''
    activeKey.value
      = documents.value.find((doc) => doc.key === requestedKey)?.key
      || documents.value[0]?.key
      || ''
  } catch (error) {
    uni.showToast({
      title: error instanceof Error ? error.message : '协议加载失败',
      icon: 'none',
    })
  }
})
</script>

<style lang="scss" scoped>
.agreement-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  padding: 0;
  background: #f5f7f3;
}

.tabs {
  display: flex;
  flex-shrink: 0;
  gap: 12rpx;
  padding: 20rpx 24rpx;
  overflow-x: auto;
  background: #ffffff;
  box-shadow: 0 4rpx 12rpx rgba(31, 47, 31, 0.04);
}

.tab {
  flex-shrink: 0;
  padding: 12rpx 24rpx;
  border-radius: 999rpx;
  background: #eef4ed;
  color: #5a6a5e;
  font-size: 26rpx;
}

.tab.active {
  background: #237a4b;
  color: #ffffff;
  font-weight: 700;
}

.content-scroll {
  flex: 1;
  padding: 24rpx 28rpx calc(40rpx + env(safe-area-inset-bottom));
}

.content-text {
  display: block;
  color: #2b352d;
  font-size: 27rpx;
  line-height: 1.8;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
