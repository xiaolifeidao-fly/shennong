package grain_purchase

import (
	baseDTO "common/base/dto"
	"common/middleware/db"
	"fmt"
	grainFarmerService "service/grain_farmer"
	grainFarmerRepository "service/grain_farmer/repository"
	grainPurchaseDTO "service/grain_purchase/dto"
	grainPurchaseRepository "service/grain_purchase/repository"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type GrainPurchaseService struct {
	farmerRepository         *grainFarmerRepository.GrainFarmerRepository
	entryRepository          *grainPurchaseRepository.GrainPurchaseEntryRepository
	snapshotRepository       *grainPurchaseRepository.GrainPurchaseEntrySnapshotRepository
	summaryRepository        *grainPurchaseRepository.GrainFarmerPurchaseSummaryRepository
	stationSummaryRepository *grainPurchaseRepository.GrainStationPurchaseSummaryRepository
	materialRepository       *grainPurchaseRepository.GrainEntryMaterialRepository
}

func NewGrainPurchaseService() *GrainPurchaseService {
	return &GrainPurchaseService{
		farmerRepository:         db.GetRepository[grainFarmerRepository.GrainFarmerRepository](),
		entryRepository:          db.GetRepository[grainPurchaseRepository.GrainPurchaseEntryRepository](),
		snapshotRepository:       db.GetRepository[grainPurchaseRepository.GrainPurchaseEntrySnapshotRepository](),
		summaryRepository:        db.GetRepository[grainPurchaseRepository.GrainFarmerPurchaseSummaryRepository](),
		stationSummaryRepository: db.GetRepository[grainPurchaseRepository.GrainStationPurchaseSummaryRepository](),
		materialRepository:       db.GetRepository[grainPurchaseRepository.GrainEntryMaterialRepository](),
	}
}

func (s *GrainPurchaseService) EnsureTable() error {
	steps := []func() error{
		s.entryRepository.EnsureTable,
		s.snapshotRepository.EnsureTable,
		s.summaryRepository.EnsureTable,
		s.stationSummaryRepository.EnsureTable,
		s.materialRepository.EnsureTable,
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func (s *GrainPurchaseService) ListEntries(query grainPurchaseDTO.GrainPurchaseEntryQueryDTO) (*baseDTO.PageDTO[grainPurchaseDTO.GrainPurchaseEntryDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	total, err := s.entryRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	entities, err := s.entryRepository.ListByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), db.ToDTOs[grainPurchaseDTO.GrainPurchaseEntryDTO](entities)), nil
}

func (s *GrainPurchaseService) CreateEntry(req *grainPurchaseDTO.GrainPurchaseEntryDTO, operatorAppUserID uint64, operatorName string) (*grainPurchaseDTO.GrainPurchaseEntryDTO, error) {
	entity := db.ToPO[grainPurchaseRepository.GrainPurchaseEntry](req)
	normalizeEntry(entity)
	entity.Version = 1
	var result *grainPurchaseRepository.GrainPurchaseEntry
	err := s.withTransaction(func(txService *GrainPurchaseService) error {
		var err error
		result, err = txService.entryRepository.Create(entity)
		if err != nil {
			return err
		}
		if err := txService.createEntrySnapshot(result, "create", operatorAppUserID, operatorName); err != nil {
			return err
		}
		return txService.applyEntryToSummary(result, 1)
	})
	if err != nil {
		return nil, err
	}
	return db.ToDTO[grainPurchaseDTO.GrainPurchaseEntryDTO](result), nil
}

func (s *GrainPurchaseService) UpdateEntry(id uint, req *grainPurchaseDTO.GrainPurchaseEntryDTO, operatorAppUserID uint64, operatorName string) (*grainPurchaseDTO.GrainPurchaseEntryDTO, error) {
	return s.updateEntry(id, req, operatorAppUserID, operatorName, 0)
}

func (s *GrainPurchaseService) UpdateEntryInStation(id uint, req *grainPurchaseDTO.GrainPurchaseEntryDTO, operatorAppUserID uint64, operatorName string, stationID uint64) (*grainPurchaseDTO.GrainPurchaseEntryDTO, error) {
	req.StationID = stationID
	return s.updateEntry(id, req, operatorAppUserID, operatorName, stationID)
}

func (s *GrainPurchaseService) updateEntry(id uint, req *grainPurchaseDTO.GrainPurchaseEntryDTO, operatorAppUserID uint64, operatorName string, stationID uint64) (*grainPurchaseDTO.GrainPurchaseEntryDTO, error) {
	var result *grainPurchaseRepository.GrainPurchaseEntry
	err := s.withTransaction(func(txService *GrainPurchaseService) error {
		entity, err := txService.entryRepository.FindById(id)
		if err != nil {
			return err
		}
		if entity.Active == 0 || (stationID > 0 && entity.StationID != stationID) {
			return gorm.ErrRecordNotFound
		}
		previous := *entity
		copier.Copy(entity, req)
		entity.Id = int(id)
		entity.Version++
		normalizeEntry(entity)
		result, err = txService.entryRepository.SaveOrUpdate(entity)
		if err != nil {
			return err
		}
		if err := txService.createEntrySnapshot(result, "update", operatorAppUserID, operatorName); err != nil {
			return err
		}
		if err := txService.applyEntryToSummary(&previous, -1); err != nil {
			return err
		}
		return txService.applyEntryToSummary(result, 1)
	})
	if err != nil {
		return nil, err
	}
	return db.ToDTO[grainPurchaseDTO.GrainPurchaseEntryDTO](result), nil
}

func (s *GrainPurchaseService) VoidEntry(id uint, operatorAppUserID uint64, operatorName string) error {
	return s.voidEntry(id, operatorAppUserID, operatorName, 0)
}

func (s *GrainPurchaseService) VoidEntryInStation(id uint, operatorAppUserID uint64, operatorName string, stationID uint64) error {
	return s.voidEntry(id, operatorAppUserID, operatorName, stationID)
}

func (s *GrainPurchaseService) voidEntry(id uint, operatorAppUserID uint64, operatorName string, stationID uint64) error {
	return s.withTransaction(func(txService *GrainPurchaseService) error {
		entity, err := txService.entryRepository.FindById(id)
		if err != nil {
			return err
		}
		if entity.Active == 0 || (stationID > 0 && entity.StationID != stationID) {
			return gorm.ErrRecordNotFound
		}
		previous := *entity
		entity.Status = "voided"
		entity.Version++
		if _, err := txService.entryRepository.SaveOrUpdate(entity); err != nil {
			return err
		}
		if err := txService.applyEntryToSummary(&previous, -1); err != nil {
			return err
		}
		return txService.createEntrySnapshot(entity, "void", operatorAppUserID, operatorName)
	})
}

func (s *GrainPurchaseService) DeleteEntry(id uint, operatorAppUserID uint64, operatorName string) error {
	return s.withTransaction(func(txService *GrainPurchaseService) error {
		entity, err := txService.entryRepository.FindById(id)
		if err != nil {
			return err
		}
		if entity.Active == 0 {
			return gorm.ErrRecordNotFound
		}
		previous := *entity
		entity.Active = 0
		entity.Version++
		if _, err := txService.entryRepository.SaveOrUpdate(entity); err != nil {
			return err
		}
		if err := txService.applyEntryToSummary(&previous, -1); err != nil {
			return err
		}
		return txService.createEntrySnapshot(entity, "delete", operatorAppUserID, operatorName)
	})
}

func (s *GrainPurchaseService) ListEntrySnapshots(query grainPurchaseDTO.GrainEntrySnapshotQueryDTO) (*baseDTO.PageDTO[grainPurchaseDTO.GrainPurchaseEntrySnapshotDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	total, err := s.snapshotRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	entities, err := s.snapshotRepository.ListByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	if err := decryptSnapshotFarmerFields(entities); err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), db.ToDTOs[grainPurchaseDTO.GrainPurchaseEntrySnapshotDTO](entities)), nil
}

func (s *GrainPurchaseService) ListFarmerPurchaseSummaries(query grainPurchaseDTO.GrainFarmerPurchaseSummaryQueryDTO) (*baseDTO.PageDTO[grainPurchaseDTO.GrainFarmerPurchaseSummaryDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	total, err := s.summaryRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	entities, err := s.summaryRepository.ListByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), db.ToDTOs[grainPurchaseDTO.GrainFarmerPurchaseSummaryDTO](entities)), nil
}

