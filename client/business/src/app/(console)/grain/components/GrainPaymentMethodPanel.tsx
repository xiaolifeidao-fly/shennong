"use client";

import {
  CrudManagementPanel,
  type CrudField,
  type CrudOption,
  type CrudTableColumn,
} from "../../components/CrudManagementPanel";
import {
  grainPaymentMethodApi,
  type GrainPaymentMethodRecord,
  type GrainPayload,
} from "../api/grain.api";

const statusOptions: CrudOption[] = [
  { label: "启用", value: "active" },
  { label: "停用", value: "inactive" },
];

const paymentFields: CrudField<GrainPaymentMethodRecord>[] = [
  { name: "methodCode", label: "付款编码", required: true, placeholder: "例如 BANK_CARD" },
  { name: "methodName", label: "付款方式", required: true },
  { name: "sortOrder", label: "排序", type: "number", min: 0, precision: 0 },
  { name: "status", label: "状态", type: "select", options: statusOptions },
];

const paymentColumns: CrudTableColumn<GrainPaymentMethodRecord>[] = [
  { name: "methodCode", label: "付款编码", width: 160, copyable: true },
  { name: "methodName", label: "付款方式", width: 160 },
  { name: "sortOrder", label: "排序", width: 90 },
  { name: "status", label: "状态", width: 100 },
];

export function GrainPaymentMethodPanel() {
  return (
    <CrudManagementPanel<GrainPaymentMethodRecord, GrainPayload>
      title="付款方式"
      createText="新增付款方式"
      searchPlaceholder="付款编码/付款方式"
      searchParam="search"
      fields={paymentFields}
      columns={paymentColumns}
      statusField="status"
      statusOptions={statusOptions}
      api={grainPaymentMethodApi}
    />
  );
}
