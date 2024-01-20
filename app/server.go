package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const respSymbols = "+-:$*"

var commandHandlers = map[string]func([]string) string{
	"ping": PONG,
	"echo": ECHO,
}

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

		command, args := ParseRESPCommand(msg)

		if handler, ok := commandHandlers[command]; ok {
			response := handler(args)
			conn.Write([]byte(response))
		} else {
			conn.Write([]byte("-ERR unknown command '" + command + "'\r\n"))
		}
	}
}

func ParseRESPCommand(command string) (string, []string) {
	args := strings.Split(command, "\r\n")

	var elements []string

	for _, arg := range args {
		if len(arg) == 0 || strings.Contains(respSymbols, string(arg[0])) {
			continue
		}

		elements = append(elements, strings.ToLower(arg))
	}

	return elements[0], elements[1:]
}

func PONG([]string) string {
	return "+PONG\r\n"
}

func ECHO(args []string) string {
	text := args[0]
	return fmt.Sprintf("+%s\r\n", text)
}
