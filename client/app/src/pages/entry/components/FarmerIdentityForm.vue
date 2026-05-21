<template>
  <view class="card form-card">
    <view class="reuse-strip">
      <text class="icon">户</text>
      <text>先选择农户。已建档农户会自动带出身份、电话、住址和银行卡信息；需要变更时可到农户详情修改。</text>
    </view>

    <view class="field">
      <text class="label">选择农户</text>
      <picker :value="farmerIndex" :range="farmerOptions" range-key="name" @change="handleFarmerChange">
        <view class="picker-value">{{ farmerOptions[farmerIndex]?.name }}</view>
      </picker>
    </view>

    <view class="field">
      <view class="label-row">
        <text class="label">农户身份</text>
        <button class="scan-btn" @click="$emit('scan-id')">拍身份证识别</button>
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
      <view class="label-row">
        <text class="label">银行卡信息</text>
        <button class="scan-btn" @click="$emit('scan-bank')">拍银行卡识别</button>
      </view>
      <input v-model="model.bankNumber" class="input" placeholder="银行卡号" />
    </view>
    <view class="field">
      <input v-model="model.bankName" class="input" placeholder="开户行" />
    </view>
    <view class="field last">
      <text class="label">农户电话</text>
      <input v-model="model.phone" class="input" placeholder="手动输入农户电话" />
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { FarmerProfile, GrainEntryDraft } from '@/types/grain'

const props = defineProps<{
  modelValue: GrainEntryDraft
  farmers: FarmerProfile[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: GrainEntryDraft]
  'farmer-change': [farmerId: string]
  'scan-id': []
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

function handleFarmerChange(event: { detail: { value: number | string } }) {
  const index = Number(event.detail.value)
  const farmerId = farmerOptions.value[index]?.id || 'new'
  emit('farmer-change', farmerId)
}
</script>

<style lang="scss" scoped>
.form-card {
  padding: 28rpx;
}

.reuse-strip {
  display: flex;
  gap: 20rpx;
  align-items: flex-start;
  padding: 24rpx;
  margin-bottom: 20rpx;
  border: 1rpx solid rgba(35, 122, 75, 0.18);
  border-radius: 8rpx;
  background: #e8f5ec;
  color: #35513b;
  font-size: 24rpx;
  line-height: 1.5;
}

.icon {
  display: flex;
  width: 54rpx;
  height: 54rpx;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 8rpx;
  background: #ffffff;
  color: #145535;
  font-weight: 800;
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

.input,
.textarea,
.picker-value {
  width: 100%;
  border: 1rpx solid #e2e8dd;
  border-radius: 8rpx;
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

.scan-btn {
  padding: 12rpx 18rpx;
  border: 0;
  border-radius: 999rpx;
  background: #eaf2fb;
  color: #2563a8;
  font-size: 24rpx;
  font-weight: 760;
  line-height: 1.2;
}
</style>
