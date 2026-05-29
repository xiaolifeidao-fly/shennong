package region

import (
	"common/middleware/db"
	_ "embed"
	"encoding/json"
	"fmt"
	regionDTO "service/region/dto"
	regionRepository "service/region/repository"
	"sort"
	"strings"
)

//go:embed data/level.json
var regionSeedData []byte

type seedRegion struct {
	Code     string       `json:"code"`
	Name     string       `json:"name"`
	Children []seedRegion `json:"children"`
}

type RegionService struct {
	repository *regionRepository.RegionRepository
}

func NewRegionService() *RegionService {
	return &RegionService{
		repository: db.GetRepository[regionRepository.RegionRepository](),
	}
}

func (s *RegionService) EnsureTable() error {
	if err := s.repository.EnsureTable(); err != nil {
		return err
	}
	return s.seedRegions()
}

func (s *RegionService) ListTree() ([]*regionDTO.RegionTreeDTO, error) {
	entities, err := s.repository.ListActive()
	if err != nil {
		return nil, err
	}

	nodes := make(map[string]*regionDTO.RegionTreeDTO, len(entities))
	sortOrders := make(map[string]int, len(entities))
	for _, entity := range entities {
		nodes[entity.Code] = &regionDTO.RegionTreeDTO{
			Code:  entity.Code,
			Name:  entity.Name,
			Level: entity.Level,
		}
		sortOrders[entity.Code] = entity.SortOrder
	}

	var roots []*regionDTO.RegionTreeDTO
	for _, entity := range entities {
		node := nodes[entity.Code]
		if entity.ParentCode == "" {
			roots = append(roots, node)
			continue
		}
		if parent, ok := nodes[entity.ParentCode]; ok {
			parent.Children = append(parent.Children, node)
		}
	}

	var sortTree func(nodes []*regionDTO.RegionTreeDTO)
	sortTree = func(items []*regionDTO.RegionTreeDTO) {
		sort.SliceStable(items, func(i, j int) bool {
			return sortOrders[items[i].Code] < sortOrders[items[j].Code]
		})
		for _, item := range items {
			if len(item.Children) > 0 {
				sortTree(item.Children)
			}
		}
	}
	sortTree(roots)
	return roots, nil
}

func (s *RegionService) seedRegions() error {
	total, err := s.repository.CountActive()
	if err != nil {
		return err
	}
	if total > 0 {
		return nil
	}

	var seeds []seedRegion
	if err := json.Unmarshal(regionSeedData, &seeds); err != nil {
		return fmt.Errorf("parse region seed data: %w", err)
	}

	var entities []*regionRepository.Region
	for provinceIndex, province := range seeds {
		provinceCode := strings.TrimSpace(province.Code)
		entities = append(entities, newRegion(provinceCode, strings.TrimSpace(province.Name), "", 1, provinceIndex))
		virtualCityAdded := false
		for cityIndex, city := range province.Children {
			if len(city.Children) == 0 {
				virtualCityCode := provinceCode + "-city"
				if !virtualCityAdded {
					entities = append(entities, newRegion(virtualCityCode, strings.TrimSpace(province.Name), provinceCode, 2, 0))
					virtualCityAdded = true
				}
				entities = append(entities, newRegion(strings.TrimSpace(city.Code), strings.TrimSpace(city.Name), virtualCityCode, 3, cityIndex))
				continue
			}
			cityCode := strings.TrimSpace(city.Code)
			entities = append(entities, newRegion(cityCode, strings.TrimSpace(city.Name), provinceCode, 2, cityIndex))
			for areaIndex, area := range city.Children {
				entities = append(entities, newRegion(strings.TrimSpace(area.Code), strings.TrimSpace(area.Name), cityCode, 3, areaIndex))
			}
		}
	}

	return s.repository.BatchCreate(entities)
}

func newRegion(code string, name string, parentCode string, level int, sortOrder int) *regionRepository.Region {
	return &regionRepository.Region{
		Code:       code,
		Name:       name,
		ParentCode: parentCode,
		Level:      level,
		SortOrder:  sortOrder,
		Status:     "active",
	}
}
