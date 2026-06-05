package dto

import baseDTO "common/base/dto"

type GrainStationDTO struct {
	baseDTO.BaseDTO
	StationName  string `json:"stationName"`
	StationCode  string `json:"stationCode"`
	TenantID     uint64 `json:"tenantId"`
	ContactName  string `json:"contactName"`
	ContactPhone string `json:"contactPhone"`
	Province     string `json:"province"`
	City         string `json:"city"`
	District     string `json:"district"`
	Address      string `json:"address"`
	Longitude    string `json:"longitude"`
	Latitude     string `json:"latitude"`
	Status       string `json:"status"`
	Remark       string `json:"remark"`
}

type GrainStationUserDTO struct {
	baseDTO.BaseDTO
	StationID uint64 `json:"stationId"`
	AppUserID uint64 `json:"appUserId"`
	RoleType  string `json:"roleType"`
	Status    string `json:"status"`
}

type GrainPurchaseTypeDTO struct {
	baseDTO.BaseDTO
	StationID uint64 `json:"stationId"`
	TypeName  string `json:"typeName"`
	Unit      string `json:"unit"`
	SortOrder int    `json:"sortOrder"`
	Status    string `json:"status"`
}

type GrainPaymentMethodDTO struct {
	baseDTO.BaseDTO
	MethodCode string `json:"methodCode"`
	MethodName string `json:"methodName"`
	SortOrder  int    `json:"sortOrder"`
	Status     string `json:"status"`
}

type GrainPurchasePlaceDTO struct {
	baseDTO.BaseDTO
	AppUserID uint64 `json:"appUserId"`
	PlaceName string `json:"placeName"`
	PlaceType string `json:"placeType"`
	Province  string `json:"province"`
	City      string `json:"city"`
	District  string `json:"district"`
	Address   string `json:"address"`
	Longitude string `json:"longitude"`
	Latitude  string `json:"latitude"`
	SortOrder int    `json:"sortOrder"`
	Status    string `json:"status"`
}

type GrainStationExtraDTO struct {
	StationID                uint64 `json:"stationId"`
	AccountHolderName        string `json:"accountHolderName"`
	BankName                 string `json:"bankName"`
	BankAccountNumber        string `json:"bankAccountNumber"`
	BusinessLicenseUrl       string `json:"businessLicenseUrl"`
	BusinessLicenseKey       string `json:"businessLicenseKey"`
	BusinessLicenseUpdatedAt int64  `json:"businessLicenseUpdatedAt"`
}

type GrainStationDetailDTO struct {
	baseDTO.BaseDTO
	StationName        string `json:"stationName"`
	StationCode        string `json:"stationCode"`
	TenantID           uint64 `json:"tenantId"`
	TenantName         string `json:"tenantName"`
	ContactName        string `json:"contactName"`
	ContactPhone       string `json:"contactPhone"`
	Province           string `json:"province"`
	City               string `json:"city"`
	District           string `json:"district"`
	Address            string `json:"address"`
	Longitude          string `json:"longitude"`
	Latitude           string `json:"latitude"`
	Status             string `json:"status"`
	Remark             string `json:"remark"`
	BusinessLicenseUrl string `json:"businessLicenseUrl"`
}

type GrainStationQueryDTO struct {
	Page       int      `form:"page"`
	PageIndex  int      `form:"pageIndex"`
	PageSize   int      `form:"pageSize"`
	StationID  uint64   `form:"stationId"`
	TenantID   uint64   `form:"tenantId"`
	StationIDs []uint64 `form:"-"`
	AppUserID  uint64   `form:"appUserId"`
	Search     string   `form:"search"`
	Status     string   `form:"status"`
}

type GrainConfigItemQueryDTO struct {
	Page       int      `form:"page"`
	PageIndex  int      `form:"pageIndex"`
	PageSize   int      `form:"pageSize"`
	StationID  uint64   `form:"stationId"`
	StationIDs []uint64 `form:"-"`
	AppUserID  uint64   `form:"appUserId"`
	Search     string   `form:"search"`
	Status     string   `form:"status"`
}
