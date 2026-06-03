package grain_config

import (
	commonRouter "common/middleware/routers"
	"manager-api/pkg/internal/tenantctx"
	grainConfigService "service/grain_config"
	grainConfigDTO "service/grain_config/dto"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GrainConfigHandler struct {
	*commonRouter.BaseHandler
	service *grainConfigService.GrainConfigService
}

func NewGrainConfigHandler() *GrainConfigHandler {
	service := grainConfigService.NewGrainConfigService()
	_ = service.EnsureTable()
	return &GrainConfigHandler{BaseHandler: &commonRouter.BaseHandler{}, service: service}
}

func (h *GrainConfigHandler) RegisterHandler(engine *gin.RouterGroup) {
	engine.GET("/grain-stations", h.listStations)
	engine.POST("/grain-stations", h.createStation)
	engine.PUT("/grain-stations/:id", h.updateStation)
	engine.DELETE("/grain-stations/:id", h.deleteStation)
	engine.GET("/grain-purchase-types", h.listPurchaseTypes)
	engine.POST("/grain-purchase-types", h.createPurchaseType)
	engine.PUT("/grain-purchase-types/:id", h.updatePurchaseType)
	engine.DELETE("/grain-purchase-types/:id", h.deletePurchaseType)
	engine.GET("/grain-payment-methods", h.listPaymentMethods)
	engine.POST("/grain-payment-methods", h.createPaymentMethod)
	engine.PUT("/grain-payment-methods/:id", h.updatePaymentMethod)
	engine.DELETE("/grain-payment-methods/:id", h.deletePaymentMethod)
	engine.GET("/grain-purchase-places", h.listPurchasePlaces)
	engine.POST("/grain-purchase-places", h.createPurchasePlace)
	engine.PUT("/grain-purchase-places/:id", h.updatePurchasePlace)
	engine.DELETE("/grain-purchase-places/:id", h.deletePurchasePlace)
}

func (h *GrainConfigHandler) listStations(context *gin.Context) {
	var query grainConfigDTO.GrainStationQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.ListStations(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) createStation(context *gin.Context) {
	var req grainConfigDTO.GrainStationDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	if tenantID := tenantctx.CurrentTenantID(context); tenantID > 0 {
		req.TenantID = tenantID
	}
	result, err := h.service.CreateStation(&req)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) updateStation(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	var req grainConfigDTO.GrainStationDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	if tenantID := tenantctx.CurrentTenantID(context); tenantID > 0 {
		req.TenantID = tenantID
	}
	result, err := h.service.UpdateStation(id, &req)
	writeResult(context, result, err, "grain station not found")
}

func (h *GrainConfigHandler) deleteStation(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	writeDelete(context, h.service.DeleteStation(id), "grain station not found")
}

func (h *GrainConfigHandler) listPurchaseTypes(context *gin.Context) {
	var query grainConfigDTO.GrainConfigItemQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	if stationIDs, ok := tenantctx.ScopedStationIDs(context); !ok {
		return
	} else if len(stationIDs) > 0 {
		query.StationIDs = stationIDs
	}
	result, err := h.service.ListPurchaseTypesPage(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) createPurchaseType(context *gin.Context) {
	var req grainConfigDTO.GrainPurchaseTypeDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.CreatePurchaseType(&req)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) updatePurchaseType(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	var req grainConfigDTO.GrainPurchaseTypeDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.UpdatePurchaseType(id, &req)
	writeResult(context, result, err, "grain purchase type not found")
}

func (h *GrainConfigHandler) deletePurchaseType(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	writeDelete(context, h.service.DeletePurchaseType(id), "grain purchase type not found")
}

func (h *GrainConfigHandler) listPaymentMethods(context *gin.Context) {
	var query grainConfigDTO.GrainConfigItemQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.ListPaymentMethodsPage(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) createPaymentMethod(context *gin.Context) {
	var req grainConfigDTO.GrainPaymentMethodDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.CreatePaymentMethod(&req)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) updatePaymentMethod(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	var req grainConfigDTO.GrainPaymentMethodDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.UpdatePaymentMethod(id, &req)
	writeResult(context, result, err, "grain payment method not found")
}

func (h *GrainConfigHandler) deletePaymentMethod(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	writeDelete(context, h.service.DeletePaymentMethod(id), "grain payment method not found")
}

func (h *GrainConfigHandler) listPurchasePlaces(context *gin.Context) {
	var query grainConfigDTO.GrainConfigItemQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.ListPurchasePlacesPage(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) createPurchasePlace(context *gin.Context) {
	var req grainConfigDTO.GrainPurchasePlaceDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.CreatePurchasePlace(&req)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) updatePurchasePlace(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	var req grainConfigDTO.GrainPurchasePlaceDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.UpdatePurchasePlace(id, &req)
	writeResult(context, result, err, "grain purchase place not found")
}

func (h *GrainConfigHandler) deletePurchasePlace(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	writeDelete(context, h.service.DeletePurchasePlace(id), "grain purchase place not found")
}

func writeResult(context *gin.Context, result interface{}, err error, notFoundMessage string) {
	if err == gorm.ErrRecordNotFound {
		commonRouter.ToError(context, notFoundMessage)
		return
	}
	commonRouter.ToJson(context, result, err)
}

func writeDelete(context *gin.Context, err error, notFoundMessage string) {
	if err == gorm.ErrRecordNotFound {
		commonRouter.ToError(context, notFoundMessage)
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
