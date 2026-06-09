package grain_farmer

import (
	appCtx "app-api/pkg/internal/appctx"
	commonRouter "common/middleware/routers"
	grainFarmerService "service/grain_farmer"
	grainFarmerDTO "service/grain_farmer/dto"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GrainFarmerHandler struct {
	*commonRouter.BaseHandler
	grainFarmerService *grainFarmerService.GrainFarmerService
}

func NewGrainFarmerHandler() *GrainFarmerHandler {
	service := grainFarmerService.NewGrainFarmerService()
	_ = service.EnsureTable()
	return &GrainFarmerHandler{
		BaseHandler:        &commonRouter.BaseHandler{},
		grainFarmerService: service,
	}
}

func (h *GrainFarmerHandler) RegisterHandler(engine *gin.RouterGroup) {
	engine.GET("/grain-farmers", h.listFarmers)
	engine.POST("/grain-farmers", h.createFarmer)
	engine.PUT("/grain-farmers/:id", h.updateFarmer)
	engine.DELETE("/grain-farmers/:id", h.deleteFarmer)
}

func (h *GrainFarmerHandler) listFarmers(context *gin.Context) {
	var query grainFarmerDTO.GrainFarmerQueryDTO
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
	result, err := h.grainFarmerService.ListFarmers(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainFarmerHandler) createFarmer(context *gin.Context) {
	var req grainFarmerDTO.GrainFarmerDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	appUserID, _ := appCtx.CurrentAppUserID(context)
	req.AppUserID = appUserID
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	req.StationID = stationID
	result, err := h.grainFarmerService.CreateFarmerForAppUser(&req, appUserID)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainFarmerHandler) updateFarmer(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	var req grainFarmerDTO.GrainFarmerDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	appUserID, _ := appCtx.CurrentAppUserID(context)
	req.AppUserID = appUserID
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	result, err := h.grainFarmerService.UpdateFarmerInStationForAppUser(id, &req, stationID, appUserID)
	if err == gorm.ErrRecordNotFound {
		commonRouter.ToError(context, "grain farmer not found")
		return
	}
	commonRouter.ToJson(context, result, err)
}

func (h *GrainFarmerHandler) deleteFarmer(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	appUserID, _ := appCtx.CurrentAppUserID(context)
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	err := h.grainFarmerService.DeleteFarmerInStationForAppUser(id, stationID, appUserID)
	if err == gorm.ErrRecordNotFound {
		commonRouter.ToError(context, "grain farmer not found")
		return
	}
	commonRouter.ToJson(context, gin.H{"deleted": true}, err)
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
