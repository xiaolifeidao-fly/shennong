package dto

import baseDTO "common/base/dto"

type GrainFarmerDTO struct {
	baseDTO.BaseDTO
	StationID  uint64 `json:"stationId"`
	AppUserID  uint64 `json:"appUserId"`
	Name       string `json:"name"`
	IDNumber   string `json:"idNumber"`
	Phone      string `json:"phone"`
	Address    string `json:"address"`
	BankNumber string `json:"bankNumber"`
	BankName   string `json:"bankName"`
	Status     string `json:"status"`
	StatusText string `json:"statusText"`
	Remark     string `json:"remark"`
}

type GrainFarmerQueryDTO struct {
	Page                 int    `form:"page"`
	PageIndex            int    `form:"pageIndex"`
	PageSize             int    `form:"pageSize"`
	StationID            uint64 `form:"stationId"`
	AppUserID            uint64 `form:"appUserId"`
	Search               string `form:"search"`
	SearchIDNumberDigest string `form:"-"`
	Status               string `form:"status"`
}
