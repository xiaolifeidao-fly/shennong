package grain_purchase

import (
	appCtx "app-api/pkg/internal/appctx"
	commonRouter "common/middleware/routers"
	grainPurchaseService "service/grain_purchase"
	grainPurchaseDTO "service/grain_purchase/dto"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GrainPurchaseHandler struct {
	*commonRouter.BaseHandler
	grainPurchaseService *grainPurchaseService.GrainPurchaseService
}

func NewGrainPurchaseHandler() *GrainPurchaseHandler {
	service := grainPurchaseService.NewGrainPurchaseService()
	_ = service.EnsureTable()
	return &GrainPurchaseHandler{
		BaseHandler:          &commonRouter.BaseHandler{},
		grainPurchaseService: service,
	}
}

func (h *GrainPurchaseHandler) RegisterHandler(engine *gin.RouterGroup) {
	engine.GET("/grain-purchase-entries", h.listEntries)
	engine.POST("/grain-purchase-entries", h.createEntry)
	engine.PUT("/grain-purchase-entries/:id", h.updateEntry)
	engine.PUT("/grain-purchase-entries/:id/void", h.voidEntry)
	engine.GET("/grain-farmer-purchase-summaries", h.listFarmerPurchaseSummaries)
	engine.GET("/grain-farmer-daily-summaries", h.listDailyFarmerSummaries)
}

func (h *GrainPurchaseHandler) listEntries(context *gin.Context) {
	var query grainPurchaseDTO.GrainPurchaseEntryQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	if query.AppUserID == 0 {
		query.AppUserID, _ = appCtx.CurrentAppUserID(context)
	}
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	query.StationID = stationID
	result, err := h.grainPurchaseService.ListEntries(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) createEntry(context *gin.Context) {
	var req grainPurchaseDTO.GrainPurchaseEntryDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	userID, _ := appCtx.CurrentAppUserID(context)
	if req.AppUserID == 0 {
		req.AppUserID = userID
	}
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	req.StationID = stationID
	result, err := h.grainPurchaseService.CreateEntry(&req, userID, appCtx.CurrentAppUserName(context))
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) updateEntry(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	var req grainPurchaseDTO.GrainPurchaseEntryDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	userID, _ := appCtx.CurrentAppUserID(context)
	if req.AppUserID == 0 {
		req.AppUserID = userID
	}
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	result, err := h.grainPurchaseService.UpdateEntryInStation(id, &req, userID, appCtx.CurrentAppUserName(context), stationID)
	if err == gorm.ErrRecordNotFound {
		commonRouter.ToError(context, "grain purchase entry not found")
		return
	}
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) voidEntry(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	userID, _ := appCtx.CurrentAppUserID(context)
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	err := h.grainPurchaseService.VoidEntryInStation(id, userID, appCtx.CurrentAppUserName(context), stationID)
	if err == gorm.ErrRecordNotFound {
		commonRouter.ToError(context, "grain purchase entry not found")
		return
	}
	commonRouter.ToJson(context, gin.H{"voided": true}, err)
}

func (h *GrainPurchaseHandler) listFarmerPurchaseSummaries(context *gin.Context) {
	var query grainPurchaseDTO.GrainFarmerPurchaseSummaryQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	if query.AppUserID == 0 {
		query.AppUserID, _ = appCtx.CurrentAppUserID(context)
	}
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	query.StationID = stationID
	result, err := h.grainPurchaseService.ListFarmerPurchaseSummaries(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) listDailyFarmerSummaries(context *gin.Context) {
	var query grainPurchaseDTO.GrainFarmerDailySummaryQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	if query.AppUserID == 0 {
		query.AppUserID, _ = appCtx.CurrentAppUserID(context)
	}
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	query.StationID = stationID
	result, err := h.grainPurchaseService.ListDailyFarmerSummaries(query)
	commonRouter.ToJson(context, result, err)
}

func parseID(context *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 32)
	if err != nil || id == 0 {
		commonRouter.ToError(context, "id必须是正整数")
		return 0, false
	}
	return uint(id), true
}

func requiredStationID(context *gin.Context) (uint64, bool) {
	stationID, ok := appCtx.CurrentStationID(context)
	if !ok {
		commonRouter.ToError(context, "粮站ID不能为空")
		return 0, false
	}
	return stationID, true
}
