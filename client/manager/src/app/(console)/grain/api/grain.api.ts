"use client";

import { getData, getPage, instance, unwrapApiResponse, type ApiResponse } from "@/utils/axios";
import { getAuthToken } from "@/utils/auth";
import type { CrudListQuery } from "../../components/CrudManagementPanel";

export class GrainStationRecord {
  id!: number;
  tenantId = 0;
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
  stationName = "";
  appUserId = 0;
  farmerId = 0;
  farmerName = "";
  farmerIdNumber = "";
  farmerPhone = "";
  farmerAddress = "";
  farmerBankNumber = "";
  farmerBankName = "";
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

export class GrainPurchaseEntrySnapshotRecord {
  id!: number;
  entryId = 0;
  snapshotVersion = 0;
  snapshotAction = "";
  snapshotTime?: string;
  operatorAppUserId = 0;
  operatorName = "";
  stationId = 0;
  appUserId = 0;
  farmerId = 0;
  farmerName = "";
  farmerIdNumber = "";
  farmerPhone = "";
  farmerAddress = "";
  farmerBankNumber = "";
  farmerBankName = "";
  purchaseTypeId = 0;
  crop = "";
  quantity = 0;
  unit = "公斤";
  amount = 0;
  unitPrice = 0;
  buyTime?: string;
  placeId = 0;
  place = "";
  locationName = "";
  locationAddress = "";
  longitude = "";
  latitude = "";
  province = "";
  city = "";
  district = "";
  paymentMethodId = 0;
  payType = "";
  entryStatus = "";
  entryRemark = "";
  createdTime?: string;
  updatedTime?: string;
  [key: string]: unknown;
}

export class GrainFarmerImagesRecord {
  idCardFront = "";
  idCardBack = "";
  bankCard = "";
}

export class GrainEntryMaterialRecord {
  id!: number;
  stationId = 0;
  entryId = 0;
  farmerId = 0;
  appUserId = 0;
  materialBizType = "";
  materialType = "";
  ossBucket = "";
  ossObjectKey = "";
  ossUrl = "";
  imageUrl = "";
  fileName = "";
  imageHash = "";
  fileSize = 0;
  mimeType = "";
  etag = "";
  sortOrder = 0;
  createdTime?: string;
  updatedTime?: string;
  [key: string]: unknown;
}

export class GrainPurchaseDashboardMetricRecord {
  newFarmerCount = 0;
  activeUserCount = 0;
  totalQuantity = 0;
  totalAmount = 0;
  entryCount = 0;
  averageUnitPrice = 0;
  averageFarmerDeal = 0;
}

export class GrainPurchaseDashboardDimensionRecord {
  key = "";
  name = "";
  stationId = 0;
  purchaseTypeId = 0;
  entryCount = 0;
  farmerCount = 0;
  activeUserCount = 0;
  totalQuantity = 0;
  totalAmount = 0;
  averageUnitPrice = 0;
  amountShare = 0;
  quantityShare = 0;
}

export class GrainPurchaseDashboardRecord {
  startDate = "";
  endDate = "";
  stationId = 0;
  overview = new GrainPurchaseDashboardMetricRecord();
  byStation: GrainPurchaseDashboardDimensionRecord[] = [];
  byCrop: GrainPurchaseDashboardDimensionRecord[] = [];
  generated?: string;
}

export class GrainCardOcrResult {
  cardType = "";
  mock = false;
  ossBucket = "";
  ossObjectKey = "";
  ossUrl = "";
  fileName = "";
  fileSize = 0;
  mimeType = "";
  name = "";
  idNumber = "";
  address = "";
  bankNumber = "";
  bankName = "";
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
export const grainPurchaseEntrySnapshotApi = {
  list: (query: CrudListQuery) => getPage(GrainPurchaseEntrySnapshotRecord, "/grain-entry-snapshots", query),
};
export const grainEntryMaterialApi = {
  list: async (query: CrudListQuery) => {
    const page = await getPage(GrainEntryMaterialRecord, "/grain-entry-materials", query);
    return {
      total: page.total,
      data: page.data.map((item) => ({
        ...item,
        imageUrl: imageApiUrl(item.imageUrl || `/grain-entry-materials?imageId=${item.id}`),
      })),
    };
  },
};
export const grainPurchaseTypeApi = crudApi(GrainPurchaseTypeRecord, "/grain-purchase-types");
export const grainPaymentMethodApi = crudApi(GrainPaymentMethodRecord, "/grain-payment-methods");
export const grainPurchasePlaceApi = crudApi(GrainPurchasePlaceRecord, "/grain-purchase-places");

export async function voidGrainPurchaseEntry(id: number) {
  const response = await instance.put<ApiResponse<{ voided: boolean }>>(`/grain-purchase-entries/${id}/void`);
  return unwrapApiResponse(response.data);
}

export async function getGrainPurchaseDashboard(params: {
  startDate?: string;
  endDate?: string;
  stationId?: number;
}) {
  return getData(GrainPurchaseDashboardRecord, "/grain-purchase-dashboard", params);
}

export async function getGrainFarmerImages(farmerId: number, appUserId?: number) {
  const response = await instance.get<ApiResponse<GrainFarmerImagesRecord>>(`/grain-farmers/${farmerId}/images`, {
    params: { appUserId },
  });
  const images = unwrapApiResponse(response.data);
  return {
    idCardFront: imageApiUrl(images.idCardFront),
    idCardBack: imageApiUrl(images.idCardBack),
    bankCard: imageApiUrl(images.bankCard),
  };
}

function imageApiUrl(path: string) {
  if (!path) {
    return "";
  }
  if (/^https?:\/\//i.test(path) || path.startsWith("data:") || path.startsWith("blob:")) {
    return path;
  }
  const [rawPath, rawQuery = ""] = path.split("?");
  const params = new URLSearchParams(rawQuery);
  const token = getAuthToken();
  if (token) {
    params.set("token", token);
  }
  const query = params.toString();
  return `/api${rawPath.startsWith("/") ? rawPath : `/${rawPath}`}${query ? `?${query}` : ""}`;
}

export async function recognizeGrainCard(
  file: File,
  payload: {
    cardType: "id-card" | "bank-card";
    stationId?: number;
    appUserId?: number;
    farmerId?: number;
    imageSide?: "front" | "back";
  },
) {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("cardType", payload.cardType);
  if (payload.stationId) {
    formData.append("stationId", String(payload.stationId));
  }
  if (payload.appUserId) {
    formData.append("appUserId", String(payload.appUserId));
  }
  if (payload.farmerId) {
    formData.append("farmerId", String(payload.farmerId));
  }
  if (payload.imageSide) {
    formData.append("imageSide", payload.imageSide);
  }
  const response = await instance.post<ApiResponse<GrainCardOcrResult>>("/file/grain-card-ocr/recognize", formData, {
    headers: { "Content-Type": "multipart/form-data" },
    timeout: 30000,
  });
  return unwrapApiResponse(response.data);
}

export async function listRegionTree() {
  const response = await instance.get<ApiResponse<RegionTreeRecord[]>>("/regions/tree");
  return unwrapApiResponse(response.data);
}
