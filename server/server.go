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

type strServerConfig struct {
	PrintLog             bool
	ListenPort           string
	ListenEncryption     string
	ListenEncryptionKey  string
	ListenAuthentication bool
	ListenUsers          []strUser
}

type strUser struct {
	ListenUserName string
	ListenPassword string
}

var jjConfig strServerConfig

func main() {
	fmt.Println("Reading server-config.json")
	readFile, err := os.Open("server-config.json")

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

	fmt.Println("Start server...")

	ln, _ := net.Listen("tcp", ":"+jjConfig.ListenPort)

	for {
		conn, _ := ln.Accept()
		//err = conn.SetDeadline(time.Time{})
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Connection Received")
		go handleSocket(conn)
	}
}

func handleSocket(client_to_proxy net.Conn) {
	buffer := make([]byte, 9*1024)
	length, err := bufio.NewReader(client_to_proxy).Read(buffer)
	if err != nil {
		fmt.Println("ERR1 ", err)
		return
	}

	message := processReceived(buffer, length, jjConfig.ListenAuthentication, jjConfig.ListenUsers,
		jjConfig.ListenEncryption, jjConfig.ListenEncryptionKey)
	if message == "" {
		return
	}

	fmt.Println("MESSAGE IS: " + message)

	var host []string
	headers := strings.Split(message, "\r\n")
	for _, header := range headers {
		if strings.HasPrefix(header, "Host") {
			host = strings.Split(header, " ")
			fmt.Println("HOST ISSSSSSSS:" + host[1])
			break
		}
	}

	if strings.HasSuffix(host[1], "443") {
		proxy_to_server, e := net.Dial("tcp", host[1])
		if e != nil {
			fmt.Println("ERROR3 ", e)
			return
		}

		fmt.Println("CONNECTED TO: " + host[1])

		//length, err := client_to_proxy.Write([]byte("TEST MESSAGE FROM GITHUB"))
		if err != nil {
			return
		}
		//writer := bufio.NewWriter(client_to_proxy)
		bytess := []byte("HTTP/1.1 200 Connection Established\r\n\r\n")
		switch jjConfig.ListenEncryption {
		case "Base64":
			bytess = encodeBase64(bytess, len(bytess))
			break

		case "AES":
			bytess = encryptAES(bytess, len(bytess), jjConfig.ListenEncryptionKey)
			break
		}
		Writelength, err := client_to_proxy.Write(bytess)
		fmt.Println("WROTED 200: ", Writelength)

		//_, e = writer.Write([]byte("TEST MESSAGE FROM GITHUB"))
		//_, e = writer.Write([]byte("HTTP/1.1 200 Connection Established\r\n"))
		//_, e = writer.Write([]byte("\r\n"))
		//err := writer.Flush()
		if err != nil {
			fmt.Println("ERROR42 ", err)
			return
		}

		go read(client_to_proxy, proxy_to_server)
		write(client_to_proxy, proxy_to_server)

	} else {
		proxy_to_server, e := net.Dial("tcp", host[1]+":80")
		if e != nil {
			fmt.Println("ERROR5 ", e)
			return
		}
		Writelength, e := proxy_to_server.Write([]byte(message))
		fmt.Println("WROTE 80 Header: " + strconv.Itoa(Writelength))
		if e != nil {
			fmt.Println("ERROR6 ", e)
			return
		}

		go read(client_to_proxy, proxy_to_server)
		write(client_to_proxy, proxy_to_server)
	}
}

func write(client_to_proxy net.Conn, proxy_to_host net.Conn) {
	defer proxy_to_host.Close()

	reader := bufio.NewReader(proxy_to_host)
	writer := bufio.NewWriter(client_to_proxy)

	for {
		_, err := reader.Peek(1)
		if err != nil {
			fmt.Println("ERROR8 ", err)
			return
		}
		n := reader.Buffered()
		if n > 0 {
			buffer := make([]byte, n)

			length, err := reader.Read(buffer)
			fmt.Println(time.Now().Format(time.Stamp) + " READ from server to proxy:" + strconv.Itoa(length))
			fmt.Println(string(buffer[:length]))
			if length > 0 {
				buffer = processToClientBuffer(buffer, length)
				fmt.Println(time.Now().Format(time.Stamp) + " Encoded Base64 WRITE from proxy to client:" + strconv.Itoa(len(buffer)))
				fmt.Println(string(buffer))
				writeLength, err := writer.Write(buffer)
				writer.Flush()
				fmt.Println(time.Now().Format(time.Stamp) + " WRITE from proxy to client:" + strconv.Itoa(writeLength))
				if err != nil {
					fmt.Println("ERR4 ", err)
					return
				}
			}
			if err != nil {
				fmt.Println("ERROR8 ", err)
				return
			}
		}
	}
}

func read(client_to_proxy net.Conn, proxy_to_host net.Conn) {
	defer client_to_proxy.Close()

	reader := bufio.NewReader(client_to_proxy)
	writer := bufio.NewWriter(proxy_to_host)
	for {
		client_to_proxy.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, err := reader.Peek(1)
		if !os.IsTimeout(err) && err != nil {
			fmt.Println("ERR51 ", err)
			return
		}
		n := reader.Buffered()
		if n > 0 {
			buffer := make([]byte, n)
			length, err := reader.Read(buffer)
			fmt.Println(time.Now().Format(time.Stamp) + "Encoded READ from client to proxy:" + strconv.Itoa(length))
			fmt.Println(string(buffer[:length]))
			if length > 0 {
				buffer = processToHostBuffer(buffer, length)
				fmt.Println(time.Now().Format(time.Stamp) + "Decoded WRITE from proxy to server :" + strconv.Itoa(len(buffer)))
				fmt.Println(string(buffer))
				writeLength, err := writer.Write(buffer)
				writer.Flush()
				fmt.Println(time.Now().Format(time.Stamp) + " WRITE from proxy to server :" + strconv.Itoa(writeLength))
				if err != nil {
					fmt.Println("ERR5 ", err)
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
