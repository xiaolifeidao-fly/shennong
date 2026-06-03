"use client";

import { useEffect, useState } from "react";
import {
  BankOutlined,
  CameraOutlined,
  FileImageOutlined,
  IdcardOutlined,
  ScanOutlined,
} from "@ant-design/icons";
import { Button, Descriptions, Drawer, Empty, Form, Image, InputNumber, message, Space, Tag, Tooltip, Typography, Upload } from "antd";
import type { FormInstance } from "antd";
import type { RcFile } from "antd/es/upload";
import {
  CrudManagementPanel,
  type CrudField,
  type CrudFormSection,
  type CrudOption,
  type CrudTableColumn,
} from "../../components/CrudManagementPanel";
import {
  getGrainFarmerImages,
  grainFarmerApi,
  grainStationApi,
  recognizeGrainCard,
  type GrainCardOcrResult,
  type GrainFarmerImagesRecord,
  type GrainFarmerRecord,
  type GrainPayload,
} from "../api/grain.api";
import { SensitiveValue } from "./SensitiveValue";
import { getCurrentAppUser, type StoredCurrentAppUser } from "@/utils/auth";

const { Text } = Typography;

const statusOptions: CrudOption[] = [
  { label: "资料完整", value: "complete" },
  { label: "银行卡待补", value: "missing-bank" },
  { label: "停用", value: "inactive" },
];

const fields: CrudField<GrainFarmerRecord>[] = [
  { name: "name", label: "农户姓名", required: true, placeholder: "请输入身份证上的姓名" },
  { name: "idNumber", label: "身份证号", placeholder: "可通过身份证识别自动回填" },
  { name: "phone", label: "手机号", placeholder: "请输入农户联系电话" },
  { name: "address", label: "户籍或联系地址", span: 2, placeholder: "可通过身份证识别自动回填" },
  { name: "bankNumber", label: "银行卡号", placeholder: "可通过银行卡识别自动回填" },
  { name: "bankName", label: "开户行", placeholder: "请输入开户银行名称" },
  { name: "status", label: "资料状态", type: "select", options: statusOptions, placeholder: "请选择资料状态" },
  { name: "statusText", label: "状态说明", placeholder: "例如：银行卡待补充" },
  { name: "remark", label: "备注", type: "textarea", span: 2, placeholder: "记录补充材料、沟通情况或特殊说明" },
];

const columns: CrudTableColumn<GrainFarmerRecord>[] = [
  { name: "name", label: "农户姓名", width: 140 },
  { name: "phone", label: "手机号", width: 140 },
  {
    name: "idNumber",
    label: "身份证",
    width: 220,
    copyable: true,
    render: (value) => <SensitiveValue value={value} keepStart={6} keepEnd={4} />,
  },
  {
    name: "bankNumber",
    label: "银行卡",
    width: 220,
    copyable: true,
    render: (value, record) => (
      <Space direction="vertical" size={0}>
        <SensitiveValue value={value} keepStart={4} keepEnd={4} />
        <Text type="secondary" style={{ fontSize: 12 }}>
          {record.bankName || "未填写开户行"}
        </Text>
      </Space>
    ),
  },
  { name: "status", label: "状态", width: 120 },
];

type OcrTarget = {
  title: string;
  cardType: "id-card" | "bank-card";
  imageSide?: "front" | "back";
};

