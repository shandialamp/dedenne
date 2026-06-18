package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	// Server
	Server ServerConfig `mapstructure:"server"`

	// JWT
	JWT JWTConfig `mapstructure:"jwt"`

	// Log
	Log LogConfig `mapstructure:"log"`

	// Database
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	Expiration int    `mapstructure:"expiration"` // seconds
}

type LogConfig struct {
	Level      string `mapstructure:"level"`      // debug, info, warn, error
	Format     string `mapstructure:"format"`     // json, console
	OutputPath string `mapstructure:"outputPath"` // stdout, stderr, file path
}

type DatabaseConfig struct {
	Type         string `mapstructure:"type"`         // sqlite, mysql
	DSN          string `mapstructure:"dsn"`          // Data Source Name
	MaxOpenConn  int    `mapstructure:"maxOpenConn"`  // 最大打开连接数
	MaxIdleConn  int    `mapstructure:"maxIdleConn"`  // 最大空闲连接数
	MaxLifetime  int    `mapstructure:"maxLifetime"`  // 连接最大生命周期（秒）
}

var globalConfig *Config

// ReadConfig 读取配置文件并初始化全局配置
func ReadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 如果提供了配置文件路径，则读取
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 绑定环境变量
	v.AutomaticEnv()
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.host", "SERVER_HOST")
	v.BindEnv("jwt.secret", "JWT_SECRET")
	v.BindEnv("jwt.expiration", "JWT_EXPIRATION")
	v.BindEnv("log.level", "LOG_LEVEL")
	v.BindEnv("log.format", "LOG_FORMAT")
	v.BindEnv("log.outputPath", "LOG_OUTPUT_PATH")
	v.BindEnv("database.type", "DATABASE_TYPE")
	v.BindEnv("database.dsn", "DATABASE_DSN")
	v.BindEnv("database.maxOpenConn", "DATABASE_MAX_OPEN_CONN")
	v.BindEnv("database.maxIdleConn", "DATABASE_MAX_IDLE_CONN")
	v.BindEnv("database.maxLifetime", "DATABASE_MAX_LIFETIME")

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 验证并修正必要字段
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	globalConfig = cfg
	return cfg, nil
}

// setDefaults 设置所有配置的默认值
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "0.0.0.0")

	// JWT defaults
	v.SetDefault("jwt.secret", "your-secret-key-change-in-production")
	v.SetDefault("jwt.expiration", 3600) // 1 hour

	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.outputPath", "stdout")

	// Database defaults
	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.dsn", "app.db")
	v.SetDefault("database.maxOpenConn", 25)
	v.SetDefault("database.maxIdleConn", 5)
	v.SetDefault("database.maxLifetime", 5*60) // 5 minutes
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		c.Server.Port = 8080
	}

	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}

	if c.JWT.Secret == "" {
		c.JWT.Secret = "your-secret-key-change-in-production"
	}

	if c.JWT.Expiration <= 0 {
		c.JWT.Expiration = 3600
	}

	if c.Log.Level == "" {
		c.Log.Level = "info"
	}

	if c.Log.Format == "" {
		c.Log.Format = "json"
	}

	if c.Log.OutputPath == "" {
		c.Log.OutputPath = "stdout"
	}

	// Database validation
	if c.Database.Type == "" {
		c.Database.Type = "sqlite"
	}

	if c.Database.DSN == "" {
		c.Database.DSN = "app.db"
	}

	if c.Database.MaxOpenConn <= 0 {
		c.Database.MaxOpenConn = 25
	}

	if c.Database.MaxIdleConn <= 0 {
		c.Database.MaxIdleConn = 5
	}

	if c.Database.MaxLifetime <= 0 {
		c.Database.MaxLifetime = 5 * 60
	}

	return nil
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		panic("config not initialized, call ReadConfig first")
	}
	return globalConfig
}
