package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type strClientConfig struct {
	PrintLog             bool
	ListenPort           string
	ListenEncryption     string
	ListenEncryptionKey  string
	ListenAuthentication bool
	ListenUsers          []strUser
	Server               string
	ServerPort           string
	SendEncryption       string
	SendEncryptionKey    string
	SendAuthentication   bool
	SendUserName         string
	SendPassword         string
}

type strUser struct {
	ListenUserName string
	ListenPassword string
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

	request := processReceived(buffer, length, jjConfig.ListenAuthentication, jjConfig.ListenUsers,
		jjConfig.ListenEncryption, jjConfig.ListenEncryptionKey)
	if request == "" {
		return
	}

	var message []byte

	if jjConfig.SendAuthentication {
		message = []byte(jjConfig.SendUserName + "," + jjConfig.SendPassword + "\r\n")
	}
	message = append(message, []byte(request)...)

	if jjConfig.PrintLog {
		fmt.Println("Message is: " + request)
	}

	if jjConfig.SendEncryption == "AES" {
		message = encryptAES(buffer, len(message), jjConfig.ListenEncryptionKey)
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

	read(client_to_server, browser_to_client)
}

func write(client_to_proxy net.Conn, browser_to_client net.Conn) {
	defer client_to_proxy.Close()
	buffer := make([]byte, 1024)
	for {
		readLeng, err := browser_to_client.Read(buffer)
		if readLeng > 0 {
			fmt.Println("from client to browser: " + strconv.Itoa(readLeng))
			//fmt.Println(string(buffer[:readLeng]))

			_, err := client_to_proxy.Write(buffer[:readLeng])
			if err != nil {
				fmt.Println("ERR6 ", err)
				return
			}
		}
		if err != nil {
			fmt.Println("ERR5 ", err)
			return
		}
	}
}

func read(client_to_proxy net.Conn, browser_to_client net.Conn) {
	defer browser_to_client.Close()

	buffer, err := bufio.NewReader(client_to_proxy).ReadBytes('\n')

	if len(buffer) > 0 {
		fmt.Println("from proxy to client: " + strconv.Itoa(len(buffer)))
		//fmt.Println(string(buffer))

		_, err := browser_to_client.Write(buffer)
		if err != nil {
			fmt.Println("ERR7 ", err)
			return
		}
		_, err = browser_to_client.Write([]byte("\r\n"))
		if err != nil {
			fmt.Println("ERR7 ", err)
			return
		}
	}
	if err != nil {
		return
	}

	go write(client_to_proxy, browser_to_client)

	buffer = make([]byte, 1024)

	for {
		length, err := bufio.NewReader(client_to_proxy).Read(buffer)
		fmt.Println("from proxy to client: " + strconv.Itoa(length))
		//fmt.Println(string(buffer[:length]))
		if length > 0 {
			_, err := browser_to_client.Write(buffer[:length])
			if err != nil {
				fmt.Println("ERR8 ", err)
				return
			}
		}
		if err != nil {
			return
		}
	}
}
