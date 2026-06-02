package grain_entry_material

import (
	appCtx "app-api/pkg/internal/appctx"
	commonRouter "common/middleware/routers"
	grainPurchaseService "service/grain_purchase"
	grainPurchaseDTO "service/grain_purchase/dto"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GrainEntryMaterialHandler struct {
	*commonRouter.BaseHandler
	grainPurchaseService *grainPurchaseService.GrainPurchaseService
}

func NewGrainEntryMaterialHandler() *GrainEntryMaterialHandler {
	service := grainPurchaseService.NewGrainPurchaseService()
	_ = service.EnsureTable()
	return &GrainEntryMaterialHandler{
		BaseHandler:          &commonRouter.BaseHandler{},
		grainPurchaseService: service,
	}
}

func (h *GrainEntryMaterialHandler) RegisterHandler(engine *gin.RouterGroup) {
	engine.GET("/grain-entry-materials", h.listMaterials)
	engine.POST("/grain-entry-materials", h.createMaterial)
	engine.DELETE("/grain-entry-materials/:id", h.deleteMaterial)
}

func (h *GrainEntryMaterialHandler) listMaterials(context *gin.Context) {
	var query grainPurchaseDTO.GrainEntryMaterialQueryDTO
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
	result, err := h.grainPurchaseService.ListMaterials(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainEntryMaterialHandler) createMaterial(context *gin.Context) {
	var req grainPurchaseDTO.GrainEntryMaterialDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	req.AppUserID, _ = appCtx.CurrentAppUserID(context)
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	req.StationID = stationID
	result, err := h.grainPurchaseService.CreateMaterial(&req)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainEntryMaterialHandler) deleteMaterial(context *gin.Context) {
	idStr := context.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		commonRouter.ToError(context, "参数错误")
		return
	}
	commonRouter.ToJson(context, nil, h.grainPurchaseService.DeleteMaterial(uint(id)))
}

func requiredStationID(context *gin.Context) (uint64, bool) {
	stationID, ok := appCtx.CurrentStationID(context)
	if !ok {
		commonRouter.ToError(context, "粮站ID不能为空")
		return 0, false
	}
	return stationID, true
}
