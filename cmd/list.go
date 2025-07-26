package cmd

import (
	"fmt"

	"github.com/q163i/snapshot-cosmos/internal/config"
	"go.uber.org/zap"
)

// listNodes displays all configured nodes and their status
func listNodes(cfg *config.Config, logger *zap.Logger) {
	fmt.Println("Configured blockchain nodes:")
	fmt.Println("=============================")

	enabledNodes := cfg.GetEnabledNodes()
	if len(enabledNodes) == 0 {
		fmt.Println("No enabled nodes found.")
		return
	}

	for _, nodeName := range enabledNodes {
		nodeCfg, err := cfg.GetNodeConfig(nodeName)
		if err != nil {
			logger.Error("Failed to get node config",
				zap.String("node", nodeName),
				zap.Error(err))
			continue
		}

		fmt.Printf("\nNode: %s\n", nodeName)
		fmt.Printf("  Chain ID: %s\n", nodeCfg.Node.ChainID)
		fmt.Printf("  Binary: %s\n", nodeCfg.Node.BinaryPath)
		fmt.Printf("  Data Path: %s\n", nodeCfg.GetNodeDataPath())
		fmt.Printf("  RPC Endpoint: %s\n", nodeCfg.Node.RPCEndpoint)
		fmt.Printf("  Snapshot Interval: %s\n", nodeCfg.Snapshot.Interval)
		fmt.Printf("  Retention: %d snapshots\n", nodeCfg.Snapshot.Retention)
		fmt.Printf("  S3 Bucket: %s\n", nodeCfg.S3.Bucket)
		fmt.Printf("  S3 Path: %s\n", nodeCfg.S3.PathPrefix)
		fmt.Printf("  Enabled: %t\n", nodeCfg.Enabled)
	}

	fmt.Printf("\nTotal enabled nodes: %d\n", len(enabledNodes))
}
