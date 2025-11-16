package config

import (
	"fmt"
	"runtime"
	"strings"

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
	File     FileConfig     `mapstructure:"file"`
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

// FileConfig 文件存储配置
type FileConfig struct {
	Mode        string        `mapstructure:"mode"`        // windows / linux（自动检测）
	WindowsPath []string      `mapstructure:"windowsPath"` // Windows允许的盘符
	LinuxPath   []string      `mapstructure:"linuxPath"`   // Linux允许的挂载点
	Storage     StorageConfig `mapstructure:"storage"`     // 存储配置
	Upload      UploadConfig  `mapstructure:"upload"`      // 上传配置
}

// StorageConfig 存储详细配置
type StorageConfig struct {
	BasePath       string `mapstructure:"basePath"`       // 存储根目录（默认：FileSync）
	UploadPath     string `mapstructure:"uploadPath"`     // 上传目录（默认：uploads）
	TempPath       string `mapstructure:"tempPath"`       // 临时目录（默认：temp）
	TrashPath      string `mapstructure:"trashPath"`      // 回收站目录（默认：trash）
	MinFreeSpace   int64  `mapstructure:"minFreeSpace"`   // 最小剩余空间（字节，默认：5GB）
	MaxStorageSize int64  `mapstructure:"maxStorageSize"` // 单盘最大使用（字节，默认：100GB）
}

// UploadConfig 上传配置
type UploadConfig struct {
	MaxFileSize         int64    `mapstructure:"maxFileSize"`         // 最大文件大小（字节，默认：10GB）
	ChunkEnabled        bool     `mapstructure:"chunkEnabled"`        // 是否启用分片（默认：true）
	ChunkSize           int64    `mapstructure:"chunkSize"`           // 分片大小（字节，默认：2MB）
	MaxChunks           int      `mapstructure:"maxChunks"`           // 最大分片数（默认：5120）
	ConcurrentUploads   int      `mapstructure:"concurrentUploads"`   // 并发上传数（默认：5）
	AllowedExtensions   []string `mapstructure:"allowedExtensions"`   // 允许的扩展名（为空则不限制）
	ForbiddenExtensions []string `mapstructure:"forbiddenExtensions"` // 禁止的扩展名
	MaxFilenameLength   int      `mapstructure:"maxFilenameLength"`   // 最大文件名长度（默认：255）
	TempCleanInterval   int      `mapstructure:"tempCleanInterval"`   // 临时文件清理间隔（秒，默认：3600）
	TempMaxAge          int      `mapstructure:"tempMaxAge"`          // 临时文件最长保留（秒，默认：86400）
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

	// 设置默认值
	config.setDefaults()

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// setDefaults 设置默认值
func (c *Config) setDefaults() {
	// 自动检测操作系统
	if c.File.Mode == "" {
		if runtime.GOOS == "windows" {
			c.File.Mode = "windows"
		} else {
			c.File.Mode = "linux"
		}
	}

	// 存储路径默认值
	if c.File.Storage.BasePath == "" {
		c.File.Storage.BasePath = "FileSync"
	}
	if c.File.Storage.UploadPath == "" {
		c.File.Storage.UploadPath = "uploads"
	}
	if c.File.Storage.TempPath == "" {
		c.File.Storage.TempPath = "temp"
	}
	if c.File.Storage.TrashPath == "" {
		c.File.Storage.TrashPath = "trash"
	}
	if c.File.Storage.MinFreeSpace == 0 {
		c.File.Storage.MinFreeSpace = 5 * 1024 * 1024 * 1024 // 5GB
	}
	if c.File.Storage.MaxStorageSize == 0 {
		c.File.Storage.MaxStorageSize = 100 * 1024 * 1024 * 1024 // 100GB
	}

	// 上传配置默认值
	if c.File.Upload.MaxFileSize == 0 {
		c.File.Upload.MaxFileSize = 10 * 1024 * 1024 * 1024 // 10GB
	}
	if c.File.Upload.ChunkSize == 0 {
		c.File.Upload.ChunkSize = 2 * 1024 * 1024 // 2MB
	}
	if c.File.Upload.MaxChunks == 0 {
		c.File.Upload.MaxChunks = 5120
	}
	if c.File.Upload.ConcurrentUploads == 0 {
		c.File.Upload.ConcurrentUploads = 5
	}
	if c.File.Upload.MaxFilenameLength == 0 {
		c.File.Upload.MaxFilenameLength = 255
	}
	if c.File.Upload.TempCleanInterval == 0 {
		c.File.Upload.TempCleanInterval = 3600 // 1小时
	}
	if c.File.Upload.TempMaxAge == 0 {
		c.File.Upload.TempMaxAge = 86400 // 24小时
	}
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

	// 验证文件存储配置
	if err := cfg.validateFileConfig(); err != nil {
		return fmt.Errorf("文件配置验证失败: %w", err)
	}

	return nil
}

// validateFileConfig 验证文件配置
func (c *Config) validateFileConfig() error {
	// 验证存储路径
	allowedPaths := c.GetAllowedPaths()
	if len(allowedPaths) == 0 {
		return fmt.Errorf("必须配置至少一个存储路径")
	}

	// 验证上传配置
	if c.File.Upload.MaxFileSize <= 0 {
		return fmt.Errorf("maxFileSize 必须大于 0")
	}
	if c.File.Upload.ChunkEnabled && c.File.Upload.ChunkSize <= 0 {
		return fmt.Errorf("启用分片时 chunkSize 必须大于 0")
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

// GetAllowedPaths 获取当前系统的允许路径
func (c *Config) GetAllowedPaths() []string {
	if c.File.Mode == "windows" {
		return c.File.WindowsPath
	}
	return c.File.LinuxPath
}

// GetDefaultPath 获取默认存储路径（第一个）
func (c *Config) GetDefaultPath() string {
	paths := c.GetAllowedPaths()
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

// IsPathAllowed 检查路径是否在允许列表中
func (c *Config) IsPathAllowed(path string) bool {
	path = strings.ToUpper(path) // 统一转大写比较（Windows盘符）
	for _, allowed := range c.GetAllowedPaths() {
		if strings.ToUpper(allowed) == path {
			return true
		}
	}
	return false
}

// GetStoragePath 获取完整存储路径
// disk: 盘符或挂载点（如 "D:" 或 "/home"）
// subPath: 子路径（如 "uploads", "temp"）
func (c *Config) GetStoragePath(disk, subPath string) string {
	if disk == "" {
		disk = c.GetDefaultPath()
	}

	// Windows: D:/FileSync/uploads
	// Linux: /home/FileSync/uploads
	separator := "/"
	if c.File.Mode == "windows" {
		separator = "/"
	}

	return fmt.Sprintf("%s%s%s%s%s",
		disk,
		separator,
		c.File.Storage.BasePath,
		separator,
		subPath,
	)
}

// GetUploadPath 获取上传目录完整路径
func (c *Config) GetUploadPath(disk string) string {
	return c.GetStoragePath(disk, c.File.Storage.UploadPath)
}

// GetTempPath 获取临时目录完整路径
func (c *Config) GetTempPath(disk string) string {
	return c.GetStoragePath(disk, c.File.Storage.TempPath)
}

// GetTrashPath 获取回收站目录完整路径
func (c *Config) GetTrashPath(disk string) string {
	return c.GetStoragePath(disk, c.File.Storage.TrashPath)
}

// IsExtensionAllowed 检查文件扩展名是否允许
func (c *Config) IsExtensionAllowed(ext string) bool {
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	// 检查禁止列表
	for _, forbidden := range c.File.Upload.ForbiddenExtensions {
		if strings.ToLower(forbidden) == ext {
			return false
		}
	}

	// 如果允许列表为空，则允许所有（除禁止列表外）
	if len(c.File.Upload.AllowedExtensions) == 0 {
		return true
	}

	// 检查允许列表
	for _, allowed := range c.File.Upload.AllowedExtensions {
		if strings.ToLower(allowed) == ext {
			return true
		}
	}

	return false
}
