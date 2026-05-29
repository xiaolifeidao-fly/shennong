<template>
  <view class="notice">
    <text class="icon notice-icon"></text>
    <text>拍身份证或选择银行卡付款时会先打开摄像头，拍照后上传到服务端识别；支付宝和微信付款请手动填写收款人姓名与收款账号。</text>
  </view>
</template>

<style lang="scss" scoped>
.notice {
  display: flex;
  gap: 20rpx;
  align-items: flex-start;
  padding: 24rpx;
  border: 1rpx solid rgba(255, 184, 77, 0.32);
  border-radius: 24rpx;
  background: linear-gradient(135deg, #fffaf0, #f5fbef);
  color: #56614f;
  font-size: 24rpx;
  line-height: 1.5;
}

.icon {
  position: relative;
  display: flex;
  width: 54rpx;
  height: 54rpx;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 18rpx;
  background: #ffffff;
  box-shadow: inset 0 0 0 1rpx rgba(35, 122, 75, 0.12);
}

.notice-icon::before {
  width: 8rpx;
  height: 28rpx;
  border-radius: 999rpx;
  background: #237a4b;
  content: '';
}

.notice-icon::after {
  position: absolute;
  left: 23rpx;
  bottom: 11rpx;
  width: 8rpx;
  height: 8rpx;
  border-radius: 50%;
  background: #ffb84d;
  content: '';
}
</style>
