package repository

import (
	"common/middleware/db"
	"fmt"
	grainPurchaseDTO "service/grain_purchase/dto"
	"strings"

	"gorm.io/gorm"
)

type GrainPurchaseEntryRepository struct {
	db.Repository[*GrainPurchaseEntry]
}

func (r *GrainPurchaseEntryRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&GrainPurchaseEntry{})
}

func (r *GrainPurchaseEntryRepository) CountByQuery(query grainPurchaseDTO.GrainPurchaseEntryQueryDTO) (int64, error) {
	dbQuery := applyEntryQuery(r.Db.Model(&GrainPurchaseEntry{}).Where("active = ?", 1), query)
	var total int64
	return total, dbQuery.Count(&total).Error
}

func (r *GrainPurchaseEntryRepository) ListByQuery(query grainPurchaseDTO.GrainPurchaseEntryQueryDTO, pageIndex, pageSize int) ([]*GrainPurchaseEntry, error) {
	dbQuery := applyEntryQuery(r.Db.Model(&GrainPurchaseEntry{}).Where("active = ?", 1), query)
	var entities []*GrainPurchaseEntry
	err := dbQuery.Order("buy_time DESC, id DESC").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&entities).Error
	return entities, err
}

type GrainPurchaseEntrySnapshotRepository struct {
	db.Repository[*GrainPurchaseEntrySnapshot]
}

func (r *GrainPurchaseEntrySnapshotRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&GrainPurchaseEntrySnapshot{})
}

func (r *GrainPurchaseEntrySnapshotRepository) CountByQuery(query grainPurchaseDTO.GrainEntrySnapshotQueryDTO) (int64, error) {
	dbQuery := applySnapshotQuery(r.Db.Model(&GrainPurchaseEntrySnapshot{}).Where("active = ?", 1), query)
	var total int64
	return total, dbQuery.Count(&total).Error
}

func (r *GrainPurchaseEntrySnapshotRepository) ListByQuery(query grainPurchaseDTO.GrainEntrySnapshotQueryDTO, pageIndex, pageSize int) ([]*GrainPurchaseEntrySnapshot, error) {
	dbQuery := applySnapshotQuery(r.Db.Model(&GrainPurchaseEntrySnapshot{}).Where("active = ?", 1), query)
	var entities []*GrainPurchaseEntrySnapshot
	err := dbQuery.Order("snapshot_version DESC, id DESC").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&entities).Error
	return entities, err
}

type GrainFarmerPurchaseSummaryRepository struct {
	db.Repository[*GrainFarmerPurchaseSummary]
}

func (r *GrainFarmerPurchaseSummaryRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&GrainFarmerPurchaseSummary{})
}

func (r *GrainFarmerPurchaseSummaryRepository) CountByQuery(query grainPurchaseDTO.GrainFarmerPurchaseSummaryQueryDTO) (int64, error) {
	dbQuery := applySummaryQuery(r.Db.Model(&GrainFarmerPurchaseSummary{}).Where("active = ?", 1), query)
	var total int64
	return total, dbQuery.Count(&total).Error
}

func (r *GrainFarmerPurchaseSummaryRepository) ListByQuery(query grainPurchaseDTO.GrainFarmerPurchaseSummaryQueryDTO, pageIndex, pageSize int) ([]*GrainFarmerPurchaseSummary, error) {
	dbQuery := applySummaryQuery(r.Db.Model(&GrainFarmerPurchaseSummary{}).Where("active = ?", 1), query)
	var entities []*GrainFarmerPurchaseSummary
	err := dbQuery.Order("summary_date DESC, farmer_id DESC, id DESC").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&entities).Error
	return entities, err
}

func (r *GrainFarmerPurchaseSummaryRepository) CountDailyFarmerSummaries(query grainPurchaseDTO.GrainFarmerDailySummaryQueryDTO) (int64, error) {
	dbQuery := applyDailySummaryQuery(r.dailySummaryBaseQuery(), query)
	var total int64
	return total, dbQuery.Distinct("f.id").Count(&total).Error
}

