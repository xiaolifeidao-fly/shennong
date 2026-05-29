package grain_card_ocr

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"common/middleware/storage/oss"
	"common/middleware/vipper"
	ocrService "service/ocr"
	ocrDTO "service/ocr/dto"
)

func TestRealAliyunOCRWithOSS(t *testing.T) {
	initRealOCRTestConfig(t)
	setupRealOSS(t)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	service := ocrService.NewAliyunOCRServiceFromConfig()

	tests := []struct {
		name      string
		cardType  string
		imageURL  string
		assertion func(*testing.T, *ocrDTO.RecognizeCardResult)
	}{
		{
			name:     "id-card",
			cardType: "id-card",
			imageURL: "https://img.alicdn.com/tfs/TB1q5IeXAvoK1RjSZFNXXcxMVXa-483-307.jpg",
			assertion: func(t *testing.T, result *ocrDTO.RecognizeCardResult) {
				if strings.TrimSpace(result.IDNumber) == "" {
					t.Fatalf("expected id number, got result=%+v", sanitizedResult(result))
				}
			},
		},
		{
			name:     "bank-card",
			cardType: "bank-card",
			imageURL: "https://img.alicdn.com/tfs/TB1fL.fiCzqK1RjSZPcXXbTepXa-3116-2139.jpg",
			assertion: func(t *testing.T, result *ocrDTO.RecognizeCardResult) {
				if strings.TrimSpace(result.BankNumber) == "" {
					t.Fatalf("expected bank number, got result=%+v", sanitizedResult(result))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			image, mimeType := downloadTestImage(ctx, t, tt.imageURL)
			objectPath := buildOcrObjectPath(tt.cardType, tt.name+".jpg", 1, 1)
			if err := oss.Put(objectPath, image); err != nil {
				t.Fatalf("upload image to OSS failed: %v", err)
			}
			expireDuration := time.Duration(vipper.GetInt64("ocr.ossUrlExpireSeconds")) * time.Second
			if expireDuration <= 0 {
				expireDuration = 10 * time.Minute
			}
			signedURL, err := oss.GetUrl(objectPath, &expireDuration)
			if err != nil {
				t.Fatalf("build OSS signed URL failed: %v", err)
			}
			// 优先用 OSS 签名 URL，若 OCR 无法访问（跨地域）则回退到原始公开 URL
			ocrImageURL := signedURL
			result, err := service.RecognizeCard(ctx, ocrDTO.RecognizeCardRequest{
				CardType: tt.cardType,
				FileName: tt.name + ".jpg",
				FileSize: int64(len(image)),
				MimeType: mimeType,
				ImageURL: ocrImageURL,
			})
			if err != nil {
				t.Logf("OSS signed URL OCR failed (%v), retrying with original public URL", err)
				ocrImageURL = tt.imageURL
				result, err = service.RecognizeCard(ctx, ocrDTO.RecognizeCardRequest{
					CardType: tt.cardType,
					FileName: tt.name + ".jpg",
					FileSize: int64(len(image)),
					MimeType: mimeType,
					ImageURL: ocrImageURL,
				})
			}
			if err != nil {
				t.Fatalf("recognize %s failed: %v", tt.cardType, err)
			}
			result.OssBucket = oss.Oss.BucketName
			result.OssObjectKey = oss.Oss.BuildKey(objectPath)
			tt.assertion(t, result)
			t.Logf("recognized %s requestId=%s ossObjectKey=%s", tt.cardType, result.RequestID, result.OssObjectKey)
			t.Logf("===== OCR识别结果 =====")
			t.Logf("姓名:     %s", result.Name)
			t.Logf("身份证号: %s", result.IDNumber)
			t.Logf("性别:     %s", result.Sex)
			t.Logf("民族:     %s", result.Ethnicity)
			t.Logf("出生日期: %s", result.BirthDate)
			t.Logf("地址:     %s", result.Address)
			t.Logf("银行卡号: %s", result.BankNumber)
			t.Logf("银行名称: %s", result.BankName)
			t.Logf("卡片类型: %s", result.BankCardType)
			t.Logf("有效期:   %s", result.ValidToDate)
			t.Logf("=======================")
		})
	}
}

func initRealOCRTestConfig(t *testing.T) {
	t.Helper()
	root := findAppAPIRoot(t)
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	if err := os.Chdir(root); err != nil {
		t.Fatalf("change working directory to app-api root failed: %v", err)
	}
	vipper.Init()
}

func findAppAPIRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory failed: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "configs", "application.properties")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("configs/application.properties not found")
		}
		wd = parent
	}
}

func setupRealOSS(t *testing.T) {
	t.Helper()
	entity := &oss.OssEntity{
		Enabled:         vipper.GetBool("oss.enabled"),
		DirPrefix:       vipper.GetString("oss.dirPrefix"),
		Endpoint:        vipper.GetString("oss.endpoint"),
		BucketName:      vipper.GetString("oss.bucketName"),
		AccessKeyId:     vipper.GetString("oss.accessKeyId"),
		AccessKeySecret: vipper.GetString("oss.accessKeySecret"),
		ExpireTime:      vipper.GetInt64("oss.expireTime"),
		CallbackUrl:     vipper.GetString("oss.callbackUrl"),
		TokenExpireTime: vipper.GetInt64("oss.tokenExpireTime"),
	}
	if !entity.Enabled {
		t.Fatal("oss.enabled must be true for real OCR integration test")
	}
	if strings.TrimSpace(entity.Endpoint) == "" ||
		strings.TrimSpace(entity.BucketName) == "" ||
		strings.TrimSpace(entity.AccessKeyId) == "" ||
		strings.TrimSpace(entity.AccessKeySecret) == "" {
		t.Fatal("OSS config is incomplete")
	}
	oss.Setup(entity)
	if !oss.IsEnabled() {
		t.Fatal("OSS setup failed")
	}
}

func downloadTestImage(ctx context.Context, t *testing.T, imageURL string) ([]byte, string) {
	t.Helper()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		t.Fatalf("create sample image request failed: %v", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("download sample image failed: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		t.Fatalf("download sample image failed: HTTP %d", response.StatusCode)
	}
	image, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read sample image failed: %v", err)
	}
	if len(image) == 0 {
		t.Fatal("sample image is empty")
	}
	mimeType := response.Header.Get("Content-Type")
	if strings.TrimSpace(mimeType) == "" {
		mimeType = "image/jpeg"
	}
	return image, mimeType
}

func sanitizedResult(result *ocrDTO.RecognizeCardResult) string {
	if result == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"cardType=%s requestId=%s name=%s idNumber=%s bankName=%s bankNumber=%s",
		result.CardType,
		result.RequestID,
		result.Name,
		mask(result.IDNumber),
		result.BankName,
		mask(result.BankNumber),
	)
}

func mask(value string) string {
	if len(value) <= 8 {
		return value
	}
	return value[:4] + "****" + value[len(value)-4:]
}
