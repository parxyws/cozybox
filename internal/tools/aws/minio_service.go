package aws

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

type UploadInput struct {
	Object      io.Reader
	ObjectName  string
	ObjectSize  int64
	BucketName  string
	ContentType string
}

type AwsService struct {
	s3Client *minio.Client
}

func NewAWSService(s3Client *minio.Client) *AwsService {
	return &AwsService{s3Client: s3Client}
}

func (aws *AwsService) PutObject(ctx context.Context, entity *UploadInput) (*minio.UploadInfo, error) {
	opts := minio.PutObjectOptions{
		UserMetadata: map[string]string{"x-amz-acl": "public-read"},
		ContentType:  entity.ContentType,
	}

	uploadInfo, err := aws.s3Client.PutObject(ctx, entity.BucketName, entity.ObjectName, entity.Object, entity.ObjectSize, opts)
	if err != nil {
		return nil, fmt.Errorf("AWSUserRepository.PutObject.s3Client.PutObject - %s", err)
	}
	return &uploadInfo, nil
}

func (aws *AwsService) GetObject(ctx context.Context, bucketName string, objectName string) (*minio.Object, error) {
	object, err := aws.s3Client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("AWSUserRepository.PutObject.s3Client.GetObject - %s", err)
	}
	defer object.Close()

	return object, err
}

func (aws *AwsService) RemoveObject(ctx context.Context, bucketName string, objectName string) error {
	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
	}
	err := aws.s3Client.RemoveObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return fmt.Errorf("AWSUserRepository.PutObject.s3Client.RemoveObject - %s", err)
	}

	return nil
}

func (aws *AwsService) PresignedGetObject(ctx context.Context, bucketName string, objectName string, expiry time.Duration) (*url.URL, error) {
	var reqParam = make(url.Values)

	presignedUrl, err := aws.s3Client.PresignedGetObject(ctx, bucketName, objectName, expiry, reqParam)
	if err != nil {
		return nil, fmt.Errorf("AWSUserRepository.PutObject.s3Client.PresignedGetObject - %s", err)
	}

	return presignedUrl, nil
}
