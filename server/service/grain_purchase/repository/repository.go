package repository

import (
	"common/middleware/db"
	"fmt"
	grainPurchaseDTO "service/grain_purchase/dto"
	"strings"
	"time"

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
	dbQuery := applyEntryQuery(r.entryListBaseQuery(), query)
	var total int64
	return total, dbQuery.Count(&total).Error
}

func (r *GrainPurchaseEntryRepository) ListByQuery(query grainPurchaseDTO.GrainPurchaseEntryQueryDTO, pageIndex, pageSize int) ([]*GrainPurchaseEntry, error) {
	dbQuery := applyEntryQuery(r.entryListBaseQuery(), query)
	var entities []*GrainPurchaseEntry
	err := dbQuery.Select("e.*").Order("e.id DESC").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Scan(&entities).Error
	return entities, err
}

func (r *GrainPurchaseEntryRepository) ListDTOByQuery(query grainPurchaseDTO.GrainPurchaseEntryQueryDTO, pageIndex, pageSize int) ([]*grainPurchaseDTO.GrainPurchaseEntryDTO, error) {
	dbQuery := applyEntryQuery(r.entryListBaseQuery(), query)
	var dtos []*grainPurchaseDTO.GrainPurchaseEntryDTO
	err := dbQuery.Joins("LEFT JOIN grain_station_extra AS gse ON gse.station_id = e.station_id AND gse.active = ?", 1).
		Select("e.*, gs.station_name AS station_name, gse.bank_account_number AS station_bank_account_number").
		Order("e.id DESC").
		Offset((pageIndex - 1) * pageSize).
		Limit(pageSize).
		Scan(&dtos).Error
	return dtos, err
}

