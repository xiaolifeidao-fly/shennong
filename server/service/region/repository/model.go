package repository

import "common/middleware/db"

type Region struct {
	db.BaseEntity
	Code       string `gorm:"column:code;type:varchar(32);uniqueIndex:uk_region_code" description:"行政区划代码"`
	Name       string `gorm:"column:name;type:varchar(100);index:idx_region_name" description:"行政区划名称"`
	ParentCode string `gorm:"column:parent_code;type:varchar(32);index:idx_parent_code" description:"上级行政区划代码"`
	Level      int    `gorm:"column:level;type:int;index:idx_region_level" description:"层级：1省 2市 3区县"`
	SortOrder  int    `gorm:"column:sort_order;type:int;default:0" description:"排序"`
	Status     string `gorm:"column:status;type:varchar(50);index:idx_status" description:"状态"`
}

func (r *Region) TableName() string {
	return "region"
}
