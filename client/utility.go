package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"strings"
)

var cc cipher.Block

func encryptAES(buffer []byte, length int, key string) []byte {
	key_length := len(key)
	plus := (length + 2) % key_length
	if plus > 0 {
		length = length + (key_length - plus)
	}

	finalBytes := make([]byte, length)
	copyArray(intTobytes(len(buffer)), finalBytes, 4)
	copyArray(buffer, finalBytes, 4)

	var err error
	if cc == nil {
		cc, err = aes.NewCipher([]byte(key))
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}
	msgByte := make([]byte, length)

	for i, j := 0, key_length; i < length-plus; i, j = i+key_length, j+key_length {
		cc.Encrypt(msgByte[i:j], buffer[i:j])
	}
	return msgByte
}

func decryptAES(buffer []byte, length int, key string) []byte {
	decrypted_buffers := make([]byte, 0)
	key_length := len(key)
	var err error
	if cc == nil {
		cc, err = aes.NewCipher([]byte(key))
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}
	if err != nil {
		fmt.Println(err)
		return nil
	}
	msgByte := make([]byte, length)

	for i, j := 0, key_length; i < length; i, j = i+key_length, j+key_length {
		cc.Decrypt(msgByte[i:j], buffer[i:j])
	}

	for index := 0; index < len(msgByte); {
		length = bytesToint(msgByte[index : index+4])
		decrypted_buffers = append(decrypted_buffers, msgByte[index+4:index+4+length]...)
		index = index + 4 + length
	}

	return decrypted_buffers
}

func processReceived(buffer []byte, length int, authentication bool, users []strUser,
	crypto string, crypto_key string) string {
	switch crypto {
	case "None":
		buffer = buffer[:length]
		break

	case "AES":
		buffer = decryptAES(buffer, length, crypto_key)
	}

	message := string(buffer)

	if !strings.Contains(message, "\r\n") {
		fmt.Println("Wrong UserPass")
		return ""
	}

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
	bytes[0] = byte(0xff & size)
	bytes[1] = byte(0xff & (size >> 8))
	bytes[1] = byte(0xff & (size >> 16))
	bytes[1] = byte(0xff & (size >> 32))

	return bytes
}

func bytesToint(bytes []byte) int {
	var result int
	result = 0
	for i := 3; i >= 0; i-- {
		result = result << 8
		result += int(bytes[i])
	}
	return result
}

func copyArray(src []byte, dst []byte, offset int) {
	for i, b := range src {
		(dst)[i+offset] = b
	}
}

func processToProxyBuffer(buffer []byte, length int) []byte {
	if jjConfig.ListenEncryption == jjConfig.SendEncryption {
		if (jjConfig.SendEncryptionKey == jjConfig.ListenEncryptionKey && jjConfig.SendEncryption == "AES") ||
			jjConfig.ListenEncryption == "None" {
			return buffer
		}
	}

	switch jjConfig.ListenEncryption {
	case "AES":
		buffer = decryptAES(buffer, length, jjConfig.ListenEncryptionKey)
		break
	}

	switch jjConfig.SendEncryption {
	case "AES":
		buffer = encryptAES(buffer, len(buffer), jjConfig.SendEncryptionKey)
		break
	}
	return buffer
}

func processToBrowserBuffer(buffer []byte, length int) []byte {
	if jjConfig.ListenEncryption == jjConfig.SendEncryption {
		if (jjConfig.SendEncryptionKey == jjConfig.ListenEncryptionKey && jjConfig.SendEncryption == "AES") ||
			jjConfig.ListenEncryption == "None" {
			return buffer
		}
	}

	switch jjConfig.SendEncryption {
	case "AES":
		buffer = decryptAES(buffer, length, jjConfig.ListenEncryptionKey)
		break
	}

	switch jjConfig.ListenEncryption {
	case "AES":
		buffer = encryptAES(buffer, len(buffer), jjConfig.SendEncryptionKey)
		break
	}
	return buffer
}
