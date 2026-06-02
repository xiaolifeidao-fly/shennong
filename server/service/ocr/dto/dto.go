package dto

type RecognizeCardRequest struct {
	CardType   string
	FileName   string
	FileSize   int64
	MimeType   string
	ImageURL   string
	ImageBytes []byte
}

type RecognizeCardResult struct {
	CardType     string                 `json:"cardType"`
	Mock         bool                   `json:"mock"`
	OssBucket    string                 `json:"ossBucket,omitempty"`
	OssObjectKey string                 `json:"ossObjectKey,omitempty"`
	OssURL       string                 `json:"ossUrl,omitempty"`
	FileName     string                 `json:"fileName"`
	FileSize     int64                  `json:"fileSize"`
	MimeType     string                 `json:"mimeType"`
	RequestID    string                 `json:"requestId,omitempty"`
	Name         string                 `json:"name,omitempty"`
	IDNumber     string                 `json:"idNumber,omitempty"`
	Address      string                 `json:"address,omitempty"`
	Sex          string                 `json:"sex,omitempty"`
	Ethnicity    string                 `json:"ethnicity,omitempty"`
	BirthDate    string                 `json:"birthDate,omitempty"`
	BankNumber   string                 `json:"bankNumber,omitempty"`
	BankName     string                 `json:"bankName,omitempty"`
	BankCardType string                 `json:"bankCardType,omitempty"`
	ValidToDate  string                 `json:"validToDate,omitempty"`
	RawData      map[string]interface{} `json:"rawData,omitempty"`
}
