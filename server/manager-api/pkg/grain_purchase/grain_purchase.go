package grain_purchase

import (
	commonRouter "common/middleware/routers"
	"common/middleware/storage/oss"
	"common/middleware/vipper"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	webAuth "manager-api/auth"
	"manager-api/pkg/internal/tenantctx"
	"net/http"
	"path/filepath"
	grainPurchaseService "service/grain_purchase"
	grainPurchaseDTO "service/grain_purchase/dto"
	authService "service/manager_auth"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const maxGrainEntryMaterialUploadSize = 10 << 20

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
	engine.GET("/grain-purchase-dashboard", h.getDashboard)
	engine.GET("/grain-purchase-entries", h.listEntries)
	engine.GET("/grain-purchase-entry-exports/count", h.countEntryExports)
	engine.GET("/grain-purchase-entry-exports", h.listEntryExports)
	engine.POST("/grain-purchase-entry-exports", h.createEntryExport)
	engine.GET("/grain-purchase-entry-exports/:batchNo/download", h.downloadEntryExport)
	engine.POST("/grain-purchase-entries", h.createEntry)
	engine.PUT("/grain-purchase-entries/:id", h.updateEntry)
	engine.PUT("/grain-purchase-entries/:id/void", h.voidEntry)
	engine.DELETE("/grain-purchase-entries/:id", h.deleteEntry)
	engine.GET("/grain-farmer-purchase-summaries", h.listFarmerPurchaseSummaries)
	engine.GET("/grain-entry-snapshots", h.listSnapshots)
	engine.GET("/grain-entry-materials", h.listMaterials)
	engine.GET("/grain-entry-materials/:id/image", h.getMaterialImage)
	engine.POST("/grain-entry-materials", h.createMaterial)
	engine.POST("/grain-entry-materials/upload", h.uploadMaterial)
	engine.DELETE("/grain-entry-materials/:id", h.deleteMaterial)
}

