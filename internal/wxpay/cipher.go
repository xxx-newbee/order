package wxpay

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

type AESCipher struct {
	key []byte
}

func NewAESCipher(apiV3Key string) (*AESCipher, error) {
	key, err := hexDecode(apiV3Key)
	if err != nil {
		return nil, fmt.Errorf("invalid api v3 key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("api v3 key must be 32 bytes, got %d", len(key))
	}
	return &AESCipher{key: key}, nil
}

func (c *AESCipher) Decrypt(ciphertext, nonce, additionalData string) ([]byte, error) {
	ct, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plain, err := aesgcm.Open(nil, []byte(nonce), ct, []byte(additionalData))
	if err != nil {
		return nil, fmt.Errorf("gcm open: %w", err)
	}
	return plain, nil
}

func hexDecode(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("hex string has odd length")
	}
	dst := make([]byte, len(s)/2)
	for i := 0; i < len(dst); i++ {
		hi := unhex(s[2*i])
		lo := unhex(s[2*i+1])
		if hi < 0 || lo < 0 {
			return nil, fmt.Errorf("invalid hex char")
		}
		dst[i] = byte(hi<<4 | lo)
	}
	return dst, nil
}

func unhex(c byte) int {
	switch {
	case '0' <= c && c <= '9':
		return int(c - '0')
	case 'a' <= c && c <= 'f':
		return int(c - 'a' + 10)
	case 'A' <= c && c <= 'F':
		return int(c - 'A' + 10)
	default:
		return -1
	}
}
