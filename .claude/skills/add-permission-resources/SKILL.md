---
name: add-permission-resources
description: 当用户要求根据本次上下文新增接口或新增页面生成管理端权限 SQL 时使用。用户只需要提供 role_id，本技能输出插入 resource_new 与 role_resource_new 的幂等 SQL，适用于本仓库 manager-api 权限资源补齐。
---

# 新增权限资源

## 目标

根据当前对话或当前代码改动中新增的管理端接口/页面，输出可直接执行的权限资源 SQL。

用户通常只会给出 `role_id`，例如：

```text
role_id = 1
```

你需要从上下文识别新增接口或页面，生成：

- `resource_new` 插入 SQL
- `role_resource_new` 绑定 SQL

只输出 SQL 和必要的简短说明；不要修改代码，除非用户明确要求落库或写文件。

---

## 资源类型一览

| resource_type | 含义         | parent_id 规则                       |
|---------------|--------------|--------------------------------------|
| `menu`        | 菜单分组     | 固定为 `0`（顶层节点）               |
| `page`        | 叶子页面     | 必须填父级 `menu` 的 id              |
| `api_group`   | 接口分组     | 固定为 `0`（顶层节点）               |
| `api`         | 单个接口     | 必须填父级 `api_group` 的 id         |

> **重要**：`page` 和 `api` 的 `parent_id` 不能为 `0`，必须挂在对应分组下。

---

## 菜单 / 页面 层级规则

```
menu（菜单分组，parent_id = 0）
└── page（叶子页面，parent_id = menu.id）
```

- **新增页面前**，先确认页面所属的 `menu` 是否已存在；若不存在，同步插入 `menu`。
- `menu` 的 `page_url` 存菜单折叠 key（如 `/user-group`），不是真实路由。
- `page` 的 `page_url` 存前端实际路由（如 `/user/maintenance`），与 `ManagerShell.tsx` 的 `items[].key` 严格对应。
- `menu` 的 `component` 填 Ant Design 图标名（如 `TeamOutlined`）；`page` 的 `component` 留空。
- `page` 的 `resource_url` 留空（页面不做后端 URL 匹配）。

**现有 menu 分组（勿重复插入）：**

```text
menu:user          → 用户管理       /user
menu:system        → 系统设置       /system-group
menu:grain_station → 粮站管理       /grain-station-group
menu:grain_farmer  → 粮户管理       /grain-farmer-group
menu:grain_purchase→ 收粮管理       /grain-purchase-group
```

**新增页面 SQL 模板：**

```sql
-- 1. 先插菜单分组（如已存在可跳过）
INSERT INTO resource_new (name, code, parent_id, resource_type, resource_url, page_url, component, redirect, menu_name, meta, sort_id, active)
SELECT '{菜单名}', 'menu:{domain}', 0, 'menu', '', '/{menu-key}', '{图标名}', '', '{菜单名}', '{副标题}', {排序}, 1
WHERE NOT EXISTS (SELECT 1 FROM resource_new WHERE code = 'menu:{domain}' AND active = 1);

-- 2. 再插叶子页面（parent_id 用子查询取菜单 id）
INSERT INTO resource_new (name, code, parent_id, resource_type, resource_url, page_url, component, redirect, menu_name, meta, sort_id, active)
SELECT '{页面名}', 'page:{domain}',
    (SELECT id FROM resource_new WHERE code = 'menu:{domain}' AND active = 1),
    'page', '', '/{实际路由}', '', '', '{页面名}', '', {排序}, 1
WHERE NOT EXISTS (SELECT 1 FROM resource_new WHERE code = 'page:{domain}' AND active = 1);
```

> MySQL 不允许 UPDATE/INSERT 子查询直接引用同表，但上面 INSERT...SELECT 中的子查询是在 FROM 子句外，MySQL 允许此写法。

---

## 接口分组 / 接口 层级规则

```
api_group（接口分组，parent_id = 0）
└── api（单个接口，parent_id = api_group.id）
```

