<template>
  <view class="page detail-page">
    <view v-if="!farmer" class="empty-state">农户信息不存在</view>
    <template v-else>
      <!-- 农户资料卡片 -->
      <view class="section-card">
        <view class="card-header">
          <view>
            <text class="farmer-name">{{ farmer.name }}</text>
            <text class="farmer-status">{{ farmer.statusText }}</text>
          </view>
          <button v-if="!editing" class="edit-btn" @click="startEdit">
            <text class="icon-edit"></text>
            <text>编辑</text>
          </button>
        </view>

        <template v-if="editing">
          <view class="field">
            <text class="label">农户姓名</text>
            <input v-model="editForm.name" class="input" placeholder="请输入姓名" />
          </view>
          <view class="field">
            <text class="label">身份证号</text>
            <input v-model="editForm.idNumber" class="input" placeholder="请输入身份证号" />
          </view>
          <view class="field">
            <text class="label">身份证住址</text>
            <textarea v-model="editForm.address" class="textarea" placeholder="请输入地址" />
          </view>
          <view class="field">
            <text class="label">农户电话</text>
            <input v-model="editForm.phone" class="input" placeholder="请输入电话" />
          </view>
          <view class="field">
            <text class="label">收款账号</text>
            <input v-model="editForm.bankNumber" class="input" placeholder="请输入银行卡号" />
          </view>
          <view class="field">
            <text class="label">收款人/开户行</text>
            <input v-model="editForm.bankName" class="input" placeholder="请输入开户行" />
          </view>
          <view class="btn-row">
            <button class="secondary-btn" @click="cancelEdit">取消</button>
            <button class="primary-btn" @click="saveFarmer">保存</button>
          </view>
        </template>
        <template v-else>
          <view class="info-grid">
            <view class="kv">
              <text class="kv-label">身份证号</text>
              <text class="kv-value">{{ farmer.idNumber || '-' }}</text>
            </view>
            <view class="kv">
              <text class="kv-label">电话</text>
              <text class="kv-value">{{ farmer.phone || '-' }}</text>
            </view>
          </view>
          <view class="kv kv-full">
            <text class="kv-label">住址</text>
            <text class="kv-value">{{ farmer.address || '-' }}</text>
          </view>
          <view class="info-grid">
            <view class="kv">
              <text class="kv-label">收款账号</text>
              <text class="kv-value">{{ farmer.bankNumber || '-' }}</text>
            </view>
            <view class="kv">
              <text class="kv-label">收款人/开户行</text>
              <text class="kv-value">{{ farmer.bankName || '-' }}</text>
            </view>
          </view>

          <!-- 证件图片 -->
          <view class="img-section">
            <text class="img-section-title">证件照片</text>
            <view class="img-row">
              <view class="img-slot">
                <text class="img-label">身份证正面</text>
                <image
                  v-if="farmerImages.idCardFront"
                  :key="farmerImages.idCardFront"
                  :src="farmerImages.idCardFront"
                  class="cert-img"
                  mode="aspectFill"
                  @click="previewImage(farmerImages.idCardFront, [farmerImages.idCardFront])"
                />
                <view v-else class="img-placeholder">
                  <text>暂无图片</text>
                </view>
                <button class="img-action" @click="uploadFarmerCardImage('id-front')">重新OCR</button>
              </view>
              <view class="img-slot">
                <text class="img-label">银行卡</text>
                <image
                  v-if="farmerImages.bankCard"
                  :key="farmerImages.bankCard"
                  :src="farmerImages.bankCard"
                  class="cert-img"
                  mode="aspectFill"
                  @click="previewImage(farmerImages.bankCard, [farmerImages.bankCard])"
                />
                <view v-else class="img-placeholder">
                  <text>暂无图片</text>
                </view>
                <view class="img-action-row">
                  <picker :value="payCardTypeIndex" :range="payCardTypeNames" @change="selectPayCardType">
                    <view class="card-type-picker">
                      <text>{{ payCardTypeLabel }}</text>
                      <text class="caret"></text>
                    </view>
                  </picker>
                  <button class="img-action" @click="uploadFarmerCardImage('bank', payCardType)">重新OCR</button>
                </view>
              </view>
            </view>
          </view>

          <button class="entry-btn" @click="goEntry">
            <text>+ 录粮</text>
          </button>
        </template>
      </view>

      <!-- 粮食类型汇总 -->
      <view class="section-title">按粮食类型汇总</view>
      <view class="section-card summary-card">
        <view v-if="grainStore.summariesLoading" class="loading-text">正在加载...</view>
        <template v-else-if="summaries.length">
          <view class="summary-head">
            <text>粮食类型</text>
            <text>付款方式</text>
            <text>总公斤</text>
            <text>总金额</text>
          </view>
          <view v-for="s in summaries" :key="s.id" class="summary-row">
            <text>{{ s.crop || '-' }}</text>
            <text>{{ s.payType || '-' }}</text>
            <text>{{ s.totalQuantity.toLocaleString('zh-CN') }}</text>
            <text>{{ formatAmount(s.totalAmount) }}</text>
          </view>
        </template>
        <view v-else class="empty-text">暂无汇总数据</view>
      </view>

      <!-- 收粮明细 -->
      <view class="section-title">收粮明细</view>
      <view class="entries-list">
        <view
          v-for="entry in pagedEntries"
          :key="entry.id"
          class="entry-card"
          @click="goEntryDetail(entry.id)"
        >
          <view class="entry-head">
            <view>
              <text class="entry-title">{{ entry.buyTime }} · {{ entry.crop }}</text>
              <text class="entry-place">{{ entry.place }}</text>
            </view>
            <view class="entry-price-badge">
              <text>{{ getEntryPrice(entry) }}</text>
            </view>
            <button class="entry-delete-btn" @click.stop="confirmDeleteEntry(entry.id)">删除</button>
          </view>
          <view class="entry-grid">
            <view class="entry-kv">
              <text>数量</text>
              <text>{{ formatQuantity(entry.quantity, entry.unit) }}</text>
            </view>
            <view class="entry-kv">
              <text>金额</text>
              <text>{{ formatAmount(entry.amount) }}</text>
            </view>
            <view class="entry-kv">
              <text>付款</text>
              <text>{{ entry.payType }}</text>
            </view>
          </view>
        </view>
      </view>

      <view v-if="grainStore.entriesLoading" class="loading-text loading-more">正在加载...</view>
      <view
        v-else-if="hasMore"
        class="load-more"
        @click="loadMore"
      >
        加载更多
      </view>
      <view v-else-if="pagedEntries.length > 0" class="no-more">已加载全部</view>
      <view v-else-if="!grainStore.entriesLoading" class="empty-text">暂无收粮明细</view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { onLoad, onReachBottom } from '@dcloudio/uni-app'
