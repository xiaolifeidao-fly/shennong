"use client";

import { getData, getPage, instance, unwrapApiResponse, type ApiResponse } from "@/utils/axios";

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

class UserRoleItem {
  id!: number;
  userId = 0;
  roleId = 0;
}

class RoleResourceItem {
  id!: number;
  roleId = 0;
  resourceId = 0;
}

class ResourceItem {
  id!: number;
  pageUrl = "";
  resourceType = "";
}

export async function fetchCurrentUserPermittedPageUrls(userId: number): Promise<Set<string>> {
  const userRoles = await getPage(UserRoleItem, "/user-roles", { userId, pageIndex: 1, pageSize: 200 });
  if (userRoles.data.length === 0) {
    return new Set();
  }

  const roleIds = userRoles.data.map((r) => r.roleId);
  const roleResourceResults = await Promise.all(
    roleIds.map((roleId) =>
      getPage(RoleResourceItem, "/role-resources", { roleId, pageIndex: 1, pageSize: 500 }),
    ),
  );

  const resourceIds = new Set(roleResourceResults.flatMap((r) => r.data.map((item) => item.resourceId)));
  if (resourceIds.size === 0) {
    return new Set();
  }

  const resources = await getPage(ResourceItem, "/resources", { pageIndex: 1, pageSize: 2000 });
  const permittedUrls = new Set<string>();
  for (const resource of resources.data) {
    if (resourceIds.has(resource.id) && resource.pageUrl) {
      permittedUrls.add(resource.pageUrl);
    }
  }
  return permittedUrls;
}
