<template>
  <view class="entry-root">
    <view class="fixed-entry-actions">
      <button class="clear-form-btn" :disabled="saving" @click="confirmClearForm">
        <text class="clear-mini"></text>
        <text>清空</text>
      </button>
      <button class="fixed-save-btn" :loading="saving" :disabled="saving" @click="saveEntry">
        <text class="submit-icon"></text>
        <text>{{ saving ? '正在保存...' : editingEntryId ? '保存修改并记录' : '保存本次录入' }}</text>
      </button>
    </view>

    <view class="page entry-page">
      <SectionHeader title="新增收粮录入" action-text="查看今日汇总" @action="goFarmers" />
      <view v-if="lastSavedEntry" class="card saved-card">
        <view class="saved-head">
          <view>
            <text class="saved-title">{{ savedActionText }}</text>
            <text class="saved-meta">{{ lastSavedEntry.buyTime }} · {{ lastSavedEntry.place || lastSavedEntry.locationName }}</text>
          </view>
          <text class="saved-badge">{{ lastSavedEntry.crop }}</text>
        </view>
        <view class="saved-grid">
          <view class="kv"><text>农户</text><text class="value">{{ savedFarmerName }}</text></view>
          <view class="kv"><text>重量</text><text class="value">{{ lastSavedEntry.quantity.toLocaleString('zh-CN') }} 公斤</text></view>
          <view class="kv"><text>金额</text><text class="value">¥{{ lastSavedEntry.amount.toLocaleString('zh-CN') }}</text></view>
          <view class="kv"><text>付款方式</text><text class="value">{{ lastSavedEntry.payType }}</text></view>
        </view>
        <view class="saved-actions">
          <button class="secondary-btn" @click="editSavedEntry">
            <text class="edit-mini"></text>
            <text>查看/修改</text>
          </button>
          <button class="primary-btn" @click="startAnotherEntry">
            <text class="plus-mini"></text>
            <text>继续录入</text>
          </button>
        </view>
      </view>
      <view class="card history-card">
        <view class="field">
          <text class="label">搜索历史提交</text>
          <input v-model="historyKeyword" class="input" placeholder="输入农户、产品、地点，一键带出原录入" @focus="loadHistoryEntries" />
        </view>
        <view v-if="grainStore.entriesLoading" class="loading-strip">正在加载历史录入...</view>
        <picker :value="historyIndex" :range="historyOptions" range-key="label" @click="loadHistoryEntries" @change="applyHistoryEntry">
          <view class="picker-value">{{ activeHistoryLabel }}</view>
        </picker>
        <view v-if="editingEntryId" class="edit-strip">
          <text>正在修改已提交记录，保存后会同步服务端并保留修改快照。</text>
          <button class="clear-edit" @click="clearEditing">
            <text class="plus-mini dark"></text>
            <text>新增</text>
          </button>
        </view>
      </view>

      <FarmerIdentityForm
        v-model="draft"
        :farmers="grainStore.farmers"
        :preset="grainStore.preset"
        @farmer-change="handleFarmerChange"
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
import { computed, ref, watch } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import FarmerIdentityForm from './components/FarmerIdentityForm.vue'
import GrainPurchaseForm from './components/GrainPurchaseForm.vue'
import SectionHeader from '@/components/business/SectionHeader.vue'
import { useGrainStore } from '@/stores/grain'
import type { GrainEntry, GrainEntryDraft } from '@/types/grain'

const grainStore = useGrainStore()
const draft = ref<GrainEntryDraft>(grainStore.createEntryDraft())
const historyKeyword = ref('')
const editingEntryId = ref('')
const hasInitializedDraft = ref(false)
const saving = ref(false)
const lastSavedEntry = ref<GrainEntry>()
const lastSavedAction = ref<'create' | 'update'>('create')

