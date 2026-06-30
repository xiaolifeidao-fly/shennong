<template>
  <view class="card form-card">
    <view class="reuse-strip">
      <text class="icon farmer-icon"></text>
      <text>先选择农户。已建档农户会自动带出身份、电话、住址和收款信息；需要变更时可到农户详情修改。</text>
    </view>

    <view class="field">
      <text class="label">选择农户</text>
      <view class="farmer-search">
        <input
          v-model="farmerKeyword"
          class="input"
          confirm-type="search"
          placeholder="输入姓名、手机号、身份证号快速查找"
          @input="handleFarmerInput"
          @focus="showFarmerResults = true"
          @confirm="handleFarmerSearch"
        />
        <button class="new-farmer-btn" @click="selectFarmer('new')">新农户</button>
      </view>
      <view class="selected-farmer">
        <text>{{ selectedFarmerText }}</text>
      </view>
      <scroll-view v-if="showFarmerResults" class="farmer-results" scroll-y>
        <view v-if="farmerSearching" class="empty-result">正在搜索农户...</view>
        <template v-else>
          <button
            v-for="farmer in filteredFarmers"
            :key="farmer.id"
            class="farmer-option"
            :class="{ active: farmer.id === model.farmerId }"
            @click="selectFarmer(farmer.id)"
          >
            <view>
              <text class="farmer-name">{{ farmer.name }}</text>
              <text class="farmer-meta">{{ farmerMeta(farmer) }}</text>
            </view>
            <text class="farmer-status">{{ farmer.statusText || '已建档' }}</text>
          </button>
        </template>
        <view v-if="!farmerSearching && !filteredFarmers.length" class="empty-result">没有匹配农户，可点“新农户”建档</view>
      </scroll-view>
    </view>

    <view class="field">
      <view class="label-row">
        <text class="label required">农户身份</text>
        <view class="scan-group">
          <button class="scan-btn" @click="$emit('scan-id-front')">
            <text class="mini-icon camera-mini"></text>
            <text>拍身份证</text>
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
    <view v-if="idCardFrontUrl || idCardBackUrl" class="field">
      <text class="label">身份证照片</text>
      <view class="card-image-row">
        <view v-if="idCardFrontUrl" class="card-image-cell">
          <image :key="idCardFrontUrl" class="card-image" :src="idCardFrontUrl" mode="aspectFill" @click="previewCardImage(idCardFrontUrl)" />
          <text class="card-image-label">身份证正面</text>
        </view>
        <view v-if="idCardBackUrl" class="card-image-cell">
          <image :key="idCardBackUrl" class="card-image" :src="idCardBackUrl" mode="aspectFill" @click="previewCardImage(idCardBackUrl)" />
          <text class="card-image-label">身份证背面</text>
        </view>
      </view>
    </view>

    <view class="field">
      <text class="label">付款方式</text>
      <picker :value="payTypeIndex" :range="payTypeNames" @change="selectPayType">
        <view class="picker-value">{{ model.payType || '请选择付款方式（可不填）' }}</view>
      </picker>
    </view>

    <view class="field">
      <view class="label-row">
        <text class="label">收款人姓名</text>
        <view v-if="isBankPayment" class="scan-group">
          <picker :value="payCardTypeIndex" :range="payCardTypeNames" @change="selectPayCardType">
            <view class="card-type-picker">
              <text>{{ payCardTypeLabel }}</text>
              <text class="caret"></text>
            </view>
          </picker>
          <button class="scan-btn" @click="$emit('scan-bank', payCardType)">
            <text class="mini-icon card-mini"></text>
            <text>拍{{ payCardTypeLabel }}</text>
          </button>
        </view>
      </view>
      <input v-model="model.bankName" class="input" placeholder="请输入收款人姓名" />
    </view>
    <view class="field">
      <text class="label">{{ accountLabel }}</text>
      <input v-model="model.bankNumber" class="input" :placeholder="accountPlaceholder" />
    </view>
    <view v-if="bankCardUrl" class="field">
      <text class="label">银行卡照片</text>
      <view class="card-image-row single">
        <view class="card-image-cell">
          <image :key="bankCardUrl" class="card-image" :src="bankCardUrl" mode="aspectFill" @click="previewCardImage(bankCardUrl)" />
          <text class="card-image-label">银行卡</text>
        </view>
      </view>
    </view>
    <view class="field last">
      <text class="label required">农户电话</text>
      <input
        v-model="model.phone"
        class="input"
        type="number"
        maxlength="11"
        placeholder="请输入 11 位农户手机号"
        @input="handlePhoneInput"
      />
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { normalizeFileUrl } from '@/utils/fileUrl'
import { PAY_CARD_OCR_OPTIONS, type FarmerProfile, type GrainEntryDraft, type GrainPreset, type PayCardOcrType } from '@/types/grain'
import { maskIdNumber, maskPhone } from '@/utils/privacy'

