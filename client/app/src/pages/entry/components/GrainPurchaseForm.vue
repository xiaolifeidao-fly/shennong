<template>
  <view class="card form-card">
    <view class="field">
      <text class="label required">购进农产品类型</text>
      <picker :value="cropIndex" :range="cropNames" @change="selectCrop">
        <view class="picker-value">{{ model.crop || '选择常用类型' }}</view>
      </picker>
    </view>
    <view class="field">
      <input v-model="model.crop" class="input" placeholder="也可手动输入或修改产品类型" @input="handleCropInput" />
    </view>

    <view class="field">
      <text class="label required">购进重量</text>
      <view class="split">
        <input v-model.number="model.quantity" class="input" type="digit" placeholder="请输入重量" />
        <view class="unit-value">公斤</view>
      </view>
    </view>

    <view class="field">
      <text class="label required">购进货物金额</text>
      <input v-model.number="model.amount" class="input" type="digit" placeholder="请输入金额" />
    </view>

    <view class="field">
      <text class="label">自动计算单价</text>
      <input class="input" :value="priceText" disabled />
    </view>

    <view class="field">
      <text class="label">收购时间</text>
      <input v-model="model.buyTime" class="input" />
    </view>

    <view class="field">
      <text class="label required">收购地点</text>
      <view class="place-row">
        <input v-model="model.place" class="input place-input" placeholder="可手动输入或调用定位回填" @input="handlePlaceInput" />
        <picker class="place-picker" :value="placeIndex" :range="placeNames" @change="selectPlace">
          <view class="place-action">常用</view>
        </picker>
        <button class="location-btn" @click="$emit('select-current-location')">调用定位</button>
      </view>
    </view>

    <view class="field">
      <view class="label-row">
        <text class="label">其他材料</text>
        <button class="scan-btn" @click="chooseMaterials">上传图片</button>
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
        <button v-if="!model.materialImages.length" class="empty-upload" @click="chooseMaterials">添加身份证补充、票据、现场照片</button>
      </view>
    </view>

    <button class="submit" :loading="saving" :disabled="saving" @click="$emit('submit')">
      {{ saving ? '正在保存...' : editing ? '保存修改并记录' : '保存本次录入' }}
    </button>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { calcUnitPrice } from '@/utils/grain'
import type { GrainEntryDraft, GrainPreset } from '@/types/grain'

const props = defineProps<{
  modelValue: GrainEntryDraft
  preset: GrainPreset
  editing?: boolean
  saving?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: GrainEntryDraft]
  'select-current-location': []
  submit: []
}>()

const model = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
const priceText = computed(() => calcUnitPrice(Number(model.value.amount), Number(model.value.quantity), model.value.unit))
const cropNames = computed(() => props.preset.purchaseTypes.length ? props.preset.purchaseTypes.map((item) => item.typeName) : props.preset.crops)
const placeNames = computed(() => props.preset.purchasePlaces.length ? props.preset.purchasePlaces.map((item) => item.placeName) : props.preset.places)
const cropIndex = computed(() => Math.max(0, cropNames.value.indexOf(model.value.crop)))
const placeIndex = computed(() => Math.max(0, placeNames.value.indexOf(model.value.place)))

function selectCrop(event: { detail: { value: number | string } }) {
  const option = props.preset.purchaseTypes[Number(event.detail.value)]
  model.value.purchaseTypeId = option?.id || 0
  model.value.crop = option?.typeName || cropNames.value[Number(event.detail.value)] || model.value.crop
  model.value.unit = option?.unit || model.value.unit || '公斤'
}

function selectPlace(event: { detail: { value: number | string } }) {
  const option = props.preset.purchasePlaces[Number(event.detail.value)]
  model.value.placeId = option?.id || 0
  model.value.place = option?.placeName || placeNames.value[Number(event.detail.value)] || model.value.place
  model.value.locationName = option?.placeName || model.value.locationName
  model.value.locationAddress = option?.address || model.value.locationAddress
  model.value.longitude = option?.longitude || model.value.longitude
  model.value.latitude = option?.latitude || model.value.latitude
  model.value.province = option?.province || model.value.province
  model.value.city = option?.city || model.value.city
  model.value.district = option?.district || model.value.district
}

function handleCropInput() {
  const option = props.preset.purchaseTypes.find((item) => item.typeName === model.value.crop)
  model.value.purchaseTypeId = option?.id || 0
  model.value.unit = option?.unit || model.value.unit || '公斤'
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
  padding: 28rpx;
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
  border-radius: 8rpx;
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

.unit-value {
  background: #eef5ec;
  color: #145535;
  font-weight: 800;
  text-align: center;
}

.place-row,
.label-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.place-picker {
  flex: 0 0 auto;
}

.place-input {
  min-width: 0;
  flex: 1;
}

.place-action {
  width: 116rpx;
  height: 88rpx;
  border: 1rpx solid #e2e8dd;
  border-radius: 8rpx;
  background: #ffffff;
  color: #243027;
  font-size: 24rpx;
  font-weight: 760;
  line-height: 88rpx;
  text-align: center;
}

.location-btn,
.scan-btn {
  flex: 0 0 auto;
  border: 0;
  background: #eaf2fb;
  color: #2563a8;
  font-weight: 760;
}

.location-btn {
  width: 172rpx;
  height: 88rpx;
  border-radius: 8rpx;
  font-size: 24rpx;
  line-height: 88rpx;
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
  grid-column: 1 / -1;
  border: 1rpx dashed rgba(35, 122, 75, 0.36);
  background: #f8faf6;
  color: #48604e;
  font-size: 24rpx;
  line-height: 148rpx;
}

.submit {
  width: 100%;
  min-height: 84rpx;
  border: 1rpx solid #237a4b;
  border-radius: 8rpx;
  background: #237a4b;
  color: #ffffff;
  font-size: 28rpx;
  font-weight: 800;
  line-height: 84rpx;
}
</style>
