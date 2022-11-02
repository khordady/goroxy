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
	buffer := make([]byte, 8*1024)
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

	if jjConfig.PrintLog {
		fmt.Println(message)
	}

	splited := strings.Split(message, " ")
	if splited[0] == "CONNECT" {
		proxy_to_server, e := net.Dial("tcp", splited[1])
		if e != nil {
			fmt.Println("ERROR3 ", e)
			return
		}
		_, e = client_to_proxy.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		if e != nil {
			fmt.Println("ERROR4 ", e)
			return
		}

		read443(client_to_proxy, proxy_to_server)
	} else if splited[0] == "GET" {
		host1 := strings.Replace(splited[1], "http://", "", 1)
		host2 := host1[:len(host1)-1]
		var final_host string
		if strings.LastIndexAny(host2, "/") > 0 {
			final_host = host2[:strings.LastIndexAny(host2, "/")]
		} else {
			final_host = host2
		}
		proxy_to_server, e := net.Dial("tcp", final_host+":80")
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
	go read80(client_to_proxy, proxy_to_server)

	buffer := make([]byte, 1024*1024)
	for {
		readLeng, err := proxy_to_server.Read(buffer)
		if err != nil {
			fmt.Println("ERROR8 ", err)
			return
		}
		//fmt.Println("WRIIIIIIIIIIIIIIIIIIIIIIT from server:")
		//fmt.Println(string(buffer[:readLeng]))
		if readLeng > 0 {
			_, err := client_to_proxy.Write(buffer[:readLeng])
			if err != nil {
				fmt.Println("ERR4 ", err)
				return
			}
		}
	}
}

func read80(client_to_proxy net.Conn, proxy_to_server net.Conn) {
	buffer := make([]byte, 1024*1024)

	for {
		readLeng, err := client_to_proxy.Read(buffer)
		if err != nil {
			return
		}
		//fmt.Println("REEEEEEEEEEEEEEEEEEEEEEED from client:")
		//fmt.Println(string(buffer[:readLeng]))
		if readLeng > 0 {
			_, err := proxy_to_server.Write(buffer[:readLeng])
			if err != nil {
				fmt.Println("ERR5 ", err)
				return
			}
		}
	}
}

func write443(client_to_proxy net.Conn, proxy_to_server net.Conn) {
	buffer := make([]byte, 1024*1024)
	for {
		readLeng, err := proxy_to_server.Read(buffer)
		if err != nil {
			fmt.Println("ERROR10 ", err)
			return
		}
		//fmt.Println("WRIIIIIIIIIIIIIIIIIIIIIIT from server:")
		//fmt.Println(string(buffer[:readLeng]))
		if readLeng > 0 {
			_, err := client_to_proxy.Write(buffer[:readLeng])
			if err != nil {
				fmt.Println("ERR11 ", err)
				return
			}
		}
	}
}

func read443(client_to_proxy net.Conn, proxy_to_server net.Conn) {
	go write443(client_to_proxy, proxy_to_server)

	buffer := make([]byte, 1024*1024)
	for {
		readLeng, err := client_to_proxy.Read(buffer)
		if err != nil {
			return
		}
		//fmt.Println("REEEEEEEEEEEEEEEEEEEEEEED from client:")
		//fmt.Println(string(buffer[:readLeng]))
		if readLeng > 0 {
			_, err := proxy_to_server.Write(buffer[:readLeng])
			if err != nil {
				fmt.Println("ERR5 ", err)
				return
			}
		}
	}
}
