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

var bufferSize = 32 * 1024

func main() {
	//a := []int{0, 1, 2, 3, 4, 5, 6}
	//b := a[:4]
	//fmt.Println(len(a), len(b))

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
		err = conn.SetReadDeadline(time.Time{})
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
	length, err := bufio.NewReader(browser_to_client).Read(buffer)
	if err != nil {
		fmt.Println("ERR1 ", err)
		return
	}

	if length == 0 {
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

	//fmt.Println("Message is: " + request)

	client_to_proxy, e := net.Dial("tcp", jjConfig.Server+":"+jjConfig.ServerPort)
	if e != nil {
		fmt.Println("ERR2 ", e)
		return
	}

	if jjConfig.SendEncryption == "AES" {
		message = encryptAES(message, len(message), jjConfig.SendEncryptionKey)
	}

	_, err = client_to_proxy.Write(intTobytes(len(message)))
	if err != nil {
		fmt.Println("ERR31 ", e)
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

	bufferReader := make([]byte, bufferSize-4)

	for {
		length, errr := browser_to_client.Read(bufferReader)
		if length > 0 {
			//fmt.Println(time.Now().Format(time.Stamp) + " READ from browser to client : " + strconv.Itoa(length))
			//fmt.Println(string(buffer[:length]))

			bufferWriter := processToProxyBuffer(bufferReader, length)
			//fmt.Println(time.Now().Format(time.Stamp) + "Decode WRITE from client to proxy: " + strconv.Itoa(len(bufferWriter)))
			//fmt.Println(string(buffer))

			_, errw := client_to_proxy.Write(intTobytes(len(bufferWriter)))
			_, errw = client_to_proxy.Write(bufferWriter)
			if errw != nil {
				fmt.Println("ERR6 ", errw)
				return
			}
			//fmt.Println(time.Now().Format(time.Stamp) + " WRITE from client to proxy: " + strconv.Itoa(writeLength))
		}
		if errr != nil {
			fmt.Println("ERR6 ", errr)
			return
		}
	}
}

func read(client_to_proxy net.Conn, browser_to_client net.Conn) {
	defer browser_to_client.Close()

	bufferReader := make([]byte, bufferSize)

	for {
		total, errr := readBuffer(bufferReader, client_to_proxy)
		if total > 0 {
			//fmt.Println(time.Now().Format(time.Stamp)+" Encoded READ from proxy to client: ", total)
			//fmt.Println(string(buffer))

			bufferWriter := processToBrowserBuffer(bufferReader, total)
			//fmt.Println(time.Now().Format(time.Stamp)+" Decoded WRITE from client to browser: ", total)
			//fmt.Println(string(buffer))

			_, errw := browser_to_client.Write(bufferWriter)
			if errw != nil {
				fmt.Println("ERR8 ", errw)
				return
			}

			//fmt.Println(time.Now().Format(time.Stamp)+" WRITE from client to browser: ", write_length)
		}

		if errr != nil {
			fmt.Println("ERR81 ", errr)
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
