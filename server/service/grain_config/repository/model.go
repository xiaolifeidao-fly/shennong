package repository

import "common/middleware/db"

type GrainStation struct {
	db.BaseEntity
	StationName  string `gorm:"column:station_name;type:varchar(100);index:idx_station_name" description:"粮站名称"`
	StationCode  string `gorm:"column:station_code;type:varchar(50);uniqueIndex:uk_station_code" description:"粮站编码"`
	TenantID     uint64 `gorm:"column:tenant_id;type:bigint unsigned;index:idx_tenant_id" description:"租户ID"`
	ContactName  string `gorm:"column:contact_name;type:varchar(50)" description:"联系人"`
	ContactPhone string `gorm:"column:contact_phone;type:varchar(32)" description:"联系电话"`
	Province     string `gorm:"column:province;type:varchar(50)" description:"省"`
	City         string `gorm:"column:city;type:varchar(50)" description:"市"`
	District     string `gorm:"column:district;type:varchar(50)" description:"区县"`
	Address      string `gorm:"column:address;type:varchar(255)" description:"详细地址"`
	Longitude    string `gorm:"column:longitude;type:varchar(50)" description:"经度"`
	Latitude     string `gorm:"column:latitude;type:varchar(50)" description:"纬度"`
	Status       string `gorm:"column:status;type:varchar(50);index:idx_status" description:"状态"`
	Remark       string `gorm:"column:remark;type:varchar(255)" description:"备注"`
}

func (g *GrainStation) TableName() string {
	return "grain_station"
}

type GrainStationUser struct {
	db.BaseEntity
	StationID uint64 `gorm:"column:station_id;type:bigint unsigned;index:idx_station_id" description:"粮站ID"`
	AppUserID uint64 `gorm:"column:app_user_id;type:bigint unsigned;index:idx_app_user_id" description:"小程序用户ID"`
	RoleType  string `gorm:"column:role_type;type:varchar(50)" description:"粮站角色"`
	Status    string `gorm:"column:status;type:varchar(50);index:idx_status" description:"状态"`
}

func (g *GrainStationUser) TableName() string {
	return "grain_station_user"
}

type GrainPurchaseType struct {
	db.BaseEntity
	StationID uint64 `gorm:"column:station_id;type:bigint unsigned;index:idx_station_id" description:"粮站ID"`
	TypeName  string `gorm:"column:type_name;type:varchar(100);index:idx_type_name" description:"收购类型名称"`
	Unit      string `gorm:"column:unit;type:varchar(20)" description:"默认单位"`
	SortOrder int    `gorm:"column:sort_order;type:int;default:0" description:"排序"`
	Status    string `gorm:"column:status;type:varchar(50);index:idx_status" description:"状态"`
}

func (g *GrainPurchaseType) TableName() string {
	return "grain_purchase_type"
}

type GrainPaymentMethod struct {
	db.BaseEntity
	MethodCode string `gorm:"column:method_code;type:varchar(50);uniqueIndex:uk_method_code" description:"付款方式编码"`
	MethodName string `gorm:"column:method_name;type:varchar(100);index:idx_method_name" description:"付款方式名称"`
	SortOrder  int    `gorm:"column:sort_order;type:int;default:0" description:"排序"`
	Status     string `gorm:"column:status;type:varchar(50);index:idx_status" description:"状态"`
}

func (g *GrainPaymentMethod) TableName() string {
	return "grain_payment_method"
}

type GrainPurchasePlace struct {
	db.BaseEntity
	AppUserID uint64 `gorm:"column:app_user_id;type:bigint unsigned;index:idx_app_user_id" description:"业务员ID"`
	PlaceName string `gorm:"column:place_name;type:varchar(100);index:idx_place_name" description:"收购地点名称"`
	PlaceType string `gorm:"column:place_type;type:varchar(50)" description:"地点类型"`
	Province  string `gorm:"column:province;type:varchar(50)" description:"省"`
	City      string `gorm:"column:city;type:varchar(50)" description:"市"`
	District  string `gorm:"column:district;type:varchar(50)" description:"区县"`
	Address   string `gorm:"column:address;type:varchar(255)" description:"详细地址"`
	Longitude string `gorm:"column:longitude;type:varchar(50)" description:"经度"`
	Latitude  string `gorm:"column:latitude;type:varchar(50)" description:"纬度"`
	SortOrder int    `gorm:"column:sort_order;type:int;default:0" description:"排序"`
	Status    string `gorm:"column:status;type:varchar(50);index:idx_status" description:"状态"`
}

func (g *GrainPurchasePlace) TableName() string {
	return "grain_purchase_place"
}
