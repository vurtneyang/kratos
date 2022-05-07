package object_storage

import (
	"context"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"io/ioutil"
	"kratos/pkg/log"
	"net/http"
	"net/url"
)

type Client interface {
	GetUrlPrefix() string                         // 获取图片路径前缀
	Put(context.Context, string, io.Reader) error // 上传
	// TODO 一些加水印的图片URL等
}

// 阿里云OSS
type OssConfig struct {
	Endpoint        string
	AccessKeyId     string
	AccessKeySecret string
	Bucket          string
	UrlPrefix       string
}

type Oss struct {
	config *OssConfig
	bucket *oss.Bucket
}

func NewOssClient(config *OssConfig) (c *Oss, err error) {
	client, err := oss.New(config.Endpoint, config.AccessKeyId, config.AccessKeySecret)
	if err != nil {
		return
	}
	bucket, err := client.Bucket(config.Bucket)
	if err != nil {
		return
	}

	c = &Oss{
		config: config,
		bucket: bucket,
	}

	return
}

func (c *Oss) Put(ctx context.Context, path string, file io.Reader) (err error) {
	err = c.bucket.PutObject(path, file)
	if err != nil {
		log.Errorc(ctx, "[dao.UploadWxMaCode] err: (%v)", err)
		return
	}
	log.Infoc(ctx, "[oss.Put] upload to oss success, path: %s", path)
	return
}

func (c *Oss) GetUrlPrefix() string {
	return c.config.UrlPrefix
}

// 腾讯云COS
type CosConfig struct {
	BucketUrl string
	SecretId  string
	SecretKey string
	UrlPrefix string
}

type Cos struct {
	config *CosConfig
	client *cos.Client
}

func NewCosClient(config *CosConfig) (c *Cos, err error) {
	u, err := url.Parse(config.BucketUrl)
	if err != nil {
		return
	}
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.SecretId,
			SecretKey: config.SecretKey,
		},
	})

	c = &Cos{
		config: config,
		client: client,
	}

	return
}

func (c *Cos) Put(ctx context.Context, path string, file io.Reader) (err error) {
	resp, err := c.client.Object.Put(ctx, path, file, &cos.ObjectPutOptions{})
	if err != nil {
		log.Errorc(ctx, "[dao.UploadImage] err: (%v)", err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Infoc(ctx, "[cos.Put] upload to cos success, body: %s, path: %s", string(body), path)
	return
}

func (c *Cos) GetUrlPrefix() string {
	return c.config.UrlPrefix
}
