"use client";

import { useEffect, useMemo, useState } from "react";
import { ApartmentOutlined } from "@ant-design/icons";
import { Button, message, Modal, Select, Tag, Tooltip } from "antd";
import {
  CrudManagementPanel,
  type CrudField,
  type CrudOption,
  type CrudTableColumn,
} from "../../components/CrudManagementPanel";
import { grainStationApi, type GrainStationRecord } from "../../grain/api/grain.api";
import { tenantApi, type TenantPayload, type TenantRecord } from "../api/tenant.api";

const statusOptions: CrudOption[] = [
  { label: "启用", value: "active" },
  { label: "停用", value: "inactive" },
];

export function TenantManagementPanel() {
  const [stations, setStations] = useState<GrainStationRecord[]>([]);

  useEffect(() => {
    grainStationApi
      .list({ pageIndex: 1, pageSize: 200 })
      .then((result) => setStations(result.data))
      .catch((error) => message.error(error instanceof Error ? error.message : "加载粮站失败"));
  }, []);

  const stationOptions = useMemo(
    () => stations.map((station) => ({ label: station.stationName, value: station.id })),
    [stations],
  );

  const fields: CrudField<TenantRecord>[] = [
    { name: "tenantName", label: "租户名称", required: true },
    { name: "tenantCode", label: "租户编码", placeholder: "留空自动生成" },
    { name: "contactName", label: "联系人" },
    { name: "contactPhone", label: "联系电话" },
    { name: "status", label: "状态", type: "select", options: statusOptions },
    { name: "remark", label: "备注", type: "textarea" },
  ];

  const columns: CrudTableColumn<TenantRecord>[] = [
    { name: "tenantName", label: "租户名称", width: 180 },
    { name: "tenantCode", label: "租户编码", width: 150 },
    { name: "contactName", label: "联系人", width: 120 },
    { name: "contactPhone", label: "联系电话", width: 150 },
    {
      name: "stationNames",
      label: "关联粮站",
      width: 260,
      render: (value) => {
        const names = Array.isArray(value) ? value.filter((item): item is string => typeof item === "string") : [];
        if (names.length === 0) {
          return "-";
        }
        return names.map((name) => (
          <Tag key={name} style={{ marginBottom: 4 }}>
            {name}
          </Tag>
        ));
      },
    },
    { name: "status", label: "状态", width: 100 },
  ];

  const handleConfigStations = (record: TenantRecord) => {
    let nextStationIds: number[] = Array.isArray(record.stationIds) ? (record.stationIds as number[]) : [];
    Modal.confirm({
      title: `配置粮站 — ${record.tenantName}`,
      width: 520,
      content: (
        <Select<number[]>
          mode="multiple"
          defaultValue={nextStationIds}
          style={{ width: "100%", marginTop: 16 }}
          placeholder="请选择关联粮站（可多选，可选全部粮站）"
          options={stationOptions}
          optionFilterProp="label"
          maxTagCount="responsive"
          onChange={(value) => {
            nextStationIds = value;
          }}
        />
      ),
      onOk: async () => {
        await tenantApi.update(record.id, { stationIds: nextStationIds });
        message.success("粮站配置已更新");
      },
    });
  };

  return (
    <CrudManagementPanel<TenantRecord, TenantPayload>
      title="租户"
      createText="新增租户"
      searchPlaceholder="租户名称/编码/联系人"
      searchParam="search"
      fields={fields}
      columns={columns}
      statusField="status"
      statusOptions={statusOptions}
      actionWidth={160}
      rowActions={(record) => (
        <Tooltip title="配置粮站">
          <Button type="text" icon={<ApartmentOutlined />} onClick={() => handleConfigStations(record)} />
        </Tooltip>
      )}
      api={tenantApi}
    />
  );
}
