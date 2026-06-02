"use client";

import { useEffect, useState } from "react";
import { message } from "antd";
import {
  CrudManagementPanel,
  type CrudCascaderOption,
  type CrudField,
  type CrudOption,
  type CrudTableColumn,
} from "../../components/CrudManagementPanel";
import {
  grainStationApi,
  listRegionTree,
  type GrainPayload,
  type GrainStationRecord,
  type RegionTreeRecord,
} from "../api/grain.api";

const statusOptions: CrudOption[] = [
  { label: "启用", value: "active" },
  { label: "停用", value: "inactive" },
];

const columns: CrudTableColumn<GrainStationRecord>[] = [
  { name: "stationName", label: "粮站名称", width: 180 },
  { name: "tenantId", label: "租户ID", width: 100 },
  { name: "contactName", label: "联系人", width: 120 },
  { name: "contactPhone", label: "联系电话", width: 150 },
  {
    name: "address",
    label: "地址",
    width: 320,
    render: (_, record) =>
      [record.province, record.city, record.district, record.address].filter(Boolean).join(" ") || "-",
  },
  { name: "status", label: "状态", width: 100 },
];

function toCascaderOptions(regions: RegionTreeRecord[]): CrudCascaderOption[] {
  return regions.map((region) => ({
    label: region.name,
    value: region.name,
    children: region.children ? toCascaderOptions(region.children) : undefined,
  }));
}

export function GrainStationPanel() {
  const [regionOptions, setRegionOptions] = useState<CrudCascaderOption[]>([]);

  useEffect(() => {
    listRegionTree()
      .then((regions) => setRegionOptions(toCascaderOptions(regions)))
      .catch((error) => message.error(error instanceof Error ? error.message : "加载地区数据失败"));
  }, []);

  const fields: CrudField<GrainStationRecord>[] = [
    { name: "stationName", label: "粮站名称", required: true },
    { name: "contactName", label: "联系人" },
    { name: "contactPhone", label: "联系电话" },
    {
      name: "regionPath",
      label: "所在地区",
      type: "cascader",
      placeholder: "请选择省 / 市 / 区县",
      cascaderOptions: regionOptions,
      linkedNames: ["province", "city", "district"],
    },
    { name: "address", label: "详细地址" },
    { name: "status", label: "状态", type: "select", options: statusOptions },
    { name: "remark", label: "备注", type: "textarea" },
  ];

  return (
    <CrudManagementPanel<GrainStationRecord, GrainPayload>
      title="粮站"
      createText="新增粮站"
      searchPlaceholder="粮站名称/联系人"
      searchParam="search"
      fields={fields}
      columns={columns}
      statusField="status"
      statusOptions={statusOptions}
      api={grainStationApi}
    />
  );
}