- **新增接口前**，先确认所属的 `api_group` 是否已存在；若不存在，同步插入 `api_group`。
- `api_group` 的 `resource_url / page_url / component / redirect / menu_name / meta` 全部留空。
- `api` 的 `page_url / component / redirect / menu_name / meta` 全部留空；`resource_url` 填 Gin 路由路径。
- 同一路径不同 HTTP 方法**不合并**，分别建资源，用 `code` 区分 action。

**现有 api_group 分组（勿重复插入）：**

```text
api_group:accounts          → 账户管理接口
api_group:permission        → 权限管理接口
api_group:users             → 用户管理接口
api_group:tenant            → 租户管理接口
api_group:app_user          → 业务员管理接口
api_group:grain_config      → 基本设置接口（粮站/品类/付款/地点）
api_group:grain_farmer      → 粮户管理接口
api_group:grain_purchase    → 收粮管理接口
```

**新增接口 SQL 模板：**

```sql
-- 1. 先插接口分组（如已存在可跳过）
INSERT INTO resource_new (name, code, parent_id, resource_type, resource_url, page_url, component, redirect, menu_name, meta, sort_id, active)
SELECT '{分组名}接口', 'api_group:{domain}', 0, 'api_group', '', '', '', '', '', '', {排序}, 1
WHERE NOT EXISTS (SELECT 1 FROM resource_new WHERE code = 'api_group:{domain}' AND active = 1);

-- 2. 再插接口资源（parent_id 用子查询取分组 id）
INSERT INTO resource_new (name, code, parent_id, resource_type, resource_url, page_url, component, redirect, menu_name, meta, sort_id, active)
SELECT '{接口名}', '{domain}:{action}',
    (SELECT id FROM resource_new WHERE code = 'api_group:{domain}' AND active = 1),
    'api', '/{gin-path}', '', '', '', '', '', {排序}, 1
WHERE NOT EXISTS (SELECT 1 FROM resource_new WHERE code = '{domain}:{action}' AND active = 1);
```

---

## 表结构约定

`resource_new` 常用字段：

```sql
name, code, parent_id, resource_type, resource_url,
page_url, component, redirect, menu_name, meta, sort_id, active
```

管理端接口权限校验主要按 `resource_url` 匹配（Gin 路由参数用 `:id`）。

---

## 输出规则

1. 必须幂等：所有 `INSERT INTO resource_new` 使用 `SELECT ... WHERE NOT EXISTS (...)`。
2. `role_resource_new` 绑定也必须幂等：使用 `NOT EXISTS` 检查同一 `role_id + resource_id + active = 1`。
3. `role_id` 使用用户提供的值；如果用户没提供，先简短询问，不要猜。
4. `parent_id` 通过子查询从父级 `code` 取 id，不要硬编码数字 id。
5. `sort_id` 从 `100` 起递增；如果上下文已有同模块排序，延续已有排序。
6. 同一路径不同方法不合并；修改密码、禁用等复用已有接口时不额外造资源。

---

## 命名规则

```text
code 格式：{domain}:{action}
```

常见 action 映射：

```text
GET  collection  → list
GET  /:id        → detail
GET  stats/count → stats
POST             → create
PUT/PATCH        → update
DELETE           → delete
```

---

## 角色绑定模板

```sql
INSERT INTO role_resource_new (role_id, resource_id, active)
SELECT
    {role_id},
    r.id,
    1
FROM resource_new r
WHERE r.active = 1
  AND r.code IN (
    '{资源编码1}',
    '{资源编码2}'
  )
  AND NOT EXISTS (
    SELECT 1
    FROM role_resource_new rr
    WHERE rr.role_id = {role_id}
      AND rr.resource_id = r.id
      AND rr.active = 1
  );
```

---

## 当前仓库参考

权限表模型：

```text
server/service/manager_permission/repository/model.go
```

管理端路由通常在：

```text
server/manager-api/pkg/{domain}/{domain}.go
```

菜单结构参考：

```text
client/manager/src/app/(console)/shell/ManagerShell.tsx
```

生成 SQL 时优先从当前对话上下文判断新增接口/页面；不确定时可读取对应 handler 的 `RegisterHandler`。
