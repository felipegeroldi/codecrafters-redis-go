package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var base = make(map[string]string)
var expire = make(map[string]time.Time)

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
	var cmdLen int
	var returnVal string
	if cmd[0] == REDIS_ARR {
		cmdLen, _ = strconv.Atoi(string(cmd[1]))
		cmd = cmd[4:]
	}

	switch cmd[0] {
	case REDIS_BULK_STR:
		args := parseBulkStr(cmd, cmdLen)
		if strings.ToLower(args[0]) == "ping" {
			returnVal = "+PONG\r\n"
		} else if strings.ToLower(args[0]) == "echo" {
			returnVal = fmt.Sprintf("+%s\r\n", args[1])
		} else if strings.ToLower(args[0]) == "set" {
			base[args[1]] = args[2]

			if len(args) > 3 {
				if strings.ToLower(args[3]) == "px" {
					ms, _ := strconv.Atoi(args[4])
					expire[args[1]] = time.Now().Add(time.Millisecond * time.Duration(ms))
				}
			}

			returnVal = "+OK\r\n"
		} else if strings.ToLower(args[0]) == "get" {
			expireTime := expire[args[1]]
			val := base[args[1]]
			if expireTime.IsZero() && val != "" || !expireTime.IsZero() && expireTime.After(time.Now()) {
				returnVal = fmt.Sprintf("+%s\r\n", base[args[1]])
			} else {
				returnVal = "$-1\r\n"
			}
		}
	default:
		returnVal = "-not supported data type\r\n"
	}

	return returnVal
}

func parseBulkStr(cmd []byte, len int) []string {
	var args []string

	for i := len; i > 0; i-- {
		values := strings.Split(string(cmd), "\r\n")

		argLen, _ := strconv.Atoi(values[0][1:])
		arg := string(values[1][:argLen])
		args = append(args, arg)

		cmd = []byte(strings.Join(values[2:], "\r\n"))
	}
	return args
}
