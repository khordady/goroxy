package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

var bufferSize = 32 * 1024

func main() {
	client_to_proxy, e := net.Dial("tcp", "192.168.1.102:7070")
	if e != nil {
		fmt.Println("ERR2 ", e)
		return
	}

	message := []byte("HI this is testServer")

	writer := bufio.NewWriter(client_to_proxy)
	_, err := writer.Write(intTobytes(len(message)))
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

	bufferReader := make([]byte, bufferSize)
	length, err := readBuffer2(bufferReader, client_to_proxy)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(length)
}

func readBuffer2(buffer []byte, src net.Conn) (int, error) {
	size := make([]byte, 4)
	var total = 0
	var leng int
	var errr error
	fmt.Println("started Reading")

	for leng == 0 {
		leng, errr = src.Read(size)
		fmt.Println("LENG is: ", leng)
		if !os.IsTimeout(errr) && errr != nil {
			fmt.Println("ERRROR IS: ", errr)
		}
	}
	if leng > 0 {
		fmt.Println("LENG > 0 ", leng)
		realSize := bytesToint(size)
		if realSize <= 0 || realSize > bufferSize {
			return 0, errr
		}
		for total < realSize {
			length, errrr := src.Read(buffer[total:realSize])
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
