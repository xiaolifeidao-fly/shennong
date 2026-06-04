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

func (r *GrainFarmerRepository) ListByQuery(query grainFarmerDTO.GrainFarmerQueryDTO, pageIndex, pageSize int) ([]*GrainFarmerListRow, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	whereSQL, values := buildFarmerListWhere(query)
	sql := `SELECT f.*, COALESCE(gs.station_name, '') AS station_name
		FROM grain_farmer f
		LEFT JOIN grain_station gs ON gs.active = 1 AND gs.id = f.station_id ` +
		whereSQL + ` ORDER BY f.id DESC LIMIT ? OFFSET ?`
	values = append(values, pageSize, (pageIndex-1)*pageSize)
	var rows []*GrainFarmerListRow
	if err := r.QueryBySQL(&rows, sql, values...); err != nil {
		return nil, err
	}
	return rows, nil
}

func buildFarmerListWhere(query grainFarmerDTO.GrainFarmerQueryDTO) (string, []interface{}) {
	clauses := []string{"WHERE f.active = 1"}
	values := make([]interface{}, 0, 8)
	if query.StationID > 0 {
		clauses = append(clauses, "f.station_id = ?")
		values = append(values, query.StationID)
	}
	if len(query.StationIDs) > 0 {
		clauses = append(clauses, "f.station_id IN ?")
		values = append(values, query.StationIDs)
	}
	if query.AppUserID > 0 {
		clauses = append(clauses, "f.app_user_id = ?")
		values = append(values, query.AppUserID)
	}
	if value := strings.TrimSpace(query.Search); value != "" {
		clauses = append(clauses, "(f.id_number_digest = ? OR f.phone = ?)")
		values = append(values, query.SearchIDNumberDigest, value)
	}
	if value := strings.TrimSpace(query.Name); value != "" {
		clauses = append(clauses, "f.name_search LIKE ?")
		values = append(values, "%"+value+"%")
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		clauses = append(clauses, "f.status = ?")
		values = append(values, value)
	}
	return strings.Join(clauses, " AND "), values
}

func (r *GrainFarmerRepository) FindActiveByIDNumberDigest(idNumberDigest, plainIDNumber string, stationID uint64) (*GrainFarmer, error) {
	return r.FindActiveByIDNumberDigestForAppUser(idNumberDigest, plainIDNumber, stationID, 0)
}

func (r *GrainFarmerRepository) FindActiveByIDNumberDigestForAppUser(idNumberDigest, plainIDNumber string, stationID, appUserID uint64) (*GrainFarmer, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var entity GrainFarmer
	dbQuery := r.Db.Where("active = ? AND station_id = ? AND app_user_id = ?", 1, stationID, appUserID)
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
		dbQuery = dbQuery.Where("(id_number_digest = ? OR phone = ?)", query.SearchIDNumberDigest, value)
	}
	if value := strings.TrimSpace(query.Name); value != "" {
		dbQuery = dbQuery.Where("name_search LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		dbQuery = dbQuery.Where("status = ?", value)
	}
	return dbQuery
}
