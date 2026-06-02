"use client";

import { useEffect, useMemo, useState } from "react";
import {
  CheckCircleOutlined,
  EditOutlined,
  LockOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
  SettingOutlined,
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
import { tenantApi, type TenantRecord } from "../../tenant/api/tenant.api";
import { createAccount, updateAccount, type UserPayload, type UserRecord } from "../api/user.api";
import { UserFormModal } from "./UserFormModal";
import { useUserManagement } from "../hooks/useUserManagement";

const { Text } = Typography;

const roleColors: Record<string, string> = {
  admin: "var(--manager-green-soft)",
  manager: "var(--manager-blue-soft)",
  auditor: "var(--manager-gold-soft)",
  member: "#f7faf5",
};

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

export function UserManagementDemo({ mode = "maintenance" }: UserManagementDemoProps) {
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
  const [tenantConfigUser, setTenantConfigUser] = useState<UserRecord | null>(null);
  const [tenantConfigTenantId, setTenantConfigTenantId] = useState<number | undefined>();
  const [tenantConfigTenants, setTenantConfigTenants] = useState<TenantRecord[]>([]);
  const [tenantConfigLoading, setTenantConfigLoading] = useState(false);
  const showTenantConfig = mode === "maintenance";

  const activeCount = users.filter((item) => resolveUserStatus(item) === "normal").length;
  const totalBalance = users.reduce((sum, item) => sum + resolveBalance(item), 0);

  const heroStats = useMemo(
    () => [
      { label: "可见用户", value: stats.visibleUsers },
      { label: "活跃用户", value: stats.activeUsers || activeCount },
      { label: "资金账户", value: stats.accountCount },
      { label: "钱包总额", value: formatCurrency(totalBalance) },
    ],
    [activeCount, stats.accountCount, stats.activeUsers, stats.visibleUsers, totalBalance],
  );

  const handleCreate = () => {
    setEditingUser(null);
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

  const handleChangeRole = (record: UserRecord) => {
    let nextRole = record.role;
    Modal.confirm({
      title: "修改角色",
      content: (
        <Select<string>
          defaultValue={record.role || "member"}
          style={{ width: "100%", marginTop: 16 }}
          onChange={(value) => {
            nextRole = value;
          }}
          options={[
            { label: "管理员", value: "admin" },
            { label: "经理", value: "manager" },
            { label: "审计", value: "auditor" },
            { label: "代理", value: "member" },
          ]}
        />
      ),
      onOk: async () => {
        await patchUser(record.id, { role: nextRole });
        message.success("角色已更新");
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
          password,
          originPassword: password,
        });
        message.success("密码已更新");
      },
    });
  };

  const handleOpenTenantConfig = (record: UserRecord) => {
    setTenantConfigUser(record);
    setTenantConfigTenantId(record.tenantId || undefined);
  };

  const handleToggleFreeze = (record: UserRecord) => {
    const currentStatus = resolveUserStatus(record);
    const nextStatus = currentStatus === "frozen" ? "normal" : "frozen";
    Modal.confirm({
      title: nextStatus === "frozen" ? "冻结账户" : "解冻账户",
      onOk: async () => {
        if (record.accountId) {
          await updateAccount(record.accountId, { accountStatus: nextStatus });
        } else {
          await createAccount({
            userId: record.id,
            accountStatus: nextStatus,
            balanceAmount: resolveBalance(record).toFixed(2),
          });
        }
        await refresh();
        message.success(nextStatus === "frozen" ? "已冻结" : "已解冻");
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
      title: "租户ID",
      dataIndex: "tenantId",
      key: "tenantId",
      width: 100,
      render: (value: number) => value || "全局",
    },
    {
      title: "角色",
      dataIndex: "role",
      key: "role",
      width: 110,
      render: (value: string) => (
        <Tag
          style={{
            width: "fit-content",
            color: "var(--manager-text)",
            background: roleColors[value] || "rgba(239,244,251,0.98)",
            border: "none",
          }}
        >
          {formatRole(value)}
        </Tag>
      ),
    },
    {
      title: "余额",
      key: "balanceAmount",
      width: 140,
      align: "right",
      render: (_, record) => {
        const balance = resolveBalance(record);
        return <Text style={{ color: "var(--manager-text)" }}>{formatNumber(balance)}</Text>;
      },
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
      width: showTenantConfig ? 248 : 208,
      fixed: "right",
      render: (_, record) => {
        const frozen = resolveUserStatus(record) === "frozen";

        return (
          <Space size={4} wrap>
            <Tooltip title="修改角色">
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
            {showTenantConfig ? (
              <Tooltip title="租户配置">
                <Button
                  size="small"
                  type="text"
                  icon={<SettingOutlined />}
                  onClick={() => handleOpenTenantConfig(record)}
                />
              </Tooltip>
            ) : null}
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

  const tenantConfigOptions = useMemo(
    () => tenantConfigTenants.map((tenant) => ({ label: tenant.tenantName, value: tenant.id })),
    [tenantConfigTenants],
  );

  useEffect(() => {
    if (!tenantConfigUser) {
      return;
    }
    setTenantConfigLoading(true);
    tenantApi
      .list({ pageIndex: 1, pageSize: 200, status: "active" })
      .then((result) => setTenantConfigTenants(result.data))
      .catch((error) => message.error(error instanceof Error ? error.message : "加载租户失败"))
      .finally(() => setTenantConfigLoading(false));
  }, [tenantConfigUser]);

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
              value={query.role || undefined}
              allowClear
              placeholder="角色筛选"
              onChange={(value) => void refresh({ pageIndex: 1, role: value ?? "" })}
              style={{ width: 160 }}
              options={[
                { label: "管理员", value: "admin" },
                { label: "经理", value: "manager" },
                { label: "审计", value: "auditor" },
                { label: "代理", value: "member" },
              ]}
            />
            <Select
              className="manager-filter-input"
              value={query.status || undefined}
              allowClear
              placeholder="状态筛选"
              onChange={(value) => void refresh({ pageIndex: 1, status: value ?? "" })}
              style={{ width: 160 }}
              options={[
                { label: "激活", value: "ACTIVE" },
                { label: "冻结", value: "EXPIRE" },
              ]}
            />
            <Button
              icon={<ReloadOutlined />}
              loading={loading || statsLoading}
              onClick={() =>
                void refresh({
                  pageIndex: 1,
                  search: searchValue,
                })
              }
            >
              刷新数据
            </Button>
          </Space>

          <Space wrap>
            <Tag style={{ color: "var(--manager-text-soft)", background: "var(--manager-green-soft)", border: "none" }}>
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
          scroll={{ x: 1540 }}
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
        showTenantField={mode === "maintenance"}
        onCancel={() => {
          setModalOpen(false);
          setEditingUser(null);
        }}
        onSubmit={handleSubmit}
      />
      <Modal
        wrapClassName="manager-form-skin"
        destroyOnClose
        open={Boolean(tenantConfigUser)}
        title={tenantConfigUser ? `租户配置：${tenantConfigUser.username}` : "租户配置"}
        okText="保存配置"
        cancelText="取消"
        confirmLoading={submitting}
        onCancel={() => {
          setTenantConfigUser(null);
          setTenantConfigTenantId(undefined);
        }}
        onOk={async () => {
          if (!tenantConfigUser) {
            return;
          }
          try {
            await patchUser(tenantConfigUser.id, { tenantId: tenantConfigTenantId ?? 0 });
            message.success("租户配置已更新");
            setTenantConfigUser(null);
            setTenantConfigTenantId(undefined);
          } catch (error) {
            message.error(error instanceof Error ? error.message : "保存租户配置失败");
          }
        }}
      >
        <Select<number>
          allowClear
          loading={tenantConfigLoading}
          options={tenantConfigOptions}
          placeholder="请选择租户"
          style={{ width: "100%" }}
          value={tenantConfigTenantId}
          onChange={(value) => setTenantConfigTenantId(value)}
        />
      </Modal>
    </div>
  );
}

function resolveBalance(record: UserRecord) {
  if (typeof record.tineBalance === "number") {
    return record.tineBalance;
  }
  return Number(record.balanceAmount || 0);
}

function resolveUserStatus(record: UserRecord) {
  return record.accountStatus || record.status || "normal";
}

function resolveDisplayStatus(record: UserRecord) {
  return record.status || record.accountStatus || "active";
}

function formatRole(value: string) {
  switch (value) {
    case "admin":
      return "管理员";
    case "manager":
      return "经理";
    case "auditor":
      return "审计";
    case "member":
      return "代理";
    default:
      return value || "-";
  }
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
    case "inactive":
    case "disabled":
    case "deleted":
      return "冻结";
    default:
      return value ? `未知(${value})` : "-";
  }
}

function formatNumber(value: number) {
  return new Intl.NumberFormat("zh-CN", {
    minimumFractionDigits: 0,
    maximumFractionDigits: 4,
  }).format(value);
}

function formatCurrency(value: number) {
  return new Intl.NumberFormat("zh-CN", {
    style: "currency",
    currency: "CNY",
    maximumFractionDigits: 2,
  }).format(value);
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