func (r *GrainFarmerPurchaseSummaryRepository) ListDailyFarmerSummaries(query grainPurchaseDTO.GrainFarmerDailySummaryQueryDTO, pageIndex, pageSize int) ([]*grainPurchaseDTO.GrainFarmerDailySummaryDTO, error) {
	dbQuery := applyDailySummaryQuery(r.dailySummaryBaseQuery(), query)
	var summaries []*grainPurchaseDTO.GrainFarmerDailySummaryDTO
	err := dbQuery.Select(`
		MAX(s.id) AS id,
		f.station_id AS station_id,
		f.app_user_id AS app_user_id,
		f.id AS farmer_id,
		f.name AS farmer_name,
		f.id_number AS farmer_id_number,
		f.phone AS farmer_phone,
		f.address AS farmer_address,
		f.bank_number AS bank_number,
		f.bank_name AS bank_name,
		f.status AS status,
		f.status_text AS status_text,
		MAX(s.summary_date) AS summary_date,
		MAX(s.crop) AS main_crop,
		SUM(s.entry_count) AS entry_count,
		SUM(s.total_amount) AS total_amount,
		SUM(s.total_quantity) AS total_quantity,
		MAX(s.updated_time) AS latest_time`,
	).Group(`
		f.station_id,
		f.app_user_id,
		f.id,
		f.name,
		f.id_number,
		f.phone,
		f.address,
		f.bank_number,
		f.bank_name,
		f.status,
		f.status_text`,
	).Order("latest_time DESC, farmer_id DESC").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Scan(&summaries).Error
	return summaries, err
}

func (r *GrainFarmerPurchaseSummaryRepository) dailySummaryBaseQuery() *gorm.DB {
	return r.Db.Table("grain_farmer_purchase_summary AS s").
		Joins("JOIN grain_farmer AS f ON f.id = s.farmer_id AND f.active = ?", 1).
		Where("s.active = ? AND s.entry_count > ?", 1, 0)
}

func (r *GrainFarmerPurchaseSummaryRepository) FindByDimension(entity *GrainFarmerPurchaseSummary) (*GrainFarmerPurchaseSummary, error) {
	var summary GrainFarmerPurchaseSummary
	err := r.Db.Where(
		"active = ? AND station_id = ? AND purchase_type_id = ? AND summary_date = ? AND farmer_id = ? AND payment_method_id = ?",
		1,
		entity.StationID,
		entity.PurchaseTypeID,
		entity.SummaryDate,
		entity.FarmerID,
		entity.PaymentMethodID,
	).First(&summary).Error
	return &summary, err
}

type GrainStationPurchaseSummaryRepository struct {
	db.Repository[*GrainStationPurchaseSummary]
}

func (r *GrainStationPurchaseSummaryRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&GrainStationPurchaseSummary{})
}

func (r *GrainStationPurchaseSummaryRepository) FindByDimension(entity *GrainStationPurchaseSummary) (*GrainStationPurchaseSummary, error) {
	var summary GrainStationPurchaseSummary
	err := r.Db.Where(
		"active = ? AND station_id = ? AND purchase_type_id = ? AND app_user_id = ? AND summary_date = ?",
		1,
		entity.StationID,
		entity.PurchaseTypeID,
		entity.AppUserID,
		entity.SummaryDate,
	).First(&summary).Error
	return &summary, err
}

type GrainEntryMaterialRepository struct {
	db.Repository[*GrainEntryMaterial]
}

func (r *GrainEntryMaterialRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&GrainEntryMaterial{})
}

func (r *GrainEntryMaterialRepository) CountByQuery(query grainPurchaseDTO.GrainEntryMaterialQueryDTO) (int64, error) {
	dbQuery := applyMaterialQuery(r.Db.Model(&GrainEntryMaterial{}).Where("active = ?", 1), query)
	var total int64
	return total, dbQuery.Count(&total).Error
}

func (r *GrainEntryMaterialRepository) ListByQuery(query grainPurchaseDTO.GrainEntryMaterialQueryDTO, pageIndex, pageSize int) ([]*GrainEntryMaterial, error) {
	dbQuery := applyMaterialQuery(r.Db.Model(&GrainEntryMaterial{}).Where("active = ?", 1), query)
	var entities []*GrainEntryMaterial
	err := dbQuery.Order("sort_order ASC, id DESC").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&entities).Error
	return entities, err
}

