"use client";

import { getPage, instance, unwrapApiResponse, type ApiResponse } from "@/utils/axios";
import type { CrudListQuery } from "../../components/CrudManagementPanel";

export class TenantRecord {
  id!: number;
  tenantName = "";
  tenantCode = "";
  contactName = "";
  contactPhone = "";
  status = "active";
  remark = "";
  stationIds: number[] = [];
  stationNames: string[] = [];
  createdTime?: string;
  updatedTime?: string;
  [key: string]: unknown;
}

export type TenantPayload = Record<string, unknown>;

export const tenantApi = {
  list: (query: CrudListQuery) => getPage(TenantRecord, "/tenants", query),
  create: async (payload: TenantPayload) => {
    const response = await instance.post<ApiResponse<TenantRecord>>("/tenants", payload);
    return unwrapApiResponse(response.data);
  },
  update: async (id: number, payload: Partial<TenantPayload>) => {
    const response = await instance.put<ApiResponse<TenantRecord>>(`/tenants/${id}`, payload);
    return unwrapApiResponse(response.data);
  },
  remove: async (id: number) => {
    const response = await instance.delete<ApiResponse<{ deleted: boolean }>>(`/tenants/${id}`);
    return unwrapApiResponse(response.data);
  },
};
