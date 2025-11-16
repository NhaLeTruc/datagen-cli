package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	// Global settings
	Verbose bool   `mapstructure:"verbose"`
	Config  string `mapstructure:"config"`

	// Generation settings
	DefaultSeed      int64  `mapstructure:"default_seed"`
	DefaultFormat    string `mapstructure:"default_format"`
	DefaultRowCount  int    `mapstructure:"default_row_count"`
	DefaultBatchSize int    `mapstructure:"default_batch_size"`

	// Performance settings
	Workers      int  `mapstructure:"workers"`
	EnableCache  bool `mapstructure:"enable_cache"`
	CacheSize    int  `mapstructure:"cache_size"`
	StreamWrites bool `mapstructure:"stream_writes"`

	// Logging settings
	LogLevel  string `mapstructure:"log_level"`
	LogFormat string `mapstructure:"log_format"`
	LogFile   string `mapstructure:"log_file"`

	// Output settings
	ColorOutput   bool `mapstructure:"color_output"`
	ProgressBar   bool `mapstructure:"progress_bar"`
	QuietMode     bool `mapstructure:"quiet_mode"`
	JSONOutput    bool `mapstructure:"json_output"`
	PrettyPrint   bool `mapstructure:"pretty_print"`
	ShowTimestamp bool `mapstructure:"show_timestamp"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		// Global settings
		Verbose: false,
		Config:  "",

		// Generation settings
		DefaultSeed:      0, // 0 means use current time
		DefaultFormat:    "sql",
		DefaultRowCount:  100,
		DefaultBatchSize: 1000,

		// Performance settings
		Workers:      4, // Number of concurrent workers
		EnableCache:  true,
		CacheSize:    10000, // LRU cache size for FK lookups
		StreamWrites: true,

		// Logging settings
		LogLevel:  "info", // debug, info, warn, error
		LogFormat: "text", // text, json
		LogFile:   "",     // empty means stderr

		// Output settings
		ColorOutput:   true, // Auto-detect terminal support
		ProgressBar:   true,
		QuietMode:     false,
		JSONOutput:    false,
		PrettyPrint:   true,
		ShowTimestamp: false,
	}
}

// InitConfig initializes the configuration system
// Configuration precedence (highest to lowest):
// 1. Command-line flags
// 2. Environment variables (DATAGEN_*)
// 3. Config file (.datagen.yaml)
// 4. Defaults
func InitConfig(cfgFile string) (*Config, error) {
	// Set config file if provided via flag
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config file in standard locations
		// 1. Current directory
		// 2. Home directory
		// 3. /etc/datagen/

		// Add current directory
		viper.AddConfigPath(".")

		// Add home directory
		if home, err := os.UserHomeDir(); err == nil {
			viper.AddConfigPath(home)
		}

		// Add system config directory
		viper.AddConfigPath("/etc/datagen")

		// Set config name (without extension)
		viper.SetConfigName(".datagen")
		viper.SetConfigType("yaml")
	}

	// Environment variables
	viper.SetEnvPrefix("DATAGEN") // will be uppercased automatically
	viper.AutomaticEnv()           // read in environment variables that match

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; using defaults and env vars
			// This is not an error - config file is optional
		} else {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal config into struct
	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Validate config
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// setDefaults sets default values for all config options
func setDefaults() {
	defaults := DefaultConfig()

	// Global settings
	viper.SetDefault("verbose", defaults.Verbose)

	// Generation settings
	viper.SetDefault("default_seed", defaults.DefaultSeed)
	viper.SetDefault("default_format", defaults.DefaultFormat)
	viper.SetDefault("default_row_count", defaults.DefaultRowCount)
	viper.SetDefault("default_batch_size", defaults.DefaultBatchSize)

	// Performance settings
	viper.SetDefault("workers", defaults.Workers)
	viper.SetDefault("enable_cache", defaults.EnableCache)
	viper.SetDefault("cache_size", defaults.CacheSize)
	viper.SetDefault("stream_writes", defaults.StreamWrites)

	// Logging settings
	viper.SetDefault("log_level", defaults.LogLevel)
	viper.SetDefault("log_format", defaults.LogFormat)
	viper.SetDefault("log_file", defaults.LogFile)

	// Output settings
	viper.SetDefault("color_output", defaults.ColorOutput)
	viper.SetDefault("progress_bar", defaults.ProgressBar)
	viper.SetDefault("quiet_mode", defaults.QuietMode)
	viper.SetDefault("json_output", defaults.JSONOutput)
	viper.SetDefault("pretty_print", defaults.PrettyPrint)
	viper.SetDefault("show_timestamp", defaults.ShowTimestamp)
}

// validateConfig validates the configuration values
func validateConfig(cfg *Config) error {
	// Validate format
	validFormats := map[string]bool{
		"sql":  true,
		"copy": true,
	}
	if !validFormats[cfg.DefaultFormat] {
		return fmt.Errorf("invalid default_format '%s', must be one of: sql, copy", cfg.DefaultFormat)
	}

	// Validate row count
	if cfg.DefaultRowCount < 0 {
		return fmt.Errorf("default_row_count must be >= 0, got %d", cfg.DefaultRowCount)
	}

	// Validate batch size
	if cfg.DefaultBatchSize <= 0 {
		return fmt.Errorf("default_batch_size must be > 0, got %d", cfg.DefaultBatchSize)
	}

	// Validate workers
	if cfg.Workers <= 0 {
		return fmt.Errorf("workers must be > 0, got %d", cfg.Workers)
	}
	if cfg.Workers > 100 {
		return fmt.Errorf("workers must be <= 100, got %d", cfg.Workers)
	}

	// Validate cache size
	if cfg.CacheSize < 0 {
		return fmt.Errorf("cache_size must be >= 0, got %d", cfg.CacheSize)
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[cfg.LogLevel] {
		return fmt.Errorf("invalid log_level '%s', must be one of: debug, info, warn, error", cfg.LogLevel)
	}

	// Validate log format
	validLogFormats := map[string]bool{
		"text": true,
		"json": true,
	}
	if !validLogFormats[cfg.LogFormat] {
		return fmt.Errorf("invalid log_format '%s', must be one of: text, json", cfg.LogFormat)
	}

	// Validate log file path (if specified)
	if cfg.LogFile != "" {
		// Check if directory exists
		dir := filepath.Dir(cfg.LogFile)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("log file directory does not exist: %s", dir)
		}
	}

	return nil
}

// GetConfigFilePath returns the path to the config file being used
func GetConfigFilePath() string {
	return viper.ConfigFileUsed()
}

// WriteConfigFile writes a sample configuration file
func WriteConfigFile(path string) error {
	// Create viper instance with defaults
	v := viper.New()

	// Set all defaults
	defaults := DefaultConfig()
	v.Set("verbose", defaults.Verbose)
	v.Set("default_seed", defaults.DefaultSeed)
	v.Set("default_format", defaults.DefaultFormat)
	v.Set("default_row_count", defaults.DefaultRowCount)
	v.Set("default_batch_size", defaults.DefaultBatchSize)
	v.Set("workers", defaults.Workers)
	v.Set("enable_cache", defaults.EnableCache)
	v.Set("cache_size", defaults.CacheSize)
	v.Set("stream_writes", defaults.StreamWrites)
	v.Set("log_level", defaults.LogLevel)
	v.Set("log_format", defaults.LogFormat)
	v.Set("log_file", defaults.LogFile)
	v.Set("color_output", defaults.ColorOutput)
	v.Set("progress_bar", defaults.ProgressBar)
	v.Set("quiet_mode", defaults.QuietMode)
	v.Set("json_output", defaults.JSONOutput)
	v.Set("pretty_print", defaults.PrettyPrint)
	v.Set("show_timestamp", defaults.ShowTimestamp)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write config file
	if err := v.WriteConfigAs(path); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// PrintConfig prints the current configuration to stdout
func PrintConfig(cfg *Config) {
	fmt.Println("Current Configuration:")
	fmt.Println()
	fmt.Println("Global Settings:")
	fmt.Printf("  Verbose:           %v\n", cfg.Verbose)
	fmt.Println()
	fmt.Println("Generation Settings:")
	fmt.Printf("  Default Seed:      %d\n", cfg.DefaultSeed)
	fmt.Printf("  Default Format:    %s\n", cfg.DefaultFormat)
	fmt.Printf("  Default Row Count: %d\n", cfg.DefaultRowCount)
	fmt.Printf("  Default Batch Size: %d\n", cfg.DefaultBatchSize)
	fmt.Println()
	fmt.Println("Performance Settings:")
	fmt.Printf("  Workers:           %d\n", cfg.Workers)
	fmt.Printf("  Enable Cache:      %v\n", cfg.EnableCache)
	fmt.Printf("  Cache Size:        %d\n", cfg.CacheSize)
	fmt.Printf("  Stream Writes:     %v\n", cfg.StreamWrites)
	fmt.Println()
	fmt.Println("Logging Settings:")
	fmt.Printf("  Log Level:         %s\n", cfg.LogLevel)
	fmt.Printf("  Log Format:        %s\n", cfg.LogFormat)
	fmt.Printf("  Log File:          %s\n", cfg.LogFile)
	fmt.Println()
	fmt.Println("Output Settings:")
	fmt.Printf("  Color Output:      %v\n", cfg.ColorOutput)
	fmt.Printf("  Progress Bar:      %v\n", cfg.ProgressBar)
	fmt.Printf("  Quiet Mode:        %v\n", cfg.QuietMode)
	fmt.Printf("  JSON Output:       %v\n", cfg.JSONOutput)
	fmt.Printf("  Pretty Print:      %v\n", cfg.PrettyPrint)
	fmt.Printf("  Show Timestamp:    %v\n", cfg.ShowTimestamp)

	if cfgFile := GetConfigFilePath(); cfgFile != "" {
		fmt.Println()
		fmt.Printf("Config file: %s\n", cfgFile)
	}
}
