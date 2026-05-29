<template>
  <view v-if="visible" class="overlay" @click.self="$emit('close')">
    <view class="sheet">
      <view class="sheet-head">
        <text class="sheet-title">{{ title }}</text>
        <button class="close" @click="$emit('close')">×</button>
      </view>

      <view class="notice">
        <text class="notice-icon">{{ type === 'bank' ? '付' : '证' }}</text>
        <text>{{ description }}</text>
      </view>

      <view class="card result-card">
        <view class="data-grid">
          <view v-for="item in fields" :key="item.label" class="kv">
            <text>{{ item.label }}</text>
            <text class="value">{{ item.value }}</text>
          </view>
        </view>
        <view class="photo-row">
          <view v-for="item in photos" :key="item" class="photo">{{ item }}</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  visible: boolean
  type: 'id' | 'bank'
}>()

defineEmits<{
  close: []
}>()

const title = computed(() => (props.type === 'bank' ? '模拟付款方式识别' : '模拟身份证识别'))
const description = computed(() =>
  props.type === 'bank'
    ? '选择银行卡时拍卡识别；选择支付宝或微信时填写收款人姓名和收款账号。'
    : '模拟身份证识别成功：姓名、身份证号、住址已自动带入录入表单，后续可接小程序相机。',
)
const fields = computed(() =>
  props.type === 'bank'
    ? [
        { label: '银行卡', value: '拍卡识别卡号与开户行' },
        { label: '支付宝/微信', value: '填写收款人和账号' },
        { label: '状态', value: '按付款方式录入' },
      ]
    : [
        { label: '姓名', value: '李建国' },
        { label: '身份证号', value: '410***********3215' },
        { label: '状态', value: '待接 OCR' },
      ],
)
const photos = computed(() => (props.type === 'bank' ? ['银行卡照片', '收款账号', 'TODO'] : ['身份证正面', '身份证反面', 'TODO']))
</script>

<style lang="scss" scoped>
.overlay {
  position: fixed;
  inset: 0;
  z-index: 30;
  display: flex;
  align-items: flex-end;
  background: rgba(15, 22, 16, 0.42);
}

.sheet {
  width: 100%;
  max-height: 86vh;
  padding: 28rpx 28rpx calc(36rpx + env(safe-area-inset-bottom));
  border-radius: 8rpx 8rpx 0 0;
  background: #ffffff;
  overflow: auto;
}

.sheet-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
  margin-bottom: 24rpx;
}

.sheet-title {
  color: #172018;
  font-size: 36rpx;
  font-weight: 760;
}

.close {
  width: 72rpx;
  height: 72rpx;
  border: 0;
  border-radius: 50%;
  background: #f2f4f0;
  color: #384338;
  font-size: 40rpx;
  line-height: 72rpx;
}

.notice {
  display: flex;
  gap: 20rpx;
  align-items: flex-start;
  padding: 24rpx;
  border: 1rpx solid #f4e1bd;
  border-radius: 8rpx;
  background: #fffaf0;
  color: #56614f;
  font-size: 24rpx;
  line-height: 1.5;
}

.notice-icon {
  display: flex;
  width: 54rpx;
  height: 54rpx;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 8rpx;
  background: #e8f5ec;
  color: #145535;
  font-weight: 800;
}

.result-card {
  margin-top: 20rpx;
  padding: 28rpx;
}

.data-grid,
.photo-row {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16rpx;
}

.kv,
.photo {
  min-width: 0;
  padding: 20rpx;
  border-radius: 8rpx;
  background: #f8faf6;
}

.kv text {
  display: block;
  color: #6d776c;
  font-size: 22rpx;
}

.kv .value {
  margin-top: 10rpx;
  color: #172018;
  font-size: 25rpx;
  font-weight: 700;
  overflow-wrap: anywhere;
}

.photo-row {
  margin-top: 20rpx;
}

.photo {
  display: flex;
  min-height: 120rpx;
  align-items: center;
  justify-content: center;
  border: 1rpx dashed rgba(35, 122, 75, 0.35);
  background: linear-gradient(135deg, rgba(35, 122, 75, 0.14), rgba(183, 121, 31, 0.16)), #eef2ea;
  color: #48604e;
  font-size: 24rpx;
  text-align: center;
}
</style>
