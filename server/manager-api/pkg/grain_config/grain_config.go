package grain_config

import (
	commonRouter "common/middleware/routers"
	"common/middleware/storage/oss"
	"fmt"
	"io"
	"manager-api/pkg/internal/tenantctx"
	"net/http"
	"path/filepath"
	grainConfigService "service/grain_config"
	grainConfigDTO "service/grain_config/dto"
	"strconv"
	"strings"
	"time"

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
	engine.GET("/grain-stations/:id/extra", h.getStationExtra)
	engine.PUT("/grain-stations/:id/extra", h.saveStationExtra)
	engine.POST("/grain-stations/:id/extra/business-license", h.uploadBusinessLicense)
	engine.GET("/grain-stations/:id/extra/business-license", h.getBusinessLicense)
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

func (h *GrainConfigHandler) getStationExtra(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	result, err := h.service.GetStationExtra(uint64(id))
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) saveStationExtra(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	var req grainConfigDTO.GrainStationExtraDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.SaveStationExtra(uint64(id), &req)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) uploadBusinessLicense(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	file, err := context.FormFile("file")
	if err != nil {
		commonRouter.ToError(context, "请上传营业执照文件")
		return
	}
	openFile, err := file.Open()
	if err != nil {
		commonRouter.ToJson(context, nil, err)
		return
	}
	defer openFile.Close()
	data, err := io.ReadAll(openFile)
	if err != nil {
		commonRouter.ToJson(context, nil, err)
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	now := time.Now()
	objectPath := fmt.Sprintf("grain-station-extra/%d/business-license/%s%s", id, now.Format("20060102150405000"), ext)
	if err := oss.Put(objectPath, data); err != nil {
		commonRouter.ToJson(context, nil, err)
		return
	}
	ossObjectKey := ""
	ossURL := ""
	if oss.Oss != nil {
		ossObjectKey = oss.Oss.BuildKey(objectPath)
		expireDuration := 365 * 24 * time.Hour
		if url, urlErr := oss.GetUrl(objectPath, &expireDuration); urlErr == nil {
			ossURL = url
		}
	}
	result, err := h.service.SaveBusinessLicense(uint64(id), ossURL, ossObjectKey)
	if result != nil {
		result.BusinessLicenseUpdatedAt = now.UnixMilli()
	}
	commonRouter.ToJson(context, result, err)
}

func (h *GrainConfigHandler) getBusinessLicense(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	content, err := h.service.GetBusinessLicenseContent(uint64(id))
	if err == gorm.ErrRecordNotFound {
		context.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}
	context.Header("Cache-Control", "private, max-age=300")
	context.Data(http.StatusOK, content.MimeType, content.Data)
}

func (h *GrainConfigHandler) listStations(context *gin.Context) {
	var query grainConfigDTO.GrainStationQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	if stationIDs, ok := tenantctx.ScopedStationIDs(context); !ok {
		return
	} else if stationIDs != nil {
		query.StationIDs = stationIDs
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
	} else if stationIDs != nil {
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
