"use client";

import { useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import {
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
} from "@ant-design/icons";
import {
  Button,
  Cascader,
  Drawer,
  Form,
  Input,
  InputNumber,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
  Tooltip,
  Typography,
  message,
} from "antd";
import type { FormInstance } from "antd";
import type { ColumnsType } from "antd/es/table";
import type { PageResult } from "@/utils/axios";

const { Text } = Typography;

export interface CrudRecord {
  id: number;
  createdTime?: string;
  updatedTime?: string;
  [key: string]: unknown;
}

export interface CrudOption {
  label: string;
  value: string | number | boolean;
}

export interface CrudCascaderOption {
  label: string;
  value: string;
  children?: CrudCascaderOption[];
}

export interface CrudField<R extends CrudRecord> {
  name: Extract<keyof R, string> | string;
  label: string;
  type?: "text" | "textarea" | "number" | "select" | "password" | "cascader";
  required?: boolean;
  placeholder?: string;
  options?: CrudOption[];
  mode?: "multiple";
  cascaderOptions?: CrudCascaderOption[];
  linkedNames?: [Extract<keyof R, string>, Extract<keyof R, string>, Extract<keyof R, string>];
  min?: number;
  precision?: number;
  hiddenOnCreate?: boolean;
  hiddenOnEdit?: boolean;
  disabledOnEdit?: boolean;
  span?: 1 | 2;
  help?: string;
}

export interface CrudFormSection<R extends CrudRecord> {
  key: string;
  title: string;
  description?: string;
  fields: Array<CrudField<R>["name"]>;
}

export interface CrudTableColumn<R extends CrudRecord> {
  name: Extract<keyof R, string>;
  label: string;
  width?: number;
  copyable?: boolean;
  render?: (value: unknown, record: R) => ReactNode;
}

export interface CrudListQuery {
  pageIndex?: number;
  pageSize?: number;
  [key: string]: string | number | undefined;
}

interface CrudApi<R extends CrudRecord, P extends Record<string, unknown>> {
  list: (query: CrudListQuery) => Promise<PageResult<R>>;
  create: (payload: P) => Promise<unknown>;
  update: (id: number, payload: Partial<P>) => Promise<unknown>;
  remove: (id: number) => Promise<unknown>;
}

interface CrudActionContext {
  reload: () => Promise<void>;
  submitting: boolean;
  setSubmitting: (submitting: boolean) => void;
}

interface CrudManagementPanelProps<R extends CrudRecord, P extends Record<string, unknown>> {
  title: string;
  createText: string;
  searchPlaceholder: string;
  searchParam: string;
  fields: CrudField<R>[];
  columns: CrudTableColumn<R>[];
  api: CrudApi<R, P>;
  statusField?: Extract<keyof R, string>;
  statusOptions?: CrudOption[];
  rowActions?: (record: R, context: CrudActionContext) => ReactNode;
  actionWidth?: number;
  modalWidth?: number;
  formSections?: CrudFormSection<R>[];
  extraFilters?: { param: string; placeholder: string; type?: "text" | "select"; options?: CrudOption[]; width?: number }[];
  formExtra?: (context: {
    form: FormInstance;
    editingRecord: R | null;
    submitting: boolean;
    setSubmitting: (submitting: boolean) => void;
    reload: () => Promise<void>;
  }) => ReactNode;
}

const defaultPageSize = 10;

function compactPayload(values: Record<string, unknown>) {
  return Object.fromEntries(
    Object.entries(values).filter(([, value]) => value !== undefined && value !== ""),
  );
}

function renderValue(value: unknown) {
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  if (typeof value === "boolean") {
    return value ? "是" : "否";
  }
  return String(value);
}

function buildInitialFormValues<R extends CrudRecord>(record: R, fields: CrudField<R>[]) {
  const values: Record<string, unknown> = { ...record };
  fields.forEach((field) => {
    if (field.type !== "cascader" || !field.linkedNames) {
      return;
    }
    values[field.name] = field.linkedNames
      .map((name) => record[name])
      .filter((value) => typeof value === "string" && value !== "");
  });
  return values;
}

function applyLinkedFieldValues<R extends CrudRecord>(values: Record<string, unknown>, fields: CrudField<R>[]) {
  fields.forEach((field) => {
    if (field.type !== "cascader" || !field.linkedNames) {
      return;
    }
    const fieldName = String(field.name);
    const fieldValue = values[fieldName];
    const selectedValues = Array.isArray(fieldValue) ? fieldValue : [];
    field.linkedNames.forEach((name, index) => {
      values[name] = selectedValues[index] ?? "";
    });
    delete values[fieldName];
  });
}

function statusTag(value: unknown, label?: string) {
  const text = label ?? renderValue(value);
  const normalized = String(value ?? "").toUpperCase();
  const color =
    normalized.includes("FAILED") ||
    normalized.includes("FROZEN") ||
    normalized.includes("EXPIRED") ||
    normalized.includes("OFFLINE") ||
    normalized.includes("INACTIVE")
      ? "red"
      : normalized.includes("LOCKED")
        ? "orange"
        : normalized.includes("SUCCESS") ||
            normalized.includes("NORMAL") ||
            normalized === "ACTIVE" ||
            normalized.includes("LOGGED_IN")
          ? "green"
          : "blue";

  return <Tag color={color}>{text}</Tag>;
}

export function CrudManagementPanel<R extends CrudRecord, P extends Record<string, unknown>>({
  title,
  createText,
  searchPlaceholder,
  searchParam,
  fields,
  columns,
  api,
  statusField,
  statusOptions,
  rowActions,
  actionWidth = 132,
  modalWidth = 720,
  formSections,
  extraFilters,
  formExtra,
}: CrudManagementPanelProps<R, P>) {
  const [form] = Form.useForm();
  const [records, setRecords] = useState<R[]>([]);
  const [total, setTotal] = useState(0);
  const [query, setQuery] = useState<Required<Pick<CrudListQuery, "pageIndex" | "pageSize">>>({
    pageIndex: 1,
    pageSize: defaultPageSize,
  });
  const [searchValue, setSearchValue] = useState("");
  const [statusValue, setStatusValue] = useState<string | number | undefined>();
  const [extraFilterValues, setExtraFilterValues] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<R | null>(null);

  const loadRecords = async (nextQuery?: CrudListQuery) => {
    const mergedQuery = { ...query, ...nextQuery };
    setLoading(true);
    try {
      const result = await api.list(mergedQuery);
      setRecords(result.data);
      setTotal(result.total);
      setQuery({
        pageIndex: Number(mergedQuery.pageIndex ?? 1),
        pageSize: Number(mergedQuery.pageSize ?? defaultPageSize),
      });
    } catch (error) {
      message.error(error instanceof Error ? error.message : "加载数据失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadRecords();
  }, []);

  const stats = useMemo(
    () => [
      { label: `${title}总数`, value: total },
      { label: "当前页数量", value: records.length },
      {
        label: "最近更新",
        value: records[0]?.updatedTime ? String(records[0].updatedTime).slice(0, 10) : "-",
      },
    ],
    [records, title, total],
  );

  const tableColumns: ColumnsType<R> = [
    {
      title: "ID",
      dataIndex: "id",
      width: 80,
      fixed: "left",
    },
    ...columns.map((column) => ({
      title: column.label,
      dataIndex: column.name,
      width: column.width,
      render: (value: unknown, record: R) => {
        if (column.render) {
          return column.render(value, record);
        }
        if (column.name === statusField) {
          const option = statusOptions?.find((item) => item.value === value);
          return statusTag(value, option?.label);
        }
        return column.copyable ? (
          <Text copyable style={{ color: "var(--manager-text)" }}>
            {renderValue(value)}
          </Text>
        ) : (
          renderValue(value)
        );
      },
    })),
    {
      title: "操作",
      key: "actions",
      width: actionWidth,
      fixed: "right",
      render: (_, record) => (
        <Space size={4}>
          {rowActions?.(record, {
            reload: () => loadRecords({ pageIndex: query.pageIndex }),
            submitting,
            setSubmitting,
          })}
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => {
                setEditingRecord(record);
                form.setFieldsValue(buildInitialFormValues(record, fields));
                setModalOpen(true);
              }}
            />
          </Tooltip>
          <Popconfirm
            title={`确认删除这条${title}记录吗？`}
            okText="删除"
            cancelText="取消"
            onConfirm={async () => {
              try {
                setSubmitting(true);
                await api.remove(record.id);
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

  const filterQuery = () => ({
    pageIndex: 1,
    [searchParam]: searchValue.trim() || undefined,
    ...(statusField ? { [statusField]: statusValue } : {}),
    ...Object.fromEntries(
      Object.entries(extraFilterValues).map(([k, v]) => [k, v.trim() || undefined]),
    ),
  });

  const visibleFields = fields
    .filter((field) => !(field.hiddenOnCreate && !editingRecord))
    .filter((field) => !(field.hiddenOnEdit && editingRecord));

  const visibleFieldNames = new Set(visibleFields.map((field) => String(field.name)));
  const visibleSections = formSections
    ?.map((section) => ({
      ...section,
      fields: section.fields
        .map((name) => visibleFields.find((field) => String(field.name) === String(name)))
        .filter((field): field is CrudField<R> => Boolean(field)),
    }))
    .filter((section) => section.fields.length > 0);

  const groupedFieldNames = new Set(
    visibleSections?.flatMap((section) => section.fields.map((field) => String(field.name))) ?? [],
  );
  const fallbackFields = visibleFields.filter((field) => !groupedFieldNames.has(String(field.name)));
  const formFieldGroups =
    visibleSections && visibleSections.length > 0
      ? [
          ...visibleSections,
          ...(fallbackFields.length > 0
            ? [{ key: "other", title: "其他信息", fields: fallbackFields }]
            : []),
        ]
      : [{ key: "basic", title: "基础信息", fields: visibleFields }];

  const renderFormControl = (field: CrudField<R>) => {
    const disabled = field.disabledOnEdit && Boolean(editingRecord);
    if (field.type === "textarea") {
      return <Input.TextArea rows={4} placeholder={field.placeholder} disabled={disabled} />;
    }
    if (field.type === "number") {
      return (
        <InputNumber
          min={field.min}
          precision={field.precision}
          style={{ width: "100%" }}
          placeholder={field.placeholder}
          disabled={disabled}
        />
      );
    }
    if (field.type === "select") {
      return (
        <Select allowClear mode={field.mode} placeholder={field.placeholder} options={field.options} disabled={disabled} />
      );
    }
    if (field.type === "cascader") {
      return (
        <Cascader
          allowClear
          showSearch
          options={field.cascaderOptions}
          placeholder={field.placeholder}
          disabled={disabled}
          changeOnSelect={false}
        />
      );
    }
    if (field.type === "password") {
      return <Input.Password placeholder={field.placeholder} disabled={disabled} />;
    }
    return <Input placeholder={field.placeholder} disabled={disabled} />;
  };

  const handleSubmit = async () => {
    const rawValues = compactPayload(form.getFieldsValue());
    applyLinkedFieldValues(rawValues, fields);
    if (editingRecord) {
      fields.forEach((field) => {
        if (field.disabledOnEdit) {
          delete rawValues[field.name];
        }
      });
    }
    const values = rawValues as P;
    setSubmitting(true);
    try {
      if (editingRecord) {
        await api.update(editingRecord.id, values);
        message.success("更新成功");
      } else {
        await api.create(values);
        message.success("创建成功");
      }
      setModalOpen(false);
      setEditingRecord(null);
      form.resetFields();
      await loadRecords({ pageIndex: editingRecord ? query.pageIndex : 1 });
    } catch (error) {
      message.error(error instanceof Error ? error.message : "保存失败");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="manager-page-stack">
      <section
        className="manager-stats-grid"
        style={{ gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))" }}
      >
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
        <div style={{ display: "flex", gap: 12, flexWrap: "wrap", justifyContent: "space-between" }}>
          <Space wrap size={12}>
            <Input
              className="manager-filter-input"
              prefix={<SearchOutlined style={{ color: "var(--manager-text-faint)" }} />}
              placeholder={searchPlaceholder}
              value={searchValue}
              onChange={(event) => setSearchValue(event.target.value)}
              onPressEnter={() => void loadRecords(filterQuery())}
              style={{ width: 260, maxWidth: "100%" }}
            />
            {extraFilters?.map((filter) =>
              filter.type === "select" ? (
                <Select
                  key={filter.param}
                  allowClear
                  placeholder={filter.placeholder}
                  value={extraFilterValues[filter.param] || undefined}
                  onChange={(value: string | undefined) =>
                    setExtraFilterValues((prev) => ({ ...prev, [filter.param]: value ?? "" }))
                  }
                  options={filter.options}
                  style={{ width: filter.width ?? 200 }}
                />
              ) : (
                <Input
                  key={filter.param}
                  className="manager-filter-input"
                  placeholder={filter.placeholder}
                  value={extraFilterValues[filter.param] ?? ""}
                  onChange={(event) =>
                    setExtraFilterValues((prev) => ({ ...prev, [filter.param]: event.target.value }))
                  }
                  onPressEnter={() => void loadRecords(filterQuery())}
                  style={{ width: filter.width ?? 200 }}
                />
              ),
            )}
            {statusField && statusOptions ? (
              <Select
                allowClear
                placeholder="状态"
                value={statusValue}
                onChange={setStatusValue}
                options={statusOptions}
                style={{ width: 180 }}
              />
            ) : null}
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
              onClick={() => {
                setEditingRecord(null);
                form.resetFields();
                setModalOpen(true);
              }}
              style={{
                color: "#ffffff",
                border: "none",
                background: "linear-gradient(135deg, #145535 0%, #237a4b 100%)",
              }}
            >
              {createText}
            </Button>
          </Space>
        </div>
      </section>

      <section className="manager-data-card manager-table">
        <Table<R>
          rowKey="id"
          loading={loading}
          dataSource={records}
          columns={tableColumns}
          scroll={{ x: Math.max(1100, columns.reduce((sum, item) => sum + (item.width ?? 160), 220)) }}
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
        title={null}
        open={modalOpen}
        width={modalWidth}
        destroyOnClose
        className="manager-crud-drawer"
        onClose={() => {
          setModalOpen(false);
          setEditingRecord(null);
          form.resetFields();
        }}
        footer={
          <div className="manager-crud-drawer-footer">
            <Button
              onClick={() => {
                setModalOpen(false);
                setEditingRecord(null);
                form.resetFields();
              }}
            >
              取消
            </Button>
            <Button type="primary" loading={submitting} onClick={() => void form.validateFields().then(handleSubmit)}>
              保存
            </Button>
          </div>
        }
      >
        <div className="manager-crud-drawer-head">
          <div>
            <div className="manager-section-label">{editingRecord ? "编辑记录" : "新增记录"}</div>
            <Typography.Title level={3} style={{ margin: "6px 0 0" }}>
              {editingRecord ? `编辑${title}` : createText}
            </Typography.Title>
          </div>
          <Tag color={editingRecord ? "blue" : "green"}>{editingRecord ? `ID ${editingRecord.id}` : "新建"}</Tag>
        </div>

        <Form form={form} layout="vertical" preserve={false} className="manager-crud-form">
          <aside className="manager-crud-form-rail">
            {formFieldGroups.map((section, index) => (
              <a key={section.key} href={`#${title}-${section.key}`}>
                <span>{String(index + 1).padStart(2, "0")}</span>
                {section.title}
              </a>
            ))}
          </aside>

          <div className="manager-crud-form-main">
            {formExtra?.({
              form,
              editingRecord,
              submitting,
              setSubmitting,
              reload: () => loadRecords({ pageIndex: query.pageIndex }),
            })}
            {formFieldGroups.map((section) => (
              <section key={section.key} id={`${title}-${section.key}`} className="manager-form-section">
                <div className="manager-form-section-head">
                  <div>
                    <Typography.Title level={5} style={{ margin: 0 }}>
                      {section.title}
                    </Typography.Title>
                    {section.description ? <Text type="secondary">{section.description}</Text> : null}
                  </div>
                </div>
                <div className="manager-form-grid">
                  {section.fields
                    .filter((field) => visibleFieldNames.has(String(field.name)))
                    .map((field) => (
                      <Form.Item
                        key={field.name}
                        className={field.span === 2 || field.type === "textarea" ? "manager-form-item-wide" : undefined}
                        name={field.name as string}
                        label={field.label}
                        help={field.help}
                        rules={field.required ? [{ required: true, message: `请输入${field.label}` }] : undefined}
                      >
                        {renderFormControl(field)}
                      </Form.Item>
                    ))}
                </div>
              </section>
            ))}
          </div>
        </Form>
      </Drawer>
    </div>
  );
}
