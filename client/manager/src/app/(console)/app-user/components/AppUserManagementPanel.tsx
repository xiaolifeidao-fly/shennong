"use client";

import { useState } from "react";
import { CheckCircleOutlined, IdcardOutlined, LockOutlined, ScanOutlined, StopOutlined } from "@ant-design/icons";
import { Button, Form, Input, Modal, Popconfirm, Space, Tooltip, Typography, Upload, message } from "antd";
import type { FormInstance } from "antd";
import type { RcFile } from "antd/es/upload";
import { CrudManagementPanel } from "../../components/CrudManagementPanel";
import type { CrudField, CrudFormSection, CrudTableColumn, CrudOption } from "../../components/CrudManagementPanel";
import { useAccessibleStations } from "../../grain/hooks/useAccessibleStations";
import { recognizeGrainCard, type GrainCardOcrResult } from "../../grain/api/grain.api";
import { SensitiveValue } from "../../grain/components/SensitiveValue";
import {
  createAppUser,
  deleteAppUser,
  fetchAppUsers,
  updateAppUser,
  updateAppUserPassword,
  updateAppUserStatus,
  type AppUserPayload,
  type AppUserRecord,
} from "../api/app-user.api";

const { Text } = Typography;

const statusOptions: CrudOption[] = [
  { label: "正常", value: "active" },
  { label: "禁用", value: "inactive" },
  { label: "锁定", value: "locked" },
];

const baseFields: CrudField<AppUserRecord>[] = [
  { name: "name", label: "姓名", required: true, placeholder: "请输入业务员真实姓名" },
  { name: "username", label: "用户名", required: true, placeholder: "用于登录管理端或小程序" },
  { name: "password", label: "登录密码", type: "password", required: true, hiddenOnEdit: true, placeholder: "至少 6 位" },
  { name: "originPassword", label: "初始密码", type: "password", hiddenOnEdit: true, placeholder: "不填则使用登录密码" },
  { name: "email", label: "邮箱", placeholder: "用于接收通知" },
  { name: "phone", label: "手机号", placeholder: "请输入常用手机号" },
  { name: "department", label: "部门", hiddenOnEdit: true, placeholder: "例如：粮食收购一组" },
  { name: "idNumber", label: "身份证号", placeholder: "可通过身份证识别自动回填（选填）" },
  { name: "status", label: "状态", type: "select", options: statusOptions, placeholder: "请选择账号状态" },
  { name: "remark", label: "备注", type: "textarea", span: 2, placeholder: "记录负责区域、交接说明或账号备注" },
];

const baseColumns: CrudTableColumn<AppUserRecord>[] = [
  { name: "username", label: "用户名", width: 150 },
  { name: "name", label: "姓名", width: 140 },
  { name: "phone", label: "手机号", width: 150 },
  {
    name: "idNumber",
    label: "身份证",
    width: 200,
    render: (value) => (value ? <SensitiveValue value={value} keepStart={6} keepEnd={4} /> : "-"),
  },
  { name: "email", label: "邮箱", width: 220 },
  { name: "status", label: "状态", width: 110 },
  { name: "banCount", label: "封禁次数", width: 110 },
  {
    name: "createdTime",
    label: "注册时间",
    width: 190,
    render: (value) => formatDateTime(value),
  },
];

