package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		buffer := make([]byte, 1024)
		len, err := conn.Read(buffer)

		if err == io.EOF {
			fmt.Println("Connection closed")
			return
		}

		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			return
		}

		msg := string(buffer[:len])

		fmt.Println("Received data: ", msg)

		_, err = conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			fmt.Println("Error writing: ", err.Error())
		}
	}
}
