package farmer_image

import (
	"common/middleware/db"
	"common/middleware/storage/oss"
	"crypto/sha256"
	"fmt"
	farmerImageRepository "service/farmer_image/repository"
	"time"
)

// FarmerImagesResult 农户证件图片汇总
type FarmerImagesResult struct {
	IDCardFront string `json:"idCardFront"`
	IDCardBack  string `json:"idCardBack"`
	BankCard    string `json:"bankCard"`
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

func refreshOssURL(ossObjectKey, fallbackURL string, expiry time.Duration) string {
	if ossObjectKey != "" {
		if oss.Oss != nil {
			if url, err := oss.Oss.GetUrlByKey(ossObjectKey, &expiry); err == nil {
				return url
			}
		}
		if url, err := oss.GetUrl(ossObjectKey, &expiry); err == nil {
			return url
		}
	}
	return fallbackURL
}

// IDCardBusinessID 生成身份证OCR业务唯一ID
func IDCardBusinessID(imageID int) string {
	return fmt.Sprintf("idcard_%d", imageID)
}

// BankCardBusinessID 生成银行卡OCR业务唯一ID
func BankCardBusinessID(imageID int) string {
	return fmt.Sprintf("bank_%d", imageID)
}
