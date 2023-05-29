package lic

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"reflect"
	"time"

	"github.com/denisbrodbeck/machineid"

	"tmaxsrv/log"
)

func IsKeyValid(licenseKey string) (bool, string, string) {
	// Get a unique machine ID based on the CPUID and Hard Disk ID
	machineIDStr, _ := machineid.ProtectedID("")

	//salt := []byte("Tscale's key 7387231296834261945") // 32 bytes
	myCipherSalt := "KJ58UTRbqrlBYjpqSY3mCjXlQU8xa743Zz6024DxZtmoIUQIffCOOfE5w6RAB7X3"
	key := []byte("MYY256Key-32Characters8376291290") // The key must be 16, 24, or 32 bytes long

	salt, _ := decrypt(key, myCipherSalt)
	machineId := []byte(machineIDStr[0:10]) // e4e13e78c5
	log.Log.Debugf("MachineId:%s\n", machineId)
	saltedData := append([]byte(machineIDStr[0:10]), []byte(salt)...)
	hash := md5.Sum(saltedData)
	hashStr := hex.EncodeToString(hash[:])
	if !reflect.DeepEqual(licenseKey[0:32], hashStr) {
		return false, machineIDStr[0:10], ""
	}
	saltedDatav := append([]byte(licenseKey[32:42]), []byte(hashStr)...) // valid date
	hashv := md5.Sum(saltedDatav)
	hashStrv := hex.EncodeToString(hashv[:])
	if !reflect.DeepEqual(licenseKey[42:74], hashStrv) {
		return false, machineIDStr[0:10], ""
	}
	layout := "2006-01-02"
	date, err := time.Parse(layout, licenseKey[32:42])
	if err != nil || time.Now().After(date) {
		fmt.Println(err)
		return false, machineIDStr[0:10], licenseKey[32:42]
	}
	return true, machineIDStr[0:10], ""
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
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
