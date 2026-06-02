package grain_purchase

import (
	commonRouter "common/middleware/routers"
	"manager-api/pkg/internal/tenantctx"
	grainPurchaseService "service/grain_purchase"
	grainPurchaseDTO "service/grain_purchase/dto"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GrainPurchaseHandler struct {
	*commonRouter.BaseHandler
	service *grainPurchaseService.GrainPurchaseService
}

func NewGrainPurchaseHandler() *GrainPurchaseHandler {
	service := grainPurchaseService.NewGrainPurchaseService()
	_ = service.EnsureTable()
	return &GrainPurchaseHandler{BaseHandler: &commonRouter.BaseHandler{}, service: service}
}

func (h *GrainPurchaseHandler) RegisterHandler(engine *gin.RouterGroup) {
	engine.GET("/grain-purchase-entries", h.listEntries)
	engine.POST("/grain-purchase-entries", h.createEntry)
	engine.PUT("/grain-purchase-entries/:id", h.updateEntry)
	engine.PUT("/grain-purchase-entries/:id/void", h.voidEntry)
	engine.DELETE("/grain-purchase-entries/:id", h.deleteEntry)
	engine.GET("/grain-farmer-purchase-summaries", h.listFarmerPurchaseSummaries)
	engine.GET("/grain-entry-snapshots", h.listSnapshots)
	engine.GET("/grain-entry-materials", h.listMaterials)
	engine.POST("/grain-entry-materials", h.createMaterial)
	engine.DELETE("/grain-entry-materials/:id", h.deleteMaterial)
}

func (h *GrainPurchaseHandler) listEntries(context *gin.Context) {
	var query grainPurchaseDTO.GrainPurchaseEntryQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	if stationIDs, ok := tenantctx.ScopedStationIDs(context); !ok {
		return
	} else if len(stationIDs) > 0 {
		query.StationIDs = stationIDs
	}
	result, err := h.service.ListEntries(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) createEntry(context *gin.Context) {
	var req grainPurchaseDTO.GrainPurchaseEntryDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.CreateEntry(&req, 0, "manager")
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
	result, err := h.service.UpdateEntry(id, &req, 0, "manager")
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
	err := h.service.VoidEntry(id, 0, "manager")
	if err == gorm.ErrRecordNotFound {
		commonRouter.ToError(context, "grain purchase entry not found")
		return
	}
	commonRouter.ToJson(context, gin.H{"voided": true}, err)
}

func (h *GrainPurchaseHandler) deleteEntry(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	err := h.service.DeleteEntry(id, 0, "manager")
	if err == gorm.ErrRecordNotFound {
		commonRouter.ToError(context, "grain purchase entry not found")
		return
	}
	commonRouter.ToJson(context, gin.H{"deleted": true}, err)
}

func (h *GrainPurchaseHandler) listSnapshots(context *gin.Context) {
	var query grainPurchaseDTO.GrainEntrySnapshotQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	if stationIDs, ok := tenantctx.ScopedStationIDs(context); !ok {
		return
	} else if len(stationIDs) > 0 {
		query.StationIDs = stationIDs
	}
	result, err := h.service.ListEntrySnapshots(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) listFarmerPurchaseSummaries(context *gin.Context) {
	var query grainPurchaseDTO.GrainFarmerPurchaseSummaryQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	if stationIDs, ok := tenantctx.ScopedStationIDs(context); !ok {
		return
	} else if len(stationIDs) > 0 {
		query.StationIDs = stationIDs
	}
	result, err := h.service.ListFarmerPurchaseSummaries(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) listMaterials(context *gin.Context) {
	var query grainPurchaseDTO.GrainEntryMaterialQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	if stationIDs, ok := tenantctx.ScopedStationIDs(context); !ok {
		return
	} else if len(stationIDs) > 0 {
		query.StationIDs = stationIDs
	}
	result, err := h.service.ListMaterials(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) createMaterial(context *gin.Context) {
	var req grainPurchaseDTO.GrainEntryMaterialDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.CreateMaterial(&req)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) deleteMaterial(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	err := h.service.DeleteMaterial(id)
	if err == gorm.ErrRecordNotFound {
		commonRouter.ToError(context, "grain entry material not found")
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
