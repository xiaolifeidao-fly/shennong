---
name: app-development
description: 面向本仓库 client/app 微信小程序端的开发技能。适用于 uni-app + Vue 3 + TypeScript 小程序初始化、页面开发、组件拆分、Pinia 状态、uni.request 请求封装、对接 server/app-api、微信开发者工具调试和小程序工程规范维护。
license: MIT
version: 1.0.0
---

# App Development Skill

本技能用于开发 `client/app` 微信小程序端。当前小程序端技术栈是 **uni-app + Vue 3 + TypeScript + Pinia + wot-design-uni**，服务端对接 `server/app-api`。

## 何时使用

- 初始化或调整 `client/app` 小程序工程
- 新增/修改小程序页面、业务组件、基础组件
- 对接 `server/app-api` 接口、封装请求、处理 token 和登录态
- 设计页面目录、组件目录、服务层、状态层、类型定义
- 调整 `pages.json`、`manifest.json`、微信开发者工具配置
- 排查 `npm run dev:mp-weixin`、`npm run build:mp-weixin`、真机/模拟器调试问题

## 当前项目边界

```text
client/manager  -> 管理端 Web，Next.js + React + Ant Design -> server/manager-api
client/app      -> 微信小程序，uni-app + Vue 3 + TS          -> server/app-api
server/service  -> 共享业务服务
```

不要把管理端的 Next.js/AntD 结构复制到小程序。小程序端保持轻量、移动优先、页面清晰、服务层集中。

## 必读参考

- [app-architecture.md](references/app-architecture.md) — 目录结构、页面/组件/API/状态分层、请求封装、抽象粒度、调试流程

## 快速工作流

### 新增页面

1. 在 `src/pages/{page}/index.vue` 新建页面
2. 在 `src/pages.json` 注册页面路径和标题
3. 页面只负责 UI 编排、事件入口、调用 store/service
4. 页面内超过 120 行或出现可复用块时，拆到同级 `components/` 或全局 `src/components/`
5. 跑 `npm run type-check` 和 `npm run build:mp-weixin`

### 新增接口调用

1. 后端接口属于 `server/app-api`，先确认路由路径和响应结构
2. 在 `src/types/` 定义请求/响应类型
3. 在 `src/services/{domain}.ts` 中封装接口函数
4. 页面或 store 只能调用 `services`，不要直接调用 `uni.request`
5. token、错误处理、loading 统一走 `src/services/request.ts`

### 新增状态

1. 跨页面共享状态放 `src/stores/{domain}.ts`
2. 单页面临时状态留在页面 `ref/reactive`
3. 服务端数据优先通过 service 获取，store 负责缓存、派生状态和跨页面行为
4. 不把表单每个字段都塞进 Pinia，除非有跨页面草稿需求

### 新增组件

按复用范围选择位置：

```text
src/pages/{page}/components/   只服务当前页面或当前业务流程
src/components/                跨页面复用的业务组件/基础组件
src/components/base/           二次封装的低层基础组件
src/components/business/       跨页面复用的业务组件
```

组件只接收明确 props、触发明确 events。不要在低层基础组件里直接请求接口或读写全局 store。

## 强制约定

- `src/pages/*/index.vue` 是页面入口，避免在页面里堆大量业务细节
- 所有 HTTP 请求必须经过 `src/services/request.ts`
- 所有接口函数必须放在 `src/services/{domain}.ts`
- 所有服务端响应类型放在 `src/types/`，不要在页面内临时写散装类型
- token 持久化只通过 `src/utils/token.ts`
- 登录态和用户信息默认走 `src/stores/user.ts`
- 页面跳转使用 `uni.navigateTo` / `uni.switchTab` / `uni.redirectTo`，不要手写路径拼接工具，除非路由增长到需要统一管理
- 用户可见错误使用 `uni.showToast({ icon: 'none' })` 或组件库反馈，不静默吞异常
- 新增 tab 页必须同步改 `pages.json` 的 `tabBar.list`
- 真实 appid 未配置前，`manifest.json` / `project.config.json` 保持占位，不要硬编码私人 appid

## 验证命令

在 `client/app` 下执行：

```bash
npm run type-check
npm run build:mp-weixin
```

开发预览：

```bash
npm run dev:mp-weixin
```

微信开发者工具导入：

```text
client/app/dist/dev/mp-weixin
```

构建产物在 `client/app/dist/`，该目录不应提交。

## 常见判断

- 只做一个页面用的 UI 块：放页面同级 `components/`
- 两个以上页面复用：提升到 `src/components/business/`
- 纯展示、无业务语义、可配置性强：放 `src/components/base/`
- 接口返回数据需要跨页面共享：service + Pinia store
- 接口返回数据只服务当前页面：页面调用 service，局部 state 即可
- 微信能力如登录、授权、支付、订阅消息：先封装到 `src/services/wechat.ts` 或领域 service，不在页面散写
