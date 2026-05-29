<template>
  <view class="card form-card">
    <view class="reuse-strip">
      <text class="icon farmer-icon"></text>
      <text>先选择农户。已建档农户会自动带出身份、电话、住址和收款信息；需要变更时可到农户详情修改。</text>
    </view>

    <view class="field">
      <text class="label">选择农户</text>
      <picker :value="farmerIndex" :range="farmerOptions" range-key="name" @change="handleFarmerChange">
        <view class="picker-value">{{ farmerOptions[farmerIndex]?.name }}</view>
      </picker>
    </view>

    <view class="field">
      <view class="label-row">
        <text class="label required">农户身份</text>
        <view class="scan-group">
          <button class="scan-btn" @click="$emit('scan-id-front')">
            <text class="mini-icon camera-mini"></text>
            <text>拍正面</text>
          </button>
          <button class="scan-btn" @click="$emit('scan-id-back')">
            <text class="mini-icon camera-mini"></text>
            <text>拍背面</text>
          </button>
        </view>
      </view>
      <input v-model="model.farmerName" class="input" placeholder="农户姓名" />
    </view>
    <view class="field">
      <input v-model="model.idNumber" class="input" placeholder="身份证号" />
    </view>
    <view class="field">
      <textarea v-model="model.address" class="textarea" placeholder="身份证住址" />
    </view>

    <view class="field">
      <text class="label required">付款方式</text>
      <picker :value="payTypeIndex" :range="payTypeNames" @change="selectPayType">
        <view class="picker-value">{{ model.payType || '请选择付款方式' }}</view>
      </picker>
    </view>

    <view v-if="isBankPayment" class="field">
      <view class="label-row">
        <text class="label required">银行卡信息</text>
        <button class="scan-btn" @click="$emit('scan-bank')">
          <text class="mini-icon card-mini"></text>
          <text>拍银行卡</text>
        </button>
      </view>
      <input v-model="model.bankNumber" class="input" placeholder="银行卡号" />
    </view>
    <view v-if="isBankPayment" class="field">
      <input v-model="model.bankName" class="input" placeholder="开户行" />
    </view>
    <view v-else-if="isAccountPayment" class="field">
      <text class="label required">收款人姓名</text>
      <input v-model="model.bankName" class="input" placeholder="请输入收款人姓名" />
    </view>
    <view v-if="isAccountPayment" class="field">
      <text class="label required">收款账号</text>
      <input v-model="model.bankNumber" class="input" :placeholder="accountPlaceholder" />
    </view>
    <view class="field last">
      <text class="label">农户电话</text>
      <input v-model="model.phone" class="input" placeholder="手动输入农户电话" />
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { FarmerProfile, GrainEntryDraft, GrainPreset } from '@/types/grain'

const props = defineProps<{
  modelValue: GrainEntryDraft
  farmers: FarmerProfile[]
  preset: GrainPreset
}>()

const emit = defineEmits<{
  'update:modelValue': [value: GrainEntryDraft]
  'farmer-change': [farmerId: string]
  'scan-id-front': []
  'scan-id-back': []
  'scan-bank': []
}>()

const model = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const farmerOptions = computed(() => [
  ...props.farmers.map((farmer) => ({ id: farmer.id, name: `${farmer.name} · 已建档` })),
  { id: 'new', name: '新农户，拍身份证建档' },
])
const farmerIndex = computed(() => Math.max(0, farmerOptions.value.findIndex((item) => item.id === props.modelValue.farmerId)))
const payTypeNames = computed(() => props.preset.paymentMethods.length ? props.preset.paymentMethods.map((item) => item.methodName) : props.preset.payTypes)
const payTypeIndex = computed(() => Math.max(0, payTypeNames.value.indexOf(model.value.payType)))
const selectedPaymentMethod = computed(() =>
  props.preset.paymentMethods.find((item) => item.id === model.value.paymentMethodId) ||
  props.preset.paymentMethods.find((item) => item.methodName === model.value.payType),
)
const paymentMethodCode = computed(() => model.value.paymentMethodCode || selectedPaymentMethod.value?.methodCode || '')
const isBankPayment = computed(() => paymentMethodCode.value === 'Bank')
const isAccountPayment = computed(() => paymentMethodCode.value === 'Alipay' || paymentMethodCode.value === 'WECHAT')
const accountPlaceholder = computed(() => {
  if (paymentMethodCode.value === 'Alipay') {
    return '请输入支付宝账号'
  }
  if (paymentMethodCode.value === 'WECHAT') {
    return '请输入微信收款账号'
  }
  return '请输入收款账号'
})

