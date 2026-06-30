package grain_farmer

import (
	baseDTO "common/base/dto"
	"common/middleware/db"
	grainFarmerDTO "service/grain_farmer/dto"
	grainFarmerRepository "service/grain_farmer/repository"
	ocrDTO "service/ocr/dto"
	"strings"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type GrainFarmerService struct {
	farmerRepository *grainFarmerRepository.GrainFarmerRepository
}

func NewGrainFarmerService() *GrainFarmerService {
	return &GrainFarmerService{
		farmerRepository: db.GetRepository[grainFarmerRepository.GrainFarmerRepository](),
	}
}

func (s *GrainFarmerService) EnsureTable() error {
	return s.farmerRepository.EnsureTable()
}

func (s *GrainFarmerService) ListFarmers(query grainFarmerDTO.GrainFarmerQueryDTO) (*baseDTO.PageDTO[grainFarmerDTO.GrainFarmerDTO], error) {
	pageIndex, pageSize := normalizePage(query.Page, query.PageIndex, query.PageSize)
	PrepareFarmerQueryIndexes(&query)
	total, err := s.farmerRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	entities, err := s.farmerRepository.ListByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	if err := decryptFarmerListRows(entities); err != nil {
		return nil, err
	}
	entities = filterFarmerListRows(entities, query)
	return baseDTO.BuildPage(int(total), db.ToDTOs[grainFarmerDTO.GrainFarmerDTO](entities)), nil
}

func (s *GrainFarmerService) CreateFarmer(req *grainFarmerDTO.GrainFarmerDTO) (*grainFarmerDTO.GrainFarmerDTO, error) {
	return s.createFarmer(req, 0)
}

func (s *GrainFarmerService) CreateFarmerForAppUser(req *grainFarmerDTO.GrainFarmerDTO, appUserID uint64) (*grainFarmerDTO.GrainFarmerDTO, error) {
	req.AppUserID = appUserID
	return s.createFarmer(req, appUserID)
}

func (s *GrainFarmerService) createFarmer(req *grainFarmerDTO.GrainFarmerDTO, scopedAppUserID uint64) (*grainFarmerDTO.GrainFarmerDTO, error) {
	entity := db.ToPO[grainFarmerRepository.GrainFarmer](req)
	if scopedAppUserID > 0 {
		entity.AppUserID = scopedAppUserID
	}
	normalizeFarmer(entity)
	digest, err := farmerIDNumberDigest(entity.IDNumber)
	if err != nil {
		return nil, err
	}
	if existing, err := s.farmerRepository.FindActiveByIDNumberDigestForAppUser(digest, entity.IDNumber, entity.StationID, entity.AppUserID); err == nil {
		existingID := existing.Id
		existingBase := existing.BaseEntity
		copier.Copy(existing, entity)
		existing.BaseEntity = existingBase
		existing.Id = existingID
		normalizeFarmer(existing)
		if err := prepareFarmerForSave(existing); err != nil {
			return nil, err
		}
		result, err := s.farmerRepository.SaveOrUpdate(existing)
		if err != nil {
			return nil, err
		}
		dto := db.ToDTO[grainFarmerDTO.GrainFarmerDTO](result)
		if err := decryptFarmerDTO(dto); err != nil {
			return nil, err
		}
		return dto, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err := prepareFarmerForSave(entity); err != nil {
		return nil, err
	}
	result, err := s.farmerRepository.Create(entity)
	if err != nil {
		return nil, err
	}
	dto := db.ToDTO[grainFarmerDTO.GrainFarmerDTO](result)
	if err := decryptFarmerDTO(dto); err != nil {
		return nil, err
	}
	return dto, nil
}

func (s *GrainFarmerService) GetFarmer(id uint) (*grainFarmerDTO.GrainFarmerDTO, error) {
	entity, err := s.farmerRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	dto := db.ToDTO[grainFarmerDTO.GrainFarmerDTO](entity)
	if err := decryptFarmerDTO(dto); err != nil {
		return nil, err
	}
	return dto, nil
}

func (s *GrainFarmerService) UpdateFarmer(id uint, req *grainFarmerDTO.GrainFarmerDTO) (*grainFarmerDTO.GrainFarmerDTO, error) {
	return s.updateFarmer(id, req, 0, 0)
}

func (s *GrainFarmerService) UpdateFarmerInStation(id uint, req *grainFarmerDTO.GrainFarmerDTO, stationID uint64) (*grainFarmerDTO.GrainFarmerDTO, error) {
	req.StationID = stationID
	return s.updateFarmer(id, req, stationID, 0)
}

func (s *GrainFarmerService) UpdateFarmerInStationForAppUser(id uint, req *grainFarmerDTO.GrainFarmerDTO, stationID, appUserID uint64) (*grainFarmerDTO.GrainFarmerDTO, error) {
	req.StationID = stationID
	req.AppUserID = appUserID
	return s.updateFarmer(id, req, stationID, appUserID)
}

func (s *GrainFarmerService) UpdateFarmerCardInfoInStation(id uint, result *ocrDTO.RecognizeCardResult, stationID uint64) (*grainFarmerDTO.GrainFarmerDTO, error) {
	entity, err := s.farmerRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 || (stationID > 0 && entity.StationID != stationID) {
		return nil, gorm.ErrRecordNotFound
	}
	if err := decryptFarmerEntity(entity); err != nil {
		return nil, err
	}

	changed := false
	if result != nil {
		switch result.CardType {
		case "id-card":
			changed = assignNonEmpty(&entity.Name, result.Name) || changed
			changed = assignNonEmpty(&entity.IDNumber, result.IDNumber) || changed
			changed = assignNonEmpty(&entity.Address, result.Address) || changed
		case "bank-card", "social-security-card":
			changed = assignNonEmpty(&entity.BankNumber, result.BankNumber) || changed
			// 社保卡无开户行，收款人取社保卡上的持卡人姓名；银行卡沿用识别到的开户行
			payeeName := result.BankName
			if result.CardType == "social-security-card" {
				payeeName = result.Name
			}
			changed = assignNonEmpty(&entity.BankName, payeeName) || changed
		}
	}
	if !changed {
		dto := db.ToDTO[grainFarmerDTO.GrainFarmerDTO](entity)
		return dto, nil
	}

	normalizeFarmerStatus(entity)
	if err := prepareFarmerForSave(entity); err != nil {
		return nil, err
	}
	saved, err := s.farmerRepository.SaveOrUpdate(entity)
	if err != nil {
		return nil, err
	}
	dto := db.ToDTO[grainFarmerDTO.GrainFarmerDTO](saved)
	if err := decryptFarmerDTO(dto); err != nil {
		return nil, err
	}
	return dto, nil
}

func (s *GrainFarmerService) updateFarmer(id uint, req *grainFarmerDTO.GrainFarmerDTO, stationID, appUserID uint64) (*grainFarmerDTO.GrainFarmerDTO, error) {
	entity, err := s.farmerRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 || (stationID > 0 && entity.StationID != stationID) || (appUserID > 0 && entity.AppUserID != appUserID) {
		return nil, gorm.ErrRecordNotFound
	}
	previousBase := entity.BaseEntity
	copier.Copy(entity, req)
	preserveFarmerBaseEntityFields(&entity.BaseEntity, previousBase)
	entity.Id = int(id)
	normalizeFarmer(entity)
	if err := prepareFarmerForSave(entity); err != nil {
		return nil, err
	}
	result, err := s.farmerRepository.SaveOrUpdate(entity)
	if err != nil {
		return nil, err
	}
	dto := db.ToDTO[grainFarmerDTO.GrainFarmerDTO](result)
	if err := decryptFarmerDTO(dto); err != nil {
		return nil, err
	}
	return dto, nil
}

func (s *GrainFarmerService) DeleteFarmer(id uint) error {
	return s.deleteFarmer(id, 0, 0)
}

func (s *GrainFarmerService) DeleteFarmerInStationForAppUser(id uint, stationID, appUserID uint64) error {
	return s.deleteFarmer(id, stationID, appUserID)
}

func (s *GrainFarmerService) deleteFarmer(id uint, stationID, appUserID uint64) error {
	entity, err := s.farmerRepository.FindById(id)
	if err != nil {
		return err
	}
	if entity.Active == 0 || (stationID > 0 && entity.StationID != stationID) || (appUserID > 0 && entity.AppUserID != appUserID) {
		return gorm.ErrRecordNotFound
	}
	entity.Active = 0
	_, err = s.farmerRepository.SaveOrUpdate(entity)
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

func normalizeFarmer(entity *grainFarmerRepository.GrainFarmer) {
	entity.Name = strings.TrimSpace(entity.Name)
	entity.IDNumber = strings.TrimSpace(entity.IDNumber)
	entity.Phone = strings.TrimSpace(entity.Phone)
	entity.Address = strings.TrimSpace(entity.Address)
	entity.BankNumber = strings.TrimSpace(entity.BankNumber)
	entity.BankName = strings.TrimSpace(entity.BankName)
	normalizeFarmerStatus(entity)
}

func normalizeFarmerStatus(entity *grainFarmerRepository.GrainFarmer) {
	if strings.TrimSpace(entity.Status) == "" {
		if strings.TrimSpace(entity.BankNumber) == "" {
			entity.Status = "missing-bank"
			entity.StatusText = "银行卡照片待补"
		} else {
			entity.Status = "complete"
			entity.StatusText = "资料完整"
		}
	}
	if strings.TrimSpace(entity.BankNumber) != "" && entity.Status == "missing-bank" {
		entity.Status = "complete"
		entity.StatusText = "资料完整"
	}
}

func preserveFarmerBaseEntityFields(base *db.BaseEntity, previous db.BaseEntity) {
	base.Active = previous.Active
	base.CreatedTime = previous.CreatedTime
	base.CreatedBy = previous.CreatedBy
}

func assignNonEmpty(target *string, value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || *target == value {
		return false
	}
	*target = value
	return true
}

func filterFarmerListRows(rows []*grainFarmerRepository.GrainFarmerListRow, query grainFarmerDTO.GrainFarmerQueryDTO) []*grainFarmerRepository.GrainFarmerListRow {
	search := strings.TrimSpace(query.Search)
	name := strings.TrimSpace(query.Name)
	if search == "" && name == "" {
		return rows
	}
	filtered := make([]*grainFarmerRepository.GrainFarmerListRow, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		if search != "" && !farmerListRowMatchesKeyword(row, search) {
			continue
		}
		if name != "" && !FarmerNameMatches(row.Name, name) {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func farmerListRowMatchesKeyword(row *grainFarmerRepository.GrainFarmerListRow, keyword string) bool {
	if FarmerProfileMatchesKeyword(row.Name, row.IDNumber, row.Phone, keyword) {
		return true
	}
	keyword = strings.TrimSpace(keyword)
	return strings.Contains(row.Address, keyword) ||
		strings.Contains(row.BankNumber, keyword) ||
		strings.Contains(row.BankName, keyword) ||
		strings.Contains(row.StatusText, keyword) ||
		strings.Contains(row.Remark, keyword)
}
