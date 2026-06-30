package grain_farmer

import (
	commonRouter "common/middleware/routers"
	"log"
	"manager-api/pkg/internal/tenantctx"
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
	} else if stationIDs != nil {
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
	log.Printf("[farmer-image] list image paths request farmerID=%d appUserIdQuery=%s imageTypeQuery=%s", id, context.Query("appUserId"), context.Query("imageType"))
	if !h.ensureFarmerAccess(context, id) {
		log.Printf("[farmer-image] list image paths access denied farmerID=%d", id)
		return
	}
	appUserID, _ := strconv.ParseUint(context.Query("appUserId"), 10, 64)
	if imageType := strings.TrimSpace(context.Query("imageType")); imageType != "" {
		h.getFarmerImageURL(context, id, appUserID, imageType)
		return
	}
	result := h.imageService.GetLatestFarmerImages(uint64(id), appUserID)
	log.Printf("[farmer-image] list image paths success farmerID=%d appUserID=%d hasIDCardFront=%t hasIDCardBack=%t hasBankCard=%t", id, appUserID, result.IDCardFront != "", result.IDCardBack != "", result.BankCard != "")
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
	h.getFarmerImageURL(context, id, appUserID, imageType)
}

func (h *GrainFarmerHandler) getFarmerImageURL(context *gin.Context, id uint, appUserID uint64, imageType string) {
	imageURL := h.imageService.LatestFarmerImageURL(uint64(id), appUserID, imageType)
	if imageURL == "" {
		commonRouter.ToError(context, "image not found")
		return
	}
	commonRouter.ToJson(context, gin.H{"imageUrl": imageURL}, nil)
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
	} else if stationIDs != nil {
		if farmer.StationID == 0 {
			return true
		}
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

func parseID(context *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 32)
	if err != nil || id == 0 {
		commonRouter.ToError(context, "id必须是正整数")
		return 0, false
	}
	return uint(id), true
}
