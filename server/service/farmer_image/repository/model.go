package repository

import "common/middleware/db"

// FarmerIDCardImage 农户身份证图片（正面/背面）
type FarmerIDCardImage struct {
	db.BaseEntity
	FarmerID     uint64 `gorm:"column:farmer_id;type:bigint unsigned;not null;index:idx_farmer_id;uniqueIndex:uniq_farmer_side_hash_user" description:"农户ID"`
	AppUserID    uint64 `gorm:"column:app_user_id;type:bigint unsigned;not null;uniqueIndex:uniq_farmer_side_hash_user" description:"操作业务员ID"`
	ImageSide    string `gorm:"column:image_side;type:varchar(10);not null;uniqueIndex:uniq_farmer_side_hash_user" description:"证件面，front正面/back背面"`
	ImageName    string `gorm:"column:image_name;type:varchar(255);not null" description:"原始文件名"`
	ImageHash    string `gorm:"column:image_hash;type:char(64);not null;uniqueIndex:uniq_farmer_side_hash_user" description:"图片SHA-256标识"`
	OssURL       string `gorm:"column:oss_url;type:text" description:"OSS地址"`
	OssObjectKey string `gorm:"column:oss_object_key;type:varchar(500)" description:"OSS对象Key，用于重新生成签名URL"`
	LastSource   string `gorm:"column:last_source;type:varchar(20);not null;default:oss" description:"图片最后来源，oss/wx_cloud"`
	WXCloudURL   string `gorm:"column:wx_cloud_url;type:varchar(1000)" description:"微信云存储地址"`
}

func (e *FarmerIDCardImage) TableName() string {
	return "farmer_idcard_image"
}

// FarmerBankCardImage 农户银行卡图片
type FarmerBankCardImage struct {
	db.BaseEntity
	FarmerID     uint64 `gorm:"column:farmer_id;type:bigint unsigned;not null;index:idx_farmer_id;uniqueIndex:uniq_farmer_hash_user" description:"农户ID"`
	AppUserID    uint64 `gorm:"column:app_user_id;type:bigint unsigned;not null;uniqueIndex:uniq_farmer_hash_user" description:"操作业务员ID"`
	ImageName    string `gorm:"column:image_name;type:varchar(255);not null" description:"原始文件名"`
	ImageHash    string `gorm:"column:image_hash;type:char(64);not null;uniqueIndex:uniq_farmer_hash_user" description:"图片SHA-256标识"`
	OssURL       string `gorm:"column:oss_url;type:text" description:"OSS地址"`
	OssObjectKey string `gorm:"column:oss_object_key;type:varchar(500)" description:"OSS对象Key，用于重新生成签名URL"`
	LastSource   string `gorm:"column:last_source;type:varchar(20);not null;default:oss" description:"图片最后来源，oss/wx_cloud"`
	WXCloudURL   string `gorm:"column:wx_cloud_url;type:varchar(1000)" description:"微信云存储地址"`
}

func (e *FarmerBankCardImage) TableName() string {
	return "farmer_bankcard_image"
}
