package token

import (
	"encoding/json"
	"errors"
	"fmt"
	_venv "github.com/sunyuanling/server/venv"
	"time"

	"github.com/sunyuanling/server/config"
	"github.com/sunyuanling/server/encryption"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// token 全局单例
var globalTokenManager *TokenManager

// InitGlobalTokenManager 初始化全局TokenManager
func InitGlobalTokenManager(cfg *config.Config) error {
	tm, err := NewTokenManager(cfg)
	if err != nil {
		return err
	}
	globalTokenManager = tm
	return nil
}

// GetGlobalTokenManager 获取token实例
func GetGlobalTokenManager() *TokenManager {
	return globalTokenManager
}

// TokenManager Token 管理器
type TokenManager struct {
	encryptor    *encryption.Encryptor
	validityDate int // 有效期（分钟）
}

// TokenPayload Token 载荷
type TokenPayload struct {
	UserID    int64                  `json:"user_id"`
	Username  string                 `json:"username"`
	Email     string                 `json:"email,omitempty"`
	Roles     []string               `json:"roles,omitempty"`
	ExtraData map[string]interface{} `json:"extra_data,omitempty"`
	IssuedAt  int64                  `json:"issued_at"`  // 签发时间戳
	ExpiresAt int64                  `json:"expires_at"` // 过期时间戳
}

// NewTokenManager 创建 Token 管理器
func NewTokenManager(cfg *config.Config) (*TokenManager, error) {
	// 获取加密密钥
	encryptionKey := _venv.GetEncryptionKey()

	// 创建加密器
	encryptor, err := encryption.NewEncryptor(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("创建加密器失败: %w", err)
	}

	return &TokenManager{
		encryptor:    encryptor,
		validityDate: cfg.Token.ValidityDate,
	}, nil
}

// GenerateToken 生成 Token（从 JSON 数据）
func (tm *TokenManager) GenerateToken(jsonData string) (string, error) {
	// 解析 JSON 到 map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return "", fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 添加时间戳
	now := time.Now()
	data["issued_at"] = now.Unix()
	data["expires_at"] = now.Add(time.Duration(tm.validityDate) * time.Minute).Unix()

	// 序列化为 JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("序列化 JSON 失败: %w", err)
	}

	// 加密生成 Token
	token, err := tm.encryptor.Encrypt(string(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("加密失败: %w", err)
	}

	return token, nil
}

// GenerateTokenFromPayload 从 Payload 生成 Token
func (tm *TokenManager) GenerateTokenFromPayload(payload *TokenPayload) (string, error) {
	// 设置时间戳
	now := time.Now()
	payload.IssuedAt = now.Unix()
	payload.ExpiresAt = now.Add(time.Duration(tm.validityDate) * time.Minute).Unix()

	// 序列化为 JSON
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("序列化失败: %w", err)
	}

	// 加密生成 Token
	token, err := tm.encryptor.Encrypt(string(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("加密失败: %w", err)
	}

	return token, nil
}

// GenerateTokenFromMap 从 Map 生成 Token
func (tm *TokenManager) GenerateTokenFromMap(data map[string]interface{}) (string, error) {
	// 添加时间戳
	now := time.Now()
	data["issued_at"] = now.Unix()
	data["expires_at"] = now.Add(time.Duration(tm.validityDate) * time.Minute).Unix()

	// 加密生成 Token
	token, err := tm.encryptor.EncryptMap(data)
	if err != nil {
		return "", fmt.Errorf("加密失败: %w", err)
	}

	return token, nil
}

// ValidateToken 验证并解析 Token
func (tm *TokenManager) ValidateToken(token string) (*TokenPayload, error) {
	// 解密 Token
	plaintext, err := tm.encryptor.Decrypt(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// 解析为 Payload
	var payload TokenPayload
	if err := json.Unmarshal([]byte(plaintext), &payload); err != nil {
		return nil, ErrInvalidToken
	}

	// 验证过期时间
	if time.Now().Unix() > payload.ExpiresAt {
		return nil, ErrExpiredToken
	}

	return &payload, nil
}

// ValidateTokenToMap 验证并解析 Token 为 Map
func (tm *TokenManager) ValidateTokenToMap(token string) (map[string]interface{}, error) {
	// 解密 Token
	data, err := tm.encryptor.DecryptToMap(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// 验证过期时间
	expiresAt, ok := data["expires_at"].(float64) // JSON 数字默认是 float64
	if !ok {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() > int64(expiresAt) {
		return nil, ErrExpiredToken
	}

	return data, nil
}

// RefreshToken 刷新 Token（延长有效期）
func (tm *TokenManager) RefreshToken(oldToken string) (string, error) {
	// 解析旧 Token
	payload, err := tm.ValidateToken(oldToken)
	if err != nil {
		return "", err
	}

	// 生成新 Token（自动更新时间戳）
	return tm.GenerateTokenFromPayload(payload)
}

// GetRemainingTime 获取 Token 剩余有效时间（秒）
func (tm *TokenManager) GetRemainingTime(token string) (int64, error) {
	payload, err := tm.ValidateToken(token)
	if err != nil {
		return 0, err
	}

	remaining := payload.ExpiresAt - time.Now().Unix()
	if remaining < 0 {
		return 0, ErrExpiredToken
	}

	return remaining, nil
}
