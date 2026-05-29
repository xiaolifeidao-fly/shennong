package grain_config

import (
	baseDTO "common/base/dto"
	"common/middleware/db"
	"fmt"
	grainConfigDTO "service/grain_config/dto"
	grainConfigRepository "service/grain_config/repository"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type GrainConfigService struct {
	stationRepository       *grainConfigRepository.GrainStationRepository
	stationUserRepository   *grainConfigRepository.GrainStationUserRepository
	purchaseTypeRepository  *grainConfigRepository.GrainPurchaseTypeRepository
	paymentMethodRepository *grainConfigRepository.GrainPaymentMethodRepository
	purchasePlaceRepository *grainConfigRepository.GrainPurchasePlaceRepository
}

func NewGrainConfigService() *GrainConfigService {
	return &GrainConfigService{
		stationRepository:       db.GetRepository[grainConfigRepository.GrainStationRepository](),
		stationUserRepository:   db.GetRepository[grainConfigRepository.GrainStationUserRepository](),
		purchaseTypeRepository:  db.GetRepository[grainConfigRepository.GrainPurchaseTypeRepository](),
		paymentMethodRepository: db.GetRepository[grainConfigRepository.GrainPaymentMethodRepository](),
		purchasePlaceRepository: db.GetRepository[grainConfigRepository.GrainPurchasePlaceRepository](),
	}
}

func (s *GrainConfigService) EnsureTable() error {
	steps := []func() error{
		s.stationRepository.EnsureTable,
		s.stationUserRepository.EnsureTable,
		s.purchaseTypeRepository.EnsureTable,
		s.paymentMethodRepository.EnsureTable,
		s.purchasePlaceRepository.EnsureTable,
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func (s *GrainConfigService) ListStations(query grainConfigDTO.GrainStationQueryDTO) (*baseDTO.PageDTO[grainConfigDTO.GrainStationDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	total, err := s.stationRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	entities, err := s.stationRepository.ListByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), db.ToDTOs[grainConfigDTO.GrainStationDTO](entities)), nil
}

func (s *GrainConfigService) CreateStation(req *grainConfigDTO.GrainStationDTO) (*grainConfigDTO.GrainStationDTO, error) {
	entity := db.ToPO[grainConfigRepository.GrainStation](req)
	entity.StationName = strings.TrimSpace(entity.StationName)
	entity.StationCode = strings.TrimSpace(entity.StationCode)
	if entity.StationCode == "" {
		entity.StationCode = generateStationCode()
	}
	if strings.TrimSpace(entity.Status) == "" {
		entity.Status = "active"
	}
	result, err := s.stationRepository.Create(entity)
	if err != nil {
		return nil, err
	}
	return db.ToDTO[grainConfigDTO.GrainStationDTO](result), nil
}

func (s *GrainConfigService) UpdateStation(id uint, req *grainConfigDTO.GrainStationDTO) (*grainConfigDTO.GrainStationDTO, error) {
	entity, err := s.stationRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	previousBase := entity.BaseEntity
	previousStationCode := strings.TrimSpace(entity.StationCode)
	copier.Copy(entity, req)
	entity.Id = int(id)
	preserveBaseEntityFields(&entity.BaseEntity, previousBase)
	entity.StationName = strings.TrimSpace(entity.StationName)
	entity.StationCode = strings.TrimSpace(entity.StationCode)
	if entity.StationCode == "" {
		entity.StationCode = previousStationCode
	}
	if entity.StationCode == "" {
		entity.StationCode = generateStationCode()
	}
	if strings.TrimSpace(entity.Status) == "" {
		entity.Status = "active"
	}
	result, err := s.stationRepository.SaveOrUpdate(entity)
	if err != nil {
		return nil, err
	}
	return db.ToDTO[grainConfigDTO.GrainStationDTO](result), nil
}

func (s *GrainConfigService) DeleteStation(id uint) error {
	entity, err := s.stationRepository.FindById(id)
	if err != nil {
		return err
	}
	if entity.Active == 0 {
		return gorm.ErrRecordNotFound
	}
	entity.Active = 0
	_, err = s.stationRepository.SaveOrUpdate(entity)
	return err
}

func (s *GrainConfigService) ListPurchaseTypes(stationID uint64) ([]*grainConfigDTO.GrainPurchaseTypeDTO, error) {
	entities, err := s.purchaseTypeRepository.ListActiveByStation(stationID)
	if err != nil {
		return nil, err
	}
	return db.ToDTOs[grainConfigDTO.GrainPurchaseTypeDTO](entities), nil
}

func (s *GrainConfigService) CreatePurchaseType(req *grainConfigDTO.GrainPurchaseTypeDTO) (*grainConfigDTO.GrainPurchaseTypeDTO, error) {
	entity := db.ToPO[grainConfigRepository.GrainPurchaseType](req)
	normalizeStatusAndOrder(&entity.Status, &entity.SortOrder)
	if entity.Unit == "" {
		entity.Unit = "公斤"
	}
	result, err := s.purchaseTypeRepository.Create(entity)
	if err != nil {
		return nil, err
	}
	return db.ToDTO[grainConfigDTO.GrainPurchaseTypeDTO](result), nil
}

func (s *GrainConfigService) ListPurchaseTypesPage(query grainConfigDTO.GrainConfigItemQueryDTO) (*baseDTO.PageDTO[grainConfigDTO.GrainPurchaseTypeDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	total, err := s.purchaseTypeRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	entities, err := s.purchaseTypeRepository.ListByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), db.ToDTOs[grainConfigDTO.GrainPurchaseTypeDTO](entities)), nil
}

func (s *GrainConfigService) UpdatePurchaseType(id uint, req *grainConfigDTO.GrainPurchaseTypeDTO) (*grainConfigDTO.GrainPurchaseTypeDTO, error) {
	entity, err := s.purchaseTypeRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	previousBase := entity.BaseEntity
	copier.Copy(entity, req)
	entity.Id = int(id)
	preserveBaseEntityFields(&entity.BaseEntity, previousBase)
	normalizeStatusAndOrder(&entity.Status, &entity.SortOrder)
	if entity.Unit == "" {
		entity.Unit = "公斤"
	}
	result, err := s.purchaseTypeRepository.SaveOrUpdate(entity)
	if err != nil {
		return nil, err
	}
	return db.ToDTO[grainConfigDTO.GrainPurchaseTypeDTO](result), nil
}

func (s *GrainConfigService) DeletePurchaseType(id uint) error {
	entity, err := s.purchaseTypeRepository.FindById(id)
	if err != nil {
		return err
	}
	if entity.Active == 0 {
		return gorm.ErrRecordNotFound
	}
	entity.Active = 0
	_, err = s.purchaseTypeRepository.SaveOrUpdate(entity)
	return err
}

func (s *GrainConfigService) ListPaymentMethods() ([]*grainConfigDTO.GrainPaymentMethodDTO, error) {
	entities, err := s.paymentMethodRepository.ListActive()
	if err != nil {
		return nil, err
	}
	return db.ToDTOs[grainConfigDTO.GrainPaymentMethodDTO](entities), nil
}

func (s *GrainConfigService) CreatePaymentMethod(req *grainConfigDTO.GrainPaymentMethodDTO) (*grainConfigDTO.GrainPaymentMethodDTO, error) {
	entity := db.ToPO[grainConfigRepository.GrainPaymentMethod](req)
	trimPaymentMethod(entity)
	if entity.MethodCode == "" {
		entity.MethodCode = generatePaymentMethodCode()
	}
	normalizeStatusAndOrder(&entity.Status, &entity.SortOrder)
	result, err := s.paymentMethodRepository.Create(entity)
	if err != nil {
		return nil, err
	}
	return db.ToDTO[grainConfigDTO.GrainPaymentMethodDTO](result), nil
}

func (s *GrainConfigService) ListPaymentMethodsPage(query grainConfigDTO.GrainConfigItemQueryDTO) (*baseDTO.PageDTO[grainConfigDTO.GrainPaymentMethodDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	total, err := s.paymentMethodRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	entities, err := s.paymentMethodRepository.ListByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), db.ToDTOs[grainConfigDTO.GrainPaymentMethodDTO](entities)), nil
}

