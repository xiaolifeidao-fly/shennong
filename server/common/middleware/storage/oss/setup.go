package oss

import (
	"log"
	"strings"
)

func Setup(entity *OssEntity) {
	Oss = nil
	if entity == nil {
		log.Println("oss entity is nil, skip oss setup")
		return
	}
	if !entity.Enabled {
		log.Println("oss disabled, skip oss setup")
		return
	}
	entity.Default()
	log.Printf(
		"oss setup config: enabled=%t endpoint=%s bucket=%s dirPrefix=%s accessKeyId=%s accessKeySecretSet=%t accessKeySecretLen=%d",
		entity.Enabled,
		entity.Endpoint,
		entity.BucketName,
		entity.DirPrefix,
		maskSecret(entity.AccessKeyId),
		strings.TrimSpace(entity.AccessKeySecret) != "",
		len(entity.AccessKeySecret),
	)

	var err error

	var oss *AliyunOss
	if oss, err = NewAliyun(entity); err != nil {
		log.Printf("oss setup failed, skip oss capability: %v", err)
		return
	}
	Oss = oss
	log.Printf("oss setup success: endpoint=%s bucket=%s dirPrefix=%s", oss.Endpoint, oss.BucketName, oss.DirPrefix)
}

func maskSecret(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "<empty>"
	}
	if len(value) <= 8 {
		return strings.Repeat("*", len(value))
	}
	return value[:4] + strings.Repeat("*", len(value)-8) + value[len(value)-4:]
}
