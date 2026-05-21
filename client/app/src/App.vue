<script setup lang="ts">
const appReady = true

void appReady
</script>

<style lang="scss">
page {
  min-height: 100%;
  background: #f5f7f3;
  color: #172018;
  font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', Helvetica, Arial, sans-serif;
}

view,
text,
button,
input,
textarea {
  box-sizing: border-box;
}

button {
  margin: 0;
}

.page {
  min-height: 100vh;
  padding: 32rpx 28rpx calc(120rpx + env(safe-area-inset-bottom));
  background: linear-gradient(180deg, rgba(35, 122, 75, 0.12), rgba(245, 247, 243, 0) 320rpx), #f5f7f3;
}

.section {
  margin-top: 24rpx;
}

.card {
  border: 1rpx solid rgba(226, 232, 221, 0.9);
  border-radius: 8rpx;
  background: #ffffff;
  box-shadow: 0 8rpx 22rpx rgba(31, 47, 31, 0.045);
}
</style>
