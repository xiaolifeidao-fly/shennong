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
            <view class="kv">
              <text class="kv-label">支付时间</text>
              <text class="kv-value">{{ entry.payTime || '未填写' }}</text>
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
                :key="img"
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
            <text class="label required">收购粮食类型</text>
            <input
              v-model="editCropKeyword"
              class="input"
              placeholder="输入关键词搜索，请从下方选择"
              @input="handleEditCropInput"
              @focus="showEditCropResults = true"
            />
            <view class="crop-helper">仅可选择粮站已维护的粮食类型，输入内容不会直接保存。</view>
            <view class="crop-options">
              <button
                v-for="crop in filteredEditCropOptions"
                :key="crop.id"
                class="crop-chip"
                :class="{ active: crop.id === editForm.purchaseTypeId }"
                @click="selectEditCrop(crop)"
              >
                {{ crop.name }}
              </button>
              <view v-if="!filteredEditCropOptions.length && editCropKeyword.trim()" class="crop-empty">未找到匹配粮食类型，请从已有类型中选择或联系管理员维护。</view>
              <view v-else-if="!grainStore.preset.purchaseTypes.length" class="crop-empty">当前粮站暂无可选粮食类型，请先联系管理员维护。</view>
            </view>
          </view>
          <view class="field">
            <text class="label">数量</text>
            <input v-model="editForm.quantity" class="input" type="digit" placeholder="请输入数量" />
          </view>
          <view class="field">
            <text class="label">单位</text>
            <picker :value="editUnitIndex" :range="unitOptions" @change="selectEditUnit">
              <view class="input" style="line-height: 88rpx;">{{ editForm.unit || '公斤' }}</view>
            </picker>
          </view>
          <view class="field">
            <text class="label">金额（元）</text>
            <input v-model="editForm.amount" class="input" type="digit" placeholder="请输入金额" />
          </view>
          <view class="field">
            <text class="label">收购时间</text>
            <view class="datetime-row">
              <picker mode="date" :value="datePart(editForm.buyTime) || todayDate" @change="setDateTimePart('buyTime', 'date', $event)">
                <view class="datetime-value">{{ datePart(editForm.buyTime) || '请选择日期' }}</view>
              </picker>
              <picker mode="time" :value="timePart(editForm.buyTime) || '00:00'" @change="setDateTimePart('buyTime', 'time', $event)">
                <view class="datetime-value">{{ timePart(editForm.buyTime) || '请选择时间' }}</view>
              </picker>
              <button v-if="editForm.buyTime" class="clear-time-btn" @click="clearDateTime('buyTime')">清空</button>
            </view>
          </view>
          <view class="field">
            <text class="label">收购地点</text>
            <input v-model="editForm.place" class="input" placeholder="请输入收购地点" />
          </view>
          <view class="field">
            <text class="label">付款方式</text>
            <input v-model="editForm.payType" class="input" placeholder="请输入付款方式（可不填）" />
          </view>
          <view class="field">
            <text class="label">支付时间</text>
            <view class="datetime-row">
              <picker mode="date" :value="datePart(editForm.payTime) || todayDate" @change="setDateTimePart('payTime', 'date', $event)">
                <view class="datetime-value">{{ datePart(editForm.payTime) || '请选择日期' }}</view>
              </picker>
              <picker mode="time" :value="timePart(editForm.payTime) || '00:00'" @change="setDateTimePart('payTime', 'time', $event)">
                <view class="datetime-value">{{ timePart(editForm.payTime) || '请选择时间' }}</view>
              </picker>
              <button v-if="editForm.payTime" class="clear-time-btn" @click="clearDateTime('payTime')">清空</button>
            </view>
          </view>

          <view class="field">
            <view class="label-row">
              <text class="label">附件图片</text>
              <button class="upload-btn" @click="addMaterials">追加上传</button>
            </view>
            <view class="img-row">
              <view
                v-for="(img, idx) in editForm.materialImages"
                :key="img + idx"
                class="img-cell"
              >
                <image
                  :key="img"
                  :src="img"
                  class="material-img"
                  mode="aspectFill"
                  @click="previewImage(img, editForm.materialImages || [])"
                />
                <text class="img-remove" @click="removeMaterial(idx)">×</text>
              </view>
              <button v-if="!(editForm.materialImages && editForm.materialImages.length)" class="empty-upload" @click="addMaterials">点击上传补充资料</button>
            </view>
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
            <text class="kv-value">{{ maskPhone(farmer.phone) || '-' }}</text>
          </view>
          <view class="kv">
            <text class="kv-label">身份证号</text>
            <text class="kv-value">{{ maskIdNumber(farmer.idNumber) || '-' }}</text>
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
import { maskIdNumber, maskPhone } from '@/utils/privacy'
import type { GrainEntryDraft, GrainPurchaseType } from '@/types/grain'

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
const todayDate = new Date().toISOString().slice(0, 10)
const editCropKeyword = ref('')
const showEditCropResults = ref(false)

