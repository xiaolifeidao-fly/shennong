package tenant

import (
	baseDTO "common/base/dto"
	"common/middleware/db"
	"fmt"
	grainConfigRepository "service/grain_config/repository"
	tenantDTO "service/tenant/dto"
	tenantRepository "service/tenant/repository"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type TenantService struct {
	tenantRepository        *tenantRepository.TenantRepository
	tenantStationRepository *tenantRepository.TenantStationRepository
	stationRepository       *grainConfigRepository.GrainStationRepository
}

func NewTenantService() *TenantService {
	return &TenantService{
		tenantRepository:        db.GetRepository[tenantRepository.TenantRepository](),
		tenantStationRepository: db.GetRepository[tenantRepository.TenantStationRepository](),
		stationRepository:       db.GetRepository[grainConfigRepository.GrainStationRepository](),
	}
}

func (s *TenantService) EnsureTable() error {
	if err := s.tenantRepository.EnsureTable(); err != nil {
		return err
	}
	return s.tenantStationRepository.EnsureTable()
}

func (s *TenantService) ListTenants(query tenantDTO.TenantQueryDTO) (*baseDTO.PageDTO[tenantDTO.TenantDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	total, err := s.tenantRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	entities, err := s.tenantRepository.ListByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	items := db.ToDTOs[tenantDTO.TenantDTO](entities)
	if err := s.fillTenantStations(items); err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), items), nil
}

func (s *TenantService) GetTenantByID(id uint) (*tenantDTO.TenantDTO, error) {
	entity, err := s.tenantRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	result := db.ToDTO[tenantDTO.TenantDTO](entity)
	return result, s.fillTenantStations([]*tenantDTO.TenantDTO{result})
}

func (s *TenantService) CreateTenant(req *tenantDTO.TenantDTO) (*tenantDTO.TenantDTO, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	entity := db.ToPO[tenantRepository.Tenant](req)
	trimTenant(entity)
	if entity.TenantName == "" {
		return nil, fmt.Errorf("tenantName is required")
	}
	if entity.TenantCode == "" {
		entity.TenantCode = generateTenantCode()
	}
	if entity.Status == "" {
		entity.Status = "active"
	}
	if err := s.ensureCodeAvailable(entity.TenantCode, 0); err != nil {
		return nil, err
	}
	created, err := s.tenantRepository.Create(entity)
	if err != nil {
		return nil, err
	}
	if err := s.saveTenantStations(uint64(created.Id), req.StationIDs); err != nil {
		return nil, err
	}
	return s.GetTenantByID(uint(created.Id))
}

func (s *TenantService) UpdateTenant(id uint, req *tenantDTO.TenantDTO) (*tenantDTO.TenantDTO, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	entity, err := s.tenantRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	previousBase := entity.BaseEntity
	previousCode := strings.TrimSpace(entity.TenantCode)
	copier.Copy(entity, req)
	entity.Id = int(id)
	entity.BaseEntity = previousBase
	trimTenant(entity)
	if entity.TenantName == "" {
		return nil, fmt.Errorf("tenantName is required")
	}
	if entity.TenantCode == "" {
		entity.TenantCode = previousCode
	}
	if entity.TenantCode == "" {
		entity.TenantCode = generateTenantCode()
	}
	if entity.Status == "" {
		entity.Status = "active"
	}
	if err := s.ensureCodeAvailable(entity.TenantCode, entity.Id); err != nil {
		return nil, err
	}
	saved, err := s.tenantRepository.SaveOrUpdate(entity)
	if err != nil {
		return nil, err
	}
	if err := s.saveTenantStations(uint64(saved.Id), req.StationIDs); err != nil {
		return nil, err
	}
	return s.GetTenantByID(uint(saved.Id))
}

