package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Security SecurityConfig `mapstructure:"security"`
	Sync     SyncConfig     `mapstructure:"sync"`
	DebugLog LogConfig      `mapstructure:"debugLog"`
	DevLog   LogConfig      `mapstructure:"devLog"`
	ProdLog  LogConfig      `mapstructure:"prodLog"`
	Token    TokenConfig    `mapstructure:"token"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, dev, prod
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	ExpireHours        int    `mapstructure:"expire_hours"`
	RefreshExpireHours int    `mapstructure:"refresh_expire_hours"`
}

type SecurityConfig struct {
	BcryptCost             int `mapstructure:"bcrypt_cost"`
	MaxLoginAttempts       int `mapstructure:"max_login_attempts"`
	LockoutDurationMinutes int `mapstructure:"lockout_duration_minutes"`
}

type SyncConfig struct {
	IntervalSeconds int `mapstructure:"interval_seconds"`
	BatchSize       int `mapstructure:"batch_size"`
}

// LogConfig 日志配置
type LogConfig struct {
	Enabled    bool   `mapstructure:"enabled"`    // 是否启用
	FileSize   int    `mapstructure:"fileSize"`   // 单个文件大小(MB)
	MaxBackups int    `mapstructure:"maxBackups"` // 保留的旧文件数
	Path       string `mapstructure:"path"`       // 日志目录
}

// TokenConfig Token 配置
type TokenConfig struct {
	ValidityDate int `mapstructure:"validityDate"` // 有效期（分钟）
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// validateConfig 验证配置的有效性
func validateConfig(cfg *Config) error {
	// 验证服务器模式
	validModes := map[string]bool{"debug": true, "dev": true, "prod": true}
	if !validModes[cfg.Server.Mode] {
		return fmt.Errorf("无效的服务器模式: %s (可选: debug, dev, prod)", cfg.Server.Mode)
	}

	// 验证日志配置
	logConfigs := map[string]LogConfig{
		"debugLog": cfg.DebugLog,
		"devLog":   cfg.DevLog,
		"prodLog":  cfg.ProdLog,
	}

	for name, logCfg := range logConfigs {
		if logCfg.Enabled {
			if logCfg.Path == "" {
				return fmt.Errorf("%s.path 不能为空", name)
			}
			if logCfg.FileSize <= 0 {
				return fmt.Errorf("%s.fileSize 必须大于 0", name)
			}
			if logCfg.MaxBackups < 0 {
				return fmt.Errorf("%s.maxBackups 不能为负数", name)
			}
		}
	}

	// 验证 Token 配置
	if cfg.Token.ValidityDate <= 0 {
		return fmt.Errorf("token.validityDate 必须大于 0")
	}

	return nil
}

// GetActiveLogConfig 获取当前激活的日志配置
func (c *Config) GetActiveLogConfig() LogConfig {
	switch c.Server.Mode {
	case "debug":
		return c.DebugLog
	case "dev":
		return c.DevLog
	case "prod":
		return c.ProdLog
	default:
		return c.DebugLog
	}
}
