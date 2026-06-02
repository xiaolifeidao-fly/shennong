package repository

import (
	"common/middleware/db"
	"fmt"
	grainConfigDTO "service/grain_config/dto"
	"strings"

	"gorm.io/gorm"
)

type GrainStationRepository struct {
	db.Repository[*GrainStation]
}

func (r *GrainStationRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&GrainStation{})
}

func (r *GrainStationRepository) CountByQuery(query grainConfigDTO.GrainStationQueryDTO) (int64, error) {
	dbQuery := applyStationQuery(r.Db.Model(&GrainStation{}).Where("grain_station.active = ?", 1), query.Search, query.Status)
	if query.StationID > 0 {
		dbQuery = dbQuery.Where("grain_station.id = ?", query.StationID)
	}
	if query.TenantID > 0 {
		dbQuery = dbQuery.Where("grain_station.tenant_id = ?", query.TenantID)
	}
	if len(query.StationIDs) > 0 {
		dbQuery = dbQuery.Where("grain_station.id IN ?", query.StationIDs)
	}
	if query.AppUserID > 0 {
		dbQuery = dbQuery.Joins("JOIN grain_station_user ON grain_station_user.station_id = grain_station.id AND grain_station_user.active = ? AND grain_station_user.status = ? AND grain_station_user.app_user_id = ?", 1, "active", query.AppUserID)
	}
	var total int64
	return total, dbQuery.Distinct("grain_station.id").Count(&total).Error
}

func (r *GrainStationRepository) ListByQuery(query grainConfigDTO.GrainStationQueryDTO, pageIndex, pageSize int) ([]*GrainStation, error) {
	dbQuery := applyStationQuery(r.Db.Model(&GrainStation{}).Where("grain_station.active = ?", 1), query.Search, query.Status)
	if query.StationID > 0 {
		dbQuery = dbQuery.Where("grain_station.id = ?", query.StationID)
	}
	if query.TenantID > 0 {
		dbQuery = dbQuery.Where("grain_station.tenant_id = ?", query.TenantID)
	}
	if len(query.StationIDs) > 0 {
		dbQuery = dbQuery.Where("grain_station.id IN ?", query.StationIDs)
	}
	if query.AppUserID > 0 {
		dbQuery = dbQuery.Joins("JOIN grain_station_user ON grain_station_user.station_id = grain_station.id AND grain_station_user.active = ? AND grain_station_user.status = ? AND grain_station_user.app_user_id = ?", 1, "active", query.AppUserID)
	}
	var entities []*GrainStation
	err := dbQuery.Distinct("grain_station.*").Order("grain_station.id DESC").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&entities).Error
	return entities, err
}

type GrainStationUserRepository struct {
	db.Repository[*GrainStationUser]
}

func (r *GrainStationUserRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&GrainStationUser{})
}

func (r *GrainStationUserRepository) FindActiveByAppUserID(appUserID uint64) (*GrainStationUser, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var entity GrainStationUser
	err := r.Db.Where("active = ? AND app_user_id = ? AND status = ?", 1, appUserID, "active").
		Order("id DESC").
		First(&entity).Error
	return &entity, err
}

func (r *GrainStationUserRepository) SaveActiveStation(appUserID, stationID uint64) (*GrainStationUser, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	if appUserID == 0 || stationID == 0 {
		return nil, fmt.Errorf("appUserId and stationId must be positive")
	}
	if err := r.Db.Model(&GrainStationUser{}).
		Where("active = ? AND app_user_id = ?", 1, appUserID).
		Updates(map[string]interface{}{"active": 0}).Error; err != nil {
		return nil, err
	}
	return r.Create(&GrainStationUser{
		StationID: stationID,
		AppUserID: appUserID,
		RoleType:  "salesman",
		Status:    "active",
	})
}

type GrainPurchaseTypeRepository struct {
	db.Repository[*GrainPurchaseType]
}

func (r *GrainPurchaseTypeRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&GrainPurchaseType{})
}

func (r *GrainPurchaseTypeRepository) ListActiveByStation(stationID uint64) ([]*GrainPurchaseType, error) {
	var entities []*GrainPurchaseType
	err := r.Db.Where("active = ? AND station_id = ? AND status = ?", 1, stationID, "active").
		Order("sort_order ASC, id ASC").
		Find(&entities).Error
	return entities, err
}

func (r *GrainPurchaseTypeRepository) CountByQuery(query grainConfigDTO.GrainConfigItemQueryDTO) (int64, error) {
	dbQuery := applyStationConfigItemQuery(r.Db.Model(&GrainPurchaseType{}).Where("active = ?", 1), query, "type_name")
	var total int64
	return total, dbQuery.Count(&total).Error
}

func (r *GrainPurchaseTypeRepository) ListByQuery(query grainConfigDTO.GrainConfigItemQueryDTO, pageIndex, pageSize int) ([]*GrainPurchaseType, error) {
	dbQuery := applyStationConfigItemQuery(r.Db.Model(&GrainPurchaseType{}).Where("active = ?", 1), query, "type_name")
	var entities []*GrainPurchaseType
	err := dbQuery.Order("sort_order ASC, id DESC").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&entities).Error
	return entities, err
}