const historyOptions = computed(() => {
  const key = historyKeyword.value.trim()
  return [...grainStore.entries]
    .sort((left, right) => new Date(right.buyTime).getTime() - new Date(left.buyTime).getTime())
    .filter((entry) => {
      const farmer = grainStore.farmers.find((item) => item.id === entry.farmerId)
      const fields = [farmer?.name || '', farmer?.phone || '', entry.crop, entry.place, entry.buyTime]
      return !key || fields.some((value) => value.includes(key))
    })
    .slice(0, 10)
    .map((entry) => {
      const farmer = grainStore.farmers.find((item) => item.id === entry.farmerId)
      return {
        id: entry.id,
        label: `${farmer?.name || '未知农户'} · ${entry.crop} · ${entry.quantity}公斤 · ${entry.buyTime}`,
      }
    })
})
const historyIndex = computed(() => Math.max(0, historyOptions.value.findIndex((item) => item.id === editingEntryId.value)))
const activeHistoryLabel = computed(() => {
  if (!historyOptions.value.length) {
    return historyKeyword.value ? '没有匹配的历史提交' : '选择一笔历史提交自动填入'
  }
  return editingEntryId.value
    ? historyOptions.value[historyIndex.value]?.label || '选择一笔历史提交自动填入'
    : '选择一笔历史提交自动填入'
})
const savedFarmerName = computed(() => {
  if (!lastSavedEntry.value) {
    return ''
  }
  return grainStore.farmers.find((farmer) => farmer.id === lastSavedEntry.value?.farmerId)?.name || draft.value.farmerName || '新农户'
})
const savedActionText = computed(() => (lastSavedAction.value === 'update' ? '修改已保存' : '录入已保存'))

onShow(async () => {
  await Promise.all([grainStore.loadPreset(), grainStore.loadFarmers()])
  if (!hasInitializedDraft.value && !editingEntryId.value) {
    draft.value = grainStore.createEntryDraft()
    hasInitializedDraft.value = true
  }
})

watch(
  () => grainStore.selectedFarmerId,
  (farmerId) => {
    if (!editingEntryId.value) {
      draft.value = grainStore.createEntryDraft(farmerId)
    }
  },
)

function handleFarmerChange(farmerId: string) {
  editingEntryId.value = ''
  grainStore.selectFarmer(farmerId)
  draft.value = grainStore.createEntryDraft(farmerId)
}

function applyHistoryEntry(event: { detail: { value: number | string } }) {
  const entryId = historyOptions.value[Number(event.detail.value)]?.id
  if (!entryId) {
    uni.showToast({ title: '没有可带出的历史记录', icon: 'none' })
    return
  }

  editingEntryId.value = entryId
  draft.value = grainStore.createEntryDraftFromHistory(entryId)
  grainStore.selectFarmer(draft.value.farmerId)
  uni.showToast({ title: '已带出历史录入', icon: 'success' })
}

function loadHistoryEntries() {
  void grainStore.loadEntries()
}