func (h *GrainPurchaseHandler) getDashboard(context *gin.Context) {
	var query grainPurchaseDTO.GrainPurchaseDashboardQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	if stationIDs, ok := tenantctx.ScopedStationIDs(context); !ok {
		return
	} else if stationIDs != nil {
		query.StationIDs = stationIDs
	}
	result, err := h.service.GetDashboard(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) listEntries(context *gin.Context) {
	query, ok := h.bindEntryQuery(context)
	if !ok {
		return
	}
	result, err := h.service.ListEntries(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) countEntryExports(context *gin.Context) {
	query, ok := h.bindEntryQuery(context)
	if !ok {
		return
	}
	result, err := h.service.CountEntriesForExport(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) listEntryExports(context *gin.Context) {
	user, ok := currentLoginUser(context)
	if !ok {
		commonRouter.ToError(context, "用户未登录")
		return
	}
	var query grainPurchaseDTO.GrainPurchaseEntryExportQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	result, err := h.service.ListEntryExportBatches(user.ID, query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) createEntryExport(context *gin.Context) {
	user, ok := currentLoginUser(context)
	if !ok {
		commonRouter.ToError(context, "用户未登录")
		return
	}
	query, ok := h.bindEntryQuery(context)
	if !ok {
		return
	}
	result, err := h.service.CreateEntryExportBatch(query, user.ID, user.Username, user.RoleIDs)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) downloadEntryExport(context *gin.Context) {
	user, ok := currentLoginUser(context)
	if !ok {
		commonRouter.ToError(context, "用户未登录")
		return
	}
	content, err := h.service.GetEntryExportFileContent(context.Param("batchNo"), user.ID)
	if err != nil {
		commonRouter.ToJson(context, nil, err)
		return
	}
	context.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", content.FileName))
	context.Data(http.StatusOK, content.MimeType, content.Data)
}

func (h *GrainPurchaseHandler) bindEntryQuery(context *gin.Context) (grainPurchaseDTO.GrainPurchaseEntryQueryDTO, bool) {
	var query grainPurchaseDTO.GrainPurchaseEntryQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return query, false
	}
	if stationIDs, ok := tenantctx.ScopedStationIDs(context); !ok {
		return query, false
	} else if stationIDs != nil {
		query.StationIDs = stationIDs
	} else if raw := strings.TrimSpace(context.Query("stationIds")); raw != "" {
		query.StationIDs = parseUint64List(raw)
	}
	if raw := strings.TrimSpace(context.Query("appUserIds")); raw != "" {
		query.AppUserIDs = parseUint64List(raw)
	}
	if raw := strings.TrimSpace(context.Query("farmerIds")); raw != "" {
		query.FarmerIDs = parseUint64List(raw)
	}
	if raw := strings.TrimSpace(context.Query("purchaseTypeIds")); raw != "" {
		query.PurchaseTypeIDs = parseUint64List(raw)
	}
	return query, true
}

func currentLoginUser(context *gin.Context) (*authService.LoginUser, bool) {
	value, ok := context.Get(webAuth.ContextUserKey)
	if !ok {
		return nil, false
	}
	user, ok := value.(*authService.LoginUser)
	return user, ok && user != nil && user.ID > 0
}

func parseUint64List(raw string) []uint64 {
	parts := strings.Split(raw, ",")
	result := make([]uint64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseUint(p, 10, 64)
		if err == nil && v > 0 {
			result = append(result, v)
		}
	}
	return result
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
	} else if stationIDs != nil {
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
	} else if stationIDs != nil {
		query.StationIDs = stationIDs
	}
	result, err := h.service.ListFarmerPurchaseSummaries(query)
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) listMaterials(context *gin.Context) {
	if imageID := strings.TrimSpace(context.Query("imageId")); imageID != "" {
		id, err := strconv.ParseUint(imageID, 10, 32)
		if err != nil || id == 0 {
			commonRouter.ToError(context, "imageId必须是正整数")
			return
		}
		log.Printf("[grain-material-image] stream image request via query imageId=%d path=%s rawQuery=%s", id, context.Request.URL.Path, context.Request.URL.RawQuery)
		h.streamMaterialImage(context, uint(id))
		return
	}
	var query grainPurchaseDTO.GrainEntryMaterialQueryDTO
	if err := context.ShouldBindQuery(&query); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	if stationIDs, ok := tenantctx.ScopedStationIDs(context); !ok {
		return
	} else if stationIDs != nil {
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

func (h *GrainPurchaseHandler) uploadMaterial(context *gin.Context) {
	file, err := context.FormFile("file")
	if err != nil {
		commonRouter.ToError(context, "请上传材料图片")
		return
	}
	if file.Size > maxGrainEntryMaterialUploadSize {
		commonRouter.ToError(context, "图片大小不能超过 10MB")
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
	stationID, ok := requiredUint64PostForm(context, "stationId")
	if !ok {
		return
	}
	if stationIDs, ok := tenantctx.ScopedStationIDs(context); !ok {
		return
	} else if stationIDs != nil && !stationAllowed(stationID, stationIDs) {
		commonRouter.ToError(context, "无权操作该粮站")
		return
	}

	openFile, err := file.Open()
	if err != nil {
		commonRouter.ToJson(context, nil, err)
		return
	}
	defer openFile.Close()
	data, err := io.ReadAll(io.LimitReader(openFile, maxGrainEntryMaterialUploadSize+1))
	if err != nil {
		commonRouter.ToJson(context, nil, err)
		return
	}
	if len(data) > maxGrainEntryMaterialUploadSize {
		commonRouter.ToError(context, "图片大小不能超过 10MB")
		return
	}

	objectPath := buildMaterialObjectPath(file.Filename, stationID, 0, entryID)
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
		MaterialBizType: strings.TrimSpace(context.PostForm("materialBizType")),
		MaterialType:    strings.TrimSpace(context.PostForm("materialType")),
		OssURL:          ossURL,
		FileName:        file.Filename,
		ImageHash:       fmt.Sprintf("%x", sha256.Sum256(data)),
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
	result, err := h.service.CreateMaterial(req)
	if result != nil && result.Id > 0 {
		result.ImageURL = result.OssURL
	}
	commonRouter.ToJson(context, result, err)
}

func (h *GrainPurchaseHandler) getMaterialImage(context *gin.Context) {
	id, ok := parseID(context)
	if !ok {
		return
	}
	h.streamMaterialImage(context, id)
}

func (h *GrainPurchaseHandler) streamMaterialImage(context *gin.Context, id uint) {
	log.Printf("[grain-material-image] image url request id=%d path=%s", id, context.Request.URL.Path)
	result, err := h.service.GetMaterialImageURL(id)
	if err == gorm.ErrRecordNotFound {
		log.Printf("[grain-material-image] image url not found id=%d", id)
		context.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("[grain-material-image] image url failed id=%d err=%v", id, err)
		commonRouter.ToError(context, err.Error())
		return
	}
	if stationIDs, ok := tenantctx.ScopedStationIDs(context); !ok {
		log.Printf("[grain-material-image] image url tenant scope missing id=%d stationID=%d", id, result.StationID)
		return
	} else if stationIDs != nil && !stationAllowed(result.StationID, stationIDs) {
		log.Printf("[grain-material-image] image url station denied id=%d stationID=%d scopedStationIDs=%v", id, result.StationID, stationIDs)
		context.Status(http.StatusNotFound)
		return
	}
	commonRouter.ToJson(context, gin.H{"imageUrl": result.ImageURL}, nil)
	log.Printf("[grain-material-image] image url success id=%d stationID=%d", id, result.StationID)
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

func stationAllowed(stationID uint64, stationIDs []uint64) bool {
	for _, scopedStationID := range stationIDs {
		if stationID == scopedStationID {
			return true
		}
	}
	return false
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

func parseID(context *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 32)
	if err != nil || id == 0 {
		commonRouter.ToError(context, "id必须是正整数")
		return 0, false
	}
	return uint(id), true
}
