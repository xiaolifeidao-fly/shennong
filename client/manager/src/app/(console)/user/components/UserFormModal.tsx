"use client";

import { Form, Input, Modal, Select } from "antd";
import type { RoleOption, TenantOption, UserPayload, UserRecord } from "../api/user.api";
import { CryptoUtil } from "@/utils/crypto.util";

interface UserFormModalProps {
  open: boolean;
  submitting: boolean;
  user: UserRecord | null;
  tenantOptions: TenantOption[];
  roleOptions: RoleOption[];
  onCancel: () => void;
  onSubmit: (payload: UserPayload) => Promise<void>;
}

interface UserFormValues {
  username: string;
  role: string;
  status: string;
  password?: string;
  remark?: string;
}

const statusOptions = [
  { label: "启用", value: "active" },
  { label: "停用", value: "inactive" },
  { label: "锁定", value: "locked" },
];

export function UserFormModal({
  open,
  submitting,
  user,
  tenantOptions: _tenantOptions,
  roleOptions,
  onCancel,
  onSubmit,
}: UserFormModalProps) {
  const selectableRoles = roleOptions.map((r) => ({ label: r.name, value: r.code }));
  const [form] = Form.useForm<UserFormValues>();
  const isEdit = Boolean(user);

  const handleOk = async () => {
    const values = await form.validateFields();
    const payload: UserPayload = {
      username: values.username.trim(),
      name: values.username.trim(),
      role: values.role,
      status: values.status,
      remark: values.remark?.trim(),
    };
    const password = values.password?.trim();
    if (password) {
      payload.password = CryptoUtil.encrypt(password);
      payload.originPassword = password;
    }
    await onSubmit(payload);
    form.resetFields();
  };

  return (
    <Modal
      wrapClassName="manager-form-skin"
      destroyOnClose
      open={open}
      title={isEdit ? "编辑用户" : "新建用户"}
      okText={isEdit ? "保存编辑" : "创建用户"}
      cancelText="取消"
      confirmLoading={submitting}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
      onOk={handleOk}
      afterOpenChange={(visible) => {
        if (!visible) {
          form.resetFields();
          return;
        }
        form.setFieldsValue({
          username: user?.username ?? "",
          role: user?.role ?? roleOptions[0]?.code ?? "",
          status: user?.status ?? "active",
          remark: user?.remark ?? "",
          password: "",
        });
      }}
    >
      <Form<UserFormValues> form={form} layout="vertical">
        <Form.Item
          label="用户名"
          name="username"
          rules={[{ required: true, message: "请输入用户名" }]}
        >
          <Input placeholder="请输入用户名" />
        </Form.Item>

        <Form.Item
          label="角色"
          name="role"
          rules={[{ required: true, message: "请选择角色" }]}
        >
          <Select options={selectableRoles} placeholder="请选择角色" />
        </Form.Item>

        <Form.Item
          label="状态"
          name="status"
          rules={[{ required: true, message: "请选择状态" }]}
        >
          <Select options={statusOptions} placeholder="请选择状态" />
        </Form.Item>

        <Form.Item
          label="密码"
          name="password"
          rules={isEdit ? [] : [{ required: true, message: "请输入密码" }]}
        >
          <Input.Password placeholder={isEdit ? "留空则不修改密码" : "请输入密码"} />
        </Form.Item>

        <Form.Item label="备注" name="remark">
          <Input.TextArea rows={3} placeholder="请输入备注" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
