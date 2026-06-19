package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type Uploader struct {
	bucket string
	region string
	client *cos.Client
}

func NewUploader(bc *BucketConfig) (*Uploader, error) {
	bucketURL, err := cos.NewBucketURL(bc.Bucket, bc.Region, true)
	if err != nil {
		return nil, fmt.Errorf("invalid bucket/region: %w", err)
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  bc.SecretID,
			SecretKey: bc.SecretKey,
		},
	})
	return &Uploader{bucket: bc.Bucket, region: bc.Region, client: client}, nil
}

func (u *Uploader) Upload(ctx context.Context, key string, body io.Reader) error {
	_, err := u.client.Object.Put(ctx, key, body, nil)
	if err != nil {
		return fmt.Errorf("upload %q: %w", key, err)
	}
	return nil
}

func (u *Uploader) BucketName() string { return u.bucket }
func (u *Uploader) Region() string     { return u.region }
