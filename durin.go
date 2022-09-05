package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

var version = "0.0.3"
var port = 8045
var data = read("keys.db")
var mu sync.Mutex
var durable bool

func read(file string) map[string]string {
	data, err := os.ReadFile(file)
	if err != nil {
		return make(map[string]string)
	}
	var result map[string]string
	json.Unmarshal([]byte(string(data)), &result)
	return result
}

func store(file string, data map[string]string) {
	s, _ := json.MarshalIndent(data, "", "  ")
	err := os.WriteFile(file, append([]byte(s), "\n"...), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func persist() {
	for {
		time.Sleep(1 * time.Second)
		d := read("keys.db")
		mu.Lock()
		eq := reflect.DeepEqual(d, data)
		if !eq {
			store("keys.db", data)
			log.Println(" * DB saved on disk")
		}
		mu.Unlock()
	}
}

func route_request(command string, key string, value string) string {
	mu.Lock()
	defer mu.Unlock()
	if command == "get" {
		if val, ok := data[key]; ok {
			return val
		}
		return "(error): key not found"
	}
	if command == "set" {
		data[key] = value
		return "OK"
	}
	if command == "del" {
		delete(data, key)
		return "OK"
	}
	if command == "key" {
		if len(data) == 0 {
			return "nil"
		}
		var response strings.Builder
		response.WriteString("[")
		for k := range data {
			response.WriteString(fmt.Sprintf("'%s',", k))
		}
		res_string := response.String()
		res_string = res_string[:len(res_string)-1]
		return res_string + "]"
	}
	return "(error): invalid command"
}

func parse_request(message string) string {
	var command string
	var key string
	var value string

	// Parse command
	if len(message) < 4 {
		return "(error): invalid syntax"
	}
	if message[0:4] != "set " && message[0:4] != "get " && message[0:4] != "del " && message[0:4] != "keys" {
		return "(error): invalid syntax"
	}
	command = message[0:3]

	// Parse key
	if idx := strings.IndexByte(message[4:], ' '); idx >= 0 {
		// set command
		key = message[4 : idx+4]
	} else {
		// get or del command, need to trim newline
		key = strings.TrimSuffix(message[4:], "\n")
	}
	if len(key) == 0 && command != "key" {
		return "(error): invalid key"
	}

	// If we got this far, the rest of the message is the value.
	if command == "set" {
		value = strings.TrimSpace(message[len(command)+len(key)+2:])
		if len(value) == 0 {
			return "(error): invalid syntax"
		}
	}
	return route_request(command, key, value)
}

func handle_connection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			conn.Close()
			return
		}
		conn.Write([]byte(parse_request(string(message)) + "\n"))
	}
}

func listen() {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("Listening at", listener.Addr().String())
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handle_connection(conn)
	}
}

func boot() {
	var mode string
	if durable {
		mode = "persistence"
	} else {
		mode = "ephemeral"
	}
	fmt.Printf(`
     ___
    /\  \
   /::\  \       Durin %s
  /:/\:\  \
 /:/  \:\  \
/:/__/ \:\__\    Running in %s mode
\:\  \ /:/  /    Port: %d
 \:\  /:/  /     PID:  %d
  \:\/:/  /
   \::/  /             https://github.com/chrisvarga/durin
    \/__/

`, version, mode, port, os.Getpid())
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-p" {
		durable = true
		go persist()
	}
	boot()
	listen()
}
