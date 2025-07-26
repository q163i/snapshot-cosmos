package s3

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/q163i/snapshot-cosmos/internal/config"
	"go.uber.org/zap"
)

// Service handles S3 operations
type Service struct {
	cfg    *config.NodeConfig
	logger *zap.Logger
	client *s3.Client
}

// NewService creates a new S3 service
func NewService(cfg *config.NodeConfig, logger *zap.Logger) *Service {
	return &Service{
		cfg:    cfg,
		logger: logger,
	}
}

// Upload uploads a file to S3
func (s *Service) Upload(filePath, s3Key string) error {
	// Load AWS configuration
	awsCfg, err := s.loadAWSConfig()
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	s.client = s3.NewFromConfig(awsCfg)

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info for content length
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	s.logger.Info("Uploading file to S3",
		zap.String("file", filePath),
		zap.String("s3_key", s3Key),
		zap.String("bucket", s.cfg.S3.Bucket),
		zap.Int64("size", fileInfo.Size()))

	// Upload to S3
	_, err = s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String(s.cfg.S3.Bucket),
		Key:           aws.String(s3Key),
		Body:          file,
		ContentLength: aws.Int64(fileInfo.Size()),
		ContentType:   aws.String("application/gzip"),
	})

	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	s.logger.Info("File uploaded successfully",
		zap.String("file", filePath),
		zap.String("s3_key", s3Key))

	return nil
}

// Download downloads a file from S3
func (s *Service) Download(s3Key, localPath string) error {
	// Load AWS configuration
	awsCfg, err := s.loadAWSConfig()
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	s.client = s3.NewFromConfig(awsCfg)

	// Create local directory if it doesn't exist
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	s.logger.Info("Downloading file from S3",
		zap.String("s3_key", s3Key),
		zap.String("local_path", localPath),
		zap.String("bucket", s.cfg.S3.Bucket))

	// Download from S3
	result, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.S3.Bucket),
		Key:    aws.String(s3Key),
	})

	if err != nil {
		return fmt.Errorf("failed to download from S3: %w", err)
	}
	defer result.Body.Close()

	// Copy content to local file
	_, err = file.ReadFrom(result.Body)
	if err != nil {
		return fmt.Errorf("failed to write to local file: %w", err)
	}

	s.logger.Info("File downloaded successfully",
		zap.String("s3_key", s3Key),
		zap.String("local_path", localPath))

	return nil
}

// List lists objects in S3 bucket with prefix
func (s *Service) List(prefix string) ([]string, error) {
	// Load AWS configuration
	awsCfg, err := s.loadAWSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	s.client = s3.NewFromConfig(awsCfg)

	var keys []string

	// List objects
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.cfg.S3.Bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to list S3 objects: %w", err)
		}

		for _, obj := range page.Contents {
			keys = append(keys, *obj.Key)
		}
	}

	return keys, nil
}

// Delete deletes an object from S3
func (s *Service) Delete(s3Key string) error {
	// Load AWS configuration
	awsCfg, err := s.loadAWSConfig()
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	s.client = s3.NewFromConfig(awsCfg)

	s.logger.Info("Deleting object from S3",
		zap.String("s3_key", s3Key),
		zap.String("bucket", s.cfg.S3.Bucket))

	// Delete object
	_, err = s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.S3.Bucket),
		Key:    aws.String(s3Key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	s.logger.Info("Object deleted successfully", zap.String("s3_key", s3Key))
	return nil
}

// loadAWSConfig loads AWS configuration
func (s *Service) loadAWSConfig() (aws.Config, error) {
	var opts []func(*awsconfig.LoadOptions) error

	// Set region
	if s.cfg.S3.Region != "" {
		opts = append(opts, awsconfig.WithRegion(s.cfg.S3.Region))
	}

	// Set custom endpoint if provided
	if s.cfg.S3.Endpoint != "" {
		opts = append(opts, awsconfig.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               s.cfg.S3.Endpoint,
					HostnameImmutable: true,
					PartitionID:       "aws",
				}, nil
			}),
		))
	}

	// Load configuration
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return cfg, nil
}
