package repository

import "common/middleware/db"

// FarmerIDCardImage 农户身份证图片（正面/背面）
type FarmerIDCardImage struct {
	db.BaseEntity
	FarmerID   uint64 `gorm:"column:farmer_id;type:bigint unsigned;not null;index:idx_farmer_id" description:"农户ID"`
	AppUserID  uint64 `gorm:"column:app_user_id;type:bigint unsigned;not null" description:"操作业务员ID"`
	ImageSide  string `gorm:"column:image_side;type:varchar(10);not null" description:"证件面，front正面/back背面"`
	ImageName  string `gorm:"column:image_name;type:varchar(255);not null" description:"原始文件名"`
	ImageHash  string `gorm:"column:image_hash;type:char(64);not null;uniqueIndex:uniq_farmer_hash_user" description:"文件名SHA-256"`
	OssURL     string `gorm:"column:oss_url;type:text" description:"OSS地址"`
}

func (e *FarmerIDCardImage) TableName() string {
	return "farmer_idcard_image"
}

// FarmerBankCardImage 农户银行卡图片
type FarmerBankCardImage struct {
	db.BaseEntity
	FarmerID  uint64 `gorm:"column:farmer_id;type:bigint unsigned;not null;index:idx_farmer_id;uniqueIndex:uniq_farmer_hash_user" description:"农户ID"`
	AppUserID uint64 `gorm:"column:app_user_id;type:bigint unsigned;not null;uniqueIndex:uniq_farmer_hash_user" description:"操作业务员ID"`
	ImageName string `gorm:"column:image_name;type:varchar(255);not null" description:"原始文件名"`
	ImageHash string `gorm:"column:image_hash;type:char(64);not null;uniqueIndex:uniq_farmer_hash_user" description:"文件名SHA-256"`
	OssURL    string `gorm:"column:oss_url;type:text" description:"OSS地址"`
}

func (e *FarmerBankCardImage) TableName() string {
	return "farmer_bankcard_image"
}
