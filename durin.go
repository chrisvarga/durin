package main

import (
	"bufio"
	js "encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

var version = "0.0.7"
var host = "localhost"
var port = 8045
var db string
var data = make(map[string]string)
var mu sync.Mutex
var cluster []string

// Read the database file into memory.
func read(file string) map[string]string {
	data, err := os.ReadFile(file)
	if err != nil {
		return make(map[string]string)
	}
	var result map[string]string
	err = js.Unmarshal([]byte(string(data)), &result)
	if err != nil {
		log.Fatal("Error reading database file:", err)
	}
	return result
}

// Store the database to disk in json format.
func store(file string, data map[string]string) {
	s, _ := js.MarshalIndent(data, "", "  ")
	err := os.WriteFile(file, append([]byte(s), "\n"...), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

// Write the database to disk if there were any data changes.
// We check for changes once every second.
func persist() {
	for {
		time.Sleep(1 * time.Second)
		d := read(db)
		mu.Lock()
		eq := reflect.DeepEqual(d, data)
		if !eq {
			store(db, data)
			log.Println(" * DB saved on disk")
		}
		mu.Unlock()
	}
}

// Get a key from the database.
func get(key string) string {
	if value, ok := data[key]; ok {
		return value
	}
	return "(error) key not found"
}

// Set a key in the database to a value.
func set(key string, value string) string {
	data[key] = value
	return "OK"
}

// Delete a key and its value from the database.
func del(key string) string {
	delete(data, key)
	return "OK"
}

// Return a list of the keys in the database.
// If they specified a prefix argument, only return the keys starting with it.
func keys(prefix string) string {
	if len(data) == 0 {
		// This means the database is completely empty.
		return "[]"
	}
	var buf strings.Builder
	buf.WriteString("[")
	if len(prefix) == 0 {
		// If they didn't specify a prefix, just return all the keys.
		for k := range data {
			buf.WriteString(fmt.Sprintf("\"%s\",", k))
		}
	} else {
		// If they specified a prefix, only return the keys starting with it.
		for k := range data {
			if strings.HasPrefix(k, prefix) {
				buf.WriteString(fmt.Sprintf("\"%s\",", k))
			}
		}
	}
	response := buf.String()
	response = response[:len(response)-1]
	if len(response) == 0 {
		// This means there were no keys starting with the specified prefix.
		return "[]"
	}
	return response + "]"
}

// Return a json object of key/values of all keys starting with the prefix.
func json(prefix string) string {
	if len(data) == 0 {
		// This means the database is completely empty.
		return "{}"
	}
	var buf strings.Builder
	buf.WriteString("{")
	for k, v := range data {
		if strings.HasPrefix(k, prefix) {
			buf.WriteString(fmt.Sprintf("\"%s\":\"%s\",", k, v))
		}
	}
	response := buf.String()
	response = response[:len(response)-1]
	if len(response) == 0 {
		// This means there were no keys starting with the specified prefix.
		return "{}"
	}
	return response + "}"
}

// Route the command and arguments to the appropriate function.
func route(command string, key string, value string) string {
	mu.Lock()
	defer mu.Unlock()
	switch command {
	case "get":
		return get(key)
	case "set":
		return set(key, value)
	case "del":
		return del(key)
	case "keys":
		return keys(key)
	case "json":
		return json(key)
	default:
		return "(error) invalid syntax"
	}
}

// Parse the message into a valid command, key, and value.
// Syntax:
//   set  <key> <value>
//   get  <key>
//   del  <key>
//   keys [prefix]
//   json <prefix>
func parse(message string) string {
	var command string
	var key string
	var value string

	// Parse command.
	if len(message) < 4 {
		return "(error) invalid syntax"
	}
	switch message[0:4] {
	case "set ", "get ", "del ":
		command = message[0:3]
	case "keys", "json":
		command = message[0:4]
	default:
		return "(error) invalid syntax"
	}

	// Parse key.
	if command == "keys" {
		// The prefix argument is optional for the keys command; check for it.
		if idx := strings.IndexByte(message[4:], '\n'); idx >= 0 {
			if idx >= 1 && message[0:5] != "keys " {
				// This means the command started with 'keys' but wasn't valid.
				// For example, they sent something like 'keysasdflsdf'.
				return "(error) invalid command; did you mean 'keys'?"
			}
			key = strings.TrimSuffix(message[5:], "\n")
			key = strings.ReplaceAll(key, " ", "")
			if len(key) != 0 && len(key)+1 != idx {
				// This means there was a space after the keys command, but
				// they never specified a non-empty prefix, i.e. only spaces.
				return "(error) invalid keys prefix"
			}
		}
	} else if command == "json" {
		// Trim trailing newline from the key.
		if idx := strings.IndexByte(message[5:], ' '); idx >= 0 {
			key = message[5 : idx+5]
		} else {
			key = strings.TrimSuffix(message[5:], "\n")
		}
	} else {
		if idx := strings.IndexByte(message[4:], ' '); idx >= 0 {
			// For the set command, we need to remove a trailing space from the key.
			key = message[4 : idx+4]
		} else {
			// For get or del, we need to trim a trailing newline from the key.
			key = strings.TrimSuffix(message[4:], "\n")
		}
	}
	// A key is required for all commands other than the keys command.
	if len(key) == 0 && command != "keys" {
		return "(error) invalid syntax"
	}

	// If we got this far, the rest of the message is the value for set.
	if command == "set" {
		value = strings.TrimSpace(message[len(command)+len(key)+2:])
		if len(value) == 0 {
			return "(error) value required for set command"
		}
	}
	return route(command, key, value)
}

// TODO: experimental clustering
// cluster = append(cluster, "localhost:8046")
func forward(message string) {
	for _, node := range cluster {
		conn, err := net.Dial("tcp", node)
		if err != nil {
			log.Println("(error) failed to connect to node at ", node)
		}
		defer conn.Close()
		fmt.Fprintf(conn, message)
	}
}

// Handle a request. All requests must be terminated by a newline.
// We keep looping so they can reuse the connection until they kill it.
func handle(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			conn.Close()
			return
		}
		// We terminate all responses with a newline.
		fmt.Fprintf(conn, "%s\n", parse(string(message)))
		// go forward(message)
	}
}

// Listen on the port specified by the port variable at the top of this file.
// We listen on the private loopback interface (i.e. localhost).
// Right now we just spin up a lightweight go routine for each connection.
func listen() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
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
		go handle(conn)
	}
}

// Parse command line arguments and set the database config accordingly.
func flags() {
	d := flag.String("d", "", "specifies a database file, enabling durable mode")
	b := flag.String("b", host, "specifies the bind address")
	p := flag.Int("p", port, "specifies the port on which to listen")
	flag.Parse()

	host = *b
	port = *p
	if *d != "" {
		db = *d
		data = read(db)
		go persist()
	}
}

// Display bootup information such as version, mode, port, and pid.
func boot() {
	var mode string
	if db == "" {
		mode = "ephemeral"
	} else {
		mode = "durable"
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

// Ye olde main.
func main() {
	flags()
	boot()
	listen()
}
