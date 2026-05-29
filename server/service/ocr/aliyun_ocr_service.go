package ocr

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"common/middleware/vipper"
	ocrDTO "service/ocr/dto"
)

const (
	cardTypeIDCard   = "id-card"
	cardTypeBankCard = "bank-card"
	defaultEndpoint  = "https://ocr-api.cn-hangzhou.aliyuncs.com"
	apiVersion       = "2021-07-07"
)

type AliyunOCRService struct {
	endpoint          string
	accessKeyID       string
	accessKeySecret   string
	timeout           time.Duration
	outputFigure      bool
	outputQualityInfo bool
	httpClient        *http.Client
}

type aliyunOCRResponse struct {
	RequestID string `json:"RequestId"`
	Data      string `json:"Data"`
	Code      string `json:"Code"`
	Message   string `json:"Message"`
}

func NewAliyunOCRServiceFromConfig() *AliyunOCRService {
	endpoint := strings.TrimSpace(vipper.GetString("ocr.endpoint"))
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	timeoutSeconds := vipper.GetInt("ocr.timeoutSeconds")
	if timeoutSeconds <= 0 {
		timeoutSeconds = 15
	}
	return &AliyunOCRService{
		endpoint:          strings.TrimRight(endpoint, "/"),
		accessKeyID:       strings.TrimSpace(vipper.GetString("ocr.accessKeyId")),
		accessKeySecret:   strings.TrimSpace(vipper.GetString("ocr.accessKeySecret")),
		timeout:           time.Duration(timeoutSeconds) * time.Second,
		outputFigure:      vipper.GetBool("ocr.idCard.outputFigure"),
		outputQualityInfo: vipper.GetBool("ocr.idCard.outputQualityInfo"),
		httpClient:        &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second},
	}
}

func (s *AliyunOCRService) RecognizeCard(ctx context.Context, req ocrDTO.RecognizeCardRequest) (*ocrDTO.RecognizeCardResult, error) {
	if strings.TrimSpace(req.CardType) == "" {
		return nil, errors.New("cardType不能为空")
	}
	if strings.TrimSpace(req.ImageURL) == "" {
		return nil, errors.New("识别图片URL不能为空")
	}
	if req.CardType == cardTypeBankCard {
		return s.RecognizeBankCard(ctx, req)
	}
	if req.CardType == cardTypeIDCard {
		return s.RecognizeIDCard(ctx, req)
	}
	return nil, errors.New("cardType必须是id-card或bank-card")
}

func (s *AliyunOCRService) RecognizeIDCard(ctx context.Context, req ocrDTO.RecognizeCardRequest) (*ocrDTO.RecognizeCardResult, error) {
	params := map[string]string{}
	if s.outputFigure {
		params["OutputFigure"] = "true"
	}
	if s.outputQualityInfo {
		params["OutputQualityInfo"] = "true"
	}
	params["Url"] = req.ImageURL
	resp, rawData, err := s.call(ctx, "RecognizeIdcard", params)
	if err != nil {
		return nil, err
	}
	result := baseResult(req, resp.RequestID, rawData)
	fillIDCardResult(result, rawData)
	return result, nil
}

func (s *AliyunOCRService) RecognizeBankCard(ctx context.Context, req ocrDTO.RecognizeCardRequest) (*ocrDTO.RecognizeCardResult, error) {
	resp, rawData, err := s.call(ctx, "RecognizeBankCard", map[string]string{"Url": req.ImageURL})
	if err != nil {
		return nil, err
	}
	result := baseResult(req, resp.RequestID, rawData)
	fillBankCardResult(result, rawData)
	return result, nil
}

func (s *AliyunOCRService) call(ctx context.Context, action string, params map[string]string) (*aliyunOCRResponse, map[string]interface{}, error) {
	if strings.TrimSpace(s.accessKeyID) == "" || strings.TrimSpace(s.accessKeySecret) == "" {
		return nil, nil, errors.New("阿里云OCR配置缺失：ocr.accessKeyId或ocr.accessKeySecret为空")
	}
	query := map[string]string{
		"Action":           action,
		"Version":          apiVersion,
		"Format":           "JSON",
		"AccessKeyId":      s.accessKeyID,
		"SignatureNonce":   nonce(),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
	}
	for k, v := range params {
		if strings.TrimSpace(v) != "" {
			query[k] = v
		}
	}
	signature := sign(http.MethodGet, query, s.accessKeySecret)
	requestURL := s.endpoint + "/?" + canonicalQuery(query) + "&Signature=" + percentEncode(signature)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, nil, err
	}

	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: s.timeout}
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, nil, fmt.Errorf("阿里云OCR调用失败：HTTP %d %s", response.StatusCode, string(responseBody))
	}

	var result aliyunOCRResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, nil, err
	}
	if result.Code != "" {
		if result.Message == "" {
			result.Message = result.Code
		}
		return nil, nil, fmt.Errorf("阿里云OCR调用失败：%s", result.Message)
	}
	rawData := map[string]interface{}{}
	if strings.TrimSpace(result.Data) != "" {
		if err := json.Unmarshal([]byte(result.Data), &rawData); err != nil {
			return nil, nil, fmt.Errorf("阿里云OCR返回Data解析失败：%w", err)
		}
	}
	return &result, rawData, nil
}

func baseResult(req ocrDTO.RecognizeCardRequest, requestID string, rawData map[string]interface{}) *ocrDTO.RecognizeCardResult {
	return &ocrDTO.RecognizeCardResult{
		CardType:  req.CardType,
		Mock:      false,
		FileName:  req.FileName,
		FileSize:  req.FileSize,
		MimeType:  req.MimeType,
		RequestID: requestID,
		RawData:   rawData,
	}
}

func fillIDCardResult(result *ocrDTO.RecognizeCardResult, rawData map[string]interface{}) {
	data := nestedMap(rawData, "data", "face", "data")
	if len(data) == 0 {
		data = nestedMap(rawData, "data", "back", "data")
	}
	if len(data) == 0 {
		data = nestedMap(rawData, "data")
	}
	result.Name = stringValue(data["name"])
	result.IDNumber = stringValue(data["idNumber"])
	result.Address = stringValue(data["address"])
	result.Sex = stringValue(data["sex"])
	result.Ethnicity = stringValue(data["ethnicity"])
	result.BirthDate = stringValue(data["birthDate"])
}

func fillBankCardResult(result *ocrDTO.RecognizeCardResult, rawData map[string]interface{}) {
	data := nestedMap(rawData, "data")
	result.BankName = stringValue(data["bankName"])
	result.BankNumber = stringValue(data["cardNumber"])
	result.BankCardType = stringValue(data["cardType"])
	result.ValidToDate = stringValue(data["validToDate"])
}

func nestedMap(data map[string]interface{}, keys ...string) map[string]interface{} {
	current := data
	for _, key := range keys {
		next, ok := current[key].(map[string]interface{})
		if !ok {
			return map[string]interface{}{}
		}
		current = next
	}
	return current
}

func stringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprintf("%v", value)
}

func sign(method string, params map[string]string, accessKeySecret string) string {
	canonical := canonicalQuery(params)
	stringToSign := method + "&" + percentEncode("/") + "&" + percentEncode(canonical)
	mac := hmac.New(sha1.New, []byte(accessKeySecret+"&"))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func canonicalQuery(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, percentEncode(key)+"="+percentEncode(params[key]))
	}
	return strings.Join(parts, "&")
}

func percentEncode(value string) string {
	encoded := url.QueryEscape(value)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}

func nonce() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
