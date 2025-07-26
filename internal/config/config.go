package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// NodeConfig represents configuration for a single blockchain node
type NodeConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Node    struct {
		HomeDir     string `mapstructure:"home_dir"`
		DataDir     string `mapstructure:"data_dir"`
		ChainID     string `mapstructure:"chain_id"`
		BinaryPath  string `mapstructure:"binary_path"`
		RPCEndpoint string `mapstructure:"rpc_endpoint"`
	} `mapstructure:"node"`
	Snapshot struct {
		Enabled     bool          `mapstructure:"enabled"`
		Interval    time.Duration `mapstructure:"interval"`
		Retention   int           `mapstructure:"retention"`
		Compression bool          `mapstructure:"compression"`
		TempDir     string        `mapstructure:"temp_dir"`
	} `mapstructure:"snapshot"`
	S3 struct {
		Bucket     string `mapstructure:"bucket"`
		Region     string `mapstructure:"region"`
		AccessKey  string `mapstructure:"access_key"`
		SecretKey  string `mapstructure:"secret_key"`
		Endpoint   string `mapstructure:"endpoint"`
		PathPrefix string `mapstructure:"path_prefix"`
		UseSSL     bool   `mapstructure:"use_ssl"`
	} `mapstructure:"s3"`
}

// GlobalS3Config represents global S3 settings
type GlobalS3Config struct {
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Endpoint  string `mapstructure:"endpoint"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Config represents the application configuration
type Config struct {
	Nodes        map[string]NodeConfig `mapstructure:"nodes"`
	GlobalS3     GlobalS3Config        `mapstructure:"global_s3"`
	Logging      LoggingConfig         `mapstructure:"logging"`
	SelectedNode string                // Currently selected node
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("nodes")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/snapshot-cosmos")

	// Environment variables
	viper.SetEnvPrefix("SNAPSHOT_COSMOS")
	viper.AutomaticEnv()

	// Default values
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Global S3 defaults
	viper.SetDefault("global_s3.region", "us-east-1")
	viper.SetDefault("global_s3.use_ssl", true)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	// Node defaults
	viper.SetDefault("nodes.cosmoshub.enabled", true)
	viper.SetDefault("nodes.cosmoshub.node.home_dir", os.Getenv("HOME")+"/.cosmos")
	viper.SetDefault("nodes.cosmoshub.node.data_dir", "data")
	viper.SetDefault("nodes.cosmoshub.node.chain_id", "cosmoshub-4")
	viper.SetDefault("nodes.cosmoshub.node.binary_path", "gaiad")
	viper.SetDefault("nodes.cosmoshub.node.rpc_endpoint", "http://localhost:26657")
	viper.SetDefault("nodes.cosmoshub.snapshot.enabled", true)
	viper.SetDefault("nodes.cosmoshub.snapshot.interval", "24h")
	viper.SetDefault("nodes.cosmoshub.snapshot.retention", 7)
	viper.SetDefault("nodes.cosmoshub.snapshot.compression", true)
	viper.SetDefault("nodes.cosmoshub.snapshot.temp_dir", "/tmp/snapshot-cosmos/cosmoshub")
	viper.SetDefault("nodes.cosmoshub.s3.bucket", "q163i-snapshots")
	viper.SetDefault("nodes.cosmoshub.s3.region", "us-east-1")
	viper.SetDefault("nodes.cosmoshub.s3.path_prefix", "snapshots/cosmoshub")
	viper.SetDefault("nodes.cosmoshub.s3.use_ssl", true)
}

// validateConfig validates the configuration
func validateConfig(cfg *Config) error {
	// Check if at least one node is enabled
	enabledNodes := 0
	for name, nodeCfg := range cfg.Nodes {
		if nodeCfg.Enabled {
			enabledNodes++
			if err := validateNodeConfig(name, &nodeCfg); err != nil {
				return err
			}
		}
	}

	if enabledNodes == 0 {
		return fmt.Errorf("no enabled nodes found in configuration")
	}

	return nil
}

// validateNodeConfig validates a single node configuration
func validateNodeConfig(name string, nodeCfg *NodeConfig) error {
	// Validate node configuration
	if nodeCfg.Node.HomeDir == "" {
		return fmt.Errorf("node %s: home_dir is required", name)
	}

	if nodeCfg.Node.ChainID == "" {
		return fmt.Errorf("node %s: chain_id is required", name)
	}

	// Validate S3 configuration
	if nodeCfg.S3.Bucket == "" {
		return fmt.Errorf("node %s: s3.bucket is required", name)
	}

	if nodeCfg.S3.Region == "" {
		return fmt.Errorf("node %s: s3.region is required", name)
	}

	// Validate snapshot configuration
	if nodeCfg.Snapshot.Interval <= 0 {
		return fmt.Errorf("node %s: snapshot.interval must be positive", name)
	}

	if nodeCfg.Snapshot.Retention < 0 {
		return fmt.Errorf("node %s: snapshot.retention cannot be negative", name)
	}

	return nil
}

// GetEnabledNodes returns a list of enabled nodes
func (c *Config) GetEnabledNodes() []string {
	var enabled []string
	for name, nodeCfg := range c.Nodes {
		if nodeCfg.Enabled {
			enabled = append(enabled, name)
		}
	}
	return enabled
}

// GetNodeConfig returns configuration for a specific node
func (c *Config) GetNodeConfig(nodeName string) (*NodeConfig, error) {
	nodeCfg, exists := c.Nodes[nodeName]
	if !exists {
		return nil, fmt.Errorf("node %s not found in configuration", nodeName)
	}

	if !nodeCfg.Enabled {
		return nil, fmt.Errorf("node %s is not enabled", nodeName)
	}

	// Merge with global S3 settings if not set
	if nodeCfg.S3.AccessKey == "" {
		nodeCfg.S3.AccessKey = c.GlobalS3.AccessKey
	}
	if nodeCfg.S3.SecretKey == "" {
		nodeCfg.S3.SecretKey = c.GlobalS3.SecretKey
	}
	if nodeCfg.S3.Endpoint == "" {
		nodeCfg.S3.Endpoint = c.GlobalS3.Endpoint
	}

	return &nodeCfg, nil
}

// GetNodeDataPath returns the full path to the node data directory
func (nc *NodeConfig) GetNodeDataPath() string {
	return filepath.Join(nc.Node.HomeDir, nc.Node.DataDir)
}

// GetSnapshotPath returns the path where snapshots should be stored
func (nc *NodeConfig) GetSnapshotPath() string {
	return filepath.Join(nc.Snapshot.TempDir, nc.Node.ChainID)
}
