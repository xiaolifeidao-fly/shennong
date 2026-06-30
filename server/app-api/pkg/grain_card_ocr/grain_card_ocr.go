package grain_card_ocr

import (
	appCtx "app-api/pkg/internal/appctx"
	commonRouter "common/middleware/routers"
	"common/middleware/storage/image_source"
	"common/middleware/storage/oss"
	"common/middleware/vipper"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
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

const maxCloudOcrImageSize = 10 << 20

type GrainCardOcrHandler struct {
	*commonRouter.BaseHandler
	ocrService         *ocrService.AliyunOCRService
	farmerImageService *farmerImageService.FarmerImageService
	grainFarmerService *grainFarmerService.GrainFarmerService
	ocrResultService   *ocrResultService.OcrResultService
}

type recognizeCardJSONRequest struct {
	CardType    string `json:"cardType"`
	FarmerID    string `json:"farmerId"`
	ImageSide   string `json:"imageSide"`
	FileName    string `json:"fileName"`
	FileSize    int64  `json:"fileSize"`
	MimeType    string `json:"mimeType"`
	ImageURL    string `json:"imageUrl"`
	CloudFileID string `json:"cloudFileId"`
}

type recognizeCardInput struct {
	cardType    string
	farmerID    string
	imageSide   string
	fileName    string
	fileSize    int64
	mimeType    string
	imageURL    string
	cloudFileID string
	image       []byte
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
	input, err := readRecognizeCardInput(context)
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}
	if input.cardType != "id-card" && input.cardType != "bank-card" && input.cardType != "social-security-card" {
		commonRouter.ToError(context, "cardType必须是id-card、bank-card或social-security-card")
		return
	}
	stationID, ok := appCtx.CurrentStationID(context)
	if !ok {
		commonRouter.ToError(context, "粮站ID不能为空")
		return
	}
	appUserID, _ := appCtx.CurrentAppUserID(context)

	if input.imageSide == "" {
		input.imageSide = "front"
	}
	log.Printf(
		"[grain-card-ocr] recognize input: cardType=%s farmerId=%s imageSide=%s fileName=%s fileSize=%d mimeType=%s cloudFileIdSet=%t imageUrlHost=%s imageBytes=%d",
		input.cardType,
		input.farmerID,
		input.imageSide,
		input.fileName,
		input.fileSize,
		input.mimeType,
		input.cloudFileID != "",
		urlHost(input.imageURL),
		len(input.image),
	)

	objectPath := buildOcrObjectPath(input.cardType, input.fileName, stationID, appUserID)
	log.Printf(
		"[grain-card-ocr] oss put image: stationId=%d appUserId=%d objectPath=%s bytes=%d",
		stationID,
		appUserID,
		objectPath,
		len(input.image),
	)
	if err := oss.Put(objectPath, input.image); err != nil {
		log.Printf(
			"[grain-card-ocr] oss put image failed: stationId=%d appUserId=%d objectPath=%s err=%v",
			stationID,
			appUserID,
			objectPath,
			err,
		)
		commonRouter.ToJson(context, nil, err)
		return
	}
	log.Printf("[grain-card-ocr] oss put image success: objectPath=%s", objectPath)
	expireDuration := time.Duration(vipper.GetInt64("ocr.ossUrlExpireSeconds")) * time.Second
	if expireDuration <= 0 {
		expireDuration = 10 * time.Minute
	}
	imageURL, err := oss.GetUrl(objectPath, &expireDuration)
	if err != nil {
		log.Printf("[grain-card-ocr] oss sign url failed: objectPath=%s expire=%s err=%v", objectPath, expireDuration, err)
		commonRouter.ToJson(context, nil, err)
		return
	}
	log.Printf("[grain-card-ocr] oss sign url success: objectPath=%s expire=%s host=%s", objectPath, expireDuration, urlHost(imageURL))

	var result *ocrDTO.RecognizeCardResult

	// 有 farmerId 时启用图片记录 + OCR缓存
	ossObjectKey := ""
	if oss.Oss != nil {
		ossObjectKey = oss.Oss.BuildKey(objectPath)
	}

	farmerID := parseUint64(input.farmerID)
	if farmerID > 0 {
		wxCloudURL := ""
		lastSource := image_source.OSS
		if image_source.IsWXCloud() && strings.TrimSpace(input.imageURL) != "" {
			wxCloudURL = strings.TrimSpace(input.imageURL)
			lastSource = image_source.WXCloud
		}
		businessID, imgErr := h.resolveBusinessIDWithCloud(farmerID, appUserID, input.cardType, input.imageSide, input.fileName, imageURL, ossObjectKey, wxCloudURL, lastSource)
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
			result, err = h.recognizeCardWithFallback(context, ocrDTO.RecognizeCardRequest{
				CardType:   input.cardType,
				FileName:   input.fileName,
				FileSize:   input.fileSize,
				MimeType:   input.mimeType,
				ImageURL:   imageURL,
				ImageBytes: input.image,
			})
			h.saveOcrResult(businessID, imageURL, result, err)
		}
	}

	// 无 farmerId 或图片记录失败时，直接调用OCR（向下兼容）
	if result == nil && err == nil {
		result, err = h.recognizeCardWithFallback(context, ocrDTO.RecognizeCardRequest{
			CardType:   input.cardType,
			FileName:   input.fileName,
			FileSize:   input.fileSize,
			MimeType:   input.mimeType,
			ImageURL:   imageURL,
			ImageBytes: input.image,
		})
	}

	if result != nil && oss.Oss != nil {
		result.OssBucket = oss.Oss.BucketName
		result.OssObjectKey = oss.Oss.BuildKey(objectPath)
		result.OssURL = imageURL
	}
	if result != nil && image_source.IsWXCloud() && strings.TrimSpace(input.imageURL) != "" {
		result.WXCloudURL = strings.TrimSpace(input.imageURL)
	}
	if err == nil && result != nil && farmerID > 0 {
		err = h.saveFarmerCardInfo(farmerID, stationID, result)
	}
	commonRouter.ToJson(context, result, err)
}

