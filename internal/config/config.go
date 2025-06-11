package config

import (
	"github.com/Azure/aks-mcp/internal/security"
	flag "github.com/spf13/pflag"
)

// ConfigData holds the global configuration
type ConfigData struct {
	// Command execution timeout in seconds
	Timeout int
	// Security configuration
	SecurityConfig *security.SecurityConfig

	// Command-line specific options
	Transport   string
	Host        string
	Port        int
	AccessLevel string
}

// NewConfig creates and returns a new configuration instance
func NewConfig() *ConfigData {
	return &ConfigData{
		Timeout:        60,
		SecurityConfig: security.NewSecurityConfig(),
		Transport:      "stdio",
		Port:           8000,
		AccessLevel:    "readonly",
	}
}

// ParseFlags parses command line arguments and updates the configuration
func (cfg *ConfigData) ParseFlags() {
	// Server configuration
	flag.StringVar(&cfg.Transport, "transport", "stdio", "Transport mechanism to use (stdio, sse or streamable-http)")
	flag.StringVar(&cfg.Host, "host", "127.0.0.1", "Host to listen for the server (only used with transport sse or streamable-http)")
	flag.IntVar(&cfg.Port, "port", 8000, "Port to listen for the server (only used with transport sse or streamable-http)")
	flag.IntVar(&cfg.Timeout, "timeout", 600, "Timeout for command execution in seconds, default is 600s")
	// Security settings
	flag.StringVar(&cfg.AccessLevel, "access-level", "readonly", "Access level (readonly, readwrite, admin)")

	flag.Parse()

	// Update security config
	cfg.SecurityConfig.AccessLevel = cfg.AccessLevel
}
