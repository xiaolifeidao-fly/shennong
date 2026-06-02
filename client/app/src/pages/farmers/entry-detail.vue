<template>
  <view class="page entry-detail-page">
    <view v-if="!entry" class="empty-state">收粮明细不存在</view>
    <template v-else>
      <!-- 基本信息卡片 -->
      <view class="section-card">
        <view class="card-header">
          <view>
            <text class="entry-title">{{ entry.crop }}</text>
            <text class="entry-time">{{ entry.buyTime }}</text>
          </view>
          <view class="price-badge">{{ getEntryPrice(entry) }}</view>
        </view>

        <template v-if="!editing">
          <view class="info-grid">
            <view class="kv">
              <text class="kv-label">数量</text>
              <text class="kv-value">{{ formatQuantityByDisplayUnit(entry.quantity, entry.displayUnit || entry.unit) }}</text>
            </view>
            <view class="kv">
              <text class="kv-label">金额</text>
              <text class="kv-value">{{ formatAmount(entry.amount) }}</text>
            </view>
            <view class="kv">
              <text class="kv-label">付款方式</text>
              <text class="kv-value">{{ entry.payType || '-' }}</text>
            </view>
          </view>
          <view class="info-grid" style="margin-top: 16rpx;">
            <view class="kv">
              <text class="kv-label">收购地点</text>
              <text class="kv-value">{{ entry.place || '-' }}</text>
            </view>
            <view class="kv">
              <text class="kv-label">位置名称</text>
              <text class="kv-value">{{ entry.locationName || '-' }}</text>
            </view>
          </view>
          <view class="kv kv-full" style="margin-top: 16rpx;">
            <text class="kv-label">详细地址</text>
            <text class="kv-value">{{ entry.locationAddress || '-' }}</text>
          </view>

          <!-- 材料图片 -->
          <view v-if="entry.materialImages.length" class="img-section">
            <text class="img-section-title">附件图片（{{ entry.materialImages.length }} 张）</text>
            <view class="img-row">
              <image
                v-for="(img, idx) in entry.materialImages"
                :key="idx"
                :src="img"
                class="material-img"
                mode="aspectFill"
                @click="previewImage(img, entry.materialImages)"
              />
            </view>
          </view>

          <view class="detail-actions">
            <button class="edit-btn-full" @click="startEdit">编辑明细</button>
            <button class="continue-btn-full" @click="continueEntry">继续录入</button>
          </view>
        </template>

        <!-- 编辑表单 -->
        <template v-else>
          <view class="field">
            <text class="label">粮食类型</text>
            <input v-model="editForm.crop" class="input" placeholder="请输入粮食类型" />
          </view>
          <view class="field">
            <text class="label">数量</text>
            <input v-model.number="editForm.quantity" class="input" type="number" placeholder="请输入数量" />
          </view>
          <view class="field">
            <text class="label">单位</text>
            <picker :value="editUnitIndex" :range="unitOptions" @change="selectEditUnit">
              <view class="input" style="line-height: 88rpx;">{{ editForm.unit || '公斤' }}</view>
            </picker>
          </view>
          <view class="field">
            <text class="label">金额（元）</text>
            <input v-model.number="editForm.amount" class="input" type="number" placeholder="请输入金额" />
          </view>
          <view class="field">
            <text class="label">收购时间</text>
            <input v-model="editForm.buyTime" class="input" placeholder="YYYY-MM-DD HH:mm" />
          </view>
          <view class="field">
            <text class="label">收购地点</text>
            <input v-model="editForm.place" class="input" placeholder="请输入收购地点" />
          </view>
          <view class="field">
            <text class="label">付款方式</text>
            <input v-model="editForm.payType" class="input" placeholder="请输入付款方式" />
          </view>
          <view class="btn-row">
            <button class="secondary-btn" @click="cancelEdit">取消</button>
            <button class="primary-btn" :disabled="saving" @click="saveEntry">{{ saving ? '保存中...' : '保存' }}</button>
          </view>
        </template>
      </view>

      <!-- 农户信息 -->
      <view v-if="farmer" class="section-card">
        <text class="section-sub-title">农户信息</text>
        <view class="info-grid">
          <view class="kv">
            <text class="kv-label">姓名</text>
            <text class="kv-value">{{ farmer.name }}</text>
          </view>
          <view class="kv">
            <text class="kv-label">电话</text>
            <text class="kv-value">{{ farmer.phone || '-' }}</text>
          </view>
          <view class="kv">
            <text class="kv-label">身份证号</text>
            <text class="kv-value">{{ farmer.idNumber || '-' }}</text>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useGrainStore } from '@/stores/grain'
import { formatAmount, formatQuantityByDisplayUnit, getEntryPrice } from '@/utils/grain'
import type { GrainEntryDraft } from '@/types/grain'

const grainStore = useGrainStore()

const unitOptions = ['公斤', '吨']
const editUnitIndex = computed(() => Math.max(0, unitOptions.indexOf(editForm.unit || '公斤')))

function selectEditUnit(event: { detail: { value: number | string } }) {
  editForm.unit = unitOptions[Number(event.detail.value)] || '公斤'
}