func (s *GrainPurchaseService) ListDailyFarmerSummaries(query grainPurchaseDTO.GrainFarmerDailySummaryQueryDTO) (*baseDTO.PageDTO[grainPurchaseDTO.GrainFarmerDailySummaryDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	applyTodayDefault(&query.StartDate, &query.EndDate)
	total, err := s.summaryRepository.CountDailyFarmerSummaries(query)
	if err != nil {
		return nil, err
	}
	summaries, err := s.summaryRepository.ListDailyFarmerSummaries(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	if err := decryptDailySummaryFarmerFields(summaries); err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), summaries), nil
}

func (s *GrainPurchaseService) ListMaterials(query grainPurchaseDTO.GrainEntryMaterialQueryDTO) (*baseDTO.PageDTO[grainPurchaseDTO.GrainEntryMaterialDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	total, err := s.materialRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	entities, err := s.materialRepository.ListByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), db.ToDTOs[grainPurchaseDTO.GrainEntryMaterialDTO](entities)), nil
}

func (s *GrainPurchaseService) CreateMaterial(req *grainPurchaseDTO.GrainEntryMaterialDTO) (*grainPurchaseDTO.GrainEntryMaterialDTO, error) {
	entity := db.ToPO[grainPurchaseRepository.GrainEntryMaterial](req)
	if strings.TrimSpace(entity.OssObjectKey) == "" && strings.TrimSpace(entity.OssURL) == "" {
		return nil, fmt.Errorf("oss object key or oss url is required")
	}
	if strings.TrimSpace(entity.MaterialBizType) == "" {
		entity.MaterialBizType = "entry"
	}
	result, err := s.materialRepository.Create(entity)
	if err != nil {
		return nil, err
	}
	return db.ToDTO[grainPurchaseDTO.GrainEntryMaterialDTO](result), nil
}

