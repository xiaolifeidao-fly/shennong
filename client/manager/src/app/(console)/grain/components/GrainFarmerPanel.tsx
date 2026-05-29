"use client";

import { useEffect, useState } from "react";
import { message, Tabs } from "antd";
import {
  CrudManagementPanel,
  type CrudCascaderOption,
  type CrudField,
  type CrudOption,
  type CrudTableColumn,
} from "../../components/CrudManagementPanel";
import { fetchAppUsers } from "../../app-user/api/app-user.api";
import {
  grainFarmerApi,
  grainPurchasePlaceApi,
  grainStationApi,
  listRegionTree,
  type GrainFarmerRecord,
  type GrainPayload,
  type GrainPurchasePlaceRecord,
  type RegionTreeRecord,
} from "../api/grain.api";

const statusOptions: CrudOption[] = [
  { label: "资料完整", value: "complete" },
  { label: "银行卡待补", value: "missing-bank" },
  { label: "停用", value: "inactive" },
];

const fields: CrudField<GrainFarmerRecord>[] = [
  { name: "name", label: "农户姓名", required: true },
  { name: "idNumber", label: "身份证号" },
  { name: "phone", label: "手机号" },
  { name: "address", label: "地址" },
  { name: "bankNumber", label: "银行卡号" },
  { name: "bankName", label: "开户行" },
  { name: "status", label: "状态", type: "select", options: statusOptions },
  { name: "statusText", label: "状态说明" },
  { name: "remark", label: "备注", type: "textarea" },
];

const columns: CrudTableColumn<GrainFarmerRecord>[] = [
  { name: "name", label: "农户姓名", width: 140 },
  { name: "phone", label: "手机号", width: 140 },
  { name: "idNumber", label: "身份证号", width: 190 },
  { name: "bankName", label: "开户行", width: 180 },
  { name: "status", label: "状态", width: 120 },
];

const placeStatusOptions: CrudOption[] = [
  { label: "启用", value: "active" },
  { label: "停用", value: "inactive" },
];

function toCascaderOptions(regions: RegionTreeRecord[]): CrudCascaderOption[] {
  return regions.map((region) => ({
    label: region.name,
    value: region.name,
    children: region.children ? toCascaderOptions(region.children) : undefined,
  }));
}

function GrainPurchasePlacePanel() {
  const [regionOptions, setRegionOptions] = useState<CrudCascaderOption[]>([]);
  const [appUserOptions, setAppUserOptions] = useState<CrudOption[]>([]);

  useEffect(() => {
    Promise.all([listRegionTree(), fetchAppUsers({ pageIndex: 1, pageSize: 200, status: "active" })])
      .then(([regions, appUsers]) => {
        setRegionOptions(toCascaderOptions(regions));
        setAppUserOptions(
          appUsers.data.map((user) => ({
            label: user.name || user.username || `业务员 ${user.id}`,
            value: user.id,
          })),
        );
      })
      .catch((error) => message.error(error instanceof Error ? error.message : "加载收购地点选项失败"));
  }, []);

  const appUserLabel = (appUserId: unknown) =>
    appUserOptions.find((option) => option.value === appUserId)?.label ?? String(appUserId || "-");

  const placeFields: CrudField<GrainPurchasePlaceRecord>[] = [
    {
      name: "appUserId",
      label: "业务员",
      type: "select",
      required: true,
      placeholder: "请选择业务员",
      options: appUserOptions,
    },
    { name: "placeName", label: "地点名称", required: true },
    { name: "placeType", label: "地点类型" },
    {
      name: "regionPath",
      label: "所在地区",
      type: "cascader",
      placeholder: "请选择省 / 市 / 区县",
      cascaderOptions: regionOptions,
      linkedNames: ["province", "city", "district"],
    },
    { name: "address", label: "地址" },
    { name: "sortOrder", label: "排序", type: "number", min: 0, precision: 0 },
    { name: "status", label: "状态", type: "select", options: placeStatusOptions },
  ];

  const placeColumns: CrudTableColumn<GrainPurchasePlaceRecord>[] = [
    { name: "appUserId", label: "业务员", width: 160, render: appUserLabel },
    { name: "placeName", label: "地点名称", width: 160 },
    { name: "placeType", label: "类型", width: 110 },
    {
      name: "address",
      label: "地址",
      width: 320,
      render: (_, record) =>
        [record.province, record.city, record.district, record.address].filter(Boolean).join(" ") || "-",
    },
    { name: "sortOrder", label: "排序", width: 90 },
    { name: "status", label: "状态", width: 100 },
  ];

  return (
    <CrudManagementPanel<GrainPurchasePlaceRecord, GrainPayload>
      title="收购地点"
      createText="新增收购地点"
      searchPlaceholder="地点名称/地址"
      searchParam="search"
      fields={placeFields}
      columns={placeColumns}
      statusField="status"
      statusOptions={placeStatusOptions}
      api={grainPurchasePlaceApi}
    />
  );
}

export function GrainFarmerPanel() {
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
      .catch((error) => message.error(error instanceof Error ? error.message : "加载农户选项失败"));
  }, []);

  const stationLabel = (stationId: unknown) =>
    stationOptions.find((option) => option.value === stationId)?.label ?? String(stationId || "-");

  const appUserLabel = (appUserId: unknown) =>
    appUserOptions.find((option) => option.value === appUserId)?.label ?? String(appUserId || "-");

  const farmerFields: CrudField<GrainFarmerRecord>[] = [
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

  const farmerColumns: CrudTableColumn<GrainFarmerRecord>[] = [
    { name: "stationId", label: "粮站", width: 180, render: stationLabel },
    { name: "appUserId", label: "业务员", width: 160, render: appUserLabel },
    ...columns,
  ];

  return (
    <Tabs
      items={[
        {
          key: "farmers",
          label: "粮户列表",
          children: (
            <CrudManagementPanel<GrainFarmerRecord, GrainPayload>
              title="粮户"
              createText="新增粮户"
              searchPlaceholder="姓名/手机号/身份证/地址"
              searchParam="search"
              fields={farmerFields}
              columns={farmerColumns}
              statusField="status"
              statusOptions={statusOptions}
              api={grainFarmerApi}
            />
          ),
        },
        {
          key: "places",
          label: "收购地点",
          children: <GrainPurchasePlacePanel />,
        },
      ]}
    />
  );
}
