<template>
  <view v-if="visible && farmer" class="overlay" @click.self="$emit('close')">
    <view class="sheet">
      <view class="sheet-head">
        <text class="sheet-title">{{ editing ? '编辑农户资料' : '农户当天汇总与明细' }}</text>
        <button class="close" @click="$emit('close')">×</button>
      </view>

      <view v-if="editing" class="card form-card">
        <view class="field">
          <text class="label">农户姓名</text>
          <input v-model="editForm.name" class="input" />
        </view>
        <view class="field">
          <text class="label">身份证号</text>
          <input v-model="editForm.idNumber" class="input" />
        </view>
        <view class="field">
          <text class="label">身份证住址</text>
          <textarea v-model="editForm.address" class="textarea" />
        </view>
        <view class="field">
          <text class="label">农户电话</text>
          <input v-model="editForm.phone" class="input" />
        </view>
        <view class="field">
          <text class="label">银行卡号</text>
          <input v-model="editForm.bankNumber" class="input" />
        </view>
        <view class="field">
          <text class="label">开户行</text>
          <input v-model="editForm.bankName" class="input" />
        </view>
        <button class="primary-btn" @click="saveFarmer">保存农户资料</button>
      </view>

      <template v-else>
        <view class="card farmer-card">
          <view class="card-head">
            <view>
              <text class="card-title">{{ farmer.name }} · 今日 {{ entries.length }} 笔</text>
              <text class="card-meta">{{ farmer.statusText }}</text>
            </view>
            <text class="badge">{{ totalQuantity.toLocaleString('zh-CN') }} 斤</text>
          </view>
          <view class="data-grid">
            <view class="kv"><text>身份证号</text><text class="value">{{ farmer.idNumber }}</text></view>
            <view class="kv"><text>电话</text><text class="value">{{ farmer.phone }}</text></view>
            <view class="kv"><text>金额合计</text><text class="value">{{ formatAmount(totalAmount) }}</text></view>
          </view>
          <view class="kv address"><text>住址</text><text class="value">{{ farmer.address }}</text></view>
          <view class="data-grid">
            <view class="kv"><text>银行卡</text><text class="value">{{ farmer.bankNumber }}</text></view>
            <view class="kv"><text>开户行</text><text class="value">{{ farmer.bankName }}</text></view>
            <view class="kv"><text>归属</text><text class="value">我的录入</text></view>
          </view>
          <view class="photo-row">
            <view class="photo">身份证照片</view>
            <view class="photo">银行卡照片</view>
            <view class="photo">现场凭证</view>
          </view>
          <view class="inline-actions">
            <button class="secondary-btn" @click="startEdit">编辑农户资料</button>
            <button class="primary-btn" @click="$emit('entry', farmer.id)">给该农户录粮</button>
          </view>
        </view>

        <SectionHeader title="多次录入明细" action-text="再录一笔" @action="$emit('entry', farmer.id)" />
        <view class="list">
          <GrainEntryCard v-for="entry in entries" :key="entry.id" :entry="entry" :farmer="farmer" />
        </view>
      </template>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import SectionHeader from '@/components/business/SectionHeader.vue'
import GrainEntryCard from '@/components/business/GrainEntryCard.vue'
import { formatAmount } from '@/utils/grain'
import type { FarmerProfile, GrainEntry } from '@/types/grain'

const props = defineProps<{
  visible: boolean
  farmer?: FarmerProfile
  entries: GrainEntry[]
  editing: boolean
}>()

const emit = defineEmits<{
  close: []
  edit: [value: boolean]
  save: [farmer: FarmerProfile]
  entry: [farmerId: string]
}>()

const editForm = reactive<FarmerProfile>({
  id: '',
  name: '',
  idNumber: '',
  phone: '',
  address: '',
  bankNumber: '',
  bankName: '',
  status: 'complete',
  statusText: '',
})

const totalQuantity = computed(() => props.entries.reduce((sum, item) => sum + item.quantity, 0))
const totalAmount = computed(() => props.entries.reduce((sum, item) => sum + item.amount, 0))

watch(
  () => props.farmer,
  (farmer) => {
    if (farmer) {
      Object.assign(editForm, farmer)
    }
  },
  { immediate: true },
)

function startEdit() {
  emit('edit', true)
}

function saveFarmer() {
  emit('save', { ...editForm })
}
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

.sheet-head,
.card-head,
.inline-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
}

.sheet-head {
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

.farmer-card,
.form-card {
  padding: 28rpx;
}

.card-title,
.card-meta {
  display: block;
}

.card-title {
  color: #172018;
  font-size: 30rpx;
  font-weight: 760;
}

.card-meta {
  margin-top: 10rpx;
  color: #6d776c;
  font-size: 24rpx;
}

.badge {
  padding: 10rpx 16rpx;
  border-radius: 999rpx;
  background: #e8f5ec;
  color: #145535;
  font-size: 22rpx;
  font-weight: 760;
}

.data-grid,
.photo-row {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16rpx;
  margin-top: 24rpx;
}

.kv,
.photo {
  min-width: 0;
  padding: 20rpx;
  border-radius: 8rpx;
  background: #f8faf6;
}

.address {
  margin-top: 16rpx;
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

.inline-actions {
  margin-top: 24rpx;
}

.primary-btn,
.secondary-btn {
  flex: 1;
  min-height: 84rpx;
  border-radius: 8rpx;
  font-size: 28rpx;
  font-weight: 800;
  line-height: 84rpx;
}

.primary-btn {
  border: 1rpx solid #237a4b;
  background: #237a4b;
  color: #ffffff;
}

.secondary-btn {
  border: 1rpx solid #e2e8dd;
  background: #ffffff;
  color: #243027;
}

.list {
  display: grid;
  gap: 20rpx;
}

.field {
  margin-bottom: 24rpx;
}

.label {
  display: block;
  margin-bottom: 14rpx;
  color: #445044;
  font-size: 26rpx;
  font-weight: 700;
}

.input,
.textarea {
  width: 100%;
  border: 1rpx solid #e2e8dd;
  border-radius: 8rpx;
  background: #fbfcfa;
  color: #172018;
  font-size: 28rpx;
}

.input {
  height: 88rpx;
  padding: 0 24rpx;
}

.textarea {
  min-height: 148rpx;
  padding: 22rpx 24rpx;
}
</style>
