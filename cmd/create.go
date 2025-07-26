package cmd

import (
	"fmt"
	"os"

	"github.com/q163i/snapshot-cosmos/internal/config"
	"github.com/q163i/snapshot-cosmos/internal/snapshot"
	"go.uber.org/zap"
)

// createSnapshot creates a new snapshot of the specified blockchain node
func createSnapshot(cfg *config.Config, logger *zap.Logger, nodeName string) error {
	// Get node configuration
	nodeCfg, err := cfg.GetNodeConfig(nodeName)
	if err != nil {
		return fmt.Errorf("failed to get node configuration: %w", err)
	}

	logger.Info("Starting snapshot creation",
		zap.String("node", nodeName),
		zap.String("chain_id", nodeCfg.Node.ChainID),
		zap.String("data_path", nodeCfg.GetNodeDataPath()))

	// Check if node data directory exists
	if _, err := os.Stat(nodeCfg.GetNodeDataPath()); os.IsNotExist(err) {
		return fmt.Errorf("node data directory does not exist: %s", nodeCfg.GetNodeDataPath())
	}

	// Create snapshot service
	snapshotSvc := snapshot.NewService(nodeCfg, logger)

	// Create snapshot
	snapshotPath, err := snapshotSvc.Create()
	if err != nil {
		logger.Error("Failed to create snapshot", zap.Error(err))
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	logger.Info("Snapshot created successfully",
		zap.String("node", nodeName),
		zap.String("path", snapshotPath),
		zap.String("chain_id", nodeCfg.Node.ChainID))

	return nil
}
