package repository

import "common/middleware/db"

type GrainFarmerListRow struct {
	db.BaseEntity
	StationID           uint64 `gorm:"column:station_id"`
	StationName         string `gorm:"column:station_name"`
	AppUserID           uint64 `gorm:"column:app_user_id"`
	Name                string `gorm:"column:name"`
	NameDigest          string `gorm:"column:name_digest"`
	IDNumber            string `gorm:"column:id_number"`
	IDNumberDigest      string `gorm:"column:id_number_digest"`
	IDNumberLast4Digest string `gorm:"column:id_number_last4_digest"`
	Phone               string `gorm:"column:phone"`
	Address             string `gorm:"column:address"`
	BankNumber          string `gorm:"column:bank_number"`
	BankName            string `gorm:"column:bank_name"`
	Status              string `gorm:"column:status"`
	StatusText          string `gorm:"column:status_text"`
	Remark              string `gorm:"column:remark"`
}

func (r *GrainFarmerListRow) TableName() string {
	return "grain_farmer"
}

type GrainFarmer struct {
	db.BaseEntity
	StationID           uint64 `gorm:"column:station_id;type:bigint unsigned;index:idx_station_id" description:"粮站ID"`
	AppUserID           uint64 `gorm:"column:app_user_id;type:bigint unsigned;index:idx_app_user_id" description:"建档业务员ID"`
	Name                string `gorm:"column:name;type:varchar(255)" description:"农户姓名"`
	NameDigest          string `gorm:"column:name_digest;type:char(64);index:idx_name_digest" description:"姓名精确检索摘要"`
	NameSearch          string `gorm:"column:name_search;type:varchar(255);index:idx_name_search" description:"姓名前缀检索编码"`
	IDNumber            string `gorm:"column:id_number;type:varchar(128)" description:"身份证号"`
	IDNumberDigest      string `gorm:"column:id_number_digest;type:char(64);index:idx_id_number_digest" description:"身份证号检索摘要"`
	IDNumberLast4Digest string `gorm:"column:id_number_last4_digest;type:char(64);index:idx_id_number_last4_digest" description:"身份证号后4位检索摘要"`
	Phone               string `gorm:"column:phone;type:varchar(32);index:idx_phone" description:"电话"`
	Address             string `gorm:"column:address;type:varchar(255)" description:"地址"`
	BankNumber          string `gorm:"column:bank_number;type:varchar(160)" description:"银行卡号"`
	BankName            string `gorm:"column:bank_name;type:varchar(255)" description:"开户行"`
	Status              string `gorm:"column:status;type:varchar(50);index:idx_status" description:"状态"`
	StatusText          string `gorm:"column:status_text;type:varchar(100)" description:"状态文本"`
	Remark              string `gorm:"column:remark;type:varchar(255)" description:"备注"`
}

func (g *GrainFarmer) TableName() string {
	return "grain_farmer"
}