const editForm = reactive<Partial<GrainEntryDraft>>({
  purchaseTypeId: 0,
  crop: '',
  quantity: 0,
  unit: '公斤',
  amount: 0,
  buyTime: '',
  payTime: '',
  place: '',
  payType: '',
  materialImages: [],
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
    editForm.purchaseTypeId = e.purchaseTypeId
    editForm.crop = e.crop
    editCropKeyword.value = e.crop
    editForm.quantity = e.quantity
    editForm.unit = e.unit
    editForm.amount = e.amount
    editForm.buyTime = e.buyTime
    editForm.payTime = e.payTime || ''
    editForm.place = e.place
    editForm.payType = e.payType
    editForm.materialImages = [...(e.materialImages || [])]
  }
}, { immediate: true })

async function ensureEntryLoaded(entryId: string) {
  await Promise.all([grainStore.loadPreset(), grainStore.loadFarmers(), grainStore.loadEntries()])
  if (entryId && !grainStore.selectedEntry) {
    grainStore.selectEntry(entryId)
  }
  const targetEntryId = entryId || grainStore.selectedEntryId
  if (targetEntryId) {
    await grainStore.loadEntryMaterials(targetEntryId)
  }
}

function startEdit() {
  editing.value = true
}

function cancelEdit() {
  if (entry.value) {
    editForm.purchaseTypeId = entry.value.purchaseTypeId
    editForm.crop = entry.value.crop
    editCropKeyword.value = entry.value.crop
    editForm.quantity = entry.value.quantity
    editForm.unit = entry.value.unit
    editForm.amount = entry.value.amount
    editForm.buyTime = entry.value.buyTime
    editForm.payTime = entry.value.payTime || ''
    editForm.place = entry.value.place
    editForm.payType = entry.value.payType
    editForm.materialImages = [...(entry.value.materialImages || [])]
  }
  editing.value = false
}

const filteredEditCropOptions = computed(() => {
  const key = editCropKeyword.value.trim()
  if (!showEditCropResults.value && editForm.crop) {
    return []
  }
  return grainStore.preset.purchaseTypes
    .filter((item) => !key || item.typeName.includes(key))
    .slice(0, 8)
    .map((item) => ({ id: item.id, name: item.typeName, unit: item.unit }))
})

function selectEditCrop(crop: Pick<GrainPurchaseType, 'id' | 'unit'> & { name: string }) {
  editForm.purchaseTypeId = crop.id
  editForm.crop = crop.name
  editForm.unit = crop.unit || editForm.unit || '公斤'
  editCropKeyword.value = crop.name
  showEditCropResults.value = false
}

function handleEditCropInput(event: unknown) {
  const value =
    typeof event === 'object' && event && 'detail' in event
      ? String((event as { detail?: { value?: string } }).detail?.value ?? editCropKeyword.value)
      : editCropKeyword.value
  editCropKeyword.value = value
  editForm.purchaseTypeId = 0
  editForm.crop = ''
  showEditCropResults.value = true
}

