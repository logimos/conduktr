package config

// Config holds the application configuration
type Config struct {
	WorkflowDir string `mapstructure:"workflow_dir"`
	HTTPPort    int    `mapstructure:"http_port"`
	LogLevel    string `mapstructure:"log_level"`
	DataDir     string `mapstructure:"data_dir"`
}

// Default returns a configuration with default values
func Default() *Config {
	return &Config{
		WorkflowDir: "./workflows",
		HTTPPort:    8000,
		LogLevel:    "info",
		DataDir:     "./data",
	}
}