func applyEntryQuery(dbQuery *gorm.DB, query grainPurchaseDTO.GrainPurchaseEntryQueryDTO) *gorm.DB {
	if query.StationID > 0 {
		dbQuery = dbQuery.Where("station_id = ?", query.StationID)
	}
	if query.AppUserID > 0 {
		dbQuery = dbQuery.Where("app_user_id = ?", query.AppUserID)
	}
	if query.FarmerID > 0 {
		dbQuery = dbQuery.Where("farmer_id = ?", query.FarmerID)
	}
	if value := strings.TrimSpace(query.Search); value != "" {
		likeValue := "%" + value + "%"
		dbQuery = dbQuery.Where("(crop LIKE ? OR place LIKE ? OR pay_type LIKE ? OR location_address LIKE ?)", likeValue, likeValue, likeValue, likeValue)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		dbQuery = dbQuery.Where("status = ?", value)
	}
	if query.StartTime != nil {
		dbQuery = dbQuery.Where("buy_time >= ?", query.StartTime)
	}
	if query.EndTime != nil {
		dbQuery = dbQuery.Where("buy_time <= ?", query.EndTime)
	}
	return dbQuery
}

func applySnapshotQuery(dbQuery *gorm.DB, query grainPurchaseDTO.GrainEntrySnapshotQueryDTO) *gorm.DB {
	if query.EntryID > 0 {
		dbQuery = dbQuery.Where("entry_id = ?", query.EntryID)
	}
	if query.StationID > 0 {
		dbQuery = dbQuery.Where("station_id = ?", query.StationID)
	}
	if query.AppUserID > 0 {
		dbQuery = dbQuery.Where("app_user_id = ?", query.AppUserID)
	}
	return dbQuery
}

func applySummaryQuery(dbQuery *gorm.DB, query grainPurchaseDTO.GrainFarmerPurchaseSummaryQueryDTO) *gorm.DB {
	if query.StationID > 0 {
		dbQuery = dbQuery.Where("station_id = ?", query.StationID)
	}
	if query.AppUserID > 0 {
		dbQuery = dbQuery.Where("app_user_id = ?", query.AppUserID)
	}
	if query.FarmerID > 0 {
		dbQuery = dbQuery.Where("farmer_id = ?", query.FarmerID)
	}
	if query.StartDate != nil {
		dbQuery = dbQuery.Where("summary_date >= ?", query.StartDate)
	}
	if query.EndDate != nil {
		dbQuery = dbQuery.Where("summary_date <= ?", query.EndDate)
	}
	if value := strings.TrimSpace(query.Search); value != "" {
		likeValue := "%" + value + "%"
		dbQuery = dbQuery.Where("(crop LIKE ? OR pay_type LIKE ?)", likeValue, likeValue)
	}
	return dbQuery
}

func applyDailySummaryQuery(dbQuery *gorm.DB, query grainPurchaseDTO.GrainFarmerDailySummaryQueryDTO) *gorm.DB {
	if query.StationID > 0 {
		dbQuery = dbQuery.Where("s.station_id = ?", query.StationID)
	}
	if query.AppUserID > 0 {
		dbQuery = dbQuery.Where("s.app_user_id = ?", query.AppUserID)
	}
	if query.FarmerID > 0 {
		dbQuery = dbQuery.Where("s.farmer_id = ?", query.FarmerID)
	}
	if query.StartDate != nil {
		dbQuery = dbQuery.Where("s.summary_date >= ?", query.StartDate)
	}
	if query.EndDate != nil {
		dbQuery = dbQuery.Where("s.summary_date <= ?", query.EndDate)
	}
	if value := strings.TrimSpace(query.Search); value != "" {
		likeValue := "%" + value + "%"
		dbQuery = dbQuery.Where("(f.name LIKE ? OR f.id_number = ? OR f.phone LIKE ? OR f.address LIKE ? OR s.crop LIKE ? OR s.pay_type LIKE ?)", likeValue, value, likeValue, likeValue, likeValue, likeValue)
	}
	return dbQuery
}

func applyMaterialQuery(dbQuery *gorm.DB, query grainPurchaseDTO.GrainEntryMaterialQueryDTO) *gorm.DB {
	if query.StationID > 0 {
		dbQuery = dbQuery.Where("station_id = ?", query.StationID)
	}
	if query.EntryID > 0 {
		dbQuery = dbQuery.Where("entry_id = ?", query.EntryID)
	}
	if query.FarmerID > 0 {
		dbQuery = dbQuery.Where("farmer_id = ?", query.FarmerID)
	}
	if query.AppUserID > 0 {
		dbQuery = dbQuery.Where("app_user_id = ?", query.AppUserID)
	}
	if value := strings.TrimSpace(query.MaterialBizType); value != "" {
		dbQuery = dbQuery.Where("material_biz_type = ?", value)
	}
	if value := strings.TrimSpace(query.MaterialType); value != "" {
		dbQuery = dbQuery.Where("material_type = ?", value)
	}
	return dbQuery
}