func (s *TenantService) DeleteTenant(id uint) error {
	entity, err := s.tenantRepository.FindById(id)
	if err != nil {
		return err
	}
	if entity.Active == 0 {
		return gorm.ErrRecordNotFound
	}
	entity.Active = 0
	if _, err := s.tenantRepository.SaveOrUpdate(entity); err != nil {
		return err
	}
	if s.stationRepository.Db != nil {
		if err := s.stationRepository.Db.Model(&grainConfigRepository.GrainStation{}).
			Where("tenant_id = ?", id).
			Update("tenant_id", 0).Error; err != nil {
			return err
		}
	}
	return s.tenantStationRepository.ReplaceActiveStations(uint64(id), nil)
}

func (s *TenantService) ListStationIDsByTenant(tenantID uint64) ([]uint64, error) {
	if tenantID == 0 {
		return nil, nil
	}
	return s.tenantStationRepository.ListActiveStationIDs(tenantID)
}

func (s *TenantService) saveTenantStations(tenantID uint64, stationIDs []uint64) error {
	if err := s.tenantStationRepository.ReplaceActiveStations(tenantID, stationIDs); err != nil {
		return err
	}
	if s.stationRepository.Db == nil {
		return nil
	}
	if err := s.stationRepository.Db.Model(&grainConfigRepository.GrainStation{}).
		Where("tenant_id = ?", tenantID).
		Update("tenant_id", 0).Error; err != nil {
		return err
	}
	if len(stationIDs) == 0 {
		return nil
	}
	return s.stationRepository.Db.Model(&grainConfigRepository.GrainStation{}).
		Where("id IN ?", stationIDs).
		Update("tenant_id", tenantID).Error
}

func (s *TenantService) fillTenantStations(items []*tenantDTO.TenantDTO) error {
	if len(items) == 0 || s.tenantStationRepository.Db == nil {
		return nil
	}
	tenantIDs := make([]int, 0, len(items))
	itemByID := make(map[int]*tenantDTO.TenantDTO, len(items))
	for _, item := range items {
		tenantIDs = append(tenantIDs, item.Id)
		itemByID[item.Id] = item
	}
	var rows []struct {
		TenantID    int
		StationID   uint64
		StationName string
	}
	err := s.tenantStationRepository.Db.Table("tenant_station AS ts").
		Select("ts.tenant_id AS tenant_id, ts.station_id AS station_id, gs.station_name AS station_name").
		Joins("LEFT JOIN grain_station AS gs ON gs.id = ts.station_id AND gs.active = ?", 1).
		Where("ts.active = ? AND ts.status = ? AND ts.tenant_id IN ?", 1, "active", tenantIDs).
		Order("ts.id ASC").
		Scan(&rows).Error
	if err != nil {
		return err
	}
	for _, row := range rows {
		if item, ok := itemByID[row.TenantID]; ok {
			item.StationIDs = append(item.StationIDs, row.StationID)
			if strings.TrimSpace(row.StationName) != "" {
				item.StationNames = append(item.StationNames, row.StationName)
			}
		}
	}
	return nil
}

func (s *TenantService) ensureCodeAvailable(code string, selfID int) error {
	existing, err := s.tenantRepository.FindByCode(code)
	if err == nil && existing != nil && existing.Active == 1 && existing.Id != selfID {
		return fmt.Errorf("tenantCode already exists")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

func trimTenant(entity *tenantRepository.Tenant) {
	entity.TenantName = strings.TrimSpace(entity.TenantName)
	entity.TenantCode = strings.TrimSpace(entity.TenantCode)
	entity.ContactName = strings.TrimSpace(entity.ContactName)
	entity.ContactPhone = strings.TrimSpace(entity.ContactPhone)
	entity.Status = strings.TrimSpace(entity.Status)
	entity.Remark = strings.TrimSpace(entity.Remark)
}

func normalizePage(page, pageIndex, pageSize int) (int, int) {
	if pageIndex <= 0 {
		pageIndex = page
	}
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return pageIndex, pageSize
}

func generateTenantCode() string {
	return fmt.Sprintf("T%d", time.Now().UnixNano())
}
