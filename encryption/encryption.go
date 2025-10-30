package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
)

var (
	ErrInvalidKey        = errors.New("invalid encryption key")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrKeyNotLoaded      = errors.New("encryption key not loaded")
)

// Encryptor 加密器
type Encryptor struct {
	key []byte
}

// NewEncryptor 创建加密器
func NewEncryptor(key string) (*Encryptor, error) {
	// Base64解码密钥
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, ErrInvalidKey
	}

	// 验证密钥长度（AES-256需要32字节）
	if len(keyBytes) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}

	return &Encryptor{key: keyBytes}, nil
}

// Encrypt 加密数据（返回Base64编码的密文）
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if e.key == nil {
		return "", ErrKeyNotLoaded
	}

	// 创建AES cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	// 使用GCM模式（提供认证加密）
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密（nonce会自动添加到密文前面）
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Base64编码返回
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// EncryptJSON 加密JSON数据
func (e *Encryptor) EncryptJSON(jsonData []byte) (string, error) {
	return e.Encrypt(string(jsonData))
}

// EncryptMap 加密map数据（自动序列化为JSON）
func (e *Encryptor) EncryptMap(data map[string]interface{}) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return e.Encrypt(string(jsonBytes))
}

// EncryptStruct 加密结构体（自动序列化为JSON）
func (e *Encryptor) EncryptStruct(v interface{}) (string, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return e.Encrypt(string(jsonBytes))
}
