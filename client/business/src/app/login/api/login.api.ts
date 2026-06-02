"use client";

import { instance, unwrapApiResponse, type ApiResponse } from "@/utils/axios";
import type { StoredCurrentAppUser } from "@/utils/auth";

export interface LoginPayload {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user?: StoredCurrentAppUser;
}

export async function login(payload: LoginPayload) {
  const response = await instance.post<ApiResponse<LoginResponse>>("/login", payload);
  return unwrapApiResponse(response.data);
}
