package farmer_image

import (
	"common/middleware/db"
	"common/middleware/storage/oss"
	"crypto/sha256"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	farmerImageRepository "service/farmer_image/repository"
	"strings"
	"time"
)

// FarmerImagesResult 农户证件图片汇总
type FarmerImagesResult struct {
	IDCardFront string `json:"idCardFront"`
	IDCardBack  string `json:"idCardBack"`
	BankCard    string `json:"bankCard"`
}

type FarmerImageContent struct {
	Data     []byte
	MimeType string
	FileName string
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
	imageHash := HashImageIdentity(imageName, ossObjectKey)
	entity := &farmerImageRepository.FarmerIDCardImage{
		FarmerID:     farmerID,
		AppUserID:    appUserID,
		ImageSide:    imageSide,
		ImageName:    imageName,
		ImageHash:    imageHash,
		OssURL:       ossURL,
		OssObjectKey: ossObjectKey,
	}
	return s.idcardRepo.FindOrCreate(entity)
}

// FindOrCreateBankCardImage 查找或创建银行卡图片记录
func (s *FarmerImageService) FindOrCreateBankCardImage(farmerID, appUserID uint64, imageName, ossURL, ossObjectKey string) (*farmerImageRepository.FarmerBankCardImage, error) {
	imageHash := HashImageIdentity(imageName, ossObjectKey)
	entity := &farmerImageRepository.FarmerBankCardImage{
		FarmerID:     farmerID,
		AppUserID:    appUserID,
		ImageName:    imageName,
		ImageHash:    imageHash,
		OssURL:       ossURL,
		OssObjectKey: ossObjectKey,
	}
	return s.bankcardRepo.FindOrCreate(entity)
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
	_, _, _, err := s.findLatestImageRecord(farmerID, appUserID, imageType)
	return err == nil
}

func (s *FarmerImageService) GetLatestFarmerImageContent(farmerID, appUserID uint64, imageType string) (*FarmerImageContent, error) {
	log.Printf("[farmer-image] get latest image content start farmerID=%d appUserID=%d imageType=%s", farmerID, appUserID, imageType)
	imageName, objectKey, ossURL, err := s.findLatestImageRecord(farmerID, appUserID, imageType)
	if err != nil {
		log.Printf("[farmer-image] find latest image record failed farmerID=%d appUserID=%d imageType=%s err=%v", farmerID, appUserID, imageType, err)
		return nil, err
	}
	log.Printf("[farmer-image] latest image record found farmerID=%d appUserID=%d imageType=%s fileName=%s ossObjectKey=%s fallbackURL=%s", farmerID, appUserID, imageType, imageName, objectKey, safeURLForLog(ossURL))
	data, err := getOssObject(objectKey, ossURL)
	if err != nil {
		log.Printf("[farmer-image] get oss object failed farmerID=%d appUserID=%d imageType=%s fileName=%s ossObjectKey=%s fallbackURL=%s err=%v", farmerID, appUserID, imageType, imageName, objectKey, safeURLForLog(ossURL), err)
		return nil, err
	}
	mimeType := detectImageMimeType(data, imageName)
	log.Printf("[farmer-image] get latest image content success farmerID=%d appUserID=%d imageType=%s fileName=%s mimeType=%s bytes=%d", farmerID, appUserID, imageType, imageName, mimeType, len(data))
	return &FarmerImageContent{
		Data:     data,
		MimeType: mimeType,
		FileName: imageName,
	}, nil
}

func (s *FarmerImageService) findLatestImageRecord(farmerID, appUserID uint64, imageType string) (string, string, string, error) {
	switch strings.TrimSpace(imageType) {
	case "id-card-front", "front":
		image, err := s.idcardRepo.FindLatestBySide(farmerID, appUserID, "front")
		if err != nil {
			return "", "", "", err
		}
		return image.ImageName, image.OssObjectKey, image.OssURL, nil
	case "id-card-back", "back":
		image, err := s.idcardRepo.FindLatestBySide(farmerID, appUserID, "back")
		if err != nil {
			return "", "", "", err
		}
		return image.ImageName, image.OssObjectKey, image.OssURL, nil
	case "bank-card", "bank":
		image, err := s.bankcardRepo.FindLatest(farmerID, appUserID)
		if err != nil {
			return "", "", "", err
		}
		return image.ImageName, image.OssObjectKey, image.OssURL, nil
	default:
		return "", "", "", fmt.Errorf("unsupported image type")
	}
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

func getOssObject(ossObjectKey, fallbackURL string) ([]byte, error) {
	key := strings.TrimSpace(ossObjectKey)
	if key == "" {
		key = objectKeyFromURL(fallbackURL)
		log.Printf("[farmer-image] oss object key empty, parsed key from fallbackURL parsedKey=%s fallbackURL=%s", key, safeURLForLog(fallbackURL))
	}
	if key == "" {
		return nil, fmt.Errorf("oss object key is empty")
	}
	if data, err := oss.GetByKey(key); err == nil {
		log.Printf("[farmer-image] oss get by key success key=%s bytes=%d", key, len(data))
		return data, nil
	} else {
		log.Printf("[farmer-image] oss get by key failed key=%s err=%v", key, err)
	}
	data, err := oss.Get(key)
	if err != nil {
		log.Printf("[farmer-image] oss get by path failed path=%s err=%v", key, err)
		return nil, err
	}
	log.Printf("[farmer-image] oss get by path success path=%s bytes=%d", key, len(data))
	return data, nil
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

func detectImageMimeType(data []byte, fileName string) string {
	if ext := strings.ToLower(filepath.Ext(fileName)); ext != "" {
		if mimeType := mime.TypeByExtension(ext); strings.HasPrefix(mimeType, "image/") {
			return mimeType
		}
	}
	if len(data) > 0 {
		return http.DetectContentType(data)
	}
	return "application/octet-stream"
}

// IDCardBusinessID 生成身份证OCR业务唯一ID
func IDCardBusinessID(imageID int) string {
	return fmt.Sprintf("idcard_%d", imageID)
}

// BankCardBusinessID 生成银行卡OCR业务唯一ID
func BankCardBusinessID(imageID int) string {
	return fmt.Sprintf("bank_%d", imageID)
}