func (s *GrainPurchaseService) DeleteMaterial(id uint) error {
	entity, err := s.materialRepository.FindById(id)
	if err != nil {
		return err
	}
	if entity.Active == 0 {
		return gorm.ErrRecordNotFound
	}
	entity.Active = 0
	_, err = s.materialRepository.SaveOrUpdate(entity)
	return err
}

func (s *GrainPurchaseService) withTransaction(fn func(*GrainPurchaseService) error) error {
	if s.entryRepository == nil || s.entryRepository.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return s.entryRepository.Db.Transaction(func(tx *gorm.DB) error {
		return fn(s.withDB(tx))
	})
}

func (s *GrainPurchaseService) withDB(tx *gorm.DB) *GrainPurchaseService {
	txService := *s

	farmerRepository := *s.farmerRepository
	farmerRepository.SetDb(tx)
	txService.farmerRepository = &farmerRepository

	entryRepository := *s.entryRepository
	entryRepository.SetDb(tx)
	txService.entryRepository = &entryRepository

	snapshotRepository := *s.snapshotRepository
	snapshotRepository.SetDb(tx)
	txService.snapshotRepository = &snapshotRepository

	summaryRepository := *s.summaryRepository
	summaryRepository.SetDb(tx)
	txService.summaryRepository = &summaryRepository

	stationSummaryRepository := *s.stationSummaryRepository
	stationSummaryRepository.SetDb(tx)
	txService.stationSummaryRepository = &stationSummaryRepository

	materialRepository := *s.materialRepository
	materialRepository.SetDb(tx)
	txService.materialRepository = &materialRepository

	return &txService
}

