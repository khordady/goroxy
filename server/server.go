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
	ListenPort          string
	ListenEncryption    string
	ListenEncryptionKey string
	Authentication      bool
	UserName            string
	Password            string
}

var jjServerConfig strServerConfig

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

	final := strings.ReplaceAll(fileLines.String(), " ", "")
	final = strings.ReplaceAll(final, "\n", "")

	readFile.Close()

	err = json.NewDecoder(bytes.NewReader([]byte(final))).Decode(&jjServerConfig)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Start server...")

	ln, _ := net.Listen("tcp", ":"+jjServerConfig.ListenPort)

	for {
		conn, _ := ln.Accept()
		err = conn.SetDeadline(time.Time{})
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleSocket(conn)
	}
}

func handleSocket(client_to_proxy net.Conn) {
	buffer := make([]byte, 8*1024)
	length, e := client_to_proxy.Read(buffer)
	if e != nil {
		fmt.Println("ERROR1 ", e)
		return
	}
	if e != nil {
		fmt.Println("ERROR1 ", e)
		return
	}

	var message string
	switch jjServerConfig.ListenEncryption {
	case "None":
		message = string(buffer[:length])
		break

	case "Base64":
		message = string(decodeBase64(buffer, length))
		break

	case "AES":
		message = string(decryptAES(buffer, length, jjServerConfig.ListenEncryptionKey))
		break
	}

	if !strings.Contains(message, "\r\n") {
		fmt.Println("Wrong UserPass")
		return
	}

	if jjServerConfig.Authentication {
		splited := strings.Split(message, "\r\n")
		splited = strings.Split(splited[0], ",")
		if len(splited) > 1 {
			if splited[0] != jjServerConfig.UserName || splited[1] != jjServerConfig.Password {
				fmt.Println("Wrong UserPass")
				return
			}
		}
	}

	splited := strings.Split(message, " ")
	if splited[0] == "CONNECT" {
		//message = strings.Replace(message, "CONNECT", "GET", 1)
		proxy_to_server, e := net.Dial("tcp", splited[1])
		if e != nil {
			fmt.Println("ERROR2 ", e)
			return
		}
		lenn, e := client_to_proxy.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		if e != nil {
			fmt.Println("ERROR8 ", e)
			return
		}
		fmt.Println(lenn)

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
			fmt.Println("ERROR7 ", e)
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
	buffer := make([]byte, 64*1024)

	readLeng, err := proxy_to_server.Read(buffer)
	if err != nil {
		fmt.Println("ERROR9 ", err)
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

	go read80(client_to_proxy, proxy_to_server)
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
				fmt.Println("ERR4 ", err)
				return
			}
		}
	}
}

func read80(client_to_proxy net.Conn, proxy_to_server net.Conn) {
	buffer := make([]byte, 32*1024)

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
	buffer := make([]byte, 32*1024)
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
				fmt.Println("ERR4 ", err)
				return
			}
		}
	}
}

func read443(client_to_proxy net.Conn, proxy_to_server net.Conn) {
	buffer := make([]byte, 32*1024)

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

	go write443(client_to_proxy, proxy_to_server)

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
