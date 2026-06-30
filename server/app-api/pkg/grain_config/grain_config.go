package grain_config

import (
	appCtx "app-api/pkg/internal/appctx"
	commonRouter "common/middleware/routers"
	"net/http"
	grainConfigService "service/grain_config"
	grainConfigDTO "service/grain_config/dto"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GrainConfigHandler struct {
	*commonRouter.BaseHandler
	grainConfigService *grainConfigService.GrainConfigService
}

func NewGrainConfigHandler() *GrainConfigHandler {
	service := grainConfigService.NewGrainConfigService()
	_ = service.EnsureTable()
	return &GrainConfigHandler{
		BaseHandler:        &commonRouter.BaseHandler{},
		grainConfigService: service,
	}
}

func (h *GrainConfigHandler) RegisterHandler(engine *gin.RouterGroup) {
	engine.GET("/grain-stations", h.listStations)
	engine.GET("/grain-stations/mine", h.getMyStationDetail)
	engine.GET("/grain-stations/mine/extra/business-license", h.getMyBusinessLicense)
	engine.POST("/grain-stations", h.createStation)
	engine.GET("/grain-purchase-types", h.listPurchaseTypes)
	engine.POST("/grain-purchase-types", h.createPurchaseType)
	engine.GET("/grain-payment-methods", h.listPaymentMethods)
	engine.POST("/grain-payment-methods", h.createPaymentMethod)
	engine.GET("/grain-purchase-places", h.listPurchasePlaces)
	engine.POST("/grain-purchase-places", h.createPurchasePlace)
}

func (h *GrainConfigHandler) getMyStationDetail(context *gin.Context) {
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	result, err := h.grainConfigService.GetStationDetail(stationID)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) getMyBusinessLicense(context *gin.Context) {
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	result, err := h.grainConfigService.GetBusinessLicenseURL(stationID)
	if err == gorm.ErrRecordNotFound {
		context.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}
	commonRouter.ToJson(context, gin.H{"imageUrl": result.ImageURL}, nil)
}

func (h *GrainConfigHandler) listStations(context *gin.Context) {
	var query grainConfigDTO.GrainStationQueryDTO
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
	result, err := h.grainConfigService.ListStations(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) createStation(context *gin.Context) {
	var req grainConfigDTO.GrainStationDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.grainConfigService.CreateStation(&req)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) listPurchaseTypes(context *gin.Context) {
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	result, err := h.grainConfigService.ListPurchaseTypes(stationID)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) createPurchaseType(context *gin.Context) {
	var req grainConfigDTO.GrainPurchaseTypeDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	req.StationID = stationID
	result, err := h.grainConfigService.CreatePurchaseType(&req)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) listPaymentMethods(context *gin.Context) {
	result, err := h.grainConfigService.ListPaymentMethods()
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) createPaymentMethod(context *gin.Context) {
	var req grainConfigDTO.GrainPaymentMethodDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.grainConfigService.CreatePaymentMethod(&req)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) listPurchasePlaces(context *gin.Context) {
	appUserID, _ := appCtx.CurrentAppUserID(context)
	result, err := h.grainConfigService.ListPurchasePlaces(appUserID)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) createPurchasePlace(context *gin.Context) {
	var req grainConfigDTO.GrainPurchasePlaceDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	req.AppUserID, _ = appCtx.CurrentAppUserID(context)
	result, err := h.grainConfigService.CreatePurchasePlace(&req)
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
