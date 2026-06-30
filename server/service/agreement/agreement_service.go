package agreement

import (
	"common/middleware/db"
	"embed"
	"strings"
	"time"

	agreementRepository "service/agreement/repository"
)

// Version 当前协议版本。协议正文有实质变更时递增，用户需重新同意。
const Version = "1.0.0"

//go:embed content/user_agreement.md content/privacy_policy.md content/privacy_guide.md
var contentFS embed.FS

// Document 一份协议文档。
type Document struct {
	Key     string `json:"key"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var documentMeta = []struct {
	Key   string
	Title string
	File  string
}{
	{Key: "user_agreement", Title: "用户协议", File: "content/user_agreement.md"},
	{Key: "privacy_policy", Title: "隐私政策", File: "content/privacy_policy.md"},
	{Key: "privacy_guide", Title: "小程序隐私保护指引", File: "content/privacy_guide.md"},
}

type AgreementService struct {
	consentRepository *agreementRepository.AgreementConsentRepository
}

func NewAgreementService() *AgreementService {
	return &AgreementService{
		consentRepository: db.GetRepository[agreementRepository.AgreementConsentRepository](),
	}
}

func (s *AgreementService) EnsureTable() error {
	return s.consentRepository.EnsureTable()
}

// Documents 返回内嵌的协议正文列表。
func (s *AgreementService) Documents() ([]Document, error) {
	docs := make([]Document, 0, len(documentMeta))
	for _, meta := range documentMeta {
		raw, err := contentFS.ReadFile(meta.File)
		if err != nil {
			return nil, err
		}
		docs = append(docs, Document{
			Key:     meta.Key,
			Title:   meta.Title,
			Content: string(raw),
		})
	}
	return docs, nil
}

// IsAgreed 判断该 openid 是否已同意当前版本协议。
func (s *AgreementService) IsAgreed(openID string) (bool, error) {
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return false, nil
	}
	record, err := s.consentRepository.FindByOpenID(openID)
	if err != nil {
		return false, err
	}
	if record == nil {
		return false, nil
	}
	return strings.TrimSpace(record.Version) == Version, nil
}

// RecordConsent 记录（或更新为最新版本）某个 openid 的同意。幂等：同一 openid 只保留一条记录。
func (s *AgreementService) RecordConsent(openID, unionID, ip string) error {
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return nil
	}
	record, err := s.consentRepository.FindByOpenID(openID)
	if err != nil {
		return err
	}
	now := time.Now()
	if record == nil {
		_, err = s.consentRepository.Create(&agreementRepository.AgreementConsent{
			OpenID:   openID,
			UnionID:  strings.TrimSpace(unionID),
			Version:  Version,
			AgreedAt: now,
			IP:       strings.TrimSpace(ip),
		})
		return err
	}
	record.UnionID = strings.TrimSpace(unionID)
	record.Version = Version
	record.AgreedAt = now
	record.IP = strings.TrimSpace(ip)
	_, err = s.consentRepository.SaveOrUpdate(record)
	return err
}
