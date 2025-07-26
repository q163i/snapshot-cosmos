package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/q163i/snapshot-cosmos/internal/config"
	"github.com/q163i/snapshot-cosmos/internal/daemon"
	"go.uber.org/zap"
)

// runDaemon runs the snapshot daemon service for the specified node
func runDaemon(cfg *config.Config, logger *zap.Logger, nodeName string) error {
	// Get node configuration
	nodeCfg, err := cfg.GetNodeConfig(nodeName)
	if err != nil {
		return fmt.Errorf("failed to get node configuration: %w", err)
	}

	logger.Info("Starting snapshot daemon",
		zap.String("node", nodeName),
		zap.String("chain_id", nodeCfg.Node.ChainID),
		zap.Duration("interval", nodeCfg.Snapshot.Interval))

	// Create daemon service
	daemonSvc := daemon.NewService(nodeCfg, logger)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()

	// Run daemon
	err = daemonSvc.Run(ctx)
	if err != nil {
		logger.Error("Daemon failed", zap.Error(err))
		return fmt.Errorf("daemon failed: %w", err)
	}

	logger.Info("Daemon stopped gracefully", zap.String("node", nodeName))
	return nil
}
