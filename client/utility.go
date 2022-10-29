package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

var cc cipher.Block

func encryptAES(buffer []byte, length int, key string) []byte {
	lenn := length % len(key)
	if lenn > 0 {
		length = length + (length % len(key))
	}
	var err error
	if cc == nil {
		cc, err = aes.NewCipher([]byte(key))
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}
	msgByte := make([]byte, length)

	for i, j := 0, 16; i < length-lenn; i, j = i+16, j+16 {
		cc.Encrypt(msgByte[i:j], buffer[i:j])
	}
	return msgByte
}

func decryptAES(buffer []byte, length int, key string) []byte {
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

	for i, j := 0, 16; i < length; i, j = i+16, j+16 {
		cc.Decrypt(msgByte[i:j], buffer[i:j])
	}
	return msgByte
}

func encodeBase64(buffer []byte) []byte {
	lengt := base64.StdEncoding.EncodedLen(len(buffer))
	b64 := make([]byte, lengt)
	base64.StdEncoding.Encode(b64, buffer)

	fmt.Println("MAX: " + strconv.Itoa(lengt) + " Real: " + strconv.Itoa(len(b64)))
	return b64
}

func decodeBase64(buffer []byte, length int) []byte {
	b64 := make([]byte, base64.StdEncoding.DecodedLen(length))
	dlength, err := base64.StdEncoding.Decode(b64, buffer[:length])
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println("MAX: " + strconv.Itoa(dlength) + " Real: " + strconv.Itoa(len(b64)))
	return b64
}

func processReceived(buffer []byte, length int, authentication bool, username string, password string, crypto string, crypto_key string) string {
	switch crypto {
	case "None":
		buffer = buffer[:length]
		break

	case "Base64":
		buffer = decodeBase64(buffer, length)
		break

	case "AES":
		buffer = decodeBase64(buffer, len(buffer))
		buffer = decryptAES(buffer, len(buffer), crypto_key)
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
			if splited[0] != username || splited[1] != password {
				fmt.Println("Wrong UserPass")
				return ""
			}
		}
	}

	return message
}
