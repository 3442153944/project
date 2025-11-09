package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
)

// Decrypt 解密数据（输入Base64编码的密文）
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	if e.key == nil {
		return "", ErrKeyNotLoaded
	}

	// Base64解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	// 创建AES cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	// 使用GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 验证密文长度
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	// 提取nonce和密文
	nonce := data[:nonceSize]
	ciphertextBytes := data[nonceSize:] //  修正：不要重复转换

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	return string(plaintext), nil
}

// DecryptJSON 解密JSON数据
func (e *Encryptor) DecryptJSON(ciphertext string) ([]byte, error) {
	plaintext, err := e.Decrypt(ciphertext)
	if err != nil {
		return nil, err
	}
	return []byte(plaintext), nil
}

// DecryptToMap 解密并反序列化为map
func (e *Encryptor) DecryptToMap(ciphertext string) (map[string]interface{}, error) {
	plaintext, err := e.Decrypt(ciphertext)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(plaintext), &result); err != nil {
		return nil, err
	}

	return result, nil
}

// DecryptToStruct 解密并反序列化到结构体
func (e *Encryptor) DecryptToStruct(ciphertext string, v interface{}) error {
	plaintext, err := e.Decrypt(ciphertext)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(plaintext), v)
}
