"use client";

const AUTH_TOKEN_KEY = "phoenix_manager_token";
const CURRENT_APP_USER_KEY = "phoenix_business_current_app_user";

export type StoredCurrentAppUser = {
  id?: number;
  name?: string;
  username?: string;
  stationId?: number;
  stationName?: string;
};

function canUseBrowserStorage() {
  return typeof window !== "undefined";
}

export function getAuthToken() {
  if (!canUseBrowserStorage()) {
    return "";
  }
  return window.localStorage.getItem(AUTH_TOKEN_KEY) || window.sessionStorage.getItem(AUTH_TOKEN_KEY) || "";
}

export function setAuthToken(token: string, remember = true) {
  if (!canUseBrowserStorage()) {
    return;
  }
  clearAuthToken();
  const storage = remember ? window.localStorage : window.sessionStorage;
  storage.setItem(AUTH_TOKEN_KEY, token);
}

export function clearAuthToken() {
  if (!canUseBrowserStorage()) {
    return;
  }
  window.localStorage.removeItem(AUTH_TOKEN_KEY);
  window.sessionStorage.removeItem(AUTH_TOKEN_KEY);
  window.localStorage.removeItem(CURRENT_APP_USER_KEY);
  window.sessionStorage.removeItem(CURRENT_APP_USER_KEY);
}

export function isAuthenticated() {
  return getAuthToken().trim().length > 0;
}

export function setCurrentAppUser(user: StoredCurrentAppUser | null | undefined, remember = true) {
  if (!canUseBrowserStorage() || !user) {
    return;
  }
  window.localStorage.removeItem(CURRENT_APP_USER_KEY);
  window.sessionStorage.removeItem(CURRENT_APP_USER_KEY);
  const storage = remember ? window.localStorage : window.sessionStorage;
  storage.setItem(CURRENT_APP_USER_KEY, JSON.stringify(user));
}

export function getCurrentAppUser() {
  if (!canUseBrowserStorage()) {
    return null;
  }
  const raw = window.localStorage.getItem(CURRENT_APP_USER_KEY) || window.sessionStorage.getItem(CURRENT_APP_USER_KEY);
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw) as StoredCurrentAppUser;
  } catch {
    return null;
  }
}
