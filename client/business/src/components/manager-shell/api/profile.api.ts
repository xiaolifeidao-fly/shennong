"use client";

import { getData, instance, unwrapApiResponse, type ApiResponse } from "@/utils/axios";

export class CurrentUserProfile {
  id!: number;

  name = "";

  username = "";

  email = "";

  phone = "";

  department = "";

  status = "";

  remark = "";

  wxNickname = "";

  wxAvatar = "";

  stationId = 0;

  stationName = "";
}

export interface UpdateCurrentUserProfilePayload {
  name: string;
  email?: string;
  phone?: string;
  department?: string;
  remark?: string;
  wxNickname?: string;
  wxAvatar?: string;
}

export interface ChangeCurrentUserPasswordPayload {
  oldPassword: string;
  newPassword: string;
}

export async function fetchCurrentUserProfile() {
  return getData(CurrentUserProfile, "/app-user-profile");
}

export async function updateCurrentUserProfile(payload: UpdateCurrentUserProfilePayload) {
  const response = await instance.put<ApiResponse<CurrentUserProfile>>("/app-user-profile", payload);
  return unwrapApiResponse(response.data);
}

export async function changeCurrentUserPassword(payload: ChangeCurrentUserPasswordPayload) {
  const response = await instance.put<ApiResponse<{ changed: boolean }>>(
    "/app-user-profile/password",
    payload,
  );
  return unwrapApiResponse(response.data);
}
