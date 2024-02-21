package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

var base = make(map[string]string)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(c)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)

	for {
		_, err := conn.Read(buf)
		if errors.Is(err, io.EOF) {
			return
		}

		conn.Write([]byte(parseCommand(buf)))
	}
}

func parseCommand(cmd []byte) string {
	var len int
	var returnVal string
	if cmd[0] == REDIS_ARR {
		len, _ = strconv.Atoi(string(cmd[1]))
		cmd = cmd[4:]
	}

	switch cmd[0] {
	case REDIS_BULK_STR:
		args := parseBulkStr(cmd, len)
		if strings.ToLower(args[0]) == "ping" {
			returnVal = "+PONG\r\n"
		} else if strings.ToLower(args[0]) == "echo" {
			returnVal = fmt.Sprintf("+%s\r\n", args[1])
		} else if strings.ToLower(args[0]) == "set" {
			base[args[1]] = args[2]
			returnVal = "+OK\r\n"
		} else if strings.ToLower(args[0]) == "get" {
			returnVal = fmt.Sprintf("+%s\r\n", base[args[1]])
		}
	default:
		returnVal = "-not supported data type\r\n"
	}

	return returnVal
}

func parseBulkStr(cmd []byte, len int) []string {
	var args []string

	for i := len; i > 0; i-- {
		argLen, _ := strconv.Atoi(string(cmd[1]))
		arg := string(cmd[4 : argLen+4])
		args = append(args, arg)

		cmd = cmd[argLen+4+2:]
	}
	return args
}
