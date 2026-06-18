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

	return nil
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		panic("config not initialized, call ReadConfig first")
	}
	return globalConfig
}
