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
		err = conn.SetDeadline(time.Time{})
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
	length, e := client_to_proxy.Read(buffer)
	if e != nil {
		fmt.Println("ERR1 ", e)
		return
	}

	message := processReceived(buffer, length, jjConfig.ListenAuthentication, jjConfig.ListenUsers,
		jjConfig.ListenEncryption, jjConfig.ListenEncryptionKey)
	if message == "" {
		return
	}

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

		_, e = client_to_proxy.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
		if e != nil {
			fmt.Println("ERROR4 ", e)
			return
		}

		read443(client_to_proxy, proxy_to_server)
	} else {
		proxy_to_server, e := net.Dial("tcp", host[1]+":80")
		if e != nil {
			fmt.Println("ERROR5 ", e)
			return
		}
		_, e = proxy_to_server.Write([]byte(message))
		if e != nil {
			fmt.Println("ERROR6 ", e)
			return
		}

		write80(client_to_proxy, proxy_to_server)
	}
}

func write80(client_to_proxy net.Conn, proxy_to_server net.Conn) {
	defer proxy_to_server.Close()
	go read80(client_to_proxy, proxy_to_server)

	buffer := make([]byte, 1024)
	for {
		readLeng, err := proxy_to_server.Read(buffer)
		fmt.Println("READ from server to proxy80:" + strconv.Itoa(readLeng))
		//fmt.Println(string(buffer[:readLeng]))
		if readLeng > 0 {
			writeLength, err := client_to_proxy.Write(buffer[:readLeng])
			fmt.Println("WRITE from server to proxy80:" + strconv.Itoa(writeLength))
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

func read80(client_to_proxy net.Conn, proxy_to_server net.Conn) {
	defer client_to_proxy.Close()
	buffer := make([]byte, 1024)

	for {
		readLeng, err := client_to_proxy.Read(buffer)
		fmt.Println("READ from proxy to client 80:" + strconv.Itoa(readLeng))
		//fmt.Println(string(buffer[:readLeng]))
		if readLeng > 0 {
			writeLength, err := proxy_to_server.Write(buffer[:readLeng])
			fmt.Println("WRITE from server to proxy80:" + strconv.Itoa(writeLength))
			if err != nil {
				fmt.Println("ERR5 ", err)
				return
			}
		}
		if err != nil {
			return
		}
	}
}

func write443(client_to_proxy net.Conn, proxy_to_server net.Conn) {
	defer proxy_to_server.Close()
	buffer := make([]byte, 1024)
	for {
		readLeng, err := proxy_to_server.Read(buffer)
		fmt.Println("READ from server to proxy443: " + strconv.Itoa(readLeng))
		//fmt.Println(string(buffer[:readLeng]))
		if readLeng > 0 {
			writeLength, err := client_to_proxy.Write(buffer[:readLeng])
			fmt.Println("WRITE from server to proxy443: " + strconv.Itoa(writeLength))
			if err != nil {
				fmt.Println("ERR11 ", err)
				return
			}
		}
		if err != nil {
			fmt.Println("ERROR10 ", err)
			return
		}
	}
}

func read443(client_to_proxy net.Conn, proxy_to_server net.Conn) {
	defer client_to_proxy.Close()
	go write443(client_to_proxy, proxy_to_server)

	buffer := make([]byte, 1024)
	for {
		readLeng, err := client_to_proxy.Read(buffer)
		fmt.Println("from proxy to client443: " + strconv.Itoa(readLeng))
		//fmt.Println(string(buffer[:readLeng]))
		if readLeng > 0 {
			writeLength, err := proxy_to_server.Write(buffer[:readLeng])
			fmt.Println("WRITE from server to proxy443: " + strconv.Itoa(writeLength))
			if err != nil {
				fmt.Println("ERR5 ", err)
				return
			}
		}
		if err != nil {
			return
		}
	}
}
