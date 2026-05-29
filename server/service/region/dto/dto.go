package dto

type RegionDTO struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	ParentCode string `json:"parentCode"`
	Level      int    `json:"level"`
	SortOrder  int    `json:"sortOrder"`
	Status     string `json:"status"`
}

type RegionTreeDTO struct {
	Code     string           `json:"code"`
	Name     string           `json:"name"`
	Level    int              `json:"level"`
	Children []*RegionTreeDTO `json:"children,omitempty"`
}
