package grain_farmer_image

import (
	appCtx "app-api/pkg/internal/appctx"
	commonRouter "common/middleware/routers"
	"common/middleware/storage/image_source"
	"net/http"
	"net/url"
	farmerImageService "service/farmer_image"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GrainFarmerImageHandler struct {
	*commonRouter.BaseHandler
	farmerImageService *farmerImageService.FarmerImageService
}

func NewGrainFarmerImageHandler() *GrainFarmerImageHandler {
	svc := farmerImageService.NewFarmerImageService()
	_ = svc.EnsureTable()
	return &GrainFarmerImageHandler{
		BaseHandler:        &commonRouter.BaseHandler{},
		farmerImageService: svc,
	}
}

func (h *GrainFarmerImageHandler) RegisterHandler(engine *gin.RouterGroup) {
	engine.GET("/grain-farmer-images", h.getFarmerImages)
	engine.POST("/grain-farmer-images", h.saveFarmerImage)
}

type saveFarmerImageRequest struct {
	FarmerID     uint64 `json:"farmerId"`
	CardType     string `json:"cardType"`
	ImageSide    string `json:"imageSide"`
	ImageName    string `json:"imageName"`
	OssURL       string `json:"ossUrl"`
	OssObjectKey string `json:"ossObjectKey"`
	WXCloudURL   string `json:"wxCloudUrl"`
	ImageURL     string `json:"imageUrl"`
}

func (h *GrainFarmerImageHandler) getFarmerImages(context *gin.Context) {
	farmerIDStr := strings.TrimSpace(context.Query("farmerId"))
	if farmerIDStr == "" {
		commonRouter.ToError(context, "farmerId不能为空")
		return
	}
	farmerID, err := strconv.ParseUint(farmerIDStr, 10, 64)
	if err != nil || farmerID == 0 {
		commonRouter.ToError(context, "farmerId必须是正整数")
		return
	}
	appUserID, _ := appCtx.CurrentAppUserID(context)
	if imageType := strings.TrimSpace(context.Query("imageType")); imageType != "" {
		h.streamFarmerImage(context, farmerID, appUserID, imageType)
		return
	}
	result := &farmerImageService.FarmerImagesResult{}
	if h.farmerImageService.HasLatestFarmerImage(farmerID, appUserID, "id-card-front") {
		result.IDCardFront = farmerImagePath(farmerID, "id-card-front")
		if image_source.IsWXCloud() {
			result.IDCardFront = h.farmerImageService.LatestFarmerImageWXCloudURL(farmerID, appUserID, "id-card-front", result.IDCardFront)
		}
	}
	if h.farmerImageService.HasLatestFarmerImage(farmerID, appUserID, "id-card-back") {
		result.IDCardBack = farmerImagePath(farmerID, "id-card-back")
		if image_source.IsWXCloud() {
			result.IDCardBack = h.farmerImageService.LatestFarmerImageWXCloudURL(farmerID, appUserID, "id-card-back", result.IDCardBack)
		}
	}
	if h.farmerImageService.HasLatestFarmerImage(farmerID, appUserID, "bank-card") {
		result.BankCard = farmerImagePath(farmerID, "bank-card")
		if image_source.IsWXCloud() {
			result.BankCard = h.farmerImageService.LatestFarmerImageWXCloudURL(farmerID, appUserID, "bank-card", result.BankCard)
		}
	}
	commonRouter.ToJson(context, result, nil)
}

func (h *GrainFarmerImageHandler) streamFarmerImage(context *gin.Context, farmerID, appUserID uint64, imageType string) {
	content, err := h.farmerImageService.GetLatestFarmerImageContent(farmerID, appUserID, imageType)
	if err == gorm.ErrRecordNotFound {
		context.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		commonRouter.ToError(context, err.Error())
		return
	}
	if content.Base64 != "" {
		commonRouter.ToJson(context, imageBase64Response{
			Mode:     "base64",
			MimeType: content.MimeType,
			FileName: content.FileName,
			Base64:   content.Base64,
			DataURL:  "data:" + content.MimeType + ";base64," + content.Base64,
		}, nil)
		return
	}
	context.Header("Cache-Control", "private, max-age=300")
	context.Data(http.StatusOK, content.MimeType, content.Data)
}

type imageBase64Response struct {
	Mode     string `json:"mode"`
	MimeType string `json:"mimeType"`
	FileName string `json:"fileName"`
	Base64   string `json:"base64"`
	DataURL  string `json:"dataUrl"`
}

func farmerImagePath(farmerID uint64, imageType string) string {
	return "/grain-farmer-images?farmerId=" + strconv.FormatUint(farmerID, 10) + "&imageType=" + url.QueryEscape(imageType)
}

func (h *GrainFarmerImageHandler) saveFarmerImage(context *gin.Context) {
	var req saveFarmerImageRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		commonRouter.ToError(context, "参数错误")
		return
	}
	if req.FarmerID == 0 {
		commonRouter.ToError(context, "farmerId不能为空")
		return
	}
	if req.CardType != "id-card" && req.CardType != "bank-card" {
		commonRouter.ToError(context, "cardType必须是id-card或bank-card")
		return
	}
	if req.CardType == "id-card" && req.ImageSide != "front" && req.ImageSide != "back" {
		commonRouter.ToError(context, "imageSide必须是front或back")
		return
	}
	if strings.TrimSpace(req.OssURL) == "" && strings.TrimSpace(req.OssObjectKey) == "" {
		commonRouter.ToError(context, "ossUrl或ossObjectKey不能为空")
		return
	}

	appUserID, _ := appCtx.CurrentAppUserID(context)
	imageName := strings.TrimSpace(req.ImageName)
	if imageName == "" {
		imageName = req.OssObjectKey
	}
	if imageName == "" {
		imageName = req.OssURL
	}
	wxCloudURL := strings.TrimSpace(req.WXCloudURL)
	if wxCloudURL == "" {
		wxCloudURL = strings.TrimSpace(req.ImageURL)
	}
	lastSource := image_source.OSS
	if wxCloudURL != "" && image_source.IsWXCloud() {
		lastSource = image_source.WXCloud
	}

	var result any
	var err error
	if req.CardType == "id-card" {
		imageSide := req.ImageSide
		if imageSide == "" {
			imageSide = "front"
		}
		result, err = h.farmerImageService.FindOrCreateIDCardImageWithCloud(req.FarmerID, appUserID, imageSide, imageName, req.OssURL, req.OssObjectKey, wxCloudURL, lastSource)
	} else {
		result, err = h.farmerImageService.FindOrCreateBankCardImageWithCloud(req.FarmerID, appUserID, imageName, req.OssURL, req.OssObjectKey, wxCloudURL, lastSource)
	}
	commonRouter.ToJson(context, result, err)
}
