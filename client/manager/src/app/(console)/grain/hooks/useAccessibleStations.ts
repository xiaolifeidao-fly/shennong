"use client";

import { useEffect, useState } from "react";
import { message } from "antd";
import { grainStationApi, type GrainStationRecord } from "../api/grain.api";

export interface AccessibleStationsResult {
  stations: GrainStationRecord[];
  stationIds: number[];
  loading: boolean;
  unrestricted: boolean;
}

/**
 * Returns the grain stations accessible to the current user.
 * Access control is resolved server-side via the current user's tenant associations.
 */
export function useAccessibleStations(): AccessibleStationsResult {
  const [stations, setStations] = useState<GrainStationRecord[]>([]);
  const [stationIds, setStationIds] = useState<number[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    void (async () => {
      setLoading(true);
      try {
        const result = await grainStationApi.list({ pageIndex: 1, pageSize: 500 });
        setStations(result.data);
        setStationIds(result.data.map((s) => s.id));
      } catch (error) {
        message.error(error instanceof Error ? error.message : "加载可访问粮站失败");
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  return { stations, stationIds, loading, unrestricted: false };
}
