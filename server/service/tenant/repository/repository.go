package repository

import (
	"common/middleware/db"
	"fmt"
	tenantDTO "service/tenant/dto"
	"strings"

	"gorm.io/gorm"
)

type TenantRepository struct {
	db.Repository[*Tenant]
}

func (r *TenantRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&Tenant{})
}

func (r *TenantRepository) CountByQuery(query tenantDTO.TenantQueryDTO) (int64, error) {
	dbQuery := r.Db.Model(&Tenant{}).Where("active = ?", 1)
	dbQuery = applyTenantQuery(dbQuery, query)
	var total int64
	return total, dbQuery.Count(&total).Error
}

func (r *TenantRepository) ListByQuery(query tenantDTO.TenantQueryDTO, pageIndex, pageSize int) ([]*Tenant, error) {
	dbQuery := r.Db.Model(&Tenant{}).Where("active = ?", 1)
	dbQuery = applyTenantQuery(dbQuery, query)
	var entities []*Tenant
	err := dbQuery.Order("id DESC").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&entities).Error
	return entities, err
}

func (r *TenantRepository) FindByCode(code string) (*Tenant, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var entity Tenant
	err := r.Db.Where("active = ? AND tenant_code = ?", 1, strings.TrimSpace(code)).First(&entity).Error
	return &entity, err
}

type TenantStationRepository struct {
	db.Repository[*TenantStation]
}

func (r *TenantStationRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&TenantStation{})
}

func (r *TenantStationRepository) ReplaceActiveStations(tenantID uint64, stationIDs []uint64) error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	if tenantID == 0 {
		return fmt.Errorf("tenantId must be positive")
	}
	if err := r.Db.Model(&TenantStation{}).
		Where("active = ? AND tenant_id = ?", 1, tenantID).
		Updates(map[string]interface{}{"active": 0}).Error; err != nil {
		return err
	}
	for _, stationID := range uniquePositiveUint64(stationIDs) {
		if _, err := r.Create(&TenantStation{TenantID: tenantID, StationID: stationID, Status: "active"}); err != nil {
			return err
		}
	}
	return nil
}

func (r *TenantStationRepository) ListActiveStationIDs(tenantID uint64) ([]uint64, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var rows []struct {
		StationID uint64
	}
	err := r.Db.Raw(`
		SELECT station_id FROM tenant_station
		WHERE active = ? AND status = ? AND tenant_id = ?
		UNION
		SELECT id AS station_id FROM grain_station
		WHERE active = ? AND tenant_id = ?`, 1, "active", tenantID, 1, tenantID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]uint64, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.StationID)
	}
	return result, nil
}

func applyTenantQuery(dbQuery *gorm.DB, query tenantDTO.TenantQueryDTO) *gorm.DB {
	if query.TenantID > 0 {
		dbQuery = dbQuery.Where("id = ?", query.TenantID)
	}
	if value := strings.TrimSpace(query.Search); value != "" {
		likeValue := "%" + value + "%"
		dbQuery = dbQuery.Where("(tenant_name LIKE ? OR tenant_code LIKE ? OR contact_name LIKE ? OR contact_phone LIKE ?)", likeValue, likeValue, likeValue, likeValue)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		dbQuery = dbQuery.Where("status = ?", value)
	}
	return dbQuery
}

func uniquePositiveUint64(values []uint64) []uint64 {
	seen := make(map[uint64]struct{}, len(values))
	result := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
