<template>
  <view class="card form-card">
    <view class="field">
      <text class="label">购进农产品类型</text>
      <picker :value="cropIndex" :range="preset.crops" @change="selectCrop">
        <view class="picker-value">{{ model.crop }}</view>
      </picker>
    </view>

    <view class="field">
      <text class="label">购进数量</text>
      <view class="split">
        <input v-model.number="model.quantity" class="input" type="digit" />
        <picker :value="unitIndex" :range="units" @change="selectUnit">
          <view class="picker-value">{{ model.unit }}</view>
        </picker>
      </view>
    </view>

    <view class="field">
      <text class="label">购进货物金额</text>
      <input v-model.number="model.amount" class="input" type="digit" />
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
      <text class="label">收购地点</text>
      <picker :value="placeIndex" :range="preset.places" @change="selectPlace">
        <view class="picker-value">{{ model.place }}</view>
      </picker>
    </view>

    <view class="field">
      <text class="label">付款方式</text>
      <picker :value="payTypeIndex" :range="preset.payTypes" @change="selectPayType">
        <view class="picker-value">{{ model.payType }}</view>
      </picker>
    </view>

    <button class="submit" @click="$emit('submit')">保存本次录入</button>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { calcUnitPrice } from '@/utils/grain'
import type { GrainEntryDraft, GrainPreset } from '@/types/grain'

const props = defineProps<{
  modelValue: GrainEntryDraft
  preset: GrainPreset
}>()

const emit = defineEmits<{
  'update:modelValue': [value: GrainEntryDraft]
  submit: []
}>()

const units = ['斤', 'kg', '吨']
const model = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
const priceText = computed(() => calcUnitPrice(Number(model.value.amount), Number(model.value.quantity), model.value.unit))
const cropIndex = computed(() => Math.max(0, props.preset.crops.indexOf(model.value.crop)))
const unitIndex = computed(() => Math.max(0, units.indexOf(model.value.unit)))
const placeIndex = computed(() => Math.max(0, props.preset.places.indexOf(model.value.place)))
const payTypeIndex = computed(() => Math.max(0, props.preset.payTypes.indexOf(model.value.payType)))

function selectCrop(event: { detail: { value: number | string } }) {
  model.value.crop = props.preset.crops[Number(event.detail.value)] || model.value.crop
}

function selectUnit(event: { detail: { value: number | string } }) {
  model.value.unit = units[Number(event.detail.value)] || model.value.unit
}

function selectPlace(event: { detail: { value: number | string } }) {
  model.value.place = props.preset.places[Number(event.detail.value)] || model.value.place
}

function selectPayType(event: { detail: { value: number | string } }) {
  model.value.payType = props.preset.payTypes[Number(event.detail.value)] || model.value.payType
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

.input,
.picker-value {
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