import { useGrainStore } from '@/stores/grain'
import { formatAmount, formatQuantity, getEntryPrice } from '@/utils/grain'
import { PAY_CARD_OCR_OPTIONS, type FarmerProfile, type GrainEntryDraft, type PayCardOcrType } from '@/types/grain'
import type { FarmerImagesResult } from '@/services/grainFarmer'

const grainStore = useGrainStore()

const farmer = computed(() => grainStore.selectedFarmer)
const summaries = computed(() => grainStore.summaries.filter((s) => s.farmerId === grainStore.selectedFarmerId))
const farmerImages = computed((): FarmerImagesResult => grainStore.farmerImages[grainStore.selectedFarmerId] ?? { idCardFront: '', idCardBack: '', bankCard: '' })

const allEntries = computed(() => grainStore.farmerEntriesSorted)
const pagedEntries = computed(() => allEntries.value)
const hasMore = computed(() => grainStore.farmerEntriesHasMore)

const payCardType = ref<PayCardOcrType>('bank-card')
const payCardTypeNames = PAY_CARD_OCR_OPTIONS.map((item) => item.label)
const payCardTypeIndex = computed(() => Math.max(0, PAY_CARD_OCR_OPTIONS.findIndex((item) => item.value === payCardType.value)))
const payCardTypeLabel = computed(() => PAY_CARD_OCR_OPTIONS[payCardTypeIndex.value]?.label || '银行卡')

