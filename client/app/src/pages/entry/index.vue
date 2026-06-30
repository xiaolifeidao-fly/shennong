<template>
  <view class="entry-root">
    <view class="fixed-entry-actions">
      <button class="clear-form-btn" :disabled="saving" @click="confirmClearForm">
        <text class="clear-mini"></text>
        <text>清空</text>
      </button>
      <button class="fixed-save-btn" :loading="saving" :disabled="saving" @click="saveEntry">
        <text class="submit-icon"></text>
        <text>{{ saving ? '正在保存...' : '保存本次录入' }}</text>
      </button>
    </view>

    <view class="page entry-page">
      <SectionHeader title="新增收粮录入" action-text="查看今日汇总" @action="goFarmers" />

      <FarmerIdentityForm
        v-model="draft"
        :farmers="grainStore.farmers"
        :farmer-searching="grainStore.farmersLoading"
        :preset="grainStore.preset"
        @farmer-change="handleFarmerChange"
        @farmer-search="handleFarmerSearch"
        @scan-id-front="applyIdScan"
        @scan-bank="applyBankScan"
      />

      <SectionHeader title="本笔粮食信息" />
      <view v-if="grainStore.presetLoading || grainStore.farmersLoading" class="loading-strip">正在加载录入配置...</view>
      <GrainPurchaseForm
        v-model="draft"
        :preset="grainStore.preset"
        @select-current-location="selectCurrentLocation"
      />
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import FarmerIdentityForm from './components/FarmerIdentityForm.vue'
import GrainPurchaseForm from './components/GrainPurchaseForm.vue'
import SectionHeader from '@/components/business/SectionHeader.vue'
import { useGrainStore } from '@/stores/grain'
import type { GrainEntryDraft, PayCardOcrType } from '@/types/grain'

const grainStore = useGrainStore()
const draft = ref<GrainEntryDraft>(grainStore.createEntryDraft())
const hasInitializedDraft = ref(false)
const saving = ref(false)

onShow(async () => {
  await Promise.all([grainStore.loadPreset(), grainStore.loadFarmers(false, 20)])
  if (!hasInitializedDraft.value) {
    draft.value = grainStore.createEntryDraft()
    hasInitializedDraft.value = true
  }
})

watch(
  () => grainStore.selectedFarmerId,
  (farmerId) => {
    draft.value = grainStore.createEntryDraft(farmerId)
  },
)

function handleFarmerChange(farmerId: string) {
  grainStore.selectFarmer(farmerId)
  draft.value = grainStore.createEntryDraft(farmerId)
}

async function handleFarmerSearch(keyword: string) {
  await grainStore.searchFarmers(keyword)
}

function confirmClearForm() {
  if (saving.value) {
    return
  }

  uni.showModal({
    title: '清空当前表单',
    content: '将清空已填写的农户、粮食、付款、地点和材料信息，是否继续？',
    confirmText: '清空',
    confirmColor: '#d14343',
    success: (res) => {
      if (res.confirm) {
        clearCurrentForm()
      }
    },
  })
}

function clearCurrentForm() {
  grainStore.selectFarmer('new')
  draft.value = grainStore.createEntryDraft('new')
  hasInitializedDraft.value = true
  uni.showToast({ title: '表单已清空', icon: 'success' })
}

async function chooseCardPhoto(): Promise<string> {
  return new Promise((resolve, reject) => {
    uni.showActionSheet({
      itemList: ['拍照', '从相册选择'],
      success: (res) => {
        const sourceType: Array<'camera' | 'album'> = res.tapIndex === 0 ? ['camera'] : ['album']
        uni.chooseImage({
          count: 1,
          sizeType: ['compressed'],
          sourceType,
          success: (imgRes) => {
            const filePath = imgRes.tempFilePaths?.[0]
            if (!filePath) {
              reject(new Error('未获取到照片'))
              return
            }
            resolve(filePath)
          },
          fail: (err) => reject(new Error(err.errMsg || '取消选择')),
        })
      },
      fail: (err) => reject(new Error(err.errMsg || '取消')),
    })
  })
}

async function applyIdScan() {
  try {
    const filePath = await chooseCardPhoto()
    draft.value = { ...draft.value, ...(await grainStore.recognizeIdCard(filePath, draft.value, 'front')) }
    uni.showToast({ title: '身份证识别完成', icon: 'success' })
  } catch (error) {
    const message = error instanceof Error ? error.message : '身份证识别失败'
    if (!message.includes('cancel') && !message.includes('取消')) {
      uni.showToast({ title: message, icon: 'none' })
    }
  }
}

async function applyBankScan(cardType: PayCardOcrType = 'bank-card') {
  const cardLabel = cardType === 'social-security-card' ? '社保卡' : '银行卡'
  try {
    const filePath = await chooseCardPhoto()
    draft.value = { ...draft.value, ...(await grainStore.recognizeBankCard(filePath, draft.value, cardType)) }
    uni.showToast({ title: `${cardLabel}识别完成`, icon: 'success' })
  } catch (error) {
    const message = error instanceof Error ? error.message : `${cardLabel}识别失败`
    if (!message.includes('cancel') && !message.includes('取消')) {
      uni.showToast({ title: message, icon: 'none' })
    }
  }
}

