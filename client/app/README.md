# shennong-app

微信小程序端，使用 uni-app + Vue 3 + TypeScript。

## Scripts

```bash
npm install
npm run dev:mp-weixin
```

开发者工具导入当前目录，或导入 `dist/dev/mp-weixin` 预览编译后的微信小程序。

## API

接口调用方式通过环境变量切换：

```txt
VITE_APP_API_TRANSPORT=cloud
VITE_APP_CLOUD_ENV=prod-d4ghi7vus72134bcc
VITE_APP_CLOUD_SERVICE=app-api-test002
VITE_APP_CLOUD_FILE_BASE_URL=https://your-cloud-container-domain.example.com
VITE_APP_API_BASE_URL=http://localhost:8191/api
```

`VITE_APP_API_TRANSPORT=cloud` 时，小程序通过 `wx.cloud.callContainer` 调用云托管服务，业务路径会从 `/login` 转成 `login`，不拼接 `/app/api` 前缀。

云开发模式下，文件上传和下载使用 `wx.uploadFile` / `wx.downloadFile`，请求地址由 `VITE_APP_CLOUD_FILE_BASE_URL` 拼接业务路径生成。这个地址应填写云托管 HTTPS 根地址，不带 `/app/api`。

`VITE_APP_API_TRANSPORT=https` 时，小程序通过 `uni.request` 调用 `VITE_APP_API_BASE_URL`。真机调试需要把 `.env.development` 中的 `VITE_APP_API_BASE_URL` 改成可访问的局域网 IP 或 HTTPS 域名。
