"use client";

import { useEffect, useMemo, useState } from "react";
import { Form, Input, Modal, Select, message } from "antd";
import { tenantApi, type TenantRecord } from "../../tenant/api/tenant.api";
import type { UserPayload, UserRecord } from "../api/user.api";

interface UserFormModalProps {
  open: boolean;
  submitting: boolean;
  user: UserRecord | null;
  showTenantField?: boolean;
  onCancel: () => void;
  onSubmit: (payload: UserPayload) => Promise<void>;
}

interface UserFormValues {
  username: string;
  tenantId?: number;
  remark?: string;
  password?: string;
}

export function UserFormModal({
  open,
  submitting,
  user,
  showTenantField = true,
  onCancel,
  onSubmit,
}: UserFormModalProps) {
  const [form] = Form.useForm<UserFormValues>();
  const isEdit = Boolean(user);
  const [tenants, setTenants] = useState<TenantRecord[]>([]);

  useEffect(() => {
    if (!open || !showTenantField) {
      return;
    }
    tenantApi
      .list({ pageIndex: 1, pageSize: 200, status: "active" })
      .then((result) => setTenants(result.data))
      .catch((error) => message.error(error instanceof Error ? error.message : "加载租户失败"));
  }, [open, showTenantField]);

  const tenantOptions = useMemo(
    () => tenants.map((tenant) => ({ label: tenant.tenantName, value: tenant.id })),
    [tenants],
  );

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
      onOk={async () => {
        const values = await form.validateFields();
        const payload: UserPayload = {
          username: values.username.trim(),
          name: values.username.trim(),
          tenantId: showTenantField ? values.tenantId : undefined,
          role: user?.role ?? "member",
          status: user?.status ?? "active",
          remark: values.remark?.trim(),
        };
        const password = values.password?.trim();
        if (password) {
          payload.password = password;
          payload.originPassword = password;
        }
        await onSubmit({
          ...payload,
        });
        form.resetFields();
      }}
      afterOpenChange={(visible) => {
        if (!visible) {
          form.resetFields();
          return;
        }
        form.setFieldsValue({
          username: user?.username ?? "",
          tenantId: user?.tenantId || undefined,
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
        {showTenantField ? (
          <Form.Item label="所属租户" name="tenantId">
            <Select allowClear options={tenantOptions} placeholder="请选择租户" />
          </Form.Item>
        ) : null}
        {!isEdit ? (
          <Form.Item
            label="密码"
            name="password"
            rules={[{ required: true, message: "请输入密码" }]}
          >
            <Input.Password placeholder="请输入密码" />
          </Form.Item>
        ) : null}
        <Form.Item label="备注" name="remark">
          <Input.TextArea rows={3} placeholder="请输入备注" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
