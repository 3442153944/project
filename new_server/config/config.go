package config

import (
	"fmt"
	"github.com/spf13/viper"
)

// Conf 全局配置实例
var Conf = new(Config)

// Config 根节点配置，完全对齐你的 YAML
type Config struct {
	DB        DBConfig     `mapstructure:"db"`
	Log       LogConfig    `mapstructure:"log"`
	Whitelist []string     `mapstructure:"whitelist"` // 白名单路由
	Auth      AuthConfig   `mapstructure:"auth"`
	Server    ServerConfig `mapstructure:"server"`
	Redis     RedisConfig  `mapstructure:"redis"`
}

// DBConfig 数据库配置 (注意：这里将 uri 拆分为 host 和 port 以适配 GORM)
type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// RedisConfig 缓存配置
type RedisConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	DB   int    `mapstructure:"db"`
}

// LogConfig 日志配置
type LogConfig struct {
	Path      string `mapstructure:"path"`
	Level     string `mapstructure:"level"`
	Format    string `mapstructure:"format"`
	Console   bool   `mapstructure:"console"`
	File      bool   `mapstructure:"file"`
	MaxSize   int    `mapstructure:"max_size"`
	MaxAge    int    `mapstructure:"max_age"`
	MaxBackup int    `mapstructure:"max_backup"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	TokenExpire   int    `mapstructure:"token_expire"`
	RefreshExpire int    `mapstructure:"refresh_expire"`
	Secret        string `mapstructure:"secret"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// Init 初始化 Viper 并解析 YAML
func Init() error {
	viper.SetConfigName("config") // 你的 yaml 文件名 (不带后缀)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config") // 假设配置文件在项目根目录

	// 开启环境变量覆盖机制 (非常重要：用于生产环境覆盖 Secret 等敏感信息)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := viper.Unmarshal(Conf); err != nil {
		return fmt.Errorf("解析配置到结构体失败: %w", err)
	}

	return nil
}
