package backup

import (
	"context"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type s3Options struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
}

func uploadFile(ctx context.Context, opt s3Options, objectKey, filePath string) error {
	cli, err := newMinioClient(opt)
	if err != nil {
		return err
	}
	_, err = cli.FPutObject(ctx, opt.Bucket, objectKey, filePath, minio.PutObjectOptions{})
	return err
}

func downloadObject(ctx context.Context, opt s3Options, objectKey, destPath string) error {
	cli, err := newMinioClient(opt)
	if err != nil {
		return err
	}
	return cli.FGetObject(ctx, opt.Bucket, objectKey, destPath, minio.GetObjectOptions{})
}

func pruneObjectsOlderThan(ctx context.Context, opt s3Options, prefix string, cutoff time.Time) error {
	cli, err := newMinioClient(opt)
	if err != nil {
		return err
	}
	prefix = strings.TrimPrefix(strings.TrimSpace(prefix), "/")
	ch := cli.ListObjects(ctx, opt.Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})
	for obj := range ch {
		if obj.Err != nil {
			return obj.Err
		}
		if obj.Key == "" {
			continue
		}
		if obj.LastModified.Before(cutoff) {
			if err := cli.RemoveObject(ctx, opt.Bucket, obj.Key, minio.RemoveObjectOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}

func newMinioClient(opt s3Options) (*minio.Client, error) {
	endpoint := strings.TrimSpace(opt.Endpoint)
	secure := true
	if endpoint == "" {
		endpoint = "s3.amazonaws.com"
	} else {
		if strings.HasPrefix(endpoint, "http://") {
			secure = false
			endpoint = strings.TrimPrefix(endpoint, "http://")
		} else if strings.HasPrefix(endpoint, "https://") {
			endpoint = strings.TrimPrefix(endpoint, "https://")
		}
		endpoint = strings.TrimRight(endpoint, "/")
	}
	region := strings.TrimSpace(opt.Region)
	if region == "" {
		region = "us-east-1"
	}
	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(opt.AccessKey, opt.SecretKey, ""),
		Secure: secure,
		Region: region,
	})
}

func removeObject(ctx context.Context, opt s3Options, objectKey string) error {
	cli, err := newMinioClient(opt)
	if err != nil {
		return err
	}
	return cli.RemoveObject(ctx, opt.Bucket, objectKey, minio.RemoveObjectOptions{})
}
