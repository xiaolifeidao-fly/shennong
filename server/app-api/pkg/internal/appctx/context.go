package appctx

import (
	webAuth "app-api/auth"
	authService "service/auth"
	"strconv"
	"strings"

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
	for _, name := range []string{"stationId", "X-Station-Id", "X-Grain-Station-Id"} {
		value := strings.TrimSpace(context.GetHeader(name))
		if value == "" {
			continue
		}
		id, err := strconv.ParseUint(value, 10, 64)
		if err == nil && id > 0 {
			return id, true
		}
	}
	return 0, false
}
