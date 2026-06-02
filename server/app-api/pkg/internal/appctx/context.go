package appctx

import (
	webAuth "app-api/auth"
	"common/middleware/db"
	authService "service/auth"
	grainConfigRepository "service/grain_config/repository"

	"github.com/gin-gonic/gin"
)

func CurrentAppUserID(context *gin.Context) (uint64, bool) {
	value, ok := context.Get(webAuth.ContextUserIDKey)
	if !ok {
		return 0, false
	}
	switch id := value.(type) {
	case uint:
		return uint64(id), id > 0
	case uint64:
		return id, id > 0
	case int:
		return uint64(id), id > 0
	case int64:
		return uint64(id), id > 0
	case float64:
		return uint64(id), id > 0
	default:
		return 0, false
	}
}

func CurrentAppUserName(context *gin.Context) string {
	value, ok := context.Get(webAuth.ContextUserKey)
	if !ok {
		return ""
	}
	user, ok := value.(*authService.LoginUser)
	if !ok || user == nil {
		return ""
	}
	if user.Name != "" {
		return user.Name
	}
	return user.Username
}

func CurrentStationID(context *gin.Context) (uint64, bool) {
	appUserID, ok := CurrentAppUserID(context)
	if !ok {
		return 0, false
	}
	repo := db.GetRepository[grainConfigRepository.GrainStationUserRepository]()
	stationUser, err := repo.FindActiveByAppUserID(appUserID)
	if err != nil || stationUser == nil || stationUser.StationID == 0 {
		return 0, false
	}
	return stationUser.StationID, true
}
