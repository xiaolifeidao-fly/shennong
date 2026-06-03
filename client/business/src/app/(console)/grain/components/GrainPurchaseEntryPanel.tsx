"use client";

import { useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import {
  BankOutlined,
  CameraOutlined,
  EditOutlined,
  EnvironmentOutlined,
  FileImageOutlined,
  HistoryOutlined,
  IdcardOutlined,
  PlusOutlined,
  ReloadOutlined,
  ScanOutlined,
  SearchOutlined,
  StopOutlined,
} from "@ant-design/icons";
import {
  Alert,
  Button,
  Descriptions,
  Divider,
  Drawer,
  Empty,
  Form,
  Image,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
  Tooltip,
  Timeline,
  Typography,
  Upload,
  message,
} from "antd";
import type { RcFile } from "antd/es/upload";
import type { ColumnsType } from "antd/es/table";
import type { CrudListQuery, CrudOption } from "../../components/CrudManagementPanel";
import {
  getGrainFarmerImages,
  grainEntryMaterialApi,
  grainFarmerApi,
  grainPurchaseEntryApi,
  grainPurchaseEntrySnapshotApi,
  grainStationApi,
  listActiveGrainPaymentMethods,
  listActiveGrainPurchaseTypes,
  recognizeGrainCard,
  voidGrainPurchaseEntry,
  type GrainCardOcrResult,
  type GrainFarmerImagesRecord,
  type GrainFarmerRecord,
  type GrainPayload,
  type GrainEntryMaterialRecord,
  type GrainPurchaseEntryRecord,
  type GrainPurchaseEntrySnapshotRecord,
} from "../api/grain.api";
import { getCurrentAppUser, type StoredCurrentAppUser } from "@/utils/auth";
import { SensitiveValue } from "./SensitiveValue";

const { Text } = Typography;

type EntryFormValues = {
  stationId?: number;
  farmerId?: number;
  farmerName?: string;
  farmerPhone?: string;
  farmerIdNumber?: string;
  farmerAddress?: string;
  farmerBankNumber?: string;
  farmerBankName?: string;
  purchaseTypeId?: number;
  crop?: string;
  quantity?: number;
  unit?: string;
  amount?: number;
  unitPrice?: number;
  locationAddress?: string;
  paymentMethodId?: number;
  payType?: string;
  status?: string;
  remark?: string;
};

type OcrTarget = {
  title: string;
  cardType: "id-card" | "bank-card";
  imageSide?: "front" | "back";
};

type ChangeRow = {
  key: string;
  label: string;
  before: ReactNode;
  after: ReactNode;
};

type PaymentMethodOption = CrudOption & {
  methodCode?: string;
};

const statusOptions: CrudOption[] = [
  { label: "已提交", value: "submitted" },
  { label: "已作废", value: "voided" },
];

const pageSize = 10;

const entryFormSections = [
  { key: "ownership", title: "基础信息" },
  { key: "grain-farmer", title: "粮食与农户信息" },
  { key: "payment", title: "收付款信息" },
  { key: "materials", title: "图片材料" },
] as const;

function renderValue(value: unknown) {
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  return String(value);
}

function statusTag(value: unknown) {
  const option = statusOptions.find((item) => item.value === value);
  return <Tag color={value === "voided" ? "red" : "green"}>{option?.label ?? renderValue(value)}</Tag>;
}

function compactPayload(values: Record<string, unknown>) {
  return Object.fromEntries(Object.entries(values).filter(([, value]) => value !== undefined && value !== ""));
}

function normalizeMoney(value?: number) {
  return typeof value === "number" && Number.isFinite(value) ? value.toFixed(2) : "-";
}

function formatTime(value?: string) {
  if (!value) {
    return "-";
  }
  return value.replace("T", " ").slice(0, 19);
}

function isBankPaymentMethod(option?: PaymentMethodOption, payType?: string) {
  const text = `${option?.methodCode || ""} ${option?.label || ""} ${payType || ""}`.toLowerCase();
  return text.includes("bank") || text.includes("银行卡") || text.includes("银行");
}

function receiptFieldMeta(isBankPayment: boolean, payType?: string) {
  if (isBankPayment) {
    return {
      accountLabel: "银行卡号",
      accountPlaceholder: "拍银行卡自动带入，或手动输入银行卡号",
      nameLabel: "开户行",
      namePlaceholder: "请输入开户行",
    };
  }
  const methodName = payType || "收款";
  return {
    accountLabel: "收款账号",
    accountPlaceholder: `请输入${methodName}账号`,
    nameLabel: "收款人姓名",
    namePlaceholder: "请输入收款人姓名",
  };
}

function actionMeta(action: string) {
  switch (action) {
    case "create":
      return { label: "创建", color: "green" };
    case "update":
      return { label: "修改", color: "blue" };
    case "void":
      return { label: "作废", color: "red" };
    case "delete":
      return { label: "删除", color: "red" };
    default:
      return { label: renderValue(action), color: "default" };
  }
}

function snapshotStatusTag(value: unknown) {
  return statusTag(value);
}

export function GrainPurchaseEntryPanel() {
  const [form] = Form.useForm<EntryFormValues>();
  const [records, setRecords] = useState<GrainPurchaseEntryRecord[]>([]);
  const [total, setTotal] = useState(0);
  const [query, setQuery] = useState({ pageIndex: 1, pageSize });
  const [searchValue, setSearchValue] = useState("");
  const [statusValue, setStatusValue] = useState<string | number | undefined>();
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<GrainPurchaseEntryRecord | null>(null);
  const [stationOptions, setStationOptions] = useState<CrudOption[]>([]);
  const [purchaseTypeOptions, setPurchaseTypeOptions] = useState<CrudOption[]>([]);
  const [paymentMethodOptions, setPaymentMethodOptions] = useState<PaymentMethodOption[]>([]);
  const [farmerImages, setFarmerImages] = useState<GrainFarmerImagesRecord | null>(null);
  const [entryMaterials, setEntryMaterials] = useState<GrainEntryMaterialRecord[]>([]);
  const [ocrTarget, setOcrTarget] = useState<OcrTarget | null>(null);
  const [ocrLoading, setOcrLoading] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyRecord, setHistoryRecord] = useState<GrainPurchaseEntryRecord | null>(null);
  const [historySnapshots, setHistorySnapshots] = useState<GrainPurchaseEntrySnapshotRecord[]>([]);
  const [selectedSnapshotId, setSelectedSnapshotId] = useState<number | null>(null);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [currentAppUser] = useState<StoredCurrentAppUser | null>(() => getCurrentAppUser());

  const watchedQuantity = Form.useWatch("quantity", form);
  const watchedAmount = Form.useWatch("amount", form);
  const watchedUnit = Form.useWatch("unit", form);
  const watchedUnitPrice = Form.useWatch("unitPrice", form);
  const watchedPaymentMethodId = Form.useWatch("paymentMethodId", form);
  const watchedPayType = Form.useWatch("payType", form);

  const loadRecords = async (nextQuery?: CrudListQuery) => {
    const mergedQuery = { ...query, ...nextQuery };
    setLoading(true);
    try {
      const result = await grainPurchaseEntryApi.list(mergedQuery);
      setRecords(result.data);
      setTotal(result.total);
      setQuery({
        pageIndex: Number(mergedQuery.pageIndex ?? 1),
        pageSize: Number(mergedQuery.pageSize ?? pageSize),
      });
    } catch (error) {
      message.error(error instanceof Error ? error.message : "加载收粮明细失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadRecords();
  }, []);

  useEffect(() => {
    grainStationApi
      .list({ pageIndex: 1, pageSize: 200, status: "active" })
      .then((stations) => {
        setStationOptions(stations.data.map((station) => ({ label: station.stationName, value: station.id })));
      })
      .catch((error) => message.error(error instanceof Error ? error.message : "加载粮站选项失败"));
  }, []);

  useEffect(() => {
    const quantity = typeof watchedQuantity === "number" ? watchedQuantity : 0;
    const amount = typeof watchedAmount === "number" ? watchedAmount : 0;
    if (quantity > 0 && amount > 0) {
      form.setFieldValue("unitPrice", Number((amount / quantity).toFixed(4)));
    }
  }, [form, watchedAmount, watchedQuantity]);

  const filterQuery = () => ({
    pageIndex: 1,
    search: searchValue.trim() || undefined,
    status: statusValue,
  });

  const optionLabel = (options: CrudOption[], value: unknown) =>
    options.find((option) => option.value === value)?.label ?? String(value || "-");
  const storedStationId =
    typeof currentAppUser?.stationId === "number" && currentAppUser.stationId > 0 ? currentAppUser.stationId : undefined;
  const currentStationOption = useMemo(() => {
    if (storedStationId) {
      const option = stationOptions.find((item) => item.value === storedStationId);
      return {
        label: currentAppUser?.stationName || option?.label || `粮站 ${storedStationId}`,
        value: storedStationId,
      };
    }
    return stationOptions[0];
  }, [currentAppUser?.stationName, stationOptions, storedStationId]);
  const currentStationId = typeof currentStationOption?.value === "number" ? currentStationOption.value : undefined;

  useEffect(() => {
    if (!currentStationId) {
      setPurchaseTypeOptions([]);
      setPaymentMethodOptions([]);
      return;
    }
    Promise.all([
      listActiveGrainPurchaseTypes(currentStationId),
      listActiveGrainPaymentMethods(currentStationId),
    ])
      .then(([purchaseTypes, paymentMethods]) => {
        setPurchaseTypeOptions(purchaseTypes.map((type) => ({ label: type.typeName, value: type.id })));
        setPaymentMethodOptions(paymentMethods.map((method) => ({ label: method.methodName, value: method.id, methodCode: method.methodCode })));
      })
      .catch((error) => message.error(error instanceof Error ? error.message : "加载收粮选项失败"));
  }, [currentStationId]);

  const selectedPaymentMethod = useMemo(
    () => paymentMethodOptions.find((option) => option.value === watchedPaymentMethodId),
    [paymentMethodOptions, watchedPaymentMethodId],
  );
  const isBankPayment = isBankPaymentMethod(selectedPaymentMethod, watchedPayType);
  const receiptMeta = receiptFieldMeta(isBankPayment, watchedPayType);

  const stationLabel = (record: GrainPurchaseEntryRecord) =>
    renderValue(record.stationName || optionLabel(stationOptions, record.stationId));

  const selectedSnapshot = useMemo(
    () => historySnapshots.find((item) => item.id === selectedSnapshotId) ?? historySnapshots[0] ?? null,
    [historySnapshots, selectedSnapshotId],
  );

  const previousSnapshot = useMemo(() => {
    if (!selectedSnapshot) {
      return null;
    }
    const index = historySnapshots.findIndex((item) => item.id === selectedSnapshot.id);
    return index >= 0 ? historySnapshots[index + 1] ?? null : null;
  }, [historySnapshots, selectedSnapshot]);

  const historyChangeRows = useMemo(() => {
    if (!selectedSnapshot) {
      return [];
    }
    const fields: Array<{
      key: keyof GrainPurchaseEntrySnapshotRecord;
      label: string;
      render?: (value: unknown) => ReactNode;
    }> = [
      { key: "farmerName", label: "农户姓名" },
      { key: "farmerIdNumber", label: "身份证号", render: (value) => <SensitiveValue value={value} keepStart={6} keepEnd={4} /> },
      { key: "farmerPhone", label: "手机号" },
      { key: "farmerBankNumber", label: "银行卡号", render: (value) => <SensitiveValue value={value} keepStart={4} keepEnd={4} /> },
      { key: "farmerBankName", label: "开户行" },
      { key: "crop", label: "收购类型" },
      { key: "quantity", label: "重量" },
      { key: "unit", label: "单位" },
      { key: "amount", label: "金额", render: (value) => `￥${normalizeMoney(Number(value))}` },
      { key: "unitPrice", label: "单价", render: (value) => `￥${normalizeMoney(Number(value))}` },
      { key: "locationAddress", label: "定位地址" },
      { key: "payType", label: "付款方式" },
      { key: "entryStatus", label: "状态", render: snapshotStatusTag },
      { key: "entryRemark", label: "备注" },
    ];
    return fields
      .filter((field) => {
        if (!previousSnapshot) {
          return selectedSnapshot.snapshotAction === "create" && renderValue(selectedSnapshot[field.key]) !== "-";
        }
        return renderValue(previousSnapshot[field.key]) !== renderValue(selectedSnapshot[field.key]);
      })
      .map<ChangeRow>((field) => ({
        key: String(field.key),
        label: field.label,
        before: previousSnapshot ? field.render?.(previousSnapshot[field.key]) ?? renderValue(previousSnapshot[field.key]) : "-",
        after: field.render?.(selectedSnapshot[field.key]) ?? renderValue(selectedSnapshot[field.key]),
      }));
  }, [previousSnapshot, selectedSnapshot]);

  const openHistory = async (record: GrainPurchaseEntryRecord) => {
    setHistoryRecord(record);
    setHistorySnapshots([]);
    setSelectedSnapshotId(null);
    setHistoryOpen(true);
    setHistoryLoading(true);
    try {
      const result = await grainPurchaseEntrySnapshotApi.list({ entryId: record.id, pageIndex: 1, pageSize: 50 });
      setHistorySnapshots(result.data);
      setSelectedSnapshotId(result.data[0]?.id ?? null);
    } catch (error) {
      message.error(error instanceof Error ? error.message : "加载修改历史失败");
    } finally {
      setHistoryLoading(false);
    }
  };

  const openCreate = () => {
    setEditingRecord(null);
    setFarmerImages(null);
    setEntryMaterials([]);
    form.resetFields();
    form.setFieldsValue({ stationId: currentStationId, unit: "公斤", status: "submitted" });
    setDrawerOpen(true);
  };

  const loadFarmerImages = async (farmerId?: number) => {
    if (!farmerId) {
      setFarmerImages(null);
      return;
    }
    try {
      const images = await getGrainFarmerImages(farmerId);
      setFarmerImages(images);
    } catch (error) {
      setFarmerImages(null);
      message.warning(error instanceof Error ? error.message : "粮户历史图片加载失败");
    }
  };

  const loadEntryMaterials = async (entryId?: number) => {
    if (!entryId) {
      setEntryMaterials([]);
      return;
    }
    try {
      const result = await grainEntryMaterialApi.list({ entryId, pageIndex: 1, pageSize: 50 });
      setEntryMaterials(result.data);
    } catch (error) {
      setEntryMaterials([]);
      message.warning(error instanceof Error ? error.message : "收粮材料图片加载失败");
    }
  };

  const openEdit = (record: GrainPurchaseEntryRecord) => {
    setEditingRecord(record);
    form.resetFields();
    form.setFieldsValue({
      stationId: currentStationId ?? record.stationId,
      farmerId: record.farmerId,
      farmerName: record.farmerName,
      farmerPhone: record.farmerPhone,
      farmerIdNumber: record.farmerIdNumber,
      farmerAddress: record.farmerAddress,
      farmerBankNumber: record.farmerBankNumber,
      farmerBankName: record.farmerBankName,
      purchaseTypeId: record.purchaseTypeId,
      crop: record.crop,
      quantity: record.quantity,
      unit: record.unit || "公斤",
      amount: record.amount,
      unitPrice: record.unitPrice,
      locationAddress: record.locationAddress,
      paymentMethodId: record.paymentMethodId,
      payType: record.payType,
      status: record.status || "submitted",
      remark: record.remark,
    });
    setDrawerOpen(true);
    void loadFarmerImages(record.farmerId);
    void loadEntryMaterials(record.id);
  };

  const applyOcrResult = (result: GrainCardOcrResult) => {
    if (result.cardType === "id-card") {
      form.setFieldsValue({
        farmerName: result.name || form.getFieldValue("farmerName"),
        farmerIdNumber: result.idNumber || form.getFieldValue("farmerIdNumber"),
        farmerAddress: result.address || form.getFieldValue("farmerAddress"),
      });
      message.success(result.mock ? "模拟身份证识别完成" : "身份证识别完成");
      return;
    }
    form.setFieldsValue({
      farmerBankNumber: result.bankNumber || form.getFieldValue("farmerBankNumber"),
      farmerBankName: result.bankName || form.getFieldValue("farmerBankName"),
    });
    message.success(result.mock ? "模拟银行卡识别完成" : "银行卡识别完成");
  };

  const beforeUpload = (file: RcFile) => {
    const isImage = file.type.startsWith("image/");
    if (!isImage) {
      message.error("请上传图片文件");
      return Upload.LIST_IGNORE;
    }
    const underLimit = file.size / 1024 / 1024 < 8;
    if (!underLimit) {
      message.error("图片需小于 8MB");
      return Upload.LIST_IGNORE;
    }
    return true;
  };

  const handleRecognize = async (file: File) => {
    if (!ocrTarget) {
      return;
    }
    const stationId = form.getFieldValue("stationId");
    if (!stationId) {
      message.warning("当前业务员未绑定粮站，无法识别");
      return;
    }
    setOcrLoading(true);
    try {
      const result = await recognizeGrainCard(file, {
        cardType: ocrTarget.cardType,
        imageSide: ocrTarget.imageSide,
        stationId,
        farmerId: form.getFieldValue("farmerId"),
      });
      applyOcrResult(result);
      await loadFarmerImages(form.getFieldValue("farmerId"));
      setOcrTarget(null);
    } catch (error) {
      message.error(error instanceof Error ? error.message : "识别失败，请手动录入");
    } finally {
      setOcrLoading(false);
    }
  };

  const ensureFarmer = async (values: EntryFormValues) => {
    const appUserId =
      typeof currentAppUser?.id === "number" && currentAppUser.id > 0 ? currentAppUser.id : undefined;
    const farmerPayload: Partial<GrainFarmerRecord> = compactPayload({
      stationId: values.stationId,
      appUserId,
      name: values.farmerName,
      phone: values.farmerPhone,
      idNumber: values.farmerIdNumber,
      address: values.farmerAddress,
      bankNumber: values.farmerBankNumber,
      bankName: values.farmerBankName,
    }) as Partial<GrainFarmerRecord>;
    const hasFarmerFields = Object.keys(farmerPayload).some((key) => !["stationId", "appUserId"].includes(key));
    if (values.farmerId) {
      if (hasFarmerFields) {
        await grainFarmerApi.update(values.farmerId, farmerPayload as GrainPayload);
      }
      return values.farmerId;
    }
    if (!hasFarmerFields) {
      return undefined;
    }
    const created = await grainFarmerApi.create(farmerPayload as GrainPayload);
    return (created as GrainFarmerRecord).id;
  };

  const handleSubmit = async () => {
    const validatedValues = await form.validateFields();
    const values = {
      ...validatedValues,
      stationId: currentStationId ?? validatedValues.stationId,
    };
    setSubmitting(true);
    try {
      const farmerId = await ensureFarmer(values);
      const payload = compactPayload({
        stationId: values.stationId,
        farmerId,
        purchaseTypeId: values.purchaseTypeId,
        crop: values.crop,
        quantity: values.quantity,
        unit: values.unit,
        amount: values.amount,
        unitPrice: values.unitPrice,
        locationAddress: values.locationAddress,
        paymentMethodId: values.paymentMethodId,
        payType: values.payType,
        status: values.status,
        remark: values.remark,
      }) as GrainPayload;
      if (editingRecord) {
        await grainPurchaseEntryApi.update(editingRecord.id, payload);
        message.success("收粮明细已更新");
      } else {
        await grainPurchaseEntryApi.create(payload);
        message.success("收粮明细已创建");
      }
      setDrawerOpen(false);
      setEditingRecord(null);
      form.resetFields();
      await loadRecords({ pageIndex: editingRecord ? query.pageIndex : 1 });
    } catch (error) {
      message.error(error instanceof Error ? error.message : "保存失败");
    } finally {
      setSubmitting(false);
    }
  };

  const columns: ColumnsType<GrainPurchaseEntryRecord> = [
    { title: "ID", dataIndex: "id", width: 80, fixed: "left" },
    { title: "粮站", dataIndex: "stationName", width: 170, render: (_, record) => stationLabel(record) },
    { title: "农户姓名", dataIndex: "farmerName", width: 130, render: renderValue },
    {
      title: "身份证号",
      dataIndex: "farmerIdNumber",
      width: 190,
      render: (value) => <SensitiveValue value={value} keepStart={6} keepEnd={4} />,
    },
    {
      title: "银行卡号",
      dataIndex: "farmerBankNumber",
      width: 190,
      render: (value) => <SensitiveValue value={value} keepStart={4} keepEnd={4} />,
    },
    { title: "收购类型", dataIndex: "crop", width: 130 },
    { title: "重量", dataIndex: "quantity", width: 110 },
    { title: "单位", dataIndex: "unit", width: 80 },
    { title: "金额", dataIndex: "amount", width: 120, render: (value) => normalizeMoney(Number(value)) },
    { title: "单价", dataIndex: "unitPrice", width: 120 },
    { title: "付款方式", dataIndex: "payType", width: 130 },
    { title: "状态", dataIndex: "status", width: 110, render: statusTag },
    {
      title: "操作",
      key: "actions",
      width: 210,
      fixed: "right",
      render: (_, record) => (
        <Space size={4}>
          <Tooltip title="修改历史">
            <Button type="text" icon={<HistoryOutlined />} onClick={() => void openHistory(record)} />
          </Tooltip>
          {record.status !== "voided" ? (
            <Popconfirm
              title={`确认作废收粮明细 #${record.id} 吗？`}
              okText="作废"
              cancelText="取消"
              onConfirm={async () => {
                try {
                  setSubmitting(true);
                  await voidGrainPurchaseEntry(record.id);
                  message.success("作废成功");
                  await loadRecords({ pageIndex: query.pageIndex });
                } catch (error) {
                  message.error(error instanceof Error ? error.message : "作废失败");
                } finally {
                  setSubmitting(false);
                }
              }}
            >
              <Tooltip title="作废">
                <Button type="text" danger icon={<StopOutlined />} disabled={submitting} />
              </Tooltip>
            </Popconfirm>
          ) : null}
          <Tooltip title="编辑">
            <Button type="text" icon={<EditOutlined />} onClick={() => openEdit(record)} />
          </Tooltip>
        </Space>
      ),
    },
  ];

  const imageCards = [
    { label: "身份证正面", value: farmerImages?.idCardFront },
    { label: "银行卡", value: farmerImages?.bankCard },
  ];

  return (
    <div className="manager-page-stack">
      <section className="manager-data-card">
        <div style={{ display: "flex", gap: 12, flexWrap: "wrap", justifyContent: "space-between" }}>
          <Space wrap size={12}>
            <Input
              className="manager-filter-input"
              prefix={<SearchOutlined style={{ color: "var(--manager-text-faint)" }} />}
              placeholder="收购类型/付款方式/地址"
              value={searchValue}
              onChange={(event) => setSearchValue(event.target.value)}
              onPressEnter={() => void loadRecords(filterQuery())}
              style={{ width: 280, maxWidth: "100%" }}
            />
            <Select
              allowClear
              placeholder="状态"
              value={statusValue}
              onChange={setStatusValue}
              options={statusOptions}
              style={{ width: 160 }}
            />
            <Button type="primary" icon={<SearchOutlined />} onClick={() => void loadRecords(filterQuery())}>
              查询
            </Button>
            <Button icon={<ReloadOutlined />} onClick={() => void loadRecords()}>
              刷新
            </Button>
          </Space>
          <Space wrap>
            <Tag style={{ color: "var(--manager-text-soft)", background: "var(--manager-green-soft)", border: "none" }}>
              共 {total} 条
            </Tag>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={openCreate}
              style={{ color: "#ffffff", border: "none", background: "linear-gradient(135deg, #145535 0%, #237a4b 100%)" }}
            >
              新增收粮明细
            </Button>
          </Space>
        </div>
      </section>

      <section className="manager-data-card manager-table">
        <Table<GrainPurchaseEntryRecord>
          rowKey="id"
          loading={loading}
          dataSource={records}
          columns={columns}
          scroll={{ x: 2010 }}
          pagination={{
            current: query.pageIndex,
            pageSize: query.pageSize,
            total,
            showSizeChanger: false,
            onChange: (page) => void loadRecords({ ...filterQuery(), pageIndex: page }),
          }}
        />
      </section>

      <Drawer
        title={editingRecord ? `编辑收粮明细 #${editingRecord.id}` : "新增收粮明细"}
        open={drawerOpen}
        width={1120}
        className="manager-crud-drawer"
        destroyOnClose
        onClose={() => {
          setDrawerOpen(false);
          setEditingRecord(null);
          setFarmerImages(null);
          setEntryMaterials([]);
          form.resetFields();
        }}
        extra={
          <Space>
            <Button onClick={() => setDrawerOpen(false)}>取消</Button>
            <Button type="primary" loading={submitting} onClick={() => void handleSubmit()}>
              保存
            </Button>
          </Space>
        }
      >
        <Form
          form={form}
          layout="vertical"
          preserve={false}
          className="manager-crud-form manager-entry-form"
        >
          <aside className="manager-crud-form-rail">
            {entryFormSections.map((section, index) => (
              <a key={section.key} href={`#entry-form-${section.key}`}>
                <span>{String(index + 1).padStart(2, "0")}</span>
                {section.title}
              </a>
            ))}
          </aside>

          <div className="manager-crud-form-main">
            <Form.Item name="farmerId" hidden>
              <InputNumber />
            </Form.Item>
            <section id="entry-form-ownership" className="manager-form-section">
              <div className="manager-form-section-head">
                <div>
                  <Typography.Title level={5} style={{ margin: 0 }}>
                    基础信息
                  </Typography.Title>
                  <Text type="secondary">确认粮站归属，保存后用于收粮记录归档。</Text>
                </div>
                <EnvironmentOutlined style={{ color: "#237a4b" }} />
              </div>
              <div className="manager-form-grid">
                <Form.Item name="stationId" label="粮站" rules={[{ required: true, message: "请选择粮站" }]}>
                  <Select
                    options={currentStationOption ? [currentStationOption] : []}
                    placeholder={currentStationOption ? undefined : "当前业务员未绑定粮站"}
                    disabled
                  />
                </Form.Item>
              </div>
            </section>

            <section id="entry-form-grain-farmer" className="manager-form-section">
              <div className="manager-form-section-head">
                <div>
                  <Typography.Title level={5} style={{ margin: 0 }}>
                    粮食与农户身份信息
                  </Typography.Title>
                  <Text type="secondary">把本笔粮食品类、重量金额和农户身份资料放在一起核对。</Text>
                </div>
                <Space wrap>
                  <Button icon={<ScanOutlined />} onClick={() => setOcrTarget({ title: "识别身份证正面", cardType: "id-card", imageSide: "front" })}>
                    身份证正面
                  </Button>
                </Space>
              </div>
              <Alert
                type="info"
                showIcon
                style={{ marginBottom: 16 }}
                message="身份证可拍照识别后自动带入，也可以直接手动输入；编辑已有记录时识别结果会同步回粮户资料。"
              />
              <div className="manager-form-grid">
                <Form.Item name="purchaseTypeId" label="收购类型">
                  <Select
                    options={purchaseTypeOptions}
                    placeholder="请选择收购类型"
                    allowClear
                    showSearch
                    optionFilterProp="label"
                    onChange={(_, option) => {
                      const selected = Array.isArray(option) ? option[0] : option;
                      if (selected?.label) {
                        form.setFieldValue("crop", String(selected.label));
                      }
                    }}
                  />
                </Form.Item>
                <Form.Item name="crop" label="收购类型名称" rules={[{ required: true, message: "请输入收购类型名称" }]}>
                  <Input placeholder="如：小麦、玉米" />
                </Form.Item>
                <Form.Item name="quantity" label="重量" rules={[{ required: true, message: "请输入重量" }]}>
                  <InputNumber min={0} precision={3} style={{ width: "100%" }} addonAfter={watchedUnit || "公斤"} />
                </Form.Item>
                <Form.Item name="unit" label="单位">
                  <Input placeholder="公斤" />
                </Form.Item>
                <Form.Item name="amount" label="金额" rules={[{ required: true, message: "请输入金额" }]}>
                  <InputNumber min={0} precision={2} style={{ width: "100%" }} addonBefore="￥" />
                </Form.Item>
                <Form.Item name="unitPrice" label="单价">
                  <InputNumber min={0} precision={4} style={{ width: "100%" }} addonBefore="￥" />
                </Form.Item>
                <Divider className="manager-form-item-wide" style={{ margin: "4px 0 16px" }} />
                <Form.Item name="farmerName" label="粮户姓名">
                  <Input placeholder="拍身份证自动带入，或手动输入" prefix={<IdcardOutlined />} />
                </Form.Item>
                <Form.Item name="farmerPhone" label="手机号">
                  <Input placeholder="请输入手机号" />
                </Form.Item>
                <Form.Item name="farmerIdNumber" label="身份证号">
                  <Input placeholder="请输入身份证号" />
                </Form.Item>
                <Form.Item name="farmerAddress" label="身份证住址" className="manager-form-item-wide">
                  <Input.TextArea rows={2} placeholder="请输入身份证住址" />
                </Form.Item>
              </div>
            </section>

            <section id="entry-form-payment" className="manager-form-section">
              <div className="manager-form-section-head">
                <div>
                  <Typography.Title level={5} style={{ margin: 0 }}>
                    收付款信息
                  </Typography.Title>
                  <Text type="secondary">先选择付款方式，再按对应方式填写收款账号信息。</Text>
                </div>
                {isBankPayment ? (
                  <Button icon={<BankOutlined />} onClick={() => setOcrTarget({ title: "识别银行卡", cardType: "bank-card" })}>
                    银行卡识别
                  </Button>
                ) : (
                  <BankOutlined style={{ color: "#237a4b" }} />
                )}
              </div>
              <div className="manager-form-grid">
                <Form.Item name="paymentMethodId" label="付款方式">
                  <Select
                    options={paymentMethodOptions}
                    placeholder="请选择付款方式"
                    allowClear
                    showSearch
                    optionFilterProp="label"
                    onChange={(_, option) => {
                      const selected = Array.isArray(option) ? option[0] : option;
                      if (selected?.label) {
                        form.setFieldValue("payType", String(selected.label));
                      } else {
                        form.setFieldValue("payType", undefined);
                      }
                    }}
                  />
                </Form.Item>
                <Form.Item name="payType" label="付款方式名称">
                  <Input placeholder="请输入付款方式" />
                </Form.Item>
                <Form.Item name="farmerBankNumber" label={receiptMeta.accountLabel}>
                  <Input
                    placeholder={receiptMeta.accountPlaceholder}
                    prefix={isBankPayment ? <BankOutlined /> : undefined}
                  />
                </Form.Item>
                <Form.Item name="farmerBankName" label={receiptMeta.nameLabel}>
                  <Input placeholder={receiptMeta.namePlaceholder} />
                </Form.Item>
                <Form.Item name="locationAddress" label="定位地址" className="manager-form-item-wide">
                  <Input placeholder="请输入定位地址" />
                </Form.Item>
                <Form.Item name="status" label="状态">
                  <Select options={statusOptions} placeholder="请选择状态" />
                </Form.Item>
                <Form.Item name="remark" label="备注" className="manager-form-item-wide">
                  <Input.TextArea rows={3} placeholder="请输入备注" />
                </Form.Item>
              </div>
            </section>
          </div>

          <aside id="entry-form-materials" style={{ position: "sticky", top: 18, alignSelf: "start" }}>
            <section className="manager-form-section" style={{ marginBottom: 16 }}>
              <Text strong>单据预览</Text>
              <Descriptions column={1} size="small" style={{ marginTop: 12 }}>
                <Descriptions.Item label="金额">
                  <Text strong>￥{normalizeMoney(watchedAmount)}</Text>
                </Descriptions.Item>
                <Descriptions.Item label="重量">{renderValue(watchedQuantity)}</Descriptions.Item>
                <Descriptions.Item label="单价">￥{renderValue(watchedUnitPrice)}</Descriptions.Item>
              </Descriptions>
            </section>

            <section className="manager-form-section">
              <Space align="center" style={{ marginBottom: 12 }}>
                <FileImageOutlined style={{ color: "#237a4b" }} />
                <Text strong>历史图片</Text>
              </Space>
              {imageCards.some((item) => item.value) ? (
                <div style={{ display: "grid", gap: 12 }}>
                  {imageCards.map((item) =>
                    item.value ? (
                      <div key={item.label}>
                        <Text type="secondary">{item.label}</Text>
                        <Image
                          src={item.value}
                          alt={item.label}
                          style={{ marginTop: 6, width: "100%", height: 120, objectFit: "cover", borderRadius: 8 }}
                        />
                      </div>
                    ) : null,
                  )}
                </div>
              ) : (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无历史图片" />
              )}
            </section>

            <section className="manager-form-section" style={{ marginTop: 16 }}>
              <Space align="center" style={{ marginBottom: 12 }}>
                <CameraOutlined style={{ color: "#237a4b" }} />
                <Text strong>收粮材料</Text>
              </Space>
              {entryMaterials.length > 0 ? (
                <div style={{ display: "grid", gap: 12 }}>
                  {entryMaterials.map((item) => (
                    <div key={item.id}>
                      <Image
                        src={item.imageUrl}
                        alt={item.fileName || "收粮材料"}
                        style={{ width: "100%", height: 120, objectFit: "cover", borderRadius: 8 }}
                      />
                    </div>
                  ))}
                </div>
              ) : (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无收粮材料图片" />
              )}
            </section>
          </aside>
        </Form>
      </Drawer>

      <Drawer
        title={historyRecord ? `修改历史 #${historyRecord.id}` : "修改历史"}
        open={historyOpen}
        width={1040}
        destroyOnClose
        onClose={() => {
          setHistoryOpen(false);
          setHistoryRecord(null);
          setHistorySnapshots([]);
          setSelectedSnapshotId(null);
        }}
      >
        <div style={{ display: "grid", gridTemplateColumns: "300px minmax(0, 1fr)", gap: 20, alignItems: "start" }}>
          <section className="manager-data-card" style={{ boxShadow: "none", padding: 16 }}>
            <Space direction="vertical" size={2} style={{ width: "100%", marginBottom: 16 }}>
              <Text strong>记录轨迹</Text>
            </Space>
            {historySnapshots.length > 0 ? (
              <Timeline
                items={historySnapshots.map((snapshot) => {
                  const meta = actionMeta(snapshot.snapshotAction);
                  const active = selectedSnapshot?.id === snapshot.id;
                  return {
                    color: meta.color,
                    children: (
                      <button
                        type="button"
                        onClick={() => setSelectedSnapshotId(snapshot.id)}
                        style={{
                          width: "100%",
                          padding: "8px 10px",
                          textAlign: "left",
                          cursor: "pointer",
                          border: active ? "1px solid var(--manager-primary)" : "1px solid transparent",
                          borderRadius: 8,
                          background: active ? "var(--manager-green-soft)" : "transparent",
                        }}
                      >
                        <Space direction="vertical" size={2}>
                          <Space size={8}>
                            <Tag color={meta.color}>{meta.label}</Tag>
                            <Text strong>版本 {snapshot.snapshotVersion}</Text>
                          </Space>
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            {formatTime(snapshot.snapshotTime)}
                          </Text>
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            {snapshot.operatorName || "manager"}
                          </Text>
                        </Space>
                      </button>
                    ),
                  };
                })}
              />
            ) : (
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={historyLoading ? "正在加载修改历史" : "暂无修改历史"} />
            )}
          </section>

          <section className="manager-data-card" style={{ boxShadow: "none", padding: 18 }}>
            {selectedSnapshot ? (
              <>
                <div style={{ display: "flex", justifyContent: "space-between", gap: 12, flexWrap: "wrap", marginBottom: 14 }}>
                  <Space align="center" wrap>
                    <Tag color={actionMeta(selectedSnapshot.snapshotAction).color}>
                      {actionMeta(selectedSnapshot.snapshotAction).label}
                    </Tag>
                    <Text strong>版本 {selectedSnapshot.snapshotVersion}</Text>
                    <Text type="secondary">{formatTime(selectedSnapshot.snapshotTime)}</Text>
                  </Space>
                  <Text type="secondary">操作人：{selectedSnapshot.operatorName || "manager"}</Text>
                </div>

                <Table<ChangeRow>
                  rowKey="key"
                  size="small"
                  pagination={false}
                  dataSource={historyChangeRows}
                  locale={{ emptyText: selectedSnapshot.snapshotAction === "create" ? "创建记录无对比项" : "本次未发现字段差异" }}
                  columns={[
                    { title: "字段", dataIndex: "label", width: 120 },
                    { title: "变更前", dataIndex: "before", width: 220 },
                    { title: "变更后", dataIndex: "after", width: 220 },
                  ]}
                />

                <Divider />

                <Descriptions bordered size="small" column={2}>
                  <Descriptions.Item label="农户姓名">{renderValue(selectedSnapshot.farmerName)}</Descriptions.Item>
                  <Descriptions.Item label="身份证号">
                    <SensitiveValue value={selectedSnapshot.farmerIdNumber} keepStart={6} keepEnd={4} />
                  </Descriptions.Item>
                  <Descriptions.Item label="银行卡号">
                    <SensitiveValue value={selectedSnapshot.farmerBankNumber} keepStart={4} keepEnd={4} />
                  </Descriptions.Item>
                  <Descriptions.Item label="开户行">{renderValue(selectedSnapshot.farmerBankName)}</Descriptions.Item>
                  <Descriptions.Item label="手机号">{renderValue(selectedSnapshot.farmerPhone)}</Descriptions.Item>
                  <Descriptions.Item label="收购类型">{renderValue(selectedSnapshot.crop)}</Descriptions.Item>
                  <Descriptions.Item label="重量">
                    {renderValue(selectedSnapshot.quantity)} {selectedSnapshot.unit || "公斤"}
                  </Descriptions.Item>
                  <Descriptions.Item label="金额">￥{normalizeMoney(Number(selectedSnapshot.amount))}</Descriptions.Item>
                  <Descriptions.Item label="单价">￥{normalizeMoney(Number(selectedSnapshot.unitPrice))}</Descriptions.Item>
                  <Descriptions.Item label="付款方式">{renderValue(selectedSnapshot.payType)}</Descriptions.Item>
                  <Descriptions.Item label="状态">{snapshotStatusTag(selectedSnapshot.entryStatus)}</Descriptions.Item>
                  <Descriptions.Item label="备注">{renderValue(selectedSnapshot.entryRemark)}</Descriptions.Item>
                  <Descriptions.Item label="定位地址" span={2}>
                    {renderValue(selectedSnapshot.locationAddress)}
                  </Descriptions.Item>
                </Descriptions>
              </>
            ) : (
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={historyLoading ? "正在加载修改历史" : "选择左侧记录查看详情"} />
            )}
          </section>
        </div>
      </Drawer>

      <Modal
        title={ocrTarget?.title}
        open={Boolean(ocrTarget)}
        onCancel={() => setOcrTarget(null)}
        footer={null}
        destroyOnClose
      >
        <Upload.Dragger
          accept="image/*"
          maxCount={1}
          showUploadList={false}
          beforeUpload={beforeUpload}
          customRequest={({ file, onSuccess }) => {
            void handleRecognize(file as File).then(() => onSuccess?.("ok"));
          }}
          disabled={ocrLoading}
        >
          <p className="ant-upload-drag-icon">
            <ScanOutlined />
          </p>
          <p className="ant-upload-text">{ocrLoading ? "正在识别，请稍候" : "点击或拖拽图片到这里识别"}</p>
          <p className="ant-upload-hint">支持身份证正面和银行卡照片；识别失败时可直接回到表单手动输入。</p>
        </Upload.Dragger>
      </Modal>
    </div>
  );
}
