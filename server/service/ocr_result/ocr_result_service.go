package ocr_result

import (
	"common/middleware/db"
	ocrResultRepository "service/ocr_result/repository"
)

type OcrResultService struct {
	repo *ocrResultRepository.OcrResultRepository
}

func NewOcrResultService() *OcrResultService {
	return &OcrResultService{
		repo: db.GetRepository[ocrResultRepository.OcrResultRepository](),
	}
}

func (s *OcrResultService) EnsureTable() error {
	return s.repo.EnsureTable()
}

// GetSuccessResult 若 businessID 已有成功结果，返回JSON字符串；否则返回空串
func (s *OcrResultService) GetSuccessResult(businessID string) string {
	entity, err := s.repo.FindByBusinessID(businessID)
	if err != nil {
		return ""
	}
	if entity.Status == ocrResultRepository.OcrStatusSuccess {
		return entity.ResultJSON
	}
	return ""
}

// SaveResult 保存或更新OCR识别结果
func (s *OcrResultService) SaveResult(businessID, ossURL, resultJSON, status string) error {
	return s.repo.Upsert(&ocrResultRepository.OcrResult{
		BusinessID: businessID,
		OssURL:     ossURL,
		ResultJSON: resultJSON,
		Status:     status,
	})
}