function selectPayCardType(event: { detail: { value: number | string } }) {
  payCardType.value = PAY_CARD_OCR_OPTIONS[Number(event.detail.value)]?.value || 'bank-card'
}

const editing = ref(false)
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

watch(farmer, (f) => {
  if (f) Object.assign(editForm, f)
}, { immediate: true })

onLoad(() => {
  const farmerId = grainStore.selectedFarmerId
  if (farmerId && farmerId !== 'new') {
    void grainStore.loadFarmerEntriesPaged(true, farmerId)
    void grainStore.loadSummaries(true, farmerId)
    void grainStore.loadFarmerImages(farmerId)
  }
})

onReachBottom(() => {
  if (hasMore.value) {
    void grainStore.loadFarmerEntriesPaged(false)
  }
})

function startEdit() {
  editing.value = true
}

function cancelEdit() {
  if (farmer.value) Object.assign(editForm, farmer.value)
  editing.value = false
}

async function saveFarmer() {
  try {
    await grainStore.updateFarmer(editForm.id, { ...editForm })
    editing.value = false
    uni.showToast({ title: '已保存农户资料', icon: 'success' })
  } catch (error) {
    uni.showToast({ title: error instanceof Error ? error.message : '保存失败', icon: 'none' })
  }
}

function goEntry() {
  uni.switchTab({ url: '/pages/entry/index' })
}

function goEntryDetail(entryId: string) {
  grainStore.selectEntry(entryId)
  uni.navigateTo({ url: '/pages/farmers/entry-detail' })
}

function confirmDeleteEntry(entryId: string) {
  uni.showModal({
    title: '删除粮食明细',
    content: '确认删除该笔粮食明细？删除后会同步更新农户汇总数据。',
    confirmText: '删除',
    confirmColor: '#c2412d',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await grainStore.voidEntry(entryId)
        uni.showToast({ title: '已删除明细', icon: 'success' })
      } catch (error) {
        uni.showToast({ title: error instanceof Error ? error.message : '删除失败', icon: 'none' })
      }
    },
  })
}

function loadMore() {
  void grainStore.loadFarmerEntriesPaged(false)
}

function previewImage(current: string, urls: string[]) {
  uni.previewImage({ current, urls })
}

function buildImageDraft(profile: FarmerProfile): GrainEntryDraft {
  return {
    farmerId: profile.id,
    farmerName: profile.name,
    idNumber: profile.idNumber,
    phone: profile.phone,
    address: profile.address,
    bankNumber: profile.bankNumber,
    bankName: profile.bankName,
    purchaseTypeId: 0,
    crop: '',
    quantity: 0,
    unit: '公斤',
    amount: 0,
    buyTime: '',
    payTime: '',
    placeId: 0,
    place: '',
    locationName: '',
    locationAddress: '',
    longitude: '',
    latitude: '',
    province: '',
    city: '',
    district: '',
    paymentMethodId: 0,
    paymentMethodCode: '',
    payType: '',
    materialImages: [],
    cardImages: {},
  }
}

function uploadFarmerCardImage(type: 'id-front' | 'bank', payCardType: PayCardOcrType = 'bank-card') {
  if (!farmer.value) return
  const farmerId = farmer.value.id
  uni.chooseImage({
    count: 1,
    success: async (res) => {
      const filePath = res.tempFilePaths[0]
      if (!filePath || !farmer.value) return
      try {
        const draft = buildImageDraft(farmer.value)
        if (type === 'bank') {
          const result = await grainStore.recognizeBankCard(filePath, draft, payCardType)
          await grainStore.saveFarmerCardImages(farmer.value.id, result.cardImages)
          await grainStore.updateFarmer(farmer.value.id, {
            ...farmer.value,
            bankNumber: result.bankNumber,
            bankName: result.bankName,
          })
        } else {
          const result = await grainStore.recognizeIdCard(filePath, draft, 'front')
          await grainStore.saveFarmerCardImages(farmer.value.id, result.cardImages)
          await grainStore.updateFarmer(farmer.value.id, {
            ...farmer.value,
            name: result.farmerName,
            idNumber: result.idNumber,
            address: result.address,
          })
        }
        await grainStore.loadFarmers(true)
        await grainStore.loadFarmerImages(farmerId)
        uni.showToast({ title: '图片已更新', icon: 'success' })
      } catch (error) {
        uni.showToast({ title: error instanceof Error ? error.message : 'OCR上传失败', icon: 'none' })
      }
    },
  })
}
</script>

