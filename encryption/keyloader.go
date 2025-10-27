package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// KeyConfig 密钥配置结构
type KeyConfig struct {
	Key string `yaml:"key"`
}

// LoadKeyFromFile 从YAML文件加载密钥
func LoadKeyFromFile(filepath string) (*Encryptor, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	var config KeyConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse key file: %w", err)
	}

	return NewEncryptor(config.Key)
}

// GenerateKey 生成新的AES-256密钥（32字节）
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// SaveKeyToFile 保存密钥到YAML文件
func SaveKeyToFile(filepath string, key string) error {
	config := KeyConfig{Key: key}
	data, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0600) // 仅所有者可读写
}
