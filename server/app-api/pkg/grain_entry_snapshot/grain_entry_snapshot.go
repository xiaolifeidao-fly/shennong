package grain_entry_snapshot

import (
	appCtx "app-api/pkg/internal/appctx"
	commonRouter "common/middleware/routers"
	grainPurchaseService "service/grain_purchase"
	grainPurchaseDTO "service/grain_purchase/dto"

	"github.com/gin-gonic/gin"
)

type GrainEntrySnapshotHandler struct {
	*commonRouter.BaseHandler
	grainPurchaseService *grainPurchaseService.GrainPurchaseService
}

func NewGrainEntrySnapshotHandler() *GrainEntrySnapshotHandler {
	service := grainPurchaseService.NewGrainPurchaseService()
	_ = service.EnsureTable()
	return &GrainEntrySnapshotHandler{
		BaseHandler:          &commonRouter.BaseHandler{},
		grainPurchaseService: service,
	}
}

func (h *GrainEntrySnapshotHandler) RegisterHandler(engine *gin.RouterGroup) {
	engine.GET("/grain-entry-snapshots", h.listEntrySnapshots)
}

func (h *GrainEntrySnapshotHandler) listEntrySnapshots(context *gin.Context) {
	var query grainPurchaseDTO.GrainEntrySnapshotQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	query.AppUserID, _ = appCtx.CurrentAppUserID(context)
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	query.StationID = stationID
	result, err := h.grainPurchaseService.ListEntrySnapshots(query)
	commonRouter.ToJson(context, result, err)
}

func requiredStationID(context *gin.Context) (uint64, bool) {
	stationID, ok := appCtx.CurrentStationID(context)
	if !ok {
		commonRouter.ToError(context, "粮站ID不能为空")
		return 0, false
	}
	return stationID, true
}
