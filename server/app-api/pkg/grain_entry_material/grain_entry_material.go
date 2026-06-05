package grain_entry_material

import (
	appCtx "app-api/pkg/internal/appctx"
	commonRouter "common/middleware/routers"
	"common/middleware/storage/oss"
	"common/middleware/vipper"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"mime"
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

type uploadMaterialJSONRequest struct {
	EntryID         uint64 `json:"entryId"`
	FarmerID        uint64 `json:"farmerId"`
	MaterialBizType string `json:"materialBizType"`
	MaterialType    string `json:"materialType"`
	FileName        string `json:"fileName"`
	FileSize        int64  `json:"fileSize"`
	MimeType        string `json:"mimeType"`
	ImageURL        string `json:"imageUrl"`
	CloudFileID     string `json:"cloudFileId"`
	SortOrder       int    `json:"sortOrder"`
}

type uploadMaterialInput struct {
	entryID         uint64
	farmerID        uint64
	materialBizType string
	materialType    string
	fileName        string
	fileSize        int64
	mimeType        string
	sortOrder       int
	data            []byte
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
	input, err := readUploadMaterialInput(context)
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}
	stationID, ok := requiredStationID(context)
	if !ok {
		return
	}
	appUserID, _ := appCtx.CurrentAppUserID(context)

	objectPath := buildMaterialObjectPath(input.fileName, stationID, appUserID, input.entryID)
	if err := oss.Put(objectPath, input.data); err != nil {
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
		EntryID:         input.entryID,
		FarmerID:        input.farmerID,
		AppUserID:       appUserID,
		MaterialBizType: input.materialBizType,
		MaterialType:    input.materialType,
		OssURL:          ossURL,
		FileName:        input.fileName,
		ImageHash:       fmt.Sprintf("%x", sha256.Sum256(input.data)),
		FileSize:        input.fileSize,
		MimeType:        input.mimeType,
		SortOrder:       input.sortOrder,
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

func readUploadMaterialInput(context *gin.Context) (*uploadMaterialInput, error) {
	if strings.HasPrefix(context.ContentType(), "application/json") {
		var req uploadMaterialJSONRequest
		if err := context.ShouldBindJSON(&req); err != nil {
			return nil, errors.New("参数错误")
		}
		if req.EntryID == 0 {
			return nil, errors.New("entryId必须是正整数")
		}
		if req.FarmerID == 0 {
			return nil, errors.New("farmerId必须是正整数")
		}
		data, mimeType, err := downloadCloudImage(strings.TrimSpace(req.ImageURL))
		if err != nil {
			return nil, err
		}
		fileName := strings.TrimSpace(req.FileName)
		if fileName == "" {
			fileName = filepath.Base(strings.TrimSpace(req.CloudFileID))
		}
		if fileName == "." || fileName == "/" || fileName == "" {
			fileName = "material.jpg"
		}
		if strings.TrimSpace(req.MimeType) != "" {
			mimeType = strings.TrimSpace(req.MimeType)
		}
		fileSize := req.FileSize
		if fileSize <= 0 {
			fileSize = int64(len(data))
		}
		return &uploadMaterialInput{
			entryID:         req.EntryID,
			farmerID:        req.FarmerID,
			materialBizType: strings.TrimSpace(req.MaterialBizType),
			materialType:    strings.TrimSpace(req.MaterialType),
			fileName:        fileName,
			fileSize:        fileSize,
			mimeType:        mimeType,
			sortOrder:       req.SortOrder,
			data:            data,
		}, nil
	}

	file, err := context.FormFile("file")
	if err != nil {
		return nil, errors.New("请上传材料图片")
	}
	entryID, err := requiredUint64PostFormValue(context, "entryId")
	if err != nil {
		return nil, err
	}
	farmerID, err := requiredUint64PostFormValue(context, "farmerId")
	if err != nil {
		return nil, err
	}
	openFile, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer openFile.Close()
	data, err := io.ReadAll(openFile)
	if err != nil {
		return nil, err
	}
	return &uploadMaterialInput{
		entryID:         entryID,
		farmerID:        farmerID,
		materialBizType: strings.TrimSpace(context.PostForm("materialBizType")),
		materialType:    strings.TrimSpace(context.PostForm("materialType")),
		fileName:        file.Filename,
		fileSize:        file.Size,
		mimeType:        file.Header.Get("Content-Type"),
		sortOrder:       parseIntPostForm(context, "sortOrder"),
		data:            data,
	}, nil
}

func downloadCloudImage(rawURL string) ([]byte, string, error) {
	if rawURL == "" {
		return nil, "", errors.New("imageUrl不能为空")
	}
	if !strings.HasPrefix(rawURL, "https://") && !strings.HasPrefix(rawURL, "http://") {
		return nil, "", errors.New("imageUrl必须是HTTP地址")
	}
	response, err := (&http.Client{Timeout: 20 * time.Second}).Get(rawURL)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, "", fmt.Errorf("下载云存储图片失败：%d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 10<<20+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) == 0 {
		return nil, "", errors.New("云存储图片为空")
	}
	if len(data) > 10<<20 {
		return nil, "", errors.New("图片不能超过10MB")
	}
	contentType := response.Header.Get("Content-Type")
	if contentType != "" {
		if parsed, _, err := mime.ParseMediaType(contentType); err == nil {
			contentType = parsed
		}
	}
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	return data, contentType, nil
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
	id, err := requiredUint64PostFormValue(context, name)
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return 0, false
	}
	return id, true
}

func requiredUint64PostFormValue(context *gin.Context, name string) (uint64, error) {
	value := strings.TrimSpace(context.PostForm(name))
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New(name + "必须是正整数")
	}
	return id, nil
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