function handleFarmerChange(event: { detail: { value: number | string } }) {
  const index = Number(event.detail.value)
  const farmerId = farmerOptions.value[index]?.id || 'new'
  emit('farmer-change', farmerId)
}

function selectPayType(event: { detail: { value: number | string } }) {
  const option = props.preset.paymentMethods[Number(event.detail.value)]
  model.value.paymentMethodId = option?.id || 0
  model.value.paymentMethodCode = option?.methodCode || ''
  model.value.payType = option?.methodName || payTypeNames.value[Number(event.detail.value)] || model.value.payType
}
</script>

<style lang="scss" scoped>
.form-card {
  padding: 30rpx;
}

.reuse-strip {
  display: flex;
  gap: 20rpx;
  align-items: flex-start;
  padding: 24rpx;
  margin-bottom: 20rpx;
  border: 1rpx solid rgba(35, 122, 75, 0.14);
  border-radius: 22rpx;
  background: linear-gradient(135deg, #e8f5ec, #f8fbf4);
  color: #35513b;
  font-size: 24rpx;
  line-height: 1.5;
}

.icon {
  position: relative;
  display: flex;
  width: 54rpx;
  height: 54rpx;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 18rpx;
  background: #ffffff;
  box-shadow: 0 10rpx 22rpx rgba(35, 122, 75, 0.08);
}

.farmer-icon::before {
  position: absolute;
  top: 12rpx;
  width: 18rpx;
  height: 18rpx;
  border-radius: 50%;
  background: #237a4b;
  content: '';
}

.farmer-icon::after {
  position: absolute;
  bottom: 11rpx;
  width: 30rpx;
  height: 16rpx;
  border-radius: 18rpx 18rpx 8rpx 8rpx;
  background: #ffb84d;
  content: '';
}

.field {
  margin-bottom: 24rpx;
}

.field.last {
  margin-bottom: 0;
}

.label,
.label-row {
  margin-bottom: 14rpx;
  color: #445044;
  font-size: 26rpx;
  font-weight: 700;
}

.label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
}

.required::after {
  content: ' *';
  color: #d14343;
}

.input,
.textarea,
.picker-value {
  width: 100%;
  border: 1rpx solid #e2e8dd;
  border-radius: 18rpx;
  background: #fbfcfa;
  color: #172018;
  font-size: 28rpx;
}

.input,
.picker-value {
  height: 88rpx;
  padding: 0 24rpx;
  line-height: 88rpx;
}

.textarea {
  min-height: 148rpx;
  padding: 22rpx 24rpx;
  line-height: 1.45;
}

.scan-group {
  display: flex;
  gap: 12rpx;
}

.scan-btn {
  display: flex;
  align-items: center;
  gap: 8rpx;
  padding: 12rpx 18rpx;
  border: 0;
  border-radius: 999rpx;
  background: #edf7ff;
  color: #1d5d99;
  font-size: 24rpx;
  font-weight: 760;
  line-height: 1.2;
}

.mini-icon {
  position: relative;
  display: inline-block;
  width: 24rpx;
  height: 24rpx;
}

.camera-mini::before {
  position: absolute;
  left: 2rpx;
  top: 6rpx;
  width: 18rpx;
  height: 14rpx;
  border: 3rpx solid #1d5d99;
  border-radius: 5rpx;
  content: '';
}

.camera-mini::after {
  position: absolute;
  left: 9rpx;
  top: 10rpx;
  width: 5rpx;
  height: 5rpx;
  border: 3rpx solid #1d5d99;
  border-radius: 50%;
  content: '';
}

.card-mini::before {
  position: absolute;
  left: 1rpx;
  top: 5rpx;
  width: 20rpx;
  height: 14rpx;
  border: 3rpx solid #1d5d99;
  border-radius: 5rpx;
  content: '';
}

.card-mini::after {
  position: absolute;
  left: 6rpx;
  top: 14rpx;
  width: 12rpx;
  height: 3rpx;
  border-radius: 999rpx;
  background: #1d5d99;
  content: '';
}
</style>
