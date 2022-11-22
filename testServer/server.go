package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

var bufferSize = 32 * 1024

func main() {
	ln, err := net.Listen("tcp", ":7070")
	conn, _ := ln.Accept()
	//err = conn.SetDeadline(time.Time{})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Connection Received")
	handleSocket(conn)
}

func handleSocket(client_to_proxy net.Conn) {
	buffer := make([]byte, 9*1024)
	reader := bufio.NewReader(client_to_proxy)

	length, _ := readBuffer(buffer, reader)
	fmt.Println(length)

	fmt.Println(string(buffer))

	writer := bufio.NewWriter(client_to_proxy)

	for {
		bytess := []byte("HTTP/1.1 200 Connection Established\r\n\r\n")

		Writelength, err := writer.Write(intTobytes(len(bytess)))
		if err != nil {
			fmt.Println(time.StampMilli, " ERROR42 ", err)
			return
		}
		Writelength, err = writer.Write(bytess)
		if err != nil {
			fmt.Println(time.StampMilli, " ERROR42 ", err)
			return
		}
		err = writer.Flush()
		if err != nil {
			fmt.Println(time.StampMilli, " ERROR42 ", err)
			return
		}
		//client_to_proxy.SetWriteDeadline(time.Now().Add(1 * time.Second))

		if err != nil {
			fmt.Println(time.StampMilli, " ERROR42 ", err)
			return
		}
		fmt.Println("WROTED 200: ", Writelength)
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
		fmt.Println("Real size is: ", realSize)
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
	fmt.Println("Total and error is: ", total, errr)
	return total, errr
}
