package grain_card_ocr

import (
	appCtx "app-api/pkg/internal/appctx"
	commonRouter "common/middleware/routers"
	"common/middleware/storage/oss"
	"common/middleware/vipper"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	farmerImageService "service/farmer_image"
	grainFarmerService "service/grain_farmer"
	ocrService "service/ocr"
	ocrDTO "service/ocr/dto"
	ocrResultService "service/ocr_result"
	ocrResultRepository "service/ocr_result/repository"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type GrainCardOcrHandler struct {
	*commonRouter.BaseHandler
	ocrService         *ocrService.AliyunOCRService
	farmerImageService *farmerImageService.FarmerImageService
	grainFarmerService *grainFarmerService.GrainFarmerService
	ocrResultService   *ocrResultService.OcrResultService
}

func NewGrainCardOcrHandler() *GrainCardOcrHandler {
	imgSvc := farmerImageService.NewFarmerImageService()
	farmerSvc := grainFarmerService.NewGrainFarmerService()
	ocrResSvc := ocrResultService.NewOcrResultService()
	_ = imgSvc.EnsureTable()
	_ = farmerSvc.EnsureTable()
	_ = ocrResSvc.EnsureTable()
	return &GrainCardOcrHandler{
		BaseHandler:        &commonRouter.BaseHandler{},
		ocrService:         ocrService.NewAliyunOCRServiceFromConfig(),
		farmerImageService: imgSvc,
		grainFarmerService: farmerSvc,
		ocrResultService:   ocrResSvc,
	}
}

func (h *GrainCardOcrHandler) RegisterHandler(engine *gin.RouterGroup) {
	engine.POST("/grain-card-ocr/recognize", h.recognize)
}

func (h *GrainCardOcrHandler) recognize(context *gin.Context) {
	cardType := strings.TrimSpace(context.PostForm("cardType"))
	if cardType != "id-card" && cardType != "bank-card" {
		commonRouter.ToError(context, "cardType必须是id-card或bank-card")
		return
	}
	file, err := context.FormFile("file")
	if err != nil {
		commonRouter.ToError(context, "请上传识别照片")
		return
	}
	stationID, ok := appCtx.CurrentStationID(context)
	if !ok {
		commonRouter.ToError(context, "粮站ID不能为空")
		return
	}
	appUserID, _ := appCtx.CurrentAppUserID(context)

	// 可选参数：农户ID和身份证面（front/back）
	farmerIDStr := strings.TrimSpace(context.PostForm("farmerId"))
	imageSide := strings.TrimSpace(context.PostForm("imageSide"))
	if imageSide == "" {
		imageSide = "front"
	}

	openFile, err := file.Open()
	if err != nil {
		commonRouter.ToJson(context, nil, err)
		return
	}
	defer openFile.Close()
	image, err := io.ReadAll(openFile)
	if err != nil {
		commonRouter.ToJson(context, nil, err)
		return
	}

	objectPath := buildOcrObjectPath(cardType, file.Filename, stationID, appUserID)
	if err := oss.Put(objectPath, image); err != nil {
		commonRouter.ToJson(context, nil, err)
		return
	}
	expireDuration := time.Duration(vipper.GetInt64("ocr.ossUrlExpireSeconds")) * time.Second
	if expireDuration <= 0 {
		expireDuration = 10 * time.Minute
	}
	imageURL, err := oss.GetUrl(objectPath, &expireDuration)
	if err != nil {
		commonRouter.ToJson(context, nil, err)
		return
	}

	var result *ocrDTO.RecognizeCardResult

	// 有 farmerId 时启用图片记录 + OCR缓存
	ossObjectKey := ""
	if oss.Oss != nil {
		ossObjectKey = oss.Oss.BuildKey(objectPath)
	}

	farmerID := parseUint64(farmerIDStr)
	if farmerID > 0 {
		businessID, imgErr := h.resolveBusinessID(farmerID, appUserID, cardType, imageSide, file.Filename, imageURL, ossObjectKey)
		if imgErr == nil && businessID != "" {
			// 命中缓存
			if cached := h.ocrResultService.GetSuccessResult(businessID); cached != "" {
				var cachedResult ocrDTO.RecognizeCardResult
				if json.Unmarshal([]byte(cached), &cachedResult) == nil {
					cachedResult.OssURL = imageURL
					if oss.Oss != nil {
						cachedResult.OssBucket = oss.Oss.BucketName
						cachedResult.OssObjectKey = oss.Oss.BuildKey(objectPath)
					}
					if err := h.saveFarmerCardInfo(farmerID, stationID, &cachedResult); err != nil {
						commonRouter.ToJson(context, nil, err)
						return
					}
					commonRouter.ToJson(context, &cachedResult, nil)
					return
				}
			}

			// 调用OCR
			result, err = h.ocrService.RecognizeCard(context.Request.Context(), ocrDTO.RecognizeCardRequest{
				CardType:   cardType,
				FileName:   file.Filename,
				FileSize:   file.Size,
				MimeType:   file.Header.Get("Content-Type"),
				ImageURL:   imageURL,
				ImageBytes: image,
			})
			h.saveOcrResult(businessID, imageURL, result, err)
		}
	}

	// 无 farmerId 或图片记录失败时，直接调用OCR（向下兼容）
	if result == nil && err == nil {
		result, err = h.ocrService.RecognizeCard(context.Request.Context(), ocrDTO.RecognizeCardRequest{
			CardType:   cardType,
			FileName:   file.Filename,
			FileSize:   file.Size,
			MimeType:   file.Header.Get("Content-Type"),
			ImageURL:   imageURL,
			ImageBytes: image,
		})
	}

	if result != nil && oss.Oss != nil {
		result.OssBucket = oss.Oss.BucketName
		result.OssObjectKey = oss.Oss.BuildKey(objectPath)
		result.OssURL = imageURL
	}
	if err == nil && result != nil && farmerID > 0 {
		err = h.saveFarmerCardInfo(farmerID, stationID, result)
	}
	commonRouter.ToJson(context, result, err)
}

// resolveBusinessID 查找或创建图片记录，返回业务唯一ID
func (h *GrainCardOcrHandler) resolveBusinessID(farmerID, appUserID uint64, cardType, imageSide, imageName, ossURL, ossObjectKey string) (string, error) {
	if cardType == "id-card" {
		record, err := h.farmerImageService.FindOrCreateIDCardImage(farmerID, appUserID, imageSide, imageName, ossURL, ossObjectKey)
		if err != nil {
			return "", err
		}
		return farmerImageService.IDCardBusinessID(record.Id), nil
	}
	record, err := h.farmerImageService.FindOrCreateBankCardImage(farmerID, appUserID, imageName, ossURL, ossObjectKey)
	if err != nil {
		return "", err
	}
	return farmerImageService.BankCardBusinessID(record.Id), nil
}

func (h *GrainCardOcrHandler) saveOcrResult(businessID, ossURL string, result *ocrDTO.RecognizeCardResult, ocrErr error) {
	status := ocrResultRepository.OcrStatusSuccess
	resultJSON := ""
	if ocrErr != nil || result == nil {
		status = ocrResultRepository.OcrStatusFailed
	} else {
		if data, err := json.Marshal(result); err == nil {
			resultJSON = string(data)
		}
	}
	_ = h.ocrResultService.SaveResult(businessID, ossURL, resultJSON, status)
}

func (h *GrainCardOcrHandler) saveFarmerCardInfo(farmerID, stationID uint64, result *ocrDTO.RecognizeCardResult) error {
	if farmerID == 0 || result == nil {
		return nil
	}
	_, err := h.grainFarmerService.UpdateFarmerCardInfoInStation(uint(farmerID), result, stationID)
	return err
}

func buildOcrObjectPath(cardType, fileName string, stationID, appUserID uint64) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		ext = ".jpg"
	}
	now := time.Now()
	return fmt.Sprintf(
		"grain-card-ocr/%d/%d/%s/%s%s",
		stationID,
		appUserID,
		now.Format("20060102"),
		now.Format("150405000000000"),
		ext,
	)
}

func parseUint64(s string) uint64 {
	if s == "" {
		return 0
	}
	var v uint64
	fmt.Sscanf(s, "%d", &v)
	return v
}
