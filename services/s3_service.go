package services

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct {
	Client     *s3.Client
	BucketName string
}

// NewS3Service initializes the S3Service.
func NewS3Service(bucketName string, region string) (*S3Service, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	// Initialize S3 client
	client := s3.NewFromConfig(cfg)

	return &S3Service{
		Client:     client,
		BucketName: bucketName,
	}, nil
}

// UploadPublicFile uploads a file to S3 and returns the public URL.
func (s *S3Service) UploadPublicFile(ctx context.Context, fileName string, fileData []byte, contentType string) (string, error) {
	// Upload file to S3
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.BucketName),
		Key:         aws.String(fileName),
		Body:        bytes.NewReader(fileData),
		ContentType: aws.String(contentType), // Define content-type here
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Generate public URL
	urlStr := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.BucketName, url.PathEscape(fileName))

	return urlStr, nil
}
