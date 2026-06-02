package grain_farmer

import (
	commonRouter "common/middleware/routers"
	"manager-api/pkg/internal/tenantctx"
	"net/http"
	"net/url"
	farmerImageService "service/farmer_image"
	grainFarmerService "service/grain_farmer"
	grainFarmerDTO "service/grain_farmer/dto"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GrainFarmerHandler struct {
	*commonRouter.BaseHandler
	service      *grainFarmerService.GrainFarmerService
	imageService *farmerImageService.FarmerImageService
}

func NewGrainFarmerHandler() *GrainFarmerHandler {
	service := grainFarmerService.NewGrainFarmerService()
	imageService := farmerImageService.NewFarmerImageService()
	_ = service.EnsureTable()
	_ = imageService.EnsureTable()
	return &GrainFarmerHandler{BaseHandler: &commonRouter.BaseHandler{}, service: service, imageService: imageService}
}

func (h *GrainFarmerHandler) RegisterHandler(engine *gin.RouterGroup) {
	engine.GET("/grain-farmers", h.listFarmers)
	engine.GET("/grain-farmers/:id/images", h.getFarmerImages)
	engine.GET("/grain-farmers/:id/images/:imageType", h.getFarmerImage)
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
	if stationIDs, ok := tenantctx.ScopedStationIDs(context); !ok {
		return
	} else if len(stationIDs) > 0 {
		query.StationIDs = stationIDs
	}
	result, err := h.service.ListFarmers(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainFarmerHandler) getFarmerImages(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	if !h.ensureFarmerAccess(context, id) {
		return
	}
	appUserID, _ := strconv.ParseUint(context.Query("appUserId"), 10, 64)
	if imageType := strings.TrimSpace(context.Query("imageType")); imageType != "" {
		h.streamFarmerImage(context, id, appUserID, imageType)
		return
	}
	result := &farmerImageService.FarmerImagesResult{}
	if h.imageService.HasLatestFarmerImage(uint64(id), appUserID, "id-card-front") {
		result.IDCardFront = h.farmerImagePath(id, appUserID, "id-card-front")
	}
	if h.imageService.HasLatestFarmerImage(uint64(id), appUserID, "id-card-back") {
		result.IDCardBack = h.farmerImagePath(id, appUserID, "id-card-back")
	}
	if h.imageService.HasLatestFarmerImage(uint64(id), appUserID, "bank-card") {
		result.BankCard = h.farmerImagePath(id, appUserID, "bank-card")
	}
	commonRouter.ToJson(context, result, nil)
}

func (h *GrainFarmerHandler) getFarmerImage(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	if !h.ensureFarmerAccess(context, id) {
		return
	}
	appUserID, _ := strconv.ParseUint(context.Query("appUserId"), 10, 64)
	imageType := strings.TrimSpace(context.Param("imageType"))
	h.streamFarmerImage(context, id, appUserID, imageType)
}

func (h *GrainFarmerHandler) streamFarmerImage(context *gin.Context, id uint, appUserID uint64, imageType string) {
	content, err := h.imageService.GetLatestFarmerImageContent(uint64(id), appUserID, imageType)
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

func (h *GrainFarmerHandler) createFarmer(context *gin.Context) {
	var req grainFarmerDTO.GrainFarmerDTO
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.CreateFarmer(&req)
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
	result, err := h.service.UpdateFarmer(id, &req)
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
	err := h.service.DeleteFarmer(id)
	if err == gorm.ErrRecordNotFound {
		commonRouter.ToError(context, "grain farmer not found")
		return
	}
	commonRouter.ToJson(context, gin.H{"deleted": true}, err)
}

func (h *GrainFarmerHandler) ensureFarmerAccess(context *gin.Context, id uint) bool {
	farmer, err := h.service.GetFarmer(id)
	if err == gorm.ErrRecordNotFound {
		commonRouter.ToError(context, "grain farmer not found")
		return false
	}
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return false
	}
	if stationIDs, ok := tenantctx.ScopedStationIDs(context); !ok {
		return false
	} else if len(stationIDs) > 0 {
		for _, stationID := range stationIDs {
			if farmer.StationID == stationID {
				return true
			}
		}
		commonRouter.ToError(context, "grain farmer not found")
		return false
	}
	return true
}

func (h *GrainFarmerHandler) farmerImagePath(farmerID uint, appUserID uint64, imageType string) string {
	path := "/grain-farmers/" + strconv.FormatUint(uint64(farmerID), 10) + "/images?imageType=" + url.QueryEscape(imageType)
	if appUserID > 0 {
		path += "&appUserId=" + strconv.FormatUint(appUserID, 10)
	}
	return path
}

func parseID(context *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 32)
	if err != nil || id == 0 {
		commonRouter.ToError(context, "id必须是正整数")
		return 0, false
	}
	return uint(id), true
}
