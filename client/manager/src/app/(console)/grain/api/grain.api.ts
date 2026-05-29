"use client";

import { getPage, instance, unwrapApiResponse, type ApiResponse } from "@/utils/axios";
import type { CrudListQuery } from "../../components/CrudManagementPanel";

export class GrainStationRecord {
  id!: number;
  stationName = "";
  stationCode = "";
  contactName = "";
  contactPhone = "";
  province = "";
  city = "";
  district = "";
  address = "";
  status = "active";
  remark = "";
  createdTime?: string;
  updatedTime?: string;
  [key: string]: unknown;
}

export class GrainFarmerRecord {
  id!: number;
  stationId = 0;
  appUserId = 0;
  name = "";
  idNumber = "";
  phone = "";
  address = "";
  bankNumber = "";
  bankName = "";
  status = "complete";
  statusText = "";
  remark = "";
  createdTime?: string;
  updatedTime?: string;
  [key: string]: unknown;
}

export class GrainPurchaseEntryRecord {
  id!: number;
  stationId = 0;
  appUserId = 0;
  farmerId = 0;
  purchaseTypeId = 0;
  crop = "";
  quantity = 0;
  unit = "公斤";
  amount = 0;
  unitPrice = 0;
  buyTime?: string;
  placeId = 0;
  place = "";
  locationAddress = "";
  paymentMethodId = 0;
  payType = "";
  status = "submitted";
  version = 1;
  remark = "";
  createdTime?: string;
  updatedTime?: string;
  [key: string]: unknown;
}

export class GrainPurchaseTypeRecord {
  id!: number;
  stationId = 0;
  typeName = "";
  unit = "公斤";
  sortOrder = 0;
  status = "active";
  createdTime?: string;
  updatedTime?: string;
  [key: string]: unknown;
}

export class GrainPaymentMethodRecord {
  id!: number;
  methodCode = "";
  methodName = "";
  sortOrder = 0;
  status = "active";
  createdTime?: string;
  updatedTime?: string;
  [key: string]: unknown;
}

export class GrainPurchasePlaceRecord {
  id!: number;
  appUserId = 0;
  placeName = "";
  placeType = "";
  province = "";
  city = "";
  district = "";
  address = "";
  sortOrder = 0;
  status = "active";
  createdTime?: string;
  updatedTime?: string;
  [key: string]: unknown;
}

export type GrainPayload = Record<string, unknown>;

export interface RegionTreeRecord {
  code: string;
  name: string;
  level: number;
  children?: RegionTreeRecord[];
}

function crudApi<R>(cls: new () => R, url: string) {
  return {
    list: (query: CrudListQuery) => getPage(cls, url, query),
    create: async (payload: GrainPayload) => {
      const response = await instance.post<ApiResponse<R>>(url, payload);
      return unwrapApiResponse(response.data);
    },
    update: async (id: number, payload: Partial<GrainPayload>) => {
      const response = await instance.put<ApiResponse<R>>(`${url}/${id}`, payload);
      return unwrapApiResponse(response.data);
    },
    remove: async (id: number) => {
      const response = await instance.delete<ApiResponse<{ deleted: boolean }>>(`${url}/${id}`);
      return unwrapApiResponse(response.data);
    },
  };
}

export const grainStationApi = crudApi(GrainStationRecord, "/grain-stations");
export const grainFarmerApi = crudApi(GrainFarmerRecord, "/grain-farmers");
export const grainPurchaseEntryApi = crudApi(GrainPurchaseEntryRecord, "/grain-purchase-entries");
export const grainPurchaseTypeApi = crudApi(GrainPurchaseTypeRecord, "/grain-purchase-types");
export const grainPaymentMethodApi = crudApi(GrainPaymentMethodRecord, "/grain-payment-methods");
export const grainPurchasePlaceApi = crudApi(GrainPurchasePlaceRecord, "/grain-purchase-places");

export async function voidGrainPurchaseEntry(id: number) {
  const response = await instance.put<ApiResponse<{ voided: boolean }>>(`/grain-purchase-entries/${id}/void`);
  return unwrapApiResponse(response.data);
}

export async function listRegionTree() {
  const response = await instance.get<ApiResponse<RegionTreeRecord[]>>("/regions/tree");
  return unwrapApiResponse(response.data);
}
