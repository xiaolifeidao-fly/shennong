package grain_farmer

import (
	baseDTO "common/base/dto"
	"common/middleware/db"
	grainFarmerDTO "service/grain_farmer/dto"
	grainFarmerRepository "service/grain_farmer/repository"
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
	if strings.TrimSpace(query.Search) != "" && farmerCryptoKey() != "" {
		query.SearchIDNumberDigest = dbFieldDigest(query.Search)
	}
	total, err := s.farmerRepository.CountByQuery(query)
	if err != nil {
		return nil, err
	}
	entities, err := s.farmerRepository.ListByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	if err := decryptFarmerEntities(entities); err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), db.ToDTOs[grainFarmerDTO.GrainFarmerDTO](entities)), nil
}

func (s *GrainFarmerService) CreateFarmer(req *grainFarmerDTO.GrainFarmerDTO) (*grainFarmerDTO.GrainFarmerDTO, error) {
	entity := db.ToPO[grainFarmerRepository.GrainFarmer](req)
	normalizeFarmer(entity)
	digest, err := farmerIDNumberDigest(entity.IDNumber)
	if err != nil {
		return nil, err
	}
	if existing, err := s.farmerRepository.FindActiveByIDNumberDigest(digest, entity.IDNumber, entity.StationID); err == nil {
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

func (s *GrainFarmerService) UpdateFarmer(id uint, req *grainFarmerDTO.GrainFarmerDTO) (*grainFarmerDTO.GrainFarmerDTO, error) {
	return s.updateFarmer(id, req, 0)
}

func (s *GrainFarmerService) UpdateFarmerInStation(id uint, req *grainFarmerDTO.GrainFarmerDTO, stationID uint64) (*grainFarmerDTO.GrainFarmerDTO, error) {
	req.StationID = stationID
	return s.updateFarmer(id, req, stationID)
}

func (s *GrainFarmerService) updateFarmer(id uint, req *grainFarmerDTO.GrainFarmerDTO, stationID uint64) (*grainFarmerDTO.GrainFarmerDTO, error) {
	entity, err := s.farmerRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 || (stationID > 0 && entity.StationID != stationID) {
		return nil, gorm.ErrRecordNotFound
	}
	copier.Copy(entity, req)
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
	entity, err := s.farmerRepository.FindById(id)
	if err != nil {
		return err
	}
	if entity.Active == 0 {
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
	if strings.TrimSpace(entity.Status) == "" {
		if strings.TrimSpace(entity.BankNumber) == "" {
			entity.Status = "missing-bank"
			entity.StatusText = "银行卡照片待补"
		} else {
			entity.Status = "complete"
			entity.StatusText = "资料完整"
		}
	}
}

func dbFieldDigest(value string) string {
	digest, err := farmerIDNumberDigest(value)
	if err != nil {
		return ""
	}
	return digest
}
