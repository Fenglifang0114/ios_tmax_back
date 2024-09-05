package lic

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"

	"tmaxsrv/log"
)

// 检验Key是否认证通过
func IsKeyValid(licenseKey string) (bool, string, string, string) {
	// Get a unique machine ID based on the CPUID and Hard Disk ID
	machineIDStr, _ := machineid.ProtectedID("")

	// salt := []byte("Tscale's key 7387231296834261945") // 32 bytes
	myCipherSalt := "KJ58UTRbqrlBYjpqSY3mCjXlQU8xa743Zz6024DxZtmoIUQIffCOOfE5w6RAB7X3"
	key := []byte("MYY256Key-32Characters8376291290") // The key must be 16, 24, or 32 bytes long

	if len(licenseKey) == 74 {
		salt, _ := decrypt(key, myCipherSalt)
		machineId := []byte(machineIDStr[0:10]) // e4e13e78c5
		log.Log.Debugf("MachineId:%s\n", machineId)
		saltedData := append([]byte(machineIDStr[0:10]), []byte(salt)...)
		hash := md5.Sum(saltedData)
		hashStr := hex.EncodeToString(hash[:])
		if !reflect.DeepEqual(licenseKey[0:32], hashStr) {
			return false, machineIDStr[0:10], "", ""
		}
		saltedDatav := append([]byte(licenseKey[32:42]), []byte(hashStr)...) // valid date
		hashv := md5.Sum(saltedDatav)
		hashStrv := hex.EncodeToString(hashv[:])
		if !reflect.DeepEqual(licenseKey[42:74], hashStrv) {
			return false, machineIDStr[0:10], "", ""
		}
		layout := "2006-01-02"
		date, err := time.Parse(layout, licenseKey[32:42])
		if err != nil || time.Now().After(date) {
			fmt.Println(err)
			return false, machineIDStr[0:10], licenseKey[32:42], ""
		}
		return true, machineIDStr[0:10], licenseKey[32:42], "T-Config"

	} else {
		salt, _ := decrypt(key, myCipherSalt)
		machineId := []byte(machineIDStr[0:10]) // e4e13e78c5
		log.Log.Debugf("MachineId:%s\n", machineId)
		saltedData := append([]byte(machineIDStr[0:10]), []byte(salt)...)
		saltedData = append([]byte(licenseKey[32:36]), saltedData...)
		// saltedData = append([]byte(licenseKey[32:36]), []byte(salt)...)
		hash := md5.Sum(saltedData)
		hashStr := hex.EncodeToString(hash[:])
		if !reflect.DeepEqual(licenseKey[0:32], hashStr) {
			return false, machineIDStr[0:10], "", ""
		}
		saltedDatav := append([]byte(licenseKey[36:46]), []byte(hashStr)...) // valid date
		hashv := md5.Sum(saltedDatav)
		hashStrv := hex.EncodeToString(hashv[:])
		if !reflect.DeepEqual(licenseKey[46:78], hashStrv) {
			return false, machineIDStr[0:10], "", ""
		}

		layout := "2006-01-02"
		date, err := time.Parse(layout, licenseKey[36:46])
		if err != nil || time.Now().After(date) {
			fmt.Println(err)
			return false, machineIDStr[0:10], licenseKey[36:46], ""
		}
		return true, machineIDStr[0:10], licenseKey[36:46], licenseKey[32:36]

	}

}

func decrypt(key []byte, ciphertext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertextBytes, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	iv := ciphertextBytes[:aes.BlockSize]
	ciphertextBytes = ciphertextBytes[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertextBytes, ciphertextBytes)

	return string(ciphertextBytes), nil
}

func ReadLicFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// func SaveKey(filePath string, content string) error {
// 	return os.WriteFile(filePath, []byte(content+"\n"), 0644)
// }

func ensureFileExists(filePath string) {
	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("creat file fail: %v\n", err)
			return
		}
		defer file.Close()
	}
}

func SaveKey(filePath string, content string) error {

	ensureFileExists(filePath)

	contentStr, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	var listContent []string
	result := []string{}
	if len(contentStr) > 0 {
		listContent = strings.Split(string(contentStr), "\r\n")
		if len(content) == 74 {
			for _, item := range listContent {
				if len(item) != 74 && len(item) == 78 {
					result = append(result, item)
				}

			}

		} else if len(content) == 78 {
			for _, item := range listContent {
				if len(item) == 78 && item[32:36] != content[32:36] {
					result = append(result, item)
				} else if len(item) == 74 {
					result = append(result, item)
				}

			}

		}
	}

	result = append(result, content)
	var lastStr string
	for _, line := range result {
		lastStr = lastStr + line + "\r\n"
	}

	return os.WriteFile(filePath, []byte(lastStr), 0644)

}
