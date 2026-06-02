package routers

import (
	"app-api/pkg/app_user"
	"app-api/pkg/grain_card_ocr"
	"app-api/pkg/grain_config"
	"app-api/pkg/grain_entry_material"
	"app-api/pkg/grain_entry_snapshot"
	"app-api/pkg/grain_farmer"
	"app-api/pkg/grain_farmer_image"
	"app-api/pkg/grain_purchase"
	"app-api/pkg/login"
	"common/middleware/routers"
)

func registerHandler() []routers.Handler {
	return []routers.Handler{
		app_user.NewAppUserHandler(),
		grain_card_ocr.NewGrainCardOcrHandler(),
		grain_config.NewGrainConfigHandler(),
		grain_farmer.NewGrainFarmerHandler(),
		grain_farmer_image.NewGrainFarmerImageHandler(),
		grain_purchase.NewGrainPurchaseHandler(),
		grain_entry_material.NewGrainEntryMaterialHandler(),
		grain_entry_snapshot.NewGrainEntrySnapshotHandler(),
		login.NewLoginHandler(),
	}
}
