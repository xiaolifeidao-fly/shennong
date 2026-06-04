package tenantctx

import (
	commonRouter "common/middleware/routers"
	"common/middleware/db"
	permissionRepository "service/manager_permission/repository"
	tenantRepository "service/tenant/repository"
	userRepository "service/manager_user/repository"

	"github.com/gin-gonic/gin"
)

const contextScopedStationIDsKey = "tenantctx.scopedStationIDs"

// ScopedStationIDs returns the station IDs accessible to the current user.
// Returns (nil, true)        → admin user, no restriction.
// Returns ([]uint64{}, true) → non-admin user with no tenant, restrict to nothing.
// Returns (ids, true)        → restrict to these station IDs.
// Returns (nil, false)       → error, response already written.
func ScopedStationIDs(c *gin.Context) ([]uint64, bool) {
	if cached, exists := c.Get(contextScopedStationIDsKey); exists {
		// cached nil means admin (no restriction); cached non-nil may be empty slice
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

	// Check if the user has super_admin role → no station restriction
	if isAdmin, _ := userHasSuperAdmin(userID); isAdmin {
		c.Set(contextScopedStationIDsKey, []uint64(nil))
		return nil, true
	}

	userTenantRepo := db.GetRepository[userRepository.UserTenantRepository]()
	tenantIDs, err := userTenantRepo.ListActiveTenantIDs(userID)
	if err != nil || len(tenantIDs) == 0 {
		// Non-admin with no tenant → can see nothing
		c.Set(contextScopedStationIDsKey, []uint64{})
		return []uint64{}, true
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

// userHasSuperAdmin checks whether the user has a role with code "super_admin".
func userHasSuperAdmin(userID uint64) (bool, error) {
	roleRepo := db.GetRepository[permissionRepository.RoleRepository]()
	if roleRepo.Db == nil {
		return false, nil
	}
	var count int64
	err := roleRepo.Db.Raw(`
		SELECT COUNT(1) FROM user_role ur
		JOIN role r ON r.id = ur.role_id AND r.active = 1
		WHERE ur.active = 1 AND ur.user_id = ? AND r.code = 'super_admin'
	`, userID).Scan(&count).Error
	return count > 0, err
}

func CurrentTenantID(context *gin.Context) uint64 {
	return 0
}
