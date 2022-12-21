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
	WriteServerFirst     bool
	ListenPort           string
	ListenEncryption     string
	ListenEncryptionKey  string
	ListenEncryptionIV   string
	ListenAuthentication bool
	ListenUsers          []strUser
}

type strUser struct {
	ListenUserName string
	ListenPassword string
}

var jjConfig strServerConfig

var bufferSize = 32 * 1024
var logger = false

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

	initializeEncrypter()

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
	writer := bufio.NewWriter(client_to_proxy)

	if jjConfig.WriteServerFirst {
		message := []byte("HI THIS IS TEST")
		if jjConfig.ListenEncryption == "AES" {
			message = encryptAES(message, len(message), jjConfig.ListenEncryptionKey)
		}
		_, err := writer.Write(intTobytes(len(message)))
		if err != nil {
			fmt.Println(time.StampMilli, " ERROR42 ", err)
			return
		}
		_, err = writer.Write(message)
		if err != nil {
			fmt.Println(time.StampMilli, " ERROR42 ", err)
			return
		}
		err = writer.Flush()
		if err != nil {
			fmt.Println(time.StampMilli, " ERROR42 ", err)
			return
		}
	}
	buffer := make([]byte, 9*1024)
	reader := bufio.NewReader(client_to_proxy)

	length, err := readBuffer(buffer, reader)
	if err != nil {
		fmt.Println(" Error 20", err)
		return
	}

	if length == 0 {
		return
	}
	message := processReceived(buffer, length, jjConfig.ListenAuthentication, jjConfig.ListenUsers,
		jjConfig.ListenEncryption, jjConfig.ListenEncryptionKey)
	if message == "" {
		return
	}

	//printer("MESSAGE IS: "+message, 0)

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
			fmt.Println(time.StampMilli, " ERROR3 ", e)
			return
		}

		printer("CONNECTED TO: "+host[1], 0)

		bytess := []byte("HTTP/1.1 200 Connection Established\r\n\r\n")
		if jjConfig.ListenEncryption == "AES" {
			bytess = encryptAES(bytess, len(bytess), jjConfig.ListenEncryptionKey)
		}

		write_length, err := writer.Write(intTobytes(len(bytess)))
		if err != nil {
			fmt.Println(time.StampMilli, " ERROR42 ", err)
			return
		}
		write_length, err = writer.Write(bytess)
		if err != nil {
			fmt.Println(time.StampMilli, " ERROR42 ", err)
			return
		}
		err = writer.Flush()
		if err != nil {
			fmt.Println(time.StampMilli, " ERROR42 ", err)
			return
		}

		if err != nil {
			fmt.Println(time.StampMilli, " ERROR42 ", err)
			return
		}
		printer("WROTED ", write_length)

		go read(client_to_proxy, proxy_to_server, reader)
		write(client_to_proxy, proxy_to_server)

	} else {
		proxy_to_server, e := net.Dial("tcp", host[1]+":80")
		if e != nil {
			fmt.Println(time.StampMilli, " ERROR5 ", e)
			return
		}

		writer2 := bufio.NewWriter(proxy_to_server)
		write_length, e := writer2.Write([]byte(message))
		if e != nil {
			fmt.Println(time.StampMilli, " ERROR6 ", e)
			return
		}
		e = writer2.Flush()
		if e != nil {
			fmt.Println(time.StampMilli, " ERROR6 ", e)
			return
		}
		printer("WROTED ", write_length)

		go read(client_to_proxy, proxy_to_server, reader)
		write(client_to_proxy, proxy_to_server)
	}
}

func printer(message string, params int) {
	if logger {
		fmt.Println(message, params)
	}
}

func write(client_to_proxy net.Conn, proxy_to_host net.Conn) {
	defer proxy_to_host.Close()

	bufferReader := make([]byte, (bufferSize)-4)

	for {
		length, err := proxy_to_host.Read(bufferReader)
		if length > 0 {
			//fmt.Println(time.StampMilli, " READ from host to proxy: ", length)
			//fmt.Println(string(buffer[:length]))
			bufferWriter := processToClientBuffer(bufferReader, length)
			//fmt.Println(time.StampMilli, " Encoded WRITE from proxy to client:", len(bufferWriter))
			writeLength, errw := client_to_proxy.Write(intTobytes(len(bufferWriter)))
			if errw != nil {
				fmt.Println(time.StampMilli, " ERROR4 ", errw)
				return
			}
			writeLength, errw = client_to_proxy.Write(bufferWriter)
			if errw != nil {
				fmt.Println(time.StampMilli, " ERROR4 ", errw)
				return
			}
			if writeLength == 0 {
				fmt.Println(time.StampMilli, " ERROR Write ")
			}
			//fmt.Println(time.StampMilli, " WRITE from proxy to client: ", writeLength)
		}
		if err != nil {
			fmt.Println(time.StampMilli, " ERROR8 ", err)
			return
		}
	}
}

func read(client_to_proxy net.Conn, proxy_to_host net.Conn, reader *bufio.Reader) {
	defer client_to_proxy.Close()
	writer := bufio.NewWriter(proxy_to_host)

	for {
		bufferReader := make([]byte, bufferSize)

		length, errr := readBuffer(bufferReader, reader)
		if length > 0 {
			//fmt.Println(time.StampMilli, " Read from client to proxy: ", length)

			bufferWriter := processToHostBuffer(bufferReader, length)
			//fmt.Println(time.StampMilli, "Decoded WRITE from proxy to host : ", length)
			//fmt.Println(string(buffer))
			writeLength, errw := writer.Write(bufferWriter)
			if errw != nil {
				fmt.Println(time.StampMilli, " ERROR5 ", errw)
				return
			}
			errw = writer.Flush()
			if errw != nil {
				fmt.Println(time.StampMilli, " ERROR5 ", errw)
				return
			}
			if writeLength == 0 {
				fmt.Println(time.StampMilli, " ERROR Write ")
			}
			//fmt.Println(time.StampMilli, " WRITE from proxy to host :", writeLength)
		}

		if errr != nil {
			fmt.Println(time.StampMilli, " ERROR51 ", errr)
			return
		}
	}
}

func readBuffer(buffer []byte, reader *bufio.Reader) (int, error) {
	size := make([]byte, 4)
	var total = 0

	fmt.Println("started Reading")
	leng, errr := reader.Read(size)
	if leng > 0 {
		realSize := bytesToint(size)
		if realSize <= 0 || realSize > bufferSize {
			return 0, fmt.Errorf(time.StampMilli, " ERROR OVER SIZE")
		}
		printer("Real size is: ", realSize)
		for total < realSize {
			length, errrr := reader.Read(buffer[total:realSize])
			printer("Readed is: ", length)
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