func (s *GrainPurchaseService) createEntrySnapshot(entry *grainPurchaseRepository.GrainPurchaseEntry, action string, operatorAppUserID uint64, operatorName string) error {
	now := time.Now()
	snapshot := &grainPurchaseRepository.GrainPurchaseEntrySnapshot{
		EntryID:           uint64(entry.Id),
		SnapshotVersion:   entry.Version,
		SnapshotAction:    action,
		SnapshotTime:      &now,
		OperatorAppUserID: operatorAppUserID,
		OperatorName:      strings.TrimSpace(operatorName),
		StationID:         entry.StationID,
		AppUserID:         entry.AppUserID,
		FarmerID:          entry.FarmerID,
		PurchaseTypeID:    entry.PurchaseTypeID,
		Crop:              entry.Crop,
		Quantity:          entry.Quantity,
		Unit:              entry.Unit,
		Amount:            entry.Amount,
		UnitPrice:         entry.UnitPrice,
		BuyTime:           entry.BuyTime,
		PlaceID:           entry.PlaceID,
		Place:             entry.Place,
		LocationName:      entry.LocationName,
		LocationAddress:   entry.LocationAddress,
		Longitude:         entry.Longitude,
		Latitude:          entry.Latitude,
		Province:          entry.Province,
		City:              entry.City,
		District:          entry.District,
		PaymentMethodID:   entry.PaymentMethodID,
		PayType:           entry.PayType,
		EntryStatus:       entry.Status,
		EntryRemark:       entry.Remark,
	}
	if farmer, err := s.farmerRepository.FindById(uint(entry.FarmerID)); err == nil && farmer.Active == 1 {
		snapshot.FarmerName = farmer.Name
		snapshot.FarmerIDNumber = farmer.IDNumber
		snapshot.FarmerPhone = farmer.Phone
		snapshot.FarmerAddress = farmer.Address
		snapshot.FarmerBankNumber = farmer.BankNumber
		snapshot.FarmerBankName = farmer.BankName
	}
	_, err := s.snapshotRepository.Create(snapshot)
	return err
}

func (s *GrainPurchaseService) applyEntryToSummary(entry *grainPurchaseRepository.GrainPurchaseEntry, sign int) error {
	if entry == nil || entry.Active == 0 || entry.Status == "voided" || entry.Status == "deleted" {
		return nil
	}
	if sign == 0 {
		return nil
	}
	if err := s.applyEntryToFarmerSummary(entry, sign); err != nil {
		return err
	}
	return s.applyEntryToStationSummary(entry, sign)
}

func (s *GrainPurchaseService) applyEntryToFarmerSummary(entry *grainPurchaseRepository.GrainPurchaseEntry, sign int) error {
	summaryDate := summaryDay(entry.BuyTime)
	deltaCount := sign
	deltaAmount := float64(sign) * entry.Amount
	deltaQuantity := float64(sign) * entry.Quantity
	dimension := &grainPurchaseRepository.GrainFarmerPurchaseSummary{
		StationID:       entry.StationID,
		AppUserID:       entry.AppUserID,
		PurchaseTypeID:  entry.PurchaseTypeID,
		Crop:            entry.Crop,
		SummaryDate:     &summaryDate,
		FarmerID:        entry.FarmerID,
		PaymentMethodID: entry.PaymentMethodID,
		PayType:         entry.PayType,
		EntryCount:      deltaCount,
		TotalAmount:     deltaAmount,
		TotalQuantity:   deltaQuantity,
	}
	existing, err := s.summaryRepository.FindByDimension(dimension)
	if err == gorm.ErrRecordNotFound {
		if sign < 0 {
			return nil
		}
		_, err = s.summaryRepository.Create(dimension)
		return err
	}
	if err != nil {
		return err
	}
	existing.AppUserID = entry.AppUserID
	existing.Crop = entry.Crop
	existing.PayType = entry.PayType
	existing.EntryCount += deltaCount
	existing.TotalAmount += deltaAmount
	existing.TotalQuantity += deltaQuantity
	if existing.EntryCount < 0 {
		existing.EntryCount = 0
	}
	if existing.TotalAmount < 0 {
		existing.TotalAmount = 0
	}
	if existing.TotalQuantity < 0 {
		existing.TotalQuantity = 0
	}
	_, err = s.summaryRepository.SaveOrUpdate(existing)
	return err
}

