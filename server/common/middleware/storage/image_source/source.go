package image_source

import (
	"common/middleware/vipper"
	"strings"
)

const (
	OSS     = "oss"
	WXCloud = "wx_cloud"
)

func Current() string {
	source := strings.ToLower(strings.TrimSpace(vipper.GetString("image.source")))
	switch source {
	case WXCloud:
		return WXCloud
	default:
		return OSS
	}
}

func IsWXCloud() bool {
	return Current() == WXCloud
}
