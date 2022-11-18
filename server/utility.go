package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"strings"
)

var cc cipher.Block

func encryptAES(buffer []byte, length int, key string) []byte {
	finalLength := 0
	key_length := len(key)
	plus := (length + 4) % key_length
	if plus > 0 {
		finalLength = length + 4 + (key_length - plus)
		plusBuffer := make([]byte, key_length-plus)
		buffer = append(buffer, plusBuffer...)
	}

	finalBytes := make([]byte, finalLength)
	copyArray(intTobytes(length), finalBytes, 0)
	copyArray(buffer, finalBytes, 4)

	var err error
	if cc == nil {
		cc, err = aes.NewCipher([]byte(key))
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}
	msgByte := make([]byte, finalLength)

	for i, j := 0, key_length; i < finalLength; i, j = i+key_length, j+key_length {
		cc.Encrypt(msgByte[i:j], finalBytes[i:j])
	}
	return msgByte
}

func decryptAES(buffer []byte, length int, key string) []byte {
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

	decrypted_buffers := make([]byte, 0)

	leng := bytesToint(msgByte[:4])
	decrypted_buffers = append(decrypted_buffers, msgByte[4:4+leng]...)

	return decrypted_buffers
}

//func encodeBase64(buffer []byte, length int) []byte {
//	lengt := base64.StdEncoding.EncodedLen(length)
//	b64 := make([]byte, lengt)
//	base64.StdEncoding.Encode(b64, buffer[:length])
//
//	return b64
//}
//
//func decodeBase64(buffer []byte, length int) []byte {
//	b64 := make([]byte, base64.StdEncoding.DecodedLen(length))
//	n, err := base64.StdEncoding.Decode(b64, buffer[:length])
//	if err != nil {
//		fmt.Println(err)
//		return nil
//	}
//	return b64[:n]
//}

func processReceived(buffer []byte, length int, authentication bool, users []strUser, crypto string, crypto_key string) string {
	switch crypto {
	case "None":
		buffer = buffer[:length]
		break

	//case "Base64":
	//	buffer = decodeBase64(buffer, length)
	//	break

	case "AES":
		buffer = decryptAES(buffer, length, crypto_key)
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
	bytes[0] = byte(0xff & size)
	bytes[1] = byte(0xff & (size >> 8))
	bytes[2] = byte(0xff & (size >> 16))
	bytes[3] = byte(0xff & (size >> 32))

	return bytes
}

func bytesToint(bytes []byte) int {
	var result int
	result = 0
	for i := 1; i >= 0; i-- {
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

func processToHostBuffer(buffer []byte, length int) []byte {
	switch jjConfig.ListenEncryption {
	case "None":
		break

	//case "Base64":
	//	buffer = decodeBase64(buffer, length)
	//	break

	case "AES":
		buffer = decryptAES(buffer, length, jjConfig.ListenEncryptionKey)
		break
	}
	return buffer
}

func processToClientBuffer(buffer []byte, length int) []byte {
	switch jjConfig.ListenEncryption {
	//case "Base64":
	//	buffer = encodeBase64(buffer, length)
	//	break

	case "AES":
		buffer = encryptAES(buffer, length, jjConfig.ListenEncryptionKey)
		break
	}
	return buffer
}
