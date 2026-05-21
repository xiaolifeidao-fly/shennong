<template>
  <view class="notice">
    <text class="icon">提</text>
    <text>当前先模拟拍照识别：点击“模拟识别”会自动填入身份证、银行卡相关信息。后续可接微信小程序相机、OCR 或第三方识别能力。</text>
  </view>
</template>

<style lang="scss" scoped>
.notice {
  display: flex;
  gap: 20rpx;
  align-items: flex-start;
  padding: 24rpx;
  border: 1rpx solid #f4e1bd;
  border-radius: 8rpx;
  background: #fffaf0;
  color: #56614f;
  font-size: 24rpx;
  line-height: 1.5;
}

.icon {
  display: flex;
  width: 54rpx;
  height: 54rpx;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 8rpx;
  background: #e8f5ec;
  color: #145535;
  font-weight: 800;
}
</style>
