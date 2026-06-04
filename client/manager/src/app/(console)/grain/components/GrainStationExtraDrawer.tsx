"use client";

import { useEffect, useState } from "react";
import { Button, Drawer, Form, Image, Input, message, Space, Typography, Upload } from "antd";
import { InboxOutlined, LoadingOutlined } from "@ant-design/icons";
import type { RcFile } from "antd/es/upload";
import { getStationExtra, saveStationExtra, uploadBusinessLicense, type GrainStationExtraRecord } from "../api/grain.api";

const { Text } = Typography;
const { Dragger } = Upload;

interface Props {
  open: boolean;
  stationId: number | null;
  stationName: string;
  onClose: () => void;
}

function beforeUpload(file: RcFile) {
  const isImage = file.type.startsWith("image/") || file.type === "application/pdf";
  if (!isImage) {
    message.error("请上传图片或 PDF 文件");
    return Upload.LIST_IGNORE;
  }
  if (file.size / 1024 / 1024 >= 10) {
    message.error("文件需小于 10MB");
    return Upload.LIST_IGNORE;
  }
  return true;
}

export function GrainStationExtraDrawer({ open, stationId, stationName, onClose }: Props) {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [extra, setExtra] = useState<GrainStationExtraRecord | null>(null);

  useEffect(() => {
    if (!open || !stationId) return;
    setLoading(true);
    getStationExtra(stationId)
      .then((data) => {
        setExtra(data);
        form.setFieldsValue({
          accountHolderName: data.accountHolderName,
          bankName: data.bankName,
          bankAccountNumber: data.bankAccountNumber,
        });
      })
      .catch((err) => message.error(err instanceof Error ? err.message : "加载失败"))
      .finally(() => setLoading(false));
  }, [open, stationId, form]);

  const handleSave = async () => {
    if (!stationId) return;
    const values = await form.validateFields();
    setLoading(true);
    try {
      const result = await saveStationExtra(stationId, values);
      setExtra(result);
      message.success("保存成功");
    } catch (err) {
      message.error(err instanceof Error ? err.message : "保存失败");
    } finally {
      setLoading(false);
    }
  };

  const handleUpload = async (file: File) => {
    if (!stationId) return;
    setUploading(true);
    try {
      const result = await uploadBusinessLicense(stationId, file);
      setExtra(result);
      message.success("营业执照上传成功");
    } catch (err) {
      message.error(err instanceof Error ? err.message : "上传失败");
    } finally {
      setUploading(false);
    }
  };

  const handleClose = () => {
    form.resetFields();
    setExtra(null);
    onClose();
  };

  return (
    <Drawer
      title={`证件材料：${stationName}`}
      open={open}
      onClose={handleClose}
      width={520}
      footer={
        <Space style={{ justifyContent: "flex-end", width: "100%" }}>
          <Button onClick={handleClose}>取消</Button>
          <Button type="primary" loading={loading} onClick={handleSave}>
            保存
          </Button>
        </Space>
      }
    >
      <Space direction="vertical" size={24} style={{ width: "100%" }}>
        <section>
          <Text strong style={{ display: "block", marginBottom: 12 }}>
            营业执照
          </Text>
          {extra?.businessLicenseUrl && (
            <div style={{ marginBottom: 12 }}>
              <Image
                src={extra.businessLicenseUrl}
                alt="营业执照"
                style={{ maxWidth: "100%", maxHeight: 220, objectFit: "contain", borderRadius: 8 }}
              />
            </div>
          )}
          <Dragger
            name="file"
            accept="image/*,.pdf"
            maxCount={1}
            showUploadList={false}
            beforeUpload={beforeUpload}
            customRequest={({ file, onSuccess }) => {
              void handleUpload(file as File).then(() => onSuccess?.("ok"));
            }}
            disabled={uploading}
            style={{ padding: "12px 0" }}
          >
            <p className="ant-upload-drag-icon">
              {uploading ? <LoadingOutlined /> : <InboxOutlined />}
            </p>
            <p className="ant-upload-text">{uploading ? "上传中..." : "点击或拖拽上传营业执照"}</p>
            <p className="ant-upload-hint">支持图片或 PDF，文件大小不超过 10MB</p>
          </Dragger>
        </section>

        <section>
          <Text strong style={{ display: "block", marginBottom: 12 }}>
            银行开户信息
          </Text>
          <Form form={form} layout="vertical">
            <Form.Item label="户名" name="accountHolderName">
              <Input placeholder="请输入开户户名" />
            </Form.Item>
            <Form.Item label="开户行" name="bankName">
              <Input placeholder="请输入开户银行名称" />
            </Form.Item>
            <Form.Item label="开户号" name="bankAccountNumber">
              <Input placeholder="请输入银行账号" />
            </Form.Item>
          </Form>
        </section>
      </Space>
    </Drawer>
  );
}
