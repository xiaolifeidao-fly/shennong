<template>
  <view v-if="visible && farmer" class="overlay" @click.self="$emit('close')">
    <view class="sheet">
      <view class="sheet-head">
        <text class="sheet-title">{{ editing ? '编辑农户资料' : '农户当天汇总与明细' }}</text>
        <button class="close" @click="$emit('close')">
          <text class="close-icon"></text>
        </button>
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
          <text class="label">收款账号</text>
          <input v-model="editForm.bankNumber" class="input" />
        </view>
        <view class="field">
          <text class="label">收款人/开户行</text>
          <input v-model="editForm.bankName" class="input" />
        </view>
        <button class="primary-btn" @click="saveFarmer">
          <text class="check-mini"></text>
          <text>保存农户资料</text>
        </button>
      </view>

      <template v-else>
        <view class="card farmer-card">
          <view class="card-head">
            <view>
              <text class="card-title">{{ farmer.name }} · 今日 {{ entries.length }} 笔</text>
              <text class="card-meta">{{ farmer.statusText }}</text>
            </view>
            <text class="badge">{{ totalQuantity.toLocaleString('zh-CN') }} 公斤</text>
          </view>
          <view class="data-grid">
            <view class="kv"><text>身份证号</text><text class="value">{{ farmer.idNumber }}</text></view>
            <view class="kv"><text>电话</text><text class="value">{{ farmer.phone }}</text></view>
            <view class="kv"><text>金额合计</text><text class="value">{{ formatAmount(totalAmount) }}</text></view>
          </view>
          <view class="kv address"><text>住址</text><text class="value">{{ farmer.address }}</text></view>
          <view class="data-grid">
            <view class="kv"><text>收款账号</text><text class="value">{{ farmer.bankNumber }}</text></view>
            <view class="kv"><text>收款人/开户行</text><text class="value">{{ farmer.bankName }}</text></view>
            <view class="kv"><text>归属</text><text class="value">我的录入</text></view>
          </view>
          <view class="photo-row">
            <view class="photo">身份证照片</view>
            <view class="photo">收款凭证</view>
            <view class="photo">现场凭证</view>
          </view>
          <view class="inline-actions">
            <button class="secondary-btn" @click="startEdit">
              <text class="edit-mini"></text>
              <text>编辑资料</text>
            </button>
            <button class="primary-btn" @click="$emit('entry', farmer.id)">
              <text class="plus-mini"></text>
              <text>录粮</text>
            </button>
          </view>
        </view>

        <SectionHeader title="农户汇总明细" action-text="再录一笔" @action="$emit('entry', farmer.id)" />
        <view class="summary-table">
          <view class="summary-row summary-head">
            <text>收购类型</text>
            <text>日期</text>
            <text>付款</text>
            <text>总公斤</text>
            <text>总金额</text>
          </view>
          <view v-for="summary in summaries" :key="summary.id" class="summary-row">
            <text>{{ summary.crop || '-' }}</text>
            <text>{{ summary.summaryDate || '-' }}</text>
            <text>{{ summary.payType || '-' }}</text>
            <text>{{ summary.totalQuantity.toLocaleString('zh-CN') }}</text>
            <text>{{ formatAmount(summary.totalAmount) }}</text>
          </view>
          <view v-if="summariesLoading" class="empty-summary">正在加载汇总...</view>
          <view v-else-if="!summaries.length" class="empty-summary">暂无汇总数据</view>
        </view>

        <SectionHeader title="多次录入明细" action-text="再录一笔" @action="$emit('entry', farmer.id)" />
        <view v-if="entriesLoading" class="detail-loading">正在加载录入明细...</view>
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
import type { FarmerProfile, GrainEntry, GrainFarmerPurchaseSummary } from '@/types/grain'

const props = defineProps<{
  visible: boolean
  farmer?: FarmerProfile
  entries: GrainEntry[]
  summaries: GrainFarmerPurchaseSummary[]
  entriesLoading?: boolean
  summariesLoading?: boolean
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
  stationId: 0,
  appUserId: 0,
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
  animation: fade-in 0.2s ease both;
}

