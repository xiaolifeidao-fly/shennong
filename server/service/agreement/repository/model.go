package repository

import (
	"common/middleware/db"
	"time"
)

// AgreementConsent 记录单个微信用户（以 openid 为唯一键）对用户协议的同意情况。
// 协议判断发生在登录之前，因此本表不与 app_user 关联，直接以 openid 标识微信用户。
type AgreementConsent struct {
	db.BaseEntity
	OpenID   string    `gorm:"column:open_id;type:varchar(100);uniqueIndex:idx_agreement_open_id" description:"微信OpenID"`
	UnionID  string    `gorm:"column:union_id;type:varchar(100);index:idx_agreement_union_id" description:"微信UnionID"`
	Version  string    `gorm:"column:version;type:varchar(50)" description:"同意的协议版本"`
	AgreedAt time.Time `gorm:"column:agreed_at;type:datetime" description:"同意时间"`
	IP       string    `gorm:"column:ip;type:varchar(50)" description:"同意时IP"`
}

func (c *AgreementConsent) TableName() string {
	return "agreement_consent"
}