func (s *GrainPurchaseService) applyEntryToStationSummary(entry *grainPurchaseRepository.GrainPurchaseEntry, sign int) error {
	summaryDate := summaryDay(entry.BuyTime)
	deltaCount := sign
	deltaAmount := float64(sign) * entry.Amount
	deltaQuantity := float64(sign) * entry.Quantity
	dimension := &grainPurchaseRepository.GrainStationPurchaseSummary{
		StationID:      entry.StationID,
		AppUserID:      entry.AppUserID,
		PurchaseTypeID: entry.PurchaseTypeID,
		Crop:           entry.Crop,
		SummaryDate:    &summaryDate,
		EntryCount:     deltaCount,
		TotalAmount:    deltaAmount,
		TotalQuantity:  deltaQuantity,
	}
	existing, err := s.stationSummaryRepository.FindByDimension(dimension)
	if err == gorm.ErrRecordNotFound {
		if sign < 0 {
			return nil
		}
		_, err = s.stationSummaryRepository.Create(dimension)
		return err
	}
	if err != nil {
		return err
	}
	existing.Crop = entry.Crop
	existing.EntryCount += deltaCount
	existing.TotalAmount += deltaAmount
	existing.TotalQuantity += deltaQuantity
	if existing.EntryCount < 0 {
		existing.EntryCount = 0
	}
	if existing.TotalAmount < 0 {
		existing.TotalAmount = 0
	}
	if existing.TotalQuantity < 0 {
		existing.TotalQuantity = 0
	}
	_, err = s.stationSummaryRepository.SaveOrUpdate(existing)
	return err
}

func summaryDay(value *time.Time) time.Time {
	source := time.Now()
	if value != nil {
		source = *value
	}
	location := grainBusinessLocation()
	source = source.In(location)
	return time.Date(source.Year(), source.Month(), source.Day(), 0, 0, 0, 0, location)
}

func applyTodayDefault(startDate, endDate **time.Time) {
	if *startDate != nil || *endDate != nil {
		return
	}
	today := summaryDay(nil)
	*startDate = &today
	*endDate = &today
}

func grainBusinessLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Local
	}
	return location
}

func decryptDailySummaryFarmerFields(summaries []*grainPurchaseDTO.GrainFarmerDailySummaryDTO) error {
	for _, summary := range summaries {
		name, idNumber, phone, address, bankNumber, bankName, err := grainFarmerService.DecryptFarmerProfileValues(
			summary.FarmerName,
			summary.FarmerIDNumber,
			summary.FarmerPhone,
			summary.FarmerAddress,
			summary.BankNumber,
			summary.BankName,
		)
		if err != nil {
			return err
		}
		summary.FarmerName = name
		summary.FarmerIDNumber = idNumber
		summary.FarmerPhone = phone
		summary.FarmerAddress = address
		summary.BankNumber = bankNumber
		summary.BankName = bankName
	}
	return nil
}

func decryptSnapshotFarmerFields(entities []*grainPurchaseRepository.GrainPurchaseEntrySnapshot) error {
	for _, entity := range entities {
		name, idNumber, phone, address, bankNumber, bankName, err := grainFarmerService.DecryptFarmerProfileValues(
			entity.FarmerName,
			entity.FarmerIDNumber,
			entity.FarmerPhone,
			entity.FarmerAddress,
			entity.FarmerBankNumber,
			entity.FarmerBankName,
		)
		if err != nil {
			return err
		}
		entity.FarmerName = name
		entity.FarmerIDNumber = idNumber
		entity.FarmerPhone = phone
		entity.FarmerAddress = address
		entity.FarmerBankNumber = bankNumber
		entity.FarmerBankName = bankName
	}
	return nil
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

func normalizeEntry(entity *grainPurchaseRepository.GrainPurchaseEntry) {
	if strings.TrimSpace(entity.Unit) == "" {
		entity.Unit = "公斤"
	}
	if entity.Quantity > 0 {
		entity.UnitPrice = entity.Amount / entity.Quantity
	}
	if strings.TrimSpace(entity.Status) == "" {
		entity.Status = "submitted"
	}
	if entity.Version <= 0 {
		entity.Version = 1
	}
}
