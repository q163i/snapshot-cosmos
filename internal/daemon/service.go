package daemon

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/q163i/snapshot-cosmos/internal/config"
	"github.com/q163i/snapshot-cosmos/internal/s3"
	"github.com/q163i/snapshot-cosmos/internal/snapshot"
	"go.uber.org/zap"
)

// Service handles the daemon functionality
type Service struct {
	cfg         *config.NodeConfig
	logger      *zap.Logger
	snapshotSvc *snapshot.Service
	s3Svc       *s3.Service
}

// NewService creates a new daemon service
func NewService(cfg *config.NodeConfig, logger *zap.Logger) *Service {
	return &Service{
		cfg:         cfg,
		logger:      logger,
		snapshotSvc: snapshot.NewService(cfg, logger),
		s3Svc:       s3.NewService(cfg, logger),
	}
}

// Run runs the daemon service
func (s *Service) Run(ctx context.Context) error {
	s.logger.Info("Starting snapshot daemon",
		zap.String("chain_id", s.cfg.Node.ChainID),
		zap.Duration("interval", s.cfg.Snapshot.Interval))

	// Create ticker for periodic snapshots
	ticker := time.NewTicker(s.cfg.Snapshot.Interval)
	defer ticker.Stop()

	// Run initial snapshot
	if err := s.runSnapshot(); err != nil {
		s.logger.Error("Initial snapshot failed", zap.Error(err))
	}

	// Main loop
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Daemon stopped by context cancellation")
			return nil
		case <-ticker.C:
			if err := s.runSnapshot(); err != nil {
				s.logger.Error("Periodic snapshot failed", zap.Error(err))
			}
		}
	}
}

// runSnapshot creates a snapshot and uploads it to S3
func (s *Service) runSnapshot() error {
	s.logger.Info("Starting periodic snapshot",
		zap.String("chain_id", s.cfg.Node.ChainID))

	// Create snapshot
	snapshotPath, err := s.snapshotSvc.Create()
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Upload to S3
	fileName := filepath.Base(snapshotPath)
	s3Key := fmt.Sprintf("%s/%s", s.cfg.S3.PathPrefix, fileName)

	if err := s.s3Svc.Upload(snapshotPath, s3Key); err != nil {
		return fmt.Errorf("failed to upload snapshot: %w", err)
	}

	// Cleanup old snapshots
	if err := s.snapshotSvc.Cleanup(); err != nil {
		s.logger.Warn("Failed to cleanup old snapshots", zap.Error(err))
	}

	// Cleanup old S3 snapshots
	if err := s.cleanupOldS3Snapshots(); err != nil {
		s.logger.Warn("Failed to cleanup old S3 snapshots", zap.Error(err))
	}

	s.logger.Info("Periodic snapshot completed successfully",
		zap.String("snapshot_path", snapshotPath),
		zap.String("s3_key", s3Key))

	return nil
}

// cleanupOldS3Snapshots removes old snapshots from S3 based on retention policy
func (s *Service) cleanupOldS3Snapshots() error {
	prefix := fmt.Sprintf("%s/", s.cfg.S3.PathPrefix)

	// List snapshots in S3
	keys, err := s.s3Svc.List(prefix)
	if err != nil {
		return fmt.Errorf("failed to list S3 snapshots: %w", err)
	}

	// If we have more snapshots than retention limit, remove oldest ones
	if len(keys) > s.cfg.Snapshot.Retention {
		keysToRemove := len(keys) - s.cfg.Snapshot.Retention
		for i := 0; i < keysToRemove; i++ {
			if err := s.s3Svc.Delete(keys[i]); err != nil {
				s.logger.Error("Failed to delete old S3 snapshot",
					zap.String("key", keys[i]),
					zap.Error(err))
			} else {
				s.logger.Info("Removed old S3 snapshot", zap.String("key", keys[i]))
			}
		}
	}

	return nil
}
