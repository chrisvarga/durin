package perf

import (
	"fmt"
	"log"
	"net"
)

func PerfSet() {
	conn, err := net.Dial("tcp", "127.0.0.1:8045")
	if err != nil {
		log.Fatal("Failed to connect to Durin: ", err)
	}
	for i := 0; i < 10000; i++ {
		fmt.Fprintf(conn, "set %d bar\n", i)
	}
}

func PerfGet() {
	conn, err := net.Dial("tcp", "127.0.0.1:8045")
	if err != nil {
		log.Fatal("Failed to connect to Durin: ", err)
	}
	for i := 0; i < 10000; i++ {
		fmt.Fprintf(conn, "get %d\n", i)
	}
}
