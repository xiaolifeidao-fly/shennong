package grain_entry_material

import (
	appCtx "app-api/pkg/internal/appctx"
	commonRouter "common/middleware/routers"
	"common/middleware/storage/oss"
	"common/middleware/vipper"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	grainPurchaseService "service/grain_purchase"
	grainPurchaseDTO "service/grain_purchase/dto"
	"strconv"
	"strings"
	"time"

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
	engine.POST("/grain-entry-materials/upload", h.uploadMaterial)
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

func (h *GrainEntryMaterialHandler) uploadMaterial(context *gin.Context) {
	file, err := context.FormFile("file")
	if err != nil {
		commonRouter.ToError(context, "请上传材料图片")
		return
	}
	entryID, ok := requiredUint64PostForm(context, "entryId")
	if !ok {
		return
	}
	farmerID, ok := requiredUint64PostForm(context, "farmerId")
	if !ok {
		return
	}
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	appUserID, _ := appCtx.CurrentAppUserID(context)

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

	objectPath := buildMaterialObjectPath(file.Filename, stationID, appUserID, entryID)
	if err := oss.Put(objectPath, data); err != nil {
		commonRouter.ToJson(context, nil, err)
		return
	}
	expireDuration := time.Duration(vipper.GetInt64("oss.expireTime")) * time.Second
	if expireDuration <= 0 {
		expireDuration = 10 * time.Minute
	}
	ossURL, err := oss.GetUrl(objectPath, &expireDuration)
	if err != nil {
		commonRouter.ToJson(context, nil, err)
		return
	}

	req := &grainPurchaseDTO.GrainEntryMaterialDTO{
		StationID:       stationID,
		EntryID:         entryID,
		FarmerID:        farmerID,
		AppUserID:       appUserID,
		MaterialBizType: strings.TrimSpace(context.PostForm("materialBizType")),
		MaterialType:    strings.TrimSpace(context.PostForm("materialType")),
		OssURL:          ossURL,
		FileName:        file.Filename,
		FileSize:        file.Size,
		MimeType:        file.Header.Get("Content-Type"),
		SortOrder:       parseIntPostForm(context, "sortOrder"),
	}
	if req.MaterialBizType == "" {
		req.MaterialBizType = "entry"
	}
	if req.MaterialType == "" {
		req.MaterialType = "image"
	}
	if oss.Oss != nil {
		req.OssBucket = oss.Oss.BucketName
		req.OssObjectKey = oss.Oss.BuildKey(objectPath)
	}
	result, err := h.grainPurchaseService.CreateMaterial(req)
	if result != nil && result.Id > 0 {
		result.ImageURL = fmt.Sprintf("/grain-entry-materials?imageId=%d", result.Id)
	}
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

func requiredUint64PostForm(context *gin.Context, name string) (uint64, bool) {
	value := strings.TrimSpace(context.PostForm(name))
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		commonRouter.ToError(context, name+"必须是正整数")
		return 0, false
	}
	return id, true
}

func parseIntPostForm(context *gin.Context, name string) int {
	value := strings.TrimSpace(context.PostForm(name))
	if value == "" {
		return 0
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return result
}

func buildMaterialObjectPath(fileName string, stationID, appUserID, entryID uint64) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		ext = ".jpg"
	}
	now := time.Now()
	return fmt.Sprintf(
		"grain-entry-material/%d/%d/%d/%s/%s%s",
		stationID,
		appUserID,
		entryID,
		now.Format("20060102"),
		now.Format("150405000000000"),
		ext,
	)
}
