package grain_entry_material

import (
	appCtx "app-api/pkg/internal/appctx"
	commonRouter "common/middleware/routers"
	"net/http"
	grainPurchaseService "service/grain_purchase"
	grainPurchaseDTO "service/grain_purchase/dto"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	if imageID := strings.TrimSpace(context.Query("imageId")); imageID != "" {
		id, err := strconv.ParseUint(imageID, 10, 32)
		if err != nil || id == 0 {
			commonRouter.ToError(context, "imageId必须是正整数")
			return
		}
		h.streamMaterialImage(context, uint(id))
		return
	}
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

func (h *GrainEntryMaterialHandler) streamMaterialImage(context *gin.Context, id uint) {
	content, err := h.grainPurchaseService.GetMaterialImageContent(id)
	if err == gorm.ErrRecordNotFound {
		context.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	if content.StationID != stationID {
		context.Status(http.StatusNotFound)
		return
	}
	context.Header("Cache-Control", "private, max-age=300")
	context.Data(http.StatusOK, content.MimeType, content.Data)
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
