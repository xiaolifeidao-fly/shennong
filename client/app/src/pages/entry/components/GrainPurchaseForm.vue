<template>
  <view class="card form-card">
    <view class="field">
      <text class="label required">收购粮食类型</text>
      <input
        v-model="cropKeyword"
        class="input"
        placeholder="输入关键词搜索，请从下方选择"
        @input="handleCropInput"
        @focus="showCropResults = true"
      />
      <view class="crop-helper">仅可选择粮站已维护的粮食类型，输入内容不会直接保存。</view>
      <view class="crop-options">
        <button
          v-for="crop in filteredCropOptions"
          :key="crop.id"
          class="crop-chip"
          :class="{ active: crop.id === model.purchaseTypeId }"
          @click="selectCrop(crop)"
        >
          {{ crop.name }}
        </button>
        <view v-if="!filteredCropOptions.length && cropKeyword.trim()" class="crop-empty">未找到匹配粮食类型，请从已有类型中选择或联系管理员维护。</view>
        <view v-else-if="!props.preset.purchaseTypes.length" class="crop-empty">当前粮站暂无可选粮食类型，请先联系管理员维护。</view>
      </view>
    </view>

    <view class="field">
      <text class="label required">购进重量</text>
      <view class="split">
        <input v-model="model.quantity" class="input" type="digit" placeholder="请输入重量" />
        <picker :value="unitIndex" :range="unitOptions" @change="selectUnit">
          <view class="unit-value">{{ model.unit || '公斤' }}</view>
        </picker>
      </view>
    </view>

    <view class="field">
      <text class="label required">购进货物金额</text>
      <input v-model="model.amount" class="input" type="digit" placeholder="请输入金额" />
    </view>

    <view class="field">
      <text class="label">自动计算单价</text>
      <input class="input" :value="priceText" disabled />
    </view>

    <view class="field">
      <text class="label">收购时间</text>
      <view class="datetime-row">
        <picker mode="date" :value="datePart(model.buyTime) || todayDate" @change="setDateTimePart('buyTime', 'date', $event)">
          <view class="datetime-value">{{ datePart(model.buyTime) || '请选择日期' }}</view>
        </picker>
        <picker mode="time" :value="timePart(model.buyTime) || '00:00'" @change="setDateTimePart('buyTime', 'time', $event)">
          <view class="datetime-value">{{ timePart(model.buyTime) || '请选择时间' }}</view>
        </picker>
        <button v-if="model.buyTime" class="clear-time-btn" @click="clearDateTime('buyTime')">清空</button>
      </view>
    </view>

    <view class="field">
      <text class="label required">收购地点</text>
      <view class="place-row">
        <input v-model="model.place" class="input place-input" placeholder="可手动输入或调用定位回填" @input="handlePlaceInput" />
        <button class="place-action" @click="$emit('select-current-location')">
          <text class="pin-mini"></text>
        </button>
      </view>
    </view>

    <view class="field">
      <text class="label">支付时间</text>
      <view class="datetime-row">
        <picker mode="date" :value="datePart(model.payTime) || todayDate" @change="setDateTimePart('payTime', 'date', $event)">
          <view class="datetime-value">{{ datePart(model.payTime) || '请选择日期' }}</view>
        </picker>
        <picker mode="time" :value="timePart(model.payTime) || '00:00'" @change="setDateTimePart('payTime', 'time', $event)">
          <view class="datetime-value">{{ timePart(model.payTime) || '请选择时间' }}</view>
        </picker>
        <button v-if="model.payTime" class="clear-time-btn" @click="clearDateTime('payTime')">清空</button>
      </view>
    </view>

    <view class="field">
      <view class="label-row">
        <text class="label">其他材料</text>
        <button class="scan-btn" @click="chooseMaterials">
          <text class="upload-mini"></text>
          <text>上传</text>
        </button>
      </view>
      <view class="material-grid">
        <image
          v-for="(image, index) in model.materialImages"
          :key="image"
          class="material-image"
          :src="image"
          mode="aspectFill"
          @click="previewMaterial(index)"
        />
        <button v-if="!model.materialImages.length" class="empty-upload" @click="chooseMaterials">
          <text class="upload-plus"></text>
          <text>添加身份证补充、票据、现场照片</text>
        </button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { calcUnitPrice } from '@/utils/grain'
