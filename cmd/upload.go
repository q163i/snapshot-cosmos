package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/q163i/snapshot-cosmos/internal/config"
	"github.com/q163i/snapshot-cosmos/internal/s3"
	"go.uber.org/zap"
)

// uploadSnapshot uploads a snapshot file to S3 for the specified node
func uploadSnapshot(cfg *config.Config, logger *zap.Logger, nodeName, filePath string) error {
	// Get node configuration
	nodeCfg, err := cfg.GetNodeConfig(nodeName)
	if err != nil {
		return fmt.Errorf("failed to get node configuration: %w", err)
	}

	logger.Info("Starting snapshot upload",
		zap.String("node", nodeName),
		zap.String("file", filePath),
		zap.String("bucket", nodeCfg.S3.Bucket))

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot file does not exist: %s", filePath)
	}

	// Create S3 service
	s3Svc := s3.NewService(nodeCfg, logger)

	// Generate S3 key
	fileName := filepath.Base(filePath)
	s3Key := fmt.Sprintf("%s/%s", nodeCfg.S3.PathPrefix, fileName)

	// Upload to S3
	err = s3Svc.Upload(filePath, s3Key)
	if err != nil {
		logger.Error("Failed to upload snapshot", zap.Error(err))
		return fmt.Errorf("failed to upload snapshot: %w", err)
	}

	logger.Info("Snapshot uploaded successfully",
		zap.String("node", nodeName),
		zap.String("file", filePath),
		zap.String("s3_key", s3Key),
		zap.String("bucket", nodeCfg.S3.Bucket))

	return nil
}
