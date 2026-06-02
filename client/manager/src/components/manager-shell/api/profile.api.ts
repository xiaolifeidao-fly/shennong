"use client";

import { getData, instance, unwrapApiResponse, type ApiResponse } from "@/utils/axios";

export class CurrentUserProfile {
  id!: number;

  name = "";

  username = "";

  email = "";

  phone = "";

  department = "";

  tenantId = 0;

  role = "";

  status = "";

  remark = "";
}

export interface UpdateCurrentUserProfilePayload {
  name: string;
  username: string;
  email?: string;
  phone?: string;
  department?: string;
  remark?: string;
}

export interface ChangeCurrentUserPasswordPayload {
  oldPassword: string;
  newPassword: string;
}

export async function fetchCurrentUserProfile() {
  return getData(CurrentUserProfile, "/user-profile");
}

export async function updateCurrentUserProfile(payload: UpdateCurrentUserProfilePayload) {
  const response = await instance.put<ApiResponse<CurrentUserProfile>>("/user-profile", payload);
  return unwrapApiResponse(response.data);
}

export async function changeCurrentUserPassword(payload: ChangeCurrentUserPasswordPayload) {
  const response = await instance.put<ApiResponse<{ changed: boolean }>>(
    "/user-profile/password",
    payload,
  );
  return unwrapApiResponse(response.data);
}
