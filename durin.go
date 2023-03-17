package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"sync"
	"time"
)

var version = "1.0.0"
var host = "localhost"
var port = 8045
var db string
var data = make(map[string]interface{})
var mu sync.Mutex
var cluster []string

// Durin HTTP API structures
type DurinRequest struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value,omitempty"`
}

type DurinError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type DurinSuccess struct {
	Key   string      `json:"key,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type DurinResponse struct {
	Data  *DurinSuccess `json:"data,omitempty"`
	Error *DurinError   `json:"error,omitempty"`
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

// Read the database file into memory.
func read(file string) map[string]interface{} {
	data, err := os.ReadFile(file)
	if err != nil {
		return make(map[string]interface{})
	}
	var result map[string]interface{}
	err = json.Unmarshal([]byte(string(data)), &result)
	if err != nil {
		log.Fatal("Error reading database file:", err)
	}
	return result
}

// Store the database to disk in json format.
func store(file string, data map[string]interface{}) {
	s, _ := json.MarshalIndent(data, "", "  ")
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

// Unpack the request body into a DurinRequest
func Unpack(r *http.Request) *DurinRequest {
	var durin DurinRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(body, &durin)
	if err != nil {
		log.Println(err)
	}

	return &durin
}

// Get a key from the database.
func get(w http.ResponseWriter, r *http.Request) {
	key := Unpack(r).Key
	var res DurinResponse

	if value, ok := data[key]; ok {
		res.Data = &DurinSuccess{
			Value: value,
		}
	} else {
		res.Error = &DurinError{
			Code:    500,
			Message: "key not found",
		}
	}

	b, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", string(b))
}

// Set a key in the database to a value.
func set(w http.ResponseWriter, r *http.Request) {
	req := Unpack(r)
	var res DurinResponse
	key, value := req.Key, req.Value

	if key == "" || value == "" || isNil(value) {
		res.Error = &DurinError{
			Code:    500,
			Message: "invalid parameters for set request",
		}
		log.Printf("value:'%v'\n", value)
	} else {
		data[key] = value
		res.Data = &DurinSuccess{
			Key: key,
		}
	}

	b, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", string(b))
}

// Delete a key and its value from the database.
func del(w http.ResponseWriter, r *http.Request) {
	key := Unpack(r).Key
	var res DurinResponse

	if key != "" {
		delete(data, key)
		res.Data = &DurinSuccess{
			Key: key,
		}
	} else {
		res.Error = &DurinError{
			Code:    500,
			Message: "del request requires key",
		}
	}

	b, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", string(b))
}

// Listen on the port specified by the port variable at the top of this file.
// We listen on the private loopback interface (i.e. localhost).
// Right now we just spin up a lightweight go routine for each connection.
func listen() {
	http.HandleFunc("/api/v1/set", set)
	http.HandleFunc("/api/v1/get", get)
	http.HandleFunc("/api/v1/del", del)
	addr := fmt.Sprintf("%s:%d", host, port)
	log.Fatal(http.ListenAndServe(addr, nil))
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