const props = defineProps<{
  modelValue: GrainEntryDraft
  farmers: FarmerProfile[]
  farmerSearching?: boolean
  preset: GrainPreset
}>()

const emit = defineEmits<{
  'update:modelValue': [value: GrainEntryDraft]
  'farmer-change': [farmerId: string]
  'farmer-search': [keyword: string]
  'scan-id-front': []
  'scan-bank': [cardType: PayCardOcrType]
}>()

const model = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const payCardType = ref<PayCardOcrType>('bank-card')
const payCardTypeNames = PAY_CARD_OCR_OPTIONS.map((item) => item.label)
const payCardTypeIndex = computed(() => Math.max(0, PAY_CARD_OCR_OPTIONS.findIndex((item) => item.value === payCardType.value)))
const payCardTypeLabel = computed(() => PAY_CARD_OCR_OPTIONS[payCardTypeIndex.value]?.label || '银行卡')

function selectPayCardType(event: { detail: { value: number | string } }) {
  payCardType.value = PAY_CARD_OCR_OPTIONS[Number(event.detail.value)]?.value || 'bank-card'
}

const farmerKeyword = ref('')
const showFarmerResults = ref(false)
const selectedFarmer = computed(() => props.farmers.find((farmer) => farmer.id === props.modelValue.farmerId))
const selectedFarmerText = computed(() => {
  if (selectedFarmer.value) {
    const meta = farmerMeta(selectedFarmer.value)
    return `当前：${selectedFarmer.value.name}${meta ? ` · ${meta}` : ' · 已建档'}`
  }
  return '当前：新农户，拍身份证建档'
})
const filteredFarmers = computed(() => {
  const key = farmerKeyword.value.trim()
  const source = key
    ? props.farmers.filter((farmer) =>
        [farmer.name, farmer.phone, farmer.idNumber, farmer.address, farmer.bankName, farmer.bankNumber, farmer.statusText].some((value) => value?.includes(key)),
      )
    : props.farmers
  return source.slice(0, 8)
})
const payTypeNames = computed(() => props.preset.paymentMethods.length ? props.preset.paymentMethods.map((item) => item.methodName) : props.preset.payTypes)
const payTypeIndex = computed(() => Math.max(0, payTypeNames.value.indexOf(model.value.payType)))
const selectedPaymentMethod = computed(() =>
  props.preset.paymentMethods.find((item) => item.id === model.value.paymentMethodId) ||
  props.preset.paymentMethods.find((item) => item.methodName === model.value.payType),
)
const idCardFrontUrl = computed(() => normalizeFileUrl(model.value.cardImages?.idCardFront?.displayUrl || model.value.cardImages?.idCardFront?.ossUrl))
const idCardBackUrl = computed(() => normalizeFileUrl(model.value.cardImages?.idCardBack?.displayUrl || model.value.cardImages?.idCardBack?.ossUrl))
const bankCardUrl = computed(() => normalizeFileUrl(model.value.cardImages?.bankCard?.displayUrl || model.value.cardImages?.bankCard?.ossUrl))
const paymentMethodCode = computed(() => model.value.paymentMethodCode || selectedPaymentMethod.value?.methodCode || '')
const isBankPayment = computed(() => paymentMethodCode.value === 'Bank')
const accountLabel = computed(() => (paymentMethodCode.value === 'Bank' ? '银行卡号' : '收款账号'))
const accountPlaceholder = computed(() => {
  if (paymentMethodCode.value === 'Bank') {
    return '请输入银行卡号'
  }
  if (paymentMethodCode.value === 'Alipay') {
    return '请输入支付宝账号'
  }
  if (paymentMethodCode.value === 'WECHAT') {
    return '请输入微信收款账号'
  }
  return '请输入收款账号'
})

watch(
  () => props.modelValue.farmerId,
  () => {
    farmerKeyword.value = selectedFarmer.value?.name || ''
    showFarmerResults.value = false
  },
  { immediate: true },
)

