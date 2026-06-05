package grain_config

import (
	baseDTO "common/base/dto"
	"common/middleware/db"
	"common/middleware/storage/image_source"
	"common/middleware/storage/oss"
	"fmt"
	"mime"
	"net/url"
	"path/filepath"
	grainConfigDTO "service/grain_config/dto"
	grainConfigRepository "service/grain_config/repository"
	tenantRepository "service/tenant/repository"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type GrainConfigService struct {
	stationRepository       *grainConfigRepository.GrainStationRepository
	stationUserRepository   *grainConfigRepository.GrainStationUserRepository
	stationExtraRepository  *grainConfigRepository.GrainStationExtraRepository
	purchaseTypeRepository  *grainConfigRepository.GrainPurchaseTypeRepository
	paymentMethodRepository *grainConfigRepository.GrainPaymentMethodRepository
	purchasePlaceRepository *grainConfigRepository.GrainPurchasePlaceRepository
}

type BusinessLicenseContent struct {
	StationID uint64
	Data      []byte
	MimeType  string
}

func NewGrainConfigService() *GrainConfigService {
	return &GrainConfigService{
		stationRepository:       db.GetRepository[grainConfigRepository.GrainStationRepository](),
		stationUserRepository:   db.GetRepository[grainConfigRepository.GrainStationUserRepository](),
		stationExtraRepository:  db.GetRepository[grainConfigRepository.GrainStationExtraRepository](),
		purchaseTypeRepository:  db.GetRepository[grainConfigRepository.GrainPurchaseTypeRepository](),
		paymentMethodRepository: db.GetRepository[grainConfigRepository.GrainPaymentMethodRepository](),
		purchasePlaceRepository: db.GetRepository[grainConfigRepository.GrainPurchasePlaceRepository](),
	}
}

func (s *GrainConfigService) EnsureTable() error {
	steps := []func() error{
		s.stationRepository.EnsureTable,
		s.stationUserRepository.EnsureTable,
		s.stationExtraRepository.EnsureTable,
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

func (s *GrainConfigService) GetStationDetail(stationID uint64) (*grainConfigDTO.GrainStationDetailDTO, error) {
	if stationID == 0 {
		return nil, fmt.Errorf("stationId is required")
	}
	station, err := s.stationRepository.FindById(uint(stationID))
	if err != nil {
		return nil, err
	}
	if station == nil || station.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	detail := &grainConfigDTO.GrainStationDetailDTO{
		BaseDTO: baseDTO.BaseDTO{
			Id:          station.Id,
			Active:      station.Active,
			CreatedTime: station.CreatedTime,
			CreatedBy:   station.CreatedBy,
			UpdatedTime: station.UpdatedTime,
			UpdatedBy:   station.UpdatedBy,
		},
		StationName:  station.StationName,
		StationCode:  station.StationCode,
		TenantID:     station.TenantID,
		ContactName:  station.ContactName,
		ContactPhone: station.ContactPhone,
		Province:     station.Province,
		City:         station.City,
		District:     station.District,
		Address:      station.Address,
		Longitude:    station.Longitude,
		Latitude:     station.Latitude,
		Status:       station.Status,
		Remark:       station.Remark,
	}
	extra, err := s.stationExtraRepository.FindByStationID(stationID)
	if err == nil && extra != nil {
		detail.BusinessLicenseUrl = businessLicenseDisplayURL(stationID, extra.BusinessLicenseKey, extra.BusinessLicenseUrl, extra.WXCloudURL, extra.LastSource, extra.UpdatedTime)
		detail.LastSource = extra.LastSource
		detail.WXCloudURL = extra.WXCloudURL
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if station.TenantID > 0 {
		tenantRepo := db.GetRepository[tenantRepository.TenantRepository]()
		tenant, tenantErr := tenantRepo.FindById(uint(station.TenantID))
		if tenantErr == nil && tenant != nil && tenant.Active == 1 {
			detail.TenantName = tenant.TenantName
		}
	}
	return detail, nil
}

func (s *GrainConfigService) GetStationExtra(stationID uint64) (*grainConfigDTO.GrainStationExtraDTO, error) {
	entity, err := s.stationExtraRepository.FindByStationID(stationID)
	if err == gorm.ErrRecordNotFound {
		return &grainConfigDTO.GrainStationExtraDTO{StationID: stationID}, nil
	}
	if err != nil {
		return nil, err
	}
	return &grainConfigDTO.GrainStationExtraDTO{
		StationID:                entity.StationID,
		AccountHolderName:        entity.AccountHolderName,
		BankName:                 entity.BankName,
		BankAccountNumber:        entity.BankAccountNumber,
		BusinessLicenseUrl:       businessLicenseDisplayURL(entity.StationID, entity.BusinessLicenseKey, entity.BusinessLicenseUrl, entity.WXCloudURL, entity.LastSource, entity.UpdatedTime),
		BusinessLicenseKey:       entity.BusinessLicenseKey,
		LastSource:               entity.LastSource,
		WXCloudURL:               entity.WXCloudURL,
		BusinessLicenseUpdatedAt: businessLicenseUpdatedAt(entity.UpdatedTime),
	}, nil
}

func (s *GrainConfigService) SaveStationExtra(stationID uint64, req *grainConfigDTO.GrainStationExtraDTO) (*grainConfigDTO.GrainStationExtraDTO, error) {
	entity := &grainConfigRepository.GrainStationExtra{
		StationID:          stationID,
		AccountHolderName:  strings.TrimSpace(req.AccountHolderName),
		BankName:           strings.TrimSpace(req.BankName),
		BankAccountNumber:  strings.TrimSpace(req.BankAccountNumber),
		BusinessLicenseUrl: strings.TrimSpace(req.BusinessLicenseUrl),
		BusinessLicenseKey: strings.TrimSpace(req.BusinessLicenseKey),
		WXCloudURL:         strings.TrimSpace(req.WXCloudURL),
	}
	if entity.BusinessLicenseUrl != "" || entity.BusinessLicenseKey != "" || entity.WXCloudURL != "" {
		entity.LastSource = normalizeImageSource(req.LastSource)
	}
	result, err := s.stationExtraRepository.Upsert(entity)
	if err != nil {
		return nil, err
	}
	return &grainConfigDTO.GrainStationExtraDTO{
		StationID:                result.StationID,
		AccountHolderName:        result.AccountHolderName,
		BankName:                 result.BankName,
		BankAccountNumber:        result.BankAccountNumber,
		BusinessLicenseUrl:       businessLicenseDisplayURL(result.StationID, result.BusinessLicenseKey, result.BusinessLicenseUrl, result.WXCloudURL, result.LastSource, result.UpdatedTime),
		BusinessLicenseKey:       result.BusinessLicenseKey,
		LastSource:               result.LastSource,
		WXCloudURL:               result.WXCloudURL,
		BusinessLicenseUpdatedAt: businessLicenseUpdatedAt(result.UpdatedTime),
	}, nil
}

func (s *GrainConfigService) SaveBusinessLicense(stationID uint64, businessLicenseURL, businessLicenseKey string) (*grainConfigDTO.GrainStationExtraDTO, error) {
	req := &grainConfigDTO.GrainStationExtraDTO{
		StationID:          stationID,
		BusinessLicenseUrl: strings.TrimSpace(businessLicenseURL),
		BusinessLicenseKey: strings.TrimSpace(businessLicenseKey),
		LastSource:         image_source.OSS,
	}
	existing, err := s.stationExtraRepository.FindByStationID(stationID)
	if err == nil && existing != nil {
		req.AccountHolderName = existing.AccountHolderName
		req.BankName = existing.BankName
		req.BankAccountNumber = existing.BankAccountNumber
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return s.SaveStationExtra(stationID, req)
}

func (s *GrainConfigService) GetBusinessLicenseContent(stationID uint64) (*BusinessLicenseContent, error) {
	entity, err := s.stationExtraRepository.FindByStationID(stationID)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	data, err := getOssObject(entity.BusinessLicenseKey, entity.BusinessLicenseUrl)
	if err != nil {
		return nil, err
	}
	return &BusinessLicenseContent{
		StationID: stationID,
		Data:      data,
		MimeType:  detectBusinessLicenseMimeType(data, entity.BusinessLicenseKey, entity.BusinessLicenseUrl),
	}, nil
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

func businessLicensePath(stationID uint64, ossObjectKey, fallbackURL string, updatedTime time.Time) string {
	if strings.TrimSpace(ossObjectKey) == "" && strings.TrimSpace(fallbackURL) == "" {
		return ""
	}
	return fmt.Sprintf("/grain-stations/%d/extra/business-license", stationID)
}

func businessLicenseDisplayURL(stationID uint64, ossObjectKey, fallbackURL, wxCloudURL, lastSource string, updatedTime time.Time) string {
	if image_source.IsWXCloud() && normalizeImageSource(lastSource) == image_source.WXCloud && strings.TrimSpace(wxCloudURL) != "" {
		return strings.TrimSpace(wxCloudURL)
	}
	return businessLicensePath(stationID, ossObjectKey, fallbackURL, updatedTime)
}

func normalizeImageSource(source string) string {
	if strings.EqualFold(strings.TrimSpace(source), image_source.WXCloud) {
		return image_source.WXCloud
	}
	return image_source.OSS
}

func businessLicenseUpdatedAt(updatedTime time.Time) int64 {
	if updatedTime.IsZero() {
		return 0
	}
	return updatedTime.UnixMilli()
}

func getOssObject(ossObjectKey, fallbackURL string) ([]byte, error) {
	key := strings.TrimSpace(ossObjectKey)
	if key == "" {
		key = parseOssKeyFromURL(fallbackURL)
	}
	if key == "" {
		return nil, fmt.Errorf("oss object key is empty")
	}
	if data, err := oss.GetByKey(key); err == nil {
		return data, nil
	}
	return oss.Get(key)
}

func parseOssKeyFromURL(rawURL string) string {
	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsedURL.Path == "" {
		return ""
	}
	return strings.TrimLeft(parsedURL.Path, "/")
}

func detectBusinessLicenseMimeType(data []byte, objectKey, fallbackURL string) string {
	name := strings.TrimSpace(objectKey)
	if name == "" {
		name = parseOssKeyFromURL(fallbackURL)
	}
	if ext := strings.ToLower(filepath.Ext(name)); ext != "" {
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			return mimeType
		}
	}
	if len(data) >= 4 && data[0] == 0x25 && data[1] == 0x50 && data[2] == 0x44 && data[3] == 0x46 {
		return "application/pdf"
	}
	if len(data) >= 3 && data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff {
		return "image/jpeg"
	}
	if len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4e && data[3] == 0x47 {
		return "image/png"
	}
	return "application/octet-stream"
}