function formatDateTime(value: unknown) {
  if (!value) {
    return "-";
  }
  const date = new Date(String(value));
  if (Number.isNaN(date.getTime())) {
    return String(value).slice(0, 19).replace("T", " ");
  }
  const pad = (num: number) => String(num).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

function beforeImageUpload(file: RcFile) {
  if (!file.type.startsWith("image/")) {
    message.error("请上传图片文件");
    return Upload.LIST_IGNORE;
  }
  if (file.size / 1024 / 1024 >= 8) {
    message.error("图片需小于 8MB");
    return Upload.LIST_IGNORE;
  }
  return true;
}

type IdCardOcrState = {
  frontLoading: boolean;
  backLoading: boolean;
};

function AppUserIdCardOcr({
  form,
  editingRecord,
}: {
  form: FormInstance;
  editingRecord: AppUserRecord | null;
}) {
  const [ocrState, setOcrState] = useState<IdCardOcrState>({ frontLoading: false, backLoading: false });

  const applyOcrResult = (result: GrainCardOcrResult, imageSide: "front" | "back") => {
    if (imageSide === "front") {
      form.setFieldsValue({
        name: result.name || form.getFieldValue("name"),
        idNumber: result.idNumber || form.getFieldValue("idNumber"),
        idCardFrontUrl: result.ossUrl || form.getFieldValue("idCardFrontUrl"),
        idCardFrontKey: result.ossObjectKey || form.getFieldValue("idCardFrontKey"),
      });
      message.success(result.mock ? "模拟身份证人像面识别完成" : "身份证人像面识别完成");
    } else {
      form.setFieldsValue({
        idCardBackUrl: result.ossUrl || form.getFieldValue("idCardBackUrl"),
        idCardBackKey: result.ossObjectKey || form.getFieldValue("idCardBackKey"),
      });
      message.success("身份证国徽面上传完成");
    }
  };

  const handleRecognize = async (file: File, imageSide: "front" | "back") => {
    const stationId = form.getFieldValue("stationId") as number | undefined;
    if (!stationId) {
      message.warning("请先选择粮站，再进行识别");
      return;
    }
    setOcrState((prev) => ({ ...prev, [`${imageSide}Loading`]: true }));
    try {
      const result = await recognizeGrainCard(file, {
        cardType: "id-card",
        imageSide,
        stationId,
        appUserId: editingRecord?.id,
      });
      applyOcrResult(result, imageSide);
    } catch (error) {
      message.error(error instanceof Error ? error.message : "识别失败，请手动录入");
    } finally {
      setOcrState((prev) => ({ ...prev, [`${imageSide}Loading`]: false }));
    }
  };

  const renderUploadButton = (imageSide: "front" | "back", label: string) => {
    const loading = imageSide === "front" ? ocrState.frontLoading : ocrState.backLoading;
    return (
      <Upload
        accept="image/*"
        maxCount={1}
        showUploadList={false}
        beforeUpload={beforeImageUpload}
        customRequest={({ file, onSuccess }) => {
          void handleRecognize(file as File, imageSide).then(() => onSuccess?.("ok"));
        }}
        disabled={loading}
      >
        <Button icon={imageSide === "front" ? <ScanOutlined /> : <IdcardOutlined />} loading={loading}>
          {label}
        </Button>
      </Upload>
    );
  };

  return (
    <section className="manager-form-assist">
      <Space align="start" style={{ width: "100%", justifyContent: "space-between" }} wrap>
        <Space direction="vertical" size={2}>
          <Text strong>身份证识别（选填）</Text>
          <Text type="secondary">
            上传身份证人像面可自动回填姓名和身份证号；身份证图片将安全存储，识别结果仍可手动修正。
          </Text>
        </Space>
        <Space wrap>
          {renderUploadButton("front", "身份证人像面")}
          {renderUploadButton("back", "身份证国徽面")}
        </Space>
      </Space>
    </section>
  );
}

export function AppUserManagementPanel() {
  const { stations: accessibleStations } = useAccessibleStations();
  const stationOptions: CrudOption[] = accessibleStations.map((s) => ({ label: s.stationName, value: s.id }));
  const [passwordForm] = Form.useForm<{ password: string; confirmPassword: string }>();
  const [passwordModalOpen, setPasswordModalOpen] = useState(false);
  const [passwordRecord, setPasswordRecord] = useState<AppUserRecord | null>(null);
  const [passwordSubmitting, setPasswordSubmitting] = useState(false);

  const stationLabel = (stationId: unknown, record: AppUserRecord) =>
    record.stationName || stationOptions.find((option) => option.value === stationId)?.label || String(stationId || "-");

  const fields: CrudField<AppUserRecord>[] = [
    {
      name: "stationId",
      label: "粮站",
      type: "select",
      required: true,
      placeholder: "请选择粮站",
      options: stationOptions,
    },
    ...baseFields,
  ];

  const formSections: CrudFormSection<AppUserRecord>[] = [
    {
      key: "assignment",
      title: "任职归属",
      description: "明确业务员所属粮站和组织信息。",
      fields: ["stationId", "department"],
    },
    {
      key: "account",
      title: "账号安全",
      description: "设置登录身份，编辑时密码通过独立操作修改。",
      fields: ["name", "username", "password", "originPassword"],
    },
    {
      key: "contact",
      title: "联系方式",
      description: "用于运营联系、通知和后续服务追踪。",
      fields: ["phone", "email"],
    },
    {
      key: "identity",
      title: "身份证信息",
      description: "身份证号加密存储，选填。",
      fields: ["idNumber"],
    },
    {
      key: "status",
      title: "状态备注",
      description: "控制账号可用状态并补充内部说明。",
      fields: ["status", "remark"],
    },
  ];

  const columns: CrudTableColumn<AppUserRecord>[] = [
    { name: "stationId", label: "粮站", width: 180, render: stationLabel },
    ...baseColumns,
  ];

  return (
    <>
      <CrudManagementPanel<AppUserRecord, AppUserPayload>
        title="业务员管理"
        createText="新增业务员"
        searchPlaceholder="用户名"
        searchParam="search"
        extraFilters={[
          { param: "name", placeholder: "业务员姓名" },
          {
            param: "stationId",
            placeholder: "粮站",
            type: "select",
            options: stationOptions,
            width: 220,
          },
        ]}
        fields={fields}
        columns={columns}
        statusField="status"
        statusOptions={statusOptions}
        actionWidth={200}
        modalWidth={940}
        formSections={formSections}
        formExtra={({ form, editingRecord }) => (
          <AppUserIdCardOcr form={form} editingRecord={editingRecord} />
        )}
        rowActions={(record, context) => {
          const isActive = record.status === "active";
          const nextStatus = isActive ? "inactive" : "active";
          const actionText = isActive ? "禁用" : "启用";

          return (
            <>
              <Tooltip title="修改密码">
                <Button
                  type="text"
                  icon={<LockOutlined />}
                  onClick={() => {
                    setPasswordRecord(record);
                    passwordForm.resetFields();
                    setPasswordModalOpen(true);
                  }}
                />
              </Tooltip>
              <Popconfirm
                title={`确认${actionText}用户「${record.username}」吗？`}
                okText={actionText}
                cancelText="取消"
                onConfirm={async () => {
                  try {
                    context.setSubmitting(true);
                    await updateAppUserStatus(record.id, nextStatus);
                    message.success(`${actionText}成功`);
                    await context.reload();
                  } catch (error) {
                    message.error(error instanceof Error ? error.message : `${actionText}失败`);
                  } finally {
                    context.setSubmitting(false);
                  }
                }}
              >
                <Tooltip title={actionText}>
                  <Button
                    type="text"
                    danger={isActive}
                    disabled={context.submitting}
                    icon={isActive ? <StopOutlined /> : <CheckCircleOutlined />}
                  />
                </Tooltip>
              </Popconfirm>
            </>
          );
        }}
        api={{
          list: fetchAppUsers,
          create: createAppUser,
          update: updateAppUser,
          remove: deleteAppUser,
        }}
      />

      <Modal
        title={passwordRecord ? `修改密码：${passwordRecord.username}` : "修改密码"}
        open={passwordModalOpen}
        okText="保存"
        cancelText="取消"
        confirmLoading={passwordSubmitting}
        destroyOnClose
        onCancel={() => {
          setPasswordModalOpen(false);
          setPasswordRecord(null);
          passwordForm.resetFields();
        }}
        onOk={() => {
          void passwordForm.validateFields().then(async (values) => {
            if (!passwordRecord) {
              return;
            }
            setPasswordSubmitting(true);
            try {
              await updateAppUserPassword(passwordRecord.id, values.password);
              message.success("密码修改成功");
              setPasswordModalOpen(false);
              setPasswordRecord(null);
              passwordForm.resetFields();
            } catch (error) {
              message.error(error instanceof Error ? error.message : "密码修改失败");
            } finally {
              setPasswordSubmitting(false);
            }
          });
        }}
      >
        <Form form={passwordForm} layout="vertical" preserve={false} style={{ marginTop: 16 }}>
          <Form.Item
            name="password"
            label="新密码"
            rules={[
              { required: true, message: "请输入新密码" },
              { min: 6, message: "密码至少 6 位" },
            ]}
          >
            <Input.Password placeholder="请输入新密码" />
          </Form.Item>
          <Form.Item
            name="confirmPassword"
            label="确认密码"
            dependencies={["password"]}
            rules={[
              { required: true, message: "请再次输入新密码" },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue("password") === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error("两次输入的密码不一致"));
                },
              }),
            ]}
          >
            <Input.Password placeholder="请再次输入新密码" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
