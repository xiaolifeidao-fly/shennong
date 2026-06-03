package tenantctx

import (
	commonRouter "common/middleware/routers"
	"common/middleware/db"
	tenantRepository "service/tenant/repository"
	userRepository "service/manager_user/repository"

	"github.com/gin-gonic/gin"
)

const contextScopedStationIDsKey = "tenantctx.scopedStationIDs"

// ScopedStationIDs returns the station IDs accessible to the current user based on
// their tenant associations. Returns (nil, true) if the user has no tenant restrictions.
// Returns (nil, false) and writes an error response if the user ID cannot be resolved.
func ScopedStationIDs(c *gin.Context) ([]uint64, bool) {
	if cached, exists := c.Get(contextScopedStationIDsKey); exists {
		ids, _ := cached.([]uint64)
		return ids, true
	}

	rawUserID, exists := c.Get("auth.userId")
	if !exists {
		commonRouter.ToError(c, "未登录")
		c.Abort()
		return nil, false
	}
	userID, _ := rawUserID.(uint64)
	if userID == 0 {
		c.Set(contextScopedStationIDsKey, []uint64(nil))
		return nil, true
	}

	userTenantRepo := db.GetRepository[userRepository.UserTenantRepository]()
	tenantIDs, err := userTenantRepo.ListActiveTenantIDs(userID)
	if err != nil || len(tenantIDs) == 0 {
		// No tenant associations → no restriction (super-admin behaviour)
		c.Set(contextScopedStationIDsKey, []uint64(nil))
		return nil, true
	}

	tenantStationRepo := db.GetRepository[tenantRepository.TenantStationRepository]()
	stationIDSet := make(map[uint64]struct{})
	for _, tenantID := range tenantIDs {
		ids, err := tenantStationRepo.ListActiveStationIDs(tenantID)
		if err != nil {
			continue
		}
		for _, id := range ids {
			stationIDSet[id] = struct{}{}
		}
	}

	stationIDs := make([]uint64, 0, len(stationIDSet))
	for id := range stationIDSet {
		stationIDs = append(stationIDs, id)
	}

	c.Set(contextScopedStationIDsKey, stationIDs)
	return stationIDs, true
}

func CurrentTenantID(context *gin.Context) uint64 {
	return 0
}
