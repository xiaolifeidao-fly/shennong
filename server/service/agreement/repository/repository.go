package repository

import (
	"common/middleware/db"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type AgreementConsentRepository struct {
	db.Repository[*AgreementConsent]
}

func (r *AgreementConsentRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&AgreementConsent{})
}

// FindByOpenID 返回指定 openid 的同意记录；不存在时返回 (nil, nil)。
func (r *AgreementConsentRepository) FindByOpenID(openID string) (*AgreementConsent, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return nil, nil
	}
	var entity AgreementConsent
	if err := r.Db.Where("open_id = ? AND active = ?", openID, 1).First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}
