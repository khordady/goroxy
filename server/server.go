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

		//writer := bufio.NewWriter(client_to_proxy)
		Writelength, err := client_to_proxy.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
		if err != nil {
			return
		}
		//Writelength, e := writer.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
		//writer.Flush()
		fmt.Println("WROTE 200 OK: " + strconv.Itoa(Writelength))
		if e != nil {
			fmt.Println("ERROR4 ", e)
			return
		}

		//read443(client_to_proxy, proxy_to_server)
		go exchange(client_to_proxy, proxy_to_server)
		exchange(proxy_to_server, client_to_proxy)

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

		//write80(client_to_proxy, proxy_to_server)
		go exchange(proxy_to_server, client_to_proxy)
		exchange(client_to_proxy, proxy_to_server)
	}
}

func exchange(src, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	io.Copy(src, dest)
}

//func write80(client_to_proxy net.Conn, proxy_to_server net.Conn) {
//	defer proxy_to_server.Close()
//	go read80(client_to_proxy, proxy_to_server)
//
//	buffer := make([]byte, 8*1024)
//	reader := bufio.NewReader(proxy_to_server)
//	writer := bufio.NewWriter(client_to_proxy)
//	for {
//		length, err := reader.Read(buffer)
//		fmt.Println(time.Now().Format(time.Stamp) + " READ from server to proxy80:" + strconv.Itoa(length))
//		//fmt.Println(string(buffer[:readLeng]))
//		if length > 0 {
//			writeLength, err := writer.Write(buffer[:length])
//			writer.Flush()
//			fmt.Println(time.Now().Format(time.Stamp) + " WRITE from server to proxy80:" + strconv.Itoa(writeLength))
//			if err != nil {
//				fmt.Println("ERR4 ", err)
//				return
//			}
//		}
//		if err != nil {
//			fmt.Println("ERROR8 ", err)
//			return
//		}
//	}
//}
//
//func read80(client_to_proxy net.Conn, proxy_to_server net.Conn) {
//	defer client_to_proxy.Close()
//	buffer := make([]byte, 8*1024)
//	reader := bufio.NewReader(client_to_proxy)
//	writer := bufio.NewWriter(proxy_to_server)
//	for {
//		length, err := reader.Read(buffer)
//		fmt.Println(time.Now().Format(time.Stamp) + " READ from proxy to client 80:" + strconv.Itoa(length))
//		//fmt.Println(string(buffer[:readLeng]))
//		if length > 0 {
//			writeLength, err := writer.Write(buffer[:length])
//			writer.Flush()
//			fmt.Println(time.Now().Format(time.Stamp) + " WRITE from server to proxy80:" + strconv.Itoa(writeLength))
//			if err != nil {
//				fmt.Println("ERR5 ", err)
//				return
//			}
//		}
//		if err != nil {
//			return
//		}
//	}
//}
//
//func write443(client_to_proxy net.Conn, proxy_to_server net.Conn) {
//	defer proxy_to_server.Close()
//	buffer := make([]byte, 8*1024)
//	reader := bufio.NewReader(proxy_to_server)
//	writer := bufio.NewWriter(client_to_proxy)
//
//	for {
//		length, err := reader.Read(buffer)
//		fmt.Println(time.Now().Format(time.Stamp) + " READ from server to proxy443: " + strconv.Itoa(length))
//		//fmt.Println(string(buffer[:readLeng]))
//		if length > 0 {
//			writeLength, err := writer.Write(buffer[:length])
//			writer.Flush()
//			fmt.Println(time.Now().Format(time.Stamp) + " WRITE from proxy to client443: " + strconv.Itoa(writeLength))
//			if err != nil {
//				fmt.Println("ERR11 ", err)
//				return
//			}
//		}
//		if err != nil {
//			fmt.Println("ERROR10 ", err)
//			return
//		}
//	}
//}
//
//func read443(client_to_proxy net.Conn, proxy_to_server net.Conn) {
//	defer client_to_proxy.Close()
//	go write443(client_to_proxy, proxy_to_server)
//
//	buffer := make([]byte, 8*1024)
//	reader := bufio.NewReader(client_to_proxy)
//	writer := bufio.NewWriter(proxy_to_server)
//	for {
//		length, err := reader.Read(buffer)
//		fmt.Println(time.Now().Format(time.Stamp) + " READ from client to proxy443: " + strconv.Itoa(length))
//		//fmt.Println(string(buffer[:readLeng]))
//		if length > 0 {
//			writeLength, err := writer.Write(buffer[:length])
//			writer.Flush()
//			fmt.Println(time.Now().Format(time.Stamp) + " WRITE from proxy to server443: " + strconv.Itoa(writeLength))
//			if err != nil {
//				fmt.Println("ERR5 ", err)
//				return
//			}
//		}
//		if err != nil {
//			return
//		}
//	}
//}
