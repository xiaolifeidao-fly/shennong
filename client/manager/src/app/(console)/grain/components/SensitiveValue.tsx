"use client";

import { Typography } from "antd";

const { Text } = Typography;

function renderValue(value: unknown) {
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  return String(value);
}

function maskSensitiveValue(text: string, keepStart: number, keepEnd: number) {
  if (text === "-" || text.length <= keepStart + keepEnd) {
    return text;
  }
  return `${text.slice(0, keepStart)}${"*".repeat(Math.max(4, text.length - keepStart - keepEnd))}${text.slice(-keepEnd)}`;
}

export function SensitiveValue({ value, keepStart, keepEnd }: { value: unknown; keepStart: number; keepEnd: number }) {
  const text = renderValue(value);
  const maskedText = maskSensitiveValue(text, keepStart, keepEnd);
  return <Text copyable={maskedText !== "-" ? { text: maskedText } : false}>{maskedText}</Text>;
}
