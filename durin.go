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

var version = "0.0.2"
var port = 8045
var data = read("dump.db")
var mu sync.Mutex

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
		d := read("dump.db")
		mu.Lock()
		eq := reflect.DeepEqual(d, data)
		if !eq {
			store("dump.db", data)
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
	if message[0:4] != "set " && message[0:4] != "get " && message[0:4] != "del " {
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
	if key == "" {
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
	go persist()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handle_connection(conn)
	}
}

func boot() {
	fmt.Printf(`
     ___
    /\  \
   /::\  \       Durin %s
  /:/\:\  \
 /:/  \:\  \
/:/__/ \:\__\    Running in persistence mode
\:\  \ /:/  /    Port: %d
 \:\  /:/  /     PID:  %d
  \:\/:/  /
   \::/  /             https://github.com/chrisvarga/durin
    \/__/

`, version, 8045, os.Getpid())
}

func main() {
	boot()
	listen()
}
