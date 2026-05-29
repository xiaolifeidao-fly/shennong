package repository

import (
	"common/middleware/db"
	"fmt"
)

type FarmerIDCardImageRepository struct {
	db.Repository[*FarmerIDCardImage]
}

func (r *FarmerIDCardImageRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&FarmerIDCardImage{})
}

// FindByUnique 按唯一条件查找
func (r *FarmerIDCardImageRepository) FindByUnique(farmerID, appUserID uint64, imageHash string) (*FarmerIDCardImage, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var entity FarmerIDCardImage
	err := r.Db.Where("farmer_id = ? AND app_user_id = ? AND image_hash = ? AND active = 1", farmerID, appUserID, imageHash).
		First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindOrCreate 存在则返回，否则创建
func (r *FarmerIDCardImageRepository) FindOrCreate(entity *FarmerIDCardImage) (*FarmerIDCardImage, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	existing, err := r.FindByUnique(entity.FarmerID, entity.AppUserID, entity.ImageHash)
	if err == nil {
		return existing, nil
	}
	if err2 := r.Db.Create(entity).Error; err2 != nil {
		return nil, err2
	}
	return entity, nil
}

// ——————————————————————————————————————————————

type FarmerBankCardImageRepository struct {
	db.Repository[*FarmerBankCardImage]
}

func (r *FarmerBankCardImageRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&FarmerBankCardImage{})
}

func (r *FarmerBankCardImageRepository) FindByUnique(farmerID, appUserID uint64, imageHash string) (*FarmerBankCardImage, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var entity FarmerBankCardImage
	err := r.Db.Where("farmer_id = ? AND app_user_id = ? AND image_hash = ? AND active = 1", farmerID, appUserID, imageHash).
		First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *FarmerBankCardImageRepository) FindOrCreate(entity *FarmerBankCardImage) (*FarmerBankCardImage, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	existing, err := r.FindByUnique(entity.FarmerID, entity.AppUserID, entity.ImageHash)
	if err == nil {
		return existing, nil
	}
	if err2 := r.Db.Create(entity).Error; err2 != nil {
		return nil, err2
	}
	return entity, nil
}