func (s *GrainConfigService) UpdatePaymentMethod(id uint, req *grainConfigDTO.GrainPaymentMethodDTO) (*grainConfigDTO.GrainPaymentMethodDTO, error) {
	entity, err := s.paymentMethodRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	previousBase := entity.BaseEntity
	previousMethodCode := strings.TrimSpace(entity.MethodCode)
	copier.Copy(entity, req)
	entity.Id = int(id)
	preserveBaseEntityFields(&entity.BaseEntity, previousBase)
	trimPaymentMethod(entity)
	if entity.MethodCode == "" {
		entity.MethodCode = previousMethodCode
	}
	if entity.MethodCode == "" {
		entity.MethodCode = generatePaymentMethodCode()
	}
	normalizeStatusAndOrder(&entity.Status, &entity.SortOrder)
	result, err := s.paymentMethodRepository.SaveOrUpdate(entity)
	if err != nil {
		return nil, err
	}
	return db.ToDTO[grainConfigDTO.GrainPaymentMethodDTO](result), nil
}

func (s *GrainConfigService) DeletePaymentMethod(id uint) error {
	entity, err := s.paymentMethodRepository.FindById(id)
	if err != nil {
		return err
	}
	if entity.Active == 0 {
		return gorm.ErrRecordNotFound
	}
	entity.Active = 0
	_, err = s.paymentMethodRepository.SaveOrUpdate(entity)
	return err
}