<style lang="scss" scoped>
.detail-page {
  padding: 24rpx;
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.empty-state {
  padding: 80rpx;
  text-align: center;
  color: #7b857b;
  font-size: 28rpx;
}

.section-title {
  color: #384338;
  font-size: 28rpx;
  font-weight: 760;
  padding: 0 4rpx;
}

.section-card {
  border-radius: 24rpx;
  background: #ffffff;
  padding: 28rpx;
  box-shadow: 0 4rpx 16rpx rgba(20, 85, 53, 0.06);
}

.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24rpx;
}

.farmer-name {
  display: block;
  color: #172018;
  font-size: 36rpx;
  font-weight: 760;
}

.farmer-status {
  display: block;
  margin-top: 8rpx;
  color: #6d776c;
  font-size: 24rpx;
}

.edit-btn {
  display: flex;
  align-items: center;
  gap: 8rpx;
  padding: 14rpx 24rpx;
  border-radius: 999rpx;
  border: 1rpx solid #e2e8dd;
  background: #ffffff;
  color: #384338;
  font-size: 24rpx;
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16rpx;
  margin-top: 16rpx;
}

.kv {
  padding: 18rpx 20rpx;
  border-radius: 16rpx;
  background: #f8faf6;
}

.kv-full {
  margin-top: 16rpx;
}

.kv-label {
  display: block;
  color: #6d776c;
  font-size: 22rpx;
}

.kv-value {
  display: block;
  margin-top: 8rpx;
  color: #172018;
  font-size: 26rpx;
  font-weight: 700;
  overflow-wrap: anywhere;
}

.img-section {
  margin-top: 24rpx;
}

.img-section-title {
  display: block;
  color: #6d776c;
  font-size: 22rpx;
  margin-bottom: 14rpx;
}

.img-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16rpx;
}

.img-slot {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
}

.img-label {
  color: #6d776c;
  font-size: 20rpx;
  text-align: center;
}

.cert-img {
  width: 100%;
  height: 140rpx;
  border-radius: 12rpx;
  border: 1rpx solid #e2e8dd;
}

.img-placeholder {
  width: 100%;
  height: 140rpx;
  border-radius: 12rpx;
  border: 1rpx dashed #c5d0bf;
  background: #f4f8f3;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9aaa96;
  font-size: 22rpx;
}

.img-action {
  width: 100%;
  min-height: 58rpx;
  border: 1rpx solid #d8e5d6;
  border-radius: 12rpx;
  background: #ffffff;
  color: #145535;
  font-size: 22rpx;
  line-height: 58rpx;
}

.img-action-row {
  display: flex;
  align-items: center;
  gap: 10rpx;
}

.img-action-row .img-action {
  flex: 1;
}

.card-type-picker {
  display: flex;
  align-items: center;
  gap: 6rpx;
  min-height: 58rpx;
  padding: 0 16rpx;
  border: 1rpx solid #d8e5d6;
  border-radius: 12rpx;
  background: #ffffff;
  color: #2d4633;
  font-size: 22rpx;
  line-height: 58rpx;
}

.card-type-picker .caret {
  width: 0;
  height: 0;
  border-left: 6rpx solid transparent;
  border-right: 6rpx solid transparent;
  border-top: 8rpx solid #6d776c;
}

