package grain_farmer_image

import (
	appCtx "app-api/pkg/internal/appctx"
	commonRouter "common/middleware/routers"
	"common/middleware/storage/image_source"
	farmerImageService "service/farmer_image"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
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
		imageURL := h.farmerImageService.LatestFarmerImageURL(farmerID, appUserID, imageType)
		if imageURL == "" {
			commonRouter.ToError(context, "image not found")
			return
		}
		commonRouter.ToJson(context, gin.H{"imageUrl": imageURL}, nil)
		return
	}
	result := h.farmerImageService.GetLatestFarmerImages(farmerID, appUserID)
	commonRouter.ToJson(context, result, nil)
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
