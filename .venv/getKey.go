package _venv

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// KeyConfig 密钥配置
type KeyConfig struct {
	JSONKey       string `yaml:"jsonKey"`
	EncryptionKey string `yaml:"encryptionKey"`
}

var (
	// 全局密钥配置
	Keys *KeyConfig
)

// init 包初始化时自动加载密钥
func init() {
	var err error
	Keys, err = LoadKeys()
	if err != nil {
		panic(fmt.Sprintf("加载密钥失败: %v", err))
	}
}

// LoadKeys 从 key.yaml 加载密钥
func LoadKeys() (*KeyConfig, error) {
	path := "key.yaml"

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取密钥文件失败: %w", err)
	}

	var config KeyConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析 key.yaml 失败: %w", err)
	}

	// 验证密钥
	if config.JSONKey == "" {
		return nil, fmt.Errorf("jsonKey 为空，请检查 key.yaml")
	}
	if config.EncryptionKey == "" {
		return nil, fmt.Errorf("encryptionKey 为空，请检查 key.yaml")
	}

	return &config, nil
}

// GetJSONKey 获取 JSON 密钥
func GetJSONKey() string {
	return Keys.JSONKey
}

// GetEncryptionKey 获取加密密钥
func GetEncryptionKey() string {
	return Keys.EncryptionKey
}
