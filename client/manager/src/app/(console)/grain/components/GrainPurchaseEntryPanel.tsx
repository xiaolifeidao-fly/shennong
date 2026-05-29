"use client";

import { useEffect, useState } from "react";
import { StopOutlined } from "@ant-design/icons";
import { Button, Popconfirm, Tooltip, message } from "antd";
import { CrudManagementPanel, type CrudField, type CrudOption, type CrudTableColumn } from "../../components/CrudManagementPanel";
import { fetchAppUsers } from "../../app-user/api/app-user.api";
import {
  grainPurchaseEntryApi,
  grainStationApi,
  voidGrainPurchaseEntry,
  type GrainPayload,
  type GrainPurchaseEntryRecord,
} from "../api/grain.api";

const statusOptions: CrudOption[] = [
  { label: "已提交", value: "submitted" },
  { label: "已作废", value: "voided" },
];

const fields: CrudField<GrainPurchaseEntryRecord>[] = [
  { name: "farmerId", label: "粮户ID", type: "number", min: 0, precision: 0, required: true },
  { name: "purchaseTypeId", label: "收购类型ID", type: "number", min: 0, precision: 0 },
  { name: "crop", label: "收购类型", required: true },
  { name: "quantity", label: "重量", type: "number", min: 0, precision: 3, required: true },
  { name: "unit", label: "单位" },
  { name: "amount", label: "金额", type: "number", min: 0, precision: 2, required: true },
  { name: "placeId", label: "地点ID", type: "number", min: 0, precision: 0 },
  { name: "place", label: "收购地点" },
  { name: "locationAddress", label: "定位地址" },
  { name: "paymentMethodId", label: "付款方式ID", type: "number", min: 0, precision: 0 },
  { name: "payType", label: "付款方式" },
  { name: "status", label: "状态", type: "select", options: statusOptions },
  { name: "remark", label: "备注", type: "textarea" },
];

const columns: CrudTableColumn<GrainPurchaseEntryRecord>[] = [
  { name: "crop", label: "收购类型", width: 140 },
  { name: "farmerId", label: "粮户ID", width: 100 },
  { name: "quantity", label: "重量", width: 110 },
  { name: "unit", label: "单位", width: 80 },
  { name: "amount", label: "金额", width: 120 },
  { name: "unitPrice", label: "单价", width: 120 },
  { name: "place", label: "地点", width: 160 },
  { name: "payType", label: "付款方式", width: 140 },
  { name: "status", label: "状态", width: 110 },
];

export function GrainPurchaseEntryPanel() {
  const [stationOptions, setStationOptions] = useState<CrudOption[]>([]);
  const [appUserOptions, setAppUserOptions] = useState<CrudOption[]>([]);

  useEffect(() => {
    Promise.all([
      grainStationApi.list({ pageIndex: 1, pageSize: 200, status: "active" }),
      fetchAppUsers({ pageIndex: 1, pageSize: 200, status: "active" }),
    ])
      .then(([stations, appUsers]) => {
        setStationOptions(
          stations.data.map((station) => ({
            label: station.stationName,
            value: station.id,
          })),
        );
        setAppUserOptions(
          appUsers.data.map((user) => ({
            label: user.name || user.username || `业务员 ${user.id}`,
            value: user.id,
          })),
        );
      })
      .catch((error) => message.error(error instanceof Error ? error.message : "加载收粮明细选项失败"));
  }, []);

  const stationLabel = (stationId: unknown) =>
    stationOptions.find((option) => option.value === stationId)?.label ?? String(stationId || "-");

  const appUserLabel = (appUserId: unknown) =>
    appUserOptions.find((option) => option.value === appUserId)?.label ?? String(appUserId || "-");

  const entryFields: CrudField<GrainPurchaseEntryRecord>[] = [
    {
      name: "stationId",
      label: "粮站",
      type: "select",
      required: true,
      placeholder: "请选择粮站",
      options: stationOptions,
    },
    {
      name: "appUserId",
      label: "业务员",
      type: "select",
      placeholder: "请选择业务员",
      options: appUserOptions,
    },
    ...fields,
  ];

  const entryColumns: CrudTableColumn<GrainPurchaseEntryRecord>[] = [
    { name: "stationId", label: "粮站", width: 180, render: stationLabel },
    { name: "appUserId", label: "业务员", width: 160, render: appUserLabel },
    ...columns,
  ];

  return (
    <CrudManagementPanel<GrainPurchaseEntryRecord, GrainPayload>
      title="收粮明细"
      createText="新增收粮明细"
      searchPlaceholder="收购类型/地点/付款方式/地址"
      searchParam="search"
      fields={entryFields}
      columns={entryColumns}
      statusField="status"
      statusOptions={statusOptions}
      actionWidth={180}
      rowActions={(record, context) =>
        record.status === "voided" ? null : (
          <Popconfirm
            title={`确认作废收粮明细 #${record.id} 吗？`}
            okText="作废"
            cancelText="取消"
            onConfirm={async () => {
              try {
                context.setSubmitting(true);
                await voidGrainPurchaseEntry(record.id);
                message.success("作废成功");
                await context.reload();
              } catch (error) {
                message.error(error instanceof Error ? error.message : "作废失败");
              } finally {
                context.setSubmitting(false);
              }
            }}
          >
            <Tooltip title="作废">
              <Button type="text" danger icon={<StopOutlined />} disabled={context.submitting} />
            </Tooltip>
          </Popconfirm>
        )
      }
      api={grainPurchaseEntryApi}
    />
  );
}
