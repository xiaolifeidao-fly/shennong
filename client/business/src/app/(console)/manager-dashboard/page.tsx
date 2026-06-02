"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  CalendarOutlined,
  DatabaseOutlined,
  FileDoneOutlined,
  GoldOutlined,
  ReloadOutlined,
  TeamOutlined,
  WalletOutlined,
} from "@ant-design/icons";
import { Button, DatePicker, Empty, Progress, Space, Typography, message } from "antd";
import dayjs from "dayjs";
import type { Dayjs } from "dayjs";
import { useRouter } from "next/navigation";
import {
  getGrainPurchaseDashboard,
  type GrainPurchaseDashboardDimensionRecord,
  type GrainPurchaseDashboardRecord,
} from "../grain/api/grain.api";

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

export default function ManagerDashboardPage() {
  const router = useRouter();
  const [range, setRange] = useState<RangeValue>([today, today]);
  const [dashboard, setDashboard] = useState<GrainPurchaseDashboardRecord | null>(null);
  const [loading, setLoading] = useState(false);

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

  const overview = dashboard?.overview;
  const weight = formatWeight(overview?.totalQuantity ?? 0);
  const cropRows = useMemo(() => sortRows(dashboard?.byCrop ?? []).slice(0, 6), [dashboard]);
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
            数据来自服务端收粮汇总表，汇总业务员、农户、品类等今日收粮业务数据。
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
            <FileDoneOutlined className="manager-dashboard-panel__icon" />
            <div>
              <h3>品类占比</h3>
              <Text>按粮食品类展示今日收粮金额占比。</Text>
            </div>
          </div>
          {cropRows.length ? (
            <div className="manager-dashboard-category-list">
              {cropRows.map((item) => (
                <Category key={item.key || item.name} item={item} />
              ))}
            </div>
          ) : (
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前时间段暂无粮食品类汇总" />
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
