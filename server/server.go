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

var bufferSize = 32 * 1024

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
	length, err := readBuffer(buffer, client_to_proxy)

	if length == 0 {
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

		case "AES":
			bytess = encryptAES(bytess, len(bytess), jjConfig.ListenEncryptionKey)
			break
		}
		Writelength, err := client_to_proxy.Write(intTobytes(len(bytess)))
		Writelength, err = client_to_proxy.Write(bytess)
		fmt.Println("WROTED 200: ", Writelength)

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

	bufferReader := make([]byte, (bufferSize)-4)

	for {
		length, err := proxy_to_host.Read(bufferReader)
		if length > 0 {
			fmt.Println(time.Now().Format(time.Stamp) + " READ from host to proxy:" + strconv.Itoa(length))
			//fmt.Println(string(buffer[:length]))
			bufferWriter := processToClientBuffer(bufferReader, length)
			fmt.Println(time.Now().Format(time.Stamp) + " Encoded WRITE from proxy to client:" + strconv.Itoa(len(bufferWriter)))
			//fmt.Println(string(buffer))
			_, errw := client_to_proxy.Write(intTobytes(len(bufferWriter)))
			if errw != nil {
				fmt.Println("ERR4 ", errw)
				return
			}
			_, errw = client_to_proxy.Write(bufferWriter)
			if errw != nil {
				fmt.Println("ERR4 ", errw)
				return
			}
			//fmt.Println(time.Now().Format(time.Stamp) + " WRITE from proxy to client:" + strconv.Itoa(writeLength))
		}
		if err != nil {
			fmt.Println("ERROR8 ", err)
			return
		}
	}
}

func read(client_to_proxy net.Conn, proxy_to_host net.Conn) {
	defer client_to_proxy.Close()
	bufferReader := make([]byte, bufferSize)

	for {
		length, errr := readBuffer(bufferReader, client_to_proxy)
		if length > 0 {
			fmt.Println(time.Now().Format(time.Stamp) + " Read from host to proxy :" + strconv.Itoa(length))

			bufferWriter := processToHostBuffer(bufferReader, length)
			fmt.Println(time.Now().Format(time.Stamp) + "Decoded WRITE from proxy to host :" + strconv.Itoa(length))
			//fmt.Println(string(buffer))
			_, errw := proxy_to_host.Write(bufferWriter)
			if errw != nil {
				fmt.Println("ERR5 ", errw)
				return
			}
			//fmt.Println(time.Now().Format(time.Stamp) + " WRITE from proxy to host :" + strconv.Itoa(writeLength))
		}

		if errr != nil {
			fmt.Println("ERR51 ", errr)
			return
		}
	}
}

func readBuffer(buffer []byte, src net.Conn) (int, error) {
	size := make([]byte, 4)

	var total = 0
	leng, err := src.Read(size)
	if leng <= 0 || leng > bufferSize {
		return 0, fmt.Errorf("ERROR")
	}
	if leng > 0 {
		realSize := bytesToint(size)
		for total < realSize {
			length, errr := src.Read(buffer[total:realSize])
			total = total + length

			if errr != nil {
				return total, errr
			}
		}
	}
	return total, err
}
