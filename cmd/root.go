package cmd

import (
	"github.com/q163i/snapshot-cosmos/internal/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd *cobra.Command

// Execute executes the root command
func Execute(cfg *config.Config, logger *zap.Logger) error {
	rootCmd = &cobra.Command{
		Use:   "snapshot-cosmos",
		Short: "Multi-node Cosmos-like blockchain snapshot service",
		Long: `A service for creating and uploading snapshots of multiple Cosmos-like blockchain nodes to S3.
		
This tool helps maintain regular snapshots of multiple blockchain node data for backup and recovery purposes.`,
	}

	// Add subcommands
	rootCmd.AddCommand(newCreateCmd(cfg, logger))
	rootCmd.AddCommand(newUploadCmd(cfg, logger))
	rootCmd.AddCommand(newDaemonCmd(cfg, logger))
	rootCmd.AddCommand(newListCmd(cfg, logger))
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd.Execute()
}

// newCreateCmd creates the create snapshot command
func newCreateCmd(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [node-name]",
		Short: "Create a new snapshot",
		Long:  "Create a new snapshot of the specified blockchain node data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return createSnapshot(cfg, logger, args[0])
		},
	}

	cmd.Flags().String("output", "", "Output file path (optional)")
	cmd.Flags().Bool("compress", true, "Enable compression")
	cmd.Flags().Bool("verify", true, "Verify snapshot integrity")

	return cmd
}

// newUploadCmd creates the upload command
func newUploadCmd(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload [node-name] [file]",
		Short: "Upload snapshot to S3",
		Long:  "Upload a snapshot file to S3 storage for the specified node",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return uploadSnapshot(cfg, logger, args[0], args[1])
		},
	}

	cmd.Flags().String("key", "", "S3 object key (optional)")
	cmd.Flags().Bool("public", false, "Make object public")

	return cmd
}

// newDaemonCmd creates the daemon command
func newDaemonCmd(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon [node-name]",
		Short: "Run snapshot daemon",
		Long:  "Run the snapshot service as a daemon with periodic snapshots for the specified node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDaemon(cfg, logger, args[0])
		},
	}

	cmd.Flags().Duration("interval", 0, "Snapshot interval (overrides config)")
	cmd.Flags().Bool("upload", true, "Automatically upload snapshots")

	return cmd
}

// newListCmd creates the list command
func newListCmd(cfg *config.Config, logger *zap.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured nodes",
		Long:  "List all configured blockchain nodes and their status",
		Run: func(cmd *cobra.Command, args []string) {
			listNodes(cfg, logger)
		},
	}
}

// newVersionCmd creates the version command
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("snapshot-cosmos v1.0.0")
		},
	}
}
