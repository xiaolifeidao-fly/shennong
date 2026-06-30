"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import {
  BankOutlined,
  CameraOutlined,
  DeleteOutlined,
  DownloadOutlined,
  EditOutlined,
  ExportOutlined,
  FileImageOutlined,
  HistoryOutlined,
  IdcardOutlined,
  PlusOutlined,
  ReloadOutlined,
  ScanOutlined,
  SearchOutlined,
  StopOutlined,
  WalletOutlined,
} from "@ant-design/icons";
import {
  Alert,
  Button,
  DatePicker,
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
  Progress,
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
import dayjs from "dayjs";
import type { Dayjs } from "dayjs";
import type { RcFile } from "antd/es/upload";
import type { ColumnsType } from "antd/es/table";
import { fetchAppUsers } from "../../app-user/api/app-user.api";
import type { CrudListQuery, CrudOption } from "../../components/CrudManagementPanel";
import {
  getGrainFarmerImages,
  grainEntryMaterialApi,
  grainFarmerApi,
  grainPaymentMethodApi,
  grainPurchaseEntryApi,
  grainPurchaseEntryExportApi,
  grainPurchaseEntrySnapshotApi,
  listAllGrainPurchaseTypes,
  recognizeGrainCard,
  voidGrainPurchaseEntry,
  type GrainCardOcrResult,
  type GrainFarmerImagesRecord,
  type GrainFarmerRecord,
  type GrainPayload,
  type GrainEntryMaterialRecord,
  type GrainPurchaseEntryExportBatchRecord,
  type GrainPurchaseEntryRecord,
  type GrainPurchaseEntrySnapshotRecord,
} from "../api/grain.api";
import { useAccessibleStations } from "../hooks/useAccessibleStations";
import { SensitiveValue } from "./SensitiveValue";

const { Text } = Typography;

type EntryFormValues = {
  stationId?: number;
  appUserId?: number;
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
  payTime?: Dayjs;
  status?: string;
  remark?: string;
};

type OcrTarget = {
  title: string;
  cardType: "id-card" | "bank-card" | "social-security-card";
  imageSide?: "front" | "back";
};

type ChangeRow = {
  key: string;
  label: string;
  before: ReactNode;
  after: ReactNode;
};

type PaymentMethodOption = CrudOption & {
  methodCode: string;
};

const statusOptions: CrudOption[] = [
  { label: "已提交", value: "submitted" },
  { label: "已作废", value: "voided" },
];

const pageSize = 10;

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

function exportStatusMeta(value: string) {
  switch (value) {
    case "pending":
      return { label: "等待中", color: "default" };
    case "running":
      return { label: "导出中", color: "processing" };
    case "success":
      return { label: "已完成", color: "success" };
    case "partial_success":
      return { label: "部分成功", color: "warning" };
    case "failed":
      return { label: "失败", color: "error" };
    default:
      return { label: renderValue(value), color: "default" };
  }
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

function actionMeta(action: string) {
  switch (action) {
    case "create":
      return { label: "创建", color: "green" };
    case "update":
      return { label: "修改", color: "blue" };
    case "material_create":
      return { label: "新增材料", color: "purple" };
    case "material_delete":
      return { label: "删除材料", color: "orange" };
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

function renderMaterialSummary(value: unknown) {
  const raw = String(value || "").trim();
  if (!raw) {
    return "-";
  }
  try {
    const parsed = JSON.parse(raw) as { count?: number; items?: Array<{ fileName?: string; materialType?: string }> };
    const count = Number(parsed.count ?? parsed.items?.length ?? 0);
    const names = (parsed.items || [])
      .map((item, index) => item.fileName || item.materialType || `材料${index + 1}`)
      .filter(Boolean)
      .slice(0, 3);
    return names.length > 0 ? `${count} 张：${names.join("、")}${count > names.length ? "等" : ""}` : `${count} 张`;
  } catch {
    return raw;
  }
}

function isBankPaymentMethod(method?: { label?: unknown; methodCode?: unknown }, payType?: string) {
  const code = String(method?.methodCode || "").toUpperCase();
  const label = String(method?.label || payType || "");
  return code.includes("BANK") || label.includes("银行") || label.includes("银行卡");
}

export function GrainPurchaseEntryPanel() {
  const { stations: accessibleStations } = useAccessibleStations();
  const stationOptions: CrudOption[] = accessibleStations.map((s) => ({ label: s.stationName, value: s.id }));
  const [form] = Form.useForm<EntryFormValues>();
  const [records, setRecords] = useState<GrainPurchaseEntryRecord[]>([]);
  const [total, setTotal] = useState(0);
  const [query, setQuery] = useState({ pageIndex: 1, pageSize });
  const [filterStationId, setFilterStationId] = useState<number | undefined>(undefined);
  const [filterPurchaseTypeIds, setFilterPurchaseTypeIds] = useState<number[]>([]);
  const [filterDateRange, setFilterDateRange] = useState<[Dayjs, Dayjs] | null>(null);
  const [filterAppUserIds, setFilterAppUserIds] = useState<number[]>([]);
  const [filterFarmerOptions, setFilterFarmerOptions] = useState<CrudOption[]>([]);
  const [filterFarmerIds, setFilterFarmerIds] = useState<number[]>([]);
  const [farmerSearching, setFarmerSearching] = useState(false);
  const [filterAppUserOptions, setFilterAppUserOptions] = useState<CrudOption[]>([]);
  const [appUserSearching, setAppUserSearching] = useState(false);
  const [filterMinAmount, setFilterMinAmount] = useState<number | undefined>(undefined);
  const [filterMaxAmount, setFilterMaxAmount] = useState<number | undefined>(undefined);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<GrainPurchaseEntryRecord | null>(null);
  const [appUserOptions, setAppUserOptions] = useState<CrudOption[]>([]);
  const [purchaseTypeOptions, setPurchaseTypeOptions] = useState<CrudOption[]>([]);
  const [purchaseTypeLoading, setPurchaseTypeLoading] = useState(false);
  const [paymentMethodOptions, setPaymentMethodOptions] = useState<PaymentMethodOption[]>([]);
  const [farmerImages, setFarmerImages] = useState<GrainFarmerImagesRecord | null>(null);
  const [entryMaterials, setEntryMaterials] = useState<GrainEntryMaterialRecord[]>([]);
  const [materialUploading, setMaterialUploading] = useState(false);
  const [ocrTarget, setOcrTarget] = useState<OcrTarget | null>(null);
  const [ocrLoading, setOcrLoading] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyRecord, setHistoryRecord] = useState<GrainPurchaseEntryRecord | null>(null);
  const [historySnapshots, setHistorySnapshots] = useState<GrainPurchaseEntrySnapshotRecord[]>([]);
  const [selectedSnapshotId, setSelectedSnapshotId] = useState<number | null>(null);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [exportOpen, setExportOpen] = useState(false);
  const [exportCount, setExportCount] = useState<number | null>(null);
  const [exportCounting, setExportCounting] = useState(false);
  const [exportStarting, setExportStarting] = useState(false);
  const [exportBatches, setExportBatches] = useState<GrainPurchaseEntryExportBatchRecord[]>([]);
  const [exportTotal, setExportTotal] = useState(0);
  const [exportPageIndex, setExportPageIndex] = useState(1);
  const [exportLoading, setExportLoading] = useState(false);

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
    Promise.all([
      fetchAppUsers({ pageIndex: 1, pageSize: 200, status: "active" }),
      grainPaymentMethodApi.list({ pageIndex: 1, pageSize: 200, status: "active" }),
    ])
      .then(([appUsers, paymentMethods]) => {
        setAppUserOptions(
          appUsers.data.map((user) => ({ label: user.name || user.username || `业务员 ${user.id}`, value: user.id })),
        );
        setPaymentMethodOptions(
          paymentMethods.data.map((method) => ({ label: method.methodName, value: method.id, methodCode: method.methodCode })),
        );
      })
      .catch((error) => message.error(error instanceof Error ? error.message : "加载收粮选项失败"));
  }, []);

  const loadPurchaseTypes = useCallback(async (stationId?: number) => {
    setPurchaseTypeLoading(true);
    try {
      const purchaseTypes = await listAllGrainPurchaseTypes({
        status: "active",
        stationId,
      });
      const nextOptions = purchaseTypes.map((type) => ({ label: type.typeName, value: type.id }));
      const optionIds = new Set(nextOptions.map((option) => Number(option.value)));
      setPurchaseTypeOptions(nextOptions);
      setFilterPurchaseTypeIds((prev) => prev.filter((id) => optionIds.has(id)));
    } catch (error) {
      message.error(error instanceof Error ? error.message : "加载收粮类型失败");
    } finally {
      setPurchaseTypeLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadPurchaseTypes(filterStationId);
  }, [filterStationId, loadPurchaseTypes]);

  useEffect(() => {
    const quantity = typeof watchedQuantity === "number" ? watchedQuantity : 0;
    const amount = typeof watchedAmount === "number" ? watchedAmount : 0;
    if (quantity > 0 && amount > 0) {
      form.setFieldValue("unitPrice", Number((amount / quantity).toFixed(4)));
    }
  }, [form, watchedAmount, watchedQuantity]);

  const stats = useMemo(
    () => [
      { label: "收粮明细总数", value: total },
      { label: "当前页重量", value: records.reduce((sum, item) => sum + Number(item.quantity || 0), 0).toFixed(3) },
      { label: "当前页金额", value: records.reduce((sum, item) => sum + Number(item.amount || 0), 0).toFixed(2) },
    ],
    [records, total],
  );

  const filterQuery = () => ({
    pageIndex: 1,
    stationIds: filterStationId !== undefined ? String(filterStationId) : undefined,
    purchaseTypeIds: filterPurchaseTypeIds.length > 0 ? filterPurchaseTypeIds.join(",") : undefined,
    startDate: filterDateRange?.[0]?.format("YYYY-MM-DD") ?? undefined,
    endDate: filterDateRange?.[1]?.format("YYYY-MM-DD") ?? undefined,
    appUserIds: filterAppUserIds.length > 0 ? filterAppUserIds.join(",") : undefined,
    farmerIds: filterFarmerIds.length > 0 ? filterFarmerIds.join(",") : undefined,
    minAmount: filterMinAmount,
    maxAmount: filterMaxAmount,
  });

  const exportFilterQuery = () => {
    return Object.fromEntries(Object.entries(filterQuery()).filter(([key]) => key !== "pageIndex")) as CrudListQuery;
  };

  const loadExportBatches = useCallback(async (pageIndex = exportPageIndex) => {
    setExportLoading(true);
    try {
      const result = await grainPurchaseEntryExportApi.list({ pageIndex, pageSize: 5 });
      setExportBatches(result.data);
      setExportTotal(result.total);
      setExportPageIndex(pageIndex);
    } catch (error) {
      message.error(error instanceof Error ? error.message : "加载导出批次失败");
    } finally {
      setExportLoading(false);
    }
  }, [exportPageIndex]);

  const openExport = async () => {
    setExportOpen(true);
    setExportCount(null);
    await loadExportBatches(1);
  };

  const handleExportCount = async () => {
    setExportCounting(true);
    try {
      const result = await grainPurchaseEntryExportApi.count(exportFilterQuery());
      setExportCount(result.totalCount);
    } catch (error) {
      message.error(error instanceof Error ? error.message : "统计导出数量失败");
    } finally {
      setExportCounting(false);
    }
  };

  const handleStartExport = async () => {
    setExportStarting(true);
    try {
      const result = await grainPurchaseEntryExportApi.create(exportFilterQuery());
      setExportCount(result.totalCount);
      message.success(`导出批次 ${result.batch.batchNo} 已创建`);
      await loadExportBatches(1);
    } catch (error) {
      message.error(error instanceof Error ? error.message : "创建导出批次失败");
    } finally {
      setExportStarting(false);
    }
  };

  useEffect(() => {
    if (!exportOpen) {
      return undefined;
    }
    const hasRunning = exportBatches.some((item) => item.status === "pending" || item.status === "running");
    if (!hasRunning) {
      return undefined;
    }
    const timer = window.setInterval(() => {
      void loadExportBatches(exportPageIndex);
    }, 5000);
    return () => window.clearInterval(timer);
  }, [exportBatches, exportOpen, exportPageIndex, loadExportBatches]);

  const searchFarmers = useCallback(async (keyword: string) => {
    if (!keyword.trim()) return;
    setFarmerSearching(true);
    try {
      const result = await grainFarmerApi.list({ pageIndex: 1, pageSize: 50, search: keyword.trim() });
      setFilterFarmerOptions(result.data.map((f) => ({ label: f.name || `农户 ${f.id}`, value: f.id })));
    } catch {
      // ignore
    } finally {
      setFarmerSearching(false);
    }
  }, []);

  const searchAppUsers = useCallback(async (keyword: string) => {
    if (!keyword.trim()) return;
    setAppUserSearching(true);
    try {
      const result = await fetchAppUsers({ pageIndex: 1, pageSize: 50, search: keyword.trim(), status: "active" });
      setFilterAppUserOptions(
        result.data.map((u) => {
          const name = u.name || "";
          const username = u.username || "";
          const label = name && username ? `${name}（${username}）` : name || username || `业务员 ${u.id}`;
          return { label, value: u.id };
        }),
      );
    } catch {
      // ignore
    } finally {
      setAppUserSearching(false);
    }
  }, []);

  const optionLabel = (options: CrudOption[], value: unknown) =>
    options.find((option) => option.value === value)?.label ?? String(value || "-");

  const stationLabel = (record: GrainPurchaseEntryRecord) =>
    renderValue(record.stationName || optionLabel(stationOptions, record.stationId));

  const selectedSnapshot = useMemo(
    () => historySnapshots.find((item) => item.id === selectedSnapshotId) ?? historySnapshots[0] ?? null,
    [historySnapshots, selectedSnapshotId],
  );
  const selectedPaymentMethod = useMemo(
    () => paymentMethodOptions.find((option) => option.value === watchedPaymentMethodId),
    [paymentMethodOptions, watchedPaymentMethodId],
  );
  const hasPaymentMethod = Boolean(watchedPaymentMethodId || watchedPayType);
  const bankPaymentSelected = hasPaymentMethod && isBankPaymentMethod(selectedPaymentMethod, watchedPayType);

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
      { key: "farmerBankName", label: "开户人" },
      { key: "crop", label: "收购类型" },
      { key: "quantity", label: "重量" },
      { key: "unit", label: "单位" },
      { key: "amount", label: "金额", render: (value) => `￥${normalizeMoney(Number(value))}` },
      { key: "unitPrice", label: "单价", render: (value) => `￥${normalizeMoney(Number(value))}` },
      { key: "locationAddress", label: "收购地址" },
      { key: "payType", label: "收款方式" },
      { key: "entryStatus", label: "状态", render: snapshotStatusTag },
      { key: "entryRemark", label: "备注" },
      { key: "materialCount", label: "材料数量" },
      { key: "materialSummary", label: "材料摘要", render: renderMaterialSummary },
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
    form.setFieldsValue({ unit: "公斤", status: "submitted" });
    setDrawerOpen(true);
  };

  const loadFarmerImages = async (farmerId?: number, appUserId?: number) => {
    if (!farmerId) {
      setFarmerImages(null);
      return;
    }
    try {
      const images = await getGrainFarmerImages(farmerId, appUserId);
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

  const handleMaterialUpload = async (file: File) => {
    if (!editingRecord) return;
    const allowedTypes = ["image/jpeg", "image/png", "image/webp", "image/heic", "image/heif"];
    if (!allowedTypes.includes(file.type) && !file.type.startsWith("image/")) {
      message.error("仅支持上传图片文件");
      return;
    }
    if (file.size > 10 * 1024 * 1024) {
      message.error("图片大小不能超过 10MB");
      return;
    }
    setMaterialUploading(true);
    try {
      await grainEntryMaterialApi.upload(file, {
        entryId: editingRecord.id,
        farmerId: editingRecord.farmerId,
        stationId: editingRecord.stationId,
      });
      await loadEntryMaterials(editingRecord.id);
      message.success("上传成功");
    } catch (error) {
      message.error(error instanceof Error ? error.message : "上传失败");
    } finally {
      setMaterialUploading(false);
    }
  };

  const handleMaterialDelete = async (id: number) => {
    try {
      await grainEntryMaterialApi.delete(id);
      setEntryMaterials((prev) => prev.filter((m) => m.id !== id));
      message.success("已删除");
    } catch (error) {
      message.error(error instanceof Error ? error.message : "删除失败");
    }
  };

  const openEdit = (record: GrainPurchaseEntryRecord) => {
    setEditingRecord(record);
    form.resetFields();
    form.setFieldsValue({
      stationId: record.stationId,
      appUserId: record.appUserId,
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
      payTime: record.payTime ? dayjs(record.payTime) : undefined,
      status: record.status || "submitted",
      remark: record.remark,
    });
    setDrawerOpen(true);
    void loadFarmerImages(record.farmerId, record.appUserId);
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
    const cardLabel = result.cardType === "social-security-card" ? "社保卡" : "银行卡";
    form.setFieldsValue({
      farmerBankNumber: result.bankNumber || form.getFieldValue("farmerBankNumber"),
      farmerBankName: result.bankName || form.getFieldValue("farmerBankName"),
    });
    message.success(result.mock ? `模拟${cardLabel}识别完成` : `${cardLabel}识别完成`);
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
      message.warning("请先选择粮站，再进行识别");
      return;
    }
    setOcrLoading(true);
    try {
      const result = await recognizeGrainCard(file, {
        cardType: ocrTarget.cardType,
        imageSide: ocrTarget.imageSide,
        stationId,
        appUserId: form.getFieldValue("appUserId"),
        farmerId: form.getFieldValue("farmerId"),
      });
      applyOcrResult(result);
      await loadFarmerImages(form.getFieldValue("farmerId"), form.getFieldValue("appUserId"));
      setOcrTarget(null);
    } catch (error) {
      message.error(error instanceof Error ? error.message : "识别失败，请手动录入");
    } finally {
      setOcrLoading(false);
    }
  };

  const ensureFarmer = async (values: EntryFormValues) => {
    const farmerPayload: Partial<GrainFarmerRecord> = compactPayload({
      stationId: values.stationId,
      appUserId: values.appUserId,
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
    const values = await form.validateFields();
    setSubmitting(true);
    try {
      const farmerId = await ensureFarmer(values);
      const payload = compactPayload({
        stationId: values.stationId,
        appUserId: values.appUserId,
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
        payTime: values.payTime?.toISOString(),
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
    { title: "业务员", dataIndex: "appUserId", width: 150, render: (value) => optionLabel(appUserOptions, value) },
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
    { title: "收购地址", dataIndex: "locationAddress", width: 220, render: renderValue },
    { title: "收款方式", dataIndex: "payType", width: 130 },
    { title: "收款时间", dataIndex: "payTime", width: 170, render: formatTime },
    { title: "创建时间", dataIndex: "createdTime", width: 170, render: formatTime },
    { title: "更新时间", dataIndex: "updatedTime", width: 170, render: formatTime },
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
          <Popconfirm
            title="确认删除这条收粮明细记录吗？"
            okText="删除"
            cancelText="取消"
            onConfirm={async () => {
              try {
                setSubmitting(true);
                await grainPurchaseEntryApi.remove(record.id);
                message.success("删除成功");
                const nextPage = records.length === 1 && query.pageIndex > 1 ? query.pageIndex - 1 : query.pageIndex;
                await loadRecords({ pageIndex: nextPage });
              } catch (error) {
                message.error(error instanceof Error ? error.message : "删除失败");
              } finally {
                setSubmitting(false);
              }
            }}
          >
            <Tooltip title="删除">
              <Button danger type="text" icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const exportColumns: ColumnsType<GrainPurchaseEntryExportBatchRecord> = [
    { title: "批次", dataIndex: "batchNo", width: 180 },
    { title: "用户名", dataIndex: "username", width: 120, render: renderValue },
    {
      title: "状态",
      dataIndex: "status",
      width: 120,
      render: (value: string) => {
        const meta = exportStatusMeta(value);
        return <Tag color={meta.color}>{meta.label}</Tag>;
      },
    },
    { title: "总数量", dataIndex: "totalCount", width: 90 },
    { title: "成功数量", dataIndex: "successCount", width: 90 },
    { title: "失败数量", dataIndex: "failCount", width: 90 },
    {
      title: "进度",
      key: "progress",
      width: 160,
      render: (_, record) => {
        const done = Number(record.successCount || 0) + Number(record.failCount || 0);
        const percent = record.totalCount > 0 ? Math.min(100, Math.round((done / record.totalCount) * 100)) : 0;
        return <Progress percent={percent} size="small" status={record.status === "failed" ? "exception" : undefined} />;
      },
    },
    { title: "时间", dataIndex: "createdTime", width: 170, render: formatTime },
    {
      title: "操作",
      key: "actions",
      width: 90,
      render: (_, record) =>
        record.status === "success" || record.status === "partial_success" ? (
          <Tooltip title="下载文件">
            <Button
              type="text"
              icon={<DownloadOutlined />}
              onClick={() => window.open(grainPurchaseEntryExportApi.downloadUrl(record.batchNo), "_blank")}
            />
          </Tooltip>
        ) : null,
    },
  ];

  const imageCards = [
    { label: "身份证人像面", value: farmerImages?.idCardFront },
    { label: "身份证国徽面", value: farmerImages?.idCardBack },
    { label: "银行卡", value: farmerImages?.bankCard },
  ];
  const formSections = [
    { key: "ownership", index: "01", title: "业务归属", description: "确定粮站和负责业务员，用于本笔收粮记录归档。" },
    {
      key: "grain-farmer",
      index: "02",
      title: "粮食与农户身份信息",
      description: "维护粮食收购信息、粮户身份与联系方式，可通过身份证识别快速回填。",
    },
    { key: "payment", index: "03", title: "收款信息", description: "按收款方式录入对应账户信息，并补充收款时间和业务备注。" },
  ];

  return (
    <div className="manager-page-stack">
      <section className="manager-stats-grid" style={{ gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))" }}>
        {stats.map((item) => (
          <div key={item.label} className="manager-data-card">
            <div className="manager-section-label">{item.label}</div>
            <div className="manager-display-title" style={{ fontSize: 32, marginTop: 12 }}>
              {item.value}
            </div>
          </div>
        ))}
      </section>

      <section className="manager-data-card">
        <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
          <div style={{ display: "flex", gap: 12, flexWrap: "wrap", justifyContent: "space-between" }}>
            <Space wrap size={12}>
              <Select
                allowClear
                showSearch
                optionFilterProp="label"
                placeholder="粮站"
                value={filterStationId}
                onChange={(value: number | undefined) => {
                  setFilterStationId(value);
                  setFilterPurchaseTypeIds([]);
                }}
                options={stationOptions}
                style={{ minWidth: 200 }}
              />
              <Select
                mode="multiple"
                allowClear
                showSearch
                optionFilterProp="label"
                placeholder="收粮类型（可多选）"
                value={filterPurchaseTypeIds}
                onChange={setFilterPurchaseTypeIds}
                options={purchaseTypeOptions}
                loading={purchaseTypeLoading}
                style={{ minWidth: 200 }}
              />
              <DatePicker.RangePicker
                value={filterDateRange}
                onChange={(dates) => setFilterDateRange(dates as [Dayjs, Dayjs] | null)}
                format="YYYY-MM-DD"
                style={{ width: 260 }}
              />
              <Select
                mode="multiple"
                allowClear
                showSearch
                filterOption={false}
                placeholder="业务员（搜索姓名/用户名）"
                value={filterAppUserIds}
                onChange={setFilterAppUserIds}
                onSearch={(keyword) => void searchAppUsers(keyword)}
                options={filterAppUserOptions}
                loading={appUserSearching}
                notFoundContent={appUserSearching ? "搜索中..." : "输入姓名或用户名搜索"}
                style={{ minWidth: 200 }}
              />
              <Select
                mode="multiple"
                allowClear
                showSearch
                filterOption={false}
                placeholder="农户（姓名前缀/身份证后4位/手机号）"
                value={filterFarmerIds}
                onChange={setFilterFarmerIds}
                onSearch={(keyword) => void searchFarmers(keyword)}
                options={filterFarmerOptions}
                loading={farmerSearching}
                notFoundContent={farmerSearching ? "搜索中..." : "输入姓名前缀、身份证或手机号搜索"}
                style={{ minWidth: 180 }}
              />
            </Space>
            <Space wrap>
              <Tag style={{ color: "var(--manager-text-soft)", background: "var(--manager-green-soft)", border: "none" }}>
                共 {total} 条
              </Tag>
            </Space>
          </div>
          <Space wrap size={12}>
            <InputNumber
              placeholder="最低金额"
              min={0}
              precision={2}
              value={filterMinAmount}
              onChange={(value) => setFilterMinAmount(value ?? undefined)}
              addonBefore="￥"
              style={{ width: 150 }}
            />
            <InputNumber
              placeholder="最高金额"
              min={0}
              precision={2}
              value={filterMaxAmount}
              onChange={(value) => setFilterMaxAmount(value ?? undefined)}
              addonBefore="￥"
              style={{ width: 150 }}
            />
            <Button type="primary" icon={<SearchOutlined />} onClick={() => void loadRecords(filterQuery())}>
              查询
            </Button>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => {
                setFilterStationId(undefined);
                setFilterPurchaseTypeIds([]);
                setFilterDateRange(null);
                setFilterAppUserIds([]);
                setFilterFarmerIds([]);
                setFilterFarmerOptions([]);
                setFilterAppUserOptions([]);
                setFilterMinAmount(undefined);
                setFilterMaxAmount(undefined);
                void loadRecords({ pageIndex: 1 });
              }}
            >
              重置
            </Button>
            <Button icon={<ExportOutlined />} onClick={() => void openExport()}>
              导出
            </Button>
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
          scroll={{ x: 2720 }}
          pagination={{
            current: query.pageIndex,
            pageSize: query.pageSize,
            total,
            showSizeChanger: false,
            onChange: (page) => void loadRecords({ ...filterQuery(), pageIndex: page }),
          }}
        />
      </section>

      <Modal
        title="导出收粮明细"
        open={exportOpen}
        width={1080}
        footer={<Button onClick={() => setExportOpen(false)}>关闭</Button>}
        onCancel={() => setExportOpen(false)}
        destroyOnClose={false}
      >
        <div style={{ display: "grid", gap: 16 }}>
          <section style={{ display: "grid", gap: 12 }}>
            <Space wrap size={12}>
              <Select
                allowClear
                showSearch
                optionFilterProp="label"
                placeholder="粮站"
                value={filterStationId}
                onChange={(value: number | undefined) => {
                  setFilterStationId(value);
                  setFilterPurchaseTypeIds([]);
                  setExportCount(null);
                }}
                options={stationOptions}
                style={{ minWidth: 180 }}
              />
              <Select
                mode="multiple"
                allowClear
                showSearch
                optionFilterProp="label"
                placeholder="收粮类型"
                value={filterPurchaseTypeIds}
                onChange={(value) => {
                  setFilterPurchaseTypeIds(value);
                  setExportCount(null);
                }}
                options={purchaseTypeOptions}
                loading={purchaseTypeLoading}
                style={{ minWidth: 180 }}
              />
              <DatePicker.RangePicker
                value={filterDateRange}
                onChange={(dates) => {
                  setFilterDateRange(dates as [Dayjs, Dayjs] | null);
                  setExportCount(null);
                }}
                format="YYYY-MM-DD"
                style={{ width: 250 }}
              />
              <Select
                mode="multiple"
                allowClear
                showSearch
                filterOption={false}
                placeholder="业务员"
                value={filterAppUserIds}
                onChange={(value) => {
                  setFilterAppUserIds(value);
                  setExportCount(null);
                }}
                onSearch={(keyword) => void searchAppUsers(keyword)}
                options={filterAppUserOptions}
                loading={appUserSearching}
                style={{ minWidth: 180 }}
              />
              <Select
                mode="multiple"
                allowClear
                showSearch
                filterOption={false}
                placeholder="农户"
                value={filterFarmerIds}
                onChange={(value) => {
                  setFilterFarmerIds(value);
                  setExportCount(null);
                }}
                onSearch={(keyword) => void searchFarmers(keyword)}
                options={filterFarmerOptions}
                loading={farmerSearching}
                style={{ minWidth: 180 }}
              />
              <InputNumber
                placeholder="最低金额"
                min={0}
                precision={2}
                value={filterMinAmount}
                onChange={(value) => {
                  setFilterMinAmount(value ?? undefined);
                  setExportCount(null);
                }}
                addonBefore="￥"
                style={{ width: 140 }}
              />
              <InputNumber
                placeholder="最高金额"
                min={0}
                precision={2}
                value={filterMaxAmount}
                onChange={(value) => {
                  setFilterMaxAmount(value ?? undefined);
                  setExportCount(null);
                }}
                addonBefore="￥"
                style={{ width: 140 }}
              />
            </Space>
            <Space wrap>
              <Button type="primary" icon={<SearchOutlined />} loading={exportCounting} onClick={() => void handleExportCount()}>
                筛选
              </Button>
              <Button
                icon={<ReloadOutlined />}
                onClick={() => {
                  setFilterStationId(undefined);
                  setFilterPurchaseTypeIds([]);
                  setFilterDateRange(null);
                  setFilterAppUserIds([]);
                  setFilterFarmerIds([]);
                  setFilterFarmerOptions([]);
                  setFilterAppUserOptions([]);
                  setFilterMinAmount(undefined);
                  setFilterMaxAmount(undefined);
                  setExportCount(null);
                }}
              >
                重置条件
              </Button>
              <Button
                type="primary"
                icon={<ExportOutlined />}
                disabled={exportCount === null || exportCount <= 0}
                loading={exportStarting}
                onClick={() => void handleStartExport()}
              >
                下一步导出
              </Button>
              {exportCount !== null ? <Tag color="green">符合条件 {exportCount} 条</Tag> : null}
            </Space>
          </section>

          <Divider style={{ margin: "4px 0" }} />
          <Table<GrainPurchaseEntryExportBatchRecord>
            rowKey="batchNo"
            size="small"
            loading={exportLoading}
            dataSource={exportBatches}
            columns={exportColumns}
            scroll={{ x: 1050 }}
            pagination={{
              current: exportPageIndex,
              pageSize: 5,
              total: exportTotal,
              showSizeChanger: false,
              onChange: (page) => void loadExportBatches(page),
            }}
          />
        </div>
      </Modal>

      <Drawer
        title={null}
        open={drawerOpen}
        width={1180}
        destroyOnClose
        className="manager-crud-drawer"
        onClose={() => {
          setDrawerOpen(false);
          setEditingRecord(null);
          setFarmerImages(null);
          setEntryMaterials([]);
          form.resetFields();
        }}
        footer={
          <div className="manager-crud-drawer-footer">
            <Button
              onClick={() => {
                setDrawerOpen(false);
                setEditingRecord(null);
                setFarmerImages(null);
                setEntryMaterials([]);
                form.resetFields();
              }}
            >
              取消
            </Button>
            <Button type="primary" loading={submitting} onClick={() => void handleSubmit()}>
              保存
            </Button>
          </div>
        }
      >
        <div className="manager-crud-drawer-head">
          <div>
            <div className="manager-section-label">{editingRecord ? "编辑记录" : "新增记录"}</div>
            <Typography.Title level={3} style={{ margin: "6px 0 0" }}>
              {editingRecord ? `编辑收粮明细 #${editingRecord.id}` : "新增收粮明细"}
            </Typography.Title>
          </div>
          <Tag color={editingRecord ? "blue" : "green"}>{editingRecord ? `ID ${editingRecord.id}` : "新建"}</Tag>
        </div>

        <Form form={form} layout="vertical" preserve={false} className="manager-crud-form">
          <aside className="manager-crud-form-rail">
            {formSections.map((section) => (
              <a key={section.key} href={`#entry-${section.key}`}>
                <span>{section.index}</span>
                {section.title}
              </a>
            ))}
          </aside>

          <div className="manager-crud-form-main grain-entry-form-main">
            <div className="grain-entry-form-fields">
              <Form.Item name="farmerId" hidden>
                <InputNumber />
              </Form.Item>

              <section id="entry-ownership" className="manager-form-section">
                <div className="manager-form-section-head">
                  <div>
                    <Typography.Title level={5} style={{ margin: 0 }}>
                      {formSections[0].title}
                    </Typography.Title>
                    <Text type="secondary">{formSections[0].description}</Text>
                  </div>
                </div>
                <div className="manager-form-grid">
                  <Form.Item name="stationId" label="粮站" rules={[{ required: true, message: "请选择粮站" }]}>
                    <Select options={stationOptions} placeholder="请选择粮站" showSearch optionFilterProp="label" />
                  </Form.Item>
                  <Form.Item name="appUserId" label="业务员">
                    <Select options={appUserOptions} placeholder="请选择业务员" allowClear showSearch optionFilterProp="label" />
                  </Form.Item>
                </div>
              </section>

              <section id="entry-grain-farmer" className="manager-form-section">
                <div className="manager-form-section-head">
                  <div>
                    <Typography.Title level={5} style={{ margin: 0 }}>
                      {formSections[1].title}
                    </Typography.Title>
                    <Text type="secondary">{formSections[1].description}</Text>
                  </div>
                  <Space wrap>
                    <Tooltip title="识别身份证正面">
                      <Button
                        icon={<ScanOutlined />}
                        onClick={() => setOcrTarget({ title: "识别身份证正面", cardType: "id-card", imageSide: "front" })}
                      >
                        身份证
                      </Button>
                    </Tooltip>
                  </Space>
                </div>
                <Alert
                  type="info"
                  showIcon
                  style={{ marginBottom: 16 }}
                  message="可拍身份证后自动带入，也可以直接手动输入；编辑已有记录时识别结果会同步回粮户资料。"
                />
                <div className="manager-form-grid">
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
                  <Form.Item name="purchaseTypeId" label="收购类型">
                    <Select
                      options={purchaseTypeOptions}
                      placeholder="请选择收购类型"
                      allowClear
                      showSearch
                      optionFilterProp="label"
                      loading={purchaseTypeLoading}
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
                  <Form.Item name="locationAddress" label="收购地址" className="manager-form-item-wide">
                    <Input placeholder="请输入收购地址" />
                  </Form.Item>
                </div>
              </section>

              <section id="entry-payment" className="manager-form-section">
                <div className="manager-form-section-head">
                  <div>
                    <Typography.Title level={5} style={{ margin: 0 }}>
                      {formSections[2].title}
                    </Typography.Title>
                    <Text type="secondary">{formSections[2].description}</Text>
                  </div>
                  {bankPaymentSelected ? (
                    <Space wrap>
                      <Tooltip title="识别银行卡">
                        <Button icon={<BankOutlined />} onClick={() => setOcrTarget({ title: "识别银行卡", cardType: "bank-card" })}>
                          银行卡
                        </Button>
                      </Tooltip>
                      <Tooltip title="识别社保卡">
                        <Button icon={<BankOutlined />} onClick={() => setOcrTarget({ title: "识别社保卡", cardType: "social-security-card" })}>
                          社保卡
                        </Button>
                      </Tooltip>
                    </Space>
                  ) : null}
                </div>
                <div className="manager-form-grid">
                  <Form.Item name="paymentMethodId" label="收款方式">
                    <Select
                      options={paymentMethodOptions}
                      placeholder="请选择收款方式"
                      allowClear
                      showSearch
                      optionFilterProp="label"
                      onChange={(_, option) => {
                        const selected = Array.isArray(option) ? option[0] : option;
                        if (selected?.label) {
                          form.setFieldValue("payType", String(selected.label));
                        }
                      }}
                    />
                  </Form.Item>
                  <Form.Item name="payType" label="收款方式名称">
                    <Input placeholder="请输入收款方式" />
                  </Form.Item>
                  <Form.Item name="payTime" label="收款时间">
                    <DatePicker showTime format="YYYY-MM-DD HH:mm:ss" placeholder="请选择收款时间" style={{ width: "100%" }} />
                  </Form.Item>
                  {hasPaymentMethod ? (
                    <>
                      <Form.Item name="farmerBankName" label={bankPaymentSelected ? "开户人" : "收款人姓名"}>
                        <Input placeholder={bankPaymentSelected ? "请输入开户人姓名" : "请输入收款人姓名"} />
                      </Form.Item>
                      <Form.Item name="farmerBankNumber" label={bankPaymentSelected ? "银行卡号" : "收款账号"}>
                        <Input
                          placeholder={bankPaymentSelected ? "拍银行卡自动带入，或手动输入" : "请输入收款账号"}
                          prefix={bankPaymentSelected ? <BankOutlined /> : <WalletOutlined />}
                        />
                      </Form.Item>
                    </>
                  ) : null}
                  <Form.Item name="status" label="状态">
                    <Select options={statusOptions} placeholder="请选择状态" />
                  </Form.Item>
                  <Form.Item name="remark" label="备注" className="manager-form-item-wide">
                    <Input.TextArea rows={3} placeholder="请输入备注" />
                  </Form.Item>
                </div>
              </section>
            </div>

            <aside className="grain-entry-form-aside">
              <section className="manager-form-assist">
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
                  <Text strong>农户照片</Text>
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
                  <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无农户照片" />
                )}
              </section>

              <section className="manager-form-section">
                <Space align="center" style={{ marginBottom: 12 }}>
                  <CameraOutlined style={{ color: "#237a4b" }} />
                  <Text strong>收粮材料</Text>
                  {editingRecord && (
                    <Upload
                      accept="image/*"
                      showUploadList={false}
                      beforeUpload={(file) => {
                        void handleMaterialUpload(file as File);
                        return false;
                      }}
                    >
                      <Button
                        size="small"
                        type="dashed"
                        icon={<PlusOutlined />}
                        loading={materialUploading}
                      >
                        上传材料
                      </Button>
                    </Upload>
                  )}
                </Space>
                {entryMaterials.length > 0 ? (
                  <div style={{ display: "grid", gap: 12 }}>
                    {entryMaterials.map((item) => (
                      <div key={item.id} style={{ position: "relative" }}>
                        <Image
                          src={item.imageUrl}
                          alt={item.fileName || "收粮材料"}
                          style={{ width: "100%", height: 120, objectFit: "cover", borderRadius: 8 }}
                        />
                        <Popconfirm
                          title="确认删除该材料图片？"
                          onConfirm={() => void handleMaterialDelete(item.id)}
                          okText="删除"
                          okButtonProps={{ danger: true }}
                          cancelText="取消"
                        >
                          <Button
                            size="small"
                            danger
                            icon={<DeleteOutlined />}
                            style={{
                              position: "absolute",
                              top: 4,
                              right: 4,
                              opacity: 0.85,
                            }}
                          />
                        </Popconfirm>
                      </div>
                    ))}
                  </div>
                ) : (
                  <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无收粮材料图片" />
                )}
              </section>
            </aside>
          </div>
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
                  <Descriptions.Item label="开户人">{renderValue(selectedSnapshot.farmerBankName)}</Descriptions.Item>
                  <Descriptions.Item label="手机号">{renderValue(selectedSnapshot.farmerPhone)}</Descriptions.Item>
                  <Descriptions.Item label="收购类型">{renderValue(selectedSnapshot.crop)}</Descriptions.Item>
                  <Descriptions.Item label="重量">
                    {renderValue(selectedSnapshot.quantity)} {selectedSnapshot.unit || "公斤"}
                  </Descriptions.Item>
                  <Descriptions.Item label="金额">￥{normalizeMoney(Number(selectedSnapshot.amount))}</Descriptions.Item>
                  <Descriptions.Item label="单价">￥{normalizeMoney(Number(selectedSnapshot.unitPrice))}</Descriptions.Item>
                  <Descriptions.Item label="收购地址" span={2}>
                    {renderValue(selectedSnapshot.locationAddress)}
                  </Descriptions.Item>
                  <Descriptions.Item label="收款方式">{renderValue(selectedSnapshot.payType)}</Descriptions.Item>
                  <Descriptions.Item label="收款时间">{formatTime(selectedSnapshot.payTime)}</Descriptions.Item>
                  <Descriptions.Item label="状态">{snapshotStatusTag(selectedSnapshot.entryStatus)}</Descriptions.Item>
                  <Descriptions.Item label="备注">{renderValue(selectedSnapshot.entryRemark)}</Descriptions.Item>
                  <Descriptions.Item label="材料数量">{renderValue(selectedSnapshot.materialCount)}</Descriptions.Item>
                  <Descriptions.Item label="材料摘要">{renderMaterialSummary(selectedSnapshot.materialSummary)}</Descriptions.Item>
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
          <p className="ant-upload-hint">支持身份证正面、银行卡和社保卡照片；识别失败时可直接回到表单手动输入。</p>
        </Upload.Dragger>
      </Modal>
    </div>
  );
}
