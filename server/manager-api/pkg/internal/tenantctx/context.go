package tenantctx

import (
	commonRouter "common/middleware/routers"
	webAuth "manager-api/auth"
	authService "service/manager_auth"
	tenantService "service/tenant"

	"github.com/gin-gonic/gin"
)

func ScopedStationIDs(context *gin.Context) ([]uint64, bool) {
	tenantID := CurrentTenantID(context)
	if tenantID == 0 {
		return nil, true
	}
	stationIDs, err := tenantService.NewTenantService().ListStationIDsByTenant(tenantID)
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return nil, false
	}
	if len(stationIDs) == 0 {
		return []uint64{0}, true
	}
	return stationIDs, true
}

func CurrentTenantID(context *gin.Context) uint64 {
	value, ok := context.Get(webAuth.ContextUserKey)
	if !ok {
		return 0
	}
	if user, ok := value.(*authService.LoginUser); ok {
		return user.TenantID
	}
	return 0
}
