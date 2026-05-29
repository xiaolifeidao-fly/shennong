package repository

import "common/middleware/db"

const (
	OcrStatusPending = "pending"
	OcrStatusSuccess = "success"
	OcrStatusFailed  = "failed"
)

// OcrResult OCR识别结果缓存
type OcrResult struct {
	db.BaseEntity
	// BusinessID 业务唯一ID，如 idcard_123 / bank_456
	BusinessID string `gorm:"column:business_id;type:varchar(128);not null;uniqueIndex:uniq_business_id" description:"业务唯一ID"`
	OssURL     string `gorm:"column:oss_url;type:text" description:"识别图片OSS地址"`
	ResultJSON string `gorm:"column:result_json;type:text" description:"OCR识别结果JSON"`
	Status     string `gorm:"column:status;type:varchar(20);not null;default:'pending'" description:"识别状态: pending/success/failed"`
}

func (e *OcrResult) TableName() string {
	return "ocr_result"
}
