package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

var cc cipher.Block

func encryptAES(buffer []byte, length int, key string) []byte {
	var err error
	if cc == nil {
		cc, err = aes.NewCipher([]byte(key))
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}
	msgByte := make([]byte, length)
	cc.Decrypt(msgByte, buffer[:length])

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
	cc.Decrypt(msgByte, buffer[:length])

	return msgByte
}

func encodeBase64(buffer []byte, length int) []byte {
	lengt := base64.StdEncoding.EncodedLen(len(buffer[:length]))
	b64 := make([]byte, lengt)
	base64.StdEncoding.Encode(b64, buffer[:length])

	return b64[:lengt]
}

func decodeBase64(buffer []byte, length int) []byte {
	b64 := make([]byte, base64.StdEncoding.DecodedLen(len(buffer[:length])))
	dlength, err := base64.StdEncoding.Decode(b64, buffer[:length])
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return b64[:dlength]
}
