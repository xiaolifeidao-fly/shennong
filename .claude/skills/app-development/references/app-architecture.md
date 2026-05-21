# App Architecture

## 技术栈

```text
框架      uni-app
语法      Vue 3 + TypeScript
构建      Vite + @dcloudio/vite-plugin-uni
UI        wot-design-uni
状态      Pinia
请求      uni.request，经 src/services/request.ts 统一封装
目标端    微信小程序 mp-weixin
服务端    server/app-api，默认 /api 前缀
```

## 目录结构

推荐结构：

```text
client/app
  src/
    App.vue
    main.ts
    manifest.json
    pages.json
    pages/
      index/
        index.vue
        components/
      login/
        index.vue
      mine/
        index.vue
    components/
      base/
      business/
    services/
      request.ts
      auth.ts
      wechat.ts
      {domain}.ts
    stores/
      user.ts
      {domain}.ts
    types/
      api.ts
      {domain}.ts
    utils/
      token.ts
      format.ts
      guard.ts
    constants/
      route.ts
      storage.ts
  package.json
  project.config.json
  tsconfig.json
  vite.config.ts
```

当前已存在的基础结构可以逐步演进，不需要一次性创建所有目录。只有出现对应职责时才新增目录。

## 页面层规范

页面文件统一为：

```text
src/pages/{page}/index.vue
```

页面职责：

- 组织页面布局和交互流程
- 管理本页面临时状态
- 调用 `stores` 或 `services`
- 展示 loading、empty、error 等页面状态

页面不做：

- 不直接调用 `uni.request`
- 不写复杂数据转换
- 不堆多个业务区域的细节实现
- 不直接读写 token storage

页面超过以下信号时应拆分：

- 单文件超过约 180 行
- template 出现 3 个以上明显业务区域
- 同一 UI 块未来可能在别的页面复用
- 事件函数开始承载业务规则而非 UI 入口

## 组件分层

### 页面私有组件

位置：

```text
src/pages/{page}/components/
```

适合：

- 只属于某个页面的 header、filter、card、panel
- 对当前页面数据结构强绑定
- 复用价值暂不明确的 UI 块

### 全局基础组件

位置：

```text
src/components/base/
```

适合：

- `BaseEmpty`
- `BasePage`
- `BaseSection`
- `BaseUpload`
- `BaseSafeArea`

要求：

- 不直接请求接口
- 不直接依赖业务 store
- props/events 清晰
- 只封装跨页面一致的视觉和交互

### 全局业务组件

位置：

```text
src/components/business/
```

适合：

- `UserProfileCard`
- `OrderStatusTag`
- `ProductCard`
- `AddressPicker`

要求：

- 可以有业务语义
- 尽量不在组件内部主动拉取页面主数据
- 复杂业务动作通过事件交给页面或 store

## 服务端调用

所有 HTTP 请求走：

```text
src/services/request.ts
```

领域接口放：

```text
src/services/{domain}.ts
```

示例：

```ts
import { http } from './request'
import type { FooDetail, FooQuery } from '@/types/foo'

export function getFooDetail(id: number) {
  return http.get<FooDetail>(`/foos/${id}`)
}

export function listFoos(query: FooQuery) {
  return http.get<FooDetail[]>('/foos', { data: query })
}
```

如果当前 `http.get` 尚未支持 query 参数，优先扩展 `request.ts`，不要在业务 service 里拼接复杂 query 字符串。

### 响应格式

`server/app-api` 当前统一响应：

```json
{
  "success": true,
  "code": 0,
  "data": {},
  "message": "请求成功",
  "error": null
}
```

前端类型放在：

```text
src/types/api.ts
```

请求封装需要处理：

- `success === false` 时抛错
- token 自动带入 header：`token` 和 `Authorization: Bearer ...`
- 未登录/登录过期时清 token
- loading 可由调用方传 `showLoading`

### 环境地址

本地默认：

```text
VITE_APP_API_BASE_URL=http://localhost:8191/api
```

真机调试时不能依赖小程序里的 `localhost` 访问电脑服务，应改为电脑局域网 IP 或 HTTPS 域名。

## 登录与微信能力

当前基础登录是账号密码接口：

```text
POST /login
GET  /auth-state
POST /logout
GET  /app-user-profile
```

后续接微信登录时建议新增：

```text
src/services/wechat.ts
```

负责封装：

- `uni.login`
- `uni.getUserProfile`
- 授权状态检查
- 微信 code 换取 app-api token
- 支付、订阅消息等微信能力入口

页面只调用 `wechat.ts` 或 `auth.ts` 暴露的方法。

## Pinia 状态边界

放入 Pinia：

- token 派生出的登录态
- 当前用户信息
- 多页面共享的业务数据
- 需要跨页面保持的筛选条件或草稿

不放入 Pinia：

- 单个输入框状态
- 单页面 loading
- 只在当前页面展示一次的接口数据
- 组件内部展开/折叠状态

store 命名：

```text
src/stores/user.ts       -> useUserStore
src/stores/order.ts      -> useOrderStore
src/stores/product.ts    -> useProductStore
```

store 可以调用 service；service 不要 import store，避免依赖反转。

## 类型文件

通用响应类型：

```text
src/types/api.ts
```

领域类型：

```text
src/types/{domain}.ts
```

优先使用 `interface` 描述服务端 DTO，避免在页面中内联类型。字段不确定时可以局部使用 `unknown`，不要扩大成 `any`。

## 样式规范

- 页面使用 `rpx`
- 全局基础样式放 `App.vue`
- 页面样式写在对应 `.vue` 的 scoped style 中
- 颜色、间距、圆角如果开始重复，抽到 `src/styles/variables.scss`
- 避免过度设计，业务小程序优先清晰、可点击区域足够、状态明确

## 路由与配置

新增页面必须改：

```text
src/pages.json
```

新增 tab 页同时改：

```text
tabBar.list
```

`manifest.json` 保存小程序名称、appid、平台配置。真实 appid 由用户提供后再填入，不要猜。

## 调试流程

开发：

```bash
cd client/app
npm run dev:mp-weixin
```

微信开发者工具导入：

```text
client/app/dist/dev/mp-weixin
```

构建检查：

```bash
npm run type-check
npm run build:mp-weixin
```

常见情况：

- 模拟器看效果：不需要上传
- 手机预览：开发者工具点“预览”扫码
- 体验版：开发者工具点“上传”，给体验成员测试
- 发布：上传后到微信公众平台提交审核

## 抽象粒度原则

先页面内实现，出现真实重复再抽象。不要为了“看起来工程化”提前制造 service/hook/component 层级。

推荐判断：

- 重复 2 次：考虑抽成函数或组件
- 重复 3 次：必须抽
- 涉及 token、错误、响应格式：一开始就统一封装
- 涉及微信原生能力：一开始就统一封装
- 涉及业务规则：优先放 service/store，不放基础组件

## 与后端协作

对接新接口时同步检查：

- `server/app-api/routers/register.go` 是否注册 handler
- handler 路由是否在 `/api` 前缀下
- 响应是否使用 `common/middleware/routers.ToJson`
- 前端 service 的路径不要重复写 `/api`
- app-api 错误现在通常 HTTP 200 + body success=false，前端必须看 body
