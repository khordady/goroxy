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
	ListenPort           string
	ListenEncryption     string
	ListenEncryptionKey  string
	ListenAuthentication bool
	ListenUserName       string
	ListenPassword       string
	Server               string
	ServerPort           string
	SendEncryption       string
	SendEncryptionKey    string
	SendAuthentication   bool
	SendUserName         string
	SendPassword         string
}

var jjConfig strClientConfig

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

	final := strings.ReplaceAll(fileLines.String(), "\n", "")

	readFile.Close()

	err = json.NewDecoder(bytes.NewReader([]byte(final))).Decode(&jjConfig)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("GOROXY: Start Client Listening ...")

	ln, _ := net.Listen("tcp", ":"+jjConfig.ListenPort)

	for {
		conn, _ := ln.Accept()
		err = conn.SetDeadline(time.Time{})
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Received")
		go handleBrowserToClient(conn)
	}
}

func handleBrowserToClient(browser_to_client net.Conn) {
	buffer := make([]byte, 8*1024)
	length, e := browser_to_client.Read(buffer)
	if e != nil {
		fmt.Println("ERR1 ", e)
		return
	}

	request := processReceived(buffer, length, jjConfig.ListenAuthentication, jjConfig.ListenUserName, jjConfig.ListenPassword,
		jjConfig.ListenEncryption, jjConfig.ListenEncryptionKey)
	if request == "" {
		return
	}

	var message []byte

	if jjConfig.SendAuthentication {
		message = []byte(jjConfig.SendUserName + "," + jjConfig.SendPassword + "\r\n")
	}
	message = append(message, []byte(request)...)

	switch jjConfig.SendEncryption {
	case "Base64":
		message = encodeBase64(message)
		break

	case "AES":
		message = encryptAES(buffer, len(message), jjConfig.ListenEncryptionKey)
		break
	}

	client_to_server, e := net.Dial("tcp", jjConfig.Server+":"+jjConfig.ServerPort)
	if e != nil {
		fmt.Println("ERR2 ", e)
		return
	}

	_, e = client_to_server.Write(message)
	if e != nil {
		fmt.Println("ERR3 ", e)
		return
	}
	//send these to ensure end of packet at server side
	_, e = client_to_server.Write([]byte("\r\n"))
	if e != nil {
		fmt.Println("ERR3 ", e)
		return
	}

	read(client_to_server, browser_to_client)
}

func write(client_to_server net.Conn, browser_to_client net.Conn) {
	buffer := make([]byte, 32*1024)
	for {
		readLeng, err := browser_to_client.Read(buffer)
		if err != nil {
			fmt.Println("ERR5 ", err)
			return
		}
		if readLeng > 0 {
			//fmt.Println("WRRRRRRRRRRRRRRRRRRRRRRRRIIIIIIIT from client:")
			//fmt.Println(string(buffer[:readLeng]))

			_, err := client_to_server.Write(buffer[:readLeng])
			if err != nil {
				fmt.Println("ERR6 ", err)
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
		//fmt.Println("REEEEEEEEEEEEEEEEEEEEEEED from client:")
		//fmt.Println(string(buffer[:readLeng]))

		_, err := browser_to_client.Write(buffer[:readLeng])
		if err != nil {
			fmt.Println("ERR7 ", err)
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
				fmt.Println("ERR8 ", err)
				return
			}
		}
	}
}
