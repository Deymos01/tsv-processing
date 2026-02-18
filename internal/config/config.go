package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Worker   WorkerConfig
	Files    FilesConfig
}

type AppConfig struct {
	Env string
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DSN builds a PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

type WorkerConfig struct {
	PoolSize     int
	QueueSize    int
	ScanInterval time.Duration
}

type FilesConfig struct {
	InputDir  string
	OutputDir string
}

// Load reads configuration from the given path and overlays env vars.
func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := &Config{}

	cfg.App.Env = viper.GetString("app.env")

	cfg.Server.Host = viper.GetString("server.host")
	cfg.Server.Port = viper.GetInt("server.port")
	cfg.Server.ReadTimeout = viper.GetDuration("server.read_timeout")
	cfg.Server.WriteTimeout = viper.GetDuration("server.write_timeout")
	cfg.Server.IdleTimeout = viper.GetDuration("server.idle_timeout")

	cfg.Database.Host = viper.GetString("database.host")
	cfg.Database.Port = viper.GetInt("database.port")
	cfg.Database.User = viper.GetString("database.user")
	cfg.Database.Password = viper.GetString("database.password")
	cfg.Database.DBName = viper.GetString("database.db_name")
	cfg.Database.SSLMode = viper.GetString("database.ssl_mode")
	cfg.Database.MaxOpenConns = viper.GetInt("database.max_open_conns")
	cfg.Database.MaxIdleConns = viper.GetInt("database.max_idle_conns")
	cfg.Database.ConnMaxLifetime = viper.GetDuration("database.conn_max_lifetime")

	cfg.Worker.PoolSize = viper.GetInt("worker.pool_size")
	cfg.Worker.QueueSize = viper.GetInt("worker.queue_size")
	cfg.Worker.ScanInterval = viper.GetDuration("worker.scan_interval")

	cfg.Files.InputDir = viper.GetString("files.input_dir")
	cfg.Files.OutputDir = viper.GetString("files.output_dir")

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// validate performs basic sanity checks on the loaded configuration.
func (c *Config) validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535, got %d", c.Server.Port)
	}
	if c.Worker.PoolSize <= 0 {
		return fmt.Errorf("worker.pool_size must be > 0")
	}
	if c.Worker.QueueSize <= 0 {
		return fmt.Errorf("worker.queue_size must be > 0")
	}
	if c.Files.InputDir == "" {
		return fmt.Errorf("files.input_dir must not be empty")
	}
	if c.Files.OutputDir == "" {
		return fmt.Errorf("files.output_dir must not be empty")
	}
	return nil
}