function renderValue(value: unknown) {
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  return String(value);
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

export function GrainFarmerPanel() {
  const [stationOptions, setStationOptions] = useState<CrudOption[]>([]);
  const [archiveOpen, setArchiveOpen] = useState(false);
  const [archiveRecord, setArchiveRecord] = useState<GrainFarmerRecord | null>(null);
  const [archiveImages, setArchiveImages] = useState<GrainFarmerImagesRecord | null>(null);
  const [archiveLoading, setArchiveLoading] = useState(false);
  const [ocrLoadingKey, setOcrLoadingKey] = useState<string | null>(null);
  const [currentAppUser] = useState<StoredCurrentAppUser | null>(() => getCurrentAppUser());

  useEffect(() => {
    grainStationApi
      .list({ pageIndex: 1, pageSize: 200, status: "active" })
      .then((stations) => {
        setStationOptions(
          stations.data.map((station) => ({
            label: station.stationName,
            value: station.id,
          })),
        );
      })
      .catch((error) => message.error(error instanceof Error ? error.message : "加载农户选项失败"));
  }, []);

  const stationLabel = (stationId: unknown) =>
    stationOptions.find((option) => option.value === stationId)?.label ?? String(stationId || "-");
  const currentStationOption = stationOptions[0];
  const currentStationId = typeof currentStationOption?.value === "number" ? currentStationOption.value : undefined;
  const currentAppUserId =
    typeof currentAppUser?.id === "number" && currentAppUser.id > 0 ? currentAppUser.id : undefined;

  const openArchive = async (record: GrainFarmerRecord) => {
    setArchiveRecord(record);
    setArchiveImages(null);
    setArchiveOpen(true);
    setArchiveLoading(true);
    try {
      const images = await getGrainFarmerImages(record.id);
      setArchiveImages(images);
    } catch (error) {
      message.warning(error instanceof Error ? error.message : "证件图片加载失败");
    } finally {
      setArchiveLoading(false);
    }
  };

  const applyOcrResult = (form: FormInstance, result: GrainCardOcrResult) => {
    if (result.cardType === "id-card") {
      form.setFieldsValue({
        name: result.name || form.getFieldValue("name"),
        idNumber: result.idNumber || form.getFieldValue("idNumber"),
        address: result.address || form.getFieldValue("address"),
      });
      message.success(result.mock ? "模拟身份证识别完成" : "身份证识别完成");
      return;
    }
    form.setFieldsValue({
      bankNumber: result.bankNumber || form.getFieldValue("bankNumber"),
      bankName: result.bankName || form.getFieldValue("bankName"),
      status: result.bankNumber ? "complete" : form.getFieldValue("status"),
      statusText: result.bankNumber ? "资料完整" : form.getFieldValue("statusText"),
    });
    message.success(result.mock ? "模拟银行卡识别完成" : "银行卡识别完成");
  };

  const handleRecognize = async (file: File, target: OcrTarget, form: FormInstance, editingRecord: GrainFarmerRecord | null) => {
    const stationId = form.getFieldValue("stationId");
    if (!stationId) {
      message.warning("请先选择粮站，再进行识别");
      return;
    }
    const key = `${target.cardType}-${target.imageSide ?? "front"}`;
    setOcrLoadingKey(key);
    try {
      const result = await recognizeGrainCard(file, {
        cardType: target.cardType,
        imageSide: target.imageSide,
        stationId,
        farmerId: editingRecord?.id,
      });
      applyOcrResult(form, result);
    } catch (error) {
      message.error(error instanceof Error ? error.message : "识别失败，请手动录入");
    } finally {
      setOcrLoadingKey(null);
    }
  };

  const renderOcrUpload = (target: OcrTarget, form: FormInstance, editingRecord: GrainFarmerRecord | null) => {
    const key = `${target.cardType}-${target.imageSide ?? "front"}`;
    return (
      <Upload
        accept="image/*"
        maxCount={1}
        showUploadList={false}
        beforeUpload={beforeImageUpload}
        customRequest={({ file, onSuccess }) => {
          void handleRecognize(file as File, target, form, editingRecord).then(() => onSuccess?.("ok"));
        }}
        disabled={ocrLoadingKey === key}
      >
        <Button icon={<ScanOutlined />} loading={ocrLoadingKey === key}>
          {target.title}
        </Button>
      </Upload>
    );
  };

  const imageCards = [
    { label: "身份证人像面", value: archiveImages?.idCardFront, icon: <IdcardOutlined /> },
    { label: "银行卡", value: archiveImages?.bankCard, icon: <BankOutlined /> },
  ];

  const farmerFields: CrudField<GrainFarmerRecord>[] = [
    {
      name: "stationId",
      label: "粮站",
      type: "select",
      required: true,
      placeholder: currentStationOption ? undefined : "当前业务员未绑定粮站",
      options: currentStationOption ? [currentStationOption] : [],
      disabled: true,
      help: currentStationOption ? "由当前业务员所属粮站自动填充" : undefined,
    },
    ...fields,
  ];

  const farmerFormSections: CrudFormSection<GrainFarmerRecord>[] = [
    {
      key: "ownership",
      title: "业务归属",
      description: "确定农户所属粮站，便于后续收购记录归档。",
      fields: ["stationId"],
    },
    {
      key: "identity",
      title: "身份资料",
      description: "维护农户实名信息和常用联系方式。",
      fields: ["name", "idNumber", "phone", "address"],
    },
    {
      key: "bank",
      title: "收款账户",
      description: "保存银行卡号与开户行，用于粮款结算核对。",
      fields: ["bankNumber", "bankName"],
    },
    {
      key: "status",
      title: "状态备注",
      description: "标记资料完整度，记录补充说明。",
      fields: ["status", "statusText", "remark"],
    },
  ];

  const farmerColumns: CrudTableColumn<GrainFarmerRecord>[] = [
    { name: "stationId", label: "粮站", width: 180, render: stationLabel },
    ...columns,
  ];

  return (
    <>
    <CrudManagementPanel<GrainFarmerRecord, GrainPayload>
      title="粮户"
      createText="新增粮户"
      searchPlaceholder="姓名/手机号/身份证/地址"
      searchParam="search"
      fields={farmerFields}
      columns={farmerColumns}
      statusField="status"
      statusOptions={statusOptions}
      actionWidth={180}
      modalWidth={920}
      showDrawerHeader={false}
      showDrawerRecordTag={false}
      showStats={false}
      initialValues={{ stationId: currentStationId, appUserId: currentAppUserId } as Partial<GrainFarmerRecord>}
      rowActions={(record) => (
        <Tooltip title="证件档案">
          <Button type="text" icon={<FileImageOutlined />} onClick={() => void openArchive(record)} />
        </Tooltip>
      )}
      formExtra={({ form, editingRecord }) => (
        <section className="manager-form-assist">
          <Form.Item name="appUserId" hidden>
            <InputNumber />
          </Form.Item>
          <Space align="start" style={{ width: "100%", justifyContent: "space-between" }} wrap>
            <Space direction="vertical" size={2}>
              <Text strong>证件与银行卡识别</Text>
              <Text type="secondary">
                上传身份证人像面或银行卡照片可自动回填表单；识别结果仍可手动修正后保存。
              </Text>
            </Space>
            <Space wrap>
              {renderOcrUpload({ title: "身份证人像面", cardType: "id-card", imageSide: "front" }, form, editingRecord)}
              {renderOcrUpload({ title: "银行卡", cardType: "bank-card" }, form, editingRecord)}
            </Space>
          </Space>
        </section>
      )}
      api={grainFarmerApi}
      formSections={farmerFormSections}
    />
    <Drawer
      title={archiveRecord ? `证件档案：${archiveRecord.name}` : "证件档案"}
      open={archiveOpen}
      onClose={() => {
        setArchiveOpen(false);
        setArchiveRecord(null);
        setArchiveImages(null);
      }}
      width={520}
    >
      {archiveRecord ? (
        <Space direction="vertical" size={18} style={{ width: "100%" }}>
          <Descriptions bordered column={1} size="small">
            <Descriptions.Item label="农户姓名">{renderValue(archiveRecord.name)}</Descriptions.Item>
            <Descriptions.Item label="身份证号">
              <SensitiveValue value={archiveRecord.idNumber} keepStart={6} keepEnd={4} />
            </Descriptions.Item>
            <Descriptions.Item label="银行卡号">
              <SensitiveValue value={archiveRecord.bankNumber} keepStart={4} keepEnd={4} />
            </Descriptions.Item>
            <Descriptions.Item label="开户行">{renderValue(archiveRecord.bankName)}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={archiveRecord.status === "inactive" ? "red" : archiveRecord.status === "missing-bank" ? "orange" : "green"}>
                {statusOptions.find((item) => item.value === archiveRecord.status)?.label ?? renderValue(archiveRecord.status)}
              </Tag>
            </Descriptions.Item>
          </Descriptions>

          <section className="manager-data-card" style={{ padding: 16, boxShadow: "none" }}>
            <Space align="center" style={{ marginBottom: 12 }}>
              <CameraOutlined style={{ color: "var(--manager-primary)" }} />
              <Text strong>图片资料</Text>
            </Space>
            {archiveLoading ? (
              <Text type="secondary">正在加载证件图片...</Text>
            ) : imageCards.some((item) => item.value) ? (
              <div style={{ display: "grid", gap: 14 }}>
                {imageCards.map((item) =>
                  item.value ? (
                    <div key={item.label}>
                      <Space size={8} style={{ marginBottom: 6 }}>
                        {item.icon}
                        <Text type="secondary">{item.label}</Text>
                      </Space>
                      <Image
                        src={item.value}
                        alt={item.label}
                        style={{ width: "100%", maxHeight: 190, objectFit: "cover", borderRadius: 8 }}
                      />
                    </div>
                  ) : null,
                )}
              </div>
            ) : (
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无身份证或银行卡图片" />
            )}
          </section>
        </Space>
      ) : null}
    </Drawer>
    </>
  );
}
