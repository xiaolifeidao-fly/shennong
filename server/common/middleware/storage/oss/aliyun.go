package oss

import (
	"bytes"
	"errors"
	"io"
	"log"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type AliyunOss struct {
	DirPrefix       string
	Endpoint        string
	BucketName      string
	AccessKeyId     string
	AccessKeySecret string

	ossClient *oss.Client
}

func NewAliyun(entity *OssEntity) (*AliyunOss, error) {
	var err error
	var ossClient *oss.Client

	aliyunOss := &AliyunOss{
		DirPrefix:       entity.DirPrefix,
		Endpoint:        entity.Endpoint,
		BucketName:      entity.BucketName,
		AccessKeyId:     entity.AccessKeyId,
		AccessKeySecret: entity.AccessKeySecret,
	}

	if ossClient, err = oss.New(entity.Endpoint, entity.AccessKeyId, entity.AccessKeySecret); err != nil {
		return nil, err
	}

	aliyunOss.ossClient = ossClient
	return aliyunOss, nil
}

func (a *AliyunOss) BuildKey(path string) string {
	if a.DirPrefix == "" {
		return path
	}

	if path[0] == '/' {
		path = path[1:]
	}
	path = a.DirPrefix + "/" + path
	return path
}

func (a *AliyunOss) Put(path string, data []byte) error {
	if len(path) == 0 || len(data) == 0 {
		return errors.New("file path or data is nil")
	}

	var err error
	var bucket *oss.Bucket
	if bucket, err = a.ossClient.Bucket(a.BucketName); err != nil {
		return err
	}

	key := a.BuildKey(path)
	log.Printf(
		"oss put start: endpoint=%s bucket=%s dirPrefix=%s key=%s bytes=%d accessKeyId=%s",
		a.Endpoint,
		a.BucketName,
		a.DirPrefix,
		key,
		len(data),
		maskSecret(a.AccessKeyId),
	)
	if err = bucket.PutObject(key, bytes.NewReader(data)); err != nil {
		log.Printf(
			"oss put failed: endpoint=%s bucket=%s key=%s bytes=%d err=%v",
			a.Endpoint,
			a.BucketName,
			key,
			len(data),
			err,
		)
		return err
	}
	log.Printf("oss put success: bucket=%s key=%s bytes=%d", a.BucketName, key, len(data))
	return nil
}

func (a *AliyunOss) Get(path string) ([]byte, error) {
	if len(path) == 0 {
		return nil, errors.New("file path is nil")
	}

	var err error
	var bucket *oss.Bucket
	if bucket, err = a.ossClient.Bucket(a.BucketName); err != nil {
		return nil, err
	}

	key := a.BuildKey(path)
	var body io.ReadCloser

	if body, err = bucket.GetObject(key); err != nil {
		return nil, err
	}
	defer body.Close()

	buf := new(bytes.Buffer)
	io.Copy(buf, body)
	return buf.Bytes(), nil
}

// GetByKey reads an OSS object using the full object key stored in the database.
func (a *AliyunOss) GetByKey(key string) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("file key is nil")
	}

	bucket, err := a.ossClient.Bucket(a.BucketName)
	if err != nil {
		return nil, err
	}

	body, err := bucket.GetObject(key)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	buf := new(bytes.Buffer)
	io.Copy(buf, body)
	return buf.Bytes(), nil
}

// 获取有效期的URL
func (a *AliyunOss) GetUrl(path string, duration *time.Duration) (string, error) {
	if len(path) == 0 {
		return "", errors.New("file path is nil")
	}

	if duration == nil {
		duration = new(time.Duration)
		*duration = time.Hour * 1
	}

	var err error
	var bucket *oss.Bucket
	if bucket, err = a.ossClient.Bucket(a.BucketName); err != nil {
		return "", err
	}

	return a.signURL(bucket, a.BuildKey(path), duration)
}

// GetUrlByKey signs a full OSS object key that already includes DirPrefix.
func (a *AliyunOss) GetUrlByKey(key string, duration *time.Duration) (string, error) {
	if len(key) == 0 {
		return "", errors.New("file key is nil")
	}
	if duration == nil {
		duration = new(time.Duration)
		*duration = time.Hour * 1
	}

	bucket, err := a.ossClient.Bucket(a.BucketName)
	if err != nil {
		return "", err
	}
	return a.signURL(bucket, key, duration)
}

func (a *AliyunOss) signURL(bucket *oss.Bucket, key string, duration *time.Duration) (string, error) {
	expiredInSec := int64((*duration).Seconds())
	return bucket.SignURL(key, oss.HTTPGet, expiredInSec)
}
