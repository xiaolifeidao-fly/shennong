"use client";

import { useEffect, useMemo, useState } from "react";
import { Col, Row, Table, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import {
  CrudManagementPanel,
  type CrudField,
  type CrudOption,
  type CrudTableColumn,
} from "../../components/CrudManagementPanel";
import {
  grainPurchaseTypeApi,
  type GrainPayload,
  type GrainPurchaseTypeRecord,
  type GrainStationRecord,
} from "../api/grain.api";
import { useAccessibleStations } from "../hooks/useAccessibleStations";

const { Text } = Typography;

const statusOptions: CrudOption[] = [
  { label: "启用", value: "active" },
  { label: "停用", value: "inactive" },
];

const purchaseTypeFields: CrudField<GrainPurchaseTypeRecord>[] = [
  { name: "typeName", label: "收粮类型", required: true },
  { name: "unit", label: "默认单位" },
  { name: "sortOrder", label: "排序", type: "number", min: 0, precision: 0 },
  { name: "status", label: "状态", type: "select", options: statusOptions },
];

const typeColumns: CrudTableColumn<GrainPurchaseTypeRecord>[] = [
  { name: "typeName", label: "收粮类型", width: 160 },
  { name: "unit", label: "单位", width: 90 },
  { name: "sortOrder", label: "排序", width: 90 },
  { name: "status", label: "状态", width: 100 },
];

const stationColumns: ColumnsType<GrainStationRecord> = [
  {
    title: "粮站名称",
    dataIndex: "stationName",
    ellipsis: true,
  },
  {
    title: "编码",
    dataIndex: "stationCode",
    width: 100,
    ellipsis: true,
  },
];

export function GrainConfigPanel() {
  const { stations, loading: stationsLoading } = useAccessibleStations();
  const [selectedStationId, setSelectedStationId] = useState<number | null>(null);
  const [stationPage, setStationPage] = useState(1);
  const stationPageSize = 10;

  useEffect(() => {
    if (stations.length === 0) {
      setSelectedStationId(null);
      return;
    }
    if (selectedStationId == null || !stations.some((station) => station.id === selectedStationId)) {
      setSelectedStationId(stations[0].id);
    }
  }, [selectedStationId, stations]);

  const selectedStation = stations.find((s) => s.id === selectedStationId);

  // Wrap api to inject stationId filter
  const typesApi = useMemo(
    () => ({
      list: (query: Parameters<typeof grainPurchaseTypeApi.list>[0]) =>
        grainPurchaseTypeApi.list({ ...query, stationId: selectedStationId ?? undefined }),
      create: (payload: GrainPayload) =>
        grainPurchaseTypeApi.create({ ...payload, stationId: selectedStationId }),
      update: (id: number, payload: Partial<GrainPayload>) =>
        grainPurchaseTypeApi.update(id, payload),
      remove: (id: number) => grainPurchaseTypeApi.remove(id),
    }),
    [selectedStationId],
  );

  return (
    <Row gutter={16} style={{ height: "100%" }}>
      {/* 左侧：粮站列表 */}
      <Col flex="320px" style={{ borderRight: "1px solid #f0f0f0", paddingRight: 16 }}>
        <Typography.Title level={5} style={{ marginBottom: 12 }}>
          粮站列表
        </Typography.Title>
        <Table<GrainStationRecord>
          rowKey="id"
          size="small"
          loading={stationsLoading}
          columns={stationColumns}
          dataSource={stations}
          rowSelection={undefined}
          onRow={(record) => ({
            onClick: () => {
              setSelectedStationId(record.id);
            },
            style: {
              cursor: "pointer",
              background: record.id === selectedStationId ? "#e6f4ff" : undefined,
            },
          })}
          pagination={{
            current: stationPage,
            pageSize: stationPageSize,
            total: stations.length,
            onChange: (page) => setStationPage(page),
            showSizeChanger: false,
            size: "small",
          }}
          scroll={{ y: "calc(100vh - 280px)" }}
        />
      </Col>

      {/* 右侧：收粮类型 */}
      <Col flex="1" style={{ minWidth: 0 }}>
        {selectedStationId == null ? (
          <div
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              height: 300,
              color: "#999",
            }}
          >
            <Text type="secondary">请点击左侧粮站查看收粮类型</Text>
          </div>
        ) : (
          <CrudManagementPanel<GrainPurchaseTypeRecord, GrainPayload>
            key={selectedStationId}
            title={`收粮类型 — ${selectedStation?.stationName ?? ""}`}
            createText="新增收粮类型"
            searchPlaceholder="收粮类型"
            searchParam="search"
            fields={purchaseTypeFields}
            columns={typeColumns}
            statusField="status"
            statusOptions={statusOptions}
            api={typesApi}
          />
        )}
      </Col>
    </Row>
  );
}
