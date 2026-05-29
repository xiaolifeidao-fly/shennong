<template>
  <view class="notice">
    <text class="icon">提</text>
    <text>拍身份证或选择银行卡付款时会先打开摄像头，拍照后上传到服务端识别；支付宝和微信付款请手动填写收款人姓名与收款账号。</text>
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