const entry = computed(() => grainStore.selectedEntry)
const farmer = computed(() =>
  grainStore.farmers.find((f) => f.id === entry.value?.farmerId) ||
  grainStore.farmerSummaries.find((f) => f.id === entry.value?.farmerId)
)

const editing = ref(false)
const saving = ref(false)

const editForm = reactive<Partial<GrainEntryDraft>>({
  crop: '',
  quantity: 0,
  unit: '公斤',
  amount: 0,
  buyTime: '',
  place: '',
  payType: '',
})

onLoad((options) => {
  const entryId = String(options?.entryId || '')
  if (entryId) {
    grainStore.selectEntry(entryId)
  }
  void ensureEntryLoaded(entryId)
})

watch(entry, (e) => {
  if (e) {
    editForm.crop = e.crop
    editForm.quantity = e.quantity
    editForm.unit = e.unit
    editForm.amount = e.amount
    editForm.buyTime = e.buyTime
    editForm.place = e.place
    editForm.payType = e.payType
  }
}, { immediate: true })

async function ensureEntryLoaded(entryId: string) {
  await Promise.all([grainStore.loadFarmers(), grainStore.loadEntries()])
  if (entryId && !grainStore.selectedEntry) {
    grainStore.selectEntry(entryId)
  }
}

function startEdit() {
  editing.value = true
}

function cancelEdit() {
  if (entry.value) {
    editForm.crop = entry.value.crop
    editForm.quantity = entry.value.quantity
    editForm.unit = entry.value.unit
    editForm.amount = entry.value.amount
    editForm.buyTime = entry.value.buyTime
    editForm.place = entry.value.place
    editForm.payType = entry.value.payType
  }
  editing.value = false
}

async function saveEntry() {
  if (!entry.value || !farmer.value) return
  saving.value = true
  try {
    const draft = grainStore.createEntryDraftFromHistory(entry.value.id)
    // 将展示单位下的数量赋给 draft（buildEntryPayload 会在保存时统一转换为公斤）
    Object.assign(draft, {
      crop: editForm.crop,
      quantity: Number(editForm.quantity),
      unit: editForm.unit || '公斤',
      amount: Number(editForm.amount),
      buyTime: editForm.buyTime,
      place: editForm.place,
      payType: editForm.payType,
    })
    await grainStore.updateEntry(entry.value.id, draft)
    editing.value = false
    uni.showToast({ title: '已保存', icon: 'success' })
  } catch (error) {
    uni.showToast({ title: error instanceof Error ? error.message : '保存失败', icon: 'none' })
  } finally {
    saving.value = false
  }
}

function previewImage(current: string, urls: string[]) {
  uni.previewImage({ current, urls })
}

function continueEntry() {
  if (entry.value?.farmerId) {
    grainStore.selectFarmer(entry.value.farmerId)
  }
  uni.switchTab({ url: '/pages/entry/index' })
}
</script>

<style lang="scss" scoped>
.entry-detail-page {
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

.section-card {
  border-radius: 24rpx;
  background: #ffffff;
  padding: 28rpx;
  box-shadow: 0 4rpx 16rpx rgba(20, 85, 53, 0.06);
}

.section-sub-title {
  display: block;
  color: #6d776c;
  font-size: 24rpx;
  font-weight: 700;
  margin-bottom: 18rpx;
}

.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 24rpx;
}

.entry-title {
  display: block;
  color: #172018;
  font-size: 34rpx;
  font-weight: 760;
}

.entry-time {
  display: block;
  margin-top: 8rpx;
  color: #6d776c;
  font-size: 24rpx;
}

.price-badge {
  padding: 10rpx 16rpx;
  border-radius: 999rpx;
  background: #e8f5ec;
  color: #145535;
  font-size: 22rpx;
  font-weight: 760;
  white-space: nowrap;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16rpx;
}

.kv {
  padding: 18rpx 20rpx;
  border-radius: 16rpx;
  background: #f8faf6;
}

.kv-full {
  padding: 18rpx 20rpx;
  border-radius: 16rpx;
  background: #f8faf6;
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
  font-size: 24rpx;
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
  display: flex;
  flex-wrap: wrap;
  gap: 14rpx;
}

.material-img {
  width: 180rpx;
  height: 180rpx;
  border-radius: 12rpx;
  border: 1rpx solid #e2e8dd;
}

.detail-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16rpx;
  margin-top: 24rpx;
}

.edit-btn-full,
.continue-btn-full {
  margin-top: 24rpx;
  width: 100%;
  min-height: 84rpx;
  border-radius: 18rpx;
  font-size: 28rpx;
  font-weight: 800;
  line-height: 84rpx;
}

.detail-actions .edit-btn-full,
.detail-actions .continue-btn-full {
  margin-top: 0;
}

.edit-btn-full {
  border: 1rpx solid #e2e8dd;
  background: #ffffff;
  color: #384338;
}

.continue-btn-full {
  border: 0;
  background: linear-gradient(135deg, #237a4b, #145535);
  color: #ffffff;
}

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

.input {
  width: 100%;
  height: 88rpx;
  padding: 0 24rpx;
  border: 1rpx solid #e2e8dd;
  border-radius: 16rpx;
  background: #fbfcfa;
  color: #172018;
  font-size: 28rpx;
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
</style>
