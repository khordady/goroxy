package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type strClientConfig struct {
	Server              string
	ClientPort          string
	ServerPort          string
	SendEncryption      string
	SendEncryptionKey   string
	ListenEncryption    string
	ListenEncryptionKey string
	Authentication      bool
	UserName            string
	Password            string
}

var jjClientConfig strClientConfig

func main() {
	fmt.Println("Reading client-config.json")
	readFile, err := os.Open("client-config.json")

	if err != nil {
		fmt.Println(err)
		return
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileLines strings.Builder

	for fileScanner.Scan() {
		fileLines.WriteString(fileScanner.Text())
	}

	final := strings.ReplaceAll(fileLines.String(), " ", "")
	final = strings.ReplaceAll(final, "\n", "")

	readFile.Close()

	err = json.NewDecoder(bytes.NewReader([]byte(final))).Decode(&jjClientConfig)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("GOROXY: Start Client Listening ...")

	ln, _ := net.Listen("tcp", ":"+jjClientConfig.ClientPort)

	for {
		conn, _ := ln.Accept()
		err = conn.SetDeadline(time.Time{})
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleBrowserTOCLient(conn)
	}
}

func handleBrowserTOCLient(browser_to_client net.Conn) {
	buffer := make([]byte, 8*1024)
	length, e := browser_to_client.Read(buffer)
	if e != nil {
		fmt.Println("ERROR1 ", e)
		return
	}

	var message []byte

	if jjClientConfig.Authentication {
		message = append([]byte(jjClientConfig.UserName+","+jjClientConfig.Password+"\r\n"), buffer[:length]...)
	}
	fmt.Println(string(message))

	switch jjClientConfig.ListenEncryption {
	case "None":
		buffer = buffer[:length]
		break

	case "Base64":
		buffer = encodeBase64(buffer, length)
		break

	case "AES":
		buffer = encryptAES(buffer, length, jjClientConfig.ListenEncryptionKey)
		break
	}

	client_to_server, e := net.Dial("tcp", jjClientConfig.Server+":"+jjClientConfig.ServerPort)
	if e != nil {
		fmt.Println("ERROR2 ", e)
		return
	}

	_, e = client_to_server.Write(message)
	if e != nil {
		fmt.Println("ERROR8 ", e)
		return
	}

	read(client_to_server, browser_to_client)
}

func write(client_to_server net.Conn, browser_to_client net.Conn) {
	buffer := make([]byte, 32*1024)
	for {
		readLeng, err := browser_to_client.Read(buffer)
		if err != nil {
			fmt.Println("ERROR10 ", err)
			return
		}
		if readLeng > 0 {
			_, err := client_to_server.Write(buffer[:readLeng])
			if err != nil {
				fmt.Println("ERR4 ", err)
				return
			}
		}
	}
}

func read(client_to_server net.Conn, browser_to_client net.Conn) {
	buffer := make([]byte, 32*1024)

	readLeng, err := client_to_server.Read(buffer)
	if err != nil {
		return
	}
	if readLeng > 0 {
		_, err := browser_to_client.Write(buffer[:readLeng])
		if err != nil {
			fmt.Println("ERR5 ", err)
			return
		}
	}

	go write(client_to_server, browser_to_client)

	for {
		readLeng, err := client_to_server.Read(buffer)
		if err != nil {
			return
		}
		//fmt.Println("REEEEEEEEEEEEEEEEEEEEEEED from client:")
		//fmt.Println(string(buffer[:readLeng]))
		if readLeng > 0 {
			_, err := browser_to_client.Write(buffer[:readLeng])
			if err != nil {
				fmt.Println("ERR5 ", err)
				return
			}
		}
	}
}
