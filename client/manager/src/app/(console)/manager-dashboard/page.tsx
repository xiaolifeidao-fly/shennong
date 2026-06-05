"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  BankOutlined,
  CalendarOutlined,
  DatabaseOutlined,
  FileDoneOutlined,
  GoldOutlined,
  ReloadOutlined,
  TeamOutlined,
  WalletOutlined,
} from "@ant-design/icons";
import { Button, DatePicker, Empty, Progress, Select, Space, Table, Typography, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import dayjs from "dayjs";
import type { Dayjs } from "dayjs";
import { useRouter } from "next/navigation";
import {
  getGrainPurchaseDashboard,
  type GrainPurchaseDashboardDimensionRecord,
  type GrainPurchaseDashboardRecord,
} from "../grain/api/grain.api";
import { useAccessibleStations } from "../grain/hooks/useAccessibleStations";

const { RangePicker } = DatePicker;
const { Text, Title } = Typography;
const today = dayjs();
type RangeValue = [Dayjs | null, Dayjs | null] | null;

function formatMoney(value: number) {
  return `¥${new Intl.NumberFormat("zh-CN", { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(value || 0)}`;
}

function formatNumber(value: number, digits = 0) {
  return new Intl.NumberFormat("zh-CN", { maximumFractionDigits: digits }).format(value || 0);
}

function formatWeight(value: number) {
  if (value >= 1000) {
    return {
      value: formatNumber(value / 1000, 2),
      unit: "吨",
      text: `${formatNumber(value / 1000, 2)} 吨`,
    };
  }
  return {
    value: formatNumber(value, 3),
    unit: "公斤",
    text: `${formatNumber(value, 3)} 公斤`,
  };
}

function sharePercent(value: number) {
  return Math.min(100, Math.max(0, Number(((value || 0) * 100).toFixed(1))));
}

function sortRows(rows: GrainPurchaseDashboardDimensionRecord[]) {
  return [...rows].sort((a, b) => b.totalAmount - a.totalAmount || b.totalQuantity - a.totalQuantity);
}

function dateParams(range: RangeValue) {
  const start = range?.[0] ?? today;
  const end = range?.[1] ?? start;
  return {
    startDate: start.format("YYYY-MM-DD"),
    endDate: end.format("YYYY-MM-DD"),
  };
}

function rangeLabel(range: RangeValue) {
  const params = dateParams(range);
  if (params.startDate === params.endDate) {
    return params.startDate === today.format("YYYY-MM-DD") ? "今日" : params.startDate;
  }
  return `${params.startDate} 至 ${params.endDate}`;
}

const columns: ColumnsType<GrainPurchaseDashboardDimensionRecord> = [
  {
    title: "粮站",
    dataIndex: "name",
    width: 180,
    render: (value: string, record) => (
      <div className="manager-dashboard-table-title">
        {value || "未命名粮站"}
        <Text className="manager-dashboard-table-subtitle">收粮 {record.entryCount} 笔</Text>
      </div>
    ),
  },
  {
    title: "业务员",
    dataIndex: "activeUserCount",
    width: 110,
    render: (value: number) => `${formatNumber(value)} 人`,
  },
  {
    title: "农户数",
    dataIndex: "farmerCount",
    width: 110,
    render: (value: number) => `${formatNumber(value)} 户`,
  },
  {
    title: "收粮量",
    dataIndex: "totalQuantity",
    width: 130,
    render: (value: number) => formatWeight(value).text,
  },
  {
    title: "金额",
    dataIndex: "totalAmount",
    width: 140,
    render: (value: number) => <span className="manager-dashboard-money">{formatMoney(value)}</span>,
  },
  {
    title: "均价",
    dataIndex: "averageUnitPrice",
    width: 130,
    render: (value: number) => `${formatMoney(value)}/公斤`,
  },
  {
    title: "金额占比",
    dataIndex: "amountShare",
    width: 150,
    render: (value: number) => <Progress percent={sharePercent(value)} size="small" strokeColor="#237a4b" />,
  },
];

export default function ManagerDashboardPage() {
  const router = useRouter();
  const { stations: accessibleStations, loading: stationsLoading } = useAccessibleStations();
  const stationOptions = useMemo(
    () => accessibleStations.map((station) => ({ label: station.stationName, value: station.id })),
    [accessibleStations],
  );
  const [range, setRange] = useState<RangeValue>([today, today]);
  const [dashboard, setDashboard] = useState<GrainPurchaseDashboardRecord | null>(null);
  const [cropDashboard, setCropDashboard] = useState<GrainPurchaseDashboardRecord | null>(null);
  const [cropStationId, setCropStationId] = useState<number | undefined>();
  const [loading, setLoading] = useState(false);
  const [cropLoading, setCropLoading] = useState(false);

  const loadDashboard = useCallback(async () => {
    setLoading(true);
    try {
      setDashboard(await getGrainPurchaseDashboard(dateParams(range)));
    } catch (error) {
      message.error(error instanceof Error ? error.message : "加载数据总览失败");
    } finally {
      setLoading(false);
    }
  }, [range]);

  useEffect(() => {
    void loadDashboard();
  }, [loadDashboard]);

  useEffect(() => {
    if (stationsLoading) {
      return;
    }
    if (accessibleStations.length === 0) {
      setCropStationId(undefined);
      return;
    }
    setCropStationId((current) =>
      current && accessibleStations.some((station) => station.id === current) ? current : accessibleStations[0].id,
    );
  }, [accessibleStations, stationsLoading]);

  const loadCropDashboard = useCallback(async () => {
    if (stationsLoading || (accessibleStations.length > 0 && !cropStationId)) {
      return;
    }
    setCropLoading(true);
    try {
      setCropDashboard(await getGrainPurchaseDashboard({ ...dateParams(range), stationId: cropStationId }));
    } catch (error) {
      message.error(error instanceof Error ? error.message : "加载粮食类型统计失败");
    } finally {
      setCropLoading(false);
    }
  }, [accessibleStations.length, cropStationId, range, stationsLoading]);

  useEffect(() => {
    void loadCropDashboard();
  }, [loadCropDashboard]);

  const overview = dashboard?.overview;
  const weight = formatWeight(overview?.totalQuantity ?? 0);
  const stationRows = useMemo(() => sortRows(dashboard?.byStation ?? []), [dashboard]);
  const cropRows = useMemo(() => sortRows(cropDashboard?.byCrop ?? []).slice(0, 6), [cropDashboard]);
  const currentRangeLabel = rangeLabel(range);

  const stats = [
    {
      label: `${currentRangeLabel}收粮量`,
      value: weight.value,
      unit: weight.unit,
      hint: `${formatNumber(overview?.entryCount ?? 0)} 笔收粮流水`,
      icon: <GoldOutlined className="manager-dashboard-metric__icon" />,
    },
    {
      label: `${currentRangeLabel}金额`,
      value: formatMoney(overview?.totalAmount ?? 0),
      unit: "",
      hint: `均价 ${formatMoney(overview?.averageUnitPrice ?? 0)}/公斤`,
      icon: <WalletOutlined className="manager-dashboard-metric__icon" />,
    },
    {
      label: "收粮农户",
      value: formatNumber(overview?.newFarmerCount ?? 0),
      unit: "户",
      hint: `户均 ${formatMoney(overview?.averageFarmerDeal ?? 0)}`,
      icon: <TeamOutlined className="manager-dashboard-metric__icon" />,
    },
  ];

  return (
    <div className="manager-page-stack manager-dashboard">
      <section className="manager-dashboard-hero">
        <div>
          <Text className="manager-section-label">{currentRangeLabel}全局概览</Text>
          <Title level={1} className="manager-dashboard-hero__title">
            {currentRangeLabel}已录入 {formatNumber(overview?.entryCount ?? 0)} 笔，覆盖 {formatNumber(overview?.newFarmerCount ?? 0)} 户农户，
            合计收粮 {weight.text}。
          </Title>
          <Text className="manager-dashboard-hero__subtitle">
            数据来自服务端收粮汇总表，按粮站、业务员、农户、品类汇总今日收粮业务数据。
          </Text>
        </div>
        <Space wrap className="manager-dashboard-hero__actions">
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
          <Button type="primary" icon={<DatabaseOutlined />} onClick={() => router.push("/grain/entries")}>
            查看收粮记录
          </Button>
          <Button icon={<ReloadOutlined />} loading={loading} onClick={loadDashboard}>
            刷新
          </Button>
        </Space>
      </section>

      <section className="manager-dashboard-metric-grid">
        {stats.map((item) => (
          <div key={item.label} className="manager-dashboard-metric">
            <div className="manager-dashboard-metric__topline">
              <span>{item.label}</span>
              {item.icon}
            </div>
            <div className="manager-dashboard-metric__value">
              {item.value}
              {item.unit ? <span>{item.unit}</span> : null}
            </div>
            <Text className="manager-dashboard-metric__hint">{item.hint}</Text>
          </div>
        ))}
      </section>

      <section className="manager-dashboard-main-grid">
        <div className="manager-dashboard-panel manager-dashboard-panel--wide">
          <div className="manager-dashboard-panel__header">
            <BankOutlined className="manager-dashboard-panel__icon" />
            <div>
              <h3>{currentRangeLabel}粮站表现</h3>
              <Text>按粮站汇总收粮规模、农户覆盖、业务员参与和金额占比。</Text>
            </div>
          </div>
          <div className="manager-table">
            <Table<GrainPurchaseDashboardDimensionRecord>
              rowKey={(record) => record.key || String(record.stationId)}
              columns={columns}
              dataSource={stationRows}
              loading={loading}
              pagination={false}
              scroll={{ x: 920 }}
              locale={{ emptyText: "当前时间段暂无粮站汇总" }}
            />
          </div>
        </div>

        <div className="manager-dashboard-panel manager-dashboard-panel--wide">
          <div className="manager-dashboard-panel__header">
            <FileDoneOutlined className="manager-dashboard-panel__icon" />
            <div>
              <h3>品类占比</h3>
              <Text>按粮食品类展示今日收粮金额占比。</Text>
            </div>
            {stationOptions.length > 1 ? (
              <Select
                showSearch
                value={cropStationId}
                options={stationOptions}
                onChange={setCropStationId}
                optionFilterProp="label"
                className="manager-dashboard-panel__select"
              />
            ) : null}
          </div>
          {cropRows.length ? (
            <div className="manager-dashboard-category-list">
              {cropRows.map((item) => (
                <Category key={item.key || item.name} item={item} />
              ))}
            </div>
          ) : (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description={cropLoading ? "正在加载粮食类型统计" : "当前时间段暂无粮食品类汇总"}
            />
          )}
        </div>
      </section>
    </div>
  );
}

function Category({ item }: { item: GrainPurchaseDashboardDimensionRecord }) {
  const percent = sharePercent(item.amountShare);
  return (
    <div className="manager-dashboard-category-row">
      <div style={{ display: "flex", justifyContent: "space-between", gap: 12 }}>
        <span className="manager-dashboard-category-name">{item.name || "未命名粮食"}</span>
        <span>
          {percent}% · {formatWeight(item.totalQuantity).text}
        </span>
      </div>
      <Progress percent={percent} showInfo={false} strokeColor="#237a4b" trailColor="#e1e7dc" />
    </div>
  );
}
