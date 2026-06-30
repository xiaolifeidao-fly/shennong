package farmer_image

import (
	"common/middleware/db"
	"common/middleware/storage/image_source"
	"common/middleware/storage/oss"
	"crypto/sha256"
	"fmt"
	"log"
	"net/url"
	farmerImageRepository "service/farmer_image/repository"
	"strings"
	"time"
)

// FarmerImagesResult 农户证件图片汇总
type FarmerImagesResult struct {
	IDCardFront           string `json:"idCardFront"`
	IDCardBack            string `json:"idCardBack"`
	BankCard              string `json:"bankCard"`
	IDCardFrontWXCloudURL string `json:"idCardFrontWxCloudUrl,omitempty"`
	IDCardBackWXCloudURL  string `json:"idCardBackWxCloudUrl,omitempty"`
	BankCardWXCloudURL    string `json:"bankCardWxCloudUrl,omitempty"`
}

type FarmerImageService struct {
	idcardRepo   *farmerImageRepository.FarmerIDCardImageRepository
	bankcardRepo *farmerImageRepository.FarmerBankCardImageRepository
}

func NewFarmerImageService() *FarmerImageService {
	return &FarmerImageService{
		idcardRepo:   db.GetRepository[farmerImageRepository.FarmerIDCardImageRepository](),
		bankcardRepo: db.GetRepository[farmerImageRepository.FarmerBankCardImageRepository](),
	}
}

func (s *FarmerImageService) EnsureTable() error {
	if err := s.idcardRepo.EnsureTable(); err != nil {
		return err
	}
	return s.bankcardRepo.EnsureTable()
}

// HashImageIdentity 计算图片唯一标识，优先使用OSS对象Key，避免小程序临时文件名重复。
func HashImageIdentity(imageName, ossObjectKey string) string {
	identity := ossObjectKey
	if identity == "" {
		identity = imageName
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(identity)))
}

// FindOrCreateIDCardImage 查找或创建身份证图片记录，返回记录ID
func (s *FarmerImageService) FindOrCreateIDCardImage(farmerID, appUserID uint64, imageSide, imageName, ossURL, ossObjectKey string) (*farmerImageRepository.FarmerIDCardImage, error) {
	return s.FindOrCreateIDCardImageWithCloud(farmerID, appUserID, imageSide, imageName, ossURL, ossObjectKey, "", image_source.OSS)
}

func (s *FarmerImageService) FindOrCreateIDCardImageWithCloud(farmerID, appUserID uint64, imageSide, imageName, ossURL, ossObjectKey, wxCloudURL, lastSource string) (*farmerImageRepository.FarmerIDCardImage, error) {
	imageHash := HashImageIdentity(imageName, ossObjectKey)
	entity := &farmerImageRepository.FarmerIDCardImage{
		FarmerID:     farmerID,
		AppUserID:    appUserID,
		ImageSide:    imageSide,
		ImageName:    imageName,
		ImageHash:    imageHash,
		OssURL:       ossURL,
		OssObjectKey: ossObjectKey,
		LastSource:   normalizeImageSource(lastSource),
		WXCloudURL:   strings.TrimSpace(wxCloudURL),
	}
	result, err := s.idcardRepo.FindOrCreate(entity)
	if err == nil && result != nil && entity.WXCloudURL != "" && (result.WXCloudURL != entity.WXCloudURL || result.LastSource != entity.LastSource) {
		if updateErr := s.idcardRepo.UpdateCloudSource(result.Id, entity.WXCloudURL, entity.LastSource); updateErr == nil {
			result.WXCloudURL = entity.WXCloudURL
			result.LastSource = entity.LastSource
		}
	}
	return result, err
}

// FindOrCreateBankCardImage 查找或创建银行卡图片记录
func (s *FarmerImageService) FindOrCreateBankCardImage(farmerID, appUserID uint64, imageName, ossURL, ossObjectKey string) (*farmerImageRepository.FarmerBankCardImage, error) {
	return s.FindOrCreateBankCardImageWithCloud(farmerID, appUserID, imageName, ossURL, ossObjectKey, "", image_source.OSS)
}

