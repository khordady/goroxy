package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

	if jjConfig.PrintLog {
		fmt.Println("Message is: " + request)
	}

	if jjConfig.SendEncryption == "AES" {
		message = encryptAES(buffer, len(message), jjConfig.ListenEncryptionKey)
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

	//go exchange(client_to_server, browser_to_client)
	//exchange(browser_to_client, client_to_server)
}

func exchange(src, dest net.Conn) {
	fmt.Println("Start SRC: ", src.RemoteAddr())
	defer func(src net.Conn) {
		err := src.Close()
		if err != nil {
			fmt.Println("ERRCPP1 ", err)
		}
	}(src)
	defer func(dest net.Conn) {
		err := dest.Close()
		if err != nil {
			fmt.Println("ERRCPP2 ", err)
		}
	}(dest)
	written, err := io.Copy(src, dest)
	if err != nil {
		fmt.Println("COPY ERROR CLIENT IS: ", written, err)
		return
	}
}

func write(client_to_proxy net.Conn, browser_to_client net.Conn) {
	defer client_to_proxy.Close()

	reader := bufio.NewReader(browser_to_client)
	//writer := bufio.NewWriter(client_to_proxy)

	for {
		reader.Peek(1)
		n := reader.Buffered()
		if n > 0 {
			fmt.Println("Size of Buffered Data: ", n)

			buffer := make([]byte, n)

			length, err := reader.Read(buffer)
			if length > 0 {
				fmt.Println(time.Now().Format(time.Stamp) + " READ from client to browser: " + strconv.Itoa(length))
				//fmt.Println(string(buffer[:readLeng]))

				writeLength, err := client_to_proxy.Write(buffer[:length])
				if writeLength > 0 {
					fmt.Println(time.Now().Format(time.Stamp) + " WRITE from client to browser: " + strconv.Itoa(writeLength))
				}
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
}

func read(client_to_proxy net.Conn, browser_to_client net.Conn) {
	defer browser_to_client.Close()

	reader := bufio.NewReader(client_to_proxy)
	writer := bufio.NewWriter(browser_to_client)

	for {
		reader.Peek(1)
		n := reader.Buffered()
		if n > 0 {
			fmt.Println("Size of Buffered Data: ", n)

			buffer := make([]byte, n)

			length, err := reader.Read(buffer)
			fmt.Println(time.Now().Format(time.Stamp)+" READ from proxy to client: ", length)
			//fmt.Println(string(buffer[:length]))
			//if length > 0 {
			write_length, err := writer.Write(buffer[:length])
			if err != nil {
				fmt.Println("ERR8 ", err)
				return
			}
			fmt.Println(time.Now().Format(time.Stamp)+" WRITE from client to browser: ", write_length)
			if err != nil {
				fmt.Println("ERR8 ", err)
				return
			}
			//}
			if err != nil {
				fmt.Println("ERR81 ", err)
				return
			}
		}
	}
}
