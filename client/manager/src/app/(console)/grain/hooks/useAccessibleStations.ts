"use client";

import { useEffect, useState } from "react";
import { message } from "antd";
import { fetchCurrentUserProfile } from "@/components/manager-shell/api/profile.api";
import { fetchUserDetail } from "../../user/api/user.api";
import { tenantApi } from "../../tenant/api/tenant.api";
import { grainStationApi, type GrainStationRecord } from "../api/grain.api";

export interface AccessibleStationsResult {
  stations: GrainStationRecord[];
  stationIds: number[];
  loading: boolean;
  // true means "no tenant restriction" (super-admin), show all stations
  unrestricted: boolean;
}

/**
 * Fetches the grain stations accessible to the current user via their tenant associations.
 * If the user has no tenant associations, all stations are returned (unrestricted).
 */
export function useAccessibleStations(): AccessibleStationsResult {
  const [stations, setStations] = useState<GrainStationRecord[]>([]);
  const [stationIds, setStationIds] = useState<number[]>([]);
  const [loading, setLoading] = useState(true);
  const [unrestricted, setUnrestricted] = useState(false);

  useEffect(() => {
    void (async () => {
      setLoading(true);
      try {
        const profile = await fetchCurrentUserProfile();
        const userDetail = await fetchUserDetail(profile.id);
        const tenantIds = userDetail.tenantIds ?? [];

        if (tenantIds.length === 0) {
          // No tenant restriction — load all stations
          const result = await grainStationApi.list({ pageIndex: 1, pageSize: 500 });
          setStations(result.data);
          setStationIds(result.data.map((s) => s.id));
          setUnrestricted(true);
          return;
        }

        // Collect station IDs from all user's tenants
        const tenantPage = await tenantApi.list({ pageIndex: 1, pageSize: 500 });
        const userTenants = tenantPage.data.filter((t) => tenantIds.includes(t.id));
        const accessibleStationIds = new Set<number>();
        for (const tenant of userTenants) {
          for (const id of tenant.stationIds ?? []) {
            accessibleStationIds.add(id);
          }
        }

        if (accessibleStationIds.size === 0) {
          // Tenants exist but no stations configured — show nothing
          setStations([]);
          setStationIds([]);
          setUnrestricted(false);
          return;
        }

        const allStations = await grainStationApi.list({ pageIndex: 1, pageSize: 500 });
        const filtered = allStations.data.filter((s) => accessibleStationIds.has(s.id));
        setStations(filtered);
        setStationIds(filtered.map((s) => s.id));
        setUnrestricted(false);
      } catch (error) {
        message.error(error instanceof Error ? error.message : "加载可访问粮站失败");
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  return { stations, stationIds, loading, unrestricted };
}