func (s *FarmerImageService) FindOrCreateBankCardImageWithCloud(farmerID, appUserID uint64, imageName, ossURL, ossObjectKey, wxCloudURL, lastSource string) (*farmerImageRepository.FarmerBankCardImage, error) {
	imageHash := HashImageIdentity(imageName, ossObjectKey)
	entity := &farmerImageRepository.FarmerBankCardImage{
		FarmerID:     farmerID,
		AppUserID:    appUserID,
		ImageName:    imageName,
		ImageHash:    imageHash,
		OssURL:       ossURL,
		OssObjectKey: ossObjectKey,
		LastSource:   normalizeImageSource(lastSource),
		WXCloudURL:   strings.TrimSpace(wxCloudURL),
	}
	result, err := s.bankcardRepo.FindOrCreate(entity)
	if err == nil && result != nil && entity.WXCloudURL != "" && (result.WXCloudURL != entity.WXCloudURL || result.LastSource != entity.LastSource) {
		if updateErr := s.bankcardRepo.UpdateCloudSource(result.Id, entity.WXCloudURL, entity.LastSource); updateErr == nil {
			result.WXCloudURL = entity.WXCloudURL
			result.LastSource = entity.LastSource
		}
	}
	return result, err
}

// GetLatestFarmerImages 获取农户最新证件图片（生成新签名URL）
func (s *FarmerImageService) GetLatestFarmerImages(farmerID, appUserID uint64) *FarmerImagesResult {
	result := &FarmerImagesResult{}
	urlExpiry := 30 * time.Minute

	if front, err := s.idcardRepo.FindLatestBySide(farmerID, appUserID, "front"); err == nil {
		result.IDCardFront = refreshOssURL(front.OssObjectKey, front.OssURL, urlExpiry)
	}
	if back, err := s.idcardRepo.FindLatestBySide(farmerID, appUserID, "back"); err == nil {
		result.IDCardBack = refreshOssURL(back.OssObjectKey, back.OssURL, urlExpiry)
	}
	if bank, err := s.bankcardRepo.FindLatest(farmerID, appUserID); err == nil {
		result.BankCard = refreshOssURL(bank.OssObjectKey, bank.OssURL, urlExpiry)
	}
	return result
}

func (s *FarmerImageService) HasLatestFarmerImage(farmerID, appUserID uint64, imageType string) bool {
	_, err := s.findLatestImageRecord(farmerID, appUserID, imageType)
	return err == nil
}

func (s *FarmerImageService) LatestFarmerImageURL(farmerID, appUserID uint64, imageType string) string {
	record, err := s.findLatestImageRecord(farmerID, appUserID, imageType)
	if err != nil || record == nil {
		return ""
	}
	return refreshOssURL(record.ossObjectKey, record.ossURL, 30*time.Minute)
}

func (s *FarmerImageService) LatestFarmerImageWXCloudURL(farmerID, appUserID uint64, imageType, fallbackURL string) string {
	record, err := s.findLatestImageRecord(farmerID, appUserID, imageType)
	if err != nil || record == nil {
		return ""
	}
	return fallbackURL
}

type latestFarmerImageRecord struct {
	imageName         string
	ossObjectKey      string
	ossURL            string
	wxCloudURL        string
	lastSource        string
	updateCloudSource func(wxCloudURL, lastSource string) error
}