function clearEditing() {
  editingEntryId.value = ''
  draft.value = grainStore.createEntryDraft(grainStore.selectedFarmerId)
  hasInitializedDraft.value = true
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
  historyKeyword.value = ''
  editingEntryId.value = ''
  lastSavedEntry.value = undefined
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

async function applyBankScan() {
  try {
    const filePath = await chooseCardPhoto()
    draft.value = { ...draft.value, ...(await grainStore.recognizeBankCard(filePath, draft.value)) }
    uni.showToast({ title: '银行卡识别完成', icon: 'success' })
  } catch (error) {
    const message = error instanceof Error ? error.message : '银行卡识别失败'
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
  uni.showLoading({ title: editingEntryId.value ? '正在保存修改' : '正在保存录入', mask: true })
  try {
    let saved: GrainEntry
    if (editingEntryId.value) {
      saved = await grainStore.updateEntry(editingEntryId.value, draft.value)
      lastSavedAction.value = 'update'
      uni.showToast({ title: '修改已同步', icon: 'success' })
      editingEntryId.value = ''
    } else {
      saved = await grainStore.saveEntry(draft.value)
      lastSavedAction.value = 'create'
      uni.showToast({ title: '保存成功', icon: 'success' })
    }
    lastSavedEntry.value = undefined
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

function editSavedEntry() {
  if (!lastSavedEntry.value) {
    return
  }
  editingEntryId.value = lastSavedEntry.value.id
  void grainStore.loadEntries(true, lastSavedEntry.value.farmerId)
  draft.value = grainStore.createEntryDraftFromHistory(lastSavedEntry.value.id)
}

function startAnotherEntry() {
  lastSavedEntry.value = undefined
  editingEntryId.value = ''
  draft.value = grainStore.createEntryDraft(grainStore.selectedFarmerId)
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
  if (!value.payType.trim()) {
    return '请选择付款方式'
  }
  const paymentMethodCode =
    value.paymentMethodCode ||
    grainStore.preset.paymentMethods.find((item) => item.id === value.paymentMethodId)?.methodCode ||
    grainStore.preset.paymentMethods.find((item) => item.methodName === value.payType)?.methodCode ||
    ''
  const isBankPayment = paymentMethodCode === 'Bank'
  const isAccountPayment = paymentMethodCode === 'Alipay' || paymentMethodCode === 'WECHAT'
  if (isBankPayment && !value.bankNumber.trim()) {
    return '请扫描或填写银行卡号'
  }
  if (isAccountPayment && !value.bankName.trim()) {
    return '请填写收款人姓名'
  }
  if (isAccountPayment && !value.bankNumber.trim()) {
    return '请填写收款账号'
  }
  if (!value.crop.trim()) {
    return '请选择或填写收购类型'
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

.history-card {
  padding: 30rpx;
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

.saved-card {
  padding: 30rpx;
  border: 1rpx solid rgba(35, 122, 75, 0.18);
  background: linear-gradient(135deg, #f4fbf5, #ffffff);
}

.saved-head,
.saved-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
}

.saved-title,
.saved-meta {
  display: block;
}

.saved-title {
  color: #145535;
  font-size: 32rpx;
  font-weight: 800;
}

.saved-meta {
  margin-top: 8rpx;
  color: #52605a;
  font-size: 24rpx;
}

.saved-badge {
  flex: 0 0 auto;
  padding: 10rpx 16rpx;
  border-radius: 999rpx;
  background: #ffffff;
  color: #145535;
  font-size: 24rpx;
  font-weight: 760;
}

.saved-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14rpx;
  margin-top: 22rpx;
}

.kv {
  min-width: 0;
  padding: 18rpx;
  border-radius: 8rpx;
  background: rgba(255, 255, 255, 0.86);
}

.kv text {
  display: block;
  color: #6d776c;
  font-size: 22rpx;
}

.kv .value {
  margin-top: 8rpx;
  color: #172018;
  font-size: 25rpx;
  font-weight: 760;
  overflow-wrap: anywhere;
}

.saved-actions {
  margin-top: 22rpx;
}

.primary-btn,
.secondary-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10rpx;
  flex: 1;
  min-height: 78rpx;
  border-radius: 18rpx;
  font-size: 26rpx;
  font-weight: 800;
  line-height: 78rpx;
}

.primary-btn {
  border: 1rpx solid #237a4b;
  background: linear-gradient(135deg, #237a4b, #145535);
  color: #ffffff;
  box-shadow: 0 14rpx 28rpx rgba(35, 122, 75, 0.18);
}

.secondary-btn {
  border: 1rpx solid #cfe0d1;
  background: #ffffff;
  color: #145535;
}

.field {
  margin-bottom: 20rpx;
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
  border-radius: 18rpx;
  background: #fbfcfa;
  color: #172018;
  font-size: 28rpx;
  line-height: 88rpx;
}

.edit-strip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  padding: 20rpx;
  margin-top: 20rpx;
  border-radius: 18rpx;
  background: #fff7ed;
  color: #8a4b11;
  font-size: 24rpx;
}

.loading-strip {
  padding: 18rpx 22rpx;
  margin-bottom: 18rpx;
  border-radius: 18rpx;
  background: #f6f8f4;
  color: #667266;
  font-size: 24rpx;
}

.clear-edit {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
  flex: 0 0 auto;
  min-width: 118rpx;
  height: 64rpx;
  border: 1rpx solid #fed7aa;
  border-radius: 16rpx;
  background: #ffffff;
  color: #9a3412;
  font-size: 24rpx;
  line-height: 64rpx;
}

.edit-mini,
.clear-mini,
.plus-mini,
.submit-icon {
  position: relative;
  display: inline-block;
  flex: 0 0 auto;
}

.edit-mini {
  width: 26rpx;
  height: 8rpx;
  border-radius: 999rpx;
  background: #145535;
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

.plus-mini.dark {
  color: #9a3412;
}

.submit-icon {
  width: 28rpx;
  height: 18rpx;
  border-left: 5rpx solid #ffffff;
  border-bottom: 5rpx solid #ffffff;
  transform: rotate(-45deg) translateY(-3rpx);
}
</style>
