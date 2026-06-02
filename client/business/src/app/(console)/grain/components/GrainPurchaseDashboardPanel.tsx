"use client";

import { useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import {
  AppstoreOutlined,
  BankOutlined,
  BarChartOutlined,
  CalendarOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  FieldTimeOutlined,
  GoldOutlined,
  LineChartOutlined,
  ReloadOutlined,
  TableOutlined,
  TeamOutlined,
  UserAddOutlined,
  WalletOutlined,
} from "@ant-design/icons";
import { Button, DatePicker, Empty, Progress, Segmented, Select, Space, Table, Tag, Tooltip, Typography, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import dayjs from "dayjs";
import type { Dayjs } from "dayjs";
import type { CrudOption } from "../../components/CrudManagementPanel";
import {
  getGrainPurchaseDashboard,
  grainStationApi,
  type GrainPurchaseDashboardDimensionRecord,
  type GrainPurchaseDashboardRecord,
} from "../api/grain.api";

const { RangePicker } = DatePicker;
const { Text, Title } = Typography;

type DimensionMode = "chart" | "table";
type RangeValue = [Dayjs | null, Dayjs | null] | null;

const today = dayjs();

function formatMoney(value: number) {
  return `¥${new Intl.NumberFormat("zh-CN", { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(value || 0)}`;
}

function formatNumber(value: number, digits = 0) {
  return new Intl.NumberFormat("zh-CN", { maximumFractionDigits: digits }).format(value || 0);
}

function formatWeight(value: number) {
  if (value >= 1000) {
    return `${formatNumber(value / 1000, 2)} 吨`;
  }
  return `${formatNumber(value, 3)} 公斤`;
}

function sharePercent(value: number) {
  return Math.min(100, Math.max(0, Number(((value || 0) * 100).toFixed(1))));
}

function dateParams(range: RangeValue) {
  const start = range?.[0] ?? today;
  const end = range?.[1] ?? start;
  return {
    startDate: start.format("YYYY-MM-DD"),
    endDate: end.format("YYYY-MM-DD"),
  };
}

function sortRows(rows: GrainPurchaseDashboardDimensionRecord[]) {
  return [...rows].sort((a, b) => b.totalAmount - a.totalAmount || b.totalQuantity - a.totalQuantity);
}

function DimensionVisual({
  rows,
  accent,
  emptyText,
}: {
  rows: GrainPurchaseDashboardDimensionRecord[];
  accent: string;
  emptyText: string;
}) {
  const sortedRows = sortRows(rows).slice(0, 8);
  const maxAmount = Math.max(...sortedRows.map((item) => item.totalAmount), 0);
  if (sortedRows.length === 0) {
    return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={emptyText} />;
  }
  return (
    <div className="grain-dashboard-rank">
      {sortedRows.map((item, index) => {
        const width = maxAmount > 0 ? Math.max(8, (item.totalAmount / maxAmount) * 100) : 8;
        return (
          <div className="grain-dashboard-rank__row" key={item.key || item.name}>
            <div className="grain-dashboard-rank__meta">
              <span className="grain-dashboard-rank__index">{index + 1}</span>
              <div>
                <strong>{item.name || "未命名"}</strong>
                <span>
                  {formatWeight(item.totalQuantity)} · {item.entryCount} 笔 · {item.farmerCount} 户
                </span>
              </div>
            </div>
            <div className="grain-dashboard-rank__track">
              <span style={{ width: `${width}%`, background: accent }} />
            </div>
            <div className="grain-dashboard-rank__value">
              <strong>{formatMoney(item.totalAmount)}</strong>
              <span>{sharePercent(item.amountShare)}%</span>
            </div>
          </div>
        );
      })}
    </div>
  );
}

function DimensionSection({
  title,
  subtitle,
  icon,
  rows,
  mode,
  onModeChange,
  accent,
  emptyText,
}: {
  title: string;
  subtitle: string;
  icon: ReactNode;
  rows: GrainPurchaseDashboardDimensionRecord[];
  mode: DimensionMode;
  onModeChange: (value: DimensionMode) => void;
  accent: string;
  emptyText: string;
}) {
  const columns: ColumnsType<GrainPurchaseDashboardDimensionRecord> = [
    {
      title: title.includes("粮站") ? "粮站" : "粮食类型",
      dataIndex: "name",
      render: (value: string, record) => (
        <Space direction="vertical" size={0}>
          <Text strong>{value || "未命名"}</Text>
          <Text type="secondary">{record.entryCount} 笔收粮</Text>
        </Space>
      ),
    },
    {
      title: "农户",
      dataIndex: "farmerCount",
      width: 96,
      render: (value: number) => `${value} 户`,
    },
    {
      title: "业务员",
      dataIndex: "activeUserCount",
      width: 96,
      render: (value: number) => `${value} 人`,
    },
    {
      title: "重量",
      dataIndex: "totalQuantity",
      render: (value: number) => formatWeight(value),
    },
    {
      title: "金额",
      dataIndex: "totalAmount",
      render: (value: number) => <Text strong>{formatMoney(value)}</Text>,
    },
    {
      title: "占比",
      dataIndex: "amountShare",
      width: 140,
      render: (value: number) => <Progress percent={sharePercent(value)} size="small" strokeColor={accent} />,
    },
  ];
  return (
    <section className="grain-dashboard-panel">
      <div className="grain-dashboard-panel__header">
        <Space size={12}>
          <span className="grain-dashboard-panel__icon">{icon}</span>
          <span>
            <Title level={3}>{title}</Title>
            <Text>{subtitle}</Text>
          </span>
        </Space>
        <Segmented
          value={mode}
          onChange={(value) => onModeChange(value as DimensionMode)}
          options={[
            { label: <Tooltip title="图表"><BarChartOutlined /></Tooltip>, value: "chart" },
            { label: <Tooltip title="表格"><TableOutlined /></Tooltip>, value: "table" },
          ]}
        />
      </div>
      {mode === "chart" ? (
        <DimensionVisual rows={rows} accent={accent} emptyText={emptyText} />
      ) : (
        <Table
          rowKey={(record) => record.key || record.name}
          columns={columns}
          dataSource={sortRows(rows)}
          pagination={false}
          scroll={{ x: 760 }}
          locale={{ emptyText }}
        />
      )}
    </section>
  );
}

export function GrainPurchaseDashboardPanel() {
  const [range, setRange] = useState<RangeValue>([today, today]);
  const [stationId, setStationId] = useState<number | undefined>();
  const [stationOptions, setStationOptions] = useState<CrudOption[]>([]);
  const [dashboard, setDashboard] = useState<GrainPurchaseDashboardRecord | null>(null);
  const [loading, setLoading] = useState(false);
  const [stationMode, setStationMode] = useState<DimensionMode>("chart");
  const [cropMode, setCropMode] = useState<DimensionMode>("chart");

  const loadDashboard = async () => {
    setLoading(true);
    try {
      const data = await getGrainPurchaseDashboard({ ...dateParams(range), stationId });
      setDashboard(data);
    } catch (error) {
      message.error(error instanceof Error ? error.message : "加载收粮大盘失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadDashboard();
  }, [range, stationId]);

  useEffect(() => {
    grainStationApi
      .list({ pageIndex: 1, pageSize: 200, status: "active" })
      .then((result) => setStationOptions(result.data.map((station) => ({ label: station.stationName, value: station.id }))))
      .catch((error) => message.error(error instanceof Error ? error.message : "加载粮站失败"));
  }, []);

  const overview = dashboard?.overview;
  const scopeLabel = useMemo(() => {
    const params = dateParams(range);
    const stationName = stationOptions.find((item) => item.value === stationId)?.label;
    return `${params.startDate} 至 ${params.endDate}${stationName ? ` · ${stationName}` : " · 全部粮站"}`;
  }, [range, stationId, stationOptions]);

  const metrics = [
    {
      label: "新增农户",
      value: `${formatNumber(overview?.newFarmerCount ?? 0)} 户`,
      hint: "当前时间段内有收粮记录的农户",
      icon: <UserAddOutlined />,
      tone: "green",
    },
    {
      label: "当前操作业务员",
      value: `${formatNumber(overview?.activeUserCount ?? 0)} 人`,
      hint: `${formatNumber(overview?.entryCount ?? 0)} 笔收粮流水`,
      icon: <TeamOutlined />,
      tone: "blue",
    },
    {
      label: "当前收粮总重量",
      value: formatWeight(overview?.totalQuantity ?? 0),
      hint: `均价 ${formatMoney(overview?.averageUnitPrice ?? 0)}/公斤`,
      icon: <GoldOutlined />,
      tone: "gold",
    },
    {
      label: "当前收粮总金额",
      value: formatMoney(overview?.totalAmount ?? 0),
      hint: `户均 ${formatMoney(overview?.averageFarmerDeal ?? 0)}`,
      icon: <WalletOutlined />,
      tone: "red",
    },
  ];

  return (
    <div className="grain-dashboard">
      <section className="grain-dashboard-hero">
        <div>
          <Tag icon={<DashboardOutlined />} bordered={false}>
            收粮管理
          </Tag>
          <Title>收粮大盘</Title>
          <Text>{scopeLabel}</Text>
        </div>
        <Space wrap className="grain-dashboard-toolbar">
          <RangePicker
            value={range}
            onChange={setRange}
            allowClear={false}
            format="YYYY-MM-DD"
            presets={[
              { label: "今日", value: [today, today] },
              { label: "本周", value: [today.startOf("week"), today.endOf("week")] },
              { label: "本月", value: [today.startOf("month"), today.endOf("month")] },
            ]}
            suffixIcon={<CalendarOutlined />}
          />
          <Select
            allowClear
            showSearch
            placeholder="全部粮站"
            value={stationId}
            options={stationOptions}
            onChange={setStationId}
            optionFilterProp="label"
            style={{ minWidth: 180 }}
          />
          <Button icon={<ReloadOutlined />} loading={loading} onClick={loadDashboard}>
            刷新
          </Button>
        </Space>
      </section>

      <section className="grain-dashboard-metrics">
        {metrics.map((metric) => (
          <div className={`grain-dashboard-metric grain-dashboard-metric--${metric.tone}`} key={metric.label}>
            <span className="grain-dashboard-metric__icon">{metric.icon}</span>
            <Text>{metric.label}</Text>
            <strong>{metric.value}</strong>
            <small>{metric.hint}</small>
          </div>
        ))}
      </section>

      <div className="grain-dashboard-grid">
        <DimensionSection
          title="按照粮站维度展示"
          subtitle="对比各粮站收粮规模、农户覆盖与业务员参与"
          icon={<BankOutlined />}
          rows={dashboard?.byStation ?? []}
          mode={stationMode}
          onModeChange={setStationMode}
          accent="#2f8f5b"
          emptyText="当前条件下暂无粮站汇总"
        />
        <DimensionSection
          title="按照粮食类型维度展示"
          subtitle="观察不同粮食类型的重量、金额与金额占比"
          icon={<DatabaseOutlined />}
          rows={dashboard?.byCrop ?? []}
          mode={cropMode}
          onModeChange={setCropMode}
          accent="#b98222"
          emptyText="当前条件下暂无粮食类型汇总"
        />
      </div>

      <div className="grain-dashboard-footer">
        <Space size={16} wrap>
          <Text>
            <FieldTimeOutlined /> 数据生成：{dashboard?.generated ? dashboard.generated.replace("T", " ").slice(0, 19) : "-"}
          </Text>
          <Text>
            <LineChartOutlined /> 数据来源：收粮汇总表
          </Text>
          <Text>
            <AppstoreOutlined /> 默认精确到日期
          </Text>
        </Space>
      </div>
    </div>
  );
}
