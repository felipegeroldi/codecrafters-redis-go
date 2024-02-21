package main

import (
	"errors"
	"flag"
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

	var portNumber uint
	flag.UintVar(&portNumber, "port", 6379, "The port number of application")
	flag.Parse()

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", portNumber))
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

func echo(msg string) string {
	return fmt.Sprintf("+%s\r\n", msg)
}

func ping() string {
	return "+PONG\r\n"
}

func set(args []string) string {
	base[args[0]] = args[1]

	if len(args) > 3 {
		if strings.ToLower(args[2]) == "px" {
			ms, _ := strconv.Atoi(args[3])
			expire[args[0]] = time.Now().Add(time.Millisecond * time.Duration(ms))
		}
	}

	return "+OK\r\n"
}

func get(args []string) string {
	expireTime := expire[args[0]]
	val := base[args[0]]
	if expireTime.IsZero() && val != "" || !expireTime.IsZero() && expireTime.After(time.Now()) {
		return fmt.Sprintf("+%s\r\n", base[args[0]])
	} else {
		return "$-1\r\n"
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

		switch strings.ToLower(args[0]) {
		case REDIS_CMD_PING:
			returnVal = ping()
		case REDIS_CMD_ECHO:
			returnVal = echo(args[1])
		case REDIS_CMD_SET:
			returnVal = set(args[1:])
		case REDIS_CMD_GET:
			returnVal = get(args[1:])
		default:
			returnVal = "-not supported command\r\n"
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