function selectFarmer(farmerId: string) {
  const farmer = props.farmers.find((item) => item.id === farmerId)
  farmerKeyword.value = farmer?.name || ''
  showFarmerResults.value = false
  emit('farmer-change', farmerId)
}

function handleFarmerInput() {
  showFarmerResults.value = true
}

function handleFarmerSearch() {
  showFarmerResults.value = true
  emit('farmer-search', farmerKeyword.value)
}

function farmerMeta(farmer: FarmerProfile) {
  const items = [maskPhone(farmer.phone), maskIdNumber(farmer.idNumber)].filter(Boolean)
  return items.length ? items.join(' · ') : '已建档'
}

function selectPayType(event: { detail: { value: number | string } }) {
  const option = props.preset.paymentMethods[Number(event.detail.value)]
  model.value.paymentMethodId = option?.id || 0
  model.value.paymentMethodCode = option?.methodCode || ''
  model.value.payType = option?.methodName || payTypeNames.value[Number(event.detail.value)] || model.value.payType
}

function handlePhoneInput(event: unknown) {
  const raw =
    typeof event === 'object' && event && 'detail' in event
      ? String((event as { detail?: { value?: string } }).detail?.value ?? model.value.phone ?? '')
      : String(model.value.phone ?? '')
  const digits = raw.replace(/\D/g, '').slice(0, 11)
  model.value.phone = digits
}

function previewCardImage(current: string) {
  const urls = [idCardFrontUrl.value, idCardBackUrl.value, bankCardUrl.value].filter(Boolean)
  uni.previewImage({ current, urls })
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

.card-image-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16rpx;
}

.card-image-row.single {
  grid-template-columns: minmax(0, 1fr);
  max-width: 320rpx;
}

.card-image-cell {
  position: relative;
  overflow: hidden;
  min-height: 176rpx;
  border: 1rpx solid #e2e8dd;
  border-radius: 18rpx;
  background: #f7faf5;
}

.card-image {
  display: block;
  width: 100%;
  height: 176rpx;
}

.card-image-label {
  position: absolute;
  left: 12rpx;
  bottom: 12rpx;
  padding: 6rpx 12rpx;
  border-radius: 999rpx;
  background: rgba(23, 32, 24, 0.72);
  color: #ffffff;
  font-size: 22rpx;
  line-height: 1.2;
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
  align-items: center;
  gap: 12rpx;
}

.card-type-picker {
  display: flex;
  align-items: center;
  gap: 8rpx;
  height: 56rpx;
  padding: 0 18rpx;
  border: 1rpx solid #cfe0d1;
  border-radius: 999rpx;
  background: #ffffff;
  color: #2d4633;
  font-size: 24rpx;
  font-weight: 760;
  line-height: 56rpx;
}

.card-type-picker .caret {
  width: 0;
  height: 0;
  border-left: 7rpx solid transparent;
  border-right: 7rpx solid transparent;
  border-top: 9rpx solid #6d776c;
}

.farmer-search {
  display: grid;
  grid-template-columns: 1fr 148rpx;
  gap: 14rpx;
  align-items: center;
}

.new-farmer-btn {
  height: 88rpx;
  border: 1rpx solid #cfe0d1;
  border-radius: 18rpx;
  background: #ffffff;
  color: #145535;
  font-size: 26rpx;
  font-weight: 800;
  line-height: 88rpx;
}

.selected-farmer {
  margin-top: 12rpx;
  color: #667266;
  font-size: 24rpx;
  line-height: 1.45;
}

.farmer-results {
  max-height: 460rpx;
  margin-top: 14rpx;
  border: 1rpx solid #e2e8dd;
  border-radius: 18rpx;
  background: #ffffff;
  overflow: hidden;
}

.farmer-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
  width: 100%;
  padding: 20rpx 22rpx;
  border: 0;
  border-bottom: 1rpx solid #edf1ea;
  border-radius: 0;
  background: #ffffff;
  color: #172018;
  line-height: 1.35;
  text-align: left;
}

.farmer-option.active {
  background: #f0f8f1;
}

.farmer-name,
.farmer-meta {
  display: block;
}

.farmer-name {
  font-size: 28rpx;
  font-weight: 800;
}

.farmer-meta,
.farmer-status {
  color: #6d776c;
  font-size: 23rpx;
}

.farmer-status {
  flex: 0 0 auto;
}

.empty-result {
  padding: 28rpx 22rpx;
  color: #667266;
  font-size: 24rpx;
  text-align: center;
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