type GrainPaymentMethodRepository struct {
	db.Repository[*GrainPaymentMethod]
}

func (r *GrainPaymentMethodRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&GrainPaymentMethod{})
}

func (r *GrainPaymentMethodRepository) ListActive() ([]*GrainPaymentMethod, error) {
	var entities []*GrainPaymentMethod
	err := r.Db.Where("active = ? AND status = ?", 1, "active").
		Order("sort_order ASC, id ASC").
		Find(&entities).Error
	return entities, err
}

func (r *GrainPaymentMethodRepository) CountByQuery(query grainConfigDTO.GrainConfigItemQueryDTO) (int64, error) {
	dbQuery := applyNamedConfigItemQuery(r.Db.Model(&GrainPaymentMethod{}).Where("active = ?", 1), query, "method_code", "method_name")
	var total int64
	return total, dbQuery.Count(&total).Error
}

func (r *GrainPaymentMethodRepository) ListByQuery(query grainConfigDTO.GrainConfigItemQueryDTO, pageIndex, pageSize int) ([]*GrainPaymentMethod, error) {
	dbQuery := applyNamedConfigItemQuery(r.Db.Model(&GrainPaymentMethod{}).Where("active = ?", 1), query, "method_code", "method_name")
	var entities []*GrainPaymentMethod
	err := dbQuery.Order("sort_order ASC, id DESC").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&entities).Error
	return entities, err
}

type GrainPurchasePlaceRepository struct {
	db.Repository[*GrainPurchasePlace]
}

func (r *GrainPurchasePlaceRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&GrainPurchasePlace{})
}

func (r *GrainPurchasePlaceRepository) ListActiveByAppUser(appUserID uint64) ([]*GrainPurchasePlace, error) {
	var entities []*GrainPurchasePlace
	err := r.Db.Where("active = ? AND app_user_id = ? AND status = ?", 1, appUserID, "active").
		Order("sort_order ASC, id ASC").
		Find(&entities).Error
	return entities, err
}

func (r *GrainPurchasePlaceRepository) CountByQuery(query grainConfigDTO.GrainConfigItemQueryDTO) (int64, error) {
	dbQuery := applyAppUserConfigItemQuery(r.Db.Model(&GrainPurchasePlace{}).Where("active = ?", 1), query, "place_name", "address")
	var total int64
	return total, dbQuery.Count(&total).Error
}

func (r *GrainPurchasePlaceRepository) ListByQuery(query grainConfigDTO.GrainConfigItemQueryDTO, pageIndex, pageSize int) ([]*GrainPurchasePlace, error) {
	dbQuery := applyAppUserConfigItemQuery(r.Db.Model(&GrainPurchasePlace{}).Where("active = ?", 1), query, "place_name", "address")
	var entities []*GrainPurchasePlace
	err := dbQuery.Order("sort_order ASC, id DESC").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&entities).Error
	return entities, err
}

func applyStationQuery(dbQuery *gorm.DB, search string, status string) *gorm.DB {
	if value := strings.TrimSpace(search); value != "" {
		likeValue := "%" + value + "%"
		dbQuery = dbQuery.Where("(grain_station.station_name LIKE ? OR grain_station.station_code LIKE ? OR grain_station.contact_name LIKE ? OR grain_station.contact_phone LIKE ? OR grain_station.address LIKE ?)", likeValue, likeValue, likeValue, likeValue, likeValue)
	}
	if value := strings.TrimSpace(status); value != "" {
		dbQuery = dbQuery.Where("grain_station.status = ?", value)
	}
	return dbQuery
}

func applyStationConfigItemQuery(dbQuery *gorm.DB, query grainConfigDTO.GrainConfigItemQueryDTO, searchColumns ...string) *gorm.DB {
	if query.StationID > 0 {
		dbQuery = dbQuery.Where("station_id = ?", query.StationID)
	}
	if len(query.StationIDs) > 0 {
		dbQuery = dbQuery.Where("station_id IN ?", query.StationIDs)
	}
	return applyNamedConfigItemQuery(dbQuery, query, searchColumns...)
}

func applyAppUserConfigItemQuery(dbQuery *gorm.DB, query grainConfigDTO.GrainConfigItemQueryDTO, searchColumns ...string) *gorm.DB {
	if query.AppUserID > 0 {
		dbQuery = dbQuery.Where("app_user_id = ?", query.AppUserID)
	}
	return applyNamedConfigItemQuery(dbQuery, query, searchColumns...)
}

func applyNamedConfigItemQuery(dbQuery *gorm.DB, query grainConfigDTO.GrainConfigItemQueryDTO, searchColumns ...string) *gorm.DB {
	if value := strings.TrimSpace(query.Search); value != "" && len(searchColumns) > 0 {
		likeValue := "%" + value + "%"
		conditions := make([]string, 0, len(searchColumns))
		args := make([]interface{}, 0, len(searchColumns))
		for _, column := range searchColumns {
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", column))
			args = append(args, likeValue)
		}
		dbQuery = dbQuery.Where("("+strings.Join(conditions, " OR ")+")", args...)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		dbQuery = dbQuery.Where("status = ?", value)
	}
	return dbQuery
}
