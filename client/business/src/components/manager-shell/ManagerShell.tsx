"use client";

import {
  AppstoreOutlined,
  BarChartOutlined,
  CompassOutlined,
  DatabaseOutlined,
  DownOutlined,
  LockOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SettingOutlined,
  UserOutlined,
  UsergroupAddOutlined,
} from "@ant-design/icons";
import {
  Avatar,
  Button,
  Dropdown,
  Form,
  Grid,
  Input,
  Layout,
  Menu,
  Modal,
  Space,
  Tag,
  Typography,
  message,
} from "antd";
import type { MenuProps } from "antd";
import { usePathname, useRouter } from "next/navigation";
import { PropsWithChildren, useCallback, useEffect, useMemo, useState } from "react";
import { clearAuthToken, setCurrentAppUser } from "@/utils/auth";
import {
  changeCurrentUserPassword,
  fetchCurrentUserProfile,
  updateCurrentUserProfile,
  type CurrentUserProfile,
  type UpdateCurrentUserProfilePayload,
} from "./api/profile.api";

const { Content, Header, Sider } = Layout;
const { Text } = Typography;
const { useBreakpoint } = Grid;

interface ManagerShellProps extends PropsWithChildren {}

type MenuItem = Required<MenuProps>["items"][number];

type ProfileFormValues = UpdateCurrentUserProfilePayload;

interface PasswordFormValues {
  oldPassword: string;
  newPassword: string;
  confirmPassword: string;
}

const pageTitleMap: Record<string, string> = {
  "/manager-dashboard": "工作台",
  "/grain/farmers": "农户管理",
  "/grain/dashboard": "粮食大盘",
  "/grain/entries": "粮食明细",
};

function getOpenKeys(pathname: string) {
  if (pathname.startsWith("/grain")) {
    if (pathname.startsWith("/grain/farmers")) {
      return ["/grain-farmer-group"];
    }
    if (pathname.startsWith("/grain/dashboard") || pathname.startsWith("/grain/entries")) {
      return ["/grain-purchase-group"];
    }
  }
  return [];
}

function renderMenuLabel(label: string, hint?: string) {
  return (
    <span className="manager-menu-label">
      <span>{label}</span>
      {hint ? <span className="manager-menu-hint">{hint}</span> : null}
    </span>
  );
}

