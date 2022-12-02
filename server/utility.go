package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"strings"
)

var encrypter cipher.BlockMode
var decrypter cipher.BlockMode

func initializeEncrypter() {
	listen_aesc, err := aes.NewCipher([]byte(jjConfig.ListenEncryptionKey))
	if err != nil {
		fmt.Println(err)
	}
	encrypter = cipher.NewCBCEncrypter(listen_aesc, []byte(jjConfig.ListenEncryptionIV))
	decrypter = cipher.NewCBCDecrypter(listen_aesc, []byte(jjConfig.ListenEncryptionIV))
}

func encryptAES(buffer []byte, length int, key string) []byte {
	finalLength := length + 4
	key_length := len(key)
	plus := (length + 4) % key_length
	if plus > 0 {
		finalLength = length + 4 + (key_length - plus)
	}

	finalBytes := make([]byte, finalLength)
	copyArray(intTobytes(length), finalBytes, 0)
	copyArray(buffer[:length], finalBytes, 4)

	msgByte := make([]byte, finalLength)

	encrypter.CryptBlocks(msgByte, finalBytes)
	return msgByte
}

func decryptAES(buffer []byte, length int) []byte {
	msgByte := make([]byte, length)

	decrypter.CryptBlocks(msgByte, buffer[:length])

	decrypted_buffers := make([]byte, 0)

	leng := bytesToint(msgByte[:4])
	decrypted_buffers = append(decrypted_buffers, msgByte[4:4+leng]...)

	return decrypted_buffers
}

func processReceived(buffer []byte, length int, authentication bool, users []strUser, crypto string, crypto_key string) string {
	switch crypto {
	case "None":
		buffer = buffer[:length]
		break

	case "AES":
		buffer = decryptAES(buffer, length)
	}

	message := string(buffer)
	if !strings.Contains(message, "\r\n") {
		fmt.Println("Wrong UserPass")
		return ""
	}

	message = message[:strings.LastIndex(message, "\r\n\r\n")+4]
	//fmt.Println(message)

	if authentication {
		splited := strings.Split(message, "\r\n")
		splited = strings.Split(splited[0], ",")
		if len(splited) > 1 {
			message = message[strings.Index(message, "\r\n")+2:]

			flag := false
			for _, user := range users {
				if splited[0] == user.ListenUserName || splited[1] == user.ListenPassword {
					flag = true
					break
				}
			}
			if !flag {
				fmt.Println("Wrong UserPass")
				return ""
			}
		}
	}

	return message
}

func intTobytes(size int) []byte {
	bytes := make([]byte, 4)
	bytes[0] = byte(0xff & (size >> 32))
	bytes[1] = byte(0xff & (size >> 16))
	bytes[2] = byte(0xff & (size >> 8))
	bytes[3] = byte(0xff & size)
	return bytes
}

func bytesToint(bytes []byte) int {
	var result int
	result = 0
	for i := 0; i < 4; i++ {
		result = result << 8
		result += int(bytes[i])
	}
	return result
}

func copyArray(src []byte, dst []byte, offset int) {
	for i := 0; i < len(src); i++ {
		(dst)[i+offset] = src[i]
	}
}

func processToHostBuffer(buffer []byte, length int) []byte {
	var newBuffr []byte
	switch jjConfig.ListenEncryption {
	case "None":
		newBuffr = buffer[:length]
		break

	case "AES":
		newBuffr = decryptAES(buffer, length)
		break
	}
	return newBuffr
}

func processToClientBuffer(buffer []byte, length int) []byte {
	var newBuffr []byte
	switch jjConfig.ListenEncryption {
	case "None":
		newBuffr = buffer[:length]
		break

	case "AES":
		newBuffr = encryptAES(buffer, length, jjConfig.ListenEncryptionKey)
		break
	}
	return newBuffr
}
