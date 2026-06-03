"use client";

export const APP_BASE_PATH = process.env.APP_BASE_PATH || "";

export function withBasePath(path: string) {
  if (!APP_BASE_PATH || path.startsWith(APP_BASE_PATH)) {
    return path;
  }
  return `${APP_BASE_PATH}${path.startsWith("/") ? path : `/${path}`}`;
}

export function withoutBasePath(path: string) {
  if (!APP_BASE_PATH || !path.startsWith(APP_BASE_PATH)) {
    return path;
  }
  return path.slice(APP_BASE_PATH.length) || "/";
}
