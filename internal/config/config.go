package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	AppName                    string `mapstructure:"app_name"`
	Environment                string `mapstructure:"environment"`
	HTTPHost                   string `mapstructure:"http_host"`
	HTTPPort                   int    `mapstructure:"http_port"`
	HTTPReadTimeoutSeconds     int    `mapstructure:"http_read_timeout_seconds"`
	HTTPWriteTimeoutSeconds    int    `mapstructure:"http_write_timeout_seconds"`
	HTTPShutdownTimeoutSeconds int    `mapstructure:"http_shutdown_timeout_seconds"`
	DBDSN                      string `mapstructure:"db_dsn"`
	DBAutoMigrate              bool   `mapstructure:"db_auto_migrate"`
	LogLevel                   string `mapstructure:"log_level"`
	LogPretty                  bool   `mapstructure:"log_pretty"`
}

func Load(cmd *cobra.Command, cfgFile string) (*Config, error) {
	v := viper.New()
	setDefaults(v)

	if cmd != nil {
		if err := bindFlags(v, cmd); err != nil {
			return nil, err
		}
	}

	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app_name", "scaffold-api")
	v.SetDefault("environment", "dev")
	v.SetDefault("http_host", "0.0.0.0")
	v.SetDefault("http_port", 8080)
	v.SetDefault("http_read_timeout_seconds", 15)
	v.SetDefault("http_write_timeout_seconds", 15)
	v.SetDefault("http_shutdown_timeout_seconds", 10)
	v.SetDefault("db_auto_migrate", true)
	v.SetDefault("log_level", "info")
	v.SetDefault("log_pretty", true)
}

func bindFlags(v *viper.Viper, cmd *cobra.Command) error {
	flagBindings := map[string]string{
		"app_name":                      "app-name",
		"environment":                   "environment",
		"http_host":                     "http-host",
		"http_port":                     "http-port",
		"http_read_timeout_seconds":     "http-read-timeout-seconds",
		"http_write_timeout_seconds":    "http-write-timeout-seconds",
		"http_shutdown_timeout_seconds": "http-shutdown-timeout-seconds",
		"db_dsn":                        "db-dsn",
		"db_auto_migrate":               "db-auto-migrate",
		"log_level":                     "log-level",
		"log_pretty":                    "log-pretty",
	}

	for key, flagName := range flagBindings {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			flag = cmd.InheritedFlags().Lookup(flagName)
		}
		if flag == nil {
			continue
		}
		if err := v.BindPFlag(key, flag); err != nil {
			return fmt.Errorf("bind flag %s: %w", flagName, err)
		}
	}

	return nil
}

func (c Config) Validate() error {
	if c.HTTPPort <= 0 || c.HTTPPort > 65535 {
		return fmt.Errorf("invalid http port: %d", c.HTTPPort)
	}
	if c.DBDSN == "" {
		return fmt.Errorf("db_dsn is required")
	}
	if c.HTTPReadTimeoutSeconds <= 0 {
		return fmt.Errorf("http_read_timeout_seconds must be positive")
	}
	if c.HTTPWriteTimeoutSeconds <= 0 {
		return fmt.Errorf("http_write_timeout_seconds must be positive")
	}
	if c.HTTPShutdownTimeoutSeconds <= 0 {
		return fmt.Errorf("http_shutdown_timeout_seconds must be positive")
	}
	return nil
}

func (c Config) Address() string {
	return fmt.Sprintf("%s:%d", c.HTTPHost, c.HTTPPort)
}

func (c Config) ReadTimeout() time.Duration {
	return time.Duration(c.HTTPReadTimeoutSeconds) * time.Second
}

func (c Config) WriteTimeout() time.Duration {
	return time.Duration(c.HTTPWriteTimeoutSeconds) * time.Second
}

func (c Config) ShutdownTimeout() time.Duration {
	return time.Duration(c.HTTPShutdownTimeoutSeconds) * time.Second
}
