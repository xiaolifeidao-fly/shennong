package repository

import "common/middleware/db"

type GrainFarmer struct {
	db.BaseEntity
	StationID      uint64 `gorm:"column:station_id;type:bigint unsigned;index:idx_station_id" description:"粮站ID"`
	AppUserID      uint64 `gorm:"column:app_user_id;type:bigint unsigned;index:idx_app_user_id" description:"建档业务员ID"`
	Name           string `gorm:"column:name;type:varchar(255)" description:"农户姓名"`
	IDNumber       string `gorm:"column:id_number;type:varchar(128)" description:"身份证号"`
	IDNumberDigest string `gorm:"column:id_number_digest;type:char(64);index:idx_id_number_digest" description:"身份证号检索摘要"`
	Phone          string `gorm:"column:phone;type:varchar(32);index:idx_phone" description:"电话"`
	Address        string `gorm:"column:address;type:varchar(255)" description:"地址"`
	BankNumber     string `gorm:"column:bank_number;type:varchar(160)" description:"银行卡号"`
	BankName       string `gorm:"column:bank_name;type:varchar(255)" description:"开户行"`
	Status         string `gorm:"column:status;type:varchar(50);index:idx_status" description:"状态"`
	StatusText     string `gorm:"column:status_text;type:varchar(100)" description:"状态文本"`
	Remark         string `gorm:"column:remark;type:varchar(255)" description:"备注"`
}

func (g *GrainFarmer) TableName() string {
	return "grain_farmer"
}
