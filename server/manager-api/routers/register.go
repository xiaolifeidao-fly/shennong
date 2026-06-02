package routers

import (
	"common/middleware/routers"
	"log"
	appUser "manager-api/pkg/app_user"
	grainCardOcr "manager-api/pkg/grain_card_ocr"
	grainConfig "manager-api/pkg/grain_config"
	grainFarmer "manager-api/pkg/grain_farmer"
	grainPurchase "manager-api/pkg/grain_purchase"
	"manager-api/pkg/login"
	"manager-api/pkg/permission"
	region "manager-api/pkg/region"
	tenant "manager-api/pkg/tenant"
	"manager-api/pkg/user"

	"time"
)

func registerHandler() []routers.Handler {
	build := func(name string, fn func() routers.Handler) routers.Handler {
		start := time.Now()
		handler := fn()
		log.Printf("Handler %s initialized in %s", name, time.Since(start))
		return handler
	}

	return []routers.Handler{
		build("login", func() routers.Handler { return login.NewLoginHandler() }),
		build("permission", func() routers.Handler { return permission.NewPermissionHandler() }),
		build("tenant", func() routers.Handler { return tenant.NewTenantHandler() }),
		build("user", func() routers.Handler { return user.NewUserHandler() }),
		build("app_user", func() routers.Handler { return appUser.NewAppUserHandler() }),
		build("region", func() routers.Handler { return region.NewRegionHandler() }),
		build("grain_config", func() routers.Handler { return grainConfig.NewGrainConfigHandler() }),
		build("grain_farmer", func() routers.Handler { return grainFarmer.NewGrainFarmerHandler() }),
		build("grain_card_ocr", func() routers.Handler { return grainCardOcr.NewGrainCardOcrHandler() }),
		build("grain_purchase", func() routers.Handler { return grainPurchase.NewGrainPurchaseHandler() }),
	}
}