func (h *GrainCardOcrHandler) recognizeCardWithFallback(context *gin.Context, req ocrDTO.RecognizeCardRequest) (*ocrDTO.RecognizeCardResult, error) {
	if strings.TrimSpace(req.ImageURL) != "" {
		urlReq := req
		urlReq.ImageBytes = nil
		log.Printf("[grain-card-ocr] aliyun ocr start: cardType=%s mode=url host=%s", req.CardType, urlHost(req.ImageURL))
		result, err := h.ocrService.RecognizeCard(context.Request.Context(), urlReq)
		if err == nil {
			log.Printf("[grain-card-ocr] aliyun ocr success: cardType=%s mode=url requestId=%s", req.CardType, result.RequestID)
			return result, nil
		}
		log.Printf("[grain-card-ocr] aliyun ocr failed: cardType=%s mode=url host=%s err=%v", req.CardType, urlHost(req.ImageURL), err)
	}

	bytesReq := req
	bytesReq.ImageURL = ""
	log.Printf("[grain-card-ocr] aliyun ocr start: cardType=%s mode=bytes bytes=%d", req.CardType, len(req.ImageBytes))
	result, err := h.ocrService.RecognizeCard(context.Request.Context(), bytesReq)
	if err != nil {
		log.Printf("[grain-card-ocr] aliyun ocr failed: cardType=%s mode=bytes bytes=%d err=%v", req.CardType, len(req.ImageBytes), err)
		return nil, err
	}
	log.Printf("[grain-card-ocr] aliyun ocr success: cardType=%s mode=bytes requestId=%s", req.CardType, result.RequestID)
	return result, nil
}

