package repository

import (
	"common/middleware/db"
	"fmt"
)

type RegionRepository struct {
	db.Repository[*Region]
}

func (r *RegionRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&Region{})
}

func (r *RegionRepository) CountActive() (int64, error) {
	if r.Db == nil {
		return 0, fmt.Errorf("database is not initialized")
	}
	var total int64
	return total, r.Db.Model(&Region{}).Where("active = ?", 1).Count(&total).Error
}

func (r *RegionRepository) BatchCreate(entities []*Region) error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	if len(entities) == 0 {
		return nil
	}
	return r.Db.CreateInBatches(entities, 300).Error
}

func (r *RegionRepository) ListActive() ([]*Region, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var entities []*Region
	err := r.Db.Where("active = ? AND status = ?", 1, "active").
		Order("level ASC, sort_order ASC, id ASC").
		Find(&entities).Error
	return entities, err
}