func (s *GrainConfigService) ListPurchasePlaces(appUserID uint64) ([]*grainConfigDTO.GrainPurchasePlaceDTO, error) {
	entities, err := s.purchasePlaceRepository.ListActiveByAppUser(appUserID)
	if err != nil {
		return nil, err
	}
	return db.ToDTOs[grainConfigDTO.GrainPurchasePlaceDTO](entities), nil
}

func (s *GrainConfigService) CreatePurchasePlace(req *grainConfigDTO.GrainPurchasePlaceDTO) (*grainConfigDTO.GrainPurchasePlaceDTO, error) {
	entity := db.ToPO[grainConfigRepository.GrainPurchasePlace](req)
	normalizeStatusAndOrder(&entity.Status, &entity.SortOrder)
	result, err := s.purchasePlaceRepository.Create(entity)
	if err != nil {
		return nil, err
	}
	return db.ToDTO[grainConfigDTO.GrainPurchasePlaceDTO](result), nil
}

func (s *GrainConfigService) ListPurchasePlacesPage(query grainConfigDTO.GrainConfigItemQueryDTO) (*baseDTO.PageDTO[grainConfigDTO.GrainPurchasePlaceDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	total, err := s.purchasePlaceRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	entities, err := s.purchasePlaceRepository.ListByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), db.ToDTOs[grainConfigDTO.GrainPurchasePlaceDTO](entities)), nil
}

func (s *GrainConfigService) UpdatePurchasePlace(id uint, req *grainConfigDTO.GrainPurchasePlaceDTO) (*grainConfigDTO.GrainPurchasePlaceDTO, error) {
	entity, err := s.purchasePlaceRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	previousBase := entity.BaseEntity
	copier.Copy(entity, req)
	entity.Id = int(id)
	preserveBaseEntityFields(&entity.BaseEntity, previousBase)
	normalizeStatusAndOrder(&entity.Status, &entity.SortOrder)
	result, err := s.purchasePlaceRepository.SaveOrUpdate(entity)
	if err != nil {
		return nil, err
	}
	return db.ToDTO[grainConfigDTO.GrainPurchasePlaceDTO](result), nil
}

func (s *GrainConfigService) DeletePurchasePlace(id uint) error {
	entity, err := s.purchasePlaceRepository.FindById(id)
	if err != nil {
		return err
	}
	if entity.Active == 0 {
		return gorm.ErrRecordNotFound
	}
	entity.Active = 0
	_, err = s.purchasePlaceRepository.SaveOrUpdate(entity)
	return err
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

func normalizeStatusAndOrder(status *string, sortOrder *int) {
	if strings.TrimSpace(*status) == "" {
		*status = "active"
	}
	if *sortOrder < 0 {
		*sortOrder = 0
	}
}

func preserveBaseEntityFields(base *db.BaseEntity, previous db.BaseEntity) {
	base.Active = previous.Active
	base.CreatedTime = previous.CreatedTime
	base.CreatedBy = previous.CreatedBy
}

func generateStationCode() string {
	return fmt.Sprintf("GS%d", time.Now().UnixNano())
}

func trimPaymentMethod(entity *grainConfigRepository.GrainPaymentMethod) {
	entity.MethodCode = strings.TrimSpace(entity.MethodCode)
	entity.MethodName = strings.TrimSpace(entity.MethodName)
}

func generatePaymentMethodCode() string {
	return fmt.Sprintf("PM%d", time.Now().UnixNano())
}
