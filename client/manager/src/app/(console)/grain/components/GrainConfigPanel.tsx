"use client";

import { useEffect, useState } from "react";
import { message, Tabs } from "antd";
import {
  CrudManagementPanel,
  type CrudField,
  type CrudOption,
  type CrudTableColumn,
} from "../../components/CrudManagementPanel";
import {
  grainPurchaseTypeApi,
  grainStationApi,
  type GrainPayload,
  type GrainPurchaseTypeRecord,
} from "../api/grain.api";

const statusOptions: CrudOption[] = [
  { label: "启用", value: "active" },
  { label: "停用", value: "inactive" },
];

const purchaseTypeFields: CrudField<GrainPurchaseTypeRecord>[] = [
  { name: "typeName", label: "收购类型", required: true },
  { name: "unit", label: "默认单位" },
  { name: "sortOrder", label: "排序", type: "number", min: 0, precision: 0 },
  { name: "status", label: "状态", type: "select", options: statusOptions },
];

export function GrainConfigPanel() {
  const [stationOptions, setStationOptions] = useState<CrudOption[]>([]);

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
      .catch((error) => message.error(error instanceof Error ? error.message : "加载基础选项失败"));
  }, []);

  const stationLabel = (stationId: unknown) =>
    stationOptions.find((option) => option.value === stationId)?.label ?? String(stationId || "-");

  const typeFields: CrudField<GrainPurchaseTypeRecord>[] = [
    {
      name: "stationId",
      label: "粮站",
      type: "select",
      required: true,
      placeholder: "请选择粮站",
      options: stationOptions,
    },
    ...purchaseTypeFields,
  ];

  const typeColumns: CrudTableColumn<GrainPurchaseTypeRecord>[] = [
    { name: "stationId", label: "粮站", width: 180, render: stationLabel },
    { name: "typeName", label: "收购类型", width: 160 },
    { name: "unit", label: "单位", width: 90 },
    { name: "sortOrder", label: "排序", width: 90 },
    { name: "status", label: "状态", width: 100 },
  ];

  return (
    <Tabs
      items={[
        {
          key: "types",
          label: "收购类型",
          children: (
            <CrudManagementPanel<GrainPurchaseTypeRecord, GrainPayload>
              title="收购类型"
              createText="新增收购类型"
              searchPlaceholder="收购类型"
              searchParam="search"
              fields={typeFields}
              columns={typeColumns}
              statusField="status"
              statusOptions={statusOptions}
              api={grainPurchaseTypeApi}
            />
          ),
        },
      ]}
    />
  );
}
