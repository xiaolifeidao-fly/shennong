package repository

import (
	"common/middleware/db"
	"fmt"
	grainFarmerDTO "service/grain_farmer/dto"
	"strings"

	"gorm.io/gorm"
)

type GrainFarmerRepository struct {
	db.Repository[*GrainFarmer]
}

func (r *GrainFarmerRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&GrainFarmer{})
}

func (r *GrainFarmerRepository) CountByQuery(query grainFarmerDTO.GrainFarmerQueryDTO) (int64, error) {
	dbQuery := applyFarmerQuery(r.Db.Model(&GrainFarmer{}).Where("active = ?", 1), query)
	var total int64
	return total, dbQuery.Count(&total).Error
}

func (r *GrainFarmerRepository) ListByQuery(query grainFarmerDTO.GrainFarmerQueryDTO, pageIndex, pageSize int) ([]*GrainFarmer, error) {
	dbQuery := applyFarmerQuery(r.Db.Model(&GrainFarmer{}).Where("active = ?", 1), query)
	var entities []*GrainFarmer
	err := dbQuery.Order("id DESC").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&entities).Error
	return entities, err
}

func (r *GrainFarmerRepository) FindActiveByIDNumberDigest(idNumberDigest, plainIDNumber string, stationID uint64) (*GrainFarmer, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var entity GrainFarmer
	dbQuery := r.Db.Where("active = ?", 1)
	if stationID > 0 {
		dbQuery = dbQuery.Where("station_id = ?", stationID)
	}
	idNumberDigest = strings.TrimSpace(idNumberDigest)
	plainIDNumber = strings.TrimSpace(plainIDNumber)
	if idNumberDigest != "" && plainIDNumber != "" {
		dbQuery = dbQuery.Where("(id_number_digest = ? OR id_number = ?)", idNumberDigest, plainIDNumber)
	} else if idNumberDigest != "" {
		dbQuery = dbQuery.Where("id_number_digest = ?", idNumberDigest)
	} else if plainIDNumber != "" {
		dbQuery = dbQuery.Where("id_number = ?", plainIDNumber)
	} else {
		return nil, gorm.ErrRecordNotFound
	}
	err := dbQuery.Order("id DESC").First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func applyFarmerQuery(dbQuery *gorm.DB, query grainFarmerDTO.GrainFarmerQueryDTO) *gorm.DB {
	if query.StationID > 0 {
		dbQuery = dbQuery.Where("station_id = ?", query.StationID)
	}
	if len(query.StationIDs) > 0 {
		dbQuery = dbQuery.Where("station_id IN ?", query.StationIDs)
	}
	if query.AppUserID > 0 {
		dbQuery = dbQuery.Where("app_user_id = ?", query.AppUserID)
	}
	if value := strings.TrimSpace(query.Search); value != "" {
		likeValue := "%" + value + "%"
		dbQuery = dbQuery.Where("(id_number_digest = ? OR phone LIKE ? OR address LIKE ?)", query.SearchIDNumberDigest, likeValue, likeValue)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		dbQuery = dbQuery.Where("status = ?", value)
	}
	return dbQuery
}
