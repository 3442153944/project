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
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
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

type LogConfig struct {
	Level    string `mapstructure:"level"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
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

	return &config, nil
}
