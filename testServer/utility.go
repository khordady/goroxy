package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

var cc cipher.Block

func encryptAES(buffer []byte, length int, key string) []byte {
	//start := time.Now()
	finalLength := length + 4
	key_length := len(key)
	plus := (length + 4) % key_length
	if plus > 0 {
		finalLength = length + 4 + (key_length - plus)
	}

	finalBytes := make([]byte, finalLength)
	copyArray(intTobytes(length), finalBytes, 0)
	copyArray(buffer[:length], finalBytes, 4)

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
	//elapsed := time.Since(start)
	//fmt.Println("Encryption elapsed ", elapsed.Milliseconds())
	return msgByte
}

func decryptAES(buffer []byte, length int, key string) []byte {
	//start := time.Now()
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

	//elapsed := time.Since(start)
	//fmt.Println("Decryption elapsed ", elapsed.Milliseconds())
	return decrypted_buffers
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
	for i, b := range src {
		(dst)[i+offset] = b
	}
}
