package farmer_image

import (
	"common/middleware/db"
	"crypto/sha256"
	"fmt"
	farmerImageRepository "service/farmer_image/repository"
)

type FarmerImageService struct {
	idcardRepo  *farmerImageRepository.FarmerIDCardImageRepository
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

// HashImageName 计算文件名SHA-256
func HashImageName(imageName string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(imageName)))
}

// FindOrCreateIDCardImage 查找或创建身份证图片记录，返回记录ID
func (s *FarmerImageService) FindOrCreateIDCardImage(farmerID, appUserID uint64, imageSide, imageName, ossURL string) (*farmerImageRepository.FarmerIDCardImage, error) {
	imageHash := HashImageName(imageName)
	entity := &farmerImageRepository.FarmerIDCardImage{
		FarmerID:  farmerID,
		AppUserID: appUserID,
		ImageSide: imageSide,
		ImageName: imageName,
		ImageHash: imageHash,
		OssURL:    ossURL,
	}
	return s.idcardRepo.FindOrCreate(entity)
}

// FindOrCreateBankCardImage 查找或创建银行卡图片记录
func (s *FarmerImageService) FindOrCreateBankCardImage(farmerID, appUserID uint64, imageName, ossURL string) (*farmerImageRepository.FarmerBankCardImage, error) {
	imageHash := HashImageName(imageName)
	entity := &farmerImageRepository.FarmerBankCardImage{
		FarmerID:  farmerID,
		AppUserID: appUserID,
		ImageName: imageName,
		ImageHash: imageHash,
		OssURL:    ossURL,
	}
	return s.bankcardRepo.FindOrCreate(entity)
}

// IDCardBusinessID 生成身份证OCR业务唯一ID
func IDCardBusinessID(imageID int) string {
	return fmt.Sprintf("idcard_%d", imageID)
}

// BankCardBusinessID 生成银行卡OCR业务唯一ID
func BankCardBusinessID(imageID int) string {
	return fmt.Sprintf("bank_%d", imageID)
}
