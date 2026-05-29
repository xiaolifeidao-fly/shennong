package grain_card_ocr

import (
	appCtx "app-api/pkg/internal/appctx"
	commonRouter "common/middleware/routers"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type GrainCardOcrHandler struct {
	*commonRouter.BaseHandler
}

type recognizeResult struct {
	CardType     string `json:"cardType"`
	Mock         bool   `json:"mock"`
	OssBucket    string `json:"ossBucket"`
	OssObjectKey string `json:"ossObjectKey"`
	OssURL       string `json:"ossUrl"`
	FileName     string `json:"fileName"`
	FileSize     int64  `json:"fileSize"`
	MimeType     string `json:"mimeType"`
	Name         string `json:"name,omitempty"`
	IDNumber     string `json:"idNumber,omitempty"`
	Address      string `json:"address,omitempty"`
	BankNumber   string `json:"bankNumber,omitempty"`
	BankName     string `json:"bankName,omitempty"`
}

func NewGrainCardOcrHandler() *GrainCardOcrHandler {
	return &GrainCardOcrHandler{BaseHandler: &commonRouter.BaseHandler{}}
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

	result := mockRecognizeResult(cardType, file.Filename, file.Size, file.Header.Get("Content-Type"), stationID, appUserID)
	commonRouter.ToJson(context, result, nil)
}

func mockRecognizeResult(cardType, fileName string, fileSize int64, mimeType string, stationID, appUserID uint64) recognizeResult {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		ext = ".jpg"
	}
	now := time.Now()
	objectKey := fmt.Sprintf(
		"mock/grain-card-ocr/%d/%d/%s/%s%s",
		stationID,
		appUserID,
		now.Format("20060102"),
		now.Format("150405000000000"),
		ext,
	)
	if strings.TrimSpace(mimeType) == "" {
		mimeType = "image/jpeg"
	}

	result := recognizeResult{
		CardType:     cardType,
		Mock:         true,
		OssBucket:    "mock-shennong-grain",
		OssObjectKey: objectKey,
		OssURL:       "https://mock-oss.local/" + objectKey,
		FileName:     fileName,
		FileSize:     fileSize,
		MimeType:     mimeType,
	}
	if cardType == "bank-card" {
		result.BankNumber = "6228480402564890018"
		result.BankName = "中国农业银行"
		return result
	}

	result.Name = "张三"
	result.IDNumber = "410105199001013215"
	result.Address = "河南省郑州市中原区示例路100号"
	return result
}