func readRecognizeCardInput(context *gin.Context) (*recognizeCardInput, error) {
	if strings.HasPrefix(context.ContentType(), "application/json") {
		var req recognizeCardJSONRequest
		if err := context.ShouldBindJSON(&req); err != nil {
			return nil, errors.New("参数错误")
		}
		image, mimeType, err := downloadCloudStorageImage(strings.TrimSpace(req.ImageURL))
		if err != nil {
			return nil, err
		}
		fileName := strings.TrimSpace(req.FileName)
		if fileName == "" {
			fileName = filepath.Base(strings.TrimSpace(req.CloudFileID))
		}
		if fileName == "." || fileName == "/" || fileName == "" {
			fileName = "card.jpg"
		}
		if req.MimeType != "" {
			mimeType = req.MimeType
		}
		fileSize := req.FileSize
		if fileSize <= 0 {
			fileSize = int64(len(image))
		}
		return &recognizeCardInput{
			cardType:    strings.TrimSpace(req.CardType),
			farmerID:    strings.TrimSpace(req.FarmerID),
			imageSide:   strings.TrimSpace(req.ImageSide),
			fileName:    fileName,
			fileSize:    fileSize,
			mimeType:    mimeType,
			imageURL:    strings.TrimSpace(req.ImageURL),
			cloudFileID: strings.TrimSpace(req.CloudFileID),
			image:       image,
		}, nil
	}

	cardType := strings.TrimSpace(context.PostForm("cardType"))
	file, err := context.FormFile("file")
	if err != nil {
		return nil, errors.New("请上传识别照片")
	}
	openFile, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer openFile.Close()
	image, err := io.ReadAll(openFile)
	if err != nil {
		return nil, err
	}
	return &recognizeCardInput{
		cardType:  cardType,
		farmerID:  strings.TrimSpace(context.PostForm("farmerId")),
		imageSide: strings.TrimSpace(context.PostForm("imageSide")),
		fileName:  file.Filename,
		fileSize:  file.Size,
		mimeType:  file.Header.Get("Content-Type"),
		image:     image,
	}, nil
}

func downloadCloudStorageImage(rawURL string) ([]byte, string, error) {
	if rawURL == "" {
		return nil, "", errors.New("imageUrl不能为空")
	}
	if !strings.HasPrefix(rawURL, "https://") && !strings.HasPrefix(rawURL, "http://") {
		return nil, "", errors.New("imageUrl必须是HTTP地址")
	}

	client := &http.Client{Timeout: 20 * time.Second}
	log.Printf("[grain-card-ocr] download cloud image start: host=%s", urlHost(rawURL))
	resp, err := client.Get(rawURL)
	if err != nil {
		log.Printf("[grain-card-ocr] download cloud image failed: host=%s err=%v", urlHost(rawURL), err)
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("[grain-card-ocr] download cloud image bad status: host=%s status=%d contentType=%s contentLength=%d", urlHost(rawURL), resp.StatusCode, resp.Header.Get("Content-Type"), resp.ContentLength)
		return nil, "", fmt.Errorf("下载云存储图片失败：%d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxCloudOcrImageSize+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) == 0 {
		return nil, "", errors.New("云存储图片为空")
	}
	if len(data) > maxCloudOcrImageSize {
		return nil, "", errors.New("图片不能超过10MB")
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "" {
		if parsed, _, err := mime.ParseMediaType(contentType); err == nil {
			contentType = parsed
		}
	}
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	log.Printf("[grain-card-ocr] download cloud image success: host=%s bytes=%d contentType=%s contentLength=%d", urlHost(rawURL), len(data), contentType, resp.ContentLength)
	return data, contentType, nil
}

func urlHost(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	return parsed.Host
}

// resolveBusinessID 查找或创建图片记录，返回业务唯一ID
func (h *GrainCardOcrHandler) resolveBusinessID(farmerID, appUserID uint64, cardType, imageSide, imageName, ossURL, ossObjectKey string) (string, error) {
	return h.resolveBusinessIDWithCloud(farmerID, appUserID, cardType, imageSide, imageName, ossURL, ossObjectKey, "", image_source.OSS)
}

func (h *GrainCardOcrHandler) resolveBusinessIDWithCloud(farmerID, appUserID uint64, cardType, imageSide, imageName, ossURL, ossObjectKey, wxCloudURL, lastSource string) (string, error) {
	if cardType == "id-card" {
		record, err := h.farmerImageService.FindOrCreateIDCardImageWithCloud(farmerID, appUserID, imageSide, imageName, ossURL, ossObjectKey, wxCloudURL, lastSource)
		if err != nil {
			return "", err
		}
		return farmerImageService.IDCardBusinessID(record.Id), nil
	}
	record, err := h.farmerImageService.FindOrCreateBankCardImageWithCloud(farmerID, appUserID, imageName, ossURL, ossObjectKey, wxCloudURL, lastSource)
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
