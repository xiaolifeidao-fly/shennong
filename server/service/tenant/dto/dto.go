package dto

import baseDTO "common/base/dto"

type TenantDTO struct {
	baseDTO.BaseDTO
	TenantName   string   `json:"tenantName"`
	TenantCode   string   `json:"tenantCode"`
	ContactName  string   `json:"contactName"`
	ContactPhone string   `json:"contactPhone"`
	Status       string   `json:"status"`
	Remark       string   `json:"remark"`
	StationIDs   []uint64 `json:"stationIds,omitempty"`
	StationNames []string `json:"stationNames,omitempty"`
}

type TenantQueryDTO struct {
	Page      int    `form:"page"`
	PageIndex int    `form:"pageIndex"`
	PageSize  int    `form:"pageSize"`
	TenantID  uint64 `form:"tenantId"`
	Search    string `form:"search"`
	Status    string `form:"status"`
}