import type { GrainEntryDraft, GrainPreset, GrainPurchaseType } from '@/types/grain'

const props = defineProps<{
  modelValue: GrainEntryDraft
  preset: GrainPreset
}>()

const emit = defineEmits<{
  'update:modelValue': [value: GrainEntryDraft]
  'select-current-location': []
}>()

const model = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
const unitOptions = ['公斤', '吨']
const unitIndex = computed(() => Math.max(0, unitOptions.indexOf(model.value.unit || '公斤')))
const priceText = computed(() => {
  const unit = model.value.unit || '公斤'
  return calcUnitPrice(Number(model.value.amount), Number(model.value.quantity), unit)
})
const cropKeyword = ref('')
const showCropResults = ref(false)
const todayDate = new Date().toISOString().slice(0, 10)
const filteredCropOptions = computed(() => {
  const key = cropKeyword.value.trim()
  if (!showCropResults.value && model.value.crop) {
    return []
  }
  return props.preset.purchaseTypes
    .filter((item) => !key || item.typeName.includes(key))
    .slice(0, 8)
    .map((item) => ({ id: item.id, name: item.typeName, unit: item.unit }))
})

watch(
  () => model.value.crop,
  (crop) => {
    cropKeyword.value = crop || ''
  },
  { immediate: true },
)

function selectUnit(event: { detail: { value: number | string } }) {
  model.value.unit = unitOptions[Number(event.detail.value)] || '公斤'
}

function selectCrop(crop: Pick<GrainPurchaseType, 'id' | 'unit'> & { name: string }) {
  model.value.purchaseTypeId = crop.id
  model.value.crop = crop.name
  model.value.unit = crop.unit || model.value.unit || '公斤'
  cropKeyword.value = model.value.crop
  showCropResults.value = false
}

function handleCropInput(event: unknown) {
  const value =
    typeof event === 'object' && event && 'detail' in event
      ? String((event as { detail?: { value?: string } }).detail?.value ?? cropKeyword.value)
      : cropKeyword.value
  cropKeyword.value = value
  model.value.purchaseTypeId = 0
  model.value.crop = ''
  showCropResults.value = true
}

function handlePlaceInput() {
  const option = props.preset.purchasePlaces.find((item) => item.placeName === model.value.place)
  model.value.placeId = option?.id || 0
  model.value.locationName = option?.placeName || ''
  model.value.locationAddress = option?.address || ''
  model.value.longitude = option?.longitude || ''
  model.value.latitude = option?.latitude || ''
  model.value.province = option?.province || ''
  model.value.city = option?.city || ''
  model.value.district = option?.district || ''
}

function datePart(value?: string) {
  return String(value || '').slice(0, 10)
}

function timePart(value?: string) {
  const match = String(value || '').match(/(\d{2}:\d{2})/)
  return match?.[1] || ''
}

function setDateTimePart(field: 'buyTime' | 'payTime', part: 'date' | 'time', event: { detail: { value: string } }) {
  const value = event.detail.value
  const nextDate = part === 'date' ? value : datePart(model.value[field]) || todayDate
  const nextTime = part === 'time' ? value : timePart(model.value[field]) || '00:00'
  model.value[field] = `${nextDate} ${nextTime}`
}

function clearDateTime(field: 'buyTime' | 'payTime') {
  model.value[field] = ''
}

function chooseMaterials() {
  uni.chooseImage({
    count: 6,
    success: (res) => {
      model.value.materialImages = [...model.value.materialImages, ...res.tempFilePaths].slice(0, 9)
    },
  })
}

function previewMaterial(index: number) {
  uni.previewImage({
    urls: model.value.materialImages,
    current: model.value.materialImages[index],
  })
}
</script>

<style lang="scss" scoped>
.form-card {
  padding: 30rpx;
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

.required::after {
  content: ' *';
  color: #d14343;
}

.input,
.picker-value,
.unit-value {
  width: 100%;
  height: 88rpx;
  padding: 0 24rpx;
  border: 1rpx solid #e2e8dd;
  border-radius: 18rpx;
  background: #fbfcfa;
  color: #172018;
  font-size: 28rpx;
  line-height: 88rpx;
}

.split {
  display: grid;
  grid-template-columns: 1fr 184rpx;
  gap: 16rpx;
}

.datetime-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 180rpx auto;
  gap: 12rpx;
  align-items: center;
}

