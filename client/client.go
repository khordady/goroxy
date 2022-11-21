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
	buffer := make([]byte, 8*1024)
	reader := bufio.NewReader(browser_to_client)
	length, err := reader.Read(buffer)
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

	writer := bufio.NewWriter(client_to_proxy)
	_, err = writer.Write(intTobytes(len(message)))
	if err != nil {
		fmt.Println("ERR31 ", e)
		return
	}
	_, e = writer.Write(message)
	if e != nil {
		fmt.Println("ERR3 ", e)
		return
	}
	err = writer.Flush()
	if err != nil {
		fmt.Println("ERR3 ", err)
		return
	}

	go write(client_to_proxy, writer, browser_to_client)
	read(client_to_proxy, browser_to_client)

	browser_to_client.Close()
	client_to_proxy.Close()
}

func write(client_to_proxy net.Conn, writer *bufio.Writer, browser_to_client net.Conn) {
	bufferReader := make([]byte, bufferSize-4)

	for {
		length, errr := browser_to_client.Read(bufferReader)

		if length > 0 {
			fmt.Println(time.StampMilli, " READ from browser to client : ", length)
			//fmt.Println(string(buffer[:length]))

			bufferWriter := processToProxyBuffer(bufferReader, length)
			fmt.Println(time.StampMilli, "Decode WRITE from client to proxy: ", len(bufferWriter))
			//fmt.Println(string(buffer))

			writeLength, errw := writer.Write(intTobytes(len(bufferWriter)))
			if errw != nil {
				fmt.Println("ERR6 ", errw)
				return
			}
			writeLength, errw = writer.Write(bufferWriter)
			if errw != nil {
				fmt.Println("ERR6 ", errw)
				return
			}
			err := writer.Flush()
			//client_to_proxy.SetWriteDeadline(time.Now().Add(1 * time.Second))
			if err != nil {
				fmt.Println("ERR6 ", errw)
				return
			}
			fmt.Println(time.StampMilli, " WRITE from client to proxy: ", writeLength)
		}
		if errr != nil {
			fmt.Println(time.StampMilli, " ERROR6 ", errr)
			return
		}
	}
}

func read(client_to_proxy net.Conn, browser_to_client net.Conn) {
	writer := bufio.NewWriter(browser_to_client)
	reader := bufio.NewReader(client_to_proxy)

	for {
		bufferReader := make([]byte, bufferSize)

		total, errr := readBuffer(bufferReader, reader, client_to_proxy)
		if total > 0 {
			fmt.Println(time.StampMilli, " Encoded READ from proxy to client: ", total)
			//fmt.Println(string(buffer))

			bufferWriter := processToBrowserBuffer(bufferReader, total)
			//fmt.Println(time.StampMilli," Decoded WRITE from client to browser: ", total)
			//fmt.Println(string(buffer))

			write_length, errw := writer.Write(bufferWriter)
			if errw != nil {
				fmt.Println(time.StampMilli, " ERROR8 ", errw)
				return
			}
			errw = writer.Flush()
			if errw != nil {
				fmt.Println(time.StampMilli, " ERROR8 ", errw)
				return
			}

			fmt.Println(time.StampMilli+" WRITE from client to browser: ", write_length)
		}

		if errr != nil {
			fmt.Println(time.StampMilli, " ERROR81 ", errr)
			return
		}
	}
}

func readBuffer(buffer []byte, reader *bufio.Reader, src net.Conn) (int, error) {
	size := make([]byte, 4)
	var total = 0

	fmt.Println("started Reading")
	for reader.Buffered() == 0 {
		fmt.Println("peaking 1")
		src.SetReadDeadline(time.Now().Add(1 * time.Second))
		peek, err := reader.Peek(1)
		if !os.IsTimeout(err) && err != nil {
			fmt.Println("ERR ", peek, err)
		}
		fmt.Println(peek)
	}
	leng, errr := reader.Read(size)
	if leng > 0 {
		realSize := bytesToint(size)
		if realSize <= 0 || realSize > bufferSize {
			return 0, fmt.Errorf(time.StampMilli, " ERROR OVER SIZE", size)
		}
		for total < realSize {
			length, errrr := reader.Read(buffer[total:realSize])
			fmt.Println("Readed is: ", length)
			total = total + length

			if errrr != nil {
				fmt.Println("Total and error is: ", total, errrr)
				return total, errrr
			}
		}
	}
	if errr != nil {
		fmt.Println("Total and error is: ", total, errr)
		return total, errr
	}
	return total, errr
}