export function ManagerShell({ children }: ManagerShellProps) {
  const pathname = usePathname();
  const router = useRouter();
  const screens = useBreakpoint();
  const activePath = pathname ?? "/manager-dashboard";
  const [openKeys, setOpenKeys] = useState<string[]>(() => getOpenKeys(activePath));
  const [collapsed, setCollapsed] = useState(false);
  const [messageApi, contextHolder] = message.useMessage();
  const [profileForm] = Form.useForm<ProfileFormValues>();
  const [passwordForm] = Form.useForm<PasswordFormValues>();
  const [profile, setProfile] = useState<CurrentUserProfile | null>(null);
  const [profileOpen, setProfileOpen] = useState(false);
  const [passwordOpen, setPasswordOpen] = useState(false);
  const [profileSaving, setProfileSaving] = useState(false);
  const [passwordSaving, setPasswordSaving] = useState(false);
  const quickActions = useMemo(
    () => [
      {
        key: "/manager-dashboard",
        label: "工作台",
        icon: <AppstoreOutlined />,
      },
      {
        key: "/grain/farmers",
        label: "农户管理",
        icon: <UsergroupAddOutlined />,
      },
      {
        key: "/grain/dashboard",
        label: "粮食大盘",
        icon: <BarChartOutlined />,
      },
      {
        key: "/grain/entries",
        label: "粮食明细",
        icon: <DatabaseOutlined />,
      },
    ],
    [],
  );

  const currentModule =
    activePath.startsWith("/grain/farmers")
      ? "农户管理"
      : activePath.startsWith("/grain/dashboard") || activePath.startsWith("/grain/entries")
        ? "粮食管理"
        : "工作台";

  const items = useMemo<MenuItem[]>(
    () => [
      {
        type: "group",
        label: "工作台",
        key: "group-workbench",
      },
      {
        key: "/manager-dashboard",
        icon: <AppstoreOutlined />,
        label: renderMenuLabel("工作台", "经营指标"),
      },
      {
        type: "group",
        label: "粮食收购",
        key: "group-grain",
      },
      {
        key: "/grain-farmer-group",
        icon: <UsergroupAddOutlined />,
        label: renderMenuLabel("粮户管理", "农户管理"),
        children: [
          {
            key: "/grain/farmers",
            label: "农户管理",
          },
        ],
      },
      {
        key: "/grain-purchase-group",
        icon: <DatabaseOutlined />,
        label: renderMenuLabel("收粮管理", "入库流水"),
        children: [
          {
            key: "/grain/dashboard",
            label: "粮食大盘",
          },
          {
            key: "/grain/entries",
            label: "粮食明细",
          },
        ],
      },
    ],
    [],
  );
  const selectedKey = activePath;

  const roleText = "业务账号";

  const displayName = profile?.name || profile?.username || "管理用户";
  const avatarText = displayName.trim().charAt(0).toUpperCase() || "A";

  const loadProfile = useCallback(async () => {
    try {
      const result = await fetchCurrentUserProfile();
      setProfile(result);
      setCurrentAppUser(result);
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "获取个人信息失败");
    }
  }, [messageApi]);

  useEffect(() => {
    const pathOpenKeys = getOpenKeys(activePath);
    if (pathOpenKeys.length === 0) {
      return;
    }
    setOpenKeys((currentKeys) => Array.from(new Set([...currentKeys, ...pathOpenKeys])));
  }, [activePath]);

  useEffect(() => {
    setCollapsed(!screens.lg);
  }, [screens.lg]);

  useEffect(() => {
    void loadProfile();
  }, [loadProfile]);

  const handleLogout = () => {
    clearAuthToken();
    router.replace("/login");
  };

  const openProfileModal = () => {
    profileForm.setFieldsValue({
      name: profile?.name ?? "",
      email: profile?.email ?? "",
      phone: profile?.phone ?? "",
      department: profile?.department ?? "",
      remark: profile?.remark ?? "",
    });
    setProfileOpen(true);
  };

  const openPasswordModal = () => {
    passwordForm.resetFields();
    setPasswordOpen(true);
  };

  const handleProfileSubmit = async () => {
    try {
      const values = await profileForm.validateFields();
      setProfileSaving(true);
      const result = await updateCurrentUserProfile(values);
      setProfile(result);
      setProfileOpen(false);
      messageApi.success("个人信息已更新");
    } catch (error) {
      if (error instanceof Error) {
        messageApi.error(error.message);
      }
    } finally {
      setProfileSaving(false);
    }
  };

  const handlePasswordSubmit = async () => {
    try {
      const values = await passwordForm.validateFields();
      setPasswordSaving(true);
      await changeCurrentUserPassword({
        oldPassword: values.oldPassword,
        newPassword: values.newPassword,
      });
      setPasswordOpen(false);
      passwordForm.resetFields();
      messageApi.success("密码已修改");
    } catch (error) {
      if (error instanceof Error) {
        messageApi.error(error.message);
      }
    } finally {
      setPasswordSaving(false);
    }
  };

  const pageTitle =
    Object.entries(pageTitleMap).find(([path]) => activePath.startsWith(path))?.[1] ??
    "管理工作台";

  return (
    <div className="manager-app-frame">
      {contextHolder}
      <div className={`manager-shell-surface ${collapsed ? "manager-shell-surface-collapsed" : ""}`}>
        <Layout
          style={{
            minHeight: "100vh",
            background: "transparent",
          }}
        >
          <Sider
            width={248}
            collapsedWidth={screens.md ? 82 : 0}
            collapsible
            collapsed={collapsed}
            trigger={null}
            style={{
              background: "transparent",
            }}
          >
            <div
              className="manager-sidebar-card manager-stagger-1"
              style={{
                height: "100%",
                padding: "20px 14px",
                display: "flex",
                flexDirection: "column",
                gap: 14,
              }}
            >
              <div className="manager-sidebar-brand">
                <Space align="center" size={12}>
                  <div className="manager-product-logo" aria-hidden="true">
                    <span />
                    <i />
                  </div>
                  <div className="manager-wordmark manager-collapse-hidden">
                    <strong style={{ color: "#fff" }}>收粮业务端</strong>
                    <span style={{ color: "rgba(255,255,255,0.66)" }}>Shennong Admin</span>
                  </div>
                </Space>
                <Tag bordered={false} className="manager-sidebar-env manager-collapse-hidden">
                  运营后台
                </Tag>
              </div>

              <Menu
                className="manager-shell-menu"
                mode="inline"
                selectedKeys={[selectedKey]}
                openKeys={openKeys}
                onOpenChange={(keys) => setOpenKeys(keys as string[])}
                items={items}
                onClick={({ key }) => {
                  if (typeof key === "string" && key.startsWith("/")) {
                    router.push(key);
                  }
                }}
                style={{
                  fontSize: 15,
                  marginTop: 2,
                }}
              />
              <div className="manager-sidebar-foot">
                <div className="manager-sidebar-status-icon">
                  <SettingOutlined />
                </div>
                <div className="manager-collapse-hidden">
                  <span>当前模块</span>
                  <strong>{currentModule}</strong>
                  <Tag bordered={false}>角色权限</Tag>
                </div>
              </div>
            </div>
          </Sider>

          <Layout style={{ background: "transparent" }}>
            <Header
              className="manager-stagger-2"
              style={{
                height: "auto",
                lineHeight: "normal",
                padding: 0,
                background: "transparent",
              }}
            >
              <div
                className="manager-shell-card"
                style={{
                  borderRadius: 0,
                  padding: "0 28px 0 30px",
                  minHeight: 76,
                  display: "grid",
                  gridTemplateColumns: "minmax(0, 1fr) auto",
                  gap: 20,
                  alignItems: "center",
                }}
              >
                <div className="manager-header-main">
                  <Space size={10} align="center" style={{ marginBottom: 8 }}>
                    <Button
                      type="text"
                      icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                      className="manager-nav-toggle"
                      onClick={() => setCollapsed((value) => !value)}
                    />
                    <CompassOutlined style={{ color: "var(--manager-primary)" }} />
                    <Text style={{ color: "var(--manager-text-soft)", fontWeight: 700 }}>
                      {pageTitle}
                    </Text>
                  </Space>
                  <Space size={10} wrap style={{ width: "100%" }}>
                    {quickActions.map((action) => {
                      const isActive = activePath === action.key;

                      return (
                        <Button
                          key={action.key}
                          type={isActive ? "primary" : "default"}
                          icon={action.icon}
                          className={isActive ? "manager-soft-button" : undefined}
                          onClick={() => router.push(action.key)}
                          style={{
                            height: 38,
                            paddingInline: 14,
                            borderRadius: 8,
                            fontWeight: 700,
                          }}
                        >
                          {action.label}
                        </Button>
                      );
                    })}
                  </Space>
                </div>

                <Space size={12} wrap className="manager-header-account">
                  <Dropdown
                    trigger={["click"]}
                    menu={{
                      items: [
                        {
                          key: "profile",
                          icon: <UserOutlined />,
                          label: "个人信息",
                        },
                        {
                          key: "password",
                          icon: <LockOutlined />,
                          label: "修改密码",
                        },
                        {
                          type: "divider",
                        },
                        {
                          key: "logout",
                          icon: <LogoutOutlined />,
                          label: "退出登录",
                        },
                      ],
                      onClick: ({ key }) => {
                        if (key === "profile") {
                          openProfileModal();
                        }
                        if (key === "password") {
                          openPasswordModal();
                        }
                        if (key === "logout") {
                          handleLogout();
                        }
                      },
                    }}
                  >
                    <Button
                      type="text"
                      style={{
                        height: 56,
                        padding: "6px 12px 6px 8px",
                        borderRadius: 8,
                        border: "1px solid var(--manager-border)",
                        background: "#ffffff",
                      }}
                    >
                      <Space size={12}>
                        <Avatar
                          style={{
                            width: 38,
                            height: 38,
                            background: "linear-gradient(135deg, var(--manager-primary), #4f8f5f)",
                            color: "#fff",
                            fontWeight: 700,
                          }}
                        >
                          {avatarText}
                        </Avatar>
                        <div style={{ textAlign: "left", lineHeight: 1.25 }}>
                          <div style={{ fontWeight: 700, color: "var(--manager-text)" }}>
                            {displayName}
                          </div>
                          <Text style={{ color: "var(--manager-text-soft)" }}>{roleText}</Text>
                        </div>
                        <DownOutlined style={{ color: "var(--manager-text-soft)", fontSize: 12 }} />
                      </Space>
                    </Button>
                  </Dropdown>
                </Space>
              </div>
            </Header>

            <Content style={{ padding: "22px 28px 40px" }}>
              <div className="manager-stagger-3">{children}</div>
            </Content>
          </Layout>
        </Layout>
      </div>
      <Modal
        title="个人信息"
        open={profileOpen}
        onCancel={() => setProfileOpen(false)}
        onOk={handleProfileSubmit}
        confirmLoading={profileSaving}
        okText="保存"
        cancelText="取消"
        width={520}
      >
        <Form form={profileForm} layout="vertical" style={{ paddingTop: 8 }}>
          <Form.Item
            name="name"
            label="姓名"
            rules={[{ required: true, message: "请输入姓名" }]}
          >
            <Input placeholder="请输入姓名" />
          </Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ type: "email", message: "邮箱格式不正确" }]}>
            <Input placeholder="请输入邮箱" />
          </Form.Item>
          <Form.Item name="phone" label="手机号">
            <Input placeholder="请输入手机号" />
          </Form.Item>
          <Form.Item name="department" label="部门">
            <Input placeholder="请输入部门" />
          </Form.Item>
          <Form.Item name="remark" label="备注">
            <Input.TextArea rows={3} placeholder="请输入备注" />
          </Form.Item>
        </Form>
      </Modal>
      <Modal
        title="修改密码"
        open={passwordOpen}
        onCancel={() => setPasswordOpen(false)}
        onOk={handlePasswordSubmit}
        confirmLoading={passwordSaving}
        okText="确认修改"
        cancelText="取消"
        width={460}
      >
        <Form form={passwordForm} layout="vertical" style={{ paddingTop: 8 }}>
          <Form.Item
            name="oldPassword"
            label="原密码"
            rules={[{ required: true, message: "请输入原密码" }]}
          >
            <Input.Password placeholder="请输入原密码" />
          </Form.Item>
          <Form.Item
            name="newPassword"
            label="新密码"
            rules={[
              { required: true, message: "请输入新密码" },
              { min: 6, message: "密码至少 6 位" },
              { max: 50, message: "密码最多 50 位" },
            ]}
          >
            <Input.Password placeholder="请输入新密码" />
          </Form.Item>
          <Form.Item
            name="confirmPassword"
            label="确认新密码"
            dependencies={["newPassword"]}
            rules={[
              { required: true, message: "请再次输入新密码" },
              ({ getFieldValue }) => ({
                validator(_, value: string | undefined) {
                  if (!value || getFieldValue("newPassword") === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error("两次输入的新密码不一致"));
                },
              }),
            ]}
          >
            <Input.Password placeholder="请再次输入新密码" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
