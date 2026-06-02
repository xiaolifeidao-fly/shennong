package repository

import "common/middleware/db"

type Tenant struct {
	db.BaseEntity
	TenantName   string `gorm:"column:tenant_name;type:varchar(100);index:idx_tenant_name" description:"租户名称"`
	TenantCode   string `gorm:"column:tenant_code;type:varchar(50);uniqueIndex:uk_tenant_code" description:"租户编码"`
	ContactName  string `gorm:"column:contact_name;type:varchar(50)" description:"联系人"`
	ContactPhone string `gorm:"column:contact_phone;type:varchar(32)" description:"联系电话"`
	Status       string `gorm:"column:status;type:varchar(50);index:idx_status" description:"状态"`
	Remark       string `gorm:"column:remark;type:varchar(255)" description:"备注"`
}

func (t *Tenant) TableName() string {
	return "tenant"
}

type TenantStation struct {
	db.BaseEntity
	TenantID  uint64 `gorm:"column:tenant_id;type:bigint unsigned;index:idx_tenant_id" description:"租户ID"`
	StationID uint64 `gorm:"column:station_id;type:bigint unsigned;index:idx_station_id" description:"粮站ID"`
	Status    string `gorm:"column:status;type:varchar(50);index:idx_status" description:"状态"`
}

func (t *TenantStation) TableName() string {
	return "tenant_station"
}