.sheet {
  width: 100%;
  max-height: 86vh;
  padding: 28rpx 28rpx calc(36rpx + env(safe-area-inset-bottom));
  border-radius: 32rpx 32rpx 0 0;
  background: #f8fbf6;
  overflow: auto;
  animation: sheet-up 0.28s ease both;
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
  display: flex;
  width: 72rpx;
  height: 72rpx;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 50%;
  background: #ffffff;
  color: #384338;
  box-shadow: 0 10rpx 24rpx rgba(31, 47, 31, 0.08);
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

.summary-table {
  overflow: hidden;
  border: 1rpx solid #e2e8dd;
  border-radius: 8rpx;
  background: #ffffff;
}

.summary-row {
  display: grid;
  grid-template-columns: 1.1fr 1.25fr 1fr 1fr 1.15fr;
  gap: 10rpx;
  padding: 18rpx 16rpx;
  border-top: 1rpx solid #eef2ea;
}

.summary-row:first-child {
  border-top: 0;
}

.summary-row text {
  min-width: 0;
  color: #384338;
  font-size: 22rpx;
  line-height: 1.35;
  overflow-wrap: anywhere;
}

.summary-head {
  background: #f6f8f4;
}

.summary-head text {
  color: #667266;
  font-weight: 760;
}

.empty-summary {
  padding: 28rpx;
  color: #7b857b;
  font-size: 24rpx;
  text-align: center;
}

.detail-loading {
  padding: 18rpx 22rpx;
  border-radius: 8rpx;
  background: #f6f8f4;
  color: #667266;
  font-size: 24rpx;
}

.kv,
.photo {
  min-width: 0;
  padding: 20rpx;
  border-radius: 18rpx;
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
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10rpx;
  flex: 1;
  min-height: 84rpx;
  border-radius: 18rpx;
  font-size: 28rpx;
  font-weight: 800;
  line-height: 84rpx;
}

.primary-btn {
  border: 1rpx solid #237a4b;
  background: linear-gradient(135deg, #237a4b, #145535);
  color: #ffffff;
  box-shadow: 0 14rpx 28rpx rgba(35, 122, 75, 0.18);
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
  border-radius: 18rpx;
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

.close-icon,
.check-mini,
.edit-mini,
.plus-mini {
  position: relative;
  display: inline-block;
  flex: 0 0 auto;
}

.close-icon {
  width: 30rpx;
  height: 30rpx;
}

.close-icon::before,
.close-icon::after {
  position: absolute;
  left: 13rpx;
  top: 1rpx;
  width: 4rpx;
  height: 28rpx;
  border-radius: 999rpx;
  background: #384338;
  content: '';
}

.close-icon::before {
  transform: rotate(45deg);
}

.close-icon::after {
  transform: rotate(-45deg);
}

.check-mini {
  width: 28rpx;
  height: 18rpx;
  border-left: 5rpx solid currentColor;
  border-bottom: 5rpx solid currentColor;
  transform: rotate(-45deg) translateY(-3rpx);
}

.edit-mini {
  width: 26rpx;
  height: 8rpx;
  border-radius: 999rpx;
  background: currentColor;
  transform: rotate(-35deg);
}

.edit-mini::after {
  position: absolute;
  right: -5rpx;
  top: -3rpx;
  width: 8rpx;
  height: 14rpx;
  border-radius: 4rpx;
  background: #ffb84d;
  content: '';
}

.plus-mini {
  width: 26rpx;
  height: 26rpx;
}

.plus-mini::before,
.plus-mini::after {
  position: absolute;
  left: 11rpx;
  top: 3rpx;
  width: 4rpx;
  height: 20rpx;
  border-radius: 999rpx;
  background: currentColor;
  content: '';
}

.plus-mini::after {
  transform: rotate(90deg);
}

@keyframes fade-in {
  from {
    opacity: 0;
  }

  to {
    opacity: 1;
  }
}

@keyframes sheet-up {
  from {
    transform: translateY(80rpx);
  }

  to {
    transform: translateY(0);
  }
}
</style>
