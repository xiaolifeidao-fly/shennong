package routers

import (
	"app-api/pkg/app_user"
	"app-api/pkg/login"
	"common/middleware/routers"
)

func registerHandler() []routers.Handler {
	return []routers.Handler{
		app_user.NewAppUserHandler(),
		login.NewLoginHandler(),
	}
}
