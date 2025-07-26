package snapshot

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/q163i/snapshot-cosmos/internal/config"
	"go.uber.org/zap"
)

// Service handles snapshot creation
type Service struct {
	cfg    *config.NodeConfig
	logger *zap.Logger
}

// NewService creates a new snapshot service
func NewService(cfg *config.NodeConfig, logger *zap.Logger) *Service {
	return &Service{
		cfg:    cfg,
		logger: logger,
	}
}

// Create creates a new snapshot of the blockchain node data
func (s *Service) Create() (string, error) {
	s.logger.Info("Creating snapshot",
		zap.String("data_path", s.cfg.GetNodeDataPath()),
		zap.String("temp_dir", s.cfg.GetSnapshotPath()))

	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(s.cfg.GetSnapshotPath(), 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate snapshot filename
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("%s-snapshot-%s.tar.gz", s.cfg.Node.ChainID, timestamp)
	snapshotPath := filepath.Join(s.cfg.GetSnapshotPath(), filename)

	// Create snapshot file
	file, err := os.Create(snapshotPath)
	if err != nil {
		return "", fmt.Errorf("failed to create snapshot file: %w", err)
	}
	defer file.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Walk through the data directory and add files to tar
	err = filepath.Walk(s.cfg.GetNodeDataPath(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
		if path == s.cfg.GetNodeDataPath() {
			return nil
		}

		// Get relative path for tar
		relPath, err := filepath.Rel(s.cfg.GetNodeDataPath(), path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return fmt.Errorf("failed to create tar header: %w", err)
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		// If it's a regular file, copy its content
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return fmt.Errorf("failed to copy file %s: %w", path, err)
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to create tar archive: %w", err)
	}

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	s.logger.Info("Snapshot created successfully",
		zap.String("path", snapshotPath),
		zap.Int64("size_bytes", fileInfo.Size()),
		zap.String("chain_id", s.cfg.Node.ChainID))

	return snapshotPath, nil
}

// Cleanup removes old snapshots based on retention policy
func (s *Service) Cleanup() error {
	s.logger.Info("Cleaning up old snapshots",
		zap.Int("retention", s.cfg.Snapshot.Retention))

	// List all snapshot files
	files, err := os.ReadDir(s.cfg.GetSnapshotPath())
	if err != nil {
		return fmt.Errorf("failed to read snapshot directory: %w", err)
	}

	// Sort files by modification time (oldest first)
	var snapshotFiles []os.FileInfo
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".gz" {
			info, err := file.Info()
			if err != nil {
				s.logger.Warn("Failed to get file info", zap.String("file", file.Name()), zap.Error(err))
				continue
			}
			snapshotFiles = append(snapshotFiles, info)
		}
	}

	// Remove files beyond retention limit
	if len(snapshotFiles) > s.cfg.Snapshot.Retention {
		filesToRemove := len(snapshotFiles) - s.cfg.Snapshot.Retention
		for i := 0; i < filesToRemove; i++ {
			filePath := filepath.Join(s.cfg.GetSnapshotPath(), snapshotFiles[i].Name())
			if err := os.Remove(filePath); err != nil {
				s.logger.Error("Failed to remove old snapshot",
					zap.String("file", filePath),
					zap.Error(err))
			} else {
				s.logger.Info("Removed old snapshot", zap.String("file", filePath))
			}
		}
	}

	return nil
}