.datetime-value {
  height: 88rpx;
  padding: 0 18rpx;
  border: 1rpx solid #e2e8dd;
  border-radius: 18rpx;
  background: #fbfcfa;
  color: #172018;
  font-size: 26rpx;
  line-height: 88rpx;
}

.clear-time-btn {
  min-width: 104rpx;
  height: 88rpx;
  padding: 0 18rpx;
  border: 1rpx solid #e2e8dd;
  border-radius: 18rpx;
  background: #ffffff;
  color: #6d776c;
  font-size: 24rpx;
  line-height: 88rpx;
}

.unit-value {
  background: #eef5ec;
  color: #145535;
  font-weight: 800;
  text-align: center;
}

.crop-options {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-top: 14rpx;
}

.crop-helper {
  margin-top: 10rpx;
  color: #7b857b;
  font-size: 23rpx;
  line-height: 1.45;
}

.crop-chip {
  min-height: 64rpx;
  padding: 0 22rpx;
  border: 1rpx solid #d8e5d6;
  border-radius: 999rpx;
  background: #ffffff;
  color: #2d4633;
  font-size: 25rpx;
  font-weight: 760;
  line-height: 64rpx;
}

.crop-chip.active {
  border-color: #237a4b;
  background: #e8f5ec;
  color: #145535;
}

.crop-empty {
  width: 100%;
  padding: 18rpx 20rpx;
  border-radius: 18rpx;
  background: #f8faf6;
  color: #667266;
  font-size: 24rpx;
  line-height: 1.45;
}

.place-row,
.label-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.place-input {
  min-width: 0;
  flex: 1;
}

.place-action {
  display: flex;
  width: 92rpx;
  height: 88rpx;
  align-items: center;
  justify-content: center;
  border: 1rpx solid #e2e8dd;
  border-radius: 18rpx;
  background: #ffffff;
}

.scan-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
  flex: 0 0 auto;
  border: 0;
  background: #edf7ff;
  color: #1d5d99;
  font-weight: 760;
}

.label-row {
  justify-content: space-between;
  margin-bottom: 14rpx;
}

.label-row .label {
  margin-bottom: 0;
}

.scan-btn {
  padding: 12rpx 18rpx;
  border-radius: 999rpx;
  font-size: 24rpx;
  line-height: 1.2;
}

.material-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14rpx;
}

.material-image,
.empty-upload {
  width: 100%;
  height: 148rpx;
  border-radius: 8rpx;
}

.material-image {
  background: #eef2ea;
}

.empty-upload {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  grid-column: 1 / -1;
  border: 1rpx dashed rgba(35, 122, 75, 0.36);
  background: #f8faf6;
  color: #48604e;
  font-size: 24rpx;
  line-height: 148rpx;
}

.pin-mini,
.upload-mini,
.upload-plus {
  position: relative;
  display: inline-block;
  flex: 0 0 auto;
}

.pin-mini {
  width: 26rpx;
  height: 26rpx;
  border: 4rpx solid #237a4b;
  border-radius: 50% 50% 50% 0;
  transform: rotate(-45deg);
}

.pin-mini::after {
  position: absolute;
  left: 6rpx;
  top: 6rpx;
  width: 6rpx;
  height: 6rpx;
  border-radius: 50%;
  background: #ffb84d;
  content: '';
}

.upload-mini,
.upload-plus {
  width: 28rpx;
  height: 28rpx;
}

.upload-mini::before,
.upload-plus::before {
  position: absolute;
  left: 12rpx;
  top: 3rpx;
  width: 4rpx;
  height: 18rpx;
  border-radius: 999rpx;
  background: currentColor;
  content: '';
}

.upload-mini::after,
.upload-plus::after {
  position: absolute;
  left: 7rpx;
  top: 3rpx;
  width: 12rpx;
  height: 12rpx;
  border-top: 4rpx solid currentColor;
  border-left: 4rpx solid currentColor;
  content: '';
  transform: rotate(45deg);
}

</style>
