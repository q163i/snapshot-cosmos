package main

import (
	"fmt"
	"os"

	"github.com/q163i/snapshot-cosmos/cmd"
	"github.com/q163i/snapshot-cosmos/internal/config"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Execute root command
	if err := cmd.Execute(cfg, logger); err != nil {
		logger.Fatal("Command execution failed", zap.Error(err))
	}
}
