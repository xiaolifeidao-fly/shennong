package repository

import (
	"common/middleware/db"
	"fmt"
)

type OcrResultRepository struct {
	db.Repository[*OcrResult]
}

func (r *OcrResultRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&OcrResult{})
}

func (r *OcrResultRepository) FindByBusinessID(businessID string) (*OcrResult, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var entity OcrResult
	err := r.Db.Where("business_id = ? AND active = 1", businessID).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Upsert 存在则更新，不存在则创建
func (r *OcrResultRepository) Upsert(entity *OcrResult) error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	existing, err := r.FindByBusinessID(entity.BusinessID)
	if err == nil {
		return r.Db.Model(existing).Updates(map[string]interface{}{
			"oss_url":     entity.OssURL,
			"result_json": entity.ResultJSON,
			"status":      entity.Status,
		}).Error
	}
	return r.Db.Create(entity).Error
}
