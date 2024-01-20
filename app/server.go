package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	RESP_SYMBOLS = "+-:$*"
	NOT_FOUND    = "$-1\r\n"
)

var commandHandlers = map[string]func([]string) string{
	"ping": Pong,
	"echo": Echo,
	"get":  Get,
	"set":  Set,
}

type Pair struct {
	value  string
	expire *time.Time
}

var db = struct {
	data map[string]Pair
}{
	data: make(map[string]Pair),
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
		if len(arg) == 0 || strings.Contains(RESP_SYMBOLS, string(arg[0])) {
			continue
		}

		elements = append(elements, strings.ToLower(arg))
	}

	return elements[0], elements[1:]
}

func Pong([]string) string {
	return "+PONG\r\n"
}

func Echo(args []string) string {
	text := args[0]
	return fmt.Sprintf("+%s\r\n", text)
}

func Get(args []string) string {
	key := args[0]
	pair := db.data[key]

	if pair == (Pair{}) {
		return NOT_FOUND
	}

	if pair.expire != nil && pair.expire.Before(time.Now()) {
		delete(db.data, key)
		return NOT_FOUND
	}

	return fmt.Sprintf("+%s\r\n", pair.value)
}

func Set(args []string) string {
	key := args[0]
	value := args[1]
	var expire *time.Time = nil

	if len(args) > 2 {
		if args[2] != "px" {
			expireMilliSeconds, _ := strconv.Atoi(args[2])
			time := time.Now().Add(time.Duration(expireMilliSeconds) * time.Millisecond)
			expire = &time
		}
	}

	db.data[key] = Pair{value, expire}

	return "+OK\r\n"
}
