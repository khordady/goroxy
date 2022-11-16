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
		//err = conn.SetReadDeadline(time.Time{})
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Received")
		go handleBrowserToClient(conn)
	}
}

func handleBrowserToClient(browser_to_client net.Conn) {
	buffer := make([]byte, 9*1024)
	length, err := bufio.NewReader(browser_to_client).Read(buffer)
	if err != nil {
		fmt.Println("ERR1 ", err)
		return
	}

	request := processReceived(buffer, length, jjConfig.ListenAuthentication, jjConfig.ListenUsers,
		jjConfig.ListenEncryption, jjConfig.ListenEncryptionKey)
	if request == "" {
		return
	}

	//fmt.Println(request)

	var message []byte

	if jjConfig.SendAuthentication {
		message = []byte(jjConfig.SendUserName + "," + jjConfig.SendPassword + "\r\n")
	}
	message = append(message, []byte(request)...)

	fmt.Println("Message is: " + request)

	switch jjConfig.SendEncryption {
	case "Base64":
		message = encodeBase64(message, len(message))
		break

	case "AES":
		message = encryptAES(message, len(message), jjConfig.ListenEncryptionKey)
		break
	}

	client_to_proxy, e := net.Dial("tcp", jjConfig.Server+":"+jjConfig.ServerPort)
	if e != nil {
		fmt.Println("ERR2 ", e)
		return
	}

	_, e = client_to_proxy.Write(message)
	if e != nil {
		fmt.Println("ERR3 ", e)
		return
	}

	go write(client_to_proxy, browser_to_client)
	read(client_to_proxy, browser_to_client)
}

func write(client_to_proxy net.Conn, browser_to_client net.Conn) {
	defer client_to_proxy.Close()

	reader := bufio.NewReader(browser_to_client)
	writer := bufio.NewWriter(client_to_proxy)

	for {
		_, err := reader.Peek(1)
		if err != nil {
			fmt.Println("ERR6 ", err)
			return
		}
		n := reader.Buffered()
		if n > 0 {
			fmt.Println("Size of Buffered Data: ", n)

			buffer := make([]byte, n)

			length, errr := reader.Read(buffer)
			if errr != nil {
				fmt.Println("ERR6 ", errr)
				return
			}
			if length > 0 {
				fmt.Println(time.Now().Format(time.Stamp) + " READ from browser to client : " + strconv.Itoa(length))
				//fmt.Println(string(buffer[:length]))

				buffer = processToProxyBuffer(buffer, length)
				fmt.Println(time.Now().Format(time.Stamp) + "Decode WRITE from client to proxy: " + strconv.Itoa(len(buffer)))
				//fmt.Println(string(buffer))

				writeLength, errw := writer.Write(buffer)
				if errw != nil {
					fmt.Println("ERR6 ", errw)
					return
				}
				errw = writer.Flush()
				if errw != nil {
					fmt.Println("ERR6 ", errw)
					return
				}
				fmt.Println(time.Now().Format(time.Stamp) + " WRITE from client to proxy: " + strconv.Itoa(writeLength))
			}
		}
	}
}

func read(client_to_proxy net.Conn, browser_to_client net.Conn) {
	defer browser_to_client.Close()

	reader := bufio.NewReader(client_to_proxy)
	writer := bufio.NewWriter(browser_to_client)

	for {
		client_to_proxy.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, err := reader.Peek(1)
		if !os.IsTimeout(err) && err != nil {
			fmt.Println("ERR81 ", err)
			return
		}
		n := reader.Buffered()
		if n > 0 {
			fmt.Println("Size of Buffered Data: ", n)

			buffer := make([]byte, n)

			length, errr := reader.Read(buffer)
			if errr != nil {
				fmt.Println("ERR8 ", errr)
				return
			}
			fmt.Println(time.Now().Format(time.Stamp)+" Encoded READ from proxy to client: ", length)
			//fmt.Println(string(buffer))

			buffer = processToBrowserBuffer(buffer, length)
			fmt.Println(time.Now().Format(time.Stamp)+" Decoded WRITE from client to browser: ", length)
			//fmt.Println(string(buffer))

			write_length, errw := writer.Write(buffer)
			if errw != nil {
				fmt.Println("ERR8 ", errw)
				return
			}

			errw = writer.Flush()
			if errw != nil {
				fmt.Println("ERR82 ", errw)
				return
			}

			fmt.Println(time.Now().Format(time.Stamp)+" WRITE from client to browser: ", write_length)
		}
	}
}
