"use client";

import { useEffect, useMemo, useState } from "react";
import {
  ApartmentOutlined,
  CheckCircleOutlined,
  EditOutlined,
  LockOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
  StopOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import {
  Button,
  Input,
  message,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  Tooltip,
  Typography,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import {
  fetchRoleOptions,
  fetchTenantOptions,
  fetchUserRoles,
  setUserRoles,
  type RoleOption,
  type TenantOption,
  type UserPayload,
  type UserRecord,
} from "../api/user.api";
import { CryptoUtil } from "@/utils/crypto.util";
import { UserFormModal } from "./UserFormModal";
import { useUserManagement } from "../hooks/useUserManagement";

const { Text } = Typography;

const statusColors: Record<string, string> = {
  normal: "rgba(95,198,163,0.14)",
  frozen: "rgba(239,107,120,0.14)",
  active: "rgba(95,198,163,0.14)",
  ACTIVE: "rgba(95,198,163,0.14)",
  expire: "rgba(239,107,120,0.14)",
  EXPIRE: "rgba(239,107,120,0.14)",
  inactive: "rgba(170,192,238,0.16)",
  locked: "rgba(239,107,120,0.14)",
  disabled: "rgba(239,107,120,0.14)",
};

interface UserManagementDemoProps {
  mode?: "maintenance" | "list";
}

export function UserManagementDemo({ mode: _mode = "maintenance" }: UserManagementDemoProps) {
  const {
    users,
    stats,
    total,
    query,
    loading,
    statsLoading,
    submitting,
    refresh,
    saveUser,
    patchUser,
  } = useUserManagement();

  const [searchValue, setSearchValue] = useState(query.search);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<UserRecord | null>(null);
  const [tenantOptions, setTenantOptions] = useState<TenantOption[]>([]);
  const [roleOptions, setRoleOptions] = useState<RoleOption[]>([]);

  useEffect(() => {
    fetchTenantOptions()
      .then((result) => setTenantOptions(result.data))
      .catch((error) => message.error(error instanceof Error ? error.message : "加载租户失败"));
    fetchRoleOptions()
      .then((result) => setRoleOptions(result.data))
      .catch((error) => message.error(error instanceof Error ? error.message : "加载角色失败"));
  }, []);

  const activeCount = users.filter((item) => resolveUserStatus(item) === "active").length;

  const heroStats = useMemo(
    () => [
      { label: "可见用户", value: stats.visibleUsers },
      { label: "活跃用户", value: stats.activeUsers || activeCount },
    ],
    [activeCount, stats.activeUsers, stats.visibleUsers],
  );

  const handleCreate = () => {
    setEditingUser(null);
    setModalOpen(true);
  };

  const handleEdit = (record: UserRecord) => {
    setEditingUser(record);
    setModalOpen(true);
  };

  const handleSubmit = async (payload: UserPayload) => {
    try {
      await saveUser(editingUser?.id ?? null, payload);
      message.success(editingUser ? "用户更新成功" : "用户创建成功");
      setModalOpen(false);
      setEditingUser(null);
    } catch (error) {
      message.error(error instanceof Error ? error.message : "保存用户失败");
    }
  };

  const handleChangeRole = async (record: UserRecord) => {
    const selectableRoles = roleOptions.map((r) => ({ label: r.name, value: r.code }));
    let currentCodes: string[] = [];
    try {
      const result = await fetchUserRoles(record.id);
      currentCodes = result.data
        .map((ur) => roleOptions.find((r) => r.id === ur.roleId)?.code)
        .filter((code): code is string => !!code);
    } catch {
      // 拉取失败时用用户当前 role 兜底
      if (record.role) currentCodes = [record.role];
    }
    let nextCodes: string[] = [...currentCodes];
    Modal.confirm({
      title: `配置角色 — ${record.username}`,
      width: 480,
      content: (
        <Select<string[]>
          mode="multiple"
          defaultValue={currentCodes}
          style={{ width: "100%", marginTop: 16 }}
          options={selectableRoles}
          optionFilterProp="label"
          placeholder="请选择角色（可多选）"
          onChange={(values) => {
            nextCodes = values;
          }}
        />
      ),
      onOk: async () => {
        await setUserRoles(record.id, nextCodes);
        await refresh();
        message.success("角色已更新");
      },
    });
  };

  const handleConfigTenants = (record: UserRecord) => {
    let nextTenantIds: number[] = record.tenantIds ?? [];
    const selectableTenants = tenantOptions
      .filter((t) => t.status !== "inactive")
      .map((t) => ({ label: t.tenantName, value: t.id }));
    Modal.confirm({
      title: `配置租户 — ${record.username}`,
      width: 480,
      content: (
        <Select<number[]>
          mode="multiple"
          defaultValue={record.tenantIds ?? []}
          style={{ width: "100%", marginTop: 16 }}
          placeholder="请选择关联租户（可多选）"
          options={selectableTenants}
          optionFilterProp="label"
          maxTagCount="responsive"
          onChange={(value) => {
            nextTenantIds = value;
          }}
        />
      ),
      onOk: async () => {
        await patchUser(record.id, { tenantIds: nextTenantIds });
        message.success("租户配置已更新");
      },
    });
  };

  const handleChangeRemark = (record: UserRecord) => {
    let nextRemark = record.remark || "";
    Modal.confirm({
      title: "修改备注",
      content: (
        <Input.TextArea
          rows={4}
          defaultValue={record.remark}
          placeholder="请输入备注"
          style={{ marginTop: 16 }}
          onChange={(event) => {
            nextRemark = event.target.value;
          }}
        />
      ),
      onOk: async () => {
        await patchUser(record.id, { remark: nextRemark });
        message.success("备注已更新");
      },
    });
  };

  const handleChangePassword = (record: UserRecord) => {
    let nextPassword = "";
    Modal.confirm({
      title: "修改密码",
      content: (
        <Input.Password
          placeholder="请输入新密码"
          style={{ marginTop: 16 }}
          onChange={(event) => {
            nextPassword = event.target.value;
          }}
        />
      ),
      onOk: async () => {
        const password = nextPassword.trim();
        if (!password) {
          throw new Error("请输入新密码");
        }
        await patchUser(record.id, {
          password: CryptoUtil.encrypt(password),
          originPassword: password,
        });
        message.success("密码已更新");
      },
    });
  };

  const handleToggleFreeze = (record: UserRecord) => {
    const frozen = isFrozenUser(record);
    const nextStatus = frozen ? "active" : "inactive";
    Modal.confirm({
      title: frozen ? "解冻用户" : "冻结用户",
      content: frozen ? "解冻后用户状态将恢复为启用。" : "冻结后用户状态将变为不可用。",
      onOk: async () => {
        await patchUser(record.id, { status: nextStatus });
        message.success(frozen ? "已解冻" : "已冻结");
      },
    });
  };

  const columns: ColumnsType<UserRecord> = [
    {
      title: "ID",
      dataIndex: "id",
      key: "id",
      width: 80,
    },
    {
      title: "用户名",
      dataIndex: "username",
      key: "username",
      width: 160,
    },
    {
      title: "关联租户",
      key: "tenantNames",
      width: 260,
      render: (_, record) => {
        const names = Array.isArray(record.tenantNames)
          ? record.tenantNames.filter((n): n is string => typeof n === "string")
          : [];
        if (names.length === 0) return <Text style={{ color: "var(--manager-text-faint)" }}>-</Text>;
        return (
          <div style={{ display: "flex", flexWrap: "wrap", gap: 4 }}>
            {names.map((name) => (
              <Tag key={name} style={{ margin: 0 }}>
                {name}
              </Tag>
            ))}
          </div>
        );
      },
    },
    {
      title: "密码",
      key: "password",
      width: 140,
      render: (_, record) => record.originPassword || record.password || "-",
    },
    {
      title: "密钥",
      dataIndex: "secretKey",
      key: "secretKey",
      width: 220,
      render: (value: string) => wrapText(value),
    },
    {
      title: "备注",
      dataIndex: "remark",
      key: "remark",
      width: 140,
      render: (value: string) => value || "-",
    },
    {
      title: "状态",
      key: "status",
      width: 110,
      render: (_, record) => {
        const value = resolveDisplayStatus(record);
        return (
          <Tag
            style={{
              color: "var(--manager-text)",
              background: statusColors[value] || "rgba(170,192,238,0.16)",
              border: "none",
            }}
          >
            {formatStatus(value)}
          </Tag>
        );
      },
    },
    {
      title: "操作",
      key: "actions",
      width: 240,
      fixed: "right",
      render: (_, record) => {
        const frozen = isFrozenUser(record);

        return (
          <Space size={4} wrap>
            <Tooltip title="编辑用户">
              <Button
                size="small"
                type="text"
                icon={<EditOutlined />}
                onClick={() => handleEdit(record)}
              />
            </Tooltip>
            <Tooltip title="配置租户">
              <Button
                size="small"
                type="text"
                icon={<ApartmentOutlined />}
                onClick={() => handleConfigTenants(record)}
              />
            </Tooltip>
            <Tooltip title="配置角色">
              <Button
                size="small"
                type="text"
                icon={<TeamOutlined />}
                onClick={() => handleChangeRole(record)}
              />
            </Tooltip>
            <Tooltip title="修改备注">
              <Button
                size="small"
                type="text"
                icon={<EditOutlined />}
                onClick={() => handleChangeRemark(record)}
              />
            </Tooltip>
            <Tooltip title="修改密码">
              <Button
                size="small"
                type="text"
                icon={<LockOutlined />}
                onClick={() => handleChangePassword(record)}
              />
            </Tooltip>
            <Tooltip title={frozen ? "解冻" : "冻结"}>
              <Button
                size="small"
                type="text"
                danger={!frozen}
                icon={frozen ? <CheckCircleOutlined /> : <StopOutlined />}
                onClick={() => handleToggleFreeze(record)}
              />
            </Tooltip>
          </Space>
        );
      },
    },
  ];

  return (
    <div className="manager-page-stack">
      <section
        className="manager-stats-grid"
        style={{ gridTemplateColumns: "repeat(auto-fit, minmax(150px, 150px))" }}
      >
        {heroStats.map((item) => (
          <div key={item.label} className="manager-metric-chip manager-metric-chip-compact">
            <Text style={{ color: "var(--manager-text-faint)", fontSize: 12 }}>{item.label}</Text>
            <div className="manager-value" style={{ marginTop: 4, fontSize: 22, lineHeight: 1.1 }}>
              {item.value}
            </div>
          </div>
        ))}
      </section>

      <section className="manager-data-card">
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            gap: 16,
            flexWrap: "wrap",
          }}
        >
          <Space wrap size={12}>
            <Input
              className="manager-filter-input"
              prefix={<SearchOutlined style={{ color: "var(--manager-text-faint)" }} />}
              placeholder="搜索姓名、账号或邮箱"
              value={searchValue}
              onChange={(event) => setSearchValue(event.target.value)}
              onPressEnter={() => void refresh({ pageIndex: 1, search: searchValue })}
              style={{ width: 280 }}
            />
            <Select
              className="manager-filter-input"
              value={query.status || undefined}
              allowClear
              placeholder="状态筛选"
              onChange={(value) => void refresh({ pageIndex: 1, status: value ?? "" })}
              style={{ width: 160 }}
              options={[
                { label: "启用", value: "active" },
                { label: "不可用", value: "inactive" },
              ]}
            />
            <Button
              icon={<ReloadOutlined />}
              loading={loading || statsLoading}
              onClick={() => void refresh({ pageIndex: 1, search: searchValue })}
            >
              刷新数据
            </Button>
          </Space>

          <Space wrap>
            <Tag
              style={{
                color: "var(--manager-text-soft)",
                background: "var(--manager-green-soft)",
                border: "none",
              }}
            >
              共 {total} 条
            </Tag>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={handleCreate}
              style={{
                color: "#ffffff",
                border: "none",
                background: "linear-gradient(135deg, #145535 0%, #237a4b 100%)",
              }}
            >
              新建用户
            </Button>
          </Space>
        </div>
      </section>

      <section className="manager-data-card manager-table">
        <Table<UserRecord>
          rowKey="id"
          scroll={{ x: 1500 }}
          loading={loading}
          dataSource={users}
          columns={columns}
          pagination={{
            current: query.pageIndex,
            pageSize: query.pageSize,
            total,
            showSizeChanger: false,
            onChange: (page) => void refresh({ pageIndex: page, search: searchValue }),
          }}
        />
      </section>

      <UserFormModal
        open={modalOpen}
        submitting={submitting}
        user={editingUser}
        tenantOptions={tenantOptions}
        roleOptions={roleOptions}
        onCancel={() => {
          setModalOpen(false);
          setEditingUser(null);
        }}
        onSubmit={handleSubmit}
      />
    </div>
  );
}

function resolveUserStatus(record: UserRecord) {
  return record.status || "active";
}

function isFrozenUser(record: UserRecord) {
  return ["inactive", "locked", "frozen", "disabled"].includes(resolveUserStatus(record));
}

function resolveDisplayStatus(record: UserRecord) {
  return record.status || record.accountStatus || "active";
}

function formatStatus(value: string) {
  switch (value) {
    case "ACTIVE":
    case "normal":
    case "active":
    case "pending":
      return "激活";
    case "expire":
    case "EXPIRE":
    case "frozen":
    case "locked":
    case "deleted":
      return "冻结";
    case "inactive":
    case "disabled":
      return "不可用";
    default:
      return value ? `未知(${value})` : "-";
  }
}

function wrapText(value?: string) {
  if (!value) {
    return "-";
  }
  return (
    <div style={{ whiteSpace: "normal", wordBreak: "break-all", color: "var(--manager-text)" }}>
      {value}
    </div>
  );
}
