"use client";

import {
  AppstoreOutlined,
  BankOutlined,
  BarChartOutlined,
  CompassOutlined,
  DatabaseOutlined,
  DownOutlined,
  LockOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SafetyCertificateOutlined,
  SettingOutlined,
  TeamOutlined,
  UserOutlined,
  UsergroupAddOutlined,
} from "@ant-design/icons";
import {
  Avatar,
  Button,
  Drawer,
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
import {
  ComponentType,
  PropsWithChildren,
  ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";
import { clearAuthToken } from "@/utils/auth";
import { CryptoUtil } from "@/utils/crypto.util";
import { withoutBasePath } from "@/utils/routes";
import {
  changeCurrentUserPassword,
  fetchCurrentUserMenuResources,
  fetchCurrentUserProfile,
  updateCurrentUserProfile,
  type CurrentUserProfile,
  type ResourceItem,
  type UpdateCurrentUserProfilePayload,
} from "./api/profile.api";

const { Content, Header, Sider } = Layout;
const { Text } = Typography;
const { useBreakpoint } = Grid;

interface ManagerShellProps extends PropsWithChildren {}

type MenuItem = Required<MenuProps>["items"][number];
type IconComponent = ComponentType;

type ProfileFormValues = UpdateCurrentUserProfilePayload;

interface PasswordFormValues {
  oldPassword: string;
  newPassword: string;
  confirmPassword: string;
}

const menuLoadingRows = [
  { key: "dashboard", width: "72%" },
  { key: "user", width: "58%" },
  { key: "permission", width: "78%" },
  { key: "grain", width: "64%" },
  { key: "settings", width: "52%" },
];

const fallbackPageTitleMap: Record<string, string> = {
  "/manager-dashboard": "数据总览",
  "/user/maintenance": "用户维护",
  "/user/tenant": "租户配置",
  "/user/list": "用户列表",
  "/user": "用户管理",
  "/permission": "权限管理",
  "/tenant": "租户配置",
  "/grain/stations": "粮站列表",
  "/grain/config": "基础设置",
  "/grain/payment-methods": "付款方式",
  "/grain/farmers": "农户管理",
  "/app-user": "业务员管理",
  "/grain/dashboard": "收粮大盘",
  "/grain/entries": "收粮明细",
};

const fallbackQuickActionIcons: Record<string, ReactNode> = {
  "/manager-dashboard": <AppstoreOutlined />,
  "/user/maintenance": <TeamOutlined />,
  "/permission": <SafetyCertificateOutlined />,
  "/app-user": <TeamOutlined />,
  "/grain/dashboard": <BarChartOutlined />,
  "/grain/entries": <DatabaseOutlined />,
};

const menuIconMap: Record<string, IconComponent> = {
  AppstoreOutlined,
  BankOutlined,
  BarChartOutlined,
  DatabaseOutlined,
  SafetyCertificateOutlined,
  TeamOutlined,
  UsergroupAddOutlined,
};

function normalizeResourceType(resourceType: string) {
  return resourceType.trim().toLowerCase();
}

function normalizePageKey(resource: ResourceItem) {
  return resource.pageUrl || resource.resourceUrl || `resource:${resource.id}`;
}

function getResourceLabel(resource: ResourceItem) {
  return resource.menuName || resource.name || resource.code || `资源 ${resource.id}`;
}

function getResourceIcon(resource: ResourceItem) {
  const IconComponent = menuIconMap[resource.component];
  return IconComponent ? <IconComponent /> : undefined;
}

function sortResources(a: ResourceItem, b: ResourceItem) {
  return a.sortId - b.sortId || a.id - b.id;
}

function buildDynamicMenuItems(resources: ResourceItem[]): MenuItem[] {
  const menuResources = resources
    .filter((resource) => ["menu", "page"].includes(normalizeResourceType(resource.resourceType)))
    .sort(sortResources);
  const childrenByParent = new Map<number, ResourceItem[]>();

  for (const resource of menuResources) {
    const siblings = childrenByParent.get(resource.parentId) ?? [];
    siblings.push(resource);
    childrenByParent.set(resource.parentId, siblings);
  }

  const buildItem = (resource: ResourceItem): MenuItem | null => {
    const resourceType = normalizeResourceType(resource.resourceType);
    const key = normalizePageKey(resource);

    if (resourceType === "page") {
      if (!resource.pageUrl) {
        return null;
      }
      return {
        key,
        icon: resource.parentId === 0 ? getResourceIcon(resource) : undefined,
        label:
          resource.parentId === 0
            ? renderMenuLabel(getResourceLabel(resource), resource.meta)
            : getResourceLabel(resource),
      };
    }

    const children = (childrenByParent.get(resource.id) ?? [])
      .sort(sortResources)
      .map(buildItem)
      .filter((item): item is MenuItem => item !== null);

    if (children.length === 0) {
      return null;
    }

    return {
      key,
      icon: getResourceIcon(resource),
      label: renderMenuLabel(getResourceLabel(resource), resource.meta),
      children,
    };
  };

  return (childrenByParent.get(0) ?? [])
    .sort(sortResources)
    .map(buildItem)
    .filter((item): item is MenuItem => item !== null);
}

function getOpenKeysFromMenuItems(pathname: string, items: MenuItem[]) {
  const keys: string[] = [];

  const visit = (item: MenuItem, parents: string[]): boolean => {
    if (!item || typeof item !== "object") {
      return false;
    }

    const menuItem = item as { key?: string; children?: MenuItem[] };
    const itemKey = typeof menuItem.key === "string" ? menuItem.key : "";
    if (itemKey && pathname.startsWith(itemKey) && !menuItem.children) {
      keys.push(...parents);
      return true;
    }

    if (Array.isArray(menuItem.children)) {
      return menuItem.children.some((child) => visit(child, itemKey ? [...parents, itemKey] : parents));
    }

    return false;
  };

  items.some((item) => visit(item, []));
  return Array.from(new Set(keys));
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
  const activePath = withoutBasePath(pathname ?? "/manager-dashboard");
  const [openKeys, setOpenKeys] = useState<string[]>([]);
  const [collapsed, setCollapsed] = useState(true);
  const [mounted, setMounted] = useState(false);
  useEffect(() => setMounted(true), []);
  const [messageApi, contextHolder] = message.useMessage();
  const [profileForm] = Form.useForm<ProfileFormValues>();
  const [passwordForm] = Form.useForm<PasswordFormValues>();
  const [profile, setProfile] = useState<CurrentUserProfile | null>(null);
  const [menuResources, setMenuResources] = useState<ResourceItem[]>([]);
  const [profileOpen, setProfileOpen] = useState(false);
  const [passwordOpen, setPasswordOpen] = useState(false);
  const [profileSaving, setProfileSaving] = useState(false);
  const [passwordSaving, setPasswordSaving] = useState(false);
  const [menuLoading, setMenuLoading] = useState(true);
  const menuItems = useMemo(() => buildDynamicMenuItems(menuResources), [menuResources]);
  const pageResources = useMemo(
    () =>
      menuResources
        .filter((resource) => normalizeResourceType(resource.resourceType) === "page" && resource.pageUrl)
        .sort(sortResources),
    [menuResources],
  );
  const permittedUrls = useMemo(() => new Set(pageResources.map((resource) => resource.pageUrl)), [pageResources]);
  const quickActions = useMemo(() => {
    const preferredOrder = [
      "/manager-dashboard",
      "/user/maintenance",
      "/permission",
      "/app-user",
      "/grain/dashboard",
      "/grain/entries",
    ];
    const resourceByUrl = new Map(pageResources.map((resource) => [resource.pageUrl, resource]));
    return preferredOrder.flatMap((pageUrl) => {
      const resource = resourceByUrl.get(pageUrl);
      if (!resource) {
        return [];
      }
      return [
        {
          key: pageUrl,
          label: getResourceLabel(resource),
          icon: fallbackQuickActionIcons[pageUrl] ?? getResourceIcon(resource),
        },
      ];
    });
  }, [pageResources]);

  const currentModule =
    activePath.startsWith("/grain/farmers") || activePath.startsWith("/app-user")
      ? "粮户运营"
      : activePath.startsWith("/grain/dashboard") || activePath.startsWith("/grain/entries")
        ? "收粮业务"
        : activePath.startsWith("/grain")
          ? "粮站配置"
          : activePath.startsWith("/permission") || activePath.startsWith("/tenant")
            ? "系统配置"
            : activePath.startsWith("/user")
              ? "用户中心"
              : "经营总览";

  const selectedKey = activePath === "/activation-code" ? "/activation-code/admin" : activePath;

  const roleText = useMemo(() => {
    switch (profile?.role) {
      case "admin":
        return "系统管理员";
      case "manager":
        return "运营管理员";
      case "auditor":
        return "审计员";
      case "member":
        return "普通成员";
      default:
        return profile?.role || "管理账号";
    }
  }, [profile?.role]);

  const displayName = profile?.name || profile?.username || "管理用户";
  const avatarText = displayName.trim().charAt(0).toUpperCase() || "A";

  const loadProfile = useCallback(async () => {
    setMenuLoading(true);
    try {
      const result = await fetchCurrentUserProfile();
      setProfile(result);
      const resources = await fetchCurrentUserMenuResources();
      setMenuResources(resources);
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "获取个人信息失败");
    } finally {
      setMenuLoading(false);
    }
  }, [messageApi]);

  useEffect(() => {
    const pathOpenKeys = getOpenKeysFromMenuItems(activePath, menuItems);
    if (pathOpenKeys.length === 0) {
      return;
    }
    setOpenKeys((currentKeys) => Array.from(new Set([...currentKeys, ...pathOpenKeys])));
  }, [activePath, menuItems]);

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
      username: profile?.username ?? "",
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
        oldPassword: CryptoUtil.encrypt(values.oldPassword),
        newPassword: CryptoUtil.encrypt(values.newPassword),
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
    pageResources.find((resource) => activePath.startsWith(resource.pageUrl))?.menuName ||
    Object.entries(fallbackPageTitleMap).find(([path]) => activePath.startsWith(path))?.[1] ||
    "管理工作台";

  // mounted 后才用真实断点，避免 SSR 时 screens 为空导致判断错误
  const isMobile = mounted && !screens.md;

  const handleMenuClick = ({ key }: { key: string }) => {
    if (typeof key === "string" && permittedUrls.has(key)) {
      router.push(key);
      if (isMobile) {
        setCollapsed(true);
      }
    }
  };

  // 账号下拉菜单（PC 和移动端共用）
  const accountDropdownItems = [
    { key: "profile", icon: <UserOutlined />, label: "个人信息" },
    { key: "password", icon: <LockOutlined />, label: "修改密码" },
    { type: "divider" as const },
    { key: "logout", icon: <LogoutOutlined />, label: "退出登录" },
  ];

  const handleAccountMenu = ({ key }: { key: string }) => {
    if (key === "profile") openProfileModal();
    if (key === "password") openPasswordModal();
    if (key === "logout") handleLogout();
  };

  return (
    <div className="manager-app-frame">
      {contextHolder}

      {/* ===== 移动端：导航 Drawer ===== */}
      {isMobile && (
        <Drawer
          open={!collapsed}
          onClose={() => setCollapsed(true)}
          placement="left"
          width={272}
          styles={{
            header: { display: "none" },
            body: { padding: 0, background: "#10251a" },
            wrapper: { boxShadow: "4px 0 24px rgba(0,0,0,0.28)" },
          }}
          style={{ background: "#10251a" }}
        >
          <div
            className="manager-sidebar-card"
            style={{
              height: "100%",
              padding: "20px 14px",
              display: "flex",
              flexDirection: "column",
              gap: 14,
            }}
          >
            {/* 品牌 + 关闭 */}
            <div className="manager-sidebar-brand">
              <Space align="center" size={12}>
                <div className="manager-product-logo" aria-hidden="true">
                  <span />
                  <i />
                </div>
                <div className="manager-wordmark">
                  <strong style={{ color: "#fff" }}>收粮管理端</strong>
                  <span style={{ color: "rgba(255,255,255,0.66)" }}>Shennong Admin</span>
                </div>
              </Space>
              <Button
                type="text"
                icon={<MenuFoldOutlined />}
                className="manager-sidebar-close-btn"
                onClick={() => setCollapsed(true)}
              />
            </div>

            {/* 菜单 */}
            {menuLoading ? (
              <div className="manager-menu-loading" aria-label="菜单加载中" aria-busy="true">
                {menuLoadingRows.map((row, index) => (
                  <div
                    className="manager-menu-loading-row"
                    key={row.key}
                    style={{ animationDelay: `${index * 70}ms` }}
                  >
                    <span />
                    <i style={{ width: row.width }} />
                  </div>
                ))}
              </div>
            ) : (
              <Menu
                className="manager-shell-menu"
                mode="inline"
                selectedKeys={[selectedKey]}
                openKeys={openKeys}
                onOpenChange={(keys) => setOpenKeys(keys as string[])}
                items={menuItems}
                onClick={handleMenuClick}
                style={{ fontSize: 15, marginTop: 2 }}
              />
            )}

            {/* 底部状态 */}
            <div className="manager-sidebar-foot">
              <div className="manager-sidebar-status-icon">
                <SettingOutlined />
              </div>
              <div>
                <span>当前模块</span>
                <strong>{currentModule}</strong>
              </div>
            </div>
          </div>
        </Drawer>
      )}

      <div className={`manager-shell-surface ${collapsed ? "manager-shell-surface-collapsed" : ""}`}>
        <Layout style={{ minHeight: "100vh", background: "transparent" }}>

          {/* ===== PC 端 Sider ===== */}
          {!isMobile && (
            <Sider
              width={248}
              collapsedWidth={82}
              collapsible
              collapsed={collapsed}
              trigger={null}
              style={{ background: "transparent" }}
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
                      <strong style={{ color: "#fff" }}>收粮管理端</strong>
                      <span style={{ color: "rgba(255,255,255,0.66)" }}>Shennong Admin</span>
                    </div>
                  </Space>
                  <Tag bordered={false} className="manager-sidebar-env manager-collapse-hidden">
                    运营后台
                  </Tag>
                </div>

                {menuLoading ? (
                  <div className="manager-menu-loading" aria-label="菜单加载中" aria-busy="true">
                    {menuLoadingRows.map((row, index) => (
                      <div
                        className="manager-menu-loading-row"
                        key={row.key}
                        style={{ animationDelay: `${index * 70}ms` }}
                      >
                        <span />
                        <i style={{ width: row.width }} />
                      </div>
                    ))}
                  </div>
                ) : (
                  <Menu
                    className="manager-shell-menu"
                    mode="inline"
                    selectedKeys={[selectedKey]}
                    openKeys={openKeys}
                    onOpenChange={(keys) => setOpenKeys(keys as string[])}
                    items={menuItems}
                    onClick={handleMenuClick}
                    style={{ fontSize: 15, marginTop: 2 }}
                  />
                )}

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
          )}

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
                  padding: isMobile ? "0 14px" : "0 28px 0 30px",
                  minHeight: isMobile ? 56 : 76,
                  display: "grid",
                  gridTemplateColumns: "minmax(0, 1fr) auto",
                  gap: isMobile ? 10 : 20,
                  alignItems: "center",
                }}
              >
                {isMobile ? (
                  /* 移动端 Header：汉堡 + 页面标题 + 账号头像 */
                  <>
                    <div style={{ display: "flex", alignItems: "center", gap: 10, minWidth: 0 }}>
                      <Button
                        type="text"
                        icon={<MenuUnfoldOutlined />}
                        className="manager-nav-toggle"
                        onClick={() => setCollapsed(false)}
                      />
                      <Text
                        style={{
                          color: "var(--manager-text-soft)",
                          fontWeight: 700,
                          fontSize: 15,
                          overflow: "hidden",
                          textOverflow: "ellipsis",
                          whiteSpace: "nowrap",
                        }}
                      >
                        {pageTitle}
                      </Text>
                    </div>
                    <Dropdown trigger={["click"]} menu={{ items: accountDropdownItems, onClick: handleAccountMenu }}>
                      <Button
                        type="text"
                        style={{
                          height: 40,
                          padding: "4px 8px",
                          borderRadius: 8,
                          border: "1px solid var(--manager-border)",
                          background: "#ffffff",
                        }}
                      >
                        <Avatar
                          style={{
                            width: 30,
                            height: 30,
                            background: "linear-gradient(135deg, var(--manager-primary), #4f8f5f)",
                            color: "#fff",
                            fontWeight: 700,
                            fontSize: 13,
                          }}
                        >
                          {avatarText}
                        </Avatar>
                      </Button>
                    </Dropdown>
                  </>
                ) : (
                  /* PC 端 Header：原有布局 */
                  <>
                    <div className="manager-header-main">
                      <Space size={10} align="center" style={{ marginBottom: 8 }}>
                        <Button
                          type="text"
                          icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                          className="manager-nav-toggle"
                          onClick={() => setCollapsed((v) => !v)}
                        />
                        <CompassOutlined style={{ color: "var(--manager-primary)" }} />
                        <Text style={{ color: "var(--manager-text-soft)", fontWeight: 700 }}>
                          {pageTitle}
                        </Text>
                      </Space>
                      <Space size={10} wrap className="manager-quick-actions" style={{ width: "100%" }}>
                        {quickActions.map((action) => {
                          const isActive = activePath === action.key;
                          return (
                            <Button
                              key={action.key}
                              type={isActive ? "primary" : "default"}
                              icon={action.icon}
                              className={`manager-quick-action-btn${isActive ? " manager-soft-button" : ""}`}
                              onClick={() => router.push(action.key)}
                              style={{ height: 38, paddingInline: 14, borderRadius: 8, fontWeight: 700 }}
                            >
                              <span className="manager-quick-action-label">{action.label}</span>
                            </Button>
                          );
                        })}
                      </Space>
                    </div>

                    <Space size={12} wrap className="manager-header-account">
                      <Dropdown trigger={["click"]} menu={{ items: accountDropdownItems, onClick: handleAccountMenu }}>
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
                              <div style={{ fontWeight: 700, color: "var(--manager-text)" }}>{displayName}</div>
                              <Text style={{ color: "var(--manager-text-soft)" }}>{roleText}</Text>
                            </div>
                            <DownOutlined style={{ color: "var(--manager-text-soft)", fontSize: 12 }} />
                          </Space>
                        </Button>
                      </Dropdown>
                    </Space>
                  </>
                )}
              </div>
            </Header>

            <Content style={{ padding: isMobile ? "14px" : "22px 28px 40px" }}>
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
          <Form.Item
            name="username"
            label="登录账号"
            rules={[{ required: true, message: "请输入登录账号" }]}
          >
            <Input placeholder="请输入登录账号" />
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
