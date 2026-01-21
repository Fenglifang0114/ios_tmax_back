package svc

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
)

// RSADecrypt RSA解密函数
func RSADecrypt(privateKeyStr, cipherText string, useOaep bool) (string, error) {
	const splitString = "T__S"
	// 1. 解析PEM格式私钥
	privateKey, err := parsePrivateKey(privateKeyStr)
	if err != nil {
		return "", fmt.Errorf("解析私钥失败: %v", err)
	}
	// 2. 分割密文
	var result strings.Builder
	parts := strings.Split(cipherText, splitString)

	// 3. 遍历每个部分解密
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}

		// Base64解码
		cipherBytes, err := base64.StdEncoding.DecodeString(part)
		if err != nil {
			return "", fmt.Errorf("Base64解码失败: %v", err)
		}

		// 解密
		var decryptedBytes []byte
		if useOaep {
			// OAEP SHA256解密
			decryptedBytes, err = rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, cipherBytes, nil)
			if err != nil {
				return "", fmt.Errorf("OAEP解密失败: %v", err)
			}
		} else {
			// PKCS1解密
			decryptedBytes, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherBytes)
			if err != nil {
				return "", fmt.Errorf("PKCS1解密失败: %v", err)
			}
		}

		// 转换为字符串并追加
		result.WriteString(string(decryptedBytes))
	}

	return result.String(), nil
}

// 解析PEM格式私钥
func parsePrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	// 解码PEM块
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("无效的PEM格式")
	}

	// 首先尝试PKCS#8格式（你提供的私钥是PKCS#8格式）
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, fmt.Errorf("不是RSA私钥")
	}

	// 尝试PKCS#1格式
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