async function saveEntry() {
  if (saving.value) {
    return
  }

  const validateMessage = validateDraft(draft.value)
  if (validateMessage) {
    uni.showToast({ title: validateMessage, icon: 'none' })
    return
  }

  saving.value = true
  uni.showLoading({ title: '正在保存录入', mask: true })
  try {
    const saved = await grainStore.saveEntry(draft.value)
    uni.showToast({ title: '保存成功', icon: 'success' })
    grainStore.selectEntry(saved.id)
    draft.value = grainStore.createEntryDraft(grainStore.selectedFarmerId)
    hasInitializedDraft.value = true
    uni.navigateTo({ url: `/pages/farmers/entry-detail?entryId=${saved.id}` })
  } catch (error) {
    uni.showToast({ title: error instanceof Error ? error.message : '保存失败', icon: 'none' })
  } finally {
    saving.value = false
    uni.hideLoading()
  }
}

function selectCurrentLocation() {
  uni.chooseLocation({
    success: (location) => {
      const place = location.name || location.address || draft.value.place
      draft.value = {
        ...draft.value,
        place,
        placeId: 0,
        locationName: location.name || place,
        locationAddress: location.address || place,
        longitude: location.longitude ? String(location.longitude) : '',
        latitude: location.latitude ? String(location.latitude) : '',
      }
    },
    fail: (error) => {
      const message = error.errMsg || ''
      const title = message.includes('cancel')
        ? '已取消选择位置'
        : message.includes('auth') || message.includes('authorize') || message.includes('deny')
          ? '请先允许定位权限'
          : '定位调用失败，请检查小程序定位配置'

      console.warn('chooseLocation failed', error)
      uni.showToast({ title, icon: 'none' })
    },
  })
}

function goFarmers() {
  uni.switchTab({ url: '/pages/farmers/index' })
}

function validateDraft(value: GrainEntryDraft) {
  if (!value.farmerName.trim()) {
    return '请填写农户姓名'
  }
  if (!value.idNumber.trim()) {
    return '请填写身份证号'
  }
  const phone = value.phone.trim()
  if (!phone) {
    return '请填写农户电话'
  }
  if (!/^\d{11}$/.test(phone)) {
    return '农户电话必须为 11 位数字'
  }
  if (!grainStore.preset.purchaseTypes.length) {
    return '当前粮站暂无粮食类型，请先联系管理员维护'
  }
  const selectedPurchaseType = grainStore.preset.purchaseTypes.find(
    (item) => item.id === Number(value.purchaseTypeId) && item.typeName === value.crop,
  )
  if (!selectedPurchaseType) {
    return '收购粮食类型为必填项，请从已有粮食类型中选择'
  }
  if (Number(value.quantity) <= 0) {
    return '请填写购进重量'
  }
  if (Number(value.amount) <= 0) {
    return '请填写购进货物金额'
  }
  if (!(value.place || value.locationName).trim()) {
    return '请选择或填写收购地点'
  }
  return ''
}
</script>

<style lang="scss" scoped>
.entry-root {
  min-height: 100vh;
}

.entry-page {
  display: flex;
  flex-direction: column;
  padding-top: 132rpx;
}

.fixed-entry-actions {
  position: fixed;
  z-index: 20;
  top: 0;
  right: 0;
  left: 0;
  display: grid;
  grid-template-columns: 180rpx minmax(0, 1fr);
  gap: 16rpx;
  padding: 18rpx 28rpx 16rpx;
  border-bottom: 1rpx solid rgba(216, 229, 214, 0.9);
  background: rgba(247, 251, 244, 0.98);
  box-shadow: 0 12rpx 30rpx rgba(31, 47, 31, 0.08);
}

.clear-form-btn,
.fixed-save-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10rpx;
  height: 84rpx;
  border-radius: 18rpx;
  font-weight: 800;
  line-height: 84rpx;
}

.clear-form-btn {
  border: 1rpx solid #f2c6c6;
  background: #ffffff;
  color: #b42323;
  font-size: 26rpx;
}

.fixed-save-btn {
  border: 1rpx solid #237a4b;
  background: linear-gradient(135deg, #237a4b, #145535);
  color: #ffffff;
  font-size: 28rpx;
  box-shadow: 0 16rpx 30rpx rgba(35, 122, 75, 0.2);
}

.clear-form-btn[disabled],
.fixed-save-btn[disabled] {
  opacity: 0.55;
}

.loading-strip {
  padding: 18rpx 22rpx;
  margin-bottom: 18rpx;
  border-radius: 18rpx;
  background: #f6f8f4;
  color: #667266;
  font-size: 24rpx;
}

.clear-mini,
.submit-icon {
  position: relative;
  display: inline-block;
  flex: 0 0 auto;
}

.clear-mini {
  width: 28rpx;
  height: 28rpx;
  border: 3rpx solid currentColor;
  border-top: 0;
  border-radius: 0 0 7rpx 7rpx;
}

.clear-mini::before {
  position: absolute;
  left: 3rpx;
  top: -7rpx;
  width: 22rpx;
  height: 4rpx;
  border-radius: 999rpx;
  background: currentColor;
  content: '';
}

.clear-mini::after {
  position: absolute;
  left: 9rpx;
  top: -12rpx;
  width: 10rpx;
  height: 4rpx;
  border-radius: 999rpx;
  background: currentColor;
  content: '';
}

.submit-icon {
  width: 28rpx;
  height: 18rpx;
  border-left: 5rpx solid #ffffff;
  border-bottom: 5rpx solid #ffffff;
  transform: rotate(-45deg) translateY(-3rpx);
}
</style>