function addMaterials() {
  uni.chooseImage({
    count: 6,
    success: (res) => {
      const next = [...(editForm.materialImages || []), ...res.tempFilePaths].slice(0, 9)
      editForm.materialImages = next
    },
  })
}

function removeMaterial(index: number) {
  const list = [...(editForm.materialImages || [])]
  list.splice(index, 1)
  editForm.materialImages = list
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
  const nextDate = part === 'date' ? value : datePart(editForm[field]) || todayDate
  const nextTime = part === 'time' ? value : timePart(editForm[field]) || '00:00'
  editForm[field] = `${nextDate} ${nextTime}`
}

function clearDateTime(field: 'buyTime' | 'payTime') {
  editForm[field] = ''
}

async function saveEntry() {
  if (!entry.value || !farmer.value) return
  const validateMessage = validateEditForm()
  if (validateMessage) {
    uni.showToast({ title: validateMessage, icon: 'none' })
    return
  }
  saving.value = true
  try {
    const draft = grainStore.createEntryDraftFromHistory(entry.value.id)
    // 将展示单位下的数量赋给 draft（buildEntryPayload 会在保存时统一转换为公斤）
    Object.assign(draft, {
      purchaseTypeId: editForm.purchaseTypeId,
      crop: editForm.crop,
      quantity: Number(editForm.quantity),
      unit: editForm.unit || '公斤',
      amount: Number(editForm.amount),
      buyTime: editForm.buyTime,
      payTime: editForm.payTime || '',
      place: editForm.place,
      payType: editForm.payType,
      materialImages: [...(editForm.materialImages || [])],
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

function validateEditForm() {
  if (!grainStore.preset.purchaseTypes.length) {
    return '当前粮站暂无粮食类型，请先联系管理员维护'
  }
  const selectedPurchaseType = grainStore.preset.purchaseTypes.find(
    (item) => item.id === Number(editForm.purchaseTypeId) && item.typeName === editForm.crop,
  )
  if (!selectedPurchaseType) {
    return '收购粮食类型为必填项，请从已有粮食类型中选择'
  }
  if (Number(editForm.quantity) <= 0) {
    return '请输入有效数量'
  }
  if (Number(editForm.amount) <= 0) {
    return '请输入有效金额'
  }
  return ''
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

.required::after {
  content: ' *';
  color: #d14343;
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

.crop-helper {
  margin-top: 10rpx;
  color: #7b857b;
  font-size: 23rpx;
  line-height: 1.45;
}

.crop-options {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-top: 14rpx;
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
  border-radius: 16rpx;
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
  border-radius: 16rpx;
  background: #ffffff;
  color: #6d776c;
  font-size: 24rpx;
  line-height: 88rpx;
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

.label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 12rpx;
}

.label-row .label {
  margin-bottom: 0;
}

.upload-btn {
  padding: 8rpx 22rpx;
  border: 1rpx solid #cfe0d1;
  border-radius: 999rpx;
  background: #ffffff;
  color: #145535;
  font-size: 24rpx;
  font-weight: 760;
}

.img-cell {
  position: relative;
  width: 180rpx;
  height: 180rpx;
}

.img-cell .material-img {
  width: 100%;
  height: 100%;
}

.img-remove {
  position: absolute;
  top: -10rpx;
  right: -10rpx;
  width: 36rpx;
  height: 36rpx;
  border-radius: 50%;
  background: rgba(20, 32, 24, 0.78);
  color: #ffffff;
  font-size: 26rpx;
  font-weight: 700;
  line-height: 36rpx;
  text-align: center;
}

.empty-upload {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  min-height: 148rpx;
  border: 1rpx dashed rgba(35, 122, 75, 0.36);
  border-radius: 16rpx;
  background: #f8faf6;
  color: #48604e;
  font-size: 24rpx;
}
</style>