.entry-btn {
  margin-top: 24rpx;
  width: 100%;
  min-height: 84rpx;
  border-radius: 18rpx;
  background: linear-gradient(135deg, #237a4b, #145535);
  color: #ffffff;
  font-size: 28rpx;
  font-weight: 800;
  border: 0;
  line-height: 84rpx;
}

/* 汇总表格 */
.summary-card {
  padding: 0;
  overflow: hidden;
}

.summary-head,
.summary-row {
  display: grid;
  grid-template-columns: 1.2fr 1fr 1fr 1.1fr;
  gap: 8rpx;
  padding: 18rpx 20rpx;
}

.summary-head {
  background: #f6f8f4;
}

.summary-head text {
  color: #667266;
  font-size: 22rpx;
  font-weight: 760;
}

.summary-row {
  border-top: 1rpx solid #eef2ea;
}

.summary-row text {
  color: #384338;
  font-size: 22rpx;
  line-height: 1.4;
  overflow-wrap: anywhere;
  min-width: 0;
}

/* 收粮明细 */
.entries-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.entry-card {
  border-radius: 20rpx;
  background: #ffffff;
  padding: 24rpx;
  box-shadow: 0 2rpx 12rpx rgba(20, 85, 53, 0.05);
  active-background: #f4f8f3;
}

.entry-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 16rpx;
}

.entry-title {
  display: block;
  color: #172018;
  font-size: 28rpx;
  font-weight: 760;
}

.entry-place {
  display: block;
  margin-top: 6rpx;
  color: #6d776c;
  font-size: 22rpx;
}

.entry-price-badge {
  padding: 8rpx 14rpx;
  border-radius: 999rpx;
  background: #e8f5ec;
  color: #145535;
  font-size: 22rpx;
  font-weight: 700;
  white-space: nowrap;
}

.entry-delete-btn {
  flex: 0 0 auto;
  min-width: 88rpx;
  min-height: 52rpx;
  padding: 0 18rpx;
  border: 1rpx solid #f0c7bf;
  border-radius: 12rpx;
  background: #fff7f5;
  color: #c2412d;
  font-size: 22rpx;
  font-weight: 760;
  line-height: 52rpx;
}

.entry-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12rpx;
}

.entry-kv {
  padding: 14rpx 16rpx;
  border-radius: 12rpx;
  background: #f8faf6;
}

.entry-kv text {
  display: block;
  font-size: 22rpx;
}

.entry-kv text:first-child {
  color: #6d776c;
  margin-bottom: 6rpx;
}

.entry-kv text:last-child {
  color: #172018;
  font-weight: 700;
}

.loading-text {
  padding: 20rpx;
  text-align: center;
  color: #7b857b;
  font-size: 24rpx;
  background: #f8faf6;
  border-radius: 12rpx;
}

.loading-more {
  margin-top: 8rpx;
}

.load-more {
  padding: 28rpx;
  text-align: center;
  color: #237a4b;
  font-size: 26rpx;
  font-weight: 700;
}

.no-more {
  padding: 28rpx;
  text-align: center;
  color: #9aaa96;
  font-size: 24rpx;
}

.empty-text {
  padding: 36rpx;
  text-align: center;
  color: #9aaa96;
  font-size: 24rpx;
}

/* 编辑表单 */
.field {
  margin-bottom: 24rpx;
}

.label {
  display: block;
  margin-bottom: 12rpx;
  color: #445044;
  font-size: 26rpx;
  font-weight: 700;
}

.input,
.textarea {
  width: 100%;
  border: 1rpx solid #e2e8dd;
  border-radius: 16rpx;
  background: #fbfcfa;
  color: #172018;
  font-size: 28rpx;
}

.input {
  height: 88rpx;
  padding: 0 24rpx;
}

.textarea {
  min-height: 140rpx;
  padding: 20rpx 24rpx;
}

.btn-row {
  display: flex;
  gap: 16rpx;
  margin-top: 8rpx;
}

.primary-btn,
.secondary-btn {
  flex: 1;
  min-height: 84rpx;
  border-radius: 18rpx;
  font-size: 28rpx;
  font-weight: 800;
  line-height: 84rpx;
}

.primary-btn {
  border: 0;
  background: linear-gradient(135deg, #237a4b, #145535);
  color: #ffffff;
}

.secondary-btn {
  border: 1rpx solid #e2e8dd;
  background: #ffffff;
  color: #384338;
}

.icon-edit {
  display: inline-block;
  width: 22rpx;
  height: 8rpx;
  border-radius: 999rpx;
  background: currentColor;
  transform: rotate(-35deg);
  position: relative;
}
</style>
