package main

import (
	"bufio"
	"bytes"
	"crypto/cipher"
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
	ListenEncryptionIV   string
	ListenAuthentication bool
	ListenChain          bool
	ListenUsers          []strUser
	ReadServerFirst      bool
	Server               string
	ServerPort           string
	SendEncryption       string
	SendEncryptionKey    string
	SendEncryptionIV     string
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
var logger = false

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

	initializeEncrypter()

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
	reader := bufio.NewReader(browser_to_client)
	length, err := reader.Read(buffer)
	if err != nil {
		fmt.Println("ERR1 ", err)
		return
	}

	if length == 0 {
		return
	}

	request := processReceived(buffer, length, jjConfig.ListenAuthentication, jjConfig.ListenUsers, jjConfig.ListenEncryption)
	if request == "" {
		return
	}

	var message []byte

	if jjConfig.SendAuthentication {
		message = []byte(jjConfig.SendUserName + "," + jjConfig.SendPassword + "\r\n")
	}
	message = append(message, []byte(request)...)

	printer("Message is: "+request, 0)

	client_to_proxy, e := net.Dial("tcp", jjConfig.Server+":"+jjConfig.ServerPort)
	if e != nil {
		fmt.Println("ERR2 ", e)
		return
	}

	if jjConfig.SendEncryption == "AES" {
		message = encryptAES(message, len(message), jjConfig.SendEncryptionKey, cipher.NewCBCEncrypter(send_aesc, []byte(jjConfig.SendEncryptionIV)))
	}

	if jjConfig.ReadServerFirst {
		_, err := readBuffer(buffer, client_to_proxy)
		if err != nil {
			fmt.Println("ERR31 ", e)
			return
		}
		fmt.Println("READED FIRST")
	}

	writer := bufio.NewWriter(client_to_proxy)
	write_length, err := writer.Write(intTobytes(len(message)))
	if err != nil {
		fmt.Println("ERR31 ", e)
		return
	}
	write_length, e = writer.Write(message)
	if e != nil {
		fmt.Println("ERR3 ", e)
		return
	}
	err = writer.Flush()
	if err != nil {
		fmt.Println("ERR3 ", err)
		return
	}

	printer("WROTED: ", write_length)

	if jjConfig.ListenChain {
		go readChain(writer, browser_to_client)
	} else {
		go readBrowser(writer, browser_to_client)
	}
	readProxy(client_to_proxy, browser_to_client)
}

func printer(message string, params int) {
	if logger {
		fmt.Println(message, params)
	}
}

func readBrowser(writer *bufio.Writer, browser_to_client net.Conn) {
	defer browser_to_client.Close()

	bufferReader := make([]byte, bufferSize-4)

	for {
		length, errr := browser_to_client.Read(bufferReader)

		if length > 0 {
			printer(" READ from browser to client : ", length)

			bufferWriter := processToProxyBuffer(bufferReader, length)
			printer("Decode WRITE from client to proxy: ", len(bufferWriter))

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
			if err != nil {
				fmt.Println("ERR6 ", errw)
				return
			}
			if writeLength == 0 {
				fmt.Println(time.StampMilli, " ERROR Write ")
			}
			printer(" WRITE from client to proxy: ", writeLength)
		}
		if errr != nil {
			fmt.Println(time.StampMilli, " ERROR6 ", errr)
			return
		}
	}
}

func readChain(writer *bufio.Writer, chain_to_client net.Conn) {
	defer chain_to_client.Close()

	bufferReader := make([]byte, bufferSize-4)

	for {
		length, errr := readBuffer(bufferReader[:bufferSize], chain_to_client)

		if length > 0 {
			printer(" READ from browser to client : ", length)

			bufferWriter := processToProxyBuffer(bufferReader, length)
			//fmt.Println(time.StampMilli, "Decode WRITE from client to proxy: ", len(bufferWriter))
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
			printer(" WRITE from client to proxy: ", writeLength)
		}
		if errr != nil {
			fmt.Println(time.StampMilli, " ERROR6 ", errr)
			return
		}
	}
}

func readProxy(client_to_proxy net.Conn, browser_to_client net.Conn) {
	defer client_to_proxy.Close()

	writer := bufio.NewWriter(browser_to_client)
	bufferReader := make([]byte, bufferSize)

	for {
		total, errr := readBuffer(bufferReader[:bufferSize], client_to_proxy)
		if total > 0 {
			printer(" Encoded READ from proxy to client: ", total)
			//fmt.Println(string(buffer))

			bufferWriter := processToBrowserBuffer(bufferReader, total)
			printer(" Decoded WRITE from client to browser: ", total)
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
			printer(" WRITE from client to browser: ", write_length)
		}

		if errr != nil {
			fmt.Println(time.StampMilli, " ERROR81 ", errr)
			return
		}
	}
}

func readBuffer(buffer []byte, src net.Conn) (int, error) {
	size := make([]byte, 4)
	var total = 0
	fmt.Println("started Reading")

	leng, errr := src.Read(size)
	//fmt.Println("LENG is: ", leng)
	if errr != nil {
		fmt.Println("ERRROR IS: ", errr)
	}
	if leng > 0 {
		//fmt.Println("LENG > 0 ", leng)
		realSize := bytesToint(size)
		if realSize <= 0 || realSize > bufferSize {
			return 0, fmt.Errorf("ERROR OVER SIZE", size)
		}
		for total < realSize {
			length, errrr := src.Read(buffer[total:realSize])
			//fmt.Println("Readed is: ", length)
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
