package grain_purchase

import (
	baseDTO "common/base/dto"
	"common/middleware/db"
	"common/middleware/storage/image_source"
	"common/middleware/storage/oss"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
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

type GrainEntryMaterialContent struct {
	Data      []byte
	MimeType  string
	FileName  string
	StationID uint64
	Base64    string
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
	dtos, err := s.entryRepository.ListDTOByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	if err := s.enrichEntryFarmerProfiles(dtos); err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), dtos), nil
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

func (s *GrainPurchaseService) UpdateEntryInStationForAppUser(id uint, req *grainPurchaseDTO.GrainPurchaseEntryDTO, operatorAppUserID uint64, operatorName string, stationID uint64) (*grainPurchaseDTO.GrainPurchaseEntryDTO, error) {
	req.StationID = stationID
	req.AppUserID = operatorAppUserID
	return s.updateEntryForAppUser(id, req, operatorAppUserID, operatorName, stationID, operatorAppUserID)
}

func (s *GrainPurchaseService) updateEntry(id uint, req *grainPurchaseDTO.GrainPurchaseEntryDTO, operatorAppUserID uint64, operatorName string, stationID uint64) (*grainPurchaseDTO.GrainPurchaseEntryDTO, error) {
	return s.updateEntryForAppUser(id, req, operatorAppUserID, operatorName, stationID, 0)
}

