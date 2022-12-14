package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

var conn net.Conn
var err error

func print(s string) {
	fmt.Print(s)
}

func durin(s string) string {
	conn.Write([]byte(s + "\n"))
	reader := bufio.NewReader(conn)
	data, err := reader.ReadString('\n')
	if err != nil {
		return "(error) connection lost"
	}
	return data
}

func main() {
	conn, err = net.Dial("tcp", "localhost:8045")
	if err != nil {
		log.Fatal("(error) failed to connect to durin")
	}
	defer conn.Close()
	print(durin("set foo bar"))
	print(durin("get foo"))
}