func (s *FarmerImageService) findLatestImageRecord(farmerID, appUserID uint64, imageType string) (*latestFarmerImageRecord, error) {
	switch strings.TrimSpace(imageType) {
	case "id-card-front", "front":
		image, err := s.idcardRepo.FindLatestBySide(farmerID, appUserID, "front")
		if err != nil {
			return nil, err
		}
		return &latestFarmerImageRecord{
			imageName: image.ImageName, ossObjectKey: image.OssObjectKey, ossURL: image.OssURL,
			wxCloudURL: image.WXCloudURL, lastSource: image.LastSource,
			updateCloudSource: func(wxCloudURL, lastSource string) error {
				return s.idcardRepo.UpdateCloudSource(image.Id, wxCloudURL, lastSource)
			},
		}, nil
	case "id-card-back", "back":
		image, err := s.idcardRepo.FindLatestBySide(farmerID, appUserID, "back")
		if err != nil {
			return nil, err
		}
		return &latestFarmerImageRecord{
			imageName: image.ImageName, ossObjectKey: image.OssObjectKey, ossURL: image.OssURL,
			wxCloudURL: image.WXCloudURL, lastSource: image.LastSource,
			updateCloudSource: func(wxCloudURL, lastSource string) error {
				return s.idcardRepo.UpdateCloudSource(image.Id, wxCloudURL, lastSource)
			},
		}, nil
	case "bank-card", "bank":
		image, err := s.bankcardRepo.FindLatest(farmerID, appUserID)
		if err != nil {
			return nil, err
		}
		return &latestFarmerImageRecord{
			imageName: image.ImageName, ossObjectKey: image.OssObjectKey, ossURL: image.OssURL,
			wxCloudURL: image.WXCloudURL, lastSource: image.LastSource,
			updateCloudSource: func(wxCloudURL, lastSource string) error {
				return s.bankcardRepo.UpdateCloudSource(image.Id, wxCloudURL, lastSource)
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported image type")
	}
}

func normalizeImageSource(source string) string {
	if strings.EqualFold(strings.TrimSpace(source), image_source.WXCloud) {
		return image_source.WXCloud
	}
	return image_source.OSS
}

func refreshOssURL(ossObjectKey, fallbackURL string, expiry time.Duration) string {
	if ossObjectKey != "" {
		if oss.Oss != nil {
			if url, err := oss.Oss.GetUrlByKey(ossObjectKey, &expiry); err == nil {
				log.Printf("[farmer-image] refresh oss url by key success ossObjectKey=%s expiry=%s", ossObjectKey, expiry)
				return url
			} else {
				log.Printf("[farmer-image] refresh oss url by key failed ossObjectKey=%s expiry=%s err=%v", ossObjectKey, expiry, err)
			}
		}
		if url, err := oss.GetUrl(ossObjectKey, &expiry); err == nil {
			log.Printf("[farmer-image] refresh oss url by path success ossObjectKey=%s expiry=%s", ossObjectKey, expiry)
			return url
		} else {
			log.Printf("[farmer-image] refresh oss url by path failed ossObjectKey=%s expiry=%s err=%v", ossObjectKey, expiry, err)
		}
	}
	log.Printf("[farmer-image] use fallback oss url ossObjectKey=%s fallbackURL=%s", ossObjectKey, safeURLForLog(fallbackURL))
	return fallbackURL
}

func objectKeyFromURL(rawURL string) string {
	value := strings.TrimSpace(rawURL)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Path == "" {
		return ""
	}
	return strings.TrimLeft(parsed.Path, "/")
}

func safeURLForLog(rawURL string) string {
	value := strings.TrimSpace(rawURL)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "<invalid-url>"
	}
	hasQuery := parsed.RawQuery != ""
	if parsed.Scheme == "" && parsed.Host == "" {
		return fmt.Sprintf("path=%s hasQuery=%t", parsed.Path, hasQuery)
	}
	return fmt.Sprintf("scheme=%s host=%s path=%s hasQuery=%t", parsed.Scheme, parsed.Host, parsed.Path, hasQuery)
}

// IDCardBusinessID 生成身份证OCR业务唯一ID
func IDCardBusinessID(imageID int) string {
	return fmt.Sprintf("idcard_%d", imageID)
}

// BankCardBusinessID 生成银行卡OCR业务唯一ID
func BankCardBusinessID(imageID int) string {
	return fmt.Sprintf("bank_%d", imageID)
}
