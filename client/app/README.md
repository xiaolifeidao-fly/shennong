# shennong-app

微信小程序端，使用 uni-app + Vue 3 + TypeScript。

## Scripts

```bash
npm install
npm run dev:mp-weixin
```

开发者工具导入当前目录，或导入 `dist/dev/mp-weixin` 预览编译后的微信小程序。

## API

默认本地接口地址：

```txt
http://localhost:8191/api
```

真机调试时需要把 `.env.development` 中的 `VITE_APP_API_BASE_URL` 改成可访问的局域网 IP 或 HTTPS 域名。