func (s *GrainPurchaseService) updateEntryForAppUser(id uint, req *grainPurchaseDTO.GrainPurchaseEntryDTO, operatorAppUserID uint64, operatorName string, stationID, ownerAppUserID uint64) (*grainPurchaseDTO.GrainPurchaseEntryDTO, error) {
	var result *grainPurchaseRepository.GrainPurchaseEntry
	err := s.withTransaction(func(txService *GrainPurchaseService) error {
		entity, err := txService.entryRepository.FindById(id)
		if err != nil {
			return err
		}
		if entity.Active == 0 || (stationID > 0 && entity.StationID != stationID) || (ownerAppUserID > 0 && entity.AppUserID != ownerAppUserID) {
			return gorm.ErrRecordNotFound
		}
		previous := *entity
		previousBase := entity.BaseEntity
		copier.Copy(entity, req)
		preserveEntryBaseEntityFields(&entity.BaseEntity, previousBase)
		entity.Id = int(id)
		normalizeEntry(entity)
		if entryBusinessEqual(&previous, entity) {
			result = entity
			return nil
		}
		entity.Version++
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

func (s *GrainPurchaseService) VoidEntryInStationForAppUser(id uint, operatorAppUserID uint64, operatorName string, stationID uint64) error {
	return s.voidEntryForAppUser(id, operatorAppUserID, operatorName, stationID, operatorAppUserID)
}

func (s *GrainPurchaseService) voidEntry(id uint, operatorAppUserID uint64, operatorName string, stationID uint64) error {
	return s.voidEntryForAppUser(id, operatorAppUserID, operatorName, stationID, 0)
}

func (s *GrainPurchaseService) voidEntryForAppUser(id uint, operatorAppUserID uint64, operatorName string, stationID, ownerAppUserID uint64) error {
	return s.withTransaction(func(txService *GrainPurchaseService) error {
		entity, err := txService.entryRepository.FindById(id)
		if err != nil {
			return err
		}
		if entity.Active == 0 || (stationID > 0 && entity.StationID != stationID) || (ownerAppUserID > 0 && entity.AppUserID != ownerAppUserID) {
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

func (s *GrainPurchaseService) GetDashboard(query grainPurchaseDTO.GrainPurchaseDashboardQueryDTO) (*grainPurchaseDTO.GrainPurchaseDashboardDTO, error) {
	applyTodayDefault(&query.StartDate, &query.EndDate)
	if s.stationSummaryRepository == nil || s.stationSummaryRepository.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	overview, err := s.stationSummaryRepository.DashboardOverview(query)
	if err != nil {
		return nil, err
	}
	newFarmerCount, err := s.summaryRepository.DashboardNewFarmerCount(query)
	if err != nil {
		return nil, err
	}
	overview.NewFarmerCount = newFarmerCount
	if overview.NewFarmerCount > 0 {
		overview.AverageFarmerDeal = overview.TotalAmount / float64(overview.NewFarmerCount)
	}
	byStation, err := s.stationSummaryRepository.DashboardByStation(query)
	if err != nil {
		return nil, err
	}
	stationFarmerCount, err := s.summaryRepository.DashboardFarmerCountByStation(query)
	if err != nil {
		return nil, err
	}
	byCrop, err := s.stationSummaryRepository.DashboardByCrop(query)
	if err != nil {
		return nil, err
	}
	cropFarmerCount, err := s.summaryRepository.DashboardFarmerCountByCrop(query)
	if err != nil {
		return nil, err
	}
	enrichDashboardRows(byStation, overview, func(row *grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO) int {
		return stationFarmerCount[row.StationID]
	})
	enrichDashboardRows(byCrop, overview, func(row *grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO) int {
		return cropFarmerCount[row.Name]
	})
	now := time.Now()
	return &grainPurchaseDTO.GrainPurchaseDashboardDTO{
		StartDate: formatDashboardDate(query.StartDate),
		EndDate:   formatDashboardDate(query.EndDate),
		StationID: query.StationID,
		Overview:  *overview,
		ByStation: byStation,
		ByCrop:    byCrop,
		Generated: &now,
	}, nil
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
	dtos := db.ToDTOs[grainPurchaseDTO.GrainEntryMaterialDTO](entities)
	for _, dto := range dtos {
		if dto != nil && dto.Id > 0 {
			dto.ImageURL = fmt.Sprintf("/grain-entry-materials?imageId=%d", dto.Id)
			if image_source.IsWXCloud() && normalizeImageSource(dto.LastSource) == image_source.WXCloud && strings.TrimSpace(dto.WXCloudURL) != "" {
				dto.ImageURL = dto.WXCloudURL
			}
		}
	}
	return baseDTO.BuildPage(int(total), dtos), nil
}

func (s *GrainPurchaseService) CreateMaterial(req *grainPurchaseDTO.GrainEntryMaterialDTO) (*grainPurchaseDTO.GrainEntryMaterialDTO, error) {
	entity := db.ToPO[grainPurchaseRepository.GrainEntryMaterial](req)
	if strings.TrimSpace(entity.OssObjectKey) == "" && strings.TrimSpace(entity.OssURL) == "" {
		return nil, fmt.Errorf("oss object key or oss url is required")
	}
	if entity.EntryID == 0 {
		return nil, fmt.Errorf("entry_id is required")
	}
	if strings.TrimSpace(entity.MaterialBizType) == "" {
		entity.MaterialBizType = "entry"
	}
	entity.LastSource = normalizeImageSource(entity.LastSource)
	if strings.TrimSpace(entity.ImageHash) == "" {
		entity.ImageHash = hashImageName(entity.FileName)
	}
	var result *grainPurchaseRepository.GrainEntryMaterial
	err := s.withTransaction(func(txService *GrainPurchaseService) error {
		var created bool
		var err error
		result, created, err = txService.materialRepository.FindOrCreate(entity)
		if err != nil || !created {
			if err == nil && result != nil && entity.WXCloudURL != "" && (result.WXCloudURL != entity.WXCloudURL || result.LastSource != entity.LastSource) {
				if updateErr := txService.materialRepository.UpdateCloudSource(result.Id, entity.WXCloudURL, entity.LastSource); updateErr != nil {
					return updateErr
				}
				result.WXCloudURL = entity.WXCloudURL
				result.LastSource = entity.LastSource
			}
			return err
		}
		return txService.createMaterialChangeSnapshot(result.EntryID, "material_create")
	})
	if err != nil {
		return nil, err
	}
	return db.ToDTO[grainPurchaseDTO.GrainEntryMaterialDTO](result), nil
}

func hashImageName(name string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(name)))
}

func (s *GrainPurchaseService) enrichEntryFarmerProfiles(entries []*grainPurchaseDTO.GrainPurchaseEntryDTO) error {
	for _, entry := range entries {
		if entry == nil || entry.FarmerID == 0 {
			continue
		}
		farmer, err := s.farmerRepository.FindById(uint(entry.FarmerID))
		if err == gorm.ErrRecordNotFound {
			continue
		}
		if err != nil {
			return err
		}
		if farmer.Active == 0 {
			continue
		}
		name, idNumber, phone, address, bankNumber, bankName, err := grainFarmerService.DecryptFarmerProfileValues(
			farmer.Name,
			farmer.IDNumber,
			farmer.Phone,
			farmer.Address,
			farmer.BankNumber,
			farmer.BankName,
		)
		if err != nil {
			return err
		}
		entry.FarmerName = name
		entry.FarmerIDNumber = idNumber
		entry.FarmerPhone = phone
		entry.FarmerAddress = address
		entry.FarmerBankNumber = bankNumber
		entry.FarmerBankName = bankName
	}
	return nil
}

func (s *GrainPurchaseService) DeleteMaterial(id uint) error {
	return s.withTransaction(func(txService *GrainPurchaseService) error {
		entity, err := txService.materialRepository.FindById(id)
		if err != nil {
			return err
		}
		if entity.Active == 0 {
			return gorm.ErrRecordNotFound
		}
		entryID := entity.EntryID
		if err = txService.materialRepository.DeletePhysicalByID(id); err != nil {
			return err
		}
		return txService.createMaterialChangeSnapshot(entryID, "material_delete")
	})
}

func (s *GrainPurchaseService) GetMaterialImageContent(id uint) (*GrainEntryMaterialContent, error) {
	log.Printf("[grain-material-image] get material image content start id=%d", id)
	entity, err := s.materialRepository.FindById(id)
	if err != nil {
		log.Printf("[grain-material-image] find material failed id=%d err=%v", id, err)
		return nil, err
	}
	if entity.Active == 0 {
		log.Printf("[grain-material-image] material inactive id=%d entryID=%d stationID=%d", id, entity.EntryID, entity.StationID)
		return nil, gorm.ErrRecordNotFound
	}
	log.Printf("[grain-material-image] material record found id=%d entryID=%d stationID=%d appUserID=%d materialBizType=%s materialType=%s fileName=%s mimeType=%s ossBucket=%s ossObjectKey=%s fallbackURL=%s", id, entity.EntryID, entity.StationID, entity.AppUserID, entity.MaterialBizType, entity.MaterialType, entity.FileName, entity.MimeType, entity.OssBucket, entity.OssObjectKey, safeURLForLog(entity.OssURL))
	data, err := getOssObject(entity.OssObjectKey, entity.OssURL)
	if err != nil {
		log.Printf("[grain-material-image] get oss object failed id=%d entryID=%d stationID=%d fileName=%s ossObjectKey=%s fallbackURL=%s err=%v", id, entity.EntryID, entity.StationID, entity.FileName, entity.OssObjectKey, safeURLForLog(entity.OssURL), err)
		return nil, err
	}
	mimeType := strings.TrimSpace(entity.MimeType)
	if !strings.HasPrefix(mimeType, "image/") {
		mimeType = detectImageMimeType(data, entity.FileName)
	}
	base64Content := ""
	if image_source.IsWXCloud() && normalizeImageSource(entity.LastSource) == image_source.OSS {
		base64Content = base64.StdEncoding.EncodeToString(data)
	}
	log.Printf("[grain-material-image] get material image content success id=%d entryID=%d stationID=%d fileName=%s mimeType=%s bytes=%d", id, entity.EntryID, entity.StationID, entity.FileName, mimeType, len(data))
	return &GrainEntryMaterialContent{
		Data:      data,
		MimeType:  mimeType,
		FileName:  entity.FileName,
		StationID: entity.StationID,
		Base64:    base64Content,
	}, nil
}

func normalizeImageSource(source string) string {
	if strings.EqualFold(strings.TrimSpace(source), image_source.WXCloud) {
		return image_source.WXCloud
	}
	return image_source.OSS
}

func getOssObject(ossObjectKey, fallbackURL string) ([]byte, error) {
	key := strings.TrimSpace(ossObjectKey)
	if key == "" {
		key = objectKeyFromURL(fallbackURL)
		log.Printf("[grain-material-image] oss object key empty, parsed key from fallbackURL parsedKey=%s fallbackURL=%s", key, safeURLForLog(fallbackURL))
	}
	if key == "" {
		return nil, fmt.Errorf("oss object key is empty")
	}
	if data, err := oss.GetByKey(key); err == nil {
		log.Printf("[grain-material-image] oss get by key success key=%s bytes=%d", key, len(data))
		return data, nil
	} else {
		log.Printf("[grain-material-image] oss get by key failed key=%s err=%v", key, err)
	}
	data, err := oss.Get(key)
	if err != nil {
		log.Printf("[grain-material-image] oss get by path failed path=%s err=%v", key, err)
		return nil, err
	}
	log.Printf("[grain-material-image] oss get by path success path=%s bytes=%d", key, len(data))
	return data, nil
}

func objectKeyFromURL(rawURL string) string {
	value := strings.TrimSpace(rawURL)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Path == "" {
		return ""
	}
	return strings.TrimLeft(parsed.Path, "/")
}

func safeURLForLog(rawURL string) string {
	value := strings.TrimSpace(rawURL)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "<invalid-url>"
	}
	hasQuery := parsed.RawQuery != ""
	if parsed.Scheme == "" && parsed.Host == "" {
		return fmt.Sprintf("path=%s hasQuery=%t", parsed.Path, hasQuery)
	}
	return fmt.Sprintf("scheme=%s host=%s path=%s hasQuery=%t", parsed.Scheme, parsed.Host, parsed.Path, hasQuery)
}

func detectImageMimeType(data []byte, fileName string) string {
	if ext := strings.ToLower(filepath.Ext(fileName)); ext != "" {
		if mimeType := mime.TypeByExtension(ext); strings.HasPrefix(mimeType, "image/") {
			return mimeType
		}
	}
	if len(data) > 0 {
		return http.DetectContentType(data)
	}
	return "application/octet-stream"
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
	materialCount, materialDigest, materialSummary, err := s.entryMaterialSnapshot(entry.Id)
	if err != nil {
		return err
	}
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
		PayTime:           entry.PayTime,
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
		MaterialCount:     materialCount,
		MaterialDigest:    materialDigest,
		MaterialSummary:   materialSummary,
	}
	if farmer, err := s.farmerRepository.FindById(uint(entry.FarmerID)); err == nil && farmer.Active == 1 {
		snapshot.FarmerName = farmer.Name
		snapshot.FarmerIDNumber = farmer.IDNumber
		snapshot.FarmerPhone = farmer.Phone
		snapshot.FarmerAddress = farmer.Address
		snapshot.FarmerBankNumber = farmer.BankNumber
		snapshot.FarmerBankName = farmer.BankName
	}
	_, err = s.snapshotRepository.Create(snapshot)
	return err
}

func (s *GrainPurchaseService) createMaterialChangeSnapshot(entryID uint64, action string) error {
	entry, err := s.entryRepository.FindById(uint(entryID))
	if err != nil {
		return err
	}
	if entry.Active == 0 {
		return gorm.ErrRecordNotFound
	}
	entry.Version++
	if _, err = s.entryRepository.SaveOrUpdate(entry); err != nil {
		return err
	}
	return s.createEntrySnapshot(entry, action, entry.AppUserID, "")
}

type entryMaterialSnapshotItem struct {
	ID              int    `json:"id"`
	MaterialBizType string `json:"materialBizType"`
	MaterialType    string `json:"materialType"`
	FileName        string `json:"fileName"`
	ImageHash       string `json:"imageHash"`
	SortOrder       int    `json:"sortOrder"`
}

type entryMaterialSnapshotSummary struct {
	Count int                         `json:"count"`
	Items []entryMaterialSnapshotItem `json:"items"`
}

func (s *GrainPurchaseService) entryMaterialSnapshot(entryID int) (int, string, string, error) {
	materials, err := s.materialRepository.ListActiveByEntryID(uint64(entryID))
	if err != nil {
		return 0, "", "", err
	}
	summary := entryMaterialSnapshotSummary{Count: len(materials), Items: make([]entryMaterialSnapshotItem, 0, len(materials))}
	hash := sha256.New()
	for _, material := range materials {
		if material == nil {
			continue
		}
		item := entryMaterialSnapshotItem{
			ID:              material.Id,
			MaterialBizType: strings.TrimSpace(material.MaterialBizType),
			MaterialType:    strings.TrimSpace(material.MaterialType),
			FileName:        strings.TrimSpace(material.FileName),
			ImageHash:       strings.TrimSpace(material.ImageHash),
			SortOrder:       material.SortOrder,
		}
		summary.Items = append(summary.Items, item)
		fmt.Fprintf(hash, "%d|%s|%s|%s|%s|%d\n", item.ID, item.MaterialBizType, item.MaterialType, item.FileName, item.ImageHash, item.SortOrder)
	}
	data, err := json.Marshal(summary)
	if err != nil {
		return 0, "", "", err
	}
	return summary.Count, fmt.Sprintf("%x", hash.Sum(nil)), string(data), nil
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

func enrichDashboardRows(rows []*grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO, overview *grainPurchaseDTO.GrainPurchaseDashboardMetricDTO, farmerCount func(*grainPurchaseDTO.GrainPurchaseDashboardDimensionDTO) int) {
	for _, row := range rows {
		if row == nil {
			continue
		}
		if strings.TrimSpace(row.Key) == "" {
			row.Key = row.Name
		}
		row.FarmerCount = farmerCount(row)
		if row.TotalQuantity > 0 {
			row.AverageUnitPrice = row.TotalAmount / row.TotalQuantity
		}
		if overview.TotalAmount > 0 {
			row.AmountShare = row.TotalAmount / overview.TotalAmount
		}
		if overview.TotalQuantity > 0 {
			row.QuantityShare = row.TotalQuantity / overview.TotalQuantity
		}
	}
}

func formatDashboardDate(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.In(grainBusinessLocation()).Format("2006-01-02")
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
	if strings.TrimSpace(entity.DisplayUnit) == "" {
		entity.DisplayUnit = "公斤"
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

func entryBusinessEqual(left, right *grainPurchaseRepository.GrainPurchaseEntry) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.StationID == right.StationID &&
		left.AppUserID == right.AppUserID &&
		left.FarmerID == right.FarmerID &&
		left.PurchaseTypeID == right.PurchaseTypeID &&
		left.Crop == right.Crop &&
		left.Quantity == right.Quantity &&
		left.Unit == right.Unit &&
		left.DisplayUnit == right.DisplayUnit &&
		left.Amount == right.Amount &&
		left.UnitPrice == right.UnitPrice &&
		timesEqual(left.BuyTime, right.BuyTime) &&
		timesEqual(left.PayTime, right.PayTime) &&
		left.PlaceID == right.PlaceID &&
		left.Place == right.Place &&
		left.LocationName == right.LocationName &&
		left.LocationAddress == right.LocationAddress &&
		left.Longitude == right.Longitude &&
		left.Latitude == right.Latitude &&
		left.Province == right.Province &&
		left.City == right.City &&
		left.District == right.District &&
		left.PaymentMethodID == right.PaymentMethodID &&
		left.PayType == right.PayType &&
		left.Status == right.Status &&
		left.Remark == right.Remark
}

func timesEqual(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Equal(*right)
}

func preserveEntryBaseEntityFields(base *db.BaseEntity, previous db.BaseEntity) {
	base.Active = previous.Active
	base.CreatedTime = previous.CreatedTime
	base.CreatedBy = previous.CreatedBy
}