func (r *GrainPurchaseEntryRepository) entryListBaseQuery() *gorm.DB {
	return r.Db.Table("grain_purchase_entry AS e").
		Joins("LEFT JOIN grain_station AS gs ON gs.id = e.station_id").
		Joins("JOIN grain_farmer AS f ON f.id = e.farmer_id AND f.active = ? AND COALESCE(f.status, '') <> ?", 1, "inactive").
		Where("e.active = ? AND e.status NOT IN ?", 1, []string{"voided", "deleted"})
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
		Joins("JOIN grain_farmer AS f ON f.id = s.farmer_id AND f.active = ? AND COALESCE(f.status, '') <> ?", 1, "inactive").
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

func (r *GrainStationPurchaseSummaryRepository) DashboardOverview(query grainPurchaseDTO.GrainPurchaseDashboardQueryDTO) (*grainPurchaseDTO.GrainPurchaseDashboardMetricDTO, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var metric grainPurchaseDTO.GrainPurchaseDashboardMetricDTO
	err := applyDashboardSummaryQuery(r.Db.Table("grain_station_purchase_summary AS s"), query).
		Select(`
			COALESCE(SUM(s.entry_count), 0) AS entry_count,
			COALESCE(SUM(s.total_quantity), 0) AS total_quantity,
			COALESCE(SUM(s.total_amount), 0) AS total_amount,
			COUNT(DISTINCT CASE WHEN s.entry_count > 0 THEN s.app_user_id END) AS active_user_count`,
		).
		Scan(&metric).Error
	if err != nil {
		return nil, err
	}
	if metric.TotalQuantity > 0 {
		metric.AverageUnitPrice = metric.TotalAmount / metric.TotalQuantity
	}
	return &metric, nil
}

func (r *GrainStationPurchaseSummaryRepository) DashboardByStation(query grainPurchaseDTO.GrainPurchaseDashboardQueryDTO) ([]*grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var rows []*grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO
	err := applyDashboardSummaryQuery(r.Db.Table("grain_station_purchase_summary AS s"), query).
		Joins("LEFT JOIN grain_station AS gs ON gs.id = s.station_id").
		Select(`
			CAST(s.station_id AS CHAR) AS ` + "`key`" + `,
			s.station_id AS station_id,
			COALESCE(NULLIF(gs.station_name, ''), CONCAT('粮站 ', s.station_id)) AS name,
			COALESCE(SUM(s.entry_count), 0) AS entry_count,
			COALESCE(SUM(s.total_quantity), 0) AS total_quantity,
			COALESCE(SUM(s.total_amount), 0) AS total_amount,
			COUNT(DISTINCT CASE WHEN s.entry_count > 0 THEN s.app_user_id END) AS active_user_count`,
		).
		Group("s.station_id, gs.station_name").
		Order("total_amount DESC, total_quantity DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *GrainStationPurchaseSummaryRepository) DashboardByCrop(query grainPurchaseDTO.GrainPurchaseDashboardQueryDTO) ([]*grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var rows []*grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO
	err := applyDashboardSummaryQuery(r.Db.Table("grain_station_purchase_summary AS s"), query).
		Select(`
			COALESCE(NULLIF(s.crop, ''), '未命名粮食') AS ` + "`key`" + `,
			COALESCE(NULLIF(s.crop, ''), '未命名粮食') AS name,
			MAX(s.purchase_type_id) AS purchase_type_id,
			COALESCE(SUM(s.entry_count), 0) AS entry_count,
			COALESCE(SUM(s.total_quantity), 0) AS total_quantity,
			COALESCE(SUM(s.total_amount), 0) AS total_amount,
			COUNT(DISTINCT CASE WHEN s.entry_count > 0 THEN s.app_user_id END) AS active_user_count`,
		).
		Group("COALESCE(NULLIF(s.crop, ''), '未命名粮食')").
		Order("total_amount DESC, total_quantity DESC").
		Scan(&rows).Error
	return rows, err
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

func (r *GrainEntryMaterialRepository) ListActiveByEntryID(entryID uint64) ([]*GrainEntryMaterial, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var entities []*GrainEntryMaterial
	err := r.Db.Where("entry_id = ? AND active = ?", entryID, 1).
		Order("sort_order ASC, id ASC").
		Find(&entities).Error
	return entities, err
}

func (r *GrainEntryMaterialRepository) ListActiveByEntryIDs(entryIDs []uint64) ([]*GrainEntryMaterial, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	if len(entryIDs) == 0 {
		return []*GrainEntryMaterial{}, nil
	}
	var entities []*GrainEntryMaterial
	err := r.Db.Where("entry_id IN ? AND active = ?", entryIDs, 1).
		Order("entry_id ASC, sort_order ASC, id ASC").
		Find(&entities).Error
	return entities, err
}

func (r *GrainEntryMaterialRepository) MaxActiveCountByEntryQuery(query grainPurchaseDTO.GrainPurchaseEntryQueryDTO) (int, error) {
	if r.Db == nil {
		return 0, fmt.Errorf("database is not initialized")
	}
	type row struct {
		MaterialCount int
	}
	var rows []row
	err := applyEntryQuery(
		r.Db.Table("grain_entry_material AS m").
			Joins("JOIN grain_purchase_entry AS e ON e.id = m.entry_id").
			Joins("JOIN grain_farmer AS f ON f.id = e.farmer_id AND f.active = ? AND COALESCE(f.status, '') <> ?", 1, "inactive").
			Where("e.active = ? AND e.status NOT IN ?", 1, []string{"voided", "deleted"}),
		query,
	).
		Where("m.active = ?", 1).
		Select("COUNT(m.id) AS material_count").
		Group("m.entry_id").
		Order("material_count DESC").
		Limit(1).
		Scan(&rows).Error
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 || rows[0].MaterialCount < 0 {
		return 0, nil
	}
	return rows[0].MaterialCount, nil
}

func (r *GrainEntryMaterialRepository) FindByUnique(entryID uint64, imageHash string) (*GrainEntryMaterial, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var entity GrainEntryMaterial
	err := r.Db.Where("entry_id = ? AND image_hash = ? AND active = 1", entryID, imageHash).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GrainEntryMaterialRepository) FindOrCreate(entity *GrainEntryMaterial) (*GrainEntryMaterial, bool, error) {
	if r.Db == nil {
		return nil, false, fmt.Errorf("database is not initialized")
	}
	existing, err := r.FindByUnique(entity.EntryID, entity.ImageHash)
	if err == nil {
		return existing, false, nil
	}
	if err2 := r.Db.Create(entity).Error; err2 != nil {
		return nil, false, err2
	}
	return entity, true, nil
}

func (r *GrainEntryMaterialRepository) UpdateCloudSource(id int, wxCloudURL, lastSource string) error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.Model(&GrainEntryMaterial{}).Where("id = ?", id).Updates(map[string]interface{}{
		"wx_cloud_url": wxCloudURL,
		"last_source":  lastSource,
	}).Error
}

func (r *GrainEntryMaterialRepository) DeletePhysicalByID(id uint) error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	result := r.Db.Delete(&GrainEntryMaterial{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

type GrainPurchaseEntryExportBatchRepository struct {
	db.Repository[*GrainPurchaseEntryExportBatch]
}

func (r *GrainPurchaseEntryExportBatchRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&GrainPurchaseEntryExportBatch{})
}

func (r *GrainPurchaseEntryExportBatchRepository) CountByUser(userID uint64, query grainPurchaseDTO.GrainPurchaseEntryExportQueryDTO) (int64, error) {
	if r.Db == nil {
		return 0, fmt.Errorf("database is not initialized")
	}
	dbQuery := applyExportBatchQuery(r.Db.Model(&GrainPurchaseEntryExportBatch{}).Where("active = ? AND user_id = ?", 1, userID), query)
	var total int64
	return total, dbQuery.Count(&total).Error
}

func (r *GrainPurchaseEntryExportBatchRepository) ListByUser(userID uint64, query grainPurchaseDTO.GrainPurchaseEntryExportQueryDTO, pageIndex, pageSize int) ([]*GrainPurchaseEntryExportBatch, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	dbQuery := applyExportBatchQuery(r.Db.Model(&GrainPurchaseEntryExportBatch{}).Where("active = ? AND user_id = ?", 1, userID), query)
	var entities []*GrainPurchaseEntryExportBatch
	err := dbQuery.Order("id DESC").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&entities).Error
	return entities, err
}

func (r *GrainPurchaseEntryExportBatchRepository) FindLatestRunningByUser(userID uint64) (*GrainPurchaseEntryExportBatch, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var entity GrainPurchaseEntryExportBatch
	err := r.Db.Where("active = ? AND user_id = ? AND status IN ?", 1, userID, []string{"pending", "running"}).
		Order("id DESC").
		First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GrainPurchaseEntryExportBatchRepository) FindByBatchNoForUser(batchNo string, userID uint64) (*GrainPurchaseEntryExportBatch, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var entity GrainPurchaseEntryExportBatch
	err := r.Db.Where("active = ? AND batch_no = ? AND user_id = ?", 1, batchNo, userID).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GrainPurchaseEntryExportBatchRepository) UpdateProgress(id int, status string, successCount, failCount int, fileName, filePath, errorMessage string, finishedAt *time.Time) error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	updates := map[string]interface{}{
		"status":        status,
		"success_count": successCount,
		"fail_count":    failCount,
		"file_name":     fileName,
		"file_path":     filePath,
		"error_message": errorMessage,
		"updated_time":  time.Now(),
	}
	if finishedAt != nil {
		updates["finished_at"] = finishedAt
	}
	return r.Db.Model(&GrainPurchaseEntryExportBatch{}).Where("id = ?", id).Updates(updates).Error
}

func applyExportBatchQuery(dbQuery *gorm.DB, query grainPurchaseDTO.GrainPurchaseEntryExportQueryDTO) *gorm.DB {
	if value := strings.TrimSpace(query.Status); value != "" {
		dbQuery = dbQuery.Where("status = ?", value)
	}
	return dbQuery
}

func applyEntryQuery(dbQuery *gorm.DB, query grainPurchaseDTO.GrainPurchaseEntryQueryDTO) *gorm.DB {
	if query.StationID > 0 {
		dbQuery = dbQuery.Where("e.station_id = ?", query.StationID)
	}
	if len(query.StationIDs) > 0 {
		dbQuery = dbQuery.Where("e.station_id IN ?", query.StationIDs)
	}
	if len(query.AppUserIDs) > 0 {
		dbQuery = dbQuery.Where("e.app_user_id IN ?", query.AppUserIDs)
	} else if query.AppUserID > 0 {
		dbQuery = dbQuery.Where("e.app_user_id = ?", query.AppUserID)
	}
	if len(query.FarmerIDs) > 0 {
		dbQuery = dbQuery.Where("e.farmer_id IN ?", query.FarmerIDs)
	} else if query.FarmerID > 0 {
		dbQuery = dbQuery.Where("e.farmer_id = ?", query.FarmerID)
	}
	if len(query.PurchaseTypeIDs) > 0 {
		dbQuery = dbQuery.Where("e.purchase_type_id IN ?", query.PurchaseTypeIDs)
	} else if query.PurchaseTypeID > 0 {
		dbQuery = dbQuery.Where("e.purchase_type_id = ?", query.PurchaseTypeID)
	}
	if value := strings.TrimSpace(query.Search); value != "" {
		likeValue := "%" + value + "%"
		farmerClauses, farmerValues := entryFarmerSearchClauses(query.Search, query.SearchIDNumberDigest, query.SearchIDNumberLast4Digest, query.SearchNameDigest, query.SearchNamePrefixCode)
		clauses := []string{"e.crop LIKE ?", "e.place LIKE ?", "e.pay_type LIKE ?", "e.location_address LIKE ?"}
		values := []interface{}{likeValue, likeValue, likeValue, likeValue}
		clauses = append(clauses, farmerClauses...)
		values = append(values, farmerValues...)
		dbQuery = dbQuery.Where("("+strings.Join(clauses, " OR ")+")", values...)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		dbQuery = dbQuery.Where("e.status = ?", value)
	}
	if startDate := strings.TrimSpace(query.StartDate); startDate != "" {
		dbQuery = dbQuery.Where("DATE(e.created_time) >= ?", startDate)
	} else if query.StartTime != nil {
		dbQuery = dbQuery.Where("e.created_time >= ?", query.StartTime)
	}
	if endDate := strings.TrimSpace(query.EndDate); endDate != "" {
		dbQuery = dbQuery.Where("DATE(e.created_time) <= ?", endDate)
	} else if query.EndTime != nil {
		dbQuery = dbQuery.Where("e.created_time <= ?", query.EndTime)
	}
	if query.MinAmount != nil {
		dbQuery = dbQuery.Where("e.amount >= ?", *query.MinAmount)
	}
	if query.MaxAmount != nil {
		dbQuery = dbQuery.Where("e.amount <= ?", *query.MaxAmount)
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
	if len(query.StationIDs) > 0 {
		dbQuery = dbQuery.Where("station_id IN ?", query.StationIDs)
	}
	if query.AppUserID > 0 {
		dbQuery = dbQuery.Where("app_user_id = ?", query.AppUserID)
	}
	return dbQuery
}

func applySummaryQuery(dbQuery *gorm.DB, query grainPurchaseDTO.GrainFarmerPurchaseSummaryQueryDTO) *gorm.DB {
	dbQuery = dbQuery.Where("entry_count > ?", 0)
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
	if len(query.StationIDs) > 0 {
		dbQuery = dbQuery.Where("s.station_id IN ?", query.StationIDs)
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
		farmerClauses, farmerValues := entryFarmerSearchClauses(query.Search, query.SearchIDNumberDigest, query.SearchIDNumberLast4Digest, query.SearchNameDigest, query.SearchNamePrefixCode)
		clauses := []string{"f.phone LIKE ?", "f.address LIKE ?", "s.crop LIKE ?", "s.pay_type LIKE ?"}
		values := []interface{}{likeValue, likeValue, likeValue, likeValue}
		clauses = append(clauses, farmerClauses...)
		values = append(values, farmerValues...)
		dbQuery = dbQuery.Where("("+strings.Join(clauses, " OR ")+")", values...)
	}
	return dbQuery
}

func entryFarmerSearchClauses(rawValue, idDigest, idLast4Digest, nameDigest, namePrefixCode string) ([]string, []interface{}) {
	value := strings.TrimSpace(rawValue)
	clauses := make([]string, 0, 5)
	values := make([]interface{}, 0, 5)
	if value == "" {
		return clauses, values
	}
	if idDigest != "" {
		clauses = append(clauses, "f.id_number_digest = ?")
		values = append(values, idDigest)
	}
	if idLast4Digest != "" && len(value) == 4 {
		clauses = append(clauses, "f.id_number_last4_digest = ?")
		values = append(values, idLast4Digest)
	}
	if nameDigest != "" {
		clauses = append(clauses, "f.name_digest = ?")
		values = append(values, nameDigest)
	}
	if namePrefixCode != "" {
		clauses = append(clauses, "f.name_search LIKE ?")
		values = append(values, namePrefixCode+"%")
	}
	clauses = append(clauses, "f.phone = ?")
	values = append(values, value)
	return clauses, values
}

func (r *GrainFarmerPurchaseSummaryRepository) DashboardNewFarmerCount(query grainPurchaseDTO.GrainPurchaseDashboardQueryDTO) (int, error) {
	if r.Db == nil {
		return 0, fmt.Errorf("database is not initialized")
	}
	var total int64
	err := applyDashboardFarmerSummaryQuery(r.activeFarmerSummaryBaseQuery(), query).
		Distinct("s.farmer_id").
		Count(&total).Error
	return int(total), err
}

func (r *GrainFarmerPurchaseSummaryRepository) DashboardOverview(query grainPurchaseDTO.GrainPurchaseDashboardQueryDTO) (*grainPurchaseDTO.GrainPurchaseDashboardMetricDTO, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var metric grainPurchaseDTO.GrainPurchaseDashboardMetricDTO
	err := applyDashboardFarmerSummaryQuery(r.activeFarmerSummaryBaseQuery(), query).
		Select(`
			COALESCE(SUM(s.entry_count), 0) AS entry_count,
			COALESCE(SUM(s.total_quantity), 0) AS total_quantity,
			COALESCE(SUM(s.total_amount), 0) AS total_amount,
			COUNT(DISTINCT CASE WHEN s.entry_count > 0 THEN s.app_user_id END) AS active_user_count,
			COUNT(DISTINCT CASE WHEN s.entry_count > 0 THEN s.farmer_id END) AS new_farmer_count`,
		).
		Scan(&metric).Error
	if err != nil {
		return nil, err
	}
	if metric.TotalQuantity > 0 {
		metric.AverageUnitPrice = metric.TotalAmount / metric.TotalQuantity
	}
	if metric.NewFarmerCount > 0 {
		metric.AverageFarmerDeal = metric.TotalAmount / float64(metric.NewFarmerCount)
	}
	return &metric, nil
}

func (r *GrainFarmerPurchaseSummaryRepository) DashboardFarmerCountByStation(query grainPurchaseDTO.GrainPurchaseDashboardQueryDTO) (map[uint64]int, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	type row struct {
		StationID   uint64
		FarmerCount int
	}
	var rows []row
	err := applyDashboardFarmerSummaryQuery(r.activeFarmerSummaryBaseQuery(), query).
		Select("s.station_id AS station_id, COUNT(DISTINCT s.farmer_id) AS farmer_count").
		Group("s.station_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint64]int, len(rows))
	for _, item := range rows {
		result[item.StationID] = item.FarmerCount
	}
	return result, nil
}

func (r *GrainFarmerPurchaseSummaryRepository) DashboardFarmerCountByCrop(query grainPurchaseDTO.GrainPurchaseDashboardQueryDTO) (map[string]int, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	type row struct {
		Name        string
		FarmerCount int
	}
	var rows []row
	err := applyDashboardFarmerSummaryQuery(r.activeFarmerSummaryBaseQuery(), query).
		Select("COALESCE(NULLIF(s.crop, ''), '未命名粮食') AS name, COUNT(DISTINCT s.farmer_id) AS farmer_count").
		Group("COALESCE(NULLIF(s.crop, ''), '未命名粮食')").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]int, len(rows))
	for _, item := range rows {
		result[item.Name] = item.FarmerCount
	}
	return result, nil
}

func (r *GrainFarmerPurchaseSummaryRepository) DashboardByStation(query grainPurchaseDTO.GrainPurchaseDashboardQueryDTO) ([]*grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var rows []*grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO
	err := applyDashboardFarmerSummaryQuery(r.activeFarmerSummaryBaseQuery(), query).
		Joins("LEFT JOIN grain_station AS gs ON gs.id = s.station_id").
		Select(`
			CAST(s.station_id AS CHAR) AS ` + "`key`" + `,
			s.station_id AS station_id,
			COALESCE(NULLIF(gs.station_name, ''), CONCAT('粮站 ', s.station_id)) AS name,
			COALESCE(SUM(s.entry_count), 0) AS entry_count,
			COALESCE(SUM(s.total_quantity), 0) AS total_quantity,
			COALESCE(SUM(s.total_amount), 0) AS total_amount,
			COUNT(DISTINCT CASE WHEN s.entry_count > 0 THEN s.app_user_id END) AS active_user_count,
			COUNT(DISTINCT CASE WHEN s.entry_count > 0 THEN s.farmer_id END) AS farmer_count`,
		).
		Group("s.station_id, gs.station_name").
		Order("total_amount DESC, total_quantity DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *GrainFarmerPurchaseSummaryRepository) DashboardByCrop(query grainPurchaseDTO.GrainPurchaseDashboardQueryDTO) ([]*grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var rows []*grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO
	err := applyDashboardFarmerSummaryQuery(r.activeFarmerSummaryBaseQuery(), query).
		Select(`
			COALESCE(NULLIF(s.crop, ''), '未命名粮食') AS ` + "`key`" + `,
			COALESCE(NULLIF(s.crop, ''), '未命名粮食') AS name,
			MAX(s.purchase_type_id) AS purchase_type_id,
			COALESCE(SUM(s.entry_count), 0) AS entry_count,
			COALESCE(SUM(s.total_quantity), 0) AS total_quantity,
			COALESCE(SUM(s.total_amount), 0) AS total_amount,
			COUNT(DISTINCT CASE WHEN s.entry_count > 0 THEN s.app_user_id END) AS active_user_count,
			COUNT(DISTINCT CASE WHEN s.entry_count > 0 THEN s.farmer_id END) AS farmer_count`,
		).
		Group("COALESCE(NULLIF(s.crop, ''), '未命名粮食')").
		Order("total_amount DESC, total_quantity DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *GrainFarmerPurchaseSummaryRepository) activeFarmerSummaryBaseQuery() *gorm.DB {
	return r.Db.Table("grain_farmer_purchase_summary AS s").
		Joins("JOIN grain_farmer AS f ON f.id = s.farmer_id AND f.active = ? AND COALESCE(f.status, '') <> ?", 1, "inactive")
}

func applyDashboardSummaryQuery(dbQuery *gorm.DB, query grainPurchaseDTO.GrainPurchaseDashboardQueryDTO) *gorm.DB {
	dbQuery = dbQuery.Where("s.active = ? AND s.entry_count > ?", 1, 0)
	if query.StationID > 0 {
		dbQuery = dbQuery.Where("s.station_id = ?", query.StationID)
	}
	if len(query.StationIDs) > 0 {
		dbQuery = dbQuery.Where("s.station_id IN ?", query.StationIDs)
	}
	if query.AppUserID > 0 {
		dbQuery = dbQuery.Where("s.app_user_id = ?", query.AppUserID)
	}
	if query.StartDate != nil {
		dbQuery = dbQuery.Where("s.summary_date >= ?", query.StartDate)
	}
	if query.EndDate != nil {
		dbQuery = dbQuery.Where("s.summary_date <= ?", query.EndDate)
	}
	return dbQuery
}

func applyDashboardFarmerSummaryQuery(dbQuery *gorm.DB, query grainPurchaseDTO.GrainPurchaseDashboardQueryDTO) *gorm.DB {
	dbQuery = dbQuery.Where("s.active = ? AND s.entry_count > ?", 1, 0)
	if query.StationID > 0 {
		dbQuery = dbQuery.Where("s.station_id = ?", query.StationID)
	}
	if len(query.StationIDs) > 0 {
		dbQuery = dbQuery.Where("s.station_id IN ?", query.StationIDs)
	}
	if query.AppUserID > 0 {
		dbQuery = dbQuery.Where("s.app_user_id = ?", query.AppUserID)
	}
	if query.StartDate != nil {
		dbQuery = dbQuery.Where("s.summary_date >= ?", query.StartDate)
	}
	if query.EndDate != nil {
		dbQuery = dbQuery.Where("s.summary_date <= ?", query.EndDate)
	}
	return dbQuery
}

func applyMaterialQuery(dbQuery *gorm.DB, query grainPurchaseDTO.GrainEntryMaterialQueryDTO) *gorm.DB {
	if query.StationID > 0 {
		dbQuery = dbQuery.Where("station_id = ?", query.StationID)
	}
	if len(query.StationIDs) > 0 {
		dbQuery = dbQuery.Where("station_id IN ?", query.StationIDs)
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
